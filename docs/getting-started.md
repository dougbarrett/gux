# Getting Started

This guide walks you through setting up Gux and building your first application.

## Prerequisites

### Required

- **Go 1.21+** — [Download Go](https://golang.org/dl/)

### Optional (Recommended)

- **TinyGo 0.30+** — For smaller WASM builds (~500KB vs ~5MB)
  ```bash
  # macOS
  brew install tinygo

  # Linux
  wget https://github.com/tinygo-org/tinygo/releases/download/v0.30.0/tinygo_0.30.0_amd64.deb
  sudo dpkg -i tinygo_0.30.0_amd64.deb

  # Windows
  scoop install tinygo
  ```

## Installation

Install the Gux CLI tool:

```bash
go install github.com/dougbarrett/gux/cmd/gux@latest
```

Verify installation:

```bash
gux version
```

## Quick Start

Create a new Gux application with a single command:

```bash
# Create new project
gux init --module github.com/myuser/myapp myapp
cd myapp

# Setup WASM runtime
gux setup              # or: gux setup --tinygo

# Install dependencies
go mod tidy

# Start development server
gux dev
```

Your app is now running at http://localhost:8080

## Project Structure

The `gux init` command creates a complete project structure:

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
├── service-worker.js     # PWA service worker
└── Dockerfile            # Multi-stage Docker build
```

## Building Your First App

### Step 1: Define Your Data Types

```go
// api/types.go
package api

type Post struct {
    ID     int    `json:"id"`
    UserID int    `json:"userId"`
    Title  string `json:"title"`
    Body   string `json:"body"`
}

type CreatePostRequest struct {
    UserID int    `json:"userId"`
    Title  string `json:"title"`
    Body   string `json:"body"`
}
```

### Step 2: Define Your API Interface

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

### Step 3: Generate Client Code

```bash
gux gen
```

This scans `api/` for interfaces with `@client` annotations and generates:
- `api/posts_client_gen.go` — Type-safe HTTP client for WASM
- `api/posts_server_gen.go` — HTTP handler wrapper for the server

### Step 4: Build Your Frontend

```go
// app/main.go
//go:build js && wasm

package main

import (
    "fmt"
    "myapp/api"
    "github.com/dougbarrett/gux/components"
    "github.com/dougbarrett/gux/state"
)

func main() {
    // Initialize UI framework
    components.LoadTailwind()
    components.InitToasts()

    // Create API client
    posts := api.NewPostsClient()

    // Create reactive state
    postsStore := state.NewAsync[[]api.Post]()

    // Setup router
    router := components.NewRouter()
    components.SetGlobalRouter(router)

    // Create layout
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
        },
    })

    // Mount app
    app := components.NewApp("#app")
    app.Mount(layout)

    // Register routes
    router.Register("/", func() {
        layout.SetPageWithHeader("Dashboard",
            components.Card(components.CardProps{},
                components.H2("Welcome to Gux!"),
                components.Text("Your Go-powered web application."),
            ),
        )
    })

    router.Register("/posts", func() {
        layout.SetPageWithHeader("Posts", renderPostsPage(posts, postsStore))
    })

    // Start router
    router.Start()

    // Keep WASM running
    select {}
}

func renderPostsPage(client *api.PostsClient, store *state.AsyncStore[[]api.Post]) js.Value {
    // Load posts
    store.Load(func() ([]api.Post, error) {
        return client.GetAll()
    })

    container := components.Div("space-y-4")

    // Show loading state
    if store.IsLoading() {
        container.Call("appendChild", components.Spinner(components.SpinnerProps{
            Size:  components.SpinnerLG,
            Label: "Loading posts...",
        }))
        return container
    }

    // Show error
    if store.HasError() {
        container.Call("appendChild", components.Alert(components.AlertProps{
            Variant: components.AlertError,
            Message: store.Err().Error(),
        }))
        return container
    }

    // Render posts
    posts := store.Data()
    for _, post := range posts {
        card := components.Card(components.CardProps{},
            components.H3(post.Title),
            components.Text(post.Body),
        )
        container.Call("appendChild", card)
    }

    return container
}
```

### Step 5: Create Your Server

```go
// server/main.go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "net/http"
    "sync"

    "myapp/api"
    "github.com/dougbarrett/gux/server"
    gqapi "github.com/dougbarrett/gux/api"
)

