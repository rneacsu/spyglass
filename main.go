package main

import (
	"embed"
	"log"

	"github.com/rneacsu/spyglass/internal/app"
)

//go:embed all:frontend/build
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

//go:embed wails.json
var wailsConfig []byte

func main() {
	if err := app.Run(app.EmbeddedResources{
		Assets:      assets,
		Icon:        icon,
		WailsConfig: wailsConfig,
	}); err != nil {
		log.Fatalf("Could not run application: %v", err)
	}
}
