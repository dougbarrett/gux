package core

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	// SessionIDLength is the length of the session ID in bytes (32 bytes = 256 bits)
	SessionIDLength = 32

	// DefaultSessionCookieName is the default name for the session cookie
	DefaultSessionCookieName = "__gux_session"

	// DefaultSessionMaxAge is the default session max age in seconds (24 hours)
	DefaultSessionMaxAge = 86400

	// SessionUserKey is the context key for the session user
	SessionUserKey = "gux_session_user"
)

// SessionUser represents an authenticated user.
type SessionUser struct {
	ID       string                 `json:"id"`
	Email    string                 `json:"email,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Roles    []string               `json:"roles,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// HasRole checks if the user has a specific role.
func (u *SessionUser) HasRole(role string) bool {
	if u == nil {
		return false
	}
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles.
func (u *SessionUser) HasAnyRole(roles ...string) bool {
	if u == nil {
		return false
	}
	for _, role := range roles {
		if u.HasRole(role) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if the user has all of the specified roles.
func (u *SessionUser) HasAllRoles(roles ...string) bool {
	if u == nil {
		return false
	}
	for _, role := range roles {
		if !u.HasRole(role) {
			return false
		}
	}
	return true
}

// AuthConfig configures the authentication system.
type AuthConfig struct {
	// SessionStore is the storage backend for sessions (required)
	SessionStore SessionStore

	// CookieName is the name of the session cookie (default: "__gux_session")
	CookieName string

	// CookieMaxAge is the max age in seconds (default: 86400 = 24 hours)
	CookieMaxAge int

	// CookieSecure sets the Secure flag on the cookie (default: false for dev)
	CookieSecure bool

	// CookieHTTPOnly sets the HttpOnly flag (default: true)
	CookieHTTPOnly bool

	// CookieSameSite sets the SameSite attribute (default: Lax)
	CookieSameSite http.SameSite

	// CookiePath sets the Path attribute (default: "/")
	CookiePath string

	// LoginPath is the path to redirect unauthenticated users (default: "/login")
	LoginPath string

	// UnauthorizedHandler is called when an unauthenticated user accesses a protected route
	// If nil, redirects to LoginPath
	UnauthorizedHandler func(w http.ResponseWriter, r *http.Request)

	// ForbiddenHandler is called when a user lacks required roles
	// If nil, returns 403 Forbidden
	ForbiddenHandler func(w http.ResponseWriter, r *http.Request)
}

// DefaultAuthConfig returns the default auth configuration.
// Note: SessionStore must be set by the user.
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		CookieName:     DefaultSessionCookieName,
		CookieMaxAge:   DefaultSessionMaxAge,
		CookieSecure:   false,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteLaxMode,
		CookiePath:     "/",
		LoginPath:      "/login",
	}
}

// SessionStore is the interface for session storage backends.
type SessionStore interface {
	// Get retrieves a session user by session ID
	Get(sessionID string) (*SessionUser, error)

	// Set stores a session user with the given session ID and max age
	Set(sessionID string, user *SessionUser, maxAge time.Duration) error

	// Delete removes a session by ID
	Delete(sessionID string) error
}

// sessionEntry holds a session user with expiration time
type sessionEntry struct {
	User      *SessionUser
	ExpiresAt time.Time
}

// MemorySessionStore is an in-memory session store for development/testing.
// Sessions are lost on server restart. For persistence, use DBSessionStore.
type MemorySessionStore struct {
	sessions map[string]*sessionEntry
	mu       sync.RWMutex
}

// NewMemorySessionStore creates a new in-memory session store.
func NewMemorySessionStore() *MemorySessionStore {
	store := &MemorySessionStore{
		sessions: make(map[string]*sessionEntry),
	}

	// Start background cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			store.cleanup()
		}
	}()

	return store
}

// Get retrieves a session user by session ID.
func (s *MemorySessionStore) Get(sessionID string) (*SessionUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.sessions[sessionID]
	if !exists {
		return nil, nil
	}

	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		return nil, nil
	}

	return entry.User, nil
}

// Set stores a session user with the given session ID and max age.
func (s *MemorySessionStore) Set(sessionID string, user *SessionUser, maxAge time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[sessionID] = &sessionEntry{
		User:      user,
		ExpiresAt: time.Now().Add(maxAge),
	}
	return nil
}

// Delete removes a session by ID.
func (s *MemorySessionStore) Delete(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
	return nil
}

