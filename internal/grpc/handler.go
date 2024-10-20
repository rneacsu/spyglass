package grpc

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rneacsu5/spyglass/internal/grpc/proto"
	"github.com/rneacsu5/spyglass/internal/kubernetes"
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

func (kh *kubeHandler) ListResource(ctx context.Context, req *connect.Request[proto.ListResourceRequest]) (*connect.Response[proto.ListResourceReply], error) {
	kubeContext := req.Msg.Context
	gvr := schema.GroupVersionResource{
		Group:    req.Msg.Gvr.Group,
		Version:  req.Msg.Gvr.Version,
		Resource: req.Msg.Gvr.Resource,
	}

	objs, err := kh.ks.ListResource(ctx, kubeContext, gvr)

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &proto.ListResourceReply{
		Resources: make([]*proto.Resource, 0, len(objs)),
	}

	for _, obj := range objs {
		response.Resources = append(response.Resources, &proto.Resource{
			Name: obj.GetName(),
			Gvr: &proto.GVR{
				Group:    obj.GroupVersionKind().Group,
				Version:  obj.GroupVersionKind().Version,
				Resource: obj.GroupVersionKind().Kind,
			},
		})
	}

	return connect.NewResponse(response), nil
}
