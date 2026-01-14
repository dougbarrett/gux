//go:build js && wasm

package auth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"syscall/js"
	"time"
)

// User represents an authenticated user
type User struct {
	ID       string         `json:"id"`
	Email    string         `json:"email"`
	Name     string         `json:"name"`
	Roles    []string       `json:"roles"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// AuthState represents the current authentication state
type AuthState struct {
	User         *User
	Token        string
	RefreshToken string
	ExpiresAt    time.Time
}

// Auth manages authentication state
type Auth struct {
	state       AuthState
	subscribers []func(AuthState)
	storageKey  string
}

var globalAuth *Auth

// Init initializes the global auth manager
func Init() *Auth {
	if globalAuth != nil {
		return globalAuth
	}

	globalAuth = &Auth{
		storageKey: "auth_state",
	}

	// Try to restore from storage
	globalAuth.restore()

	return globalAuth
}

// GetAuth returns the global auth instance
func GetAuth() *Auth {
	if globalAuth == nil {
		Init()
	}
	return globalAuth
}

// Login sets the authentication state
func Login(token string, user *User) {
	auth := GetAuth()

	auth.state = AuthState{
		User:      user,
		Token:     token,
		ExpiresAt: extractExpiry(token),
	}

	auth.persist()
	auth.notify()
}

// LoginWithTokens sets auth state with both access and refresh tokens
func LoginWithTokens(token, refreshToken string, user *User) {
	auth := GetAuth()

	auth.state = AuthState{
		User:         user,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    extractExpiry(token),
	}

	auth.persist()
	auth.notify()
}

// Logout clears the authentication state
func Logout() {
	auth := GetAuth()
	auth.state = AuthState{}
	js.Global().Get("localStorage").Call("removeItem", auth.storageKey)
	auth.notify()
}

// IsAuthenticated returns true if user is logged in
func IsAuthenticated() bool {
	auth := GetAuth()
	return auth.state.Token != "" && auth.state.User != nil
}

// IsTokenValid returns true if token hasn't expired
func IsTokenValid() bool {
	auth := GetAuth()
	if !IsAuthenticated() {
		return false
	}
	return time.Now().Before(auth.state.ExpiresAt)
}

// GetUser returns the current user
func GetUser() *User {
	return GetAuth().state.User
}

// GetToken returns the current JWT token
func GetToken() string {
	return GetAuth().state.Token
}

// GetRefreshToken returns the refresh token
func GetRefreshToken() string {
	return GetAuth().state.RefreshToken
}

// HasRole checks if user has a specific role
func HasRole(role string) bool {
	user := GetUser()
	if user == nil {
		return false
	}
	for _, r := range user.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if user has any of the specified roles
func HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if HasRole(role) {
			return true
		}
	}
	return false
}

// OnAuthChange subscribes to auth state changes
func OnAuthChange(fn func(AuthState)) func() {
	auth := GetAuth()
	auth.subscribers = append(auth.subscribers, fn)
	idx := len(auth.subscribers) - 1

	// Call immediately with current state
	fn(auth.state)

	return func() {
		auth.subscribers = append(
			auth.subscribers[:idx],
			auth.subscribers[idx+1:]...,
		)
	}
}

// SetToken updates just the token (e.g., after refresh)
func SetToken(token string) {
	auth := GetAuth()
	auth.state.Token = token
	auth.state.ExpiresAt = extractExpiry(token)
	auth.persist()
	auth.notify()
}

func (a *Auth) persist() {
	data, _ := json.Marshal(a.state)
	js.Global().Get("localStorage").Call("setItem", a.storageKey, string(data))
}

func (a *Auth) restore() {
	stored := js.Global().Get("localStorage").Call("getItem", a.storageKey)
	if stored.IsNull() || stored.IsUndefined() {
		return
	}

	var state AuthState
	if err := json.Unmarshal([]byte(stored.String()), &state); err != nil {
		return
	}

	a.state = state
}

func (a *Auth) notify() {
	for _, fn := range a.subscribers {
		fn(a.state)
	}
}

// extractExpiry attempts to extract expiry from JWT (without verification)
func extractExpiry(token string) time.Time {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Now().Add(24 * time.Hour) // Default to 24h
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Now().Add(24 * time.Hour)
	}

	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return time.Now().Add(24 * time.Hour)
	}

	if claims.Exp == 0 {
		return time.Now().Add(24 * time.Hour)
	}

	return time.Unix(claims.Exp, 0)
}

// AuthHeader returns the Authorization header value
func AuthHeader() string {
	token := GetToken()
	if token == "" {
		return ""
	}
	return "Bearer " + token
}
