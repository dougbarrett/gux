# Gux Framework

Gux is a full-stack Go framework for building modern web applications with WebAssembly.

## Architecture

### Core Framework (`core/`)

The core package provides a universal rendering system that works identically on server (SSR) and client (WASM):

- **Node System** (`node.go`) - Abstract UI tree representation
- **Elements** (`elements.go`) - HTML element helpers (Div, Button, Input, Video, Audio, etc.)
- **Renderers** - HTML (`html_renderer.go`) for server, DOM (`dom_renderer.go`) for WASM
- **App & Routing** (`app.go`, `router_server.go`, `router_wasm.go`) - Hybrid SSR+WASM routing
- **CRUD** (`crud.go`) - Automatic REST API generation with DTOs and hooks
- **CSRF** (`csrf.go`) - Automatic CSRF protection for mutating operations
- **Script Loader** (`script_loader_wasm.go`) - Dynamic JS/CSS loading with deduplication

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

**CRUD with DTOs and Authentication**:
```go
// Protected by default (requires authentication when auth is enabled)
app.CRUD(models.User{},
    core.WithListDTO(dto.UserList{}),
    core.WithDetailDTO(dto.UserDetail{}, "Posts"), // with preload
    core.WithCreateHook(func(data map[string]interface{}) (interface{}, error) {
        // Custom creation logic (e.g., password hashing)
    }),
    core.WithRoles("admin"), // Require admin role
)

// Public CRUD (no authentication required)
app.CRUD(models.Product{},
    core.WithPublic(), // Anyone can access
    core.WithListDTO(dto.ProductList{}),
)
```

**Typed API Endpoints** (`endpoint.go`):

For custom API endpoints with compile-time type safety:

```go
// Define request/response types (in dto/ package)
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}
type LoginResponse struct {
    Success  bool   `json:"success"`
    Redirect string `json:"redirect,omitempty"`
}

// POST endpoint - public (e.g., login doesn't require auth)
core.API(app, "POST", "/api/login", func(ctx *core.APIContext, req LoginRequest) (LoginResponse, error) {
    db := ctx.DB().(*gorm.DB)
    // ... business logic
    ctx.Login(&core.SessionUser{ID: "123", Email: user.Email})
    return LoginResponse{Success: true, Redirect: "/dashboard"}, nil
}).Public()  // No auth required for login

// GET endpoint - protected by default
core.APIGet(app, "/api/users/:id", func(ctx *core.APIContext) (dto.UserDetail, error) {
    id := ctx.ParamUint("id")
    // ... fetch user
    return userDetail, nil
})  // Requires authentication when auth is enabled

// GET endpoint - public
core.APIGet(app, "/api/products", func(ctx *core.APIContext) ([]dto.Product, error) {
    // ... fetch products
    return products, nil
}).Public()  // Anyone can access

// DELETE endpoint - requires admin role
core.APIDelete(app, "/api/users/:id", func(ctx *core.APIContext) error {
    id := ctx.ParamUint("id")
    // ... delete user
    return nil
}).WithRoles("admin")  // Only admins can delete users
```

**APIContext Methods**:
- `ctx.Param("id")` - Get path parameter as string
- `ctx.ParamInt("id")`, `ctx.ParamUint("id")` - Get path parameter as int/uint
- `ctx.Query("search")` - Get query parameter
- `ctx.Header("X-Key")` - Get request header
- `ctx.User()` - Get authenticated user (or nil)
- `ctx.DB()` - Get database connection
- `ctx.Login(user)` - Create session for user
- `ctx.Logout()` - Destroy current session

**Error Handling**: Return `api.NotFound()`, `api.BadRequest()`, `api.Unauthorized()`, etc. for proper HTTP status codes.

## Project Structure

```
goquery/
├── core/                    # Universal rendering framework
│   ├── node.go             # Node interface & types
│   ├── elements.go         # Element helpers
│   ├── app.go              # App, Router, State
│   ├── crud.go             # CRUD API generation
│   ├── endpoint.go         # Typed API endpoints (APIContext, API, APIGet, APIDelete)
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
gux build         # Production build (always runs gux gen first)
gux gen           # Generate API client/server code
```

