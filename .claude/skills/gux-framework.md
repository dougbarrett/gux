# Gux Framework Development Guide

You are an expert in **Gux**, a full-stack Go framework for building modern web applications with WebAssembly. This skill provides comprehensive knowledge for developing Gux applications.

## Framework Overview

Gux enables writing entire web applications in Go:
- **Frontend**: Compiles to WebAssembly, runs natively in the browser
- **Backend**: Standard Go HTTP server with generated handlers
- **API**: Type-safe clients and servers generated from Go interfaces
- **Components**: 45+ production-ready UI components with Tailwind CSS

## CLI Reference

### Installation

```bash
go install github.com/dougbarrett/gux/cmd/gux@latest
```

### Commands

| Command | Description |
|---------|-------------|
| `gux init --module <path> <name>` | Create new Gux application |
| `gux setup [--tinygo]` | Copy wasm_exec.js from Go/TinyGo installation |
| `gux gen [--dir <api-dir>]` | Generate API client/server code from interfaces |
| `gux build [--tinygo]` | Build WASM module |
| `gux dev [--port <port>] [--tinygo]` | Build and run dev server |
| `gux version` | Show version |
| `gux help` | Show help |

### Project Scaffolding

```bash
# Create new project
gux init --module github.com/youruser/myapp myapp
cd myapp

# Setup WASM runtime
gux setup              # Standard Go (~5MB WASM)
gux setup --tinygo     # TinyGo (~500KB WASM)

# Install dependencies and run
go mod tidy
gux dev
```

#### Generated Project Structure

```
myapp/
├── app/
│   └── main.go           # WASM frontend entry point
├── server/
│   └── main.go           # HTTP server
├── api/
│   ├── types.go          # Shared data types
│   └── example.go        # Example API interface
├── go.mod                # Go module file
├── index.html            # PWA entry point
├── manifest.json         # PWA manifest
├── offline.html          # Offline fallback page
├── service-worker.js     # PWA caching
└── Dockerfile            # Multi-stage Docker build
```

## API Code Generation

### Defining API Interfaces

Use annotations to define type-safe APIs:

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
- Parameters must be `int` or `string` types

### Request Bodies

- Struct types (not primitives) are treated as request bodies
- `context.Context` is always skipped
- Path parameters are extracted, remaining structs become the body

### Generate Code

```bash
gux gen                  # Scans ./api directory
gux gen --dir src/api    # Custom directory
```

Generates:
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

## Component Library

### Initialization

```go
//go:build js && wasm

import "github.com/dougbarrett/gux/components"

func main() {
    components.LoadTailwind()  // Load Tailwind CSS
    components.InitToasts()    // Initialize toast system

    app := components.NewApp("#app")
    app.Mount(/* your components */)

    select {} // Keep WASM running
}
```

### Form Components

```go
// Button
btn := components.Button(components.ButtonProps{
    Text:      "Save",
    Variant:   components.ButtonSuccess,  // Primary, Secondary, Success, Warning, Danger, Info, Ghost
    Size:      components.ButtonLG,       // SM, MD, LG
    OnClick:   handleSave,
})

// Convenience buttons
components.PrimaryButton("Submit", handleSubmit)
components.SecondaryButton("Cancel", handleCancel)
components.DangerButton("Delete", handleDelete)

// Input
input := components.Input(components.InputProps{
    Label:       "Email",
    Type:        components.InputEmail,  // Text, Email, Password, Number, URL
    Placeholder: "you@example.com",
    Value:       "initial@value.com",
    OnChange:    func(value string) { /* handle */ },
})
email := input.Value()
input.SetValue("new@email.com")
input.Focus()
input.Clear()

// Select
sel := components.SimpleSelect("Country", []components.SelectOption{
    {Label: "United States", Value: "us"},
    {Label: "Canada", Value: "ca"},
}, func(value string) { /* handle */ })

// Checkbox & Toggle
cb := components.Checkbox(components.CheckboxProps{
    Label:    "Accept terms",
    Checked:  false,
    OnChange: func(checked bool) { /* handle */ },
})

toggle := components.Toggle(components.ToggleProps{
    Label:    "Enable notifications",
    OnChange: func(enabled bool) { /* handle */ },
})

// DatePicker
picker := components.DatePicker(components.DatePickerProps{
    Label:    "Start Date",
    OnChange: func(date time.Time) { /* handle */ },
})

// Combobox (searchable dropdown)
combo := components.Combobox(components.ComboboxProps{
    Label:       "Assign to",
    Placeholder: "Search users...",
    Options: []components.ComboboxOption{
        {Label: "John Doe", Value: "1", Description: "Engineering"},
    },
    OnChange: func(value string) { /* handle */ },
})

// FormBuilder (dynamic forms)
form := components.NewFormBuilder(components.FormBuilderProps{
    Fields: []components.BuilderField{
        {Name: "email", Type: components.BuilderFieldEmail, Label: "Email",
         Rules: []components.ValidationRule{components.Required, components.Email}},
        {Name: "password", Type: components.BuilderFieldPassword, Label: "Password",
         Rules: []components.ValidationRule{components.Required, components.MinLength(8)}},
    },
    SubmitText: "Create Account",
    OnSubmit: func(values map[string]string) { /* handle */ },
})
```

