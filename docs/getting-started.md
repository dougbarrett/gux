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

```bash
go get github.com/yourusername/gux
```

## Project Setup

### 1. Create Project Structure

```bash
mkdir myapp && cd myapp
go mod init myapp

mkdir -p app api server static
```

Your project structure should look like:

```
myapp/
├── app/           # WASM frontend
│   └── main.go
├── api/           # API definitions
│   ├── posts.go
│   └── types.go
├── server/        # Go backend
│   └── main.go
├── static/        # Static files
│   └── index.html
├── go.mod
└── Makefile
```

### 2. Copy WASM Runtime

Gux requires the Go WASM runtime file:

```bash
# For standard Go
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" static/

# For TinyGo (recommended)
cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" static/
```

### 3. Create HTML Entry Point

```html
<!-- static/index.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>My Gux App</title>
</head>
<body>
    <div id="app"></div>

    <script src="/wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("/main.wasm"), go.importObject)
            .then(result => go.run(result.instance));
    </script>
</body>
</html>
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

//go:generate go run gux/cmd/apigen -source=posts.go -output=posts_client_gen.go

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
go generate ./api/...
```

This creates `api/posts_client_gen.go` with a type-safe HTTP client.

### Step 4: Build Your Frontend

```go
// app/main.go
//go:build js && wasm

package main

import (
    "fmt"
    "myapp/api"
    "gux/components"
    "gux/state"
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
    "gux/server"
    gqapi "gux/api"
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

### Step 6: Create Makefile

```makefile
# Makefile
.PHONY: build dev clean generate setup

setup:
	cp "$$(tinygo env TINYGOROOT)/targets/wasm_exec.js" static/

generate:
	go generate ./api/...

build:
	tinygo build -o static/main.wasm -target wasm -no-debug ./app

dev: build
	go run ./server -port 8080 -dir ./static

clean:
	rm -f static/main.wasm
```

### Step 7: Build and Run

```bash
# First time setup
make setup
make generate

# Build and run
make dev

# Open http://localhost:8080
```

## Next Steps

- [API Generation](api-generation.md) — Learn about all annotation options
- [Components](components.md) — Explore the full component library
- [State Management](state-management.md) — Master reactive state patterns
- [WebSocket](websocket.md) — Add real-time features
- [Deployment](deployment.md) — Deploy to production

## Troubleshooting

### WASM file not loading

- Ensure `wasm_exec.js` is copied to your static directory
- Check browser console for errors
- Verify MIME type is set correctly (should be `application/wasm`)

### Build errors with TinyGo

- Some standard library features aren't supported in TinyGo
- Check [TinyGo compatibility](https://tinygo.org/docs/reference/lang-support/)
- Use `GOOS=js GOARCH=wasm go build` for full compatibility

### API client errors

- Ensure you've run `go generate ./api/...` after changing interfaces
- Check that your server is running and accessible
- Verify CORS is enabled if running on different ports