// PostsService implements api.PostsAPI
type PostsService struct {
    mu     sync.RWMutex
    posts  map[int]api.Post
    nextID int
}

func NewPostsService() *PostsService {
    return &PostsService{
        posts:  make(map[int]api.Post),
        nextID: 1,
    }
}

func (s *PostsService) GetAll(ctx context.Context) ([]api.Post, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    posts := make([]api.Post, 0, len(s.posts))
    for _, p := range s.posts {
        posts = append(posts, p)
    }
    return posts, nil
}

func (s *PostsService) GetByID(ctx context.Context, id int) (*api.Post, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    post, ok := s.posts[id]
    if !ok {
        return nil, gqapi.NotFoundf("post %d not found", id)
    }
    return &post, nil
}

func (s *PostsService) Create(ctx context.Context, req api.CreatePostRequest) (*api.Post, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    post := api.Post{
        ID:     s.nextID,
        UserID: req.UserID,
        Title:  req.Title,
        Body:   req.Body,
    }
    s.posts[post.ID] = post
    s.nextID++
    return &post, nil
}

func (s *PostsService) Update(ctx context.Context, id int, req api.CreatePostRequest) (*api.Post, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    if _, ok := s.posts[id]; !ok {
        return nil, gqapi.NotFoundf("post %d not found", id)
    }

    post := api.Post{
        ID:     id,
        UserID: req.UserID,
        Title:  req.Title,
        Body:   req.Body,
    }
    s.posts[id] = post
    return &post, nil
}

func (s *PostsService) Delete(ctx context.Context, id int) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if _, ok := s.posts[id]; !ok {
        return gqapi.NotFoundf("post %d not found", id)
    }
    delete(s.posts, id)
    return nil
}

func main() {
    port := flag.Int("port", 8080, "Server port")
    dir := flag.String("dir", "./static", "Static files directory")
    flag.Parse()

    mux := http.NewServeMux()

    // Create service and handler
    service := NewPostsService()
    handler := api.NewPostsAPIHandler(service)

    // Add middleware
    handler.Use(
        server.Logger(),
        server.CORS(server.CORSOptions{}),
        server.Recover(),
    )
    handler.RegisterRoutes(mux)

    // SPA handler
    spa := server.NewSPAHandler(*dir)
    mux.HandleFunc("/", spa.ServeHTTP)

    addr := fmt.Sprintf(":%d", *port)
    fmt.Printf("Server running at http://localhost%s\n", addr)
    log.Fatal(http.ListenAndServe(addr, mux))
}
```

### Step 6: Build and Run

```bash
# Generate API client/server code
gux gen

# Build and start dev server
gux dev

# Open http://localhost:8080
```

For production builds:

```bash
# Build with TinyGo for smaller output
gux build --tinygo
```

## Next Steps

- [CLI Reference](cli.md) — Full command-line tool documentation
- [API Generation](api-generation.md) — Learn about all annotation options
- [Components](components.md) — Explore the full component library
- [State Management](state-management.md) — Master reactive state patterns
- [WebSocket](websocket.md) — Add real-time features
- [Deployment](deployment.md) — Deploy to production

## Troubleshooting

### WASM file not loading

- Run `gux setup` to copy `wasm_exec.js` to your project
- Check browser console for errors
- Verify MIME type is set correctly (should be `application/wasm`)

### Build errors with TinyGo

- Some standard library features aren't supported in TinyGo
- Check [TinyGo compatibility](https://tinygo.org/docs/reference/lang-support/)
- Use `gux build` (without `--tinygo`) for full compatibility

### API client errors

- Run `gux gen` after changing API interfaces
- Check that your server is running and accessible
- Verify CORS is enabled if running on different ports

### "wasm_exec.js not found"

Run `gux setup` to copy the runtime file:

```bash
gux setup          # For standard Go
gux setup --tinygo # For TinyGo
```
