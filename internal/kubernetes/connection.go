package kubernetes

import (
	"context"
	"fmt"
	"net"
	"path"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	// DefaultDialTimeout is the default timeout for the connection
	DefaultDialTimeout = 5 * time.Second

	// MaxResourceWatchers is the maximum number of resource watchers that can be active at the same time
	MaxWatchers = 10
)

type KubeConnection struct {
	kubeContext  string
	clientConfig *rest.Config
	watchers     map[string]Watcher
	discovery    *disk.CachedDiscoveryClient

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

	// TODO: use XDG Base Directory Specification
	discoveryCacheDir := path.Join(clientcmd.RecommendedConfigDir, "cache", "discovery")

	discoveryClient, err := disk.NewCachedDiscoveryClientForConfig(clientConfig, discoveryCacheDir, "", 10*time.Minute)

	if err != nil {
		return nil, err
	}

	return &KubeConnection{
		kubeContext:  kubeContext,
		clientConfig: clientConfig,
		watchers:     make(map[string]Watcher),
		discovery:    discoveryClient,
		LastUsed:     time.Now(),
	}, nil
}

func (kc *KubeConnection) Discover(ctx context.Context) ([]*metav1.APIResourceList, error) {
	kc.LastUsed = time.Now()

	resources, err := kc.discovery.ServerPreferredResources()

	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (kc *KubeConnection) GetWatcher(gvr schema.GroupVersionResource, namespace string, watcherType WatcherType) (Watcher, error) {
	kc.LastUsed = time.Now()

	key := FormatWatcherID(gvr, namespace, watcherType)

	if watcher, ok := kc.watchers[key]; ok {
		watcher.UpdateLastUsed()
		return watcher, nil
	}

	if len(kc.watchers) > MaxWatchers {
		// Limit the number of watchers to avoid performance and rate limiting issues
		var oldestWatcher Watcher
		var oldestKey string
		for k, w := range kc.watchers {
			if oldestWatcher == nil || w.GetLastUsed().Before(oldestWatcher.GetLastUsed()) {
				oldestWatcher = w
				oldestKey = k
			}
		}

		oldestWatcher.Stop()
		delete(kc.watchers, oldestKey)
	}

	watcherConfig := WatcherConfig{
		KubeContext: kc.kubeContext,
		GVR:         gvr,
		Namespace:   namespace,
	}

	var watcher Watcher
	var err error

	switch watcherType {
	case WatcherTypeList:
		watcher, err = NewListWatcher(kc.clientConfig, watcherConfig)
	case WatcherTypeTable:
		watcher, err = NewTableWatcher(kc.clientConfig, watcherConfig)
	default:
		err = fmt.Errorf("unsupported watcher type: %s", watcherType)
	}

	if err != nil {
		return nil, err
	}

	kc.watchers[key] = watcher

	return watcher, nil
}

func (kc *KubeConnection) Stop() {
	for _, watcher := range kc.watchers {
		watcher.Stop()
	}
}
