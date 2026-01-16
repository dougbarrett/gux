# Server Utilities

Gux provides server-side utilities for building Go HTTP backends that complement the WASM frontend.

## Middleware

### Middleware Pattern

All middleware follows the standard pattern:

```go
type Middleware func(http.Handler) http.Handler
```

### Composing Middleware

```go
import "github.com/dougbarrett/gux/server"

// Chain multiple middleware
handler := server.Chain(
    server.Logger(),
    server.CORS(server.CORSOptions{}),
    server.Recover(),
    server.RequestID(),
)(yourHandler)
```

### Logger

Logs HTTP requests with method, path, status, and duration:

```go
handler := server.Logger()(yourHandler)

// Output:
// 2024/01/15 10:30:45 GET /api/posts 200 15.234ms
// 2024/01/15 10:30:46 POST /api/posts 201 23.456ms
```

### CORS

Cross-Origin Resource Sharing for API access from browsers:

```go
// Default options (permissive)
handler := server.CORS(server.CORSOptions{})(yourHandler)

// Custom options
handler := server.CORS(server.CORSOptions{
    AllowOrigin:  "https://myapp.com",
    AllowMethods: "GET, POST, PUT, DELETE",
    AllowHeaders: "Content-Type, Authorization, X-Request-ID",
})(yourHandler)
```

**Default values:**
- `AllowOrigin`: `"*"`
- `AllowMethods`: `"GET, POST, PUT, DELETE, OPTIONS"`
- `AllowHeaders`: `"Content-Type, Authorization"`

### Recover

Catches panics and returns 500 Internal Server Error:

```go
handler := server.Recover()(yourHandler)

// If handler panics:
// - Logs the panic
// - Returns {"error": {"code": "internal_error", "message": "Internal server error"}}
// - Status code 500
```

### RequestID

Adds unique request ID to each request:

```go
handler := server.RequestID()(yourHandler)

// Adds X-Request-ID header to response
// Useful for tracing requests through logs
```

### Using with Generated Handlers

```go
// Create handler from generated code
postsHandler := api.NewPostsAPIHandler(postsService)

// Add middleware
postsHandler.Use(
    server.Logger(),
    server.CORS(server.CORSOptions{}),
    server.Recover(),
)

// Register routes
postsHandler.RegisterRoutes(mux)
```

### JWT Authentication

Validates JWT tokens and injects user claims into the request context:

```go
handler := server.JWT(server.JWTOptions{
    Secret: []byte("your-secret-key"),
})(yourHandler)
```

#### JWT Options

```go
server.JWTOptions{
    // Required: HMAC secret key for HS256 tokens
    Secret: []byte("your-256-bit-secret"),

    // Skip authentication for certain paths
    SkipPaths: []string{
        "/health",
        "/api/auth/login",
        "/api/auth/signup",
        "/api/public/*",  // Wildcard prefix match
    },

    // Skip authentication for certain methods (e.g., CORS preflight)
    SkipMethods: []string{"OPTIONS"},

    // Where to find the token (default: "header:Authorization")
    TokenLookup: "header:Authorization",  // or "query:token", "cookie:jwt"

    // Auth scheme in header (default: "Bearer")
    AuthScheme: "Bearer",

    // Require tokens to have an expiry claim
    RequireExpiry: true,

    // Custom error handling
    ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
        // Custom error response
    },

    // Called after successful authentication
    SuccessHandler: func(r *http.Request, claims *server.Claims) {
        log.Printf("User %s authenticated", claims.UserID)
    },
}
```

#### Claims Structure

The JWT middleware extracts these claims:

```go
type Claims struct {
    // Standard JWT claims
    Subject   string   // "sub"
    Issuer    string   // "iss"
    Audience  string   // "aud"
    ExpiresAt int64    // "exp"
    IssuedAt  int64    // "iat"
    NotBefore int64    // "nbf"
    JWTID     string   // "jti"

    // Common custom claims
    UserID   string   // "user_id"
    Email    string   // "email"
    Name     string   // "name"
    Roles    []string // "roles"
    OrgID    string   // "org_id"
    TenantID string   // "tenant_id"

    // Access any claim
    Raw map[string]any
}
```

#### Accessing Claims in Handlers

```go
func (s *Service) GetProfile(ctx context.Context) (*User, error) {
    // Get full claims
    claims := server.GetClaims(ctx)
    if claims == nil {
        return nil, api.Unauthorized("not authenticated")
    }

    // Or use convenience functions
    userID := server.GetUserID(ctx)
    email := server.GetUserEmail(ctx)
    roles := server.GetUserRoles(ctx)
    orgID := server.GetOrgID(ctx)

    return s.db.GetUser(userID)
}
```

#### Role-Based Access Control

```go
// Require specific roles for a handler
adminHandler := server.Chain(
    server.JWT(jwtOpts),
    server.RequireRoles("admin"),
)(adminOnlyHandler)

// Check roles in service methods
func (s *Service) DeleteUser(ctx context.Context, id int) error {
    claims := server.GetClaims(ctx)
    if !claims.HasRole("admin") {
        return api.Forbidden("admin access required")
    }
    return s.db.DeleteUser(id)
}

// Check any of multiple roles
if claims.HasAnyRole("admin", "moderator") {
    // Allow moderation actions
}
```

