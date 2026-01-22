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
	Routes      []Route
	APIPrefix   string
	wasmBinary  []byte
	wasmExecJS  []byte
	title       string
}

// New creates a new App instance.
func New() *App {
	return &App{
		APIPrefix: "/__gux_api",
		title:     "Gux App",
	}
}

// SetAssets sets the WASM binary and wasm_exec.js.
func (a *App) SetAssets(wasmBinary, wasmExecJS []byte) *App {
	a.wasmBinary = wasmBinary
	a.wasmExecJS = wasmExecJS
	return a
}

// SetTitle sets the page title.
func (a *App) SetTitle(title string) *App {
	a.title = title
	return a
}

// GET registers a GET route (SSR only by default).
func (a *App) GET(path string, handler PageFunc) *App {
	a.Routes = append(a.Routes, Route{
		Method:  "GET",
		Path:    path,
		Handler: handler,
		Hybrid:  false,
	})
	return a
}

// POST registers a POST route.
func (a *App) POST(path string, handler PageFunc) *App {
	a.Routes = append(a.Routes, Route{
		Method:  "POST",
		Path:    path,
		Handler: handler,
		Hybrid:  false,
	})
	return a
}

// Hybrid registers a route with SSR + WASM hydration.
func (a *App) Hybrid(path string, handler PageFunc) *App {
	a.Routes = append(a.Routes, Route{
		Method:  "GET",
		Path:    path,
		Handler: handler,
		Hybrid:  true,
	})
	return a
}

// Run starts the HTTP server on the given address.
func (a *App) Run(addr string) error {
	mux := http.NewServeMux()

	// Serve WASM binary
	mux.HandleFunc("/app.wasm", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/wasm")
		w.Write(a.wasmBinary)
	})

	// Serve wasm_exec.js
	mux.HandleFunc("/wasm_exec.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(a.wasmExecJS)
	})

	// Register page routes
	for _, route := range a.Routes {
		handler := route.Handler
		hybrid := route.Hybrid
		mux.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			router := NewRouter(nil)
			component := handler(router)
			html := component().Render(HTML()).HTML()

			w.Header().Set("Content-Type", "text/html")

			if hybrid {
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
}

// NewRouter creates a new router instance.
func NewRouter(rerender func()) *Router {
	return &Router{
		state:    make(map[string]any),
		rerender: rerender,
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

// StateInt creates an integer state.
func (r *Router) StateInt(key string, initial int) *State[int] {
	if _, ok := r.state[key]; !ok {
		r.state[key] = initial
	}
	return &State[int]{key: key, router: r}
}

// StateBool creates a boolean state.
func (r *Router) StateBool(key string, initial bool) *State[bool] {
	if _, ok := r.state[key]; !ok {
		r.state[key] = initial
	}
	return &State[bool]{key: key, router: r}
}

// StateString creates a string state.
func (r *Router) StateString(key string, initial string) *State[string] {
	if _, ok := r.state[key]; !ok {
		r.state[key] = initial
	}
	return &State[string]{key: key, router: r}
}
