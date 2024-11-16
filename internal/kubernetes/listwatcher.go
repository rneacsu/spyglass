package kubernetes

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/rneacsu5/spyglass/internal/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type ListWatcher struct {
	*baseWatcher
	client      *dynamic.DynamicClient
	objListLock sync.RWMutex
	objList     map[string]*unstructured.Unstructured
}

func NewListWatcher(clientConfig *rest.Config, config WatcherConfig) (*ListWatcher, error) {
	client, err := dynamic.NewForConfig(clientConfig)

	if err != nil {
		return nil, err
	}

	return &ListWatcher{
		baseWatcher: NewBaseWatcher(config, WatcherTypeList),
		client:      client,
		objList:     make(map[string]*unstructured.Unstructured),
	}, nil
}

func (lw *ListWatcher) List(ctx context.Context) ([]*unstructured.Unstructured, error) {
	lw.watchLock.Lock()
	defer lw.watchLock.Unlock()

	if lw.watch == nil {
		// Start by listing the resources
		listOpt := metav1.ListOptions{}

		var listResult *unstructured.UnstructuredList
		var err error

		if lw.config.Namespace != "" {
			listResult, err = lw.client.Resource(lw.config.GVR).Namespace(lw.config.Namespace).List(ctx, listOpt)
		} else {
			listResult, err = lw.client.Resource(lw.config.GVR).List(ctx, listOpt)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list (context: %s, resource: %s): %w", lw.config.KubeContext, lw.config.GVR, err)
		}

		// Then start a background watch
		timeout := int64(DefaultWatchTimeout.Seconds())
		watchOpts := metav1.ListOptions{
			ResourceVersion: listResult.GetResourceVersion(),
			TimeoutSeconds:  &timeout,
		}
		watcher, err := lw.client.Resource(lw.config.GVR).Watch(context.Background(), watchOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to watch list (context: %s, resource: %s): %w", lw.config.KubeContext, lw.config.GVR, err)
		}

		lw.watch = watcher

		// Add the initial list to the object list
		lw.objListLock.Lock()
		for idk := range listResult.Items {
			obj := &listResult.Items[idk]
			lw.objList[string(obj.GetUID())] = obj
		}
		lw.objListLock.Unlock()

		lw.watchWG.Add(1)
		go func() {
			defer lw.watchWG.Done()
			logger.Infow("background watching started", lw.logContext...)

			for event := range watcher.ResultChan() {
				lw.objListLock.Lock()
				switch event.Type {
				case watch.Added, watch.Modified:
					obj := event.Object.(*unstructured.Unstructured)
					lw.objList[string(obj.GetUID())] = obj
				case watch.Deleted:
					obj := event.Object.(*unstructured.Unstructured)
					delete(lw.objList, string(obj.GetUID()))
				case watch.Error:
					var logMsg string
					if status, ok := event.Object.(*metav1.Status); ok {
						logMsg = fmt.Sprintf("watch event error: %s, reason: %s", status.Message, status.Reason)
					} else {
						logMsg = fmt.Sprintf("watch event error: %v", event.Object)
					}
					logger.Errorw(logMsg, lw.logContext...)
				}
				lw.objListLock.Unlock()
			}

			lw.watchLock.Lock()
			defer lw.watchLock.Unlock()
			lw.watch = nil

			logger.Infow("background watching finished", lw.logContext...)
		}()
	}

	lw.objListLock.RLock()
	defer lw.objListLock.RUnlock()
	resourceList := make([]*unstructured.Unstructured, 0, len(lw.objList))
	for _, obj := range lw.objList {
		resourceList = append(resourceList, obj)
	}

	// Sort by name
	sort.Slice(resourceList, func(i, j int) bool {
		return resourceList[i].GetName() < resourceList[j].GetName()
	})

	return resourceList, nil
}
