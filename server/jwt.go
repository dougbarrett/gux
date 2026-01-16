package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dougbarrett/gux/api"
)

// Claims represents the JWT claims extracted from a token
type Claims struct {
	// Standard claims
	Subject   string `json:"sub"`
	Issuer    string `json:"iss"`
	Audience  string `json:"aud"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	NotBefore int64  `json:"nbf"`
	JWTID     string `json:"jti"`

	// Common custom claims
	UserID   string   `json:"user_id,omitempty"`
	Email    string   `json:"email,omitempty"`
	Name     string   `json:"name,omitempty"`
	Roles    []string `json:"roles,omitempty"`
	OrgID    string   `json:"org_id,omitempty"`
	TenantID string   `json:"tenant_id,omitempty"`

	// Raw claims for custom access
	Raw map[string]any `json:"-"`
}

// HasRole checks if the claims include a specific role
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the claims include any of the specified roles
func (c *Claims) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// IsExpired checks if the token has expired
func (c *Claims) IsExpired() bool {
	if c.ExpiresAt == 0 {
		return false
	}
	return time.Now().Unix() > c.ExpiresAt
}

// JWTOptions configures the JWT middleware
type JWTOptions struct {
	// Secret is the HMAC secret key for HS256 tokens (required)
	Secret []byte

	// SkipPaths are paths that don't require authentication
	// Supports exact matches and prefix matches with trailing *
	// Example: []string{"/health", "/api/public/*"}
	SkipPaths []string

	// SkipMethods are HTTP methods that don't require authentication
	// Example: []string{"OPTIONS"} for CORS preflight
	SkipMethods []string

	// TokenLookup defines where to find the token
	// Default: "header:Authorization"
	// Options: "header:<name>", "query:<name>", "cookie:<name>"
	TokenLookup string

	// AuthScheme is the scheme in Authorization header
	// Default: "Bearer"
	AuthScheme string

	// ErrorHandler is called when authentication fails
	// Default: writes JSON error response
	ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

	// SuccessHandler is called after successful authentication
	// Can be used for logging or custom context injection
	SuccessHandler func(r *http.Request, claims *Claims)

	// ClaimsContextKey is the context key for storing claims
	// Default: "jwt_claims"
	ClaimsContextKey any

	// RequireExpiry rejects tokens without an exp claim
	RequireExpiry bool
}

// contextKey is the type for context keys
type contextKey string

// Default context key for JWT claims
const defaultClaimsKey contextKey = "jwt_claims"

// JWT returns middleware that validates JWT tokens
func JWT(opts JWTOptions) Middleware {
	// Set defaults
	if opts.TokenLookup == "" {
		opts.TokenLookup = "header:Authorization"
	}
	if opts.AuthScheme == "" {
		opts.AuthScheme = "Bearer"
	}
	if opts.ClaimsContextKey == nil {
		opts.ClaimsContextKey = defaultClaimsKey
	}
	if opts.ErrorHandler == nil {
		opts.ErrorHandler = defaultJWTErrorHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path should be skipped
			if shouldSkipPath(r.URL.Path, opts.SkipPaths) {
				next.ServeHTTP(w, r)
				return
			}

			// Check if method should be skipped
			for _, method := range opts.SkipMethods {
				if r.Method == method {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract token
			token, err := extractToken(r, opts.TokenLookup, opts.AuthScheme)
			if err != nil {
				opts.ErrorHandler(w, r, err)
				return
			}

			// Validate and parse token
			claims, err := validateToken(token, opts.Secret)
			if err != nil {
				opts.ErrorHandler(w, r, err)
				return
			}

			// Check expiry
			if opts.RequireExpiry && claims.ExpiresAt == 0 {
				opts.ErrorHandler(w, r, api.Unauthorized("token missing expiry"))
				return
			}
			if claims.IsExpired() {
				opts.ErrorHandler(w, r, api.Unauthorized("token expired"))
				return
			}

			// Call success handler if provided
			if opts.SuccessHandler != nil {
				opts.SuccessHandler(r, claims)
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), opts.ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRoles returns middleware that requires specific roles
func RequireRoles(roles ...string) Middleware {
	return RequireRolesWithKey(defaultClaimsKey, roles...)
}

// RequireRolesWithKey returns middleware that requires specific roles using a custom context key
func RequireRolesWithKey(claimsKey any, roles ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaimsWithKey(r.Context(), claimsKey)
			if claims == nil {
				api.WriteError(w, api.Unauthorized("authentication required"))
				return
			}

			if !claims.HasAnyRole(roles...) {
				api.WriteError(w, api.Forbidden("insufficient permissions"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetClaims retrieves JWT claims from the request context
func GetClaims(ctx context.Context) *Claims {
	return GetClaimsWithKey(ctx, defaultClaimsKey)
}

// GetClaimsWithKey retrieves JWT claims using a custom context key
func GetClaimsWithKey(ctx context.Context, key any) *Claims {
	claims, ok := ctx.Value(key).(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// GetUserID retrieves the user ID from context claims
func GetUserID(ctx context.Context) string {
	claims := GetClaims(ctx)
	if claims == nil {
		return ""
	}
	// Try UserID first, fall back to Subject
	if claims.UserID != "" {
		return claims.UserID
	}
	return claims.Subject
}

// GetUserEmail retrieves the user email from context claims
func GetUserEmail(ctx context.Context) string {
	claims := GetClaims(ctx)
	if claims == nil {
		return ""
	}
	return claims.Email
}

// GetUserRoles retrieves the user roles from context claims
func GetUserRoles(ctx context.Context) []string {
	claims := GetClaims(ctx)
	if claims == nil {
		return nil
	}
	return claims.Roles
}

// GetOrgID retrieves the organization ID from context claims
func GetOrgID(ctx context.Context) string {
	claims := GetClaims(ctx)
	if claims == nil {
		return ""
	}
	return claims.OrgID
}

// defaultJWTErrorHandler writes a JSON error response
func defaultJWTErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	api.WriteError(w, err)
}

// shouldSkipPath checks if a path matches any skip pattern
func shouldSkipPath(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if pattern == path {
			return true
		}
		// Prefix match with trailing *
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}
	return false
}

// extractToken extracts the JWT token from the request
func extractToken(r *http.Request, lookup, scheme string) (string, error) {
	parts := strings.SplitN(lookup, ":", 2)
	if len(parts) != 2 {
		return "", api.InternalError("invalid token lookup configuration")
	}

	source, name := parts[0], parts[1]

	switch source {
	case "header":
		return extractFromHeader(r, name, scheme)
	case "query":
		return extractFromQuery(r, name)
	case "cookie":
		return extractFromCookie(r, name)
	default:
		return "", api.InternalError("unsupported token lookup source: " + source)
	}
}

func extractFromHeader(r *http.Request, name, scheme string) (string, error) {
	header := r.Header.Get(name)
	if header == "" {
		return "", api.Unauthorized("missing authorization header")
	}

	// Check scheme
	if scheme != "" {
		prefix := scheme + " "
		if !strings.HasPrefix(header, prefix) {
			return "", api.Unauthorized("invalid authorization scheme")
		}
		return strings.TrimPrefix(header, prefix), nil
	}

	return header, nil
}

func extractFromQuery(r *http.Request, name string) (string, error) {
	token := r.URL.Query().Get(name)
	if token == "" {
		return "", api.Unauthorized("missing token query parameter")
	}
	return token, nil
}

func extractFromCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", api.Unauthorized("missing token cookie")
	}
	return cookie.Value, nil
}

// validateToken validates a JWT token and returns the claims
func validateToken(tokenString string, secret []byte) (*Claims, error) {
	// Split token into parts
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, api.Unauthorized("invalid token format")
	}

	headerB64, payloadB64, signatureB64 := parts[0], parts[1], parts[2]

	// Verify signature
	message := headerB64 + "." + payloadB64
	expectedSig, err := computeHS256Signature(message, secret)
	if err != nil {
		return nil, api.Unauthorized("invalid token")
	}

	// Decode provided signature
	providedSig, err := base64.RawURLEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, api.Unauthorized("invalid token signature encoding")
	}

	// Compare signatures using constant-time comparison
	if !hmac.Equal(providedSig, expectedSig) {
		return nil, api.Unauthorized("invalid token signature")
	}

	// Decode header to verify algorithm
	headerJSON, err := base64.RawURLEncoding.DecodeString(headerB64)
	if err != nil {
		return nil, api.Unauthorized("invalid token header encoding")
	}

	var header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, api.Unauthorized("invalid token header")
	}

	if header.Alg != "HS256" {
		return nil, api.Unauthorized("unsupported token algorithm: " + header.Alg)
	}

	// Decode payload
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, api.Unauthorized("invalid token payload encoding")
	}

	// Parse into Claims struct
	var claims Claims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, api.Unauthorized("invalid token payload")
	}

	// Also store raw claims for custom access
	var raw map[string]any
	if err := json.Unmarshal(payloadJSON, &raw); err == nil {
		claims.Raw = raw
	}

	return &claims, nil
}

// computeHS256Signature computes HMAC-SHA256 signature
func computeHS256Signature(message string, secret []byte) ([]byte, error) {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(message))
	return h.Sum(nil), nil
}

// GenerateToken creates a new JWT token with the given claims
// This is a convenience function for testing and simple use cases
// For production, consider using a dedicated JWT library
func GenerateToken(claims *Claims, secret []byte) (string, error) {
	// Create header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Create payload
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature
	message := headerB64 + "." + payloadB64
	sig, err := computeHS256Signature(message, secret)
	if err != nil {
		return "", err
	}
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	return message + "." + sigB64, nil
}

// NewClaims creates a new Claims struct with common defaults
func NewClaims(userID, email string, roles []string, expiresIn time.Duration) *Claims {
	now := time.Now()
	return &Claims{
		Subject:   userID,
		UserID:    userID,
		Email:     email,
		Roles:     roles,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(expiresIn).Unix(),
	}
}
