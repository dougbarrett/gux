package core

import "time"

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

// SessionStore is the interface for session storage backends.
type SessionStore interface {
	// Get retrieves a session user by session ID
	Get(sessionID string) (*SessionUser, error)

	// Set stores a session user with the given session ID and max age
	Set(sessionID string, user *SessionUser, maxAge time.Duration) error

	// Delete removes a session by ID
	Delete(sessionID string) error
}
