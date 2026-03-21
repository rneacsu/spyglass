//go:build dev

package main

import (
	"log/slog"
	"os"

	"github.com/rneacsu/spyglass/internal/webviewgo"
)

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	frontend = webviewgo.NewExternalUrlFrontend(logger, "http://localhost:5173/")
}
