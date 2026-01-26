# Getting Started

This guide walks you through setting up Gux and building your first application.

**[See Gux in action](https://gux-demo.production.app.dbb1.dev/)** — Explore the live demo before getting started.

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

# Start development server
gux dev
```

Your app is now running at http://localhost:8080

> **Note:** Default files (index.html, manifest.json, service-worker.js, wasm_exec.js) are automatically injected at build time. To customize the HTML shell, create a `public/` directory with your own files.

## Project Structure

The `gux init` command creates a complete project structure:

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

Default files (index.html, manifest.json, service-worker.js, wasm_exec.js) are injected automatically at build time. To override any of these, create a `public/` directory with your own versions.

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

### Step 4: Build Your Application

For a complete working example with hybrid rendering (SSR + WASM), see the [examples/minimal](https://github.com/dougbarrett/gux/tree/main/examples/minimal) directory.

The minimal example demonstrates:

- **Hybrid rendering** — Server-side rendering with WASM hydration
- **Route groups** — Public and admin routes with separate WASM bundles
- **CRUD with DTOs** — Users and Posts with secure field filtering
- **State management** — Data loading with server-to-client hydration
- **Security hooks** — Password hashing in create/update hooks

Key patterns:

```go
import "github.com/dougbarrett/gux/core"

// Page function with loader and component
func MyPage(r *core.Router) func() core.Node {
    var items []Item

    // OnLoad runs on server for SSR, via API for client navigation
    r.OnLoad(func() {
        // Fetch data
    })

    // Component returns UI
    return func() core.Node {
        count := r.StateInt("count", 0)

        return core.Div(core.Class("container"),
            core.H1(core.Attrs{}, core.Text("My Page")),
            core.Button(core.Attrs{
                OnClick: func() { count.Set(count.Get() + 1) },
            }, core.Text("Increment")),
        )
    }
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

- [API Generation](api-generation.md) — Learn about all annotation options
- [Templates](templates.md) — Page templates and common patterns
- [Server Utilities](server.md) — Middleware and backend helpers
- [Deployment](deployment.md) — Deploy to production

For complete examples, see [examples/minimal](https://github.com/dougbarrett/gux/tree/main/examples/minimal).

## Troubleshooting

### WASM file not loading

- Check browser console for errors
- Verify MIME type is set correctly (should be `application/wasm`)
- Ensure TinyGo is installed if using the default build (or use `--go` flag)

### Build errors with TinyGo

- Some standard library features aren't supported in TinyGo
- Check [TinyGo compatibility](https://tinygo.org/docs/reference/lang-support/)
- Use `gux build --go` or `gux dev --go` for full Go compatibility

### API client errors

- Run `gux gen` after changing API interfaces
- Check that your server is running and accessible
- Verify CORS is enabled if running on different ports

### TinyGo not found

Install TinyGo or use the `--go` flag for standard Go builds:

```bash
gux dev --go      # Use standard Go instead of TinyGo
gux build --go    # Build with standard Go
```
