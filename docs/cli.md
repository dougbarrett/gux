# CLI Reference

The `gux` command-line tool provides scaffolding, code generation, building, and development utilities for Gux applications.

## Installation

```bash
go install github.com/dougbarrett/gux/cmd/gux@latest
```

Verify installation:

```bash
gux version
```

## Commands

| Command | Description |
|---------|-------------|
| `gux init` | Create a new Gux application |
| `gux setup` | Copy wasm_exec.js from Go/TinyGo |
| `gux gen` | Generate API client and server code |
| `gux build` | Build the WASM module |
| `gux dev` | Build and run development server |
| `gux version` | Show version |
| `gux help` | Show help |

---

## gux init

Creates a new Gux application with a complete project structure.

```bash
gux init [--module <module-path>] <appname>
```

### Options

| Flag | Description |
|------|-------------|
| `--module` | Go module path (e.g., `github.com/user/myapp`) |

### Examples

```bash
# Basic (uses appname as module path)
gux init myapp

# With full module path (recommended)
gux init --module github.com/myuser/myapp myapp
```

### Generated Structure

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

### App Name Rules

- Lowercase letters only (`a-z`)
- Numbers (`0-9`)
- Hyphens (`-`) and underscores (`_`)
- No spaces or special characters

### After Initialization

```bash
cd myapp
gux setup         # Copy wasm_exec.js
go mod tidy       # Download dependencies
gux dev           # Start development server
```

---

## gux setup

Copies the `wasm_exec.js` runtime file from your TinyGo (default) or Go installation into the current directory.

```bash
gux setup [--go]
```

### Options

| Flag | Description |
|------|-------------|
| `--go` | Copy from standard Go instead of TinyGo |

### Examples

```bash
# Copy from TinyGo installation (default)
gux setup

# Copy from standard Go installation
gux setup --go
```

### Notes

- Run this command from your project root
- The `wasm_exec.js` file is required to run Go WASM in the browser
- TinyGo is the default (smaller WASM binaries ~500KB)
- Must match your build toolchain (TinyGo or Go)

---

## gux gen

Generates type-safe API client and server code from Go interface definitions.

```bash
gux gen [--dir <api-dir>]
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `--dir` | `api` | Directory containing API interface files |

### Examples

```bash
# Generate from default api/ directory
gux gen

# Generate from custom directory
gux gen --dir ./internal/api
```

### How It Works

1. Scans the specified directory for `.go` files
2. Finds interfaces with the `@client` annotation
3. Generates two files per interface:
   - `*_client_gen.go` — WASM HTTP client
   - `*_server_gen.go` — HTTP handler wrapper

### Interface Annotations

Your API interfaces must include annotations:

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

### Annotation Reference

| Annotation | Location | Description |
|------------|----------|-------------|
| `@client <Name>` | Interface comment | Names the generated client struct |
| `@basepath <path>` | Interface comment | Base URL path for all routes |
| `@route <METHOD> <path>` | Method comment | HTTP method and path |

### Generated Output

For `api/posts.go`:

```
api/
├── posts.go              # Your interface (input)
├── posts_client_gen.go   # Generated WASM client
└── posts_server_gen.go   # Generated HTTP handlers
```

### Usage in Code

**Client (WASM frontend):**

```go
client := api.NewPostsClient(
    api.WithBaseURL("https://api.example.com"),
    api.WithHeader("Authorization", "Bearer token"),
)

posts, err := client.GetAll()
post, err := client.GetByID(123)
```

**Server (Go backend):**

```go
service := &MyPostsService{}
handler := api.NewPostsAPIHandler(service)
handler.RegisterRoutes(mux)
```

See [API Generation](api-generation.md) for complete documentation.

---

## gux build

Builds a production-ready binary with WASM and all static assets embedded.

```bash
gux build [--go]
```

### Options

| Flag | Description |
|------|-------------|
| `--go` | Use standard Go instead of TinyGo (~5MB vs ~500KB) |

### Examples

```bash
# Build with TinyGo (default, smaller WASM ~500KB)
gux build

# Build with standard Go (larger WASM ~5MB, full stdlib)
gux build --go

