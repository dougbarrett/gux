# GoQuery

A Go framework for building web applications with WebAssembly.

Write your frontend and backend in Go, with type-safe API clients generated from interface definitions.

## Features

- **UI Components** - Buttons, layouts, routing, sidebars, and more
- **Type-Safe API Clients** - Generate HTTP clients from Go interfaces
- **SPA Support** - Server utilities for single-page application routing
- **Tailwind CSS** - Automatic Tailwind loading for styling

## Quick Start

```bash
# Clone and run the example
cd example
make setup  # First time only: copies wasm_exec.js
make dev    # Build WASM and start server

# Open http://localhost:8080
```

## Project Structure

```
goquery/
├── components/     # UI components (WASM)
├── server/         # Server utilities
├── cmd/apigen/     # API client generator
└── example/        # Complete working example
```

## Usage

### 1. Define your API interface

```go
// api/posts.go
package api

//go:generate go run goquery/cmd/apigen -source=posts.go -output=posts_client_gen.go

// @client PostsClient
// @basepath /api/posts
type PostsAPI interface {
    // @route GET /
    GetAll(ctx context.Context) ([]Post, error)

    // @route GET /{id}
    GetByID(ctx context.Context, id int) (*Post, error)

    // @route POST /
    Create(ctx context.Context, req CreatePostRequest) (*Post, error)
}
```

### 2. Generate the client

```bash
go generate ./api/...
```

### 3. Use in your WASM app

```go
// app/main.go
package main

import (
    "context"
    "goquery/components"
    "yourapp/api"
)

func main() {
    components.LoadTailwind()

    posts := api.NewPostsClient()

    // Fetch data
    allPosts, err := posts.GetAll(context.Background())
    // ...
}
```

### 4. Implement the server

```go
// server/main.go
package main

import (
    "net/http"
    "goquery/server"
)

func main() {
    mux := http.NewServeMux()

    // Your API handlers
    mux.HandleFunc("/api/posts", handlePosts)

    // SPA handler for static files
    spa := server.NewSPAHandler("./static")
    mux.HandleFunc("/", spa.ServeHTTP)

    http.ListenAndServe(":8080", mux)
}
```

## Components

### Layout Components

```go
layout := components.NewLayout(components.LayoutProps{
    Sidebar: components.SidebarProps{
        Title: "My App",
        Items: []components.NavItem{
            {Label: "Home", Icon: "🏠", Path: "/"},
            {Label: "Settings", Icon: "⚙️", Path: "/settings"},
        },
    },
    Header: components.HeaderProps{
        Title: "Dashboard",
    },
})
```

### Router

```go
router := components.NewRouter()
components.SetGlobalRouter(router)

router.Register("/", showHome)
router.Register("/settings", showSettings)

router.Start()
```

### Button

```go
btn := components.Button(components.ButtonProps{
    Text: "Click Me",
    OnClick: func() {
        // Handle click
    },
})
```

## API Generator Annotations

- `@client ClientName` - Name for the generated client struct
- `@basepath /api/resource` - Base path for all endpoints
- `@route METHOD /path` - HTTP method and path for each method

Path parameters use `{name}` syntax and must match function parameter names.

## License

MIT
