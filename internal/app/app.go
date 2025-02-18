package app

import (
	"context"

	"github.com/rneacsu/spyglass/internal/grpc"
	"github.com/rneacsu/spyglass/internal/logger"
)

var (
	AppEnv = "prod"
)

func IsDev() bool {
	return AppEnv == "dev"
}

// App struct
type App struct {
	ctx        context.Context
	grpcServer *grpc.GRPCServer
	info       *AppInfo
}

type AppInfo struct {
	Version   string
	Name      string
	Copyright string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		grpcServer: grpc.NewGRPCServer(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	logger.Info("Application starting up")

	// Start the gRPC server
	if err := a.grpcServer.Start(); err != nil {
		logger.Fatalf("Failed to start gRPC server: %v", err)
	}
	logger.Info("gRPC server started")
}

func (a *App) Shutdown(ctx context.Context) {
	logger.Info("Application shutting down")
	a.grpcServer.Stop(ctx)
}
