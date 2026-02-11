//go:build !js || !wasm

package core

import (
	"fmt"
	"net/http"
	"time"
)

// ScheduleRerender on server-side just calls rerender immediately.
// Server-side rendering doesn't need deferred re-renders since it's a single pass.
func ScheduleRerender(r *Router) {
	if r.rerender != nil {
		r.rerender()
	}
}

// SetInterval is a no-op on server — timers only run in WASM.
func (r *Router) SetInterval(callback func(), ms int) *TimerHandle {
	return &TimerHandle{cleared: true}
}

// SetTimeout is a no-op on server — timers only run in WASM.
func (r *Router) SetTimeout(callback func(), ms int) *TimerHandle {
	return &TimerHandle{cleared: true}
}

func clearTimer(h *TimerHandle) {}

// SSRSessionSetter is a function that sets the session context for SSR API calls.
// This is used to propagate authentication to API calls made during page rendering.
// Set this to your generated api.SetEndpointSession function.
var SSRSessionSetter func(sessionID, cookieName string)

// SSRSessionClearer is a function that clears the session context after SSR rendering.
// Set this to your generated api.ClearEndpointSession function.
var SSRSessionClearer func()

// SSRCookieSetter is a function that forwards all cookies from the incoming request
// to SSR API calls. This enables endpoints that depend on non-session cookies
// (e.g., org selection, preferences) to work correctly during server-side rendering.
// Set this to your generated api.SetEndpointCookies function.
var SSRCookieSetter func(cookies []*http.Cookie)

// SSRCookieClearer is a function that clears forwarded cookies after SSR rendering.
// Set this to your generated api.ClearEndpointCookies function.
var SSRCookieClearer func()

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

// Path returns the current request path.
// On server, returns the actual request URL path.
func (r *Router) Path() string {
	if r.request != nil {
		if req, ok := r.request.(*http.Request); ok {
			return req.URL.Path
		}
	}
	if path, ok := r.routeParams["__path"]; ok {
		return path
	}
	return ""
}

// Query returns a URL query parameter value by name.
// On server, reads from the HTTP request URL.
func (r *Router) Query(name string) string {
	if r.request != nil {
		if req, ok := r.request.(*http.Request); ok {
			return req.URL.Query().Get(name)
		}
	}
	return ""
}

// Login creates a session for the user.
// Only works on server-side. Returns an error if session creation fails.
func (r *Router) Login(user *SessionUser) error {
	authConfig, ok := r.authConfig.(*AuthConfig)
	if !ok || authConfig == nil || authConfig.SessionStore == nil {
		return fmt.Errorf("auth not configured")
	}
	response, ok := r.response.(http.ResponseWriter)
	if !ok || response == nil {
		return fmt.Errorf("Login() can only be called server-side")
	}

	// Generate new session ID (session fixation prevention)
	sessionID, err := generateSessionID()
	if err != nil {
		return fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Store session
	maxAge := time.Duration(authConfig.CookieMaxAge) * time.Second
	if err := authConfig.SessionStore.Set(sessionID, user, maxAge); err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	// Set session cookie
	setSessionCookie(response, sessionID, *authConfig)

	// Update router's user
	r.user = user

	return nil
}

// Logout destroys the current session.
// Only works on server-side.
func (r *Router) Logout() error {
	authConfig, ok := r.authConfig.(*AuthConfig)
	if !ok || authConfig == nil || authConfig.SessionStore == nil {
		return fmt.Errorf("auth not configured")
	}
	response, ok := r.response.(http.ResponseWriter)
	if !ok || response == nil {
		return fmt.Errorf("Logout() can only be called server-side")
	}
	request, ok := r.request.(*http.Request)
	if !ok || request == nil {
		return fmt.Errorf("Logout() can only be called server-side")
	}

	// Get current session ID
	sessionID := getSessionIDFromCookie(request, authConfig.CookieName)
	if sessionID != "" {
		// Delete from store
		authConfig.SessionStore.Delete(sessionID)
	}

	// Clear session cookie
	clearSessionCookie(response, *authConfig)

	// Clear router's user
	r.user = nil

	return nil
}

// Redirect performs a full-page navigation.
// In WASM: uses window.location.assign() for a full page load (cross-bundle safe).
// On server: performs an HTTP 302 redirect.
// Use this instead of Navigate() when crossing WASM bundle boundaries
// or when you need a guaranteed full page load (e.g., after login/logout).
func (r *Router) Redirect(path string) {
	if r.redirect != nil {
		r.redirect(path)
		return
	}
	if r.response != nil && r.request != nil {
		if resp, ok := r.response.(http.ResponseWriter); ok {
			if req, ok := r.request.(*http.Request); ok {
				http.Redirect(resp, req, path, http.StatusFound)
			}
		}
	}
}
