# Authentication

Gux provides a client-side authentication module for managing user sessions, JWT tokens, and role-based access control in WASM applications.

## Overview

The auth package handles:
- User session management with automatic persistence to localStorage
- JWT token storage and expiry tracking
- Refresh token support
- Role-based access control
- Reactive state changes via subscriptions

## Getting Started

```go
import "github.com/dougbarrett/gux/auth"

// Initialize auth (typically at app startup)
auth.Init()
```

## User Type

The `User` struct represents an authenticated user:

```go
type User struct {
    ID       string         `json:"id"`
    Email    string         `json:"email"`
    Name     string         `json:"name"`
    Roles    []string       `json:"roles"`
    Metadata map[string]any `json:"metadata,omitempty"`
}
```

## Authentication State

The `AuthState` struct holds the complete authentication state:

```go
type AuthState struct {
    User         *User
    Token        string
    RefreshToken string
    ExpiresAt    time.Time
}
```

## Login and Logout

### Basic Login

```go
// After successful authentication from your backend
user := &auth.User{
    ID:    "user-123",
    Email: "user@example.com",
    Name:  "John Doe",
    Roles: []string{"user", "admin"},
}

auth.Login(jwtToken, user)
```

### Login with Refresh Token

```go
// When your backend provides both access and refresh tokens
auth.LoginWithTokens(accessToken, refreshToken, user)
```

### Logout

```go
// Clears all auth state and removes from localStorage
auth.Logout()
```

## Checking Authentication Status

```go
// Check if user is logged in
if auth.IsAuthenticated() {
    // User has valid session
}

// Check if token hasn't expired
if auth.IsTokenValid() {
    // Token is still valid
}
```

## Accessing User Data

```go
// Get current user
user := auth.GetUser()
if user != nil {
    fmt.Println("Welcome,", user.Name)
}

// Get JWT token (for API calls)
token := auth.GetToken()

// Get refresh token
refreshToken := auth.GetRefreshToken()

// Get formatted Authorization header
header := auth.AuthHeader() // Returns "Bearer <token>"
```

## Role-Based Access Control

```go
// Check single role
if auth.HasRole("admin") {
    // Show admin panel
}

// Check any of multiple roles
if auth.HasAnyRole("admin", "moderator") {
    // Show moderation tools
}
```

### Example: Protected UI Component

```go
func AdminPanel() js.Value {
    if !auth.HasRole("admin") {
        return components.Alert(components.AlertProps{
            Variant: components.AlertWarning,
            Message: "Access denied. Admin privileges required.",
        })
    }

    return components.Div("admin-panel",
        components.H2("Admin Dashboard"),
        // Admin content...
    )
}
```

## Subscribing to Auth Changes

React to authentication state changes in your UI:

```go
// Subscribe returns an unsubscribe function
unsubscribe := auth.OnAuthChange(func(state auth.AuthState) {
    if state.User != nil {
        showUserMenu(state.User)
    } else {
        showLoginButton()
    }
})

// Clean up when component unmounts
defer unsubscribe()
```

## Token Management

### Updating Token After Refresh

```go
// When you refresh the access token from your backend
auth.SetToken(newAccessToken)
```

### JWT Expiry Extraction

The auth package automatically extracts the `exp` claim from JWT tokens to track expiry. If the token doesn't contain an expiry or is malformed, it defaults to 24 hours.

> **Note:** Token expiry is extracted without cryptographic verification. Your backend should always validate tokens server-side.

## Integration with API Calls

### Using with fetch

```go
func fetchProtectedResource(url string) (js.Value, error) {
    if !auth.IsTokenValid() {
        return js.Null(), errors.New("not authenticated")
    }

    headers := map[string]string{
        "Authorization": auth.AuthHeader(),
        "Content-Type":  "application/json",
    }

    return api.Fetch(url, api.WithHeaders(headers))
}
```

### Using with Generated API Client

```go
// Configure API client with auth header
client := api.NewClient(
    api.WithBaseURL("https://api.example.com"),
    api.WithHeader("Authorization", auth.AuthHeader()),
)
```

### Auto-Refresh Pattern

