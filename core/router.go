package core

import "encoding/json"

// PageFunc is a page handler that returns a component function.
// The outer function runs on server (loader), the inner runs on client (component).
type PageFunc func(r *Router) func() Node

// Router provides page context and state management.
type Router struct {
	state          map[string]any
	rerender       func()
	navigate       func(path string)
	redirect       func(path string)
	db             interface{}       // Database connection (server-side only)
	hydrated       bool              // True if state was hydrated from server
	routeParams    map[string]string // Route parameters (e.g., :id -> "123")
	suppressRender bool              // When true, Set() won't trigger re-render (for input events)
	user           *SessionUser      // Current authenticated user (nil if not authenticated)
	request        any               // *http.Request on server, nil in WASM
	response       any               // http.ResponseWriter on server, nil in WASM
	authConfig     any               // *AuthConfig on server, nil in WASM
	sessionID      string            // Session ID for SSR API calls
	activeTimers   []*TimerHandle    // Tracked timers/intervals for cleanup on navigation
}

// TimerHandle represents a JavaScript timer (setTimeout or setInterval).
// Call Clear() to cancel the timer manually, or it will be automatically
// cleared when the user navigates to a different page.
type TimerHandle struct {
	id       int
	interval bool // true = setInterval, false = setTimeout
	cleared  bool
	router   *Router
}

// Clear cancels the timer. Safe to call multiple times.
func (h *TimerHandle) Clear() {
	if h.cleared {
		return
	}
	h.cleared = true
	clearTimer(h)
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
//
//	user := core.UseState(r, "user", User{Name: "John"})
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

// NewRouter creates a new router instance.
func NewRouter(rerender func()) *Router {
	return &Router{
		state:       make(map[string]any),
		rerender:    rerender,
		routeParams: make(map[string]string),
	}
}

// SuppressRender temporarily suppresses re-renders during the callback.
// Used by the DOM renderer for input events to avoid re-render on every keystroke.
func (r *Router) SuppressRender(fn func()) {
	r.suppressRender = true
	fn()
	r.suppressRender = false
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

// DB returns the database connection for server-side queries.
// Returns nil in WASM - use api package for client-side data fetching.
func (r *Router) DB() interface{} {
	return r.db
}

// SetNavigate sets the navigation callback (called by WASM runtime).
func (r *Router) SetNavigate(fn func(path string)) {
	r.navigate = fn
}

// SetRedirect sets the redirect callback (called by WASM runtime).
// This is used for full-page navigation (window.location.assign).
func (r *Router) SetRedirect(fn func(path string)) {
	r.redirect = fn
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

	// Clear all active timers/intervals
	for _, h := range r.activeTimers {
		h.Clear()
	}
	r.activeTimers = nil

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
