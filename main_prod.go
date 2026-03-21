//go:build !dev

package main

import (
	"embed"
	"log/slog"
	"os"

	"github.com/rneacsu/spyglass/internal/webviewgo"
)

//go:embed all:frontend/build
var assets embed.FS

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	frontend = webviewgo.NewEmbeddedFrontend(logger, &assets, "frontend/build")
}
