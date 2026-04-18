package webviewgo

/*
#cgo LDFLAGS: -ldl
#cgo pkg-config: gtk4 webkitgtk-6.0

#include <gtk/gtk.h>
#include <webkit/webkit.h>

*/
import "C"
import (
	"context"
	"fmt"
	"runtime"
	"sync"
)

type LinuxApplication struct {
	app              *Application
	mainThreadWG     sync.WaitGroup
	mainThreadEvents chan func()
	mainThreadDone   chan struct{}
}

func newApplicationImpl(app *Application) ApplicationImpl {
	return &LinuxApplication{
		app:              app,
		mainThreadEvents: make(chan func()),
		mainThreadDone:   make(chan struct{}),
	}
}

func (a *LinuxApplication) init() error {
	var initResult C.gboolean
	var initWG sync.WaitGroup
	initWG.Add(1)

	a.mainThreadWG.Go(func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		a.app.logger.Info("main loop started")

		a.app.logger.Info("initialising GTK")
		initResult = C.gtk_init_check()
		initWG.Done()
		if initResult == 0 {
			a.app.logger.Error("failed to initialise GTK")
			return
		}

		for {
			select {
			case f := <-a.mainThreadEvents:
				f()
			case <-a.mainThreadDone:
				a.app.logger.Info("exiting main loop")
				return
			}
		}
	})

	initWG.Wait()
	if initResult == 0 {
		a.mainThreadWG.Wait()
		return fmt.Errorf("failed to initialize GTK")
	}
	return nil
}

func (a *LinuxApplication) newWindow(win *Window) WindowImpl {
	w := newLinuxWindow(a, win)
	return w
}

func (a *LinuxApplication) run(ctx context.Context) error {
	a.app.logger.Info("running GTK application")
	running := true
	for running {
		select {
		case <-ctx.Done():
			a.app.logger.Info("context cancelled, exiting application")
			a.mainThreadDone <- struct{}{}
			running = false
		default:
			if !a.app.hasWindows {
				a.app.logger.Info("no windows, exiting application")
				a.mainThreadDone <- struct{}{}
				running = false
			} else {
				a.mainThreadRun(func() {
					C.g_main_context_iteration(nil, 0)
				})
			}
		}
	}
	a.mainThreadWG.Wait()
	return nil
}

func (a *LinuxApplication) mainThreadRun(f func()) {
	var wg sync.WaitGroup
	wg.Add(1)
	a.mainThreadEvents <- func() {
		defer wg.Done()
		f()
	}
	wg.Wait()
}
