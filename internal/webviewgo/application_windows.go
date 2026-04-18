package webviewgo

/*
#cgo CFLAGS: -I${SRCDIR}/../../libs/webview2/build/native/include
#cgo CXXFLAGS: -I${SRCDIR}/../../libs/webview2/build/native/include -std=c++20
#cgo LDFLAGS: -ldl

#include "webviewgo_windows.h"

*/
import "C"
import (
	"context"
	"fmt"
)

type WindowsApplication struct {
	app *Application
}

func newApplicationImpl(app *Application) ApplicationImpl {
	return &WindowsApplication{
		app: app,
	}
}

func (a *WindowsApplication) init() error {

	rt := C.wv2_init()

	if rt != 0 {
		return fmt.Errorf("failed to initialise WebView2")
	}

	return nil
}

func (a *WindowsApplication) newWindow(win *Window) WindowImpl {
	w := newWindowsWindow(a, win)
	return w
}

func (a *WindowsApplication) run(ctx context.Context) error {

	return nil
}