### Layout Components

```go
// Main Layout with Sidebar and Header
layout := components.Layout(components.LayoutProps{
    Sidebar: components.SidebarProps{
        Title: "My App",
        Items: []components.NavItem{
            {Label: "Dashboard", Path: "/", Icon: "home"},
            {Label: "Posts", Path: "/posts", Icon: "file-text"},
        },
    },
    Header: components.HeaderProps{
        Title: "Dashboard",
        Actions: []js.Value{
            components.Button(components.ButtonProps{Text: "New"}),
        },
    },
})
layout.SetContent(myContent)
layout.SetPageWithHeader("Posts", postsContent)

// Card
card := components.Card(components.CardProps{}, content...)
card := components.TitledCard("Card Title", content...)
card := components.SectionCard("Title", "Description", content...)

// Tabs
tabs := components.Tabs(components.TabsProps{
    Tabs: []components.Tab{
        {Label: "Profile", Content: profileContent},
        {Label: "Settings", Content: settingsContent},
    },
})

// Accordion
accordion := components.Accordion(components.AccordionProps{
    AllowMultiple: true,
    Items: []components.AccordionItem{
        {Title: "Section 1", Content: content1},
    },
})

// Drawer (side panel)
drawer := components.RightDrawer(components.DrawerProps{
    Title:   "Details",
    Content: detailsContent,
})
drawer.Open()
drawer.Close()
```

### Header Components

```go
// UserMenu
menu := components.NewUserMenu(components.UserMenuProps{
    Name:       "John Doe",
    Email:      "john@example.com",
    AvatarSrc:  "/avatar.jpg",
    OnProfile:  func() { router.Navigate("/profile") },
    OnSettings: func() { router.Navigate("/settings") },
    OnLogout:   func() { handleLogout() },
})

// NotificationCenter
nc := components.NewNotificationCenter(components.NotificationCenterProps{
    Notifications: []components.Notification{
        {ID: "1", Title: "New message", Message: "From Alice", Time: "2 min ago", Read: false, Type: "info"},
    },
    OnMarkRead:    func(id string) { /* mark read */ },
    OnMarkAllRead: func() { /* mark all */ },
    OnClear:       func() { /* clear */ },
})
nc.SetNotifications(newNotifications)

// ConnectionStatus (WebSocket indicator)
status := components.NewConnectionStatus(components.ConnectionStatusProps{
    Variant:   components.ConnectionStatusDotVariant,
    Size:      components.ConnectionStatusMD,
    ShowLabel: true,
})
status.BindToWebSocket(wsStore)
```

### Data Display Components

```go
// Table
table := components.Table(components.TableProps{
    Columns: []components.TableColumn{
        {Header: "ID", Key: "id", Width: "w-16"},
        {Header: "Name", Key: "name"},
        {Header: "Status", Key: "status", Render: func(row map[string]any) js.Value {
            return components.Badge(components.BadgeProps{
                Text:    row["status"].(string),
                Variant: components.BadgeSuccess,
            })
        }},
    },
    Data:       tableData,
    Striped:    true,
    Hoverable:  true,
    OnRowClick: func(row map[string]any) { /* handle */ },
})
table.UpdateData(newData)

// Badge
badge := components.Badge(components.BadgeProps{
    Text:    "Active",
    Variant: components.BadgeSuccess,  // Primary, Secondary, Success, Warning, Error, Info
})

// Avatar
avatar := components.Avatar(components.AvatarProps{
    Name:   "John Doe",
    Size:   components.AvatarLG,  // SM, MD, LG
    Status: "online",             // online, away, offline, busy
})

// Pagination
pagination := components.Pagination(components.PaginationProps{
    CurrentPage:  1,
    TotalPages:   10,
    TotalItems:   100,
    ItemsPerPage: 10,
    OnPageChange: func(page int) { /* load page */ },
})

// VirtualList (efficient large lists)
list := components.VirtualList(components.VirtualListProps{
    Items:      largeDataset,
    ItemHeight: 48,
    OnRender: func(item any, index int) js.Value {
        return components.Div("p-2", components.Text(item.(MyType).Name))
    },
})

// Data Export
components.ExportCSV(data, []string{"id", "name", "email"}, "users.csv")
components.ExportJSON(data, "users.json")
components.ExportPDF(data, headers, keys, "report.pdf", components.PDFExportOptions{
    Title: "User Report",
    Orientation: "landscape",
})
```

