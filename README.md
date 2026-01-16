# Gux

A full-stack Go framework for building modern web applications with WebAssembly. Write your entire application in Go — from type-safe API clients to reactive UI components.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Features

- **Type-Safe API Generation** — Define Go interfaces, generate HTTP clients and server handlers automatically
- **45+ UI Components** — Forms, layouts, data display, feedback, and charts with Tailwind CSS
- **WCAG 2.1 AA Accessible** — Screen reader support, keyboard navigation, focus management
- **Command Palette** — Quick actions with Cmd/Ctrl+K
- **Data Export** — CSV, JSON, and PDF export for tables
- **Reactive State Management** — Generic stores, persistence, async loading, and SWR-style query caching
- **WebSocket Support** — Type-safe real-time communication with automatic reconnection
- **PWA Ready** — Installable with offline support
- **Server Utilities** — Middleware composition, SPA handler, CORS, logging, and error handling
- **Go-Powered Frontend** — Compile to WebAssembly, run natively in the browser

## Quick Start

### Prerequisites

- Go 1.21+
- TinyGo 0.30+ (optional, for smaller WASM builds ~500KB vs ~5MB)

### Installation

```bash
# Install the Gux CLI tool
go install github.com/dougbarrett/gux/cmd/gux@latest
```

### Create a New App

```bash
# Scaffold a new application
gux init --module github.com/youruser/myapp myapp

# Setup and run
cd myapp
make setup    # Copy wasm_exec.js from Go installation
go mod tidy   # Download dependencies
make dev      # Build and run at http://localhost:8080
```

This creates a minimal Gux application with:
- `app/main.go` — WASM frontend with router and layout
- `server/main.go` — HTTP server with SPA handler
- `api/` — Example API interface for code generation
- `Makefile` — Build commands (setup, build, dev, clean)
- PWA files — manifest.json, service-worker.js, offline.html

### Generate API Code

```bash
# Generate client and server code from API interfaces
gux gen                  # Scans ./api directory
gux gen --dir src/api    # Custom directory
```

This finds all `.go` files with `@client` annotations and generates type-safe HTTP clients and server handlers.

### Run the Example

```bash
cd example
make setup-tinygo   # First time: copy wasm_exec.js
make dev-tinygo     # Build and start server

# Open http://localhost:8093
```

## How It Works

### 1. Define Your API Interface

```go
// api/posts.go
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
}
```

### 2. Generate Client & Server Code

```bash
gux gen
```

This scans the `api/` directory and generates:
- `posts_client_gen.go` — Type-safe HTTP client for WASM
- `posts_server_gen.go` — HTTP handler with automatic routing

### 3. Build Your Frontend

```go
//go:build js && wasm
package main

import (
    "yourapp/api"
    "github.com/dougbarrett/gux/components"
)

func main() {
    components.LoadTailwind()
    components.InitToasts()

    // Type-safe API client
    posts := api.NewPostsClient()

    // Build reactive UI
    app := components.NewApp("#app")
    app.Mount(
        components.Layout(components.LayoutProps{
            Sidebar: components.SidebarProps{
                Title: "My App",
                Items: []components.NavItem{
                    {Label: "Dashboard", Path: "/", Icon: "home"},
                    {Label: "Posts", Path: "/posts", Icon: "file-text"},
                },
            },
        }),
    )

    // Fetch and display
    allPosts, err := posts.GetAll()
    if err != nil {
        components.Toast(err.Error(), components.ToastError)
        return
    }

    // Use posts...
    select {} // Keep WASM running
}
```

### 4. Compile to WebAssembly

```bash
# Standard Go (larger output, ~5MB)
GOOS=js GOARCH=wasm go build -o main.wasm ./app

# TinyGo (smaller output, ~500KB)
tinygo build -o main.wasm -target wasm -no-debug ./app
```

### 5. Create Your Server

