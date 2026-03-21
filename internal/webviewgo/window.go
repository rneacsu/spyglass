package webviewgo

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type WindowOnCloseHandler func(id string)

type WindowImpl interface {
	SetTitle(title string)
	SetMinSize(width int, height int)
	// SetSize(width int, height int)
	open(url string)
	Close()
	setInitScript(script string)
	getPostMessageFunction() string
	runScript(script string)
}

type scriptMessage struct {
	Id     string            `json:"id"`
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
}

type scriptBinding struct {
	f reflect.Value
}

type Window struct {
	WindowImpl
	app             *Application
	id              string
	path            string
	onCloseHandlers []WindowOnCloseHandler
	scriptBindings  map[string]scriptBinding
}

func newWindow(app *Application, id string, path string) *Window {
	if path[0] != '/' {
		path = "/" + path
	}

	w := &Window{
		app:             app,
		id:              id,
		path:            path,
		onCloseHandlers: make([]WindowOnCloseHandler, 0),
		scriptBindings:  make(map[string]scriptBinding),
	}

	w.WindowImpl = app.newWindow(w)
	app.OnApplicationStart(func() {
		w.Open()
	})

	return w
}

func (w *Window) OnClose(handler WindowOnCloseHandler) {
	w.onCloseHandlers = append(w.onCloseHandlers, handler)
}

func (w *Window) onCloseSignal() {
	w.app.logger.Info("window closing")
	for _, handler := range w.onCloseHandlers {
		handler(w.id)
	}
	w.onCloseHandlers = make([]WindowOnCloseHandler, 0)
}

func (w *Window) Bind(name string, f any) error {
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

	w.setInitScript(w.generateInitScript())

	return nil
}

func (w *Window) generateInitScript() string {
	methods := make([]string, 0, len(w.scriptBindings))
	for method := range w.scriptBindings {
		methods = append(methods, method)
	}
	methodsJson, _ := json.Marshal(methods)

	return fmt.Sprintf(`(function () {
		const methods = %s;
		var promises = {};
		var lastId = 0;
		for (const method of methods) {
			window[method] = (...params) => {
				id = String(lastId++);
				const promise = new Promise((resolve, reject) => {
					promises[id] = { resolve, reject };
				});
				msg = { id, method, params };
				%s(JSON.stringify(msg));
				return promise;
			};
		}
		window.__webview_onReply = (id, result, ok) => {
			if (promises[id]) {
				const { resolve, reject } = promises[id];
				if (ok) {
					try {
						result = JSON.parse(result);
						resolve(result);
					} catch (e) {
						reject(new Error("failed to parse JSON result: " + e));
					}
				} else {
					reject(new Error(result));
				}
				delete promises[id];
			}
		};
	})()`, string(methodsJson), w.getPostMessageFunction())
}

func (w *Window) onScriptMessage(msg string) {
	w.app.logger.Info("received script message", "msg", msg)
	scriptMsg := scriptMessage{}
	err := json.Unmarshal([]byte(msg), &scriptMsg)
	if err != nil {
		w.app.logger.Error("failed to unmarshal script message", "error", err)
		return
	}

	method := scriptMsg.Method
	id := scriptMsg.Id

	if id == "" {
		w.app.logger.Error("failed to handle message: id is empty")
		return
	}

	callBinding := func() (string, error) {
		binding, ok := w.scriptBindings[method]
		if !ok {
			return "", fmt.Errorf("method %s not found", method)
		}

		numParams := binding.f.Type().NumIn()
		if numParams != len(scriptMsg.Params) {
			return "", fmt.Errorf("invalid number of parameters: expected %d, got %d", numParams, len(scriptMsg.Params))
		}

		params := make([]reflect.Value, numParams)
		for i := 0; i < numParams; i++ {
			paramType := binding.f.Type().In(i)
			param := reflect.New(paramType)
			err := json.Unmarshal(scriptMsg.Params[i], param.Interface())
			if err != nil {
				return "", fmt.Errorf("failed to unmarshal parameter %d: %w", i, err)
			}
			params[i] = param.Elem()
		}

		results := binding.f.Call(params)
		errorType := reflect.TypeFor[error]()

		switch len(results) {
		case 0:
			return "", nil
		case 1:
			if results[0].Type().Implements(errorType) && !results[0].IsNil() {
				return "", results[0].Interface().(error)
			}
		case 2:
			if !results[1].Type().Implements(errorType) {
				return "", fmt.Errorf("second return value must be an error")
			}
			if !results[1].IsNil() {
				return "", results[1].Interface().(error)
			}
		default:
			return "", fmt.Errorf("invalid number of return values: expected 0, 1 or 2, got %d", len(results))
		}

		result := results[0].Interface()
		resultBytes, err := json.Marshal(result)
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %w", err)
		}

		return string(resultBytes), nil
	}

	result, err := callBinding()

	ok := err == nil
	if !ok {
		result = err.Error()
	}

	escapedResultBytes, err := json.Marshal(result)
	if err != nil {
		w.app.logger.Error("failed to escape result", "error", err)
		return
	}

	w.runScript(fmt.Sprintf("window.__webview_onReply(\"%s\", %s, %t);", id, string(escapedResultBytes), ok))
}

func (w *Window) Open() {
	url := fmt.Sprintf("%s%s", w.app.frontend.GetUrl(), w.path)
	w.open(url)
}
