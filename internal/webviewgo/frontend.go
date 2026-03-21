package webviewgo

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Frontend interface {
	Start() error
	Stop() error
	GetUrl() string
}

type BaseFrontend struct {
	url    string
	logger *slog.Logger
}

func (f *BaseFrontend) GetUrl() string {
	return f.url
}

type EmbeddedFrontend struct {
	*BaseFrontend
	assets *embed.FS
	prefix string
	url    string
	server *http.Server
	wg     sync.WaitGroup
}

func NewEmbeddedFrontend(logger *slog.Logger, assets *embed.FS, prefix string) *EmbeddedFrontend {
	if prefix[0] != '/' {
		prefix = "/" + prefix
	}

	f := &EmbeddedFrontend{
		BaseFrontend: &BaseFrontend{
			logger: logger,
		},
		assets: assets,
		prefix: prefix,
		server: &http.Server{
			Handler: http.FileServer(http.FS(assets)),
		},
	}
	return f
}

func (f *EmbeddedFrontend) GetUrl() string {
	return f.url
}

func (f *EmbeddedFrontend) Start() error {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return fmt.Errorf("could not open port for listening: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	f.url = fmt.Sprintf("http://localhost:%d%s", port, f.prefix)

	f.logger.Info("frontend starting", slog.String("url", f.url))

	f.wg.Go(func() {
		if err := f.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			f.logger.Error("frontend server error", slog.Any("error", err))
		}
	})

	return nil
}

func (f *EmbeddedFrontend) Stop() error {
	f.logger.Info("frontend shutting down")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := f.server.Shutdown(timeoutCtx)
	f.wg.Wait()
	return err
}

type ExternalUrlFrontend struct {
	*BaseFrontend
}

func NewExternalUrlFrontend(logger *slog.Logger, url string) *ExternalUrlFrontend {
	f := &ExternalUrlFrontend{
		BaseFrontend: &BaseFrontend{
			logger: logger,
			url:    strings.TrimSuffix(url, "/"),
		},
	}
	return f
}

func (f *ExternalUrlFrontend) Start() error {
	return nil
}

func (f *ExternalUrlFrontend) Stop() error {
	return nil
}

func (f *ExternalUrlFrontend) GetUrl() string {
	return f.url
}