```go
package main

import (
    "net/http"
    "yourapp/api"
    "github.com/dougbarrett/gux/server"
)

func main() {
    mux := http.NewServeMux()

    // Wire up generated handler with your service
    service := NewPostsService()
    handler := api.NewPostsAPIHandler(service)

    // Add middleware
    handler.Use(
        server.Logger(),
        server.CORS(server.CORSOptions{}),
        server.Recover(),
    )
    handler.RegisterRoutes(mux)

    // Serve static files with SPA routing
    spa := server.NewSPAHandler("./static")
    mux.HandleFunc("/", spa.ServeHTTP)

    http.ListenAndServe(":8080", mux)
}
```

## Documentation

| Guide | Description |
|-------|-------------|
| [Getting Started](docs/getting-started.md) | Installation, setup, and first app |
| [API Generation](docs/api-generation.md) | Code generation annotations and usage |
| [Components](docs/components.md) | Complete UI component reference |
| [State Management](docs/state-management.md) | Stores, persistence, and async data |
| [WebSocket](docs/websocket.md) | Real-time communication patterns |
| [Server Utilities](docs/server.md) | Middleware and backend helpers |
| [Keyboard Shortcuts](docs/keyboard-shortcuts.md) | Complete keyboard navigation reference |
| [Accessibility](docs/accessibility.md) | ARIA patterns and a11y guidelines |
| [Deployment](docs/deployment.md) | Docker and production setup |

## Component Library

Gux includes 45+ production-ready UI components:

### Forms
```go
// Text input with validation
input := components.Input(components.InputProps{
    Label:       "Email",
    Type:        components.InputEmail,
    Placeholder: "you@example.com",
    OnChange:    func(value string) { /* validate */ },
})

// Dynamic form builder
form := components.NewFormBuilder(components.FormBuilderProps{
    Fields: []components.BuilderField{
        {Name: "email", Type: components.BuilderFieldEmail, Label: "Email",
         Rules: []components.ValidationRule{components.Required, components.Email}},
        {Name: "password", Type: components.BuilderFieldPassword, Label: "Password",
         Rules: []components.ValidationRule{components.Required, components.MinLength(8)}},
    },
    OnSubmit: func(values map[string]string) { /* handle */ },
})
```

### Data Display
```go
// Interactive table
table := components.Table(components.TableProps{
    Columns: []components.TableColumn{
        {Header: "Name", Key: "name"},
        {Header: "Status", Key: "status", Render: func(row map[string]any) js.Value {
            return components.Badge(components.BadgeProps{
                Text: row["status"].(string),
                Variant: components.BadgeSuccess,
            })
        }},
    },
    Data:       tableData,
    Striped:    true,
    OnRowClick: func(row map[string]any) { /* handle */ },
})

// Charts
chart := components.BarChart(components.ChartProps{
    Data: []components.ChartData{
        {Label: "Jan", Value: 100},
        {Label: "Feb", Value: 150},
        {Label: "Mar", Value: 120},
    },
    Height:     200,
    ShowValues: true,
})
```

### Feedback
```go
// Toast notifications
components.Toast("Post created!", components.ToastSuccess)

// Modal dialogs
modal := components.Modal(components.ModalProps{
    Title: "Confirm Delete",
    Content: components.Text("Are you sure?"),
    Footer: components.Div("flex gap-2",
        components.Button(components.ButtonProps{Text: "Cancel", OnClick: modal.Close}),
        components.Button(components.ButtonProps{Text: "Delete", Variant: components.ButtonDanger}),
    ),
})
modal.Open()
```