### Feedback Components

```go
// Modal
modal := components.Modal(components.ModalProps{
    Title:      "Confirm Action",
    CloseOnEsc: true,
    Content:    components.Text("Are you sure?"),
    Footer: components.Div("flex gap-2 justify-end",
        components.Button(components.ButtonProps{Text: "Cancel", OnClick: func() { modal.Close() }}),
        components.Button(components.ButtonProps{Text: "Confirm", OnClick: handleConfirm}),
    ),
})
modal.Open()

// Toast (initialize once with InitToasts())
components.Toast("Operation successful!", components.ToastSuccess)
components.Toast("Something went wrong", components.ToastError)
components.Toast("Please note...", components.ToastInfo)
components.Toast("Be careful!", components.ToastWarning)

// Alert
alert := components.Alert(components.AlertProps{
    Variant: components.AlertWarning,
    Message: "Your session will expire soon.",
})

// ConfirmDialog
dialog := components.ConfirmDanger("Delete Account", "This cannot be undone.", func() {
    deleteAccount()
})
dialog.Open()

// Spinner
spinner := components.Spinner(components.SpinnerProps{
    Size:  components.SpinnerLG,
    Label: "Loading...",
})

// Skeleton loaders
components.SkeletonText(3)
components.SkeletonCard()

// EmptyState
empty := components.NewEmptyState(components.EmptyStateProps{
    Icon:        "📭",
    Title:       "No messages",
    Description: "You don't have any messages yet.",
    ActionLabel: "Compose",
    OnAction:    func() { compose() },
})
```

### Navigation Components

```go
// Router
router := components.NewRouter()
components.SetGlobalRouter(router)

router.Register("/", showHome)
router.Register("/posts", showPosts)
router.Register("/posts/:id", showPost)
router.Start()

router.Navigate("/posts")
currentPath := router.CurrentPath()

// Link
link := components.Link(components.LinkProps{
    Path: "/posts",
    Text: "View Posts",
})

// CommandPalette (Cmd/Ctrl+K)
palette := components.NewCommandPalette(components.CommandPaletteProps{
    Commands: []components.Command{
        {ID: "new", Label: "Create New Post", Category: "Actions", OnExecute: handleNew},
        {ID: "search", Label: "Search", Category: "Navigation", Shortcut: "Ctrl+F", OnExecute: openSearch},
    },
})
palette.RegisterKeyboardShortcut()
```

### Chart Components

```go
// BarChart
chart := components.BarChart(components.ChartProps{
    Data: []components.ChartData{
        {Label: "Jan", Value: 100},
        {Label: "Feb", Value: 150},
    },
    Height:     200,
    ShowLabels: true,
    ShowValues: true,
})

// LineChart, PieChart, DonutChart - same ChartProps interface

// Sparkline (inline mini charts)
components.LineSparkline([]float64{10, 25, 15, 30})
components.BarSparkline([]float64{10, 25, 15, 30})
```

### Utility Components

```go
// Theme toggle
components.ThemeToggle()

// Animations
components.Bounce(element)
components.Shake(element)
components.Pulse(element)

// Tooltip
buttonWithTooltip := components.WithTooltip(button, "Tooltip text", components.TooltipTop)

// Focus trap (for modals)
trap := components.FocusTrap(modalContent)

// Skip links (accessibility)
skipLinks := components.SkipLinks()
```

### Element Helpers

```go
// Create elements
div := components.Div("flex gap-4 p-2", child1, child2)
components.Text("Paragraph text")
components.H1("Heading 1")
components.H2("Heading 2")
components.H3("Heading 3")
components.Section("Section Title", content...)
```

## State Management

### Basic Store

```go
import "github.com/dougbarrett/gux/state"

// Create store with initial value
type AppState struct {
    User  *User
    Count int
    Theme string
}

store := state.New(AppState{Count: 0, Theme: "light"})

// Read state
current := store.Get()

// Update state
store.Update(func(s *AppState) {
    s.Count++
    s.Theme = "dark"
})

// Subscribe to changes
unsubscribe := store.Subscribe(func(s AppState) {
    fmt.Println("Count:", s.Count)
})
defer unsubscribe()
```

