//go:build !js || !wasm

package core

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	// CSRFTokenLength is the length of the random token in bytes (32 bytes = 256 bits)
	CSRFTokenLength = 32

	// CSRFCookieName is the name of the CSRF cookie
	CSRFCookieName = "__gux_csrf"

	// CSRFHeaderName is the header name for CSRF token validation
	CSRFHeaderName = "X-CSRF-Token"

	// CSRFMetaName is the meta tag name in HTML for client-side access
	CSRFMetaName = "csrf-token"
)

// CSRFConfig holds configuration for CSRF protection
type CSRFConfig struct {
	// Enabled controls whether CSRF protection is active (default: true)
	Enabled bool

	// CookiePath sets the path for the CSRF cookie (default: "/")
	CookiePath string

	// CookieMaxAge sets the max age in seconds (default: 12 hours)
	CookieMaxAge int

	// CookieSameSite sets the SameSite attribute (default: Strict)
	CookieSameSite http.SameSite

	// Secure sets the Secure flag (default: true in production)
	Secure bool

	// ExemptPaths are paths that skip CSRF validation
	ExemptPaths []string
}

// DefaultCSRFConfig returns the default CSRF configuration
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		Enabled:        true,
		CookiePath:     "/",
		CookieMaxAge:   43200, // 12 hours
		CookieSameSite: http.SameSiteStrictMode,
		Secure:         false, // Set to true in production
		ExemptPaths:    nil,
	}
}

// csrfToken generates a cryptographically secure random token
func generateCSRFToken() (string, error) {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// csrfTokenStore provides thread-safe token storage per request
// Uses a simple in-memory store with expiration
type csrfTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
}

var tokenStore = &csrfTokenStore{
	tokens: make(map[string]time.Time),
}

// Add stores a token with expiration
func (s *csrfTokenStore) Add(token string, expiry time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = time.Now().Add(expiry)
}

// Valid checks if a token exists and is not expired
func (s *csrfTokenStore) Valid(token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	expiry, exists := s.tokens[token]
	if !exists {
		return false
	}
	return time.Now().Before(expiry)
}

// Cleanup removes expired tokens (call periodically)
func (s *csrfTokenStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for token, expiry := range s.tokens {
		if now.After(expiry) {
			delete(s.tokens, token)
		}
	}
}

// setCSRFCookie sets the CSRF token cookie on the response
func setCSRFCookie(w http.ResponseWriter, token string, config CSRFConfig) {
	cookie := &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     config.CookiePath,
		MaxAge:   config.CookieMaxAge,
		HttpOnly: false, // Must be readable by JavaScript/WASM
		Secure:   config.Secure,
		SameSite: config.CookieSameSite,
	}
	http.SetCookie(w, cookie)
}

// getCSRFFromCookie retrieves the CSRF token from the request cookie
func getCSRFFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(CSRFCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// getCSRFFromHeader retrieves the CSRF token from the request header
func getCSRFFromHeader(r *http.Request) string {
	return r.Header.Get(CSRFHeaderName)
}

// validateCSRF checks if the CSRF token in the header matches the cookie
func validateCSRF(r *http.Request) bool {
	cookieToken := getCSRFFromCookie(r)
	headerToken := getCSRFFromHeader(r)

	if cookieToken == "" || headerToken == "" {
		return false
	}

	// Constant-time comparison to prevent timing attacks
	if len(cookieToken) != len(headerToken) {
		return false
	}

	result := 0
	for i := 0; i < len(cookieToken); i++ {
		result |= int(cookieToken[i] ^ headerToken[i])
	}
	return result == 0
}

// isMutatingMethod returns true for methods that modify data
func isMutatingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

// isExemptPath checks if the path is exempt from CSRF validation.
// Supports exact matches and prefix patterns ending with * (e.g., "/api/*").
func isExemptPath(path string, exemptPaths []string) bool {
	for _, exempt := range exemptPaths {
		if strings.HasSuffix(exempt, "*") {
			if strings.HasPrefix(path, strings.TrimSuffix(exempt, "*")) {
				return true
			}
		} else if path == exempt {
			return true
		}
	}
	return false
}

// CSRFMiddleware returns middleware that handles CSRF protection
func CSRFMiddleware(config CSRFConfig) func(http.Handler) http.Handler {
	// Start background cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			tokenStore.Cleanup()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Check if path is exempt
			if isExemptPath(r.URL.Path, config.ExemptPaths) {
				next.ServeHTTP(w, r)
				return
			}

			// For mutating methods, validate the token
			if isMutatingMethod(r.Method) {
				if !validateCSRF(r) {
					http.Error(w, "CSRF token invalid", http.StatusForbidden)
					return
				}
			}

			// Ensure a token exists (either from cookie or generate new)
			token := getCSRFFromCookie(r)
			if token == "" {
				var err error
				token, err = generateCSRFToken()
				if err != nil {
					http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
					return
				}
				tokenStore.Add(token, time.Duration(config.CookieMaxAge)*time.Second)
				setCSRFCookie(w, token, config)
			}

			next.ServeHTTP(w, r)
		})
	}
}
