//go:build prod

package main

import (
	"embed"
	"log/slog"
	"os"

	"github.com/rneacsu/spyglass/internal/webviewgo"
)

//go:embed frontend/build
var assets embed.FS

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
var frontend = webviewgo.NewEmbeddedFrontend(logger, &assets, "frontend/build")
