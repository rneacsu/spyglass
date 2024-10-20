package kubernetes

import (
	"net"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	// DefaultDialTimeout is the default timeout for the connection
	DefaultDialTimeout = 5 * time.Second

	// MaxResourceWatchers is the maximum number of resource watchers that can be active at the same time
	MaxResourceWatchers = 10
)

type KubeConnection struct {
	kubeContext      string
	client           *dynamic.DynamicClient
	resourceWatchers map[string]*ResourceWatcher

	LastUsed time.Time
}

func NewKubeConnection(kubeConfig *api.Config, kubeContext string) (*KubeConnection, error) {
	clientConfig, err := clientcmd.NewDefaultClientConfig(*kubeConfig, &clientcmd.ConfigOverrides{
		CurrentContext: kubeContext,
	}).ClientConfig()

	if err != nil {
		return nil, err
	}

	clientConfig.Dial = (&net.Dialer{
		Timeout: DefaultDialTimeout,
	}).DialContext
	client, err := dynamic.NewForConfig(clientConfig)

	if err != nil {
		return nil, err
	}

	return &KubeConnection{
		kubeContext:      kubeContext,
		client:           client,
		resourceWatchers: make(map[string]*ResourceWatcher),
		LastUsed:         time.Now(),
	}, nil
}

func (kc *KubeConnection) GetResourceWatcher(gvr schema.GroupVersionResource) *ResourceWatcher {
	kc.LastUsed = time.Now()

	key := gvr.String()
	if watcher, ok := kc.resourceWatchers[key]; ok {
		return watcher
	}

	if len(kc.resourceWatchers) > MaxResourceWatchers {
		// Limit the number of watchers to avoid performance and rate limiting issues
		var oldestWatcher *ResourceWatcher
		var oldestKey string
		for k, w := range kc.resourceWatchers {
			if oldestWatcher == nil || w.LastUsed.Before(oldestWatcher.LastUsed) {
				oldestWatcher = w
				oldestKey = k
			}
		}

		oldestWatcher.Stop()
		delete(kc.resourceWatchers, oldestKey)
	}

	watcher := NewResourceWatcher(kc.client, kc.kubeContext, gvr)
	kc.resourceWatchers[key] = watcher

	return watcher
}

func (kc *KubeConnection) Stop() {
	for _, watcher := range kc.resourceWatchers {
		watcher.Stop()
	}
}
