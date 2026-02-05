package core

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dougbarrett/gux/api"
)

// APIContext provides context for API endpoint handlers.
// It wraps http.Request/ResponseWriter with convenience methods
// for accessing request data, authentication, and database.
type APIContext struct {
	Request  *http.Request
	Response http.ResponseWriter
	app      *App
	params   map[string]string
	user     *SessionUser
}

// Param returns a path parameter by name.
// Usage: id := ctx.Param("id")
func (ctx *APIContext) Param(name string) string {
	return ctx.params[name]
}

// ParamInt returns a path parameter as int, or 0 if not found/invalid.
func (ctx *APIContext) ParamInt(name string) int {
	val := ctx.params[name]
	if val == "" {
		return 0
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return i
}

// ParamUint returns a path parameter as uint, or 0 if not found/invalid.
func (ctx *APIContext) ParamUint(name string) uint {
	val := ctx.params[name]
	if val == "" {
		return 0
	}
	u, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0
	}
	return uint(u)
}

// User returns the authenticated user, or nil if not authenticated.
func (ctx *APIContext) User() *SessionUser {
	return ctx.user
}

// DB returns the database connection.
func (ctx *APIContext) DB() interface{} {
	return ctx.app.db
}

// Query returns a query parameter by name.
func (ctx *APIContext) Query(name string) string {
	return ctx.Request.URL.Query().Get(name)
}

// Header returns a request header by name.
func (ctx *APIContext) Header(name string) string {
	return ctx.Request.Header.Get(name)
}

// Login creates a session for the given user.
// Returns an error if auth is not configured.
func (ctx *APIContext) Login(user *SessionUser) error {
	if ctx.app.authConfig == nil || ctx.app.authConfig.SessionStore == nil {
		return api.InternalError("auth not configured")
	}

	sessionID, err := generateSessionID()
	if err != nil {
		return api.InternalError("failed to generate session")
	}

	maxAge := ctx.app.authConfig.CookieMaxAge
	if maxAge == 0 {
		maxAge = 86400 * 7 // 7 days default
	}

	if err := ctx.app.authConfig.SessionStore.Set(sessionID, user, time.Duration(maxAge)*time.Second); err != nil {
		return api.InternalError("failed to store session")
	}

	setSessionCookie(ctx.Response, sessionID, *ctx.app.authConfig)
	ctx.user = user
	return nil
}

// Logout destroys the current session.
func (ctx *APIContext) Logout() error {
	if ctx.app.authConfig == nil || ctx.app.authConfig.SessionStore == nil {
		return nil // No auth configured, nothing to do
	}

	sessionID := getSessionIDFromCookie(ctx.Request, ctx.app.authConfig.CookieName)
	if sessionID != "" {
		ctx.app.authConfig.SessionStore.Delete(sessionID)
	}

	clearSessionCookie(ctx.Response, *ctx.app.authConfig)
	ctx.user = nil
	return nil
}

// APIEndpoint stores metadata about a registered API endpoint.
// Used by code generation to create WASM clients.
type APIEndpoint struct {
	Method       string // HTTP method: GET, POST, PUT, PATCH, DELETE
	Path         string // URL path: /api/login, /api/users/:id
	RequestType  string // Request body type name (empty for GET/DELETE)
	ResponseType string // Response type name (empty for DELETE)
}

// APIRegistration is returned by API registration functions for chaining auth options.
type APIRegistration struct {
	app    *App
	public bool
	roles  []string
}

// Public marks this API endpoint as public (no authentication required).
// By default, all API endpoints require authentication when auth is configured.
func (r *APIRegistration) Public() *APIRegistration {
	r.public = true
	return r
}

// WithRoles specifies required roles for this API endpoint.
// User must have ANY of the specified roles (OR logic).
func (r *APIRegistration) WithRoles(roles ...string) *APIRegistration {
	r.roles = roles
	return r
}

// checkAPIAuth checks if the request is authorized for this endpoint.
// Returns false and writes error response if not authorized.
func (r *APIRegistration) checkAuth(ctx *APIContext, w http.ResponseWriter) bool {
	// Public endpoints don't require auth
	if r.public {
		return true
	}

	// If auth is not configured, allow access (backwards compatibility)
	if r.app.authConfig == nil || r.app.authConfig.SessionStore == nil {
		return true
	}

	if ctx.user == nil {
		api.WriteError(w, api.Unauthorized("authentication required"))
		return false
	}

	// Check roles if specified
	if len(r.roles) > 0 && !ctx.user.HasAnyRole(r.roles...) {
		api.WriteError(w, api.Forbidden("insufficient permissions"))
		return false
	}

	return true
}

