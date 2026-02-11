//go:build !js || !wasm

package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

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
	storage        Storage                      // File storage backend (nil = disabled)
	server         *http.Server                 // Underlying server (set by Run)
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

// SetStorage configures file storage for the application.
// This enables the upload endpoint at /__gux_api/upload and file serving at /__gux_api/files/{key}.
//
// Usage:
//
//	storage := core.NewLocalStorage("uploads", core.WithMaxSize(10<<20), core.WithAllowedTypes("image/*"))
//	app.SetStorage(storage)
func (a *App) SetStorage(s Storage) *App {
	a.storage = s
	return a
}

// Storage returns the configured storage backend, or nil if not set.
func (a *App) Storage() Storage {
	return a.storage
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

// Handler returns the configured http.Handler with all routes and middleware applied.
// Useful for testing with httptest.NewServer or httptest.NewRecorder.
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()

	// Register CRUD handlers
	a.registerCRUDHandlers(mux)

	// Register upload and file serving handlers (if storage configured)
	if a.storage != nil {
		a.registerUploadHandlers(mux)
	}

	// Register custom handlers — convert :param to {param} for Go 1.22+ ServeMux
	for pattern, handler := range a.customHandlers {
		mux.HandleFunc(convertColonParams(pattern), handler)
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
				if SSRCookieSetter != nil {
					SSRCookieSetter(r.Cookies())
				}

				router := NewRouterWithAuth(a.db, params, user, r, w, a.authConfig)
				router.sessionID = sessionID
				component := route.Handler(router)
				component()

				// Clear session/cookie context after rendering
				if SSRSessionClearer != nil {
					SSRSessionClearer()
				}
				if SSRCookieClearer != nil {
					SSRCookieClearer()
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
				if SSRCookieSetter != nil {
					SSRCookieSetter(r.Cookies())
				}

				// Create router with auth context
				router := NewRouterWithAuth(a.db, params, user, r, w, a.authConfig)
				router.sessionID = sessionID
				component := route.Handler(router)
				html := component().Render(HTML()).HTML()

				// Clear session/cookie context after rendering
				if SSRSessionClearer != nil {
					SSRSessionClearer()
				}
				if SSRCookieClearer != nil {
					SSRCookieClearer()
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

				// Get or generate CSRF token
				csrfToken := ""
				if a.csrfConfig.Enabled {
					// Reuse existing cookie token to avoid meta/cookie desync
					csrfToken = getCSRFFromCookie(r)
					if csrfToken == "" {
						var err error
						csrfToken, err = generateCSRFToken()
						if err == nil {
							setCSRFCookie(w, csrfToken, a.csrfConfig)
						}
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
    <meta charset="utf-8">
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
    <meta charset="utf-8">
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

	// Apply CSRF middleware if enabled
	var handler http.Handler = mux
	if a.csrfConfig.Enabled {
		handler = CSRFMiddleware(a.csrfConfig)(mux)
	}

	return handler
}

// Run starts the HTTP server on the given address with graceful shutdown.
// On SIGTERM or SIGINT, it stops accepting new connections and waits up to
// 30 seconds for in-flight requests to complete before returning.
func (a *App) Run(addr string) error {
	fmt.Printf("http://localhost%s\n", addr)
	return a.listenAndServe(addr)
}

// Server returns the underlying *http.Server after Run has been called.
// Returns nil if Run has not been called yet. Use this for custom shutdown
// logic (e.g., stopping background workers before draining connections).
func (a *App) Server() *http.Server {
	return a.server
}