### Persistent Store (localStorage)

```go
// Auto-saves to localStorage on every change
userStore := state.NewPersistentStore("currentUser", User{Name: "Guest"})

userStore.Update(func(u *User) {
    u.Name = "John"
})
// localStorage["currentUser"] = {"Name":"John"}

// On page reload, state is restored
```

### Session Store (sessionStorage)

```go
// Clears when browser closes
cartStore := state.NewSessionStore("shoppingCart", Cart{Items: []CartItem{}})
```

### Async Store

```go
// Manages loading/error states
postsStore := state.NewAsync[[]Post]()

postsStore.Load(func() ([]Post, error) {
    return api.GetPosts()
})

if postsStore.IsLoading() { /* show spinner */ }
if postsStore.HasError() { /* show error: postsStore.Err() */ }
posts := postsStore.Data()
```

### Query Cache (SWR Pattern)

```go
// Cached data fetching with stale-while-revalidate
result := state.UseQuery("posts", func() (any, error) {
    return api.GetPosts()
}, state.QueryOptions{
    StaleTime:      5 * time.Minute,
    CacheTime:      10 * time.Minute,
    RefetchOnFocus: true,
})

if result.IsLoading { /* ... */ }
if result.IsSuccess { posts := result.Data.([]Post) }
result.Refetch()

// Cache operations
cache := state.GetQueryCache()
cache.Invalidate("posts")
cache.SetData("posts", updatedPosts)  // Optimistic updates
```

### WebSocket Store

```go
wsStore := state.NewWebSocketStore(state.WebSocketConfig{
    URL: "ws://localhost:8080/ws",
    OnOpen: func() { fmt.Println("Connected") },
    OnMessage: func(data []byte) { /* handle */ },
    ReconnectInterval: 5 * time.Second,
})

wsStore.Connect()
wsStore.SendJSON(MyMessage{Type: "ping"})

wsStore.On("chat.message", func(data []byte) {
    var msg ChatMessage
    json.Unmarshal(data, &msg)
})

// Access state
state := wsStore.State()
// state.Connected, state.Connecting, state.Error
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

## Accessibility

All components are WCAG 2.1 AA compliant:
- Proper ARIA labels and roles
- Full keyboard navigation
- Visible focus indicators
- Color contrast compliance
- Focus traps for modals
- Screen reader support

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl+K` | Open command palette |
| `Cmd/Ctrl+B` | Toggle sidebar |
| `Escape` | Close modal/dropdown |
| `Arrow Up/Down` | Navigate menus |
| `Tab` | Move focus |

## Best Practices

1. **Always use `gux gen`** after modifying API interfaces
2. **Keep state normalized** - flat collections, not nested
3. **Clean up subscriptions** - store and call unsubscribe functions
4. **Use AsyncStore for API data** - handles loading/error states
5. **Leverage query cache** for repeated fetches
6. **Initialize components properly**:
   - Call `LoadTailwind()` first
   - Call `InitToasts()` before using toasts
   - End `main()` with `select {}` to keep WASM running

## Common Patterns

### Page with Data Loading

```go
func renderPage(client *api.PostsClient) js.Value {
    store := state.NewAsync[[]api.Post]()
    store.Load(client.GetAll)

    if store.IsLoading() {
        return components.Spinner(components.SpinnerProps{Label: "Loading..."})
    }
    if store.HasError() {
        return components.Alert(components.AlertProps{
            Variant: components.AlertError,
            Message: store.Err().Error(),
        })
    }

    container := components.Div("space-y-4")
    for _, post := range store.Data() {
        container.Call("appendChild", components.Card(components.CardProps{},
            components.H3(post.Title),
            components.Text(post.Body),
        ))
    }
    return container
}
```

### Form Submission

```go
func createPostForm(client *api.PostsClient) js.Value {
    form := components.NewFormBuilder(components.FormBuilderProps{
        Fields: []components.BuilderField{
            {Name: "title", Type: components.BuilderFieldText, Label: "Title",
             Rules: []components.ValidationRule{components.Required}},
            {Name: "body", Type: components.BuilderFieldTextarea, Label: "Body"},
        },
        SubmitText: "Create Post",
        OnSubmit: func(values map[string]string) {
            _, err := client.Create(api.CreatePostRequest{
                Title: values["title"],
                Body:  values["body"],
            })
            if err != nil {
                components.Toast(err.Error(), components.ToastError)
                return
            }
            components.Toast("Post created!", components.ToastSuccess)
        },
    })
    return form.Element()
}
```
