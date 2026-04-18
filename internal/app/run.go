package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rneacsu/spyglass/internal/webview"
)

func Run(logger *slog.Logger, frontend webview.Frontend) error {
	app := NewApp(logger)

	if err := app.useInteractiveShellPath(); err != nil {
		logger.Warn("could not set PATH from interactive shell", "error", err)
	}

	wApp, err := webview.InitApplication(logger, frontend)
	if err != nil {
		return fmt.Errorf("could not initialize application: %w", err)
	}

	w, err := wApp.NewWindow("main", "/")
	if err != nil {
		return fmt.Errorf("could not create new window: %w", err)
	}

	w.SetTitle(AppName)
	w.SetMinSize(512, 512)

	getUrlFn := func() string {
		return app.grpcServer.GetUrl()
	}

	if err := w.Bind("GetGRPCUrl", getUrlFn); err != nil {
		return fmt.Errorf("could not bind GetGRPCUrl function: %w", err)
	}

	ctx := context.Background()
	app.Startup()
	err = wApp.Run(ctx)
	app.Shutdown()

	if err != nil {
		return fmt.Errorf("could not run application: %w", err)
	}

	return nil
}
