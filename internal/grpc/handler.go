package grpc

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/rneacsu5/spyglass/internal/grpc/proto"
	"github.com/rneacsu5/spyglass/internal/kubernetes"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type kubeHandler struct {
	ks *kubernetes.KubeService
}

func NewKubeHandler() *kubeHandler {
	return &kubeHandler{
		ks: kubernetes.NewKubeService(),
	}
}

func (kh *kubeHandler) Stop() {
	kh.ks.Stop()
}

func (kh *kubeHandler) GetContexts(context.Context, *connect.Request[proto.Empty]) (*connect.Response[proto.ContextsReply], error) {
	return connect.NewResponse(&proto.ContextsReply{Contexts: kh.ks.GetContextNames()}), nil
}

func (kh *kubeHandler) GetDefaultContext(context.Context, *connect.Request[proto.Empty]) (*connect.Response[proto.ContextReply], error) {
	return connect.NewResponse(&proto.ContextReply{Context: kh.ks.GetDefaultContext()}), nil
}

func (kh *kubeHandler) Discover(ctx context.Context, req *connect.Request[proto.DiscoverRequest]) (*connect.Response[proto.DiscoverReply], error) {
	kubeContext := req.Msg.Context

	resources, err := kh.ks.Discover(ctx, kubeContext)

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &proto.DiscoverReply{
		Apis: make(map[string]*proto.DiscoverApi, len(resources)),
	}

	for _, resource := range resources {
		gvSplit := strings.Split(resource.GroupVersion, "/")
		var group string
		var version string
		if len(gvSplit) == 2 {
			group = gvSplit[0]
			version = gvSplit[1]
		} else {
			group = ""
			version = gvSplit[0]
		}

		api := &proto.DiscoverApi{
			Group:     group,
			Version:   version,
			Resources: make([]*proto.DiscoverResource, 0, len(resource.APIResources)),
		}

		for _, res := range resource.APIResources {
			api.Resources = append(api.Resources, &proto.DiscoverResource{
				Name:       res.Name,
				Namespaced: res.Namespaced,
			})
		}

		response.Apis[resource.GroupVersion] = api
	}

	return connect.NewResponse(response), nil
}

func (kh *kubeHandler) ListResource(ctx context.Context, req *connect.Request[proto.ListResourceRequest]) (*connect.Response[proto.ListResourceReply], error) {
	kubeContext := req.Msg.Context
	gvr := schema.GroupVersionResource{
		Group:    req.Msg.Gvr.Group,
		Version:  req.Msg.Gvr.Version,
		Resource: req.Msg.Gvr.Resource,
	}

	namespace := ""
	if req.Msg.Namespace != nil {
		namespace = *req.Msg.Namespace
	}

	objs, err := kh.ks.ListResource(ctx, kubeContext, gvr, namespace)

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &proto.ListResourceReply{
		Resources: make([]*proto.Resource, 0, len(objs)),
	}

	for _, obj := range objs {
		raw, err := structpb.NewStruct(obj.Object)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		response.Resources = append(response.Resources, &proto.Resource{
			Name: obj.GetName(),
			Gvk: &proto.GVK{
				Group:   obj.GroupVersionKind().Group,
				Version: obj.GroupVersionKind().Version,
				Kind:    obj.GroupVersionKind().Kind,
			},
			Namespace: obj.GetNamespace(),
			Raw:       raw,
		})
	}

	return connect.NewResponse(response), nil
}

func (kh *kubeHandler) ListResourceTabular(ctx context.Context, req *connect.Request[proto.ListResourceRequest]) (*connect.Response[proto.ListResourceTabularReply], error) {
	kubeContext := req.Msg.Context
	gvr := schema.GroupVersionResource{
		Group:    req.Msg.Gvr.Group,
		Version:  req.Msg.Gvr.Version,
		Resource: req.Msg.Gvr.Resource,
	}

	namespace := ""
	if req.Msg.Namespace != nil {
		namespace = *req.Msg.Namespace
	}

	table, err := kh.ks.ListResourceTabular(ctx, kubeContext, gvr, namespace)

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &proto.ListResourceTabularReply{
		Columns: make([]*proto.ListResourceTabularReply_TabularColumn, 0, len(table.ColumnDefinitions)),
		Rows:    make([]*proto.ListResourceTabularReply_TabularRow, 0, len(table.Rows)),
	}

	for _, col := range table.ColumnDefinitions {
		response.Columns = append(response.Columns, &proto.ListResourceTabularReply_TabularColumn{
			Name: col.Name,
			Type: col.Type,
		})
	}
	for _, row := range table.Rows {
		r := &proto.ListResourceTabularReply_TabularRow{
			Cells: make([]string, 0, len(row.Cells)),
		}
		for _, cell := range row.Cells {
			r.Cells = append(r.Cells, fmt.Sprintf("%v", cell))
		}
		pom := row.Object.Object.(*v1.PartialObjectMetadata)

		objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pom)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		raw, err := structpb.NewStruct(objMap)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}

		r.Resource = &proto.Resource{
			Name:      pom.Name,
			Namespace: pom.Namespace,
			Gvk: &proto.GVK{
				Group:   pom.GroupVersionKind().Group,
				Version: pom.GroupVersionKind().Version,
				Kind:    pom.GroupVersionKind().Kind,
			},
			Raw:     raw,
			Created: timestamppb.New(pom.CreationTimestamp.Time),
		}
		response.Rows = append(response.Rows, r)
	}

	return connect.NewResponse(response), nil
}