**Note**: `gux build` always runs `gux gen` first, regenerating all files in `guxgen/`. Never manually edit files in `guxgen/` — they will be overwritten.

### Customizing Generated Admin Pages

All admin pages (layout, dashboard, settings, model CRUD pages) live in `guxgen/admin/` and are regenerated on every `gux gen` or `gux build`. **Do not edit these files directly.**

To customize a generated admin page:

1. **Copy** the file from `guxgen/admin/` to your project's `admin/` directory
2. **Change the package** declaration if needed (both directories use `package admin`)
3. **Update `app.go`** to import your project's `admin` package instead of `guxgen/admin`
4. **Remove the route** that referenced the generated version and point it to your custom version

For example, to customize the dashboard:
```bash
cp guxgen/admin/dashboard.go admin/dashboard.go
```
Then update `app.go` imports to use your local `admin` package for that page.

**For minor customizations**, prefer using [Admin Page Hooks](#admin-page-hooks) instead of copying files — hooks survive regeneration.

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

## API Authentication

All CRUD and typed API endpoints are **protected by default** when authentication is configured. This provides secure-by-default behavior.

### CRUD Authentication

```go
// Protected by default (requires login)
app.CRUD(models.Order{})

// Public endpoint (no auth required)
app.CRUD(models.Product{}, core.WithPublic())

// Requires specific role(s) - user must have ANY of these roles
app.CRUD(models.User{}, core.WithRoles("admin"))
app.CRUD(models.AuditLog{}, core.WithRoles("admin", "auditor"))
```

### Typed API Authentication

```go
// Protected by default (requires login)
core.APIGet(app, "/api/orders", listOrders)

// Public endpoint (no auth required)
core.APIGet(app, "/api/products", listProducts).Public()

// Requires admin role
core.APIGet(app, "/api/stats", getStats).WithRoles("admin")
core.API(app, "POST", "/api/admin/users", createUser).WithRoles("admin")
core.APIDelete(app, "/api/users/:id", deleteUser).WithRoles("admin")
```

### SSR Session Propagation

When pages call protected APIs during server-side rendering, the user's session is automatically propagated. To enable this, wire up the generated API functions in your main():

```go
import "yourapp/guxgen/api"

func main() {
    // Wire up SSR session propagation
    core.SSRSessionSetter = api.SetEndpointSession
    core.SSRSessionClearer = api.ClearEndpointSession

    // ... rest of setup
}
```

This allows page loaders to call protected API endpoints during SSR while inheriting the user's authentication.

### Backwards Compatibility

If auth is not configured on the app (`app.EnableAuth()` not called), all endpoints remain accessible (backwards compatible behavior). Auth enforcement only applies when auth is explicitly configured.

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

## Model Scaffolding

The `gux model` commands generate complete CRUD scaffolding from `gux.config.json`.

### Auth Preset

When `gux init --auth` or `gux init --auth-public` is used, a User model is automatically created with the `"preset": "auth"` setting. This preset:

1. **Auto-adds authentication fields**:
   - `PasswordHash string` with `json:"-"` tag (excluded from JSON responses)
   - `Verified bool` for email verification

2. **Generates password methods** in `models/user_gen.go`:
   ```go
   func (u *User) SetPassword(password string) error {
       hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
       if err != nil {
           return err
       }
       u.PasswordHash = string(hash)
       return nil
   }

   func (u *User) CheckPassword(password string) bool {
       return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
   }
   ```

3. **Excludes PasswordHash from DTOs** - DTOs only include safe fields

### Customizable Roles

User roles are defined at the top level of `gux.config.json` and automatically applied to auth preset models during regeneration:

```json
{
  "module": "myapp",
  "auth": "private",
  "roles": [
    {"value": "user", "label": "User"},
    {"value": "moderator", "label": "Moderator"},
    {"value": "admin", "label": "Admin"}
  ],
  "models": {
    "User": {
      "preset": "auth",
      "sections": {
        "Account": [
          {"name": "Email", "type": "string", "required": true, "table": true, "input": "email"},
          {"name": "Name", "type": "string", "required": true, "table": true},
          {"name": "Role", "type": "string", "table": true, "input": "select"}
        ]
      }
    }
  }
}
```

**To customize roles:**

1. Edit the top-level `roles` array in `gux.config.json`
2. Run `gux model regen User`
3. The generated admin pages will have the updated role dropdown options

The `roles` array is the source of truth - when regenerating auth preset models, the top-level roles override inline options in the model's Role field.

### Generated Files

For a model named `User` with auth preset:

| File | Description |
|------|-------------|
| `guxgen/models/user_gen.go` | GORM model with SetPassword/CheckPassword methods |
| `guxgen/dto/user_gen.go` | UserList and UserDetail DTOs (excludes PasswordHash) |
| `guxgen/admin/user_list_gen.go` | Admin list page |
| `guxgen/admin/user_new_gen.go` | Admin create form (includes Password field) |
| `guxgen/admin/user_detail_gen.go` | Admin detail view |
| `guxgen/admin/user_edit_gen.go` | Admin edit form (includes Password field) |
| `admin/hooks_gen.go` | Hook setter functions for all models |

## Admin Page Hooks

Scaffolded admin pages support **view hooks** that allow customization without modifying generated code. Generated admin pages live in `guxgen/admin/`, and hook setter functions are generated in `admin/` for each model.

### Available Hooks

| Setter Function | Purpose | Signature |
|-----------------|---------|-----------|
| `Set{Model}ListActions` | Add buttons to list page header | `func(ctx HookContext) []core.Node` |
| `Set{Model}ListRowActions` | Add buttons per table row | `func(ctx HookContext, item T) []core.Node` |
| `Set{Model}DetailActions` | Add buttons to detail page header | `func(ctx HookContext, item T) []core.Node` |
| `Set{Model}DetailSections` | Add sections to detail page body | `func(ctx HookContext, item T) []core.Node` |
| `Set{Model}FormSections` | Add sections to create/edit forms | `func(ctx HookContext, bool) []core.Node` |
| `Set{Model}BeforeSave` | Validate/transform before API call | `func(ctx HookContext, data map[string]any, isEdit bool) error` |
| `Set{Model}AfterSave` | Run after successful save | `func(ctx HookContext, id uint, isEdit bool)` |

### Usage Example

Create a hooks file (e.g., `admin/client_hooks.go`) - this file is never overwritten by regeneration:

```go
package admin

import (
    "fmt"
    "strings"

    "github.com/dougbarrett/gux/core"
    "github.com/dougbarrett/gux/ui"
    "myapp/guxgen/dto"
)

func init() {
    // Add "Export CSV" button to list page header
    SetClientListActions(func(ctx HookContext) []core.Node {
        return []core.Node{
            ui.Button(ui.ButtonProps{
                Variant:  ui.ButtonSecondary,
                Children: []core.Node{core.Text("Export CSV")},
                OnClick:  func() { /* export logic */ },
            }),
        }
    })

    // Add "Send Email" button to detail page
    SetClientDetailActions(func(ctx HookContext, item dto.ClientDetail) []core.Node {
        return []core.Node{
            ui.Button(ui.ButtonProps{
                Variant:  ui.ButtonSecondary,
                Children: []core.Node{core.Text("Send Email")},
                OnClick:  func() { /* email logic */ },
            }),
        }
    })

    // Add custom section to detail page
    SetClientDetailSections(func(ctx HookContext, item dto.ClientDetail) []core.Node {
        return []core.Node{
            core.Div(core.Class("mt-6 pt-4 border-t border-gray-200 dark:border-gray-700"),
                core.H3(core.Class("text-lg font-semibold text-gray-900 dark:text-white mb-4"),
                    core.Text("Activity Log"),
                ),
                // ... activity log content
            ),
        }
    })

    // Validate before save (return error to cancel)
    SetClientBeforeSave(func(ctx HookContext, data map[string]any, isEdit bool) error {
        email, _ := data["email"].(string)
        if email != "" && !strings.Contains(email, "@") {
            return fmt.Errorf("invalid email format")
        }
        return nil
    })
}
```

### Hook Context

`HookContext` provides access to the router for navigation and state (re-exported from `guxgen/admin`):

```go
type HookContext struct {
    Router *core.Router
}

// Example: Navigate on button click
OnClick: func() {
    ctx.Router.Navigate("/admin/clients/export")
}
```

### Generated Files

| File | Location | Purpose |
|------|----------|---------|
| `{model}_*_gen.go` | `guxgen/admin/` | Admin pages with hook slot variables |
| `hooks_gen.go` | `admin/` | All hook setter functions (`Set*` functions) for all models |

### Regeneration Safety

- Generated files (`*_gen.go`) are regenerated on each `gux model regen`
- Hook files (e.g., `client_hooks.go`) in `admin/` are user-owned and never overwritten
- Running `gux model regen` preserves all custom hooks

## Parent-Child Entity Management

Child models are displayed as inline tables on the parent's detail page. Configure in `gux.config.json`:

```json
"OrderItem": {
    "parent": "Order",
    "parentField": "OrderID",
    "sidebar": false,
    "sections": { ... }
}
```

- `parent`: Parent model name
- `parentField`: FK field linking to parent (e.g., `"OrderID"`)
- `sidebar`: Set to `false` to exclude child model from sidebar navigation

This generates inline child tables on the parent detail page with "Add" links that pre-fill the parent FK via query parameter.

## Brief DTOs

Brief DTOs are auto-generated in `guxgen/dto/briefs_gen.go` for all relation fields. They provide lightweight representations used in list/detail DTOs:

```go
type OrderBrief struct {
    ID          uint   `json:"id" gux:"Order.ID"`
    OrderNumber string `json:"order_number" gux:"Order.OrderNumber"`
}
```

The display field is determined by the `"display"` model config option, or auto-detected from `Name`, `Title`, `FirstName`, or `Label`.

## Audit Logging

Track who changed what and when for regulatory compliance. Enable per-model with `WithAuditLog()`:

```go
app.CRUD(models.Document{}, core.WithAuditLog())
app.CRUD(models.User{}, core.WithAuditLog("PasswordHash"))

// View audit logs (admin only)
app.CRUD(core.AuditEntry{}, core.WithRoles("admin"))
```

- Auto-logs create/update/delete operations
- Field-level diffs for updates
- Captures user ID, email, IP address, timestamp
- Auto-migrates `audit_entries` table
- Async logging (never blocks responses)
- Configurable field exclusion

## CRUD Query Parameter Filtering

CRUD list endpoints support URL query parameter filtering:

```
GET /__gux_api/crud/orderitems?order_id=2
GET /__gux_api/crud/clients?active=true
```

Generated API clients also have a `ListFiltered` method:

```go
api.OrderItems.ListFiltered(map[string]string{"order_id": "2"}, func(result []dto.OrderItemList, err error) {
    // filtered results
})
```

## Router.Query()

Access URL query parameters in page functions (works on both server and WASM):

```go
orderID := r.Query("order_id") // reads ?order_id=X from URL
```

## Testing

All bug fixes and new features **must** include unit tests. Run the full test suite before committing:

```bash
go test ./... -count=1 -timeout 120s
```

### Test Files

| File | Coverage |
|------|----------|
| `cmd/gux/build_test.go` | CRUD parsing, endpoint generation, pointer type mapping, `isNumericParam`, `generateFieldMapping`, `generateServerAPICode` |
| `cmd/gux/modelgen_test.go` | Template field conversion, form field generation, detail field generation, pointer types, display field detection, DTO type overrides, GORM tags |
| `cmd/gux/generate_test.go` | Dockerfile regeneration |
| `core/endpoint_test.go` | Path parameter extraction, `:param` to `{param}` conversion, param name extraction |
| `core/crud_test.go` | CRUD API operations |
| `core/audit_test.go` | Audit logging |

### Testing Requirements

- **Pointer type handling**: Every code generation path that maps between model fields and DTO fields must be tested with pointer types (`*string`, `*float64`, `*uint`, `*time.Time`). This includes `generateFieldMapping`, `generateServerAPICode`, `convertToTemplateField`, and `generateDetailFieldCode`.
- **Path parameters**: Test both `:param` and `{param}` syntax. Test numeric params (`id`, `user_id`) and string params (`slug`, `code`).
- **Table-driven tests**: Use Go table-driven test patterns with `t.Run()` subtests.
- **Temp directories**: Tests that need file system access should use `t.TempDir()` and `os.Chdir()` with defer to restore.
- **Cache clearing**: Tests that use `getModelFieldTypes()` must reset `modelFieldTypesCache` before running.

## Lifecycle Hooks: OnMount & OnUnmount

`OnMount` fires after an element's DOM node is created and all children are appended. `OnUnmount` fires before the element's DOM tree is replaced on the next render. Both only run in WASM — they're ignored during SSR.

```go
core.Div(core.Attrs{
    ID: "my-element",
    OnMount: func(el any) {
        // el is a syscall/js.Value in WASM
        // Initialize JS libraries, add DOM listeners, etc.
        jsEl := el.(js.Value)
        initLibrary(jsEl)
    },
    OnUnmount: func() {
        // Clean up before DOM replacement
        // Dispose JS library instances, remove global listeners, etc.
        disposeLibrary()
    },
}, children...)
```

**Key behavior:**
- `OnMount` is called on **every render** (gux replaces DOM trees on re-render)
- `OnUnmount` is called **before** the next render clears the DOM — use it for cleanup
- `OnMount` receives the DOM element as `any` (cast to `syscall/js.Value` in WASM)
- `OnUnmount` takes no arguments — capture references in closures if needed
- The VideoPlayer component uses `OnUnmount` to call `videojs.dispose()` automatically

## Script & CSS Loading

Dynamic script/CSS loading with deduplication for WASM builds:

```go
// Load external JS (only loads once per URL, even with multiple calls)
core.LoadScript("https://cdn.example.com/lib.js", func(err error) {
    if err != nil {
        println("Failed:", err.Error())
        return
    }
    // Library is now available via js.Global()
})

// Load external CSS
core.LoadCSS("https://cdn.example.com/styles.css", func(err error) {
    // Stylesheet loaded
})
```

On non-WASM builds, both functions are no-ops that immediately call the callback with nil.

## Media Elements

HTML element helpers for media content:

```go
core.Video(attrs, children...)   // <video>
core.Audio(attrs, children...)   // <audio>
core.Source(attrs)               // <source> (self-closing)
core.Track(attrs)                // <track> (self-closing)
core.Iframe(attrs, children...)  // <iframe>
core.Canvas(attrs, children...)  // <canvas>
```

**Media attributes on `Attrs`:**

| Field | Type | Description |
|-------|------|-------------|
| `Poster` | `string` | Video poster image URL |
| `Autoplay` | `bool` | Auto-play media |
| `Loop` | `bool` | Loop playback |
| `Muted` | `bool` | Start muted |
| `Controls` | `bool` | Show playback controls |
| `Preload` | `string` | `"none"`, `"metadata"`, `"auto"` |
| `Width` | `string` | Element width |
| `Height` | `string` | Element height |

## VideoPlayer Component

`ui.VideoPlayer` provides an integrated video player with optional VideoJS and Google IMA ads support.

### Basic Usage

```go
// Native HTML5 video
ui.VideoPlayer(ui.VideoPlayerProps{
    Sources:  []ui.VideoSource{{Src: "/video.mp4", Type: "video/mp4"}},
    Controls: true,
})

// With VideoJS (enhanced player UI, loaded from CDN)
ui.VideoPlayer(ui.VideoPlayerProps{
    Sources:       []ui.VideoSource{{Src: "/video.mp4", Type: "video/mp4"}},
    EnableVideoJS: true,
    Controls:      true,
    OnPlay:        func() { println("playing") },
})

// With Google IMA pre-roll ads
ui.VideoPlayer(ui.VideoPlayerProps{
    Sources:       []ui.VideoSource{{Src: "/video.mp4", Type: "video/mp4"}},
    EnableVideoJS: true,
    EnableAds:     true,
    AdTagURL:      "https://pubads.g.doubleclick.net/gampad/ads?...",
    Controls:      true,
})
```

### Props

| Prop | Type | Description |
|------|------|-------------|
| `Sources` | `[]VideoSource` | Video sources (`Src` + `Type` MIME) |
| `Poster` | `string` | Poster image URL |
| `Width`, `Height` | `string` | Dimensions (default: `"100%"`, `"auto"`) |
| `AutoPlay`, `Loop`, `Muted`, `Controls` | `bool` | Standard video attributes |
| `Preload` | `string` | Default: `"metadata"` |
| `EnableVideoJS` | `bool` | Use VideoJS enhanced player |
| `VideoJSCSS`, `VideoJSScript` | `string` | Custom VideoJS CDN URLs |
| `EnableAds` | `bool` | Enable Google IMA ads (requires VideoJS) |
| `AdTagURL` | `string` | VAST/VMAP ad tag URL |

### Video Event Handlers (WASM only)

| Handler | Signature | Description |
|---------|-----------|-------------|
| `OnPlay` | `func()` | Video started playing |
| `OnPause` | `func()` | Video paused |
| `OnEnded` | `func()` | Video finished |
| `OnTimeUpdate` | `func(currentTime float64)` | Periodic playback position (~4x/sec) in seconds |
| `OnError` | `func(msg string)` | Playback error |
| `OnReady` | `func()` | Player fully initialized |

### Google IMA Ad Event Handlers (WASM only)

All ad event handlers receive `ui.AdInfo` with ad metadata:

```go
type AdInfo struct {
    AdID, AdTitle, AdSystem string
    Duration, CurrentTime   float64
    IsSkippable             bool
    ContentType             string
    AdPosition, TotalAds    int
    ClickURL                string
}
```

| Handler | Signature | Description |
|---------|-----------|-------------|
| `OnAdLoaded` | `func(AdInfo)` | Ad creative loaded |
| `OnAdStarted` | `func(AdInfo)` | Ad playback began |
| `OnAdComplete` | `func(AdInfo)` | Ad finished playing |
| `OnAdSkipped` | `func(AdInfo)` | User skipped the ad |
| `OnAdProgress` | `func(AdInfo)` | Periodic progress tick |
| `OnAdClicked` | `func(AdInfo)` | User clicked the ad |
| `OnAdError` | `func(string)` | Ad error (falls back to content) |
| `OnContentPause` | `func()` | Content pausing for ad break |
| `OnContentResume` | `func()` | Content resuming after ads |

## Timer Helpers

`r.SetInterval()` and `r.SetTimeout()` provide managed JavaScript timers with automatic cleanup on navigation. WASM only — no-ops during SSR.

```go
func MyPage(r *core.Router) func() core.Node {
    count := r.StateInt("count", 0)

    r.OnLoad(func() {
        // Increment every second — automatically cleared on navigation
        r.SetInterval(func() {
            count.Set(count.Get() + 1)
        }, 1000)

        // One-shot after 5 seconds
        r.SetTimeout(func() {
            println("timeout fired")
        }, 5000)
    })

    return func() core.Node {
        return core.Text(fmt.Sprintf("Count: %d", count.Get()))
    }
}
```

Both return a `*TimerHandle` with a `Clear()` method for manual cancellation. All active timers are automatically cleared when the user navigates away (via `Router.ClearState()`).

## Development Notes

- **core/** is the low-level universal rendering system (focus of current development)
- Always use `core.OnLoad()` for data fetching to support hydration
- DTOs should exclude sensitive fields (passwords, internal IDs)
- Use `core.External: true` in Attrs for links that cross bundle boundaries
- CSRF protection is automatic for all CRUD operations
