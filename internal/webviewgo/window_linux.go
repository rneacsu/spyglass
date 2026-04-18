package webviewgo

/*
#cgo LDFLAGS: -ldl
#cgo pkg-config: gtk4 webkitgtk-6.0

#include <stdlib.h>
#include <gtk/gtk.h>
#include <webkit/webkit.h>

gboolean goOnWindowClose(GtkWindow *, gpointer);
void goOnScriptMessage(WebKitUserContentManager *, JSCValue *, gpointer);

*/
import "C"
import (
	"fmt"
	"runtime"
	"runtime/cgo"
	"unsafe"
)

const WEBKIT_HANDLER_NAME = "go"

type LinuxWindow struct {
	win                      *Window
	app                      *LinuxApplication
	window                   *C.GtkWindow
	webview                  *C.WebKitWebView
	webkitSettings           *C.WebKitSettings
	webkitUserContentManager *C.WebKitUserContentManager
	initScript               *C.WebKitUserScript
	handlePinner             runtime.Pinner
}

//export goOnWindowClose
func goOnWindowClose(win *C.GtkWindow, data C.gpointer) C.gboolean {
	handle := (*cgo.Handle)(data)
	if w, ok := handle.Value().(*LinuxWindow); ok {
		handle.Delete()
		w.handlePinner.Unpin()
		go w.win.onCloseSignal()
	}
	return 0
}

//export goOnScriptMessage
func goOnScriptMessage(ucm *C.WebKitUserContentManager, msg *C.JSCValue, data C.gpointer) {
	handle := (*cgo.Handle)(data)
	if w, ok := handle.Value().(*LinuxWindow); ok {
		s := C.jsc_value_to_string(msg)
		defer C.g_free((C.gpointer)(s))
		msgStr := C.GoString(s)
		go w.win.onScriptMessage(msgStr)
	}
}

func newLinuxWindow(app *LinuxApplication, win *Window) *LinuxWindow {
	w := &LinuxWindow{
		win: win,
		app: app,
	}

	w.app.mainThreadRun(func() {
		w.window = (*C.GtkWindow)(unsafe.Pointer(C.gtk_window_new()))

		handle := cgo.NewHandle(w)
		w.handlePinner.Pin(&handle)

		s := C.CString("close-request")
		defer C.free(unsafe.Pointer(s))
		C.g_signal_connect_data((C.gpointer)(w.window), s, (C.GCallback)(C.goOnWindowClose), (C.gpointer)(&handle), nil, 0)

		w.webview = (*C.WebKitWebView)(unsafe.Pointer(C.webkit_web_view_new()))
		C.gtk_window_set_child(w.window, (*C.GtkWidget)(unsafe.Pointer(w.webview)))

		w.webkitSettings = C.webkit_web_view_get_settings(w.webview)
		C.webkit_settings_set_javascript_can_access_clipboard(w.webkitSettings, 1)
		C.webkit_settings_set_enable_developer_extras(w.webkitSettings, 1)

		w.webkitUserContentManager = C.webkit_web_view_get_user_content_manager(w.webview)
		sHandlerName := C.CString(WEBKIT_HANDLER_NAME)
		defer C.free(unsafe.Pointer(sHandlerName))
		sSignalName := C.CString(fmt.Sprintf("script-message-received::%s", WEBKIT_HANDLER_NAME))
		defer C.free(unsafe.Pointer(sSignalName))
		C.webkit_user_content_manager_register_script_message_handler(w.webkitUserContentManager, sHandlerName, nil)
		C.g_signal_connect_data((C.gpointer)(w.webkitUserContentManager), sSignalName, (C.GCallback)(C.goOnScriptMessage), (C.gpointer)(&handle), nil, 0)
	})

	return w
}

func (w *LinuxWindow) Close() {
	w.app.mainThreadRun(func() {
		C.gtk_window_close(w.window)
	})
}

func (w *LinuxWindow) SetTitle(title string) {
	s := C.CString(title)
	defer C.free(unsafe.Pointer(s))
	w.app.mainThreadRun(func() {
		C.gtk_window_set_title(w.window, s)
	})
}

func (w *LinuxWindow) SetMinSize(width int, height int) {
	w.app.mainThreadRun(func() {
		C.gtk_widget_set_size_request((*C.GtkWidget)(unsafe.Pointer(w.window)), C.int(width), C.int(height))
	})
}

func (w *LinuxWindow) open(url string) {
	s := C.CString(url)
	defer C.free(unsafe.Pointer(s))
	w.app.mainThreadRun(func() {
		C.gtk_window_present(w.window)
		C.webkit_web_view_load_uri(w.webview, s)
	})
}

func (w *LinuxWindow) getPostMessageFunction() string {
	return fmt.Sprintf("window.webkit.messageHandlers.%s.postMessage", WEBKIT_HANDLER_NAME)
}

func (w *LinuxWindow) setInitScript(script string) {
	s := C.CString(script)
	defer C.free(unsafe.Pointer(s))

	w.app.mainThreadRun(func() {
		if w.initScript != nil {
			C.webkit_user_content_manager_remove_script(w.webkitUserContentManager, w.initScript)
			C.webkit_user_script_unref(w.initScript)
		}

		w.initScript = C.webkit_user_script_new(s, C.WEBKIT_USER_CONTENT_INJECT_TOP_FRAME, C.WEBKIT_USER_SCRIPT_INJECT_AT_DOCUMENT_START, nil, nil)
		C.webkit_user_content_manager_add_script(w.webkitUserContentManager, w.initScript)
	})
}

func (w *LinuxWindow) runScript(script string) {
	s := C.CString(script)
	defer C.free(unsafe.Pointer(s))

	w.app.mainThreadRun(func() {
		if C.webkit_web_view_get_uri(w.webview) == nil {
			return
		}
		C.webkit_web_view_evaluate_javascript(w.webview, s, -1, nil, nil, nil, nil, nil)
	})
}