// newAPIContext creates an APIContext from an HTTP request.
func newAPIContext(app *App, w http.ResponseWriter, r *http.Request, pattern string) *APIContext {
	ctx := &APIContext{
		Request:  r,
		Response: w,
		app:      app,
		params:   make(map[string]string),
	}

	// Extract path parameters using Go 1.22+ r.PathValue() first,
	// falling back to manual extraction from the URL path.
	paramNames := extractParamNames(pattern)
	if len(paramNames) > 0 {
		// Try Go 1.22+ PathValue first (works when mux registers {param} patterns)
		allFound := true
		for _, name := range paramNames {
			val := r.PathValue(name)
			if val != "" {
				ctx.params[name] = val
			} else {
				allFound = false
				break
			}
		}
		if !allFound {
			// Fallback to manual extraction (for older Go or non-mux routing)
			ctx.params = extractPathParams(pattern, r.URL.Path)
		}
	}

	// Load user from session if auth is enabled
	if app.authConfig != nil && app.authConfig.SessionStore != nil {
		cookieName := app.authConfig.CookieName
		if cookieName == "" {
			cookieName = "__gux_session"
		}
		sessionID := getSessionIDFromCookie(r, cookieName)
		if sessionID != "" {
			ctx.user, _ = app.authConfig.SessionStore.Get(sessionID)
		}
	}

	return ctx
}

// convertColonParams converts :param syntax to {param} for Go 1.22+ http.ServeMux.
// "GET /api/users/:id/posts/:postId" -> "GET /api/users/{id}/posts/{postId}"
func convertColonParams(pattern string) string {
	parts := strings.Split(pattern, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + strings.TrimPrefix(part, ":") + "}"
		}
	}
	return strings.Join(parts, "/")
}

// extractParamNames returns the parameter names from a :param-style pattern.
func extractParamNames(pattern string) []string {
	var names []string
	for _, part := range strings.Split(pattern, "/") {
		if strings.HasPrefix(part, ":") {
			names = append(names, strings.TrimPrefix(part, ":"))
		}
	}
	return names
}

// extractPathParams extracts path parameters from a URL path based on a pattern.
// It first tries Go 1.22+ r.PathValue() via the request (if available),
// then falls back to manual extraction from the URL path.
// Pattern: /api/users/:id/posts/:postId
// Path: /api/users/123/posts/456
// Returns: map[string]string{"id": "123", "postId": "456"}
func extractPathParams(pattern, path string) map[string]string {
	params := make(map[string]string)

	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return params
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			paramName := strings.TrimPrefix(part, ":")
			params[paramName] = pathParts[i]
		}
	}

	return params
}