### Header Components
```go
// UserMenu with avatar and dropdown
userMenu := components.UserMenu(components.UserMenuProps{
    Name:      "John Doe",
    Email:     "john@example.com",
    AvatarURL: "/avatar.png",
    OnLogout:  func() { /* handle logout */ },
})

// NotificationCenter with real-time updates
notifications := components.NotificationCenter(components.NotificationCenterProps{
    Notifications: notificationList,
    OnMarkRead:    func(id string) { /* mark read */ },
    OnClear:       func() { /* clear all */ },
})

// Connection status indicator for WebSocket state
status := components.ConnectionStatus(components.ConnectionStatusProps{
    Connected: wsClient.IsConnected(),
    Variant:   components.StatusDot,  // or StatusBanner
})
```

### Command Palette
```go
// Command Palette (Cmd/Ctrl+K)
palette := components.CommandPalette(components.CommandPaletteProps{
    Commands: []components.CommandItem{
        {ID: "new", Label: "Create New Post", Category: "Actions", Action: handleNew},
        {ID: "search", Label: "Search", Category: "Navigation", Action: openSearch},
    },
})
```

### Data Export
```go
// Export table data to CSV, JSON, or PDF
exporter := components.DataExport(components.DataExportProps{
    Data:     tableData,
    Columns:  []string{"Name", "Email", "Status"},
    Filename: "users",
    Formats:  []string{"csv", "json", "pdf"},
})
```

### Full Component List

| Category | Components |
|----------|------------|
| **Forms** | Button, Input, TextArea, Select, Checkbox, Toggle, DatePicker, Combobox, FileUpload, FormBuilder |
| **Layout** | Layout, Sidebar, Header, Card, Tabs, Accordion, Drawer |
| **Header** | UserMenu, NotificationCenter, ConnectionStatus |
| **Navigation** | Router, Link, Stepper, CommandPalette |
| **Data** | Table, Badge, Avatar, Breadcrumbs, Pagination, VirtualList, DataExport |
| **Feedback** | Modal, Toast, Alert, Progress, Spinner, Skeleton, Tooltip, EmptyState |
| **Charts** | BarChart, LineChart, PieChart, DonutChart, Sparkline |
| **Utilities** | Theme, Animation, Clipboard, FocusTrap, SkipLinks, Inspector |

## State Management

```go
// Generic reactive store
store := state.New(AppState{Count: 0, User: nil})

// Subscribe to changes
unsubscribe := store.Subscribe(func(s AppState) {
    fmt.Println("Count:", s.Count)
})
defer unsubscribe()

// Update state
store.Update(func(s *AppState) {
    s.Count++
})

// Persistent store (auto-saves to localStorage)
userStore := state.NewPersistentStore("currentUser", User{})

// Async data with loading states
posts := state.NewAsync[[]Post]()
posts.Load(func() ([]Post, error) {
    return api.GetPosts()
})

if posts.IsLoading() {
    // Show spinner
}
if posts.HasError() {
    // Show error
}
data := posts.Data()

// Query caching (SWR pattern)
result := state.UseQuery("posts", fetchPosts, state.QueryOptions{
    StaleTime:      5 * time.Minute,
    RefetchOnFocus: true,
})
```

## WebSocket Support

### Type-Safe Subscriptions

```go
// Subscribe to real-time events (mirrors HTTP client pattern)
sub, err := posts.Subscribe(func(event api.PostEvent) {
    switch event.Type {
    case "created":
        fmt.Println("New post:", event.Post.Title)
    case "updated":
        fmt.Println("Updated:", event.Post.Title)
    case "deleted":
        fmt.Println("Deleted ID:", event.ID)
    }
})
if err != nil {
    log.Fatal(err)
}
defer sub.Close()
```

### Low-Level WebSocket Client

```go
client := ws.NewClient("ws://localhost:8080/ws",
    ws.WithOnOpen(func() { fmt.Println("Connected") }),
    ws.WithOnClose(func(code int, reason string) { fmt.Println("Closed") }),
)
client.Connect()

// Typed message handlers
ws.OnTyped(client, "chat.message", func(msg ChatMessage) {
    fmt.Println(msg.Author, ":", msg.Text)
})

// Send typed messages
client.Send("chat.join", JoinRequest{Room: "general"})

// Request/response pattern
resp, err := ws.RequestTyped[JoinReq, JoinResp](client, "room.join", req)
```

