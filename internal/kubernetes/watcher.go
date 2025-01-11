package kubernetes

import (
	"sync"
	"time"

	"github.com/rneacsu/spyglass/internal/logger"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	DefaultWatchTimeout = 120 * time.Second
)

type Watcher interface {
	Stop()
	GetLastUsed() time.Time
	UpdateLastUsed()
	GetType() WatcherType
	GetID() string
}

type WatcherType string

const (
	WatcherTypeList     WatcherType = "list"
	WatcherTypeTable    WatcherType = "table"
	WatcherTypeResource WatcherType = "resource"
)

type WatcherConfig struct {
	KubeContext string
	GVR         schema.GroupVersionResource
	Namespace   string

	watcherType WatcherType
}

type baseWatcher struct {
	config     WatcherConfig
	watch      watch.Interface
	watchLock  sync.Mutex
	watchWG    sync.WaitGroup
	logContext []interface{}

	LastUsed time.Time
}

func NewBaseWatcher(config WatcherConfig, watcherType WatcherType) *baseWatcher {
	config.watcherType = watcherType
	return &baseWatcher{
		config:   config,
		LastUsed: time.Now(),
		logContext: []interface{}{
			"context", config.KubeContext,
			"resource", config.GVR,
			"type", watcherType,
		},
	}
}

func (bw *baseWatcher) GetLastUsed() time.Time {
	return bw.LastUsed
}

func (bw *baseWatcher) UpdateLastUsed() {
	bw.LastUsed = time.Now()
}

func (bw *baseWatcher) Stop() {
	bw.watchLock.Lock()
	if bw.watch != nil {
		bw.watch.Stop()
	}
	bw.watchLock.Unlock()
	bw.watchWG.Wait()
	logger.Infow("Stopped watching", bw.logContext...)
}

func (bw *baseWatcher) GetType() WatcherType {
	return bw.config.watcherType
}

func (bw *baseWatcher) GetID() string {
	return FormatWatcherID(bw.config.GVR, bw.config.Namespace, bw.config.watcherType)
}

func FormatWatcherID(gvr schema.GroupVersionResource, namespace string, watcherType WatcherType) string {
	return gvr.String() + "#" + namespace + "#" + string(watcherType)
}
