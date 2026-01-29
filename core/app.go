package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// PageFunc is a page handler that returns a component function.
// The outer function runs on server (loader), the inner runs on client (component).
type PageFunc func(r *Router) func() Node

// Route represents a registered route.
type Route struct {
	Method        string   // GET, POST, etc.
	Path          string   // /posts, /posts/:id
	Handler       PageFunc // The page function
	Hybrid        bool     // If true, SSR + WASM hydration
	Bundle        string   // Bundle name (empty = default "app" bundle)
	Protected     bool     // If true, requires authentication
	RequiredRoles []string // Roles required to access (any of these roles)
}

// App is the main application container.
type App struct {
	routes         []Route
	apiPrefix      string
	wasmBinary     []byte                       // Default bundle (backward compatibility)
	wasmBundles    map[string][]byte            // Named bundles: "admin" -> admin.wasm bytes
	wasmExecJS     []byte
	stylesCSS      []byte
	title          string
	db             interface{}                  // Database connection (e.g., *gorm.DB)
	crudModels     []CRUDModel                  // Registered CRUD models
	csrfConfig     CSRFConfig                   // CSRF protection configuration
	darkMode       bool                         // Enable dark mode (adds class="dark" to html element)
	authConfig     *AuthConfig                  // Authentication configuration (nil = disabled)
	customHandlers map[string]http.HandlerFunc // Custom HTTP handlers
	apiEndpoints   []APIEndpoint                // Registered typed API endpoints
	auditMigrated  bool                         // Whether audit_entries table has been auto-migrated
	// Cache busting hashes
	stylesHash     string            // Hash for styles.css
	wasmHashes     map[string]string // Hash for each WASM bundle
}

// Default assets (set by generated code)
var defaultWasmBinary []byte
var defaultWasmBundles = make(map[string][]byte)
var defaultWasmExecJS []byte
var defaultStylesCSS []byte
var defaultStylesHash string
var defaultWasmHashes = make(map[string]string)

// matchRoute checks if a URL path matches a route pattern and extracts parameters.
// Pattern "/users/:id" matches path "/users/123" and returns {"id": "123"}.
func matchRoute(pattern, path string) (bool, map[string]string) {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return false, nil
	}

	params := make(map[string]string)
	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			// Parameter segment
			params[part[1:]] = pathParts[i]
		} else if part != pathParts[i] {
			// Static segment doesn't match
			return false, nil
		}
	}

	return true, params
}


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

// SetDefaultAssetHashes sets the cache-busting hashes for assets.
// Called by generated assets_gen.go in init().
func SetDefaultAssetHashes(stylesHash string, wasmHashes map[string]string) {
	defaultStylesHash = stylesHash
	for k, v := range wasmHashes {
		defaultWasmHashes[k] = v
	}
}

// New creates a new App instance.
// Automatically loads .env file if present.
func New() *App {
	// Load .env file if present (before any config)
	LoadEnv(".env")

	// Copy default bundles
	bundles := make(map[string][]byte)
	for k, v := range defaultWasmBundles {
		bundles[k] = v
	}

	// Copy default hashes
	hashes := make(map[string]string)
	for k, v := range defaultWasmHashes {
		hashes[k] = v
	}

	return &App{
		apiPrefix:   "/__gux_api",
		title:       "Gux App",
		wasmBinary:  defaultWasmBinary,
		wasmBundles: bundles,
		wasmExecJS:  defaultWasmExecJS,
		stylesCSS:   defaultStylesCSS,
		stylesHash:  defaultStylesHash,
		wasmHashes:  hashes,
		csrfConfig:  DefaultCSRFConfig(), // CSRF enabled by default
	}
}

// EnableCSRF configures CSRF protection.
// CSRF is enabled by default. Use this to customize or disable.
//
// Usage:
//
//	app.EnableCSRF(core.CSRFConfig{Enabled: false}) // Disable
//	app.EnableCSRF(core.CSRFConfig{Secure: true})   // Require HTTPS
func (a *App) EnableCSRF(config CSRFConfig) *App {
	a.csrfConfig = config
	return a
}

