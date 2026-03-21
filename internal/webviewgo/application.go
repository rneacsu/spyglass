package webviewgo

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type ApplicationImpl interface {
	newWindow(*Window) WindowImpl
	run(ctx context.Context) error
	init() error
}

type onApplicationStartHandler func()

type Application struct {
	ApplicationImpl
	logger                     *slog.Logger
	frontend                   Frontend
	windows                    map[string]*Window
	windowsLock                sync.Mutex
	hasWindows                 bool
	onApplicationStartHandlers []onApplicationStartHandler
}

var appInitOnce sync.Once
var app *Application
var appInitErr error

func InitApplication(logger *slog.Logger, frontend Frontend) (*Application, error) {
	appInitOnce.Do(func() {
		app = &Application{
			logger:                     logger.With("component", "webviewgo"),
			frontend:                   frontend,
			windows:                    make(map[string]*Window),
			hasWindows:                 false,
			onApplicationStartHandlers: make([]onApplicationStartHandler, 0),
		}

		app.ApplicationImpl = newApplicationImpl(app)

		logger.Info("initialising application")
		if err := app.init(); err != nil {
			appInitErr = fmt.Errorf("failed to initialize application: %w", err)
			app = nil
			return
		}
	})
	return app, appInitErr
}

func (a *Application) NewWindow(id string, path string) (*Window, error) {
	a.windowsLock.Lock()
	defer a.windowsLock.Unlock()

	if _, exists := a.windows[id]; exists {
		return nil, fmt.Errorf("window with id %s already exists", id)
	}

	w := newWindow(a, id, path)

	a.windows[id] = w
	a.hasWindows = true
	w.OnClose(func(id string) {
		a.windowsLock.Lock()
		defer a.windowsLock.Unlock()
		delete(a.windows, id)
		a.hasWindows = len(a.windows) > 0
	})

	return w, nil
}

func (a *Application) Run(ctx context.Context) error {
	if err := a.frontend.Start(); err != nil {
		return fmt.Errorf("could not start frontend: %w", err)
	}
	defer func() {
		if err := a.frontend.Stop(); err != nil {
			a.logger.Error("could not stop frontend", "error", err)
		}
	}()

	a.logger.Info("running application")
	for _, handler := range a.onApplicationStartHandlers {
		handler()
	}
	a.onApplicationStartHandlers = make([]onApplicationStartHandler, 0)

	return a.run(ctx)
}

func (a *Application) OnApplicationStart(handler onApplicationStartHandler) {
	a.onApplicationStartHandlers = append(a.onApplicationStartHandlers, handler)
}
