package main

import (
	"log/slog"

	"github.com/rneacsu/spyglass/internal/app"
)

func main() {
	if err := app.Run(logger, frontend); err != nil {
		slog.Error("could not run application", "error", err)
	}
}