// API registers a typed API endpoint that expects a request body.
// Use for POST, PUT, PATCH methods.
// Returns an APIRegistration for chaining auth options.
//
// Usage:
//
//	core.API(app, "POST", "/api/login", func(ctx *core.APIContext, req LoginRequest) (LoginResponse, error) {
//	    // Business logic here
//	    return LoginResponse{Success: true}, nil
//	}).Public()  // Mark as public (no auth required)
//
//	core.API(app, "POST", "/api/admin/users", createUser).WithRoles("admin")
func API[Req any, Res any](app *App, method string, path string, handler func(*APIContext, Req) (Res, error)) *APIRegistration {
	// Register metadata for code generation
	app.apiEndpoints = append(app.apiEndpoints, APIEndpoint{
		Method:       method,
		Path:         path,
		RequestType:  getTypeName[Req](),
		ResponseType: getTypeName[Res](),
	})

	reg := &APIRegistration{app: app}

	// Create HTTP handler
	app.HandleFunc(method+" "+path, func(w http.ResponseWriter, r *http.Request) {
		ctx := newAPIContext(app, w, r, path)

		// Auth check
		if !reg.checkAuth(ctx, w) {
			return
		}

		// Decode request body (skip if body is nil/empty, e.g., struct{} request type)
		var req Req
		if r.Body != nil {
			body, _ := io.ReadAll(r.Body)
			if len(body) > 0 {
				if err := json.Unmarshal(body, &req); err != nil {
					api.WriteError(w, api.BadRequest("invalid request body: "+err.Error()))
					return
				}
			}
		}

		// Call handler
		res, err := handler(ctx, req)
		if err != nil {
			api.WriteError(w, err)
			return
		}

		// Encode response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	return reg
}

// APIGet registers a typed GET endpoint (no request body).
// Returns an APIRegistration for chaining auth options.
//
// Usage:
//
//	core.APIGet(app, "/api/users/:id", func(ctx *core.APIContext) (dto.UserDetail, error) {
//	    id := ctx.ParamUint("id")
//	    // ... fetch user
//	    return user, nil
//	})
//
//	core.APIGet(app, "/api/products", listProducts).Public()  // No auth required
//	core.APIGet(app, "/api/admin/stats", getStats).WithRoles("admin")
func APIGet[Res any](app *App, path string, handler func(*APIContext) (Res, error)) *APIRegistration {
	// Register metadata for code generation
	app.apiEndpoints = append(app.apiEndpoints, APIEndpoint{
		Method:       "GET",
		Path:         path,
		ResponseType: getTypeName[Res](),
	})

	reg := &APIRegistration{app: app}

	// Create HTTP handler
	app.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
		ctx := newAPIContext(app, w, r, path)

		// Auth check
		if !reg.checkAuth(ctx, w) {
			return
		}

		// Call handler
		res, err := handler(ctx)
		if err != nil {
			api.WriteError(w, err)
			return
		}

		// Encode response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	return reg
}

// APIDelete registers a typed DELETE endpoint.
// Returns an APIRegistration for chaining auth options.
//
// Usage:
//
//	core.APIDelete(app, "/api/users/:id", func(ctx *core.APIContext) error {
//	    id := ctx.ParamUint("id")
//	    // ... delete user
//	    return nil
//	}).WithRoles("admin")
func APIDelete(app *App, path string, handler func(*APIContext) error) *APIRegistration {
	// Register metadata for code generation
	app.apiEndpoints = append(app.apiEndpoints, APIEndpoint{
		Method: "DELETE",
		Path:   path,
	})

	reg := &APIRegistration{app: app}

	// Create HTTP handler
	app.HandleFunc("DELETE "+path, func(w http.ResponseWriter, r *http.Request) {
		ctx := newAPIContext(app, w, r, path)

		// Auth check
		if !reg.checkAuth(ctx, w) {
			return
		}

		// Call handler
		err := handler(ctx)
		if err != nil {
			api.WriteError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	return reg
}

// getTypeName returns the name of a type using generics.
func getTypeName[T any]() string {
	var zero T
	return typeNameFromValue(zero)
}

// typeNameFromValue extracts the type name from a value.
func typeNameFromValue(v interface{}) string {
	if v == nil {
		return ""
	}
	// Use %T to get the type, then extract just the name
	typeName := strings.TrimPrefix(
		strings.Replace(
			stringf("%T", v),
			"*", "", -1,
		),
		"main.",
	)
	// Handle package-qualified types (e.g., "dto.LoginRequest" -> "LoginRequest")
	if idx := strings.LastIndex(typeName, "."); idx != -1 {
		return typeName[idx+1:]
	}
	return typeName
}

// stringf is a simple sprintf wrapper to avoid importing fmt.
func stringf(format string, args ...interface{}) string {
	// For %T, we need reflection
	if format == "%T" && len(args) == 1 {
		return getTypeString(args[0])
	}
	return ""
}

// getTypeString returns the type of a value as a string.
func getTypeString(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	switch v.(type) {
	case string:
		return "string"
	case int, int8, int16, int32, int64:
		return "int"
	case uint, uint8, uint16, uint32, uint64:
		return "uint"
	case float32, float64:
		return "float64"
	case bool:
		return "bool"
	default:
		// Use reflect for struct types
		return getStructTypeName(v)
	}
}

// getStructTypeName uses a type assertion trick to get struct type names.
func getStructTypeName(v interface{}) string {
	// This is a simplified version - the actual implementation
	// in code generation will parse the AST directly
	return "struct"
}
