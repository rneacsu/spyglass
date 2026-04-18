package webview

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type onApplicationStartHandler func()

type Application struct {
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

		logger.Info("initialising application")
	})
	return app, appInitErr
}

func (a *Application) NewWindow(id string, path string) (*Window, error) {
	a.windowsLock.Lock()
	defer a.windowsLock.Unlock()

	if _, exists := a.windows[id]; exists {
		return nil, fmt.Errorf("window with id %s already exists", id)
	}

	w, err := newWindow(a, id, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}

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
	if !a.hasWindows {
		return fmt.Errorf("no windows to run")
	}

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

	// Run each window in a separate go routine, wait for them and collect all the errors
	wg := sync.WaitGroup{}
	errChan := make(chan error, len(a.windows))
	for _, window := range a.windows {
		wg.Add(1)
		go func(window *Window) {
			defer wg.Done()
			err := window.run(ctx)
			if err != nil {
				errChan <- err
			}
		}(window)
	}
	wg.Wait()
	close(errChan)

	// Collect all the errors
	errs := make([]error, 0)
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("application encountered errors: %v", errs)
	}

	return nil
}

func (a *Application) OnApplicationStart(handler onApplicationStartHandler) {
	a.onApplicationStartHandlers = append(a.onApplicationStartHandlers, handler)
}