```go
func authenticatedRequest(url string) ([]byte, error) {
    // Check if token is expired
    if !auth.IsTokenValid() && auth.GetRefreshToken() != "" {
        // Attempt to refresh
        newToken, err := refreshAccessToken(auth.GetRefreshToken())
        if err != nil {
            auth.Logout()
            return nil, errors.New("session expired")
        }
        auth.SetToken(newToken)
    }

    if !auth.IsAuthenticated() {
        return nil, errors.New("not authenticated")
    }

    // Make request with fresh token
    return api.Get(url, api.WithHeader("Authorization", auth.AuthHeader()))
}
```

## Persistence

Auth state is automatically persisted to localStorage under the key `auth_state`. On page reload, the auth module restores the previous session:

```go
// On app startup, state is automatically restored
auth.Init()

// Check if we have a restored session
if auth.IsAuthenticated() {
    // User was previously logged in
    if auth.IsTokenValid() {
        // Token is still valid
    } else {
        // Token expired, may need to refresh or re-login
    }
}
```

## Complete Example

```go
package main

import (
    "github.com/dougbarrett/gux/auth"
    "github.com/dougbarrett/gux/components"
    "syscall/js"
)

func main() {
    // Initialize auth
    auth.Init()

    // Subscribe to auth changes
    auth.OnAuthChange(func(state auth.AuthState) {
        renderApp(state)
    })

    // Keep the Go runtime alive
    select {}
}

func renderApp(state auth.AuthState) {
    app := js.Global().Get("document").Call("getElementById", "app")
    app.Set("innerHTML", "")

    if state.User != nil {
        app.Call("appendChild", renderLoggedIn(state.User))
    } else {
        app.Call("appendChild", renderLoginForm())
    }
}

func renderLoggedIn(user *auth.User) js.Value {
    return components.Div("p-4",
        components.H1("Welcome, "+user.Name),
        components.Text("Email: "+user.Email),
        components.Button(components.ButtonProps{
            Text:    "Logout",
            OnClick: func(e js.Value) { auth.Logout() },
        }),
    )
}

func renderLoginForm() js.Value {
    return components.Div("p-4",
        components.H1("Please Log In"),
        components.Button(components.ButtonProps{
            Text:    "Login",
            OnClick: handleLogin,
        }),
    )
}

func handleLogin(e js.Value) {
    // In a real app, this would call your auth API
    user := &auth.User{
        ID:    "123",
        Email: "user@example.com",
        Name:  "Demo User",
        Roles: []string{"user"},
    }
    auth.Login("eyJhbGciOiJIUzI1NiIs...", user)
}
```

## Best Practices

### 1. Initialize Early

```go
func main() {
    // Initialize auth before rendering UI
    auth.Init()

    // Now render app
    renderApp()
}
```

### 2. Always Check Authentication Before Protected Actions

```go
func deleteResource(id string) error {
    if !auth.IsAuthenticated() {
        return errors.New("must be logged in")
    }
    if !auth.HasRole("admin") {
        return errors.New("admin access required")
    }
    // Proceed with deletion
}
```

### 3. Handle Token Expiry Gracefully

```go
// Wrap API calls with expiry checking
func secureAPICall(fn func() error) error {
    if !auth.IsTokenValid() {
        // Attempt refresh or redirect to login
        return errors.New("session expired")
    }
    return fn()
}
```

### 4. Clean Up Subscriptions

```go
var cleanupFns []func()

func mountComponent() {
    unsub := auth.OnAuthChange(func(s auth.AuthState) {
        // Handle state change
    })
    cleanupFns = append(cleanupFns, unsub)
}

func unmountComponent() {
    for _, fn := range cleanupFns {
        fn()
    }
    cleanupFns = nil
}
```

### 5. Don't Store Sensitive Data in Metadata

```go
// Good: Store user preferences
user := &auth.User{
    Metadata: map[string]any{
        "theme":    "dark",
        "language": "en",
    },
}

// Bad: Don't store sensitive info client-side
user := &auth.User{
    Metadata: map[string]any{
        "ssn": "123-45-6789", // Never do this!
    },
}
```

### 6. Use Role-Based Checks for UI, Server-Side for Security

```go
// Client-side role checks are for UX (hiding/showing UI)
if auth.HasRole("admin") {
    showAdminButton()
}

// Always enforce permissions server-side too!
// The client can be manipulated - never trust it for security
```
