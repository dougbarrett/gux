//go:build js && wasm

package components

import (
	"strings"
	"syscall/js"
)

// RouteHandler is called when a route is matched
type RouteHandler func()

// NavigateCallback is called after navigation completes
type NavigateCallback func(path string)

// route represents a registered route with its pattern and handler
type route struct {
	pattern  string       // Original pattern like "/users/:id"
	segments []string     // Split segments: ["users", ":id"]
	handler  RouteHandler
}

// Router handles client-side routing with browser history
type Router struct {
	routes      []route
	routeParams map[string]string
	onNavigate  NavigateCallback
	currentPath string
}

// NewRouter creates a new Router instance
func NewRouter() *Router {
	return &Router{
		routes:      []route{},
		routeParams: make(map[string]string),
	}
}

// parseSegments splits a path into segments, filtering empty strings
func parseSegments(path string) []string {
	parts := strings.Split(path, "/")
	segments := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			segments = append(segments, p)
		}
	}
	return segments
}

// matchRoute finds a matching route and extracts parameters
func (r *Router) matchRoute(path string) (*route, map[string]string) {
	pathSegments := parseSegments(path)

	for i := range r.routes {
		rt := &r.routes[i]
		if len(rt.segments) != len(pathSegments) {
			continue
		}

		params := make(map[string]string)
		match := true

		for j, seg := range rt.segments {
			if strings.HasPrefix(seg, ":") {
				// Parameter segment - extract value
				paramName := seg[1:]
				params[paramName] = pathSegments[j]
			} else if seg != pathSegments[j] {
				// Static segment must match exactly
				match = false
				break
			}
		}

		if match {
			return rt, params
		}
	}

	return nil, nil
}

// Register adds a route handler
func (r *Router) Register(pattern string, handler RouteHandler) {
	r.routes = append(r.routes, route{
		pattern:  pattern,
		segments: parseSegments(pattern),
		handler:  handler,
	})
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
	if rt, params := r.matchRoute(path); rt != nil {
		r.routeParams = params
		rt.handler()
	} else {
		r.routeParams = make(map[string]string)
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

		if rt, params := r.matchRoute(path); rt != nil {
			r.routeParams = params
			rt.handler()
		} else {
			r.routeParams = make(map[string]string)
		}

		if r.onNavigate != nil {
			r.onNavigate(path)
		}

		return nil
	}))

	// Handle initial URL
	path := js.Global().Get("location").Get("pathname").String()
	r.currentPath = path

	if rt, params := r.matchRoute(path); rt != nil {
		r.routeParams = params
		rt.handler()
	} else {
		r.routeParams = make(map[string]string)
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

// GetRouteParams returns the current route parameters
func GetRouteParams() map[string]string {
	if globalRouter == nil {
		return nil
	}
	return globalRouter.routeParams
}