#### Complete JWT Example

```go
func main() {
    mux := http.NewServeMux()

    // JWT configuration
    jwtOpts := server.JWTOptions{
        Secret: []byte(os.Getenv("JWT_SECRET")),
        SkipPaths: []string{
            "/health",
            "/api/auth/login",
            "/api/auth/signup",
            "/api/auth/refresh",
        },
        SkipMethods: []string{"OPTIONS"},
        RequireExpiry: true,
    }

    // Public auth endpoints (no JWT required)
    authHandler := api.NewAuthAPIHandler(authService)
    authHandler.Use(
        server.Logger(),
        server.CORS(server.CORSOptions{}),
    )
    authHandler.RegisterRoutes(mux)

    // Protected API endpoints
    postsHandler := api.NewPostsAPIHandler(postsService)
    postsHandler.Use(
        server.Logger(),
        server.CORS(server.CORSOptions{}),
        server.JWT(jwtOpts),
        server.Recover(),
    )
    postsHandler.RegisterRoutes(mux)

    // Admin-only endpoints
    adminHandler := api.NewAdminAPIHandler(adminService)
    adminHandler.Use(
        server.Logger(),
        server.CORS(server.CORSOptions{}),
        server.JWT(jwtOpts),
        server.RequireRoles("admin"),
        server.Recover(),
    )
    adminHandler.RegisterRoutes(mux)

    http.ListenAndServe(":8080", mux)
}
```

#### Generating Tokens

For testing or simple use cases, use the built-in token generator:

```go
// Create claims
claims := server.NewClaims(
    "user-123",           // userID
    "user@example.com",   // email
    []string{"user"},     // roles
    24*time.Hour,         // expires in
)

// Add custom claims
claims.OrgID = "org-456"
claims.Name = "John Doe"

// Generate token
token, err := server.GenerateToken(claims, []byte("your-secret"))
```

For production, consider using a dedicated JWT library like `golang-jwt/jwt`.

## SPA Handler

Serves static files with fallback to `index.html` for client-side routing:

```go
spa := server.NewSPAHandler("./static")
mux.HandleFunc("/", spa.ServeHTTP)
```

### How It Works

1. Checks if requested file exists in static directory
2. If file exists, serves it with correct MIME type
3. If file doesn't exist, serves `index.html` (for SPA routing)

### Supported MIME Types

| Extension | MIME Type |
|-----------|-----------|
| `.html` | `text/html` |
| `.js` | `application/javascript` |
| `.wasm` | `application/wasm` |
| `.css` | `text/css` |
| `.json` | `application/json` |
| `.svg` | `image/svg+xml` |
| `.png` | `image/png` |
| `.ico` | `image/x-icon` |

### Example Setup

```go
func main() {
    mux := http.NewServeMux()

    // API routes (more specific, registered first)
    mux.HandleFunc("/api/", apiHandler)
    mux.HandleFunc("/ws/", wsHandler)

    // SPA handler (catch-all, registered last)
    spa := server.NewSPAHandler("./static")
    mux.HandleFunc("/", spa.ServeHTTP)

    http.ListenAndServe(":8080", mux)
}
```

### Directory Structure

```
static/
├── index.html      # Entry point (served for all non-file routes)
├── main.wasm       # Compiled WASM
├── wasm_exec.js    # Go WASM runtime
├── favicon.ico
└── assets/
    ├── styles.css
    └── logo.png
```

## Error Handling

### Error Types

```go
import "github.com/dougbarrett/gux/api"

// Not Found (404)
return nil, api.NotFound("resource not found")
return nil, api.NotFoundf("user %d not found", id)

// Bad Request (400)
return nil, api.BadRequest("invalid email format")
return nil, api.BadRequestf("field %s is required", fieldName)

// Unauthorized (401)
return nil, api.Unauthorized("invalid credentials")

// Forbidden (403)
return nil, api.Forbidden("access denied")

// Conflict (409)
return nil, api.Conflict("resource already exists")

// Internal Error (500)
return nil, api.InternalError("database connection failed")
return nil, api.InternalErrorf("failed to process: %v", err)
```

### Error Response Format

All errors return JSON:

```json
{
    "error": {
        "code": "not_found",
        "message": "user 123 not found"
    }
}
```

### Writing Errors Manually

```go
func handler(w http.ResponseWriter, r *http.Request) {
    user, err := getUser(id)
    if err != nil {
        api.WriteError(w, err)
        return
    }
    // ...
}
```

### Custom Error Handling

```go
type Error struct {
    Status  int    `json:"-"`           // HTTP status code
    Code    string `json:"code"`        // Machine-readable code
    Message string `json:"message"`     // Human-readable message
}

func (e *Error) Error() string {
    return e.Message
}
```

## Query Utilities

### Query Parameters

