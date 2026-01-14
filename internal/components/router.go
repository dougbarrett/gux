//go:build js && wasm

package components

import "syscall/js"

type Route struct {
	Path    string
	Handler func()
}

type Router struct {
	routes       map[string]func()
	currentPath  string
	onNavigate   func(path string)
}

var globalRouter *Router

func NewRouter() *Router {
	r := &Router{
		routes: make(map[string]func()),
	}
	globalRouter = r

	// Listen for browser back/forward
	js.Global().Get("window").Call("addEventListener", "popstate", js.FuncOf(func(this js.Value, args []js.Value) any {
		path := js.Global().Get("location").Get("pathname").String()
		r.handleRoute(path)
		return nil
	}))

	return r
}

func (r *Router) Register(path string, handler func()) {
	r.routes[path] = handler
}

func (r *Router) OnNavigate(fn func(path string)) {
	r.onNavigate = fn
}

func (r *Router) Navigate(path string) {
	if path == r.currentPath {
		return
	}

	// Update browser URL
	js.Global().Get("history").Call("pushState", nil, "", path)
	r.handleRoute(path)
}

func (r *Router) handleRoute(path string) {
	r.currentPath = path

	if r.onNavigate != nil {
		r.onNavigate(path)
	}

	if handler, ok := r.routes[path]; ok {
		handler()
	} else if handler, ok := r.routes["/"]; ok {
		// Fallback to root
		handler()
	}
}

func (r *Router) Start() {
	// Handle initial route
	path := js.Global().Get("location").Get("pathname").String()
	if path == "" {
		path = "/"
	}
	r.handleRoute(path)
}

func (r *Router) CurrentPath() string {
	return r.currentPath
}

// GetRouter returns the global router instance
func GetRouter() *Router {
	return globalRouter
}
