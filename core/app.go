package core

import (
	"fmt"
	"net/http"
)

// PageFunc is a page handler that returns a component function.
// The outer function runs on server (loader), the inner runs on client (component).
type PageFunc func(r *Router) func() Node

// Route represents a registered route.
type Route struct {
	Method  string   // GET, POST, etc.
	Path    string   // /posts, /posts/:id
	Handler PageFunc // The page function
	Hybrid  bool     // If true, SSR + WASM hydration
}

// App is the main application container.
type App struct {
	routes     []Route
	apiPrefix  string
	wasmBinary []byte
	wasmExecJS []byte
	title      string
}

// Default assets (set by generated code)
var defaultWasmBinary []byte
var defaultWasmExecJS []byte

// SetDefaultAssets sets the default WASM assets.
// Called by generated assets_gen.go in init().
func SetDefaultAssets(wasmBinary, wasmExecJS []byte) {
	defaultWasmBinary = wasmBinary
	defaultWasmExecJS = wasmExecJS
}

// New creates a new App instance.
func New() *App {
	return &App{
		apiPrefix:  "/__gux_api",
		title:      "Gux App",
		wasmBinary: defaultWasmBinary,
		wasmExecJS: defaultWasmExecJS,
	}
}

// SetAssets sets the WASM binary and wasm_exec.js.
// This is called by generated code, not by the developer.
func (a *App) SetAssets(wasmBinary, wasmExecJS []byte) {
	a.wasmBinary = wasmBinary
	a.wasmExecJS = wasmExecJS
}

// SetTitle sets the page title.
func (a *App) SetTitle(title string) {
	a.title = title
}

// RouteBuilder provides methods for registering routes.
type RouteBuilder struct {
	app *App
}

// Routes returns a RouteBuilder for registering routes.
func (a *App) Routes() *RouteBuilder {
	return &RouteBuilder{app: a}
}

// GET registers a GET route (SSR only).
func (rb *RouteBuilder) GET(path string, handler PageFunc) *RouteBuilder {
	rb.app.routes = append(rb.app.routes, Route{
		Method:  "GET",
		Path:    path,
		Handler: handler,
		Hybrid:  false,
	})
	return rb
}

// POST registers a POST route.
func (rb *RouteBuilder) POST(path string, handler PageFunc) *RouteBuilder {
	rb.app.routes = append(rb.app.routes, Route{
		Method:  "POST",
		Path:    path,
		Handler: handler,
		Hybrid:  false,
	})
	return rb
}

// Hybrid registers a route with SSR + WASM hydration.
func (rb *RouteBuilder) Hybrid(path string, handler PageFunc) *RouteBuilder {
	rb.app.routes = append(rb.app.routes, Route{
		Method:  "GET",
		Path:    path,
		Handler: handler,
		Hybrid:  true,
	})
	return rb
}

// Run starts the HTTP server on the given address.
func (a *App) Run(addr string) error {
	mux := http.NewServeMux()

	// Serve WASM binary (if assets are set)
	if len(a.wasmBinary) > 0 {
		mux.HandleFunc("/app.wasm", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/wasm")
			w.Write(a.wasmBinary)
		})
	}

	// Serve wasm_exec.js (if assets are set)
	if len(a.wasmExecJS) > 0 {
		mux.HandleFunc("/wasm_exec.js", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/javascript")
			w.Write(a.wasmExecJS)
		})
	}

	// Register page routes
	for _, route := range a.routes {
		handler := route.Handler
		hybrid := route.Hybrid
		mux.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			router := NewRouter(nil)
			component := handler(router)
			html := component().Render(HTML()).HTML()

			w.Header().Set("Content-Type", "text/html")

			if hybrid && len(a.wasmBinary) > 0 {
				// Include WASM loader for hydration
				fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body>
    <div id="app">%s</div>
    <script src="/wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("/app.wasm"), go.importObject)
            .then(result => go.run(result.instance));
    </script>
</body>
</html>`, a.title, html)
			} else {
				// SSR only, no WASM
				fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body>
    <div id="app">%s</div>
</body>
</html>`, a.title, html)
			}
		})
	}

	fmt.Printf("http://localhost%s\n", addr)
	return http.ListenAndServe(addr, mux)
}

// Router provides page context and state management.
type Router struct {
	state    map[string]any
	rerender func()
	navigate func(path string)
}

// NewRouter creates a new router instance.
func NewRouter(rerender func()) *Router {
	return &Router{
		state:    make(map[string]any),
		rerender: rerender,
	}
}

// SetNavigate sets the navigation callback (called by WASM runtime).
func (r *Router) SetNavigate(fn func(path string)) {
	r.navigate = fn
}

// Navigate programmatically navigates to a path.
// Only works in WASM - on server this is a no-op.
func (r *Router) Navigate(path string) {
	if r.navigate != nil {
		r.navigate(path)
	}
}

// State represents a reactive state value.
type State[T any] struct {
	key    string
	router *Router
}

// Get returns the current state value.
func (s *State[T]) Get() T {
	if val, ok := s.router.state[s.key]; ok {
		return val.(T)
	}
	var zero T
	return zero
}

// Set updates the state and triggers re-render.
func (s *State[T]) Set(val T) {
	s.router.state[s.key] = val
	if s.router.rerender != nil {
		s.router.rerender()
	}
}

// UseState creates a reactive state for any type.
// Usage: count := core.UseState(r, "count", 0)
//        user := core.UseState(r, "user", User{Name: "John"})
func UseState[T any](r *Router, key string, initial T) *State[T] {
	if _, ok := r.state[key]; !ok {
		r.state[key] = initial
	}
	return &State[T]{key: key, router: r}
}

// StateInt creates an integer state.
func (r *Router) StateInt(key string, initial int) *State[int] {
	return UseState(r, key, initial)
}

// StateBool creates a boolean state.
func (r *Router) StateBool(key string, initial bool) *State[bool] {
	return UseState(r, key, initial)
}

// StateString creates a string state.
func (r *Router) StateString(key string, initial string) *State[string] {
	return UseState(r, key, initial)
}