// DisableCSRF disables CSRF protection.
// Only use this for APIs that use other authentication methods (e.g., API keys).
func (a *App) DisableCSRF() *App {
	a.csrfConfig.Enabled = false
	return a
}

// EnableAuth configures session-based authentication.
// Requires a SessionStore implementation for session storage.
//
// Usage:
//
//	app.EnableAuth(core.AuthConfig{
//	    SessionStore: core.NewMemorySessionStore(), // Use Redis in production
//	    LoginPath:    "/login",
//	})
func (a *App) EnableAuth(config AuthConfig) *App {
	// Apply defaults for unset values
	if config.CookieName == "" {
		config.CookieName = DefaultSessionCookieName
	}
	if config.CookieMaxAge == 0 {
		config.CookieMaxAge = DefaultSessionMaxAge
	}
	if config.CookiePath == "" {
		config.CookiePath = "/"
	}
	if config.CookieSameSite == 0 {
		config.CookieSameSite = http.SameSiteLaxMode
	}
	if config.LoginPath == "" {
		config.LoginPath = "/login"
	}
	// CookieHTTPOnly defaults to true if not explicitly set
	// This is handled by checking the config when it matters

	a.authConfig = &config
	return a
}

// Auth returns the auth configuration, or nil if auth is disabled.
func (a *App) Auth() *AuthConfig {
	return a.authConfig
}

// customHandler represents a custom HTTP handler registration.
type customHandler struct {
	pattern string
	handler http.HandlerFunc
}

// HandleFunc registers a custom HTTP handler.
// Use this for API endpoints like login, logout, etc.
//
// Usage:
//
//	app.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
//	    // Handle login
//	})
func (a *App) HandleFunc(pattern string, handler http.HandlerFunc) *App {
	if a.customHandlers == nil {
		a.customHandlers = make(map[string]http.HandlerFunc)
	}
	a.customHandlers[pattern] = handler
	return a
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

// DarkMode enables dark mode for the application.
// This adds class="dark" to the html element, activating Tailwind dark mode variants.
func (a *App) DarkMode() *App {
	a.darkMode = true
	return a
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
func (rb *RouteBuilder) Hybrid(path string, handler PageFunc) *RouteEntry {
	route := Route{
		Method:  "GET",
		Path:    path,
		Handler: handler,
		Hybrid:  true,
	}
	rb.app.routes = append(rb.app.routes, route)
	return &RouteEntry{
		app:   rb.app,
		index: len(rb.app.routes) - 1,
	}
}

// RouteEntry represents a registered route that can be further configured.
type RouteEntry struct {
	app   *App
	index int
}

// Protected marks this route as requiring authentication.
// Unauthenticated users will be redirected to the login page.
func (re *RouteEntry) Protected() *RouteEntry {
	re.app.routes[re.index].Protected = true
	return re
}

// WithRoles specifies required roles for this route.
// User must have ANY of the specified roles (OR logic).
// Automatically marks the route as protected.
func (re *RouteEntry) WithRoles(roles ...string) *RouteEntry {
	re.app.routes[re.index].Protected = true
	re.app.routes[re.index].RequiredRoles = roles
	return re
}

// Hybrid allows chaining back to RouteBuilder for additional routes.
func (re *RouteEntry) Hybrid(path string, handler PageFunc) *RouteEntry {
	return (&RouteBuilder{app: re.app}).Hybrid(path, handler)
}

// GET allows chaining back to RouteBuilder for additional routes.
func (re *RouteEntry) GET(path string, handler PageFunc) *RouteEntry {
	route := Route{
		Method:  "GET",
		Path:    path,
		Handler: handler,
		Hybrid:  false,
	}
	re.app.routes = append(re.app.routes, route)
	return &RouteEntry{
		app:   re.app,
		index: len(re.app.routes) - 1,
	}
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
	app           *App
	prefix        string
	bundle        string   // WASM bundle name (empty = default)
	protected     bool     // If true, all routes in group require auth
	requiredRoles []string // Roles required for all routes in group
}

// RouteGroup creates a new route group with the given prefix and options.
// Use WithBundle() to assign the group to a separate WASM bundle.
//
// Usage:
//
//	app.RouteGroup("/admin", core.WithBundle("admin")).
//	    Protected().
//	    WithRoles("admin").
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

// Protected marks all routes in this group as requiring authentication.
func (rg *RouteGroup) Protected() *RouteGroup {
	rg.protected = true
	return rg
}

// WithRoles specifies required roles for all routes in this group.
// User must have ANY of the specified roles (OR logic).
// Automatically marks the group as protected.
func (rg *RouteGroup) WithRoles(roles ...string) *RouteGroup {
	rg.protected = true
	rg.requiredRoles = roles
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
		Method:        "GET",
		Path:          fullPath,
		Handler:       handler,
		Hybrid:        true,
		Bundle:        rg.bundle,
		Protected:     rg.protected,
		RequiredRoles: rg.requiredRoles,
	})
	return rg
}

