# Gux Framework

Gux is a full-stack Go framework for building modern web applications with WebAssembly.

## Architecture

### Core Framework (`core/`)

The core package provides a universal rendering system that works identically on server (SSR) and client (WASM):

- **Node System** (`node.go`) - Abstract UI tree representation
- **Elements** (`elements.go`) - HTML element helpers (Div, Button, Input, etc.)
- **Renderers** - HTML (`html_renderer.go`) for server, DOM (`dom_renderer.go`) for WASM
- **App & Routing** (`app.go`, `router_server.go`, `router_wasm.go`) - Hybrid SSR+WASM routing
- **CRUD** (`crud.go`) - Automatic REST API generation with DTOs and hooks
- **CSRF** (`csrf.go`) - Automatic CSRF protection for mutating operations

### Key Patterns

**Page Functions**: Return a loader that returns a component:
```go
func MyPage(r *core.Router) func() core.Node {
    // Loader - runs on server, fetches data
    var data []Model
    r.OnLoad(func() {
        api.Models.List(func(result []Model, err error) {
            data = result
        })
    })

    // Component - reactive rendering
    return func() core.Node {
        return core.Div(core.Class("..."),
            core.Text("Hello"),
        )
    }
}
```

**State Management**:
```go
count := r.StateInt("count", 0)       // int state
name := r.StateString("name", "")     // string state
active := r.StateBool("active", false) // bool state
user := core.UseState(r, "user", User{}) // any type

// Read and update
current := count.Get()
count.Set(current + 1)
```

**CRUD with DTOs**:
```go
app.CRUD(models.User{},
    core.WithListDTO(dto.UserList{}),
    core.WithDetailDTO(dto.UserDetail{}, "Posts"), // with preload
    core.WithCreateHook(func(data map[string]interface{}) (interface{}, error) {
        // Custom creation logic (e.g., password hashing)
    }),
)
```

## Project Structure

```
goquery/
├── core/                    # Universal rendering framework
│   ├── node.go             # Node interface & types
│   ├── elements.go         # Element helpers
│   ├── app.go              # App, Router, State
│   ├── crud.go             # CRUD API generation
│   ├── csrf.go             # CSRF protection
│   ├── html_renderer.go    # Server-side rendering
│   ├── dom_renderer.go     # WASM DOM rendering
│   └── router_*.go         # Platform-specific routing
├── fetch/                   # WASM HTTP client with auto CSRF
├── api/                    # API utilities & error types
├── server/                 # Server middleware
├── cmd/gux/               # CLI tool
└── examples/
    └── minimal/           # Reference implementation
        ├── app.go         # App setup, routes, CRUD registration
        ├── models/        # GORM models
        ├── dto/           # Data transfer objects
        ├── pages/         # Public pages
        └── admin/         # Admin panel pages
```

## Commands

```bash
gux dev           # Build and run with hot reload
gux build         # Production build
gux gen           # Generate API client/server code
```

## Examples/Minimal

The minimal example demonstrates:

- **Hybrid rendering**: SSR with WASM hydration
- **Route groups with bundles**: Public and admin routes use separate WASM files
- **CRUD with DTOs**: Users and Posts with relationship preloading
- **Security hooks**: Password hashing in create/update hooks
- **State hydration**: Server state transferred to client

### Models (`models/`)

```go
type User struct {
    gorm.Model
    Email        string `gorm:"uniqueIndex"`
    Name         string
    PasswordHash string // Hidden from API via DTOs
    Role         string
    Posts        []Post `gorm:"foreignKey:UserID"`
}

type Post struct {
    gorm.Model
    Title   string
    Content string
    UserID  uint
    User    User
}
```

### DTOs (`dto/`)

DTOs control what data is exposed via API:

```go
// UserList excludes PasswordHash
type UserList struct {
    ID    uint   `json:"id" gux:"User.ID"`
    Email string `json:"email" gux:"User.Email"`
    Name  string `json:"name" gux:"User.Name"`
    Role  string `json:"role" gux:"User.Role"`
}

// UserDetail includes related posts
type UserDetail struct {
    ID    uint        `json:"id"`
    Email string      `json:"email"`
    Name  string      `json:"name"`
    Posts []PostBrief `json:"posts" gux:"User.Posts" preload:"Posts"`
}
```

### Route Setup

```go
// Public routes (default bundle)
app.Routes().
    Hybrid("/", pages.Home).
    Hybrid("/about", pages.About)

// Admin routes (separate "admin" bundle)
app.RouteGroup("/admin", core.WithBundle("admin")).
    Hybrid("/", admin.Dashboard).
    Hybrid("/users", admin.Users).
    Hybrid("/users/:id", admin.UserDetail)
```

## CSRF Protection

CSRF protection is enabled by default. It uses the Double Submit Cookie pattern:

1. Server generates token, embeds in `<meta name="csrf-token">` and sets `__gux_csrf` cookie
2. `fetch` package reads token and includes `X-CSRF-Token` header on POST/PUT/PATCH/DELETE
3. Server middleware validates header matches cookie

**Configuration**:
```go
app := core.New()                     // CSRF enabled by default
app.DisableCSRF()                     // Disable for API-key authenticated endpoints
app.EnableCSRF(core.CSRFConfig{...})  // Custom configuration
```

**Transparency**: Developers don't need to handle CSRF manually - it's automatic.

## Testing Examples

**IMPORTANT**: When testing or running example apps, ALWAYS use `gux dev`:

```bash
cd examples/marketing
gux dev   # Builds WASM, generates Tailwind CSS, and runs with hot reload
```

**DO NOT** use `go build` or `go run` directly - these will not generate the required assets (WASM, CSS) and the app will not render correctly.

If something appears broken (missing styles, assets not loading):
```bash
gux clean   # Removes .gux/ directory and generated files
gux dev     # Fresh build with all assets
```

**Example Ports**:
- Auth: 8082
- Marketing: 8083
- SaaS: 8082 (run separately from Auth)
- Admin: 8084

## Development Notes

- **core/** is the low-level universal rendering system (focus of current development)
- Always use `core.OnLoad()` for data fetching to support hydration
- DTOs should exclude sensitive fields (passwords, internal IDs)
- Use `core.External: true` in Attrs for links that cross bundle boundaries
- CSRF protection is automatic for all CRUD operations