// cleanup removes expired sessions.
func (s *MemorySessionStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, entry := range s.sessions {
		if now.After(entry.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}

// generateSessionID generates a cryptographically secure session ID.
func generateSessionID() (string, error) {
	bytes := make([]byte, SessionIDLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// setSessionCookie sets the session cookie on the response.
func setSessionCookie(w http.ResponseWriter, sessionID string, config AuthConfig) {
	cookie := &http.Cookie{
		Name:     config.CookieName,
		Value:    sessionID,
		Path:     config.CookiePath,
		MaxAge:   config.CookieMaxAge,
		HttpOnly: config.CookieHTTPOnly,
		Secure:   config.CookieSecure,
		SameSite: config.CookieSameSite,
	}
	http.SetCookie(w, cookie)
}

// clearSessionCookie removes the session cookie.
func clearSessionCookie(w http.ResponseWriter, config AuthConfig) {
	cookie := &http.Cookie{
		Name:     config.CookieName,
		Value:    "",
		Path:     config.CookiePath,
		MaxAge:   -1,
		HttpOnly: config.CookieHTTPOnly,
		Secure:   config.CookieSecure,
		SameSite: config.CookieSameSite,
	}
	http.SetCookie(w, cookie)
}

// getSessionIDFromCookie retrieves the session ID from the request cookie.
func getSessionIDFromCookie(r *http.Request, cookieName string) string {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// sessionUserToJSON converts a SessionUser to JSON for hydration.
func sessionUserToJSON(user *SessionUser) string {
	if user == nil {
		return ""
	}
	data, err := json.Marshal(user)
	if err != nil {
		return ""
	}
	return string(data)
}

// sessionUserFromJSON parses a SessionUser from JSON.
func sessionUserFromJSON(data string) *SessionUser {
	if data == "" {
		return nil
	}
	var user SessionUser
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil
	}
	return &user
}

// GetSessionUser retrieves the authenticated user from a raw http.Request.
// Returns nil if not authenticated or auth is not configured.
//
// Use this in raw http.Handler functions registered via app.HandleFunc():
//
//	app.HandleFunc("GET /api/me", func(w http.ResponseWriter, r *http.Request) {
//	    user := core.GetSessionUser(app, r)
//	    if user == nil {
//	        http.Error(w, "unauthorized", 401)
//	        return
//	    }
//	    // user.ID, user.Email, user.Roles, etc.
//	})
func GetSessionUser(app *App, r *http.Request) *SessionUser {
	return app.getUserFromRequest(r)
}

// LoginFromHTTP creates a session for the user and sets the session cookie.
// Use this in raw http.Handler functions registered via app.HandleFunc():
//
//	app.HandleFunc("POST /api/register", func(w http.ResponseWriter, r *http.Request) {
//	    user := createUser(r)
//	    if err := core.LoginFromHTTP(app, w, &core.SessionUser{
//	        ID:    fmt.Sprint(user.ID),
//	        Email: user.Email,
//	        Roles: []string{user.Role},
//	    }); err != nil {
//	        http.Error(w, err.Error(), 500)
//	        return
//	    }
//	})
func LoginFromHTTP(app *App, w http.ResponseWriter, user *SessionUser) error {
	if app.authConfig == nil || app.authConfig.SessionStore == nil {
		return fmt.Errorf("auth not configured")
	}

	sessionID, err := generateSessionID()
	if err != nil {
		return fmt.Errorf("failed to generate session: %w", err)
	}

	maxAge := app.authConfig.CookieMaxAge
	if maxAge == 0 {
		maxAge = DefaultSessionMaxAge
	}

	if err := app.authConfig.SessionStore.Set(sessionID, user, time.Duration(maxAge)*time.Second); err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	setSessionCookie(w, sessionID, *app.authConfig)
	return nil
}

// LogoutFromHTTP destroys the session and clears the session cookie.
// Use this in raw http.Handler functions registered via app.HandleFunc():
//
//	app.HandleFunc("POST /api/logout", func(w http.ResponseWriter, r *http.Request) {
//	    core.LogoutFromHTTP(app, w, r)
//	})
func LogoutFromHTTP(app *App, w http.ResponseWriter, r *http.Request) {
	if app.authConfig == nil || app.authConfig.SessionStore == nil {
		return
	}

	cookieName := app.authConfig.CookieName
	if cookieName == "" {
		cookieName = DefaultSessionCookieName
	}

	sessionID := getSessionIDFromCookie(r, cookieName)
	if sessionID != "" {
		app.authConfig.SessionStore.Delete(sessionID)
	}

	clearSessionCookie(w, *app.authConfig)
}