// buildTime is set when the server starts, used for hot reload detection
var buildTime = fmt.Sprintf("%d", os.Getpid())

// Run starts the HTTP server on the given address.
func (a *App) Run(addr string) error {
	mux := http.NewServeMux()

	// Register CRUD handlers
	a.registerCRUDHandlers(mux)

	// Register custom handlers
	for pattern, handler := range a.customHandlers {
		mux.HandleFunc(pattern, handler)
	}

	// Hot reload ping endpoint (only active when GUX_HOT_RELOAD=1)
	if os.Getenv("GUX_HOT_RELOAD") == "1" {
		mux.HandleFunc("/__gux_dev/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"buildTime":"%s"}`, buildTime)
		})
	}

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
	// We need custom matching for parameterized routes
	mux.HandleFunc(a.apiPrefix+"/pages/", func(w http.ResponseWriter, r *http.Request) {
		// Extract the page path from the URL
		pagePath := strings.TrimPrefix(r.URL.Path, a.apiPrefix+"/pages")
		if pagePath == "/index" {
			pagePath = "/"
		}

		// Find matching route
		for _, route := range a.routes {
			if !route.Hybrid {
				continue
			}
			if matches, params := matchRoute(route.Path, pagePath); matches {
				// Load session user if auth is enabled
				var user *SessionUser
				if a.authConfig != nil && a.authConfig.SessionStore != nil {
					sessionID := getSessionIDFromCookie(r, a.authConfig.CookieName)
					if sessionID != "" {
						user, _ = a.authConfig.SessionStore.Get(sessionID)
					}
				}

				// Check authentication for protected routes
				if route.Protected {
					if user == nil {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusUnauthorized)
						json.NewEncoder(w).Encode(map[string]string{
							"error":    "unauthorized",
							"redirect": a.authConfig.LoginPath,
						})
						return
					}

					// Check required roles
					if len(route.RequiredRoles) > 0 {
						hasRole := false
						for _, role := range route.RequiredRoles {
							if user.HasRole(role) {
								hasRole = true
								break
							}
						}
						if !hasRole {
							w.Header().Set("Content-Type", "application/json")
							w.WriteHeader(http.StatusForbidden)
							json.NewEncoder(w).Encode(map[string]string{"error": "forbidden"})
							return
						}
					}
				}

				// Set session context for SSR API calls
				sessionID := ""
				cookieName := DefaultSessionCookieName
				if a.authConfig != nil {
					sessionID = getSessionIDFromCookie(r, a.authConfig.CookieName)
					if a.authConfig.CookieName != "" {
						cookieName = a.authConfig.CookieName
					}
				}
				if SSRSessionSetter != nil && sessionID != "" {
					SSRSessionSetter(sessionID, cookieName)
				}

				router := NewRouterWithAuth(a.db, params, user, r, w, a.authConfig)
				router.sessionID = sessionID
				component := route.Handler(router)
				component()

				// Clear session context after rendering
				if SSRSessionClearer != nil {
					SSRSessionClearer()
				}

				// Include user in state for hydration
				if user != nil {
					router.state["__gux_user"] = user
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(router.state)
				return
			}
		}
		http.NotFound(w, r)
	})

	// Register page routes with custom matching for parameterized routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Skip API and static asset routes
		if strings.HasPrefix(path, a.apiPrefix) ||
			path == "/app.wasm" ||
			strings.HasSuffix(path, ".wasm") ||
			path == "/wasm_exec.js" ||
			path == "/styles.css" {
			http.NotFound(w, r)
			return
		}

		// Find matching route
		for _, route := range a.routes {
			if matches, params := matchRoute(route.Path, path); matches {
				// Load session user if auth is enabled
				var user *SessionUser
				if a.authConfig != nil && a.authConfig.SessionStore != nil {
					sessionID := getSessionIDFromCookie(r, a.authConfig.CookieName)
					if sessionID != "" {
						user, _ = a.authConfig.SessionStore.Get(sessionID)
					}
				}

				// Check authentication for protected routes
				if route.Protected {
					if user == nil {
						// Not authenticated - redirect to login
						if a.authConfig != nil && a.authConfig.UnauthorizedHandler != nil {
							a.authConfig.UnauthorizedHandler(w, r)
						} else {
							loginPath := "/login"
							if a.authConfig != nil && a.authConfig.LoginPath != "" {
								loginPath = a.authConfig.LoginPath
							}
							http.Redirect(w, r, loginPath+"?redirect="+path, http.StatusFound)
						}
						return
					}

					// Check required roles (if any)
					if len(route.RequiredRoles) > 0 {
						hasRole := false
						for _, role := range route.RequiredRoles {
							if user.HasRole(role) {
								hasRole = true
								break
							}
						}
						if !hasRole {
							// User doesn't have required role - forbidden
							if a.authConfig != nil && a.authConfig.ForbiddenHandler != nil {
								a.authConfig.ForbiddenHandler(w, r)
							} else {
								http.Error(w, "Forbidden", http.StatusForbidden)
							}
							return
						}
					}
				}

				// Set session context for SSR API calls
				sessionID := ""
				cookieName := DefaultSessionCookieName
				if a.authConfig != nil {
					sessionID = getSessionIDFromCookie(r, a.authConfig.CookieName)
					if a.authConfig.CookieName != "" {
						cookieName = a.authConfig.CookieName
					}
				}
				if SSRSessionSetter != nil && sessionID != "" {
					SSRSessionSetter(sessionID, cookieName)
				}

				// Create router with auth context
				router := NewRouterWithAuth(a.db, params, user, r, w, a.authConfig)
				router.sessionID = sessionID
				component := route.Handler(router)
				html := component().Render(HTML()).HTML()

				// Clear session context after rendering
				if SSRSessionClearer != nil {
					SSRSessionClearer()
				}

				w.Header().Set("Content-Type", "text/html")

				// Determine which WASM to use
				wasmPath := "/app.wasm"
				hasWasm := len(a.wasmBinary) > 0
				if route.Bundle != "" {
					wasmPath = "/" + route.Bundle + ".wasm"
					_, hasWasm = a.wasmBundles[route.Bundle]
				}

				// Check if hot reload is enabled (gux dev --watch)
				hotReloadScript := ""
				if os.Getenv("GUX_HOT_RELOAD") == "1" {
					hotReloadScript = `
    <script>
        (function() {
            let lastCheck = Date.now();
            let checking = false;
            async function checkServer() {
                if (checking) return;
                checking = true;
                try {
                    const res = await fetch('/__gux_dev/ping', {cache: 'no-store'});
                    if (res.ok) {
                        const data = await res.json();
                        if (data.buildTime && data.buildTime !== window.__guxBuildTime) {
                            if (window.__guxBuildTime) {
                                console.log('[gux] Reloading...');
                                location.reload();
                            }
                            window.__guxBuildTime = data.buildTime;
                        }
                    }
                } catch (e) {
                    // Server down, will retry
                }
                checking = false;
            }
            setInterval(checkServer, 1000);
            checkServer();
        })();
    </script>`
				}

				// Generate CSRF token and set cookie
				csrfToken := ""
				if a.csrfConfig.Enabled {
					var err error
					csrfToken, err = generateCSRFToken()
					if err == nil {
						setCSRFCookie(w, csrfToken, a.csrfConfig)
					}
				}

				// Build CSRF meta tag
				csrfMeta := ""
				if csrfToken != "" {
					csrfMeta = fmt.Sprintf(`<meta name="%s" content="%s">`, CSRFMetaName, csrfToken)
				}

				// Build dark mode class
				htmlClass := ""
				if a.darkMode {
					htmlClass = ` class="dark"`
				}

				// Build cache-busted asset URLs
				stylesURL := "/styles.css"
				if a.stylesHash != "" {
					stylesURL = fmt.Sprintf("/styles.css?v=%s", a.stylesHash)
				}

				wasmURL := wasmPath
				if bundleName := route.Bundle; bundleName != "" {
					if hash, ok := a.wasmHashes[bundleName]; ok && hash != "" {
						wasmURL = fmt.Sprintf("/%s.wasm?v=%s", bundleName, hash)
					}
				} else if hash, ok := a.wasmHashes["app"]; ok && hash != "" {
					wasmURL = fmt.Sprintf("/app.wasm?v=%s", hash)
				}

				if route.Hybrid && hasWasm {
					// Include user in state for hydration
					if user != nil {
						router.state["__gux_user"] = user
					}
					// Serialize state for hydration
					stateJSON, _ := json.Marshal(router.state)

					// Include WASM loader for hydration
					fmt.Fprintf(w, `<!DOCTYPE html>
<html%s>
<head>
    <title>%s</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    %s
    <link rel="stylesheet" href="%s">
</head>
<body>
    <div id="app">%s</div>
    <script id="__gux_state" type="application/json">%s</script>
    <script src="/wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("%s"), go.importObject)
            .then(result => go.run(result.instance));
    </script>%s
</body>
</html>`, htmlClass, a.title, csrfMeta, stylesURL, html, stateJSON, wasmURL, hotReloadScript)
				} else {
					// SSR only, no WASM
					fmt.Fprintf(w, `<!DOCTYPE html>
<html%s>
<head>
    <title>%s</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    %s
    <link rel="stylesheet" href="%s">
</head>
<body>
    <div id="app">%s</div>%s
</body>
</html>`, htmlClass, a.title, csrfMeta, stylesURL, html, hotReloadScript)
				}
				return
			}
		}
		http.NotFound(w, r)
	})

	fmt.Printf("http://localhost%s\n", addr)

	// Apply CSRF middleware if enabled
	var handler http.Handler = mux
	if a.csrfConfig.Enabled {
		handler = CSRFMiddleware(a.csrfConfig)(mux)
	}

	return http.ListenAndServe(addr, handler)
}

