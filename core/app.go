package core

import (
	"encoding/json"
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
	Bundle  string   // Bundle name (empty = default "app" bundle)
}

// App is the main application container.
type App struct {
	routes      []Route
	apiPrefix   string
	wasmBinary  []byte            // Default bundle (backward compatibility)
	wasmBundles map[string][]byte // Named bundles: "admin" -> admin.wasm bytes
	wasmExecJS  []byte
	stylesCSS   []byte
	title       string
	db          interface{}   // Database connection (e.g., *gorm.DB)
	crudModels  []CRUDModel   // Registered CRUD models
}

// Default assets (set by generated code)
var defaultWasmBinary []byte
var defaultWasmBundles = make(map[string][]byte)
var defaultWasmExecJS []byte
var defaultStylesCSS []byte

// SetDefaultAssets sets the default WASM and CSS assets.
// Called by generated assets_gen.go in init().
func SetDefaultAssets(wasmBinary, wasmExecJS, stylesCSS []byte) {
	defaultWasmBinary = wasmBinary
	defaultWasmExecJS = wasmExecJS
	defaultStylesCSS = stylesCSS
}

// SetDefaultBundle registers a named WASM bundle.
// Called by generated assets_gen.go for each bundle.
func SetDefaultBundle(name string, wasmBinary []byte) {
	defaultWasmBundles[name] = wasmBinary
}

