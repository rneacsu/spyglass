package webview

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

/*

#cgo CFLAGS: -I${SRCDIR}/../../libs/webview/core/include/webview
#cgo CXXFLAGS: -I${SRCDIR}/../../libs/webview/core/include/webview -std=c++14 -DWEBVIEW_STATIC

#cgo linux LDFLAGS: -ldl
#cgo linux pkg-config: gtk4 webkitgtk-6.0

#cgo darwin LDFLAGS: -framework WebKit -ldl

#cgo windows CXXFLAGS: -I${SRCDIR}/../../libs/webview/core/include/webview -I${SRCDIR}/../../libs/webview2/build/native/include -std=c++14 -DWEBVIEW_STATIC
#cgo windows LDFLAGS: -static -ladvapi32 -lole32 -lshell32 -lshlwapi -luser32 -lversion

#include "webview.h"

#include <stdlib.h>
#include <stdint.h>

*/
import "C"

type WindowOnCloseHandler func(id string)

type scriptBinding struct {
	f reflect.Value
}

type Window struct {
	app             *Application
	logger          *slog.Logger
	id              string
	path            string
	wv              C.webview_t
	onCloseHandlers []WindowOnCloseHandler
	scriptBindings  map[string]scriptBinding
	running         bool
	uiThreadWG      sync.WaitGroup
	uiThreadEvents  chan func()
	uiThreadDone    chan struct{}
}

func newWindow(app *Application, id string, path string) (*Window, error) {
	if path[0] != '/' {
		path = "/" + path
	}

	w := &Window{
		app:             app,
		id:              id,
		path:            path,
		logger:          app.logger.With("window", id),
		running:         false,
		onCloseHandlers: make([]WindowOnCloseHandler, 0),
		scriptBindings:  make(map[string]scriptBinding),
		uiThreadEvents:  make(chan func()),
		uiThreadDone:    make(chan struct{}),
	}

	initWG := sync.WaitGroup{}
	initWG.Add(1)

	w.uiThreadWG.Go(func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		w.logger.Info("window loop started")

		w.wv = C.webview_create(0, nil)
		initWG.Done()

		for {
			select {
			case f := <-w.uiThreadEvents:
				f()
			case <-w.uiThreadDone:
				w.logger.Info("exiting window loop")
				return
			}
		}
	})

	initWG.Wait()

	app.OnApplicationStart(func() {
		err := w.Open()
		if err != nil {
			w.logger.Error("failed to open window", "error", err)
		}
	})

	return w, nil
}

func (w *Window) OnClose(handler WindowOnCloseHandler) {
	w.onCloseHandlers = append(w.onCloseHandlers, handler)
}

func (w *Window) Bind(name string, f any) error {
	if w.wv == nil {
		return fmt.Errorf("window destroyed")
	}

	fVal := reflect.ValueOf(f)
	if fVal.Kind() != reflect.Func {
		return fmt.Errorf("f must be a function")
	}

	if _, exists := w.scriptBindings[name]; exists {
		return fmt.Errorf("binding with name %s already exists", name)
	}

	w.scriptBindings[name] = scriptBinding{
		f: fVal,
	}

	// w.setInitScript(w.generateInitScript())

	return nil
}

// func (w *Window) generateInitScript() string {
// 	methods := make([]string, 0, len(w.scriptBindings))
// 	for method := range w.scriptBindings {
// 		methods = append(methods, method)
// 	}
// 	methodsJson, _ := json.Marshal(methods)

// 	return fmt.Sprintf(`(function () {
// 		const methods = %s;
// 		var promises = {};
// 		var lastId = 0;
// 		for (const method of methods) {
// 			window[method] = (...params) => {
// 				id = String(lastId++);
// 				const promise = new Promise((resolve, reject) => {
// 					promises[id] = { resolve, reject };
// 				});
// 				msg = { id, method, params };
// 				%s(JSON.stringify(msg));
// 				return promise;
// 			};
// 		}
// 		window.__webview_onReply = (id, result, ok) => {
// 			if (promises[id]) {
// 				const { resolve, reject } = promises[id];
// 				if (ok) {
// 					try {
// 						result = JSON.parse(result);
// 						resolve(result);
// 					} catch (e) {
// 						reject(new Error("failed to parse JSON result: " + e));
// 					}
// 				} else {
// 					reject(new Error(result));
// 				}
// 				delete promises[id];
// 			}
// 		};
// 	})()`, string(methodsJson), w.getPostMessageFunction())
// }

