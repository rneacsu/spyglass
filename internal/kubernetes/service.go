package kubernetes

import (
	"context"
	"sort"

	"github.com/rneacsu/spyglass/internal/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	// MaxConnections is the maximum number of connections that can be active at the same time
	MaxConnections = 3
)

type KubeService struct {
	kubeConfig  *api.Config
	connections map[string]*KubeConnection
}

func NewKubeService() *KubeService {
	loader := clientcmd.NewDefaultClientConfigLoadingRules()

	kubeConfig, err := loader.Load()
	if err != nil {
		logger.Warnw("some errors encountered while loading default kubeconfig", "error", err)
	}

	return &KubeService{
		kubeConfig:  kubeConfig,
		connections: make(map[string]*KubeConnection),
	}
}

func (ks *KubeService) Stop() {
	for _, conn := range ks.connections {
		conn.Stop()
	}
}

func (ks *KubeService) GetContextNames() []string {
	contexts := make([]string, 0)

	for name := range ks.kubeConfig.Contexts {
		contexts = append(contexts, name)
	}
	sort.Strings(contexts)

	return contexts
}

func (ks *KubeService) GetDefaultContext() string {
	return ks.kubeConfig.CurrentContext
}

func (ks *KubeService) getConnection(kubeContext string) (*KubeConnection, error) {
	if connection, ok := ks.connections[kubeContext]; ok {
		return connection, nil
	}

	if len(ks.connections) >= MaxConnections {
		// Limit the number of connections to avoid performance and rate limiting issues
		var oldestConnection *KubeConnection
		var oldestKey string
		for k, c := range ks.connections {
			if oldestConnection == nil || c.LastUsed.Before(oldestConnection.LastUsed) {
				oldestConnection = c
				oldestKey = k
			}
		}

		oldestConnection.Stop()
		delete(ks.connections, oldestKey)
	}

	connection, err := NewKubeConnection(ks.kubeConfig, kubeContext)

	if err != nil {
		return nil, err
	}

	ks.connections[kubeContext] = connection

	return connection, nil
}

func (ks *KubeService) Discover(ctx context.Context, kubeContext string) ([]*metav1.APIResourceList, error) {
	conn, err := ks.getConnection(kubeContext)
	if err != nil {
		return nil, err
	}

	resources, err := conn.Discover(ctx)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (ks *KubeService) ListResource(ctx context.Context, kubeContext string, gvr schema.GroupVersionResource, namespace string) ([]*unstructured.Unstructured, error) {
	conn, err := ks.getConnection(kubeContext)
	if err != nil {
		return nil, err
	}

	watcher, err := conn.GetWatcher(gvr, namespace, WatcherTypeList)

	if err != nil {
		return nil, err
	}

	objects, err := watcher.(*ListWatcher).List(ctx)
	if err != nil {
		return nil, err
	}

	return objects, nil
}

func (ks *KubeService) ListResourceTabular(ctx context.Context, kubeContext string, gvr schema.GroupVersionResource, namespace string) (*metav1.Table, error) {
	conn, err := ks.getConnection(kubeContext)
	if err != nil {
		return nil, err
	}

	watcher, err := conn.GetWatcher(gvr, namespace, WatcherTypeTable)

	if err != nil {
		return nil, err
	}

	table, err := watcher.(*TableWatcher).GetTable(ctx)
	if err != nil {
		return nil, err
	}

	return table, nil
}
