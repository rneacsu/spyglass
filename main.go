package main

import (
	"log/slog"

	"github.com/rneacsu/spyglass/internal/app"
	"github.com/rneacsu/spyglass/internal/webviewgo"
)

var logger *slog.Logger
var frontend webviewgo.Frontend

func main() {
	if err := app.Run(logger, frontend); err != nil {
		slog.Error("could not run application", "error", err)
	}
}
