package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/rneacsu5/spyglass/internal/logger"
	metainternalversionscheme "k8s.io/apimachinery/pkg/apis/meta/internalversion/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
)

type TableWatcher struct {
	*baseWatcher

	client    *rest.RESTClient
	tableLock sync.RWMutex
	table     metav1.Table
}

func NewTableWatcher(clientConfig *rest.Config, config WatcherConfig) (*TableWatcher, error) {
	restConfig := rest.CopyConfig(clientConfig)
	restConfig.AcceptContentTypes = "application/json;as=Table;v=v1;g=meta.k8s.io,application/json"
	restConfig.ContentType = "application/json"
	if restConfig.UserAgent == "" {
		restConfig.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	restConfig.GroupVersion = &schema.GroupVersion{}
	var apiPath string
	if config.GVR.Group != "" {
		apiPath = "/apis/" + config.GVR.Group
	} else {
		apiPath = "/api"
	}
	apiPath = apiPath + "/" + config.GVR.Version
	restConfig.APIPath = apiPath

	restConfig.NegotiatedSerializer = metainternalversionscheme.Codecs.WithoutConversion()

	restClient, err := rest.RESTClientFor(restConfig)
	if err != nil {
		return nil, err
	}

	return &TableWatcher{
		baseWatcher: NewBaseWatcher(config, WatcherTypeTable),
		client:      restClient,
	}, nil
}

func decodeTableRows(table *metav1.Table) error {
	for i := range table.Rows {
		pom := &metav1.PartialObjectMetadata{}
		err := json.Unmarshal(table.Rows[i].Object.Raw, pom)
		if err != nil {
			return fmt.Errorf("failed to decode object metadata: %w", err)
		}
		table.Rows[i].Object.Object = pom
	}
	return nil
}

func (tw *TableWatcher) GetTable(ctx context.Context) (*metav1.Table, error) {
	tw.watchLock.Lock()
	defer tw.watchLock.Unlock()

	if tw.watch == nil {
		// Start by listing the resources
		listOpt := metav1.ListOptions{}

		listRequest := tw.client.Get()
		if tw.config.Namespace != "" {
			listRequest = listRequest.Namespace(tw.config.Namespace)
		}
		listRequest = listRequest.Resource(tw.config.GVR.Resource).SpecificallyVersionedParams(&listOpt, metav1.ParameterCodec, metav1.Unversioned)

		logger.Info(listRequest.URL().String())

		listResult := metav1.Table{}
		err := listRequest.Do(ctx).Into(&listResult)

		if err != nil {
			return nil, fmt.Errorf("failed to list (context: %s, resource: %s, namespace: %s): %w", tw.config.KubeContext, tw.config.GVR, tw.config.Namespace, err)
		}

		// Then start a background watch
		timeout := int64(DefaultWatchTimeout.Seconds())
		watchOpts := metav1.ListOptions{
			Watch:           true,
			ResourceVersion: listResult.GetResourceVersion(),
			TimeoutSeconds:  &timeout,
		}

		watchRequest := tw.client.Get()
		if tw.config.Namespace != "" {
			watchRequest = watchRequest.Namespace(tw.config.Namespace)
		}

		watcher, err := watchRequest.Resource(tw.config.GVR.Resource).SpecificallyVersionedParams(&watchOpts, metav1.ParameterCodec, metav1.Unversioned).Watch(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to watch (context: %s, resource: %s, namespace: %s): %w", tw.config.KubeContext, tw.config.Namespace, tw.config.GVR, err)
		}

		tw.watch = watcher

		// Add the initial table
		tw.tableLock.Lock()
		tw.table = listResult
		if err = decodeTableRows(&tw.table); err != nil {
			return nil, err
		}

		tw.tableLock.Unlock()

		tw.watchWG.Add(1)
		go func() {
			defer tw.watchWG.Done()
			logger.Infow("background watching started", tw.logContext...)

			for event := range watcher.ResultChan() {
				tw.tableLock.Lock()

				switch event.Type {
				case watch.Added, watch.Modified, watch.Deleted:
					table := event.Object.(*metav1.Table)
					if err = decodeTableRows(table); err != nil {
						logger.Errorw(fmt.Sprintf("failed to decode table rows: %v", err), tw.logContext...)
						tw.tableLock.Unlock()
						continue
					}
					tableRow := table.Rows[0]
					objUID := tableRow.Object.Object.(*metav1.PartialObjectMetadata).UID

					switch event.Type {
					case watch.Added:
						tw.table.Rows = append(tw.table.Rows, tableRow)
					case watch.Modified, watch.Deleted:
						for i := range tw.table.Rows {
							curObjId := tw.table.Rows[i].Object.Object.(*metav1.PartialObjectMetadata).UID

							if curObjId == objUID {
								switch event.Type {
								case watch.Modified:
									tw.table.Rows[i] = tableRow
								case watch.Deleted:
									tw.table.Rows = append(tw.table.Rows[:i], tw.table.Rows[i+1:]...)
								}
								break
							}
						}
					}

				case watch.Error:
					var logMsg string
					if status, ok := event.Object.(*metav1.Status); ok {
						logMsg = fmt.Sprintf("watch event error: %s, reason: %s", status.Message, status.Reason)
					} else {
						logMsg = fmt.Sprintf("watch event error: %v", event.Object)
					}
					logger.Errorw(logMsg, tw.logContext...)
				}

				tw.tableLock.Unlock()
			}

			tw.watchLock.Lock()
			defer tw.watchLock.Unlock()
			tw.watch = nil

			logger.Infow("background watching finished", tw.logContext...)
		}()
	}

	tw.tableLock.RLock()
	defer tw.tableLock.RUnlock()

	tableResult := &metav1.Table{
		ColumnDefinitions: tw.table.ColumnDefinitions,
		Rows:              make([]metav1.TableRow, len(tw.table.Rows)),
	}
	copy(tableResult.Rows, tw.table.Rows)

	// Sort by name
	sort.Slice(tableResult.Rows, func(i, j int) bool {
		aName := tableResult.Rows[i].Object.Object.(*metav1.PartialObjectMetadata).Name
		bName := tableResult.Rows[j].Object.Object.(*metav1.PartialObjectMetadata).Name
		return aName < bName
	})

	return tableResult, nil
}
