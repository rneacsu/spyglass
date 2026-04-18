//go:build dev

package main

import (
	"log/slog"
	"os"

	"github.com/rneacsu/spyglass/internal/webview"
)

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	frontend = webview.NewExternalUrlFrontend(logger, "http://localhost:5173/")
}