```go
func handler(w http.ResponseWriter, r *http.Request) {
    q := api.Query(r)

    // String with default
    search := q.String("search", "")

    // Integer with default
    limit := q.Int("limit", 10)

    // Boolean with default
    active := q.Bool("active", true)
}
```

### Pagination

```go
func handler(w http.ResponseWriter, r *http.Request) {
    q := api.Query(r)
    page := q.Pagination()

    // page.Page     - Current page (1-indexed)
    // page.PerPage  - Items per page (default 20)
    // page.Offset   - Calculated offset for database

    // Use with database
    items := db.Query().
        Offset(page.Offset).
        Limit(page.PerPage).
        Find()

    total := db.Query().Count()

    // Return paginated response
    result := api.NewPaginatedResult(items, page, total)
    json.NewEncoder(w).Encode(result)
}
```

### Paginated Response Format

```json
{
    "data": [...],
    "page": 1,
    "per_page": 20,
    "total": 100,
    "total_pages": 5
}
```

### Default Pagination

```go
// Default values
var DefaultPagination = Pagination{
    Page:    1,
    PerPage: 20,
}

// Query params: ?page=2&per_page=50
```

## Complete Server Example

```go
package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"

    "yourapp/api"
    "github.com/dougbarrett/gux/server"
)

func main() {
    port := flag.Int("port", 8080, "Server port")
    staticDir := flag.String("static", "./static", "Static files directory")
    flag.Parse()

    mux := http.NewServeMux()

    // Create services
    postsService := NewPostsService()
    usersService := NewUsersService()

    // Create and configure handlers
    postsHandler := api.NewPostsAPIHandler(postsService)
    postsHandler.Use(
        server.Logger(),
        server.CORS(server.CORSOptions{}),
        server.Recover(),
    )
    postsHandler.RegisterRoutes(mux)

    usersHandler := api.NewUsersAPIHandler(usersService)
    usersHandler.Use(
        server.Logger(),
        server.CORS(server.CORSOptions{}),
        server.Recover(),
    )
    usersHandler.RegisterRoutes(mux)

    // WebSocket handler
    wsHandler := NewWebSocketHandler(postsService)
    mux.HandleFunc("/ws/posts", wsHandler.ServeHTTP)

    // Wire up real-time events
    postsService.SetEventCallbacks(
        func(p api.Post) { wsHandler.Broadcast("post.created", p) },
        func(p api.Post) { wsHandler.Broadcast("post.updated", p) },
        func(id int) { wsHandler.Broadcast("post.deleted", id) },
    )

    // Health check endpoint
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    // SPA handler for static files (catch-all)
    spa := server.NewSPAHandler(*staticDir)
    mux.HandleFunc("/", spa.ServeHTTP)

    // Start server
    addr := fmt.Sprintf(":%d", *port)
    fmt.Printf("Server running at http://localhost%s\n", addr)
    fmt.Printf("Static files from: %s\n", *staticDir)
    fmt.Println("\nEndpoints:")
    fmt.Println("  GET    /api/posts      - List posts")
    fmt.Println("  POST   /api/posts      - Create post")
    fmt.Println("  GET    /api/posts/:id  - Get post")
    fmt.Println("  PUT    /api/posts/:id  - Update post")
    fmt.Println("  DELETE /api/posts/:id  - Delete post")
    fmt.Println("  WS     /ws/posts       - Real-time updates")
    fmt.Println("  GET    /health         - Health check")

    log.Fatal(http.ListenAndServe(addr, mux))
}
```

## Best Practices

### 1. Order Middleware Correctly

```go
// Recommended order
handler := server.Chain(
    server.RequestID(),  // First: adds ID for tracing
    server.Logger(),     // Second: logs with request ID
    server.Recover(),    // Third: catches panics from below
    server.CORS(opts),   // Fourth: handles CORS preflight
)(yourHandler)
```

### 2. Use Structured Errors

```go
// Good: Structured error
return nil, api.NotFoundf("post %d not found", id)

// Avoid: Generic error
return nil, errors.New("not found")
```

### 3. Register Specific Routes First

```go
mux := http.NewServeMux()

// More specific routes first
mux.HandleFunc("/api/posts", postsHandler)
mux.HandleFunc("/api/users", usersHandler)
mux.HandleFunc("/ws/", wsHandler)

// Catch-all last
mux.HandleFunc("/", spaHandler)
```

### 4. Validate Input Early

```go
func (s *Service) Create(ctx context.Context, req CreateRequest) (*Item, error) {
    // Validate first
    if req.Title == "" {
        return nil, api.BadRequest("title is required")
    }
    if len(req.Title) > 100 {
        return nil, api.BadRequest("title too long (max 100 chars)")
    }

    // Then process
    // ...
}
```

### 5. Use Context for Cancellation

```go
func (s *Service) GetAll(ctx context.Context) ([]Item, error) {
    // Check context before expensive operations
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Database query with context
    items, err := s.db.QueryContext(ctx, "SELECT * FROM items")
    // ...
}
```
