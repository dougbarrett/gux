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
│   ├── html_renderer.go    # Server-side rendering
│   ├── dom_renderer.go     # WASM DOM rendering
│   └── router_*.go         # Platform-specific routing
├── components/             # Pre-built UI components (Gux library)
├── state/                  # State management utilities
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

## Development Notes

- **core/** is the low-level universal rendering system (focus of current development)
- **components/** contains the pre-built Gux component library (separate from core)
- Always use `core.OnLoad()` for data fetching to support hydration
- DTOs should exclude sensitive fields (passwords, internal IDs)
- Use `core.External: true` in Attrs for links that cross bundle boundaries
