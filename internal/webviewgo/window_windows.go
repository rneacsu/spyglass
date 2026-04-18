package webviewgo

/*
#cgo LDFLAGS: -ldl

*/
import "C"

type WindowsWindow struct {
	win *Window
	app *WindowsApplication
}

func newWindowsWindow(app *WindowsApplication, win *Window) *WindowsWindow {
	w := &WindowsWindow{
		win: win,
		app: app,
	}

	return w
}

func (w *WindowsWindow) Close() {

}

func (w *WindowsWindow) SetTitle(title string) {

}

func (w *WindowsWindow) SetMinSize(width int, height int) {

}

func (w *WindowsWindow) open(url string) {

}

func (w *WindowsWindow) getPostMessageFunction() string {
	return ""
}

func (w *WindowsWindow) setInitScript(script string) {

}

func (w *WindowsWindow) runScript(script string) {

}
