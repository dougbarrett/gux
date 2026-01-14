//go:build js && wasm

package components

import "syscall/js"

// RouteHandler is called when a route is matched
type RouteHandler func()

// NavigateCallback is called after navigation completes
type NavigateCallback func(path string)

// Router handles client-side routing with browser history
type Router struct {
	routes      map[string]RouteHandler
	onNavigate  NavigateCallback
	currentPath string
}

// NewRouter creates a new Router instance
func NewRouter() *Router {
	return &Router{
		routes: make(map[string]RouteHandler),
	}
}

// Register adds a route handler
func (r *Router) Register(path string, handler RouteHandler) {
	r.routes[path] = handler
}

// OnNavigate sets a callback for navigation events
func (r *Router) OnNavigate(cb NavigateCallback) {
	r.onNavigate = cb
}

// Navigate programmatically navigates to a path
func (r *Router) Navigate(path string) {
	if path == r.currentPath {
		return
	}

	r.currentPath = path

	// Update browser URL
	js.Global().Get("history").Call("pushState", nil, "", path)

	// Call route handler
	if handler, ok := r.routes[path]; ok {
		handler()
	}

	// Notify listeners
	if r.onNavigate != nil {
		r.onNavigate(path)
	}
}

// Start initializes the router and handles the current URL
func (r *Router) Start() {
	// Handle browser back/forward
	js.Global().Call("addEventListener", "popstate", js.FuncOf(func(this js.Value, args []js.Value) any {
		path := js.Global().Get("location").Get("pathname").String()
		r.currentPath = path

		if handler, ok := r.routes[path]; ok {
			handler()
		}

		if r.onNavigate != nil {
			r.onNavigate(path)
		}

		return nil
	}))

	// Handle initial URL
	path := js.Global().Get("location").Get("pathname").String()
	r.currentPath = path

	if handler, ok := r.routes[path]; ok {
		handler()
	}

	if r.onNavigate != nil {
		r.onNavigate(path)
	}
}

// CurrentPath returns the current route path
func (r *Router) CurrentPath() string {
	return r.currentPath
}

// global router instance for Link component
var globalRouter *Router

// SetGlobalRouter sets the router instance used by Link components
func SetGlobalRouter(r *Router) {
	globalRouter = r
}

// GetGlobalRouter returns the global router instance
func GetGlobalRouter() *Router {
	return globalRouter
}
