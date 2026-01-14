# Codebase Structure

**Analysis Date:** 2026-01-14

## Directory Layout

```
goquery/
├── api/                # Backend API utilities
├── auth/               # Authentication helpers
├── cmd/                # CLI tools
│   └── apigen/        # API code generator
├── components/         # UI component library (WASM)
├── docs/               # Documentation
├── example/            # Complete example application
│   ├── api/           # API definitions (shared)
│   ├── app/           # WASM frontend entry
│   └── server/        # Backend server
├── fetch/              # Browser fetch wrapper (WASM)
├── server/             # Server utilities
├── state/              # State management (WASM)
├── storage/            # Data persistence
├── ws/                 # WebSocket client (WASM)
├── go.mod              # Module definition
└── README.md           # Documentation
```

## Directory Purposes

**api/**
- Purpose: Backend API utilities
- Contains: Error handling, query parsing, pagination
- Key files: `errors.go`, `query.go`
- Build tag: None (backend only)

**auth/**
- Purpose: Authentication helpers
- Contains: Token management, auth state
- Key files: `auth.go`
- Build tag: `//go:build js && wasm`

**cmd/apigen/**
- Purpose: API code generation CLI
- Contains: AST parser, template generator
- Key files: `main.go` (576 lines)
- Usage: `go run gux/cmd/apigen -source=api.go -output=*_client_gen.go`

**components/**
- Purpose: UI component library (45+ components)
- Contains: Form, layout, data display, feedback, charts
- Key files: `button.go`, `input.go`, `table.go`, `modal.go`, `charts.go`, `router.go`
- Build tag: `//go:build js && wasm`
- Subdirectories: None (flat structure)

**docs/**
- Purpose: Framework documentation
- Contains: Getting started, API generation, components, deployment guides
- Key files: `getting-started.md`, `api-generation.md`, `components.md`, `websocket.md`

**example/**
- Purpose: Complete working example application
- Contains: Full-stack demo with posts CRUD + WebSocket

**example/api/**
- Purpose: Shared API definitions
- Contains: Interface definitions, types, generated code
- Key files: `posts.go` (interface), `types.go`, `posts_client_gen.go`, `posts_server_gen.go`

**example/app/**
- Purpose: WASM frontend application
- Contains: Entry point, routing, page components
- Key files: `main.go`
- Build tag: `//go:build js && wasm`

**example/server/**
- Purpose: Backend HTTP server
- Contains: Service implementation, WebSocket handler
- Key files: `main.go`, `posts.go`, `posts_ws.go`

**fetch/**
- Purpose: Browser fetch API wrapper
- Contains: HTTP request helpers
- Key files: `fetch.go`
- Build tag: `//go:build js && wasm`

**server/**
- Purpose: Backend server utilities
- Contains: Middleware composition, SPA handler
- Key files: `middleware.go`, `spa.go`
- Build tag: None (backend only)

**state/**
- Purpose: Reactive state management
- Contains: Generic stores, async patterns, persistence
- Key files: `store.go`, `async.go`, `storage.go`, `querycache.go`, `websocket.go`
- Build tag: `//go:build js && wasm`

**storage/**
- Purpose: Browser storage abstraction
- Contains: localStorage/sessionStorage wrappers
- Key files: `local.go`
- Build tag: `//go:build js && wasm`

**ws/**
- Purpose: WebSocket client
- Contains: Type-safe WebSocket with request/response
- Key files: `ws.go` (397 lines)
- Build tag: `//go:build js && wasm`

## Key File Locations

**Entry Points:**
- `example/app/main.go` - WASM frontend entry
- `example/server/main.go` - Backend server entry
- `cmd/apigen/main.go` - Code generator CLI

**Configuration:**
- `go.mod` - Module definition (Go 1.24.3)
- `example/Makefile` - Build automation
- `example/Dockerfile` - Container build
- `fly.toml` - Fly.io deployment

**Core Logic:**
- `components/*.go` - UI component library
- `state/store.go` - Reactive state management
- `api/errors.go` - Error handling
- `server/middleware.go` - HTTP middleware

**Code Generation:**
- `cmd/apigen/main.go` - Generator implementation
- `example/api/posts.go` - Input interface with annotations
- `example/api/posts_client_gen.go` - Generated WASM client
- `example/api/posts_server_gen.go` - Generated HTTP handler

**Documentation:**
- `README.md` - Main documentation
- `COMPONENTS.md` - Component reference
- `docs/*.md` - Detailed guides

## Naming Conventions

**Files:**
- `snake_case.go` for multi-word files (`data_display.go`, `focus_trap.go`)
- `lowercase.go` for single-word files (`button.go`, `input.go`)
- `*_gen.go` suffix for generated code (`posts_client_gen.go`)
- `*_test.go` suffix for tests (none present)

**Directories:**
- lowercase singular/plural (`components`, `state`, `api`)
- `cmd/` for CLI tools

**Special Patterns:**
- `//go:build js && wasm` - WASM-only code (first line)
- `//go:generate` - Code generation triggers

## Where to Add New Code

**New Component:**
- Implementation: `components/{name}.go`
- Props struct: `{Name}Props`
- Build tag: `//go:build js && wasm`

**New API Endpoint:**
- Interface: `example/api/{resource}.go` with `@client`, `@route` annotations
- Types: `example/api/types.go`
- Generate: `go generate ./example/api/...`

**New Server Middleware:**
- Implementation: `server/middleware.go` or new file
- Pattern: `func(http.Handler) http.Handler`

**New State Store:**
- Pattern: `state.NewStore[T](initial)` or `state.NewAsyncStore[T]()`
- Location: Application code (`example/app/`)

## Special Directories

**example/**
- Purpose: Reference implementation and demo
- Source: Application code using the framework
- Committed: Yes

**docs/**
- Purpose: Framework documentation
- Source: Markdown files
- Committed: Yes

**.planning/codebase/**
- Purpose: Codebase analysis documents
- Source: Generated by GSD workflow
- Committed: Yes

---

*Structure analysis: 2026-01-14*
*Update when directory structure changes*
