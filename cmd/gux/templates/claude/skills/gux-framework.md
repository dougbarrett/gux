# Gux Framework Development Guide

You are an expert in **Gux**, a full-stack Go framework for building modern web applications with WebAssembly. This skill provides comprehensive knowledge for developing Gux applications.

## Documentation Reference

**IMPORTANT**: If you have any questions about Gux functionality or patterns that aren't fully covered in this skill, always fetch the official documentation:

**Docs Site**: https://dougbarrett.github.io/gux/

Key documentation pages:
- [Getting Started](https://dougbarrett.github.io/gux/#/getting-started)
- [API Generation](https://dougbarrett.github.io/gux/#/api-generation)
- [Templates](https://dougbarrett.github.io/gux/#/templates)
- [Server Utilities](https://dougbarrett.github.io/gux/#/server)
- [Deployment](https://dougbarrett.github.io/gux/#/deployment)

Use your web fetching capabilities to read these pages before answering questions or implementing features.

## Framework Overview

Gux enables writing entire web applications in Go:
- **Frontend**: Compiles to WebAssembly, runs natively in the browser
- **Backend**: Standard Go HTTP server with generated handlers
- **API**: Type-safe clients and servers generated from Go interfaces
- **Core**: Universal rendering system for SSR + WASM hydration

---

## Core Framework (`core/` package)

The `core` package provides the low-level universal rendering system. Components built with `core` work identically on server (SSR) and client (WASM).

### Architecture

The key insight: components return **Nodes**, not DOM elements or HTML strings. Nodes are an intermediate representation rendered to either target.

```
Component -> Node Tree -> Renderer -> HTML (server) or DOM (client)
```

### Node System

```go
// Node is the universal currency - components produce Nodes
type Node interface {
    Render(r Renderer) RenderResult
}

// Available node types:
TextNode{Content: "Hello"}
Element{Tag: "div", Attrs: Attrs{}, Children: []Node{}}
Fragment{Children: []Node{}}
```

### Element Helpers

```go
import "github.com/dougbarrett/gux/core"

// Create elements with attributes and children
core.Div(core.Class("flex gap-4"),
    core.H1(core.Attrs{Class: "text-2xl"}, core.Text("Title")),
    core.P(core.Attrs{}, core.Text("Content")),
)

// Available helpers:
core.Div, core.Span, core.P, core.H1, core.H2, core.H3
core.A, core.Button, core.Input, core.Form, core.Label
core.Ul, core.Li, core.Nav, core.Header, core.Footer
core.Main, core.Section, core.Article, core.Img
core.Table, core.Thead, core.Tbody, core.Tr, core.Th, core.Td
core.Select, core.Option, core.Textarea

// Generic element
core.El("custom-tag", core.Attrs{}, children...)

// Fragment (multiple nodes without wrapper)
core.Frag(node1, node2, node3)

// Conditional rendering
core.If(condition, node)

// Shorthand for class-only attrs
core.Class("my-class") // equivalent to core.Attrs{Class: "my-class"}
```

### Attrs (Attributes)

```go
core.Attrs{
    ID:    "my-id",
    Class: "flex items-center",
    Href:  "/path",           // For <a> elements
    Src:   "/image.png",      // For <img> elements
    Alt:   "Description",
    Type:  "text",            // For <input> elements
    Name:  "field-name",
    Value: "initial-value",

    // External links (bypass client-side navigation)
    External: true,

    // Data attributes (rendered as data-*)
    Data: map[string]string{"id": "123"},

    // Event handlers (WASM only, ignored in SSR)
    OnClick:  func() { /* handle click */ },
    OnSubmit: func() { /* handle form submit */ },
    OnChange: func(value string) { /* handle input change */ },
    OnEnter:  func() { /* handle Enter key */ },

    // Additional attributes
    Extra: map[string]string{"aria-label": "Label"},
}
```

### Page Functions

Pages follow a loader + component pattern:

```go
func MyPage(r *core.Router) func() core.Node {
    // LOADER: Runs on server (SSR) and via API for client navigation
    var items []Item

    // OnLoad only runs when NOT hydrated (server-side or fresh navigation)
    r.OnLoad(func() {
        api.Items.List(func(result []Item, err error) {
            if err == nil {
                items = result
            }
        })
    })

    // COMPONENT: Returns the UI, re-runs on state changes
    return func() core.Node {
        // Store data in state for hydration
        itemsState := r.StateString("items", serializeItems(items))

        return core.Div(core.Class("container"),
            core.H1(core.Attrs{}, core.Text("Items")),
            renderItemList(parseItems(itemsState.Get())),
        )
    }
}
```

### State Management

```go
// Typed state helpers
count := r.StateInt("count", 0)        // *State[int]
name := r.StateString("name", "")      // *State[string]
active := r.StateBool("active", false) // *State[bool]

// Generic state for any type
user := core.UseState(r, "user", User{Name: "Guest"}) // *State[User]

// Read state
current := count.Get()

// Update state (triggers re-render)
count.Set(current + 1)

// Update without re-render (for input tracking)
name.SetQuiet("new value")
```

### Routing

```go
app := core.New()
app.SetTitle("My App")
app.SetDB(db) // Set database for CRUD

// Simple routes (SSR only)
app.Routes().
    GET("/", pages.Home).
    POST("/submit", pages.Submit)

// Hybrid routes (SSR + WASM hydration)
app.Routes().
    Hybrid("/", pages.Home).
    Hybrid("/about", pages.About)

// Route groups with separate WASM bundles
app.RouteGroup("/admin", core.WithBundle("admin")).
    Hybrid("/", admin.Dashboard).
    Hybrid("/users", admin.Users).
    Hybrid("/users/:id", admin.UserDetail)

// Protected routes (require authentication)
app.RouteGroup("/admin", core.WithBundle("admin")).
    Protected().  // All routes in this group require login
    Hybrid("/", admin.Dashboard).
    Hybrid("/users", admin.Users)

// Single route protection
app.Routes().
    Hybrid("/profile", pages.Profile).Protected()

// Auth configuration
app.EnableAuth(core.AuthConfig{
    SessionStore: core.NewMemorySessionStore(),
    CookieName:   "__myapp_session",
    CookieMaxAge: 86400 * 7, // 7 days
    LoginPath:    "/login",  // Redirect here when unauthorized
})

// Route parameters
func UserDetail(r *core.Router) func() core.Node {
    params := r.GetRouteParams() // {"id": "123"}
    userID := params["id"]
    // Or use the shorthand:
    userID := r.Param("id")
    // ...
}

// Current path
currentPath := r.Path() // e.g., "/admin/users/123"

// URL query parameters (works on server and WASM)
orderID := r.Query("order_id") // reads ?order_id=X from URL

// Programmatic navigation
r.Navigate("/admin/users")

// Authenticated user access
func Dashboard(r *core.Router) func() core.Node {
    // Check if user is authenticated
    if user := r.User(); user != nil {
        fmt.Println(user.ID, user.Email, user.Name, user.Roles)
    }

    // Helper methods
    if r.IsAuthenticated() { /* ... */ }
    if r.HasRole("admin") { /* ... */ }
    if r.HasAnyRole("admin", "moderator") { /* ... */ }
}

// External links (cross-bundle navigation)
core.A(core.Attrs{Href: "/admin", External: true}, core.Text("Admin"))

// Start server
app.Run(":8080")
```

### CRUD API Generation

Automatically generate REST endpoints for GORM models:

```go
// Basic CRUD - creates endpoints:
// GET/POST    /__gux_api/crud/counters
// GET/PUT/DEL /__gux_api/crud/counters/:id
// Protected by default when auth is enabled
app.CRUD(models.Counter{})

// Public CRUD (no auth required)
app.CRUD(models.Product{}, core.WithPublic())

// Restricted to specific roles
app.CRUD(models.User{}, core.WithRoles("admin"))

// With DTOs (control what's exposed)
app.CRUD(models.User{},
    core.WithListDTO(dto.UserList{}),                    // For list responses
    core.WithDetailDTO(dto.UserDetail{}, "Posts"),       // For single item, with preload
    core.WithDTO(dto.User{}),                            // Same DTO for both
    core.WithRoles("admin"),                             // Admin only
)

// With create/update hooks (custom logic)
app.CRUD(models.User{},
    core.WithCreateHook(func(data map[string]interface{}) (interface{}, error) {
        user := &models.User{}
        user.Email = data["email"].(string)
        user.Name = data["name"].(string)
        // Hash password before storing
        if password, ok := data["password"].(string); ok {
            user.SetPassword(password)
        }
        return user, nil
    }),
    core.WithUpdateHook(func(existing interface{}, data map[string]interface{}) (interface{}, error) {
        user := existing.(*models.User)
        if email, ok := data["email"].(string); ok {
            user.Email = email
        }
        // Only update password if provided
        if password, ok := data["password"].(string); ok && password != "" {
            user.SetPassword(password)
        }
        return user, nil
    }),
)
```

### Typed API Endpoints

For custom API endpoints beyond CRUD, use the typed endpoint helpers:

```go
import "github.com/dougbarrett/gux/core"

// POST endpoint with request body - auto JSON decode/encode
core.API(app, "POST", "/api/login", func(ctx *core.APIContext, req dto.LoginRequest) (dto.LoginResponse, error) {
    db := ctx.DB().(*gorm.DB)

    // Validate credentials...
    var user models.User
    if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
        return dto.LoginResponse{}, api.Unauthorized("invalid credentials")
    }

    // Create session
    ctx.Login(&core.SessionUser{
        ID:    fmt.Sprint(user.ID),
        Email: user.Email,
        Name:  user.Name,
        Roles: []string{user.Role},
    })

    return dto.LoginResponse{Success: true, Redirect: "/dashboard"}, nil
})

// GET endpoint without request body
core.APIGet(app, "/api/users/:id", func(ctx *core.APIContext) (dto.UserDetail, error) {
    id := ctx.ParamUint("id")

    var user models.User
    db := ctx.DB().(*gorm.DB)
    if err := db.Preload("Posts").First(&user, id).Error; err != nil {
        return dto.UserDetail{}, api.NotFound("user not found")
    }

    return dto.UserDetail{
        ID:    user.ID,
        Email: user.Email,
        Name:  user.Name,
    }, nil
})

// DELETE endpoint
core.APIDelete(app, "/api/users/:id", func(ctx *core.APIContext) error {
    id := ctx.ParamUint("id")

    db := ctx.DB().(*gorm.DB)
    if err := db.Delete(&models.User{}, id).Error; err != nil {
        return api.InternalError("failed to delete user")
    }

    return nil
})
```

**APIContext Methods:**
```go
ctx.Param("id")       // Get path parameter as string
ctx.ParamInt("id")    // Get path parameter as int
ctx.ParamUint("id")   // Get path parameter as uint
ctx.Query("search")   // Get query parameter
ctx.Header("X-Key")   // Get request header
ctx.User()            // Get authenticated user (or nil)
ctx.DB()              // Get database connection
ctx.Login(user)       // Create session for user
ctx.Logout()          // Destroy current session
```

**Error Responses:**
```go
import "github.com/dougbarrett/gux/api"

return nil, api.NotFound("user not found")
return nil, api.BadRequest("invalid email format")
return nil, api.Unauthorized("login required")
return nil, api.Forbidden("access denied")
return nil, api.Conflict("email already exists")
return nil, api.InternalError("database error")
```

### API Authentication

All CRUD and typed API endpoints are **protected by default** when auth is configured on the app.

**CRUD Authentication:**
```go
// Protected by default (requires login)
app.CRUD(models.Order{})

// Public endpoint (no auth required)
app.CRUD(models.Product{}, core.WithPublic())

// Requires specific role(s) - user must have ANY of these
app.CRUD(models.User{}, core.WithRoles("admin"))
app.CRUD(models.AuditLog{}, core.WithRoles("admin", "auditor"))
```

**Typed API Authentication:**
```go
// Protected by default (requires login)
core.APIGet(app, "/api/orders", listOrders)

// Public endpoint (no auth required)
core.APIGet(app, "/api/products", listProducts).Public()

// Mark login endpoint as public
core.API(app, "POST", "/api/login", handleLogin).Public()

// Requires specific role
core.APIGet(app, "/api/stats", getStats).WithRoles("admin")
core.APIDelete(app, "/api/users/:id", deleteUser).WithRoles("admin")
```

**SSR Session Propagation:**

When pages call protected APIs during server-side rendering, the user's session must be propagated. Wire this up in your app.go:

```go
import "yourapp/guxgen/api"

func init() {
    // Enable SSR session propagation
    core.SSRSessionSetter = api.SetEndpointSession
    core.SSRSessionClearer = api.ClearEndpointSession
}
```

This allows page loaders that call `api.Orders.List()` etc. during SSR to inherit the user's authentication.

**Backwards Compatibility:**

If auth is not configured (`app.EnableAuth()` not called), all endpoints remain accessible. Auth enforcement only applies when auth is explicitly configured.

### DTO Mapping

DTOs use `gux` tags for automatic field mapping:

```go
// Model (has sensitive fields)
type User struct {
    gorm.Model
    Email        string
    Name         string
    PasswordHash string // Should NOT be exposed
    Posts        []Post `gorm:"foreignKey:UserID"`
}

// DTO (safe for API responses)
type UserList struct {
    ID    uint   `json:"id" gux:"User.ID"`
    Email string `json:"email" gux:"User.Email"`
    Name  string `json:"name" gux:"User.Name"`
}

// DTO with nested relationship
type UserDetail struct {
    ID    uint        `json:"id" gux:"User.ID"`
    Email string      `json:"email" gux:"User.Email"`
    Name  string      `json:"name" gux:"User.Name"`
    Posts []PostBrief `json:"posts" gux:"User.Posts"`
}

type PostBrief struct {
    ID    uint   `json:"id" gux:"Post.ID"`
    Title string `json:"title" gux:"Post.Title"`
}
```

### Database Access

```go
// Server-side only - r.DB() returns the GORM connection
func LoadData(r *core.Router) func() core.Node {
    var users []models.User

    if db := r.DB(); db != nil {
        // Direct GORM queries (server-side)
        db.(*gorm.DB).Preload("Posts").Find(&users)
    }

    return func() core.Node { /* ... */ }
}

// Client-side - use generated API client
r.OnLoad(func() {
    api.Users.List(func(users []dto.UserList, err error) {
        // Handle response
    })
})
```

### CSRF Protection

CSRF protection is **enabled by default** for all mutating operations (POST, PUT, PATCH, DELETE).

**How it works:**
1. Server generates a CSRF token and embeds it in the HTML (`<meta name="csrf-token">`)
2. Server also sets a cookie (`__gux_csrf`) with the same token
3. Client-side `fetch` package automatically reads the token and includes it in `X-CSRF-Token` header
4. Server validates that header matches cookie (Double Submit Cookie pattern)

**Configuration:**
```go
app := core.New()

// CSRF is enabled by default - no configuration needed

// Customize CSRF settings
app.EnableCSRF(core.CSRFConfig{
    Enabled:        true,
    CookiePath:     "/",
    CookieMaxAge:   43200, // 12 hours
    CookieSameSite: http.SameSiteStrictMode,
    Secure:         true,  // Require HTTPS (for production)
    ExemptPaths:    []string{"/api/webhooks"}, // Skip CSRF for webhooks
})

// Disable CSRF (only for APIs using other auth methods)
app.DisableCSRF()
```

**Transparency:**
- CRUD operations are automatically protected
- The `fetch` package automatically includes CSRF tokens for POST/PUT/PATCH/DELETE
- Generated API clients automatically include CSRF tokens
- No manual token handling required by developers

### UI Components (`ui/` package)

The `ui` package provides styled, accessible components with Tailwind CSS:

```go
import "github.com/dougbarrett/gux/ui"

// Button variants
ui.Button(ui.ButtonProps{
    Variant:  ui.ButtonPrimary,     // Primary, Secondary, Outline, Ghost, Destructive
    Size:     ui.ButtonMD,          // SM, MD, LG
    Type:     "submit",             // button, submit, reset
    Disabled: false,
    OnClick:  func() { /* ... */ },
    Children: []core.Node{core.Text("Click me")},
})

// Input types
ui.Input(ui.InputProps{
    Type:        ui.InputText,      // Text, Email, Password, Number, Search, Tel, URL
    Size:        ui.InputMD,        // SM, MD, LG
    ID:          "email",
    Name:        "email",
    Value:       email.Get(),
    Placeholder: "Enter your email",
    Disabled:    false,
    Required:    true,
    Error:       "",                // Shows error styling when non-empty
    OnChange:    func(v string) { email.SetQuiet(v) },
    OnEnter:     func() { /* submit form */ },
})

// Select dropdown
ui.Select(ui.SelectProps{
    ID:          "role",
    Name:        "role",
    Value:       role.Get(),
    Placeholder: "Select a role",
    Options: []ui.SelectOption{
        {Value: "user", Label: "User"},
        {Value: "admin", Label: "Admin"},
        {Value: "mod", Label: "Moderator", Disabled: true},
    },
    OnChange: func(v string) { role.Set(v) },
})

// Checkbox
ui.Checkbox(ui.CheckboxProps{
    ID:       "remember",
    Name:     "remember",
    Label:    "Remember me",
    Checked:  remember.Get(),
    OnChange: func(checked bool) { remember.Set(checked) },
})

// Card component
ui.Card(ui.CardProps{
    Class: "max-w-md",
    Children: []core.Node{
        ui.CardHeader(ui.CardHeaderProps{
            Title:       "Account Settings",
            Description: "Manage your profile",
        }),
        ui.CardContent(ui.CardContentProps{
            Children: []core.Node{ /* form fields */ },
        }),
        ui.CardFooter(ui.CardFooterProps{
            Children: []core.Node{ /* action buttons */ },
        }),
    },
})

// Alert component
ui.Alert(ui.AlertProps{
    Variant: ui.AlertSuccess, // Success, Error, Warning, Info
    Title:   "Success!",
    Message: "Your changes have been saved.",
})

// Badge component
ui.Badge(ui.BadgeProps{
    Variant: ui.BadgePrimary, // Primary, Secondary, Success, Warning, Error
    Text:    "New",
})

// Modal component
ui.Modal(ui.ModalProps{
    Open:    showModal.Get(),
    OnClose: func() { showModal.Set(false) },
    Title:   "Confirm Delete",
    Children: []core.Node{
        core.P(core.Class("mb-4"), core.Text("Are you sure?")),
        ui.Button(ui.ButtonProps{
            Variant:  ui.ButtonDestructive,
            OnClick:  func() { deleteItem(); showModal.Set(false) },
            Children: []core.Node{core.Text("Delete")},
        }),
    },
})

// DataTable component
ui.DataTable(ui.DataTableProps{
    Columns: []ui.DataTableColumn{
        {Key: "name", Label: "Name", Sortable: true},
        {Key: "email", Label: "Email"},
        {Key: "role", Label: "Role"},
    },
    Data: users,  // []map[string]any or similar
    OnRowClick: func(row map[string]any) {
        r.Navigate(fmt.Sprintf("/users/%v", row["id"]))
    },
})

// Available components:
// Alert, Avatar, Badge, Breadcrumb, Button, Card, Checkbox,
// DataTable, Dropdown, Form, Icon, Input, List, Modal,
// Pagination, Radio, Select, Sidebar, SidebarLayout, Switch,
// Table, Tabs, Textarea, Toast, Tooltip
```

### Hydration Flow

1. **Server renders HTML** with initial state
2. **State serialized** to `<script id="__gux_state">`
3. **WASM loads** and calls `r.Hydrate(state)`
4. **r.IsHydrated()** returns true, `r.OnLoad()` skips data fetching
5. **Component re-renders** with hydrated state
6. **r.ClearHydrated()** allows future navigations to fetch fresh data

---

## CLI Reference

### Installation

```bash
go install github.com/dougbarrett/gux/cmd/gux@latest
```

### Commands

| Command | Description |
|---------|-------------|
| `gux init <name>` | Create new Gux application (interactive) |
| `gux init [--auth\|--auth-public] [--admin] [--claude] <name>` | Create with explicit options |
| `gux gen [--watch]` | Generate API client/server code and WASM entry points |
| `gux build [--go]` | Build WASM and server binary with embedded assets |
| `gux dev [--go] [--watch]` | Build and run dev server (with optional hot reload) |
| `gux model add` | Interactive model builder |
| `gux model add --from-config` | Generate all models from gux.config.json |
| `gux model regen <Name>` | Regenerate files for a model |
| `gux model list` | List models defined in config |
| `gux clean` | Remove generated files (.gux/, guxgen/, assets_gen.go) |
| `gux claude` | Install Claude Code skill and CLAUDE.md |
| `gux update [--check]` | Update gux to latest version |
| `gux version` | Show version |
| `gux help [pattern]` | Show help or boilerplate for a pattern |

### Project Scaffolding

```bash
# Create new project (interactive)
gux init myapp

# Create with explicit options
gux init --module github.com/youruser/myapp myapp
gux init --auth --admin --module github.com/user/app myapp
gux init --auth-public --module github.com/user/app myapp

# Start development server
cd myapp
gux dev
```

### Config File (`gux.config.json`)

Project settings are saved during init and can be reused:

```json
{
  "module": "github.com/user/myapp",
  "auth": "private",   // "none", "private", "public"
  "admin": true,
  "claude": true,      // Claude Code integration
  "port": "8080"
}
```

### Model Scaffolding

Generate complete CRUD scaffolding for models:

```bash
# Interactive model builder
gux model add

# Generate from config
gux model add --from-config

# Regenerate after changes
gux model regen Client
```

**Config-based models** in `gux.config.json`:

```json
{
  "models": {
    "Client": {
      "formLayout": "grid",
      "sections": {
        "Contact": [
          { "name": "FirstName", "type": "string", "required": true, "table": true, "priority": 1 },
          { "name": "Email", "type": "string", "input": "email", "table": true }
        ],
        "Lead Info": [
          { "name": "SalespersonID", "type": "*uint", "relation": "User", "label": "Salesperson", "table": true },
          { "name": "ClosedLead", "type": "bool", "table": true, "badge": { "true": "success", "false": "secondary", "trueLabel": "Closed", "falseLabel": "Open" } }
        ],
        "Business": [
          { "name": "State", "type": "string", "input": "select", "options": [{"value": "CA", "label": "California"}] },
          { "name": "Description", "type": "string", "input": "textarea", "fullWidth": true }
        ]
      },
      "preloads": ["Salesperson"],
      "public": false,
      "roles": ["admin"]
    }
  }
}
```

**Model configuration options**:

| Property | Description |
|----------|-------------|
| `formLayout` | Form field layout: `"stacked"` (default, vertical) or `"grid"` (two-column responsive) |
| `display` | Display field name for Brief DTOs, breadcrumbs, and page titles (auto-detects Name/Title/FirstName if omitted) |
| `preloads` | Relations to preload for detail view |
| `public` | CRUD is public (no auth required) |
| `roles` | Required roles for CRUD access |
| `parent` | Parent model name for child entities (e.g., `"Order"`) |
| `parentField` | FK field linking to parent (e.g., `"OrderID"`) |
| `sidebar` | Set to `false` to exclude child model from sidebar navigation |

**Field configuration options**:

| Property | Description |
|----------|-------------|
| `name` | Field name (PascalCase) |
| `type` | Go type: `string`, `int`, `uint`, `float64`, `bool`, `*uint` (FK), `time.Time`, `[]string` |
| `required` | Not null constraint |
| `input` | Input type: `text`, `email`, `tel`, `textarea`, `select`, `multiselect` |
| `relation` | Related model for FK fields |
| `label` | Display label |
| `table` | Show in list view |
| `priority` | Column priority (1=high, shown first) |
| `sortable` | Column is sortable |
| `filterable` | Can filter by field |
| `fullWidth` | Span full width in grid layout (useful for textareas) |
| `options` | Static select options: `[{value, label}]` |
| `optionsRef` | Reference to optionSets |
| `badge` | Badge display for booleans: `{true, false, trueLabel, falseLabel}` |
| `many2many` | Join table name for M2M relations |

**Generated admin page features** (when `admin: true` in config):

- Human-readable labels ("LeadSource" → "Lead Source")
- List pages with record count subtitle and link-style add button
- Clickable display field (Name/Title) linking to detail page
- Detail pages with avatar initials and delete confirmation modal
- Form pages respect `formLayout` setting for field arrangement

**Generated files** for model `Client`:

```
guxgen/models/client_gen.go       # GORM model
guxgen/dto/client_gen.go          # List and Detail DTOs
guxgen/dto/briefs_gen.go          # Brief DTOs for all relations (shared)
guxgen/admin/client_list_gen.go   # List page
guxgen/admin/client_new_gen.go    # Create form
guxgen/admin/client_detail_gen.go # Detail view
guxgen/admin/client_edit_gen.go   # Edit form
admin/hooks_gen.go                # Hook setter functions (all models, in user-owned admin/)
```

**Important**: All `guxgen/` files (including `guxgen/admin/` layout, dashboard, settings, and model pages) are regenerated on every `gux build` and `gux gen`. Do not manually edit them — changes will be overwritten. The `admin/hooks_gen.go` file is also regenerated, but user-created hook files (e.g., `admin/client_hooks.go`) are never overwritten. To customize a generated admin page, copy it from `guxgen/admin/` to `admin/` and update `app.go` imports accordingly.

```bash
# Re-init project from config (regenerate files)
gux init --config .

# Auto-detected when initializing in existing directory
gux init .
```

Default files (index.html, manifest.json, service-worker.js, wasm_exec.js) are automatically injected at build time. To customize, create a `public/` directory with your own versions.

#### Generated Project Structure

```
myapp/
├── cmd/
│   ├── app/
│   │   └── main.go           # WASM frontend entry point
│   └── server/
│       └── main.go           # HTTP server
├── internal/
│   └── api/
│       ├── types.go          # Shared data types
│       └── example.go        # Example API interface
├── go.mod                    # Go module file
└── Dockerfile                # Multi-stage Docker build
```

### Brief DTOs

When a field has `"relation"` in the config, the generator automatically creates Brief DTOs in `guxgen/dto/briefs_gen.go`. These are lightweight representations used in list/detail DTOs instead of raw FK uint fields.

```go
// Auto-generated in guxgen/dto/briefs_gen.go
type OrderBrief struct {
    ID          uint   `json:"id" gux:"Order.ID"`
    OrderNumber string `json:"order_number" gux:"Order.OrderNumber"` // Uses display field
}
```

The display field is determined by the `display` config option, or auto-detected from `Name`, `Title`, `FirstName`, or `Label`. List/Detail DTOs then use pointer fields:

```go
type OrderItemList struct {
    ID      uint        `json:"id"`
    Order   *OrderBrief `json:"order"`    // Not OrderID uint
    Product *ProductBrief `json:"product"` // Not ProductID uint
}
```

### Parent-Child Entity Management

Child models are displayed as inline tables on the parent's detail page. Configure in `gux.config.json`:

```json
{
  "models": {
    "Order": {
      "display": "OrderNumber",
      "sections": {
        "Details": [
          { "name": "OrderNumber", "type": "string", "table": true }
        ]
      }
    },
    "OrderItem": {
      "parent": "Order",
      "parentField": "OrderID",
      "sidebar": false,
      "sections": {
        "Item Details": [
          { "name": "OrderID", "type": "uint", "relation": "Order" },
          { "name": "ProductID", "type": "uint", "relation": "Product", "table": true },
          { "name": "Quantity", "type": "int", "table": true }
        ]
      }
    }
  }
}
```

This generates:
- **Inline child table** on Order's detail page showing OrderItem records
- **"+ Add Order Item" link** with pre-filled FK query param (`?order_id=2`)
- **`ListFiltered` calls** to load child records by parent FK
- **Sidebar exclusion** — child models with `"sidebar": false` are not added to the admin sidebar

### Admin Page Hooks

Generated admin pages support **view hooks** for customization without modifying generated code. Hook setter functions are generated in `admin/hooks_gen.go`.

**Available hooks** (for each model):

| Setter Function | Purpose | Signature |
|-----------------|---------|-----------|
| `Set{Model}ListActions` | Add buttons to list page header | `func(ctx HookContext) []core.Node` |
| `Set{Model}ListRowActions` | Add buttons per table row | `func(ctx HookContext, item T) []core.Node` |
| `Set{Model}DetailActions` | Add buttons to detail page header | `func(ctx HookContext, item T) []core.Node` |
| `Set{Model}DetailSections` | Add sections to detail page body | `func(ctx HookContext, item T) []core.Node` |
| `Set{Model}FormSections` | Add sections to create/edit forms | `func(ctx HookContext, bool) []core.Node` |
| `Set{Model}BeforeSave` | Validate/transform before API call | `func(ctx HookContext, data map[string]any, isEdit bool) error` |
| `Set{Model}AfterSave` | Run after successful save | `func(ctx HookContext, id uint, isEdit bool)` |

**Usage**: Create a hooks file in `admin/` (e.g., `admin/order_hooks.go`). This file is user-owned and never overwritten:

```go
package admin

import (
    "github.com/dougbarrett/gux/core"
    "github.com/dougbarrett/gux/ui"
    "myapp/guxgen/dto"
)

func init() {
    // Add "Export CSV" button to list page header
    SetOrderListActions(func(ctx HookContext) []core.Node {
        return []core.Node{
            ui.Button(ui.ButtonProps{
                Variant:  ui.ButtonSecondary,
                Children: []core.Node{core.Text("Export CSV")},
                OnClick:  func() { /* export logic */ },
            }),
        }
    })

    // Add custom section to detail page
    SetOrderDetailSections(func(ctx HookContext, item dto.OrderDetail) []core.Node {
        return []core.Node{
            core.Div(core.Class("mt-6 pt-4 border-t border-gray-200 dark:border-gray-700"),
                core.H3(core.Class("text-lg font-semibold text-gray-900 dark:text-white mb-4"),
                    core.Text("Activity Log"),
                ),
                // ... custom content
            ),
        }
    })

    // Validate before save (return error to cancel)
    SetOrderBeforeSave(func(ctx HookContext, data map[string]any, isEdit bool) error {
        if data["order_number"] == "" {
            return fmt.Errorf("order number is required")
        }
        return nil
    })
}
```

`HookContext` provides `Router *core.Router` for navigation and state access.

### CRUD Query Parameter Filtering

All CRUD list endpoints support URL query parameter filtering:

```
GET /__gux_api/crud/orderitems?order_id=2
GET /__gux_api/crud/clients?active=true
GET /__gux_api/crud/products?name=Widget
```

- Filters by any field on the model (matched by JSON tag or snake_case field name)
- Supports `uint`, `int`, `float64`, `bool`, `string`, and pointer types
- Unknown parameters are silently ignored

**Generated API client method**:

```go
// List all records
api.OrderItems.List(func(result []dto.OrderItemList, err error) { ... })

// List with filters (WASM client appends query params to URL)
api.OrderItems.ListFiltered(map[string]string{
    "order_id": "2",
}, func(result []dto.OrderItemList, err error) { ... })
```

### Audit Logging

Enable automatic audit logging for any CRUD model to track who changed what and when.

**Enable Audit Logging:**

```go
// Audit all fields
app.CRUD(models.Document{}, core.WithAuditLog())

// Ignore sensitive fields in audit diffs
app.CRUD(models.User{},
    core.WithAuditLog("PasswordHash"),
    core.WithRoles("admin"),
)
```

**What Gets Logged:**

| Action | Changes Field |
|--------|--------------|
| Create | Full snapshot of created entity (minus ignored fields) |
| Update | Field-level diff: `{"name": {"old": "Alice", "new": "Bob"}}` |
| Delete | Entity type and ID |

**AuditEntry Model:**

```go
type AuditEntry struct {
    ID         uint        `json:"id"`
    CreatedAt  time.Time   `json:"created_at"`
    Action     AuditAction `json:"action"`      // "create", "update", "delete"
    EntityType string      `json:"entity_type"` // Model name
    EntityID   uint        `json:"entity_id"`
    UserID     string      `json:"user_id"`
    UserEmail  string      `json:"user_email"`
    IPAddress  string      `json:"ip_address"`
    Changes    string      `json:"changes"`     // JSON
}
```

**Viewing Audit Logs:**

Register `AuditEntry` as a read-only CRUD endpoint:

```go
app.CRUD(core.AuditEntry{}, core.WithRoles("admin"))
```

**Behavior:**
- Audit entries auto-migrate when any model enables audit
- Logging is async (goroutine) — never blocks the response
- Failures are logged to stderr, never cause request errors
- Public endpoints log with empty UserID/UserEmail
- `UpdatedAt` and `DeletedAt` are always excluded from diffs

### Router.Query() Method

Access URL query parameters in page functions:

```go
func OrderItemNew(r *core.Router) func() core.Node {
    // Read query parameter (works on both server and WASM)
    parentID := r.Query("order_id") // reads ?order_id=X from URL

    return func() core.Node {
        // Pre-fill form with parent FK if present
        // ...
    }
}
```

### Build Regeneration

`gux build` always runs `gux gen` first, which regenerates all files in `guxgen/`. This means:

- **Never manually edit** files in `guxgen/` — they will be overwritten
- **Custom code** belongs in `admin/` (hooks), `models/` (custom methods), `dto/` (custom DTOs)
- After changing `gux.config.json`, run `gux model regen <Name>` to update generated files
- If generated code appears stale, run `gux clean && gux gen` for a fresh build

### Customizing Generated Admin Pages

All admin pages (layout, dashboard, settings, and model CRUD pages) live in `guxgen/admin/` and are regenerated on every `gux gen` or `gux build`. **Do not edit these files directly** — changes will be lost.

To customize a generated admin page:

1. **Copy** the file from `guxgen/admin/` to your project's `admin/` directory:
   ```bash
   cp guxgen/admin/dashboard.go admin/dashboard.go
   ```
2. **Update `app.go`** to import your local `admin` package instead of (or in addition to) `guxgen/admin`, and update route references to point to your local copy
3. Your local copy in `admin/` will **not** be overwritten by regeneration

**For minor customizations**, prefer using [Admin Page Hooks](#admin-page-hooks) instead of copying files — hooks survive regeneration automatically.

## API Code Generation

### Defining API Interfaces

Use annotations to define type-safe APIs:

```go
// internal/api/posts.go
package api

import "context"

// @client PostsClient
// @basepath /api/posts
type PostsAPI interface {
    // @route GET /
    GetAll(ctx context.Context) ([]Post, error)

    // @route GET /{id}
    GetByID(ctx context.Context, id int) (*Post, error)

    // @route POST /
    Create(ctx context.Context, req CreatePostRequest) (*Post, error)

    // @route PUT /{id}
    Update(ctx context.Context, id int, req CreatePostRequest) (*Post, error)

    // @route DELETE /{id}
    Delete(ctx context.Context, id int) error

    // @route GET /{userId}/posts/{postId}
    GetUserPost(ctx context.Context, userId int, postId int) (*Post, error)
}
```

### Annotations Reference

| Annotation | Description | Example |
|------------|-------------|---------|
| `@client <Name>` | Names the generated client struct | `@client PostsClient` |
| `@basepath <path>` | Base URL path for all endpoints | `@basepath /api/posts` |
| `@route <METHOD> <path>` | HTTP method and path for endpoint | `@route GET /{id}` |

### Path Parameters

- Use `{paramName}` syntax in paths
- Parameter names must match function argument names exactly
- Parameters can be `int` or `string` types
- The generator automatically detects the type from your function signature
- `int` parameters are validated on the server and return 400 if invalid
- `string` parameters are extracted directly without conversion

### Request Bodies

- Struct types (not primitives) are treated as request bodies
- `context.Context` is always skipped
- Path parameters are extracted, remaining structs become the body

### Generate Code

```bash
gux gen                       # Scans ./internal/api directory
gux gen --dir internal/api    # Explicit directory
```

Generates:
- `client_shared_gen.go` - Shared client types and functions (generated once per directory)
- `*_client_gen.go` - Type-safe HTTP client (WASM only, `//go:build js && wasm`)
- `*_server_gen.go` - HTTP handler wrapper

### Using Generated Client (WASM)

```go
//go:build js && wasm

client := api.NewPostsClient()

// With options
client := api.NewPostsClient(
    api.WithBaseURL("https://api.example.com"),
    api.WithHeader("Authorization", "Bearer token"),
)

// With dynamic auth (token refreshed on each request)
client := api.NewPostsClient(
    api.WithAuthProvider(func() string {
        return "Bearer " + getAuthToken()
    }),
)

// Make requests
posts, err := client.GetAll()
post, err := client.GetByID(123)
created, err := client.Create(api.CreatePostRequest{Title: "Hello"})
err := client.Delete(123)
```

### Using Generated Server Handler

```go
// Implement the interface
type PostsService struct { /* ... */ }

func (s *PostsService) GetAll(ctx context.Context) ([]api.Post, error) {
    // implementation
}
// ... implement all methods

// Wire up in server
service := NewPostsService()
handler := api.NewPostsAPIHandler(service)

// Add middleware
handler.Use(
    server.Logger(),
    server.CORS(server.CORSOptions{}),
    server.Recover(),
)

// Register routes
handler.RegisterRoutes(mux)
```

### Server-Side Errors

```go
import gqapi "github.com/dougbarrett/gux/api"

// Return structured errors
return nil, gqapi.NotFoundf("post %d not found", id)
return nil, gqapi.BadRequest("invalid email format")
return nil, gqapi.Unauthorized("authentication required")
return nil, gqapi.Forbidden("access denied")
return nil, gqapi.Conflict("resource already exists")
return nil, gqapi.InternalError("unexpected error")
```

## Server Utilities

### Middleware

```go
import "github.com/dougbarrett/gux/server"

// Chain middleware
handler := server.Chain(
    server.Logger(),
    server.CORS(server.CORSOptions{}),
    server.Recover(),
    server.RequestID(),
)(apiHandler)

// Or use with generated handler
handler.Use(
    server.Logger(),
    server.CORS(server.CORSOptions{}),
)
```

### SPA Handler

```go
// Serves static files with SPA fallback
spa := server.NewSPAHandler("./static")
mux.HandleFunc("/", spa.ServeHTTP)
```

## Build & Deployment

### Building WASM

```bash
# Standard Go (~5MB)
GOOS=js GOARCH=wasm go build -o main.wasm ./app

# TinyGo (~500KB)
tinygo build -o main.wasm -target wasm -no-debug ./app
```

### Docker

The scaffold includes a multi-stage Dockerfile:
1. TinyGo stage - compiles WASM
2. Go stage - builds server binary
3. Alpine stage - minimal production image (~20MB)

```bash
docker build -t myapp .
docker run -p 8080:8080 myapp
```

## Examples/Minimal Reference

The `examples/minimal/` directory provides a complete reference implementation:

### Structure

```
examples/minimal/
├── app.go              # Main entry, routes, CRUD registration
├── models/
│   ├── counter.go      # Simple counter model
│   ├── user.go         # User with password hashing
│   └── post.go         # Post with User relationship
├── dto/
│   ├── user.go         # UserList, UserDetail (excludes PasswordHash)
│   └── post.go         # PostList, PostDetail (includes Author)
├── pages/
│   ├── home.go         # Counter page with API persistence
│   ├── about.go        # Static page
│   └── nav.go          # Shared navigation
├── admin/
│   ├── dashboard.go    # Admin overview
│   ├── users.go        # User list
│   ├── user_detail.go  # User detail view
│   ├── user_edit.go    # User edit form
│   ├── user_new.go     # Create user form
│   ├── posts.go        # Post management
│   └── nav.go          # Admin navigation
└── guxgen/
    ├── api/            # Generated API client
    └── wasm/           # WASM entry points
```

### Key Patterns Demonstrated

**Route Groups with Bundles**:
```go
// Public routes use default "app" bundle
app.Routes().
    Hybrid("/", pages.Home).
    Hybrid("/about", pages.About)

// Admin routes use separate "admin" bundle
app.RouteGroup("/admin", core.WithBundle("admin")).
    Hybrid("/", admin.Dashboard).
    Hybrid("/users", admin.Users)
```

**CRUD with Security Hooks**:
```go
app.CRUD(models.User{},
    core.WithListDTO(dto.UserList{}),           // Hides PasswordHash
    core.WithDetailDTO(dto.UserDetail{}, "Posts"),
    core.WithCreateHook(func(data map[string]interface{}) (interface{}, error) {
        user := &models.User{}
        user.Email = data["email"].(string)
        if pw, ok := data["password"].(string); ok {
            user.SetPassword(pw) // bcrypt hash
        }
        return user, nil
    }),
)
```

**State Serialization for Lists**:
```go
func Posts(r *core.Router) func() core.Node {
    var posts []dto.PostList

    r.OnLoad(func() {
        api.Posts.List(func(result []dto.PostList, err error) {
            if err == nil {
                posts = result
            }
        })
    })

    return func() core.Node {
        // Serialize to state for hydration
        postsJSON, _ := json.Marshal(posts)
        postsState := r.StateString("posts", string(postsJSON))

        // Parse from state (works on server and client)
        var displayPosts []dto.PostList
        json.Unmarshal([]byte(postsState.Get()), &displayPosts)

        return renderPostTable(displayPosts)
    }
}
```

**Cross-Bundle Navigation**:
```go
// From admin back to public (different WASM bundle)
core.A(core.Attrs{Href: "/", External: true, Class: "..."},
    core.Text("Back to Home"),
)
```

## Best Practices

### Core Framework

1. **Use `r.OnLoad()` for data fetching** - automatically skipped when hydrated
2. **Serialize complex state** - use JSON for lists/objects in state
3. **Keep loaders simple** - move complex logic to separate functions
4. **Use DTOs to hide sensitive data** - never expose password hashes, internal IDs
5. **Use hooks for validation** - create/update hooks for business logic
6. **Mark cross-bundle links as External** - prevents WASM from handling them
7. **Use `SetQuiet()` for input tracking** - avoids unnecessary re-renders

## Common Patterns

### Page with Data Loading

```go
func MyPage(r *core.Router) func() core.Node {
    var items []Item

    r.OnLoad(func() {
        api.Items.List(func(result []Item, err error) {
            if err == nil {
                items = result
            }
        })
    })

    return func() core.Node {
        itemsJSON, _ := json.Marshal(items)
        itemsState := r.StateString("items", string(itemsJSON))

        var displayItems []Item
        json.Unmarshal([]byte(itemsState.Get()), &displayItems)

        return core.Div(core.Class("container"),
            renderItemList(displayItems),
        )
    }
}
```

### Form Submission

```go
func CreateForm(r *core.Router) func() core.Node {
    return func() core.Node {
        title := r.StateString("title", "")
        body := r.StateString("body", "")

        onSubmit := func() {
            api.Posts.Create(CreateRequest{
                Title: title.Get(),
                Body:  body.Get(),
            }, func(post Post, err error) {
                if err != nil {
                    // Handle error
                    return
                }
                // Success - redirect or update
                r.Navigate("/posts")
            })
        }

        return core.Form(core.Attrs{OnSubmit: onSubmit},
            core.Input(core.Attrs{
                Value:    title.Get(),
                OnChange: func(v string) { title.SetQuiet(v) },
            }),
            core.Textarea(core.Attrs{
                Value:    body.Get(),
                OnChange: func(v string) { body.SetQuiet(v) },
            }),
            core.Button(core.Attrs{Type: "submit"}, core.Text("Create")),
        )
    }
}
```