// Router provides page context and state management.
type Router struct {
	state          map[string]any
	rerender       func()
	navigate       func(path string)
	db             interface{}          // Database connection (server-side only)
	hydrated       bool                 // True if state was hydrated from server
	routeParams    map[string]string    // Route parameters (e.g., :id -> "123")
	suppressRender bool                 // When true, Set() won't trigger re-render (for input events)
	user           *SessionUser         // Current authenticated user (nil if not authenticated)
	request        *http.Request        // Current HTTP request (server-side only)
	response       http.ResponseWriter  // Current HTTP response (server-side only)
	authConfig     *AuthConfig          // Auth configuration (server-side only)
	sessionID      string               // Session ID for SSR API calls
}

// SSRSessionSetter is a function that sets the session context for SSR API calls.
// This is used to propagate authentication to API calls made during page rendering.
// Set this to your generated api.SetEndpointSession function.
var SSRSessionSetter func(sessionID, cookieName string)

// SSRSessionClearer is a function that clears the session context after SSR rendering.
// Set this to your generated api.ClearEndpointSession function.
var SSRSessionClearer func()

// SuppressRender temporarily suppresses re-renders during the callback.
// Used by the DOM renderer for input events to avoid re-render on every keystroke.
func (r *Router) SuppressRender(fn func()) {
	r.suppressRender = true
	fn()
	r.suppressRender = false
}

