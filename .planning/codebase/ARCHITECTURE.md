# Architecture

**Analysis Date:** 2026-01-14

## Pattern Overview

**Overall:** Full-Stack Monolithic Framework with WASM Frontend

**Key Characteristics:**
- Single language (Go) for frontend and backend
- Frontend compiles to WebAssembly
- Type-safe API generation from Go interfaces
- Reactive component-based UI
- Clear separation via build tags (`//go:build js && wasm`)

## Layers

**Component Layer (WASM):**
- Purpose: DOM manipulation and UI rendering
- Contains: 45+ UI components (Button, Input, Modal, Table, Charts, etc.)
- Location: `components/*.go`
- Depends on: State layer, browser APIs via `syscall/js`
- Used by: Application code (`example/app/main.go`)

**State Layer (WASM):**
- Purpose: Reactive state management
- Contains: Generic stores, async loading, persistence, query caching
- Location: `state/store.go`, `state/async.go`, `state/storage.go`, `state/querycache.go`
- Depends on: Browser localStorage
- Used by: Component layer

**API Client Layer (WASM):**
- Purpose: Type-safe HTTP communication
- Contains: Generated HTTP clients from interface definitions
- Location: `example/api/posts_client_gen.go` (generated)
- Depends on: Fetch wrapper (`fetch/fetch.go`)
- Used by: Component layer

**Communication Layer (WASM):**
- Purpose: Network communication
- Contains: Fetch API wrapper, WebSocket client
- Location: `fetch/fetch.go`, `ws/ws.go`, `state/websocket.go`
- Depends on: Browser APIs
- Used by: API clients, real-time features

**Server Layer (Backend):**
- Purpose: HTTP handling and middleware
- Contains: Middleware composition, SPA handler, generated API handlers
- Location: `server/middleware.go`, `server/spa.go`, `example/api/posts_server_gen.go`
- Depends on: Go stdlib `net/http`
- Used by: Application server

**API Utilities Layer (Backend):**
- Purpose: Request/response helpers
- Contains: Error types, query parsing, pagination
- Location: `api/errors.go`, `api/query.go`
- Depends on: Go stdlib
- Used by: Server handlers

## Data Flow

**HTTP Request (WASM → Server):**

1. User triggers action in component
2. Component calls generated API client method (`example/api/posts_client_gen.go`)
3. Client serializes to JSON, calls `fetch.Fetch()` (`fetch/fetch.go`)
4. Server receives request via generated handler (`example/api/posts_server_gen.go`)
5. Handler parses request, calls service implementation (`example/server/posts.go`)
6. Response serialized and returned
7. Client deserializes, component updates state

**WebSocket Event Flow:**

1. WASM client subscribes via `ws.Client.On()` (`ws/ws.go`)
2. Server broadcasts event via `PostsWSHandler.broadcastEvent()` (`example/server/posts_ws.go`)
3. All clients receive typed message
4. Handlers execute, state updates, UI re-renders

**State Management:**
- File-based: Application state in reactive stores
- Stores notify subscribers on update (`state/store.go`)
- Optional persistence to localStorage (`state/storage.go`)

## Key Abstractions

**Store[T]:**
- Purpose: Generic reactive state container
- Location: `state/store.go`
- Pattern: Pub/sub with typed subscriptions
- Examples: `state.NewStore[[]Post](nil)`

**Component:**
- Purpose: UI element with props and rendering
- Location: `components/*.go`
- Pattern: Props struct + factory function returning `js.Value`
- Examples: `components.Button(ButtonProps{...})`, `components.NewTable(TableProps{...})`

**Generated Client:**
- Purpose: Type-safe API access
- Location: `*_client_gen.go` files
- Pattern: Interface methods → HTTP calls
- Examples: `PostsClient.GetAll()`, `PostsClient.Create(req)`

**Middleware:**
- Purpose: HTTP handler composition
- Location: `server/middleware.go`
- Pattern: `func(http.Handler) http.Handler`
- Examples: `server.Logger()`, `server.CORS()`, `server.Chain()`

## Entry Points

**WASM Frontend:**
- Location: `example/app/main.go`
- Build tag: `//go:build js && wasm`
- Triggers: WASM loaded in browser
- Responsibilities: Initialize app, mount components, run event loop

**HTTP Backend:**
- Location: `example/server/main.go`
- Triggers: Server process start
- Responsibilities: Setup routes, middleware, start HTTP listener

**Code Generator:**
- Location: `cmd/apigen/main.go`
- Triggers: `go generate` or direct invocation
- Responsibilities: Parse interfaces, generate client and server code

## Error Handling

**Strategy:** Typed errors with HTTP status codes

**Patterns:**
- API errors: `api.Error` struct with Code, Message, Status (`api/errors.go`)
- Helper constructors: `api.NotFound()`, `api.BadRequest()`, `api.InternalError()`
- JSON response format: `{"error": {"code": "...", "message": "..."}}`

## Cross-Cutting Concerns

**Logging:**
- Server: `server.Logger()` middleware (`server/middleware.go`)
- Console output with request/response timing

**Validation:**
- Form validation rules in components (`components/form.go`, `components/validation.go`)
- Server-side: Manual validation in service implementations

**Authentication:**
- Auth helpers: `auth/auth.go`
- Token storage in localStorage
- Middleware-compatible design

**Theming:**
- Theme management: `components/theme.go`
- Light/dark mode support
- CSS variable injection

---

*Architecture analysis: 2026-01-14*
*Update when major patterns change*
