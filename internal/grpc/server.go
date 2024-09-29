package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"github.com/rneacsu5/spyglass/internal/grpc/proto"
	"github.com/rneacsu5/spyglass/internal/grpc/proto/protoconnect"
	"github.com/rneacsu5/spyglass/internal/logger"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type GRPCServer struct {
	url       string
	server    *http.Server
	handler   *greeterHandler
	wgReady   sync.WaitGroup
	wgStopped sync.WaitGroup
}

type greeterHandler struct {
}

// GetContexts implements protoconnect.GreeterHandler.
func (s *greeterHandler) GetContexts(context.Context, *connect.Request[proto.Empty]) (*connect.Response[proto.GetContextsReply], error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.LoadFromFile(kubeconfig)

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to load kubeconfig: %w", err))
	}

	contexts := make([]string, 0, len(config.Contexts))
	for context := range config.Contexts {
		contexts = append(contexts, context)
	}

	return connect.NewResponse(&proto.GetContextsReply{Contexts: contexts}), nil
}

func (s *greeterHandler) SayHello(ctx context.Context, req *connect.Request[proto.HelloRequest]) (*connect.Response[proto.HelloReply], error) {
	return connect.NewResponse(&proto.HelloReply{Message: "Hello " + req.Msg.Name + "! The gRPC connection is working 🥳"}), nil
}

func NewGRPCServer() *GRPCServer {
	server := &GRPCServer{
		handler: &greeterHandler{},
	}
	server.wgReady.Add(1)
	return server
}

func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.url = fmt.Sprintf("http://localhost:%d", lis.Addr().(*net.TCPAddr).Port)

	mux := http.NewServeMux()
	path, handler := protoconnect.NewGreeterHandler(s.handler)
	mux.Handle(path, handler)
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
	})

	h2s := &http2.Server{}
	s.server = &http.Server{
		Handler: h2c.NewHandler(corsMiddleware.Handler(mux), h2s),
	}
	if err := http2.ConfigureServer(s.server, h2s); err != nil {
		return fmt.Errorf("failed to configure http server: %w", err)
	}

	s.wgStopped.Add(1)
	go func() {
		defer s.wgStopped.Done()
		s.wgReady.Done()
		logger.Infof("Starting gRPC server on %s", s.url)
		if err := s.server.Serve(lis); !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	return nil
}

func (s *GRPCServer) Stop() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("Failed to shutdown gRPC server: %v", err)
	}
	s.wgStopped.Wait()
}

func (s *GRPCServer) GetUrl() string {
	s.wgReady.Wait()
	return s.url
}
