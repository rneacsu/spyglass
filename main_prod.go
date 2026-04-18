//go:build !dev

package main

import (
	"embed"
	"log/slog"
	"os"

	"github.com/rneacsu/spyglass/internal/webview"
)

//go:embed all:frontend/build
var assets embed.FS

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	frontend = webview.NewEmbeddedFrontend(logger, &assets, "frontend/build")
}