// func (w *Window) onScriptMessage(msg string) {
// 	w.app.logger.Info("received script message", "msg", msg)
// 	scriptMsg := scriptMessage{}
// 	err := json.Unmarshal([]byte(msg), &scriptMsg)
// 	if err != nil {
// 		w.app.logger.Error("failed to unmarshal script message", "error", err)
// 		return
// 	}

// 	method := scriptMsg.Method
// 	id := scriptMsg.Id

// 	if id == "" {
// 		w.app.logger.Error("failed to handle message: id is empty")
// 		return
// 	}

// 	callBinding := func() (string, error) {
// 		binding, ok := w.scriptBindings[method]
// 		if !ok {
// 			return "", fmt.Errorf("method %s not found", method)
// 		}

// 		numParams := binding.f.Type().NumIn()
// 		if numParams != len(scriptMsg.Params) {
// 			return "", fmt.Errorf("invalid number of parameters: expected %d, got %d", numParams, len(scriptMsg.Params))
// 		}

// 		params := make([]reflect.Value, numParams)
// 		for i := range numParams {
// 			paramType := binding.f.Type().In(i)
// 			param := reflect.New(paramType)
// 			err := json.Unmarshal(scriptMsg.Params[i], param.Interface())
// 			if err != nil {
// 				return "", fmt.Errorf("failed to unmarshal parameter %d: %w", i, err)
// 			}
// 			params[i] = param.Elem()
// 		}

// 		results := binding.f.Call(params)
// 		errorType := reflect.TypeFor[error]()

// 		switch len(results) {
// 		case 0:
// 			return "", nil
// 		case 1:
// 			if results[0].Type().Implements(errorType) && !results[0].IsNil() {
// 				return "", results[0].Interface().(error)
// 			}
// 		case 2:
// 			if !results[1].Type().Implements(errorType) {
// 				return "", fmt.Errorf("second return value must be an error")
// 			}
// 			if !results[1].IsNil() {
// 				return "", results[1].Interface().(error)
// 			}
// 		default:
// 			return "", fmt.Errorf("invalid number of return values: expected 0, 1 or 2, got %d", len(results))
// 		}

// 		result := results[0].Interface()
// 		resultBytes, err := json.Marshal(result)
// 		if err != nil {
// 			return "", fmt.Errorf("failed to marshal result: %w", err)
// 		}

// 		return string(resultBytes), nil
// 	}

// 	result, err := callBinding()

// 	ok := err == nil
// 	if !ok {
// 		result = err.Error()
// 	}

// 	escapedResultBytes, err := json.Marshal(result)
// 	if err != nil {
// 		w.app.logger.Error("failed to escape result", "error", err)
// 		return
// 	}

// 	w.runScript(fmt.Sprintf("window.__webview_onReply(\"%s\", %s, %t);", id, string(escapedResultBytes), ok))
// }

func (w *Window) runOnUIThread(f func() error) error {
	if w.wv == nil {
		return fmt.Errorf("window destroyed")
	}

	if w.running {
		return fmt.Errorf("window already running, cannot run on UI thread")
	}

	var wg sync.WaitGroup
	var err error
	wg.Add(1)
	w.uiThreadEvents <- func() {
		defer wg.Done()
		err = f()
	}
	wg.Wait()

	return err
}

func (w *Window) Open() error {
	if w.wv == nil {
		return fmt.Errorf("window destroyed")
	}

	url := fmt.Sprintf("%s%s", w.app.frontend.GetUrl(), w.path)
	s := C.CString(url)
	defer C.free(unsafe.Pointer(s))

	w.logger.Info("opening url", "url", url)
	return w.runOnUIThread(func() error {
		rc := C.webview_navigate(w.wv, s)
		if rc != C.WEBVIEW_ERROR_OK {
			return fmt.Errorf("failed to open url")
		}
		return nil
	})
}

func (w *Window) run(ctx context.Context) error {
	if w.wv == nil {
		return fmt.Errorf("window destroyed")
	}

	w.logger.Info("running window")
	w.running = true
	closeChan := make(chan struct{})

	w.uiThreadEvents <- func() {
		C.webview_run(w.wv)
		closeChan <- struct{}{}
	}

	for w.running {
		select {
		case <-ctx.Done():
			C.webview_terminate(w.wv)
		case <-closeChan:
			w.uiThreadDone <- struct{}{}
			w.running = false
		}
	}
	w.uiThreadWG.Wait()

	w.app.logger.Info("window closing")
	for _, handler := range w.onCloseHandlers {
		handler(w.id)
	}
	w.onCloseHandlers = make([]WindowOnCloseHandler, 0)

	rt := C.webview_destroy(w.wv)

	if rt != C.WEBVIEW_ERROR_OK {
		return fmt.Errorf("failed to destroy window")
	}

	w.wv = nil

	return nil
}
