package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	connectcors "connectrpc.com/cors"
	"github.com/rneacsu/spyglass/internal/grpc/proto/protoconnect"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type GRPCServer struct {
	logger    *slog.Logger
	url       string
	server    *http.Server
	handler   *kubeHandler
	wgReady   sync.WaitGroup
	wgStopped sync.WaitGroup
}

func NewGRPCServer(logger *slog.Logger) *GRPCServer {
	server := &GRPCServer{
		logger:  logger.With("component", "grpc_server"),
		handler: NewKubeHandler(logger),
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
	path, handler := protoconnect.NewKubeHandler(s.handler)
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

	s.wgStopped.Go(func() {
		s.wgReady.Done()
		s.logger.Info("starting gRPC server", "url", s.url)
		if err := s.server.Serve(lis); !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Errorf("failed to start gRPC server: %w", err))
		}
	})

	return nil
}

func (s *GRPCServer) Stop() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("failed to shutdown gRPC server", "error", err)
	}
	s.wgStopped.Wait()
	s.handler.Stop()
}

func (s *GRPCServer) GetUrl() string {
	s.wgReady.Wait()
	return s.url
}
