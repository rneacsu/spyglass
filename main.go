package main

import (
	"log/slog"

	"github.com/rneacsu/spyglass/internal/app"
	"github.com/rneacsu/spyglass/internal/webview"
)

var logger *slog.Logger
var frontend webview.Frontend

func main() {
	if err := app.Run(logger, frontend); err != nil {
		slog.Error("could not run application", "error", err)
	}
}