# Run the production binary
./server
```

### Build Process

1. Compiles `./cmd/app` to WebAssembly (`public/main.wasm`)
2. Builds `./cmd/server` with all `public/` assets embedded
3. Outputs single `./server` binary

### Output

```
Building WASM module...
Built public/main.wasm (0.48 MB) with TinyGo
Building server binary with embedded assets...
Built ./server (1.23 MB) with all assets embedded

Run with: ./server
```

### Single Binary Deployment

The output binary contains everything:
- `main.wasm` (your WASM frontend)
- `wasm_exec.js` (Go/TinyGo WASM runtime)
- `index.html`, `manifest.json`, etc.
- Any CSS, JS, images, PDFs in `public/`

Cache-busting is handled automatically at runtime—the server computes a hash of `main.wasm` and injects it into `index.html` when served.

### Requirements

- Must run from project root (with `cmd/app/` and `cmd/server/` directories)
- Run `gux setup` first to copy `wasm_exec.js`
- TinyGo must be installed for `--tinygo` flag

### Build Size Comparison

| Toolchain | WASM Size | Notes |
|-----------|-----------|-------|
| Standard Go | ~5 MB | Full standard library support |
| TinyGo | ~500 KB | Some stdlib limitations |

### Large Assets

For large files (videos, large images), consider using a CDN instead of embedding them in the binary.

---

## gux dev

Builds the WASM module and starts a development server.

```bash
gux dev [--port <port>] [--go]
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8080` | Port to run the server on |
| `--go` | `false` | Use standard Go instead of TinyGo |

### Examples

```bash
# Run on default port 8080 (uses TinyGo)
gux dev

# Run on custom port
gux dev --port 3000

# Build with standard Go and run
gux dev --go
```

### What It Does

1. Checks for `wasm_exec.js` (run `gux setup` first)
2. Builds the WASM module to `public/main.wasm`
3. Starts the Go server from `./cmd/server` in dev mode
4. Serves static files from filesystem (not embedded) for hot reload

### Requirements

- `cmd/app/` directory with WASM frontend code
- `cmd/server/` directory with Go server code
- `public/wasm_exec.js` (run `gux setup` first)

### Output

```
Building WASM module...
Built public/main.wasm (0.48 MB) with TinyGo

Starting dev server on http://localhost:8080
```

### Dev Mode vs Production

| Mode | Command | Static Files | Cache-busting |
|------|---------|--------------|---------------|
| Dev | `gux dev` | Filesystem (`public/`) | None (reload to update) |
| Prod | `./server` | Embedded in binary | Runtime hash injection |

---

## Workflow

### New Project

```bash
# 1. Create project
gux init --module github.com/myuser/myapp myapp
cd myapp

# 2. Setup runtime
gux setup              # or: gux setup --tinygo

# 3. Install dependencies
go mod tidy

# 4. Generate API code (if you've defined interfaces)
gux gen

# 5. Start developing
gux dev
```

### Daily Development

```bash
# Start dev server (rebuilds WASM automatically, uses TinyGo)
gux dev

# After changing API interfaces
gux gen

# Production build (single binary with embedded assets)
gux build
./server
```

### Adding New APIs

1. Create interface in `api/` with annotations
2. Run `gux gen`
3. Implement the interface in `server/`
4. Use the generated client in `app/`

---

## Troubleshooting

### "wasm_exec.js not found"

Run `gux setup` to copy the file from your TinyGo installation:

```bash
gux setup       # For TinyGo (default)
gux setup --go  # For standard Go
```

### "no app/ directory found"

You're not in a Gux project root. Either:
- `cd` to your project directory
- Run `gux init` to create a new project

### "TinyGo not found"

Install TinyGo:

```bash
# macOS
brew install tinygo

# Linux
wget https://github.com/tinygo-org/tinygo/releases/download/v0.30.0/tinygo_0.30.0_amd64.deb
sudo dpkg -i tinygo_0.30.0_amd64.deb
```

### "No API interface files found"

Ensure your interface files have the `@client` annotation:

```go
// @client MyClient
type MyAPI interface { ... }
```

### Build errors with TinyGo

Some standard library features aren't supported. Either:
- Use `gux build --go` for full standard library compatibility
- Check [TinyGo compatibility](https://tinygo.org/docs/reference/lang-support/)
