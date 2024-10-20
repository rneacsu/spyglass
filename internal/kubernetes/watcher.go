package kubernetes

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/rneacsu5/spyglass/internal/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

const (
	DefaultWatchTimeout = 120 * time.Second
)

type ResourceWatcher struct {
	kubeContext string
	client      *dynamic.DynamicClient
	gkr         schema.GroupVersionResource
	watch       watch.Interface
	watchLock   sync.Mutex
	watchWG     sync.WaitGroup
	objListLock sync.RWMutex
	objList     map[string]*unstructured.Unstructured

	LastUsed time.Time
}

func NewResourceWatcher(client *dynamic.DynamicClient, kubeContext string, gkr schema.GroupVersionResource) *ResourceWatcher {
	return &ResourceWatcher{
		kubeContext: kubeContext,
		client:      client,
		gkr:         gkr,
		objList:     make(map[string]*unstructured.Unstructured),
		LastUsed:    time.Now(),
	}
}

func (rw *ResourceWatcher) Stop() {
	rw.watchLock.Lock()
	if rw.watch != nil {
		rw.watch.Stop()
	}
	rw.watchLock.Unlock()
	rw.watchWG.Wait()
	logger.Infow("Stopped watching", "context", rw.kubeContext, "resource", rw.gkr)
}

func (rw *ResourceWatcher) ListAll(ctx context.Context) ([]*unstructured.Unstructured, error) {
	rw.LastUsed = time.Now()

	rw.watchLock.Lock()
	defer rw.watchLock.Unlock()

	if rw.watch == nil {
		// If no context is set, we need to start the watch

		// Start by listing the resources
		listOpt := metav1.ListOptions{}

		listResult, err := rw.client.Resource(rw.gkr).List(ctx, listOpt)

		if err != nil {
			return nil, fmt.Errorf("failed to list (context: %s, resource: %s): %w", rw.kubeContext, rw.gkr, err)
		}

		// Then start a background watch
		timeout := int64(DefaultWatchTimeout.Seconds())
		watchOpts := metav1.ListOptions{
			ResourceVersion: listResult.GetResourceVersion(),
			TimeoutSeconds:  &timeout,
		}
		watcher, err := rw.client.Resource(rw.gkr).Watch(context.Background(), watchOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to watch (context: %s, resource: %s): %w", rw.kubeContext, rw.gkr, err)
		}

		rw.watch = watcher

		// Add the initial list to the object list
		rw.objListLock.Lock()
		for idk := range listResult.Items {
			obj := &listResult.Items[idk]
			rw.objList[string(obj.GetUID())] = obj
		}
		rw.objListLock.Unlock()

		rw.watchWG.Add(1)
		go func() {
			defer rw.watchWG.Done()
			logger.Infow("background watching started", "context", rw.kubeContext, "resource", rw.gkr)

			for event := range watcher.ResultChan() {
				rw.objListLock.Lock()
				switch event.Type {
				case watch.Added, watch.Modified:
					obj := event.Object.(*unstructured.Unstructured)
					rw.objList[string(obj.GetUID())] = obj
				case watch.Deleted:
					obj := event.Object.(*unstructured.Unstructured)
					delete(rw.objList, string(obj.GetUID()))
				case watch.Error:
					var logMsg string
					if status, ok := event.Object.(*metav1.Status); ok {
						logMsg = fmt.Sprintf("watch event error: %s, reason: %s", status.Message, status.Reason)
					} else {
						logMsg = fmt.Sprintf("watch event error: %v", event.Object)
					}
					logger.Errorw(logMsg, "context", rw.kubeContext, "resource", rw.gkr)
				}
				rw.objListLock.Unlock()
			}

			rw.watchLock.Lock()
			defer rw.watchLock.Unlock()
			rw.watch = nil

			logger.Infow("background watching finished", "context", rw.kubeContext, "resource", rw.gkr)
		}()
	}

	rw.objListLock.RLock()
	defer rw.objListLock.RUnlock()
	resourceList := make([]*unstructured.Unstructured, 0, len(rw.objList))
	for _, obj := range rw.objList {
		resourceList = append(resourceList, obj)
	}

	// Sort by name
	sort.Slice(resourceList, func(i, j int) bool {
		return resourceList[i].GetName() < resourceList[j].GetName()
	})

	return resourceList, nil
}