// NewRouter creates a new router instance.
func NewRouter(rerender func()) *Router {
	return &Router{
		state:       make(map[string]any),
		rerender:    rerender,
		routeParams: make(map[string]string),
	}
}

// NewRouterWithDB creates a router with database access (server-side).
func NewRouterWithDB(db interface{}) *Router {
	return &Router{
		state:       make(map[string]any),
		db:          db,
		routeParams: make(map[string]string),
	}
}

// NewRouterWithParams creates a router with database access and route params.
func NewRouterWithParams(db interface{}, params map[string]string) *Router {
	return &Router{
		state:       make(map[string]any),
		db:          db,
		routeParams: params,
	}
}

// NewRouterWithAuth creates a router with full context including auth.
func NewRouterWithAuth(db interface{}, params map[string]string, user *SessionUser, r *http.Request, w http.ResponseWriter, authConfig *AuthConfig) *Router {
	return &Router{
		state:       make(map[string]any),
		db:          db,
		routeParams: params,
		user:        user,
		request:     r,
		response:    w,
		authConfig:  authConfig,
	}
}

// SetRouteParams sets the route parameters.
func (r *Router) SetRouteParams(params map[string]string) {
	r.routeParams = params
}

// GetRouteParams returns the current route parameters.
func (r *Router) GetRouteParams() map[string]string {
	return r.routeParams
}

