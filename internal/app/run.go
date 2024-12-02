package app

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/rneacsu5/spyglass/internal/logger"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

func parseWailsConfig(wailsConfig []byte) (*AppInfo, error) {
	var config map[string]interface{}

	if err := json.Unmarshal(wailsConfig, &config); err != nil {
		return nil, fmt.Errorf("Could not parse Wails json: %w", err)
	}

	var info = config["info"].(map[string]interface{})

	return &AppInfo{
		Version:   info["productVersion"].(string),
		Name:      info["productName"].(string),
		Copyright: info["copyright"].(string),
	}, nil
}

func useLoginShellPath() error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "echo %PATH%")
	} else {
		shell, exists := os.LookupEnv("SHELL")
		if !exists {
			shell = "/bin/bash"
			logger.Warnf("could not find SHELL environment variable, using %s", shell)
		}
		cmd = exec.Command(shell, "-i", "-c", "echo $PATH")
	}

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("could not get PATH from shell: %w", err)
	} else {
		logger.Infof("setting PATH from shell: %s", output)
		path := string(output)
		if err = os.Setenv("PATH", path); err != nil {
			return fmt.Errorf("could not set PATH from shell: %w", err)
		}
	}

	return nil
}

type EmbeddedResources struct {
	Assets      embed.FS
	Icon        []byte
	WailsConfig []byte
}

func Run(emb EmbeddedResources) error {
	if err := logger.InitGlobalLogger(IsDev()); err != nil {
		return fmt.Errorf("could not initialize global logger: %w", err)
	}
	defer func() { _ = logger.GlobalSync() }()

	if err := useLoginShellPath(); err != nil {
		logger.Warnw("could not set PATH from shell", "error", err)
	}

	app := NewApp()

	info, err := parseWailsConfig(emb.WailsConfig)
	if err != nil {
		return fmt.Errorf("could not parse Wails configuration: %w", err)
	}
	app.info = info

	err = wails.Run(&options.App{
		Title:     info.Name,
		MinWidth:  512,
		MinHeight: 512,
		AssetServer: &assetserver.Options{
			Assets: emb.Assets,
		},
		Logger:     logger.NewWailsLogger(logger.GlobalLogger()),
		OnStartup:  app.Startup,
		OnShutdown: app.Shutdown,
		Bind: []interface{}{
			&AppApi{
				app: app,
			},
		},
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title:   info.Name,
				Message: fmt.Sprintf("Version: %s\n\n%s", info.Version, info.Copyright),
				Icon:    emb.Icon,
			},
		},
	})

	if err != nil {
		return fmt.Errorf("could not run application: %w", err)
	}

	return nil
}