## Server Utilities

### Middleware

```go
// Compose middleware
handler := server.Chain(
    server.Logger(),      // Request logging
    server.CORS(opts),    // Cross-origin support
    server.Recover(),     // Panic recovery
    server.RequestID(),   // X-Request-ID header
)(apiHandler)
```

### Error Handling

```go
// Structured errors with HTTP status codes
if user == nil {
    return nil, api.NotFoundf("user %d not found", id)
}

if !valid {
    return nil, api.BadRequest("invalid email format")
}

// Automatic JSON error responses
// {"error": {"code": "not_found", "message": "user 123 not found"}}
```

### Pagination

```go
func handleList(w http.ResponseWriter, r *http.Request) {
    q := api.Query(r)
    search := q.String("search", "")
    page := q.Pagination() // Reads ?page=1&per_page=20

    items := fetchItems(search, page.Offset, page.PerPage)
    total := countItems(search)

    result := api.NewPaginatedResult(items, page, total)
    json.NewEncoder(w).Encode(result)
}
```

## Accessibility

Gux components are built with WCAG 2.1 AA compliance:

- **Screen Reader Support** — All components have proper ARIA labels, roles, and live regions
- **Keyboard Navigation** — Full keyboard access with visible focus indicators
- **Reduced Motion** — Respects `prefers-reduced-motion` system setting
- **Color Contrast** — Meets WCAG 2.1 AA contrast requirements (4.5:1 minimum)
- **Focus Management** — Modal focus traps, focus restoration, skip links

### Accessibility Testing

Automated accessibility testing with axe-core and Playwright:

```bash
make test-a11y        # Run accessibility tests
make test-a11y-debug  # Run with visible browser
```

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl + K` | Open command palette |
| `Cmd/Ctrl + B` | Toggle sidebar |
| `Escape` | Close modal/dropdown/palette |
| `Enter` | Confirm selection |
| `Arrow Up/Down` | Navigate dropdown/menu items |
| `Arrow Left/Right` | Switch tabs |
| `Home/End` | Jump to first/last tab |
| `Tab` | Move between focusable elements |

## Project Structure

```
gux/
├── api/           # Error handling, query utilities, pagination
├── auth/          # Authentication helpers
├── cmd/gux/       # CLI tool (gux init, gux gen)
├── components/    # 45+ UI components (WASM)
├── example/       # Complete working application
│   ├── app/       # WASM frontend
│   ├── server/    # Go backend
│   ├── api/       # API definitions
│   └── Dockerfile # Production deployment
├── fetch/         # Browser fetch API wrapper
├── server/        # Middleware and SPA handler
├── state/         # Reactive state management
├── storage/       # Data persistence layer
└── ws/            # WebSocket client
```

## Deployment

### Docker

```bash
cd example
make docker       # Build image
make docker-run   # Run locally on :8080
```

The Dockerfile uses multi-stage builds:
1. **TinyGo** — Compiles WASM frontend (~500KB)
2. **Go** — Builds server binary
3. **Alpine** — Minimal production image (~20MB)

See [Deployment Guide](docs/deployment.md) for Kubernetes, fly.io, and other platforms.

## PWA Support

Gux applications can be installed as Progressive Web Apps:

- **Installable** — Add to home screen on mobile and desktop
- **Offline Support** — Service worker caches static assets
- **Asset Caching** — Cache-first strategy for optimal performance

The example application includes:
- `manifest.json` — App metadata, icons, theme colors
- `sw.js` — Service worker with intelligent caching
- Install prompt component with 7-day dismissal cooldown

```bash
# PWA files are in example/server/static/
example/server/static/manifest.json
example/server/static/sw.js
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

## License

MIT License — see [LICENSE](LICENSE) for details.