// New creates a new App instance.
func New() *App {
	// Copy default bundles
	bundles := make(map[string][]byte)
	for k, v := range defaultWasmBundles {
		bundles[k] = v
	}

	return &App{
		apiPrefix:   "/__gux_api",
		title:       "Gux App",
		wasmBinary:  defaultWasmBinary,
		wasmBundles: bundles,
		wasmExecJS:  defaultWasmExecJS,
		stylesCSS:   defaultStylesCSS,
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

// RouteGroupOption configures a route group.
type RouteGroupOption func(*RouteGroup)

// WithBundle assigns routes in this group to a separate WASM bundle.
// Routes in different bundles are compiled into separate WASM files,
// reducing initial download size for each section of your app.
//
// Usage:
//
//	app.RouteGroup("/admin", core.WithBundle("admin")).
//	    Hybrid("/", admin.Dashboard).
//	    Hybrid("/account", admin.Account)
func WithBundle(name string) RouteGroupOption {
	return func(rg *RouteGroup) {
		rg.bundle = name
	}
}

// RouteGroup represents a group of routes with a common prefix and options.
type RouteGroup struct {
	app    *App
	prefix string
	bundle string // WASM bundle name (empty = default)
}

// RouteGroup creates a new route group with the given prefix and options.
// Use WithBundle() to assign the group to a separate WASM bundle.
//
// Usage:
//
//	app.RouteGroup("/admin", core.WithBundle("admin")).
//	    Hybrid("/", admin.Dashboard).
//	    Hybrid("/users", admin.Users)
func (a *App) RouteGroup(prefix string, opts ...RouteGroupOption) *RouteGroup {
	rg := &RouteGroup{
		app:    a,
		prefix: prefix,
	}
	for _, opt := range opts {
		opt(rg)
	}
	return rg
}

// GET registers a GET route in the group (SSR only).
func (rg *RouteGroup) GET(path string, handler PageFunc) *RouteGroup {
	fullPath := rg.prefix + path
	if path == "/" {
		fullPath = rg.prefix
	}
	rg.app.routes = append(rg.app.routes, Route{
		Method:  "GET",
		Path:    fullPath,
		Handler: handler,
		Hybrid:  false,
		Bundle:  rg.bundle,
	})
	return rg
}

// Hybrid registers a route with SSR + WASM hydration in the group.
func (rg *RouteGroup) Hybrid(path string, handler PageFunc) *RouteGroup {
	fullPath := rg.prefix + path
	if path == "/" {
		fullPath = rg.prefix
	}
	rg.app.routes = append(rg.app.routes, Route{
		Method:  "GET",
		Path:    fullPath,
		Handler: handler,
		Hybrid:  true,
		Bundle:  rg.bundle,
	})
	return rg
}

// Run starts the HTTP server on the given address.
func (a *App) Run(addr string) error {
	mux := http.NewServeMux()

	// Register CRUD handlers
	a.registerCRUDHandlers(mux)

	// Serve default WASM binary (if assets are set)
	if len(a.wasmBinary) > 0 {
		mux.HandleFunc("/app.wasm", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/wasm")
			w.Write(a.wasmBinary)
		})
	}

	// Serve named WASM bundles (e.g., /admin.wasm)
	for name, binary := range a.wasmBundles {
		bundleBinary := binary // Capture for closure
		mux.HandleFunc("/"+name+".wasm", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/wasm")
			w.Write(bundleBinary)
		})
	}

	// Serve wasm_exec.js (if assets are set)
	if len(a.wasmExecJS) > 0 {
		mux.HandleFunc("/wasm_exec.js", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/javascript")
			w.Write(a.wasmExecJS)
		})
	}

	// Serve styles.css (if assets are set)
	if len(a.stylesCSS) > 0 {
		mux.HandleFunc("/styles.css", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/css")
			w.Write(a.stylesCSS)
		})
	}

	// Register page loader endpoints (for client-side navigation)
	for _, route := range a.routes {
		if route.Hybrid {
			handler := route.Handler
			loaderPath := a.apiPrefix + "/pages" + route.Path
			if route.Path == "/" {
				loaderPath = a.apiPrefix + "/pages/index"
			}
			mux.HandleFunc(loaderPath, func(w http.ResponseWriter, r *http.Request) {
				router := NewRouterWithDB(a.db)
				component := handler(router) // Run loader
				component()                  // Run component to populate state
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(router.state)
			})
		}
	}

	// Register page routes
	for _, route := range a.routes {
		handler := route.Handler
		hybrid := route.Hybrid
		bundle := route.Bundle
		mux.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			router := NewRouterWithDB(a.db)
			component := handler(router)
			html := component().Render(HTML()).HTML()

			w.Header().Set("Content-Type", "text/html")

			// Determine which WASM to use
			wasmPath := "/app.wasm"
			hasWasm := len(a.wasmBinary) > 0
			if bundle != "" {
				wasmPath = "/" + bundle + ".wasm"
				_, hasWasm = a.wasmBundles[bundle]
			}

			if hybrid && hasWasm {
				// Serialize state for hydration
				stateJSON, _ := json.Marshal(router.state)

				// Include WASM loader for hydration
				fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
    <link rel="stylesheet" href="/styles.css">
</head>
<body>
    <div id="app">%s</div>
    <script id="__gux_state" type="application/json">%s</script>
    <script src="/wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("%s"), go.importObject)
            .then(result => go.run(result.instance));
    </script>
</body>
</html>`, a.title, html, stateJSON, wasmPath)
			} else {
				// SSR only, no WASM
				fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
    <link rel="stylesheet" href="/styles.css">
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
	db       interface{} // Database connection (server-side only)
	hydrated bool        // True if state was hydrated from server
}

// NewRouter creates a new router instance.
func NewRouter(rerender func()) *Router {
	return &Router{
		state:    make(map[string]any),
		rerender: rerender,
	}
}

// NewRouterWithDB creates a router with database access (server-side).
func NewRouterWithDB(db interface{}) *Router {
	return &Router{
		state: make(map[string]any),
		db:    db,
	}
}

// DB returns the database connection for server-side queries.
// Returns nil in WASM - use api package for client-side data fetching.
func (r *Router) DB() interface{} {
	return r.db
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

// Hydrate restores state from SSR.
// Called by WASM runtime during hydration.
func (r *Router) Hydrate(state map[string]any) {
	for k, v := range state {
		r.state[k] = v
	}
	r.hydrated = true
}

// IsHydrated returns true if state was hydrated from server.
// Use this in loaders to skip data fetching when state is already available.
func (r *Router) IsHydrated() bool {
	return r.hydrated
}

// ClearHydrated resets the hydrated flag.
// Called after render to allow future navigations to fetch fresh data.
func (r *Router) ClearHydrated() {
	r.hydrated = false
}

// OnLoad runs the loader function only if state was NOT hydrated from server.
// Use this to wrap data loading logic - it will automatically be skipped
// when the client already has the data from SSR or client-side navigation.
//
// Usage:
//
//	r.OnLoad(func() {
//	    api.Counters.Get(1, func(c *api.Counter, err error) {
//	        if c != nil {
//	            initialCount = c.Value
//	        }
//	    })
//	})
func (r *Router) OnLoad(loader func()) {
	if !r.hydrated {
		loader()
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
		// Handle JSON unmarshaling: numbers come as float64
		if result, ok := val.(T); ok {
			return result
		}
		// Try converting float64 to int (common case from JSON)
		if f, ok := val.(float64); ok {
			var zero T
			switch any(zero).(type) {
			case int:
				return any(int(f)).(T)
			case int64:
				return any(int64(f)).(T)
			case uint:
				return any(uint(f)).(T)
			}
		}
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
