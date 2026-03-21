package app

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"

	"github.com/rneacsu/spyglass/internal/grpc"
)

var (
	AppName    = "SpyGlass"
	AppVersion = "0.0.0"
)

// App struct
type App struct {
	logger     *slog.Logger
	grpcServer *grpc.GRPCServer
}

// NewApp creates a new App application struct
func NewApp(logger *slog.Logger) *App {
	return &App{
		logger:     logger.With("component", "app"),
		grpcServer: grpc.NewGRPCServer(logger),
	}
}

func (a *App) useInteractiveShellPath() error {
	if runtime.GOOS != "darwin" {
		return nil
	}

	shell, exists := os.LookupEnv("SHELL")
	if !exists {
		shell = "/bin/zsh"
		a.logger.Warn("could not find SHELL environment variable, using fallback", "fallback", shell)
	}

	cmd := exec.Command(shell, "-i", "-c", "echo $PATH")
	output, err := cmd.Output()

	if err != nil {
		return fmt.Errorf("could not get PATH from shell: %w", err)
	} else {
		path := string(output)
		a.logger.Info("setting PATH from shell", "path", path)
		if err = os.Setenv("PATH", path); err != nil {
			return fmt.Errorf("could not set PATH from shell: %w", err)
		}
	}

	return nil
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup() {
	a.logger.Info("application starting up")

	// Start the gRPC server
	if err := a.grpcServer.Start(); err != nil {
		a.logger.Error("failed to start gRPC server", "error", err)
		os.Exit(1)
	}
	a.logger.Info("gRPC server started")
}

func (a *App) Shutdown() {
	a.logger.Info("Application shutting down")
	a.grpcServer.Stop()
}