// Param returns a route parameter by name.
// Shortcut for r.GetRouteParams()[name].
func (r *Router) Param(name string) string {
	if r.routeParams == nil {
		return ""
	}
	return r.routeParams[name]
}

// Path returns the current request path.
// On server, returns the actual request URL path.
// In WASM, returns the path from route params or empty string.
func (r *Router) Path() string {
	// Server-side: get from request
	if r.request != nil {
		return r.request.URL.Path
	}
	// WASM: would be set via route params or navigation
	if path, ok := r.routeParams["__path"]; ok {
		return path
	}
	return ""
}

// Query returns a URL query parameter value by name.
// On server, reads from the HTTP request URL.
// In WASM, reads from window.location.search.
func (r *Router) Query(name string) string {
	if r.request != nil {
		return r.request.URL.Query().Get(name)
	}
	return wasmQuery(name)
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

	// Extract user from hydrated state
	if userData, ok := state["__gux_user"]; ok && userData != nil {
		// The user data comes as a map from JSON
		if userMap, ok := userData.(map[string]interface{}); ok {
			r.user = &SessionUser{}
			if id, ok := userMap["id"].(string); ok {
				r.user.ID = id
			}
			if email, ok := userMap["email"].(string); ok {
				r.user.Email = email
			}
			if name, ok := userMap["name"].(string); ok {
				r.user.Name = name
			}
			if roles, ok := userMap["roles"].([]interface{}); ok {
				r.user.Roles = make([]string, len(roles))
				for i, role := range roles {
					if roleStr, ok := role.(string); ok {
						r.user.Roles[i] = roleStr
					}
				}
			}
			if metadata, ok := userMap["metadata"].(map[string]interface{}); ok {
				r.user.Metadata = metadata
			}
		}
	}
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

// ClearState clears all page-specific state while preserving system state.
// Called during SPA navigation to ensure fresh data is loaded for the new page.
// Preserves keys starting with "__gux_" (like user session data).
func (r *Router) ClearState() {
	// Preserve system state keys
	preserved := make(map[string]any)
	for k, v := range r.state {
		if len(k) > 6 && k[:6] == "__gux_" {
			preserved[k] = v
		}
	}

	// Clear all state
	r.state = make(map[string]any)

	// Restore preserved keys
	for k, v := range preserved {
		r.state[k] = v
	}

	// Reset hydrated flag so OnLoad will run
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

// StateDebugLog is a debug callback for tracing state changes.
// Set this in WASM code to enable state debugging.
// Parameters: action ("SET" or "GET"), key, value, router
var StateDebugLog func(action, key string, value any, router *Router)

// State represents a reactive state value.
type State[T any] struct {
	key    string
	router *Router
}

// Get returns the current state value.
func (s *State[T]) Get() T {
	// Debug: log state access (uses callback if set in WASM mode)
	if StateDebugLog != nil {
		StateDebugLog("GET", s.key, s.router.state[s.key], s.router)
	}
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
		// Handle complex types from JSON (slices, maps, structs)
		// Re-marshal and unmarshal to convert []interface{} to []T etc.
		if jsonBytes, err := json.Marshal(val); err == nil {
			var result T
			if err := json.Unmarshal(jsonBytes, &result); err == nil {
				// Cache the converted value for future Gets
				s.router.state[s.key] = result
				return result
			}
		}
	}
	var zero T
	return zero
}

// Set updates the state and schedules a re-render.
// If called during an input event (via SuppressRender), the re-render is skipped
// to avoid losing focus on every keystroke. This is similar to how React and Vue
// automatically batch state updates during input events.
func (s *State[T]) Set(val T) {
	// Debug: log state change (uses callback if set in WASM mode)
	if StateDebugLog != nil {
		StateDebugLog("SET", s.key, val, s.router)
	}
	s.router.state[s.key] = val
	if !s.router.suppressRender {
		ScheduleRerender(s.router)
	}
}

// SetQuiet updates the state WITHOUT triggering a re-render.
// Use this for input field updates where you want to track the value
// but don't need to re-render the UI on every keystroke.
func (s *State[T]) SetQuiet(val T) {
	s.router.state[s.key] = val
}

// UseState creates a reactive state for any type.
// Usage: count := core.UseState(r, "count", 0)
//        user := core.UseState(r, "user", User{Name: "John"})
func UseState[T any](r *Router, key string, initial T) *State[T] {
	if _, ok := r.state[key]; !ok {
		r.state[key] = initial
		if StateDebugLog != nil {
			StateDebugLog("INIT", key, initial, r)
		}
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

// User returns the current authenticated user, or nil if not authenticated.
func (r *Router) User() *SessionUser {
	return r.user
}

// IsAuthenticated returns true if the user is authenticated.
func (r *Router) IsAuthenticated() bool {
	return r.user != nil
}

// HasRole checks if the current user has a specific role.
func (r *Router) HasRole(role string) bool {
	if r.user == nil {
		return false
	}
	return r.user.HasRole(role)
}

// HasAnyRole checks if the current user has any of the specified roles.
func (r *Router) HasAnyRole(roles ...string) bool {
	if r.user == nil {
		return false
	}
	return r.user.HasAnyRole(roles...)
}

// HasAllRoles checks if the current user has all of the specified roles.
func (r *Router) HasAllRoles(roles ...string) bool {
	if r.user == nil {
		return false
	}
	return r.user.HasAllRoles(roles...)
}

// Login creates a session for the user.
// Only works on server-side. Returns an error if session creation fails.
func (r *Router) Login(user *SessionUser) error {
	if r.authConfig == nil || r.authConfig.SessionStore == nil {
		return fmt.Errorf("auth not configured")
	}
	if r.response == nil {
		return fmt.Errorf("Login() can only be called server-side")
	}

	// Generate new session ID (session fixation prevention)
	sessionID, err := generateSessionID()
	if err != nil {
		return fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Store session
	maxAge := time.Duration(r.authConfig.CookieMaxAge) * time.Second
	if err := r.authConfig.SessionStore.Set(sessionID, user, maxAge); err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	// Set session cookie
	setSessionCookie(r.response, sessionID, *r.authConfig)

	// Update router's user
	r.user = user

	return nil
}

// Logout destroys the current session.
// Only works on server-side.
func (r *Router) Logout() error {
	if r.authConfig == nil || r.authConfig.SessionStore == nil {
		return fmt.Errorf("auth not configured")
	}
	if r.response == nil || r.request == nil {
		return fmt.Errorf("Logout() can only be called server-side")
	}

	// Get current session ID
	sessionID := getSessionIDFromCookie(r.request, r.authConfig.CookieName)
	if sessionID != "" {
		// Delete from store
		r.authConfig.SessionStore.Delete(sessionID)
	}

	// Clear session cookie
	clearSessionCookie(r.response, *r.authConfig)

	// Clear router's user
	r.user = nil

	return nil
}

// Redirect performs an HTTP redirect. Only works server-side.
func (r *Router) Redirect(path string) {
	if r.response != nil && r.request != nil {
		http.Redirect(r.response, r.request, path, http.StatusFound)
	}
}
