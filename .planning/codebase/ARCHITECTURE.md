# Architecture

**Analysis Date:** 2026-02-01

## Pattern Overview

**Overall:** Universal Component System with Hybrid Server-Side Rendering + WebAssembly Hydration

Gux uses an abstraction-based architecture where components produce `Node` objects that can be rendered to either HTML (server) or DOM elements (WASM). The framework enables seamless hybrid rendering: pages render on the server for initial load, then hydrate to interactive WASM clients. Authentication, CSRF protection, and routing work identically on both platforms.

**Key Characteristics:**
- **Platform Abstraction**: The `Node` interface is universal; renderers (`HTMLRenderer`, `DOMRenderer`) handle platform-specific output
- **Hybrid Rendering**: Single codebase, dual execution (SSR for initial HTML, WASM for interactivity)
- **Type-Safe API Endpoints**: Compiled endpoints with request/response types (no reflection at call site)
- **Automatic Security**: CSRF protection enabled by default; authentication enforced by default for protected endpoints
- **Bundle-Based Code Splitting**: Routes can be grouped into named WASM bundles (e.g., "admin", "public"), reducing per-section download size
- **DTO Pattern**: Data Transfer Objects control API response shapes; sensitive fields excluded automatically

## Layers

**Rendering Layer (Core Abstractions):**
- Purpose: Abstract UI representation and rendering strategy
- Location: `core/node.go`, `core/renderer.go`, `core/html_renderer.go`, `core/dom_renderer.go`
- Contains: `Node` interface, `RenderResult`, `Element`, `TextNode`, `Renderer` implementations
- Depends on: Nothing (pure abstractions)
- Used by: All components, page functions, element helpers

**Element & Component Layer:**
- Purpose: HTML element helpers and page components
- Location: `core/elements.go` (element builders), `ui/` (component library)
- Contains: `Div()`, `Button()`, `Input()`, `Video()`, `Audio()`, `Canvas()`, etc. Element helpers; UI components like `Button`, `Card`, `Modal`, `DataTable`
- Depends on: Rendering layer (`Node`, `Attrs`)
- Used by: Page functions; applications build UIs from these

**State Management Layer:**
- Purpose: Reactive state with hydration from server to client
- Location: `core/app.go` (`Router` struct), component state methods
- Contains: `Router.StateInt()`, `Router.StateString()`, `Router.StateBool()`, `Router.UseState()`, `OnLoad()` for loaders
- Depends on: Nothing special
- Used by: Page functions for reactive data; hydration transfers state to WASM

**Routing & Page Layer:**
- Purpose: Request routing, page resolution, parameter extraction
- Location: `core/app.go` (route registration, matching), `core/router_server.go`, `core/router_wasm.go` (platform-specific)
- Contains: `Route` struct, `matchRoute()`, `RouteBuilder`, `RouteGroup`, hybrid route handling
- Depends on: Rendering layer, state management
- Used by: App request handlers; page functions receive `Router` context

**API & Endpoint Layer:**
- Purpose: Type-safe HTTP endpoints with automatic CSRF, auth, role validation
- Location: `core/endpoint.go` (typed endpoints), `core/crud.go` (CRUD operations)
- Contains: `APIContext` (request context), `API()`, `APIGet()`, `APIDelete()` helpers; `CRUD()` for model endpoints; `DTOMapper` interface for response transformation
- Depends on: State management (auth/roles), database (`DB` interface)
- Used by: Application setup to register custom and CRUD endpoints

**Security Layer:**
- Purpose: Authentication, authorization, session management, CSRF protection
- Location: `core/auth.go` (session management), `core/csrf.go` (CSRF tokens/validation), `core/endpoint.go` (auth checks)
- Contains: `SessionUser`, `AuthConfig`, `SessionStore` interface, `CSRFConfig`, token generation/validation
- Depends on: HTTP layer (cookies, headers)
- Used by: App request handlers for auth enforcement; endpoint handlers for role checks

**Database & ORM Layer:**
- Purpose: Database abstraction; works with GORM
- Location: `core/crud.go` (CRUD model metadata), `core/audit.go` (audit logging)
- Contains: `CRUDModel` struct, `DB` interface (GORM compatible), `CreateHook`, `UpdateHook`, audit log tracking
- Depends on: Nothing directly (interface-based)
- Used by: App.CRUD() registration; generated API code calls DB methods

**HTTP & Server Layer:**
- Purpose: HTTP request handling, middleware, asset serving
- Location: `core/app.go` (main `Handler()`), `server/middleware.go`, `server/spa.go`
- Contains: Route matching, request dispatch, WASM/CSS asset serving, middleware chain (Logger, CORS, Gzip, Recover)
- Depends on: Routing layer, security layer, rendering layer
- Used by: `app.Run()` starts HTTP server with this handler

**Client (WASM) HTTP Layer:**
- Purpose: Browser-based HTTP client with automatic CSRF
- Location: `fetch/fetch.go` (WASM build only, `//go:build js && wasm`)
- Contains: `Fetch()` function, CSRF token extraction (meta tag or cookie), automatic header injection
- Depends on: Browser fetch API (`syscall/js`)
- Used by: Generated API clients for client-side requests

**Code Generation Layer:**
- Purpose: CLI tool for scaffolding, building, generating API clients
- Location: `cmd/gux/` (main, build, generate, model, etc.)
- Contains: Project init, CRUD scaffolding, WASM compilation, API client generation, Dockerfile generation
- Depends on: Everything above (generates code using patterns from all layers)
- Used by: `gux init`, `gux build`, `gux dev`, `gux gen` commands

## Data Flow

**Server-Side Page Render (Initial Load):**

1. HTTP request arrives at `App.Handler()`
2. Route matching finds matching `Route` for path (e.g., `/users/:id`)
3. Session loaded from cookie if auth enabled
4. Protected route check: redirect to login if unauthenticated
5. `Router` created with params, user, database connection
6. Page function called: `component := route.Handler(router)` → loader runs
7. Loader fetches data via `router.OnLoad()`, fills state
8. Component function called: `component()` → renders to `Node` tree
9. Node tree rendered to HTML string via `HTMLRenderer`
10. HTML with `<meta name="csrf-token">` and state JSON embedded sent to browser

**Client-Side Hydration (WASM):**

1. WASM bundle loads in browser
2. Generated main() code initializes app
3. State from server injected into `Router` state map
4. Page component function called with existing state
5. `DOMRenderer` creates DOM elements from Node tree
6. Event listeners attached (OnClick, OnChange, etc.)
7. Client routes registered; client-side navigation works without page reload

**Client-Side Navigation (Hybrid Update):**

1. User clicks link or calls `router.Navigate()`
2. Page loader endpoint called: `/__gux_api/pages/{path}`
3. Server-side loader runs (fetches data), state returned as JSON
4. WASM component function re-runs with new state
5. DOM replaced with new Node tree via `DOMRenderer`

**API Endpoint Call (from WASM):**

1. Page calls generated API function (e.g., `api.Users.Create()`)
2. CSRF token extracted from meta tag or cookie by `fetch/fetch.go`
3. Request sent with token in `X-CSRF-Token` header
4. Server receives POST/PUT/PATCH/DELETE
5. CSRF middleware validates header == cookie
6. Auth middleware loads session user, checks roles
7. Handler function called with `APIContext`
8. Handler reads params, query, body from context
9. Business logic executes (may use database)
10. Response marshaled to JSON, written to client
11. Client code calls callback with result

**CRUD Operation (Admin List/Detail):**

1. Generated page calls `api.{Model}.List()` or `api.{Model}.Detail(id)`
2. Endpoint reaches server, auth checked
3. Database query built using preloads from `CRUDModel.ListPreloads`
4. Results wrapped in ListDTO via `DTOMapper.FromModel()`
5. DTO response sent to client (sensitive fields excluded)
6. Client page renders list or detail view from DTO

**State Hydration Flow:**

```
Server-side state → JSON in HTML → WASM loads JSON → Router.state → Component re-renders with state
```

On server: `router.StateInt("count", 0).Set(42)` → stored in `router.state["count"]`
HTML includes: `<script>__initialState = {count: 42}</script>`
WASM loads: reads `__initialState`, populates new Router's state map
Component sees same state value when it runs on client

## Key Abstractions

**Node Interface:**
- Purpose: Universal representation of any UI element
- Examples: `TextNode` for text, `Element` for HTML tags, fragments
- Pattern: Implementations provide `Render(renderer Renderer) RenderResult` method

**Renderer Interface:**
- Purpose: Strategy pattern for output format
- Examples: `HTMLRenderer` produces strings, `DOMRenderer` produces `js.Value` DOM nodes
- Pattern: Renderers are stateless; `Render()` calls renderer method that handles type conversion

**Router Context:**
- Purpose: Provides access to request data, state, database during page load and render
- Examples: Params (from `:id`), state (reactive), user (auth), database connection
- Pattern: Server-side receives `*http.Request`, client-side stub for consistency

**DTOMapper Interface:**
- Purpose: Type-safe response transformation; hides sensitive fields
- Examples: `UserList` DTO excludes `PasswordHash`; `PostBrief` for relations
- Pattern: CRUD handler calls `dto.FromModel()` to transform; generated code uses reflection to match fields

**APIContext:**
- Purpose: Convenient request context for endpoint handlers
- Examples: `ctx.ParamUint("id")`, `ctx.User()`, `ctx.DB()`, `ctx.Login()`, `ctx.Logout()`
- Pattern: Wraps `*http.Request`, `http.ResponseWriter`, params map; auth user set by middleware

**CRUDModel & CRUDOption:**
- Purpose: Declarative CRUD registration
- Examples: `WithListDTO()`, `WithDetailDTO()`, `WithCreateHook()`, `WithRoles()`
- Pattern: Builder pattern; options chain to configure before `App.CRUD()` processes

**PageFunc:**
- Purpose: Universal page handler type
- Pattern: `func(r *Router) func() Node` - outer function is loader (server-side), inner is component (both sides)

**SessionStore Interface:**
- Purpose: Pluggable session backend
- Examples: `MemorySessionStore` for dev, Redis/database for production
- Pattern: `Get(id)`, `Set(id, user, ttl)`, `Delete(id)` - app doesn't care about storage

## Entry Points

**App.Handler() for HTTP:**
- Location: `core/app.go` (line ~430)
- Triggers: All HTTP requests to the server
- Responsibilities: Route matching, asset serving (WASM, CSS, wasm_exec.js), page rendering, CRUD endpoints, authentication checks

**Page Function for Rendering:**
- Location: User-defined, e.g., `examples/test-admin/pages/dashboard.go`
- Triggers: `app.Routes().Hybrid("/path", PageFunc)` or server page load
- Responsibilities: Run loader (server-side data fetch), return component function

**Generated main() for WASM:**
- Location: `guxgen/wasm/main_gen.go` or similar
- Triggers: Browser loads WASM bundle
- Responsibilities: Initialize WASM runtime, hydrate state, attach event listeners, set up client-side router

**API Endpoint for Custom Logic:**
- Location: `core.API()` or `core.APIGet()`, etc. registered in app.go
- Triggers: `POST /api/login`, `GET /api/data`, etc.
- Responsibilities: Handle request, return typed response, write JSON

## Error Handling

**Strategy:** Typed API errors with HTTP status codes; runtime panics caught by middleware

**Patterns:**
- API errors: Return `error` from endpoint handler; if `*api.Error`, status code honored; else 500
- Route auth failures: Redirect to login (HTML) or return 401/403 (API)
- CSRF failures: Return 403 Forbidden
- WASM runtime errors: Browser console logs; doesn't crash page
- Database errors: Returned from CRUD handler; endpoint decides HTTP status

## Cross-Cutting Concerns

**Logging:**
- Server: `server.Logger()` middleware logs request method, path, duration
- Client: `console.log()` via WASM; access via browser DevTools
- Audit: `WithAuditLog()` on CRUD tracks user, IP, timestamp, field diffs for changes

**Validation:**
- Client: Form validation via UI components (e.g., email input type, required fields)
- Server: Endpoint handlers validate request, return `api.BadRequest()` if invalid
- CRUD: DTOs implicitly validate (only mapped fields accepted); hooks can add custom validation

**Authentication:**
- Flow: Browser sends session cookie → middleware extracts SessionUser from session store → context for handlers
- Session creation: `ctx.Login(user)` generates session ID, stores in SessionStore, sets cookie
- Session destruction: `ctx.Logout()` deletes from store, clears cookie
- Protected routes: Redirect unauthenticated requests to login page

**CSRF Protection:**
- Server generates token, embeds in `<meta name="csrf-token">` and `__gux_csrf` cookie
- Client fetch code reads token, adds `X-CSRF-Token` header on POST/PUT/PATCH/DELETE
- Server middleware validates header matches cookie
- Automatic on all CRUD operations; custom endpoints inherit protection via fetch layer

**Authorization:**
- Role-based: `user.HasRole("admin")` checks roles in SessionUser.Roles
- Route-level: `.WithRoles("admin")` on routes checks user roles; returns 403 if missing
- Endpoint-level: `.WithRoles("admin")` on API endpoints, CRUD operations
- Default: All protected by default when auth configured; `WithPublic()` or `.Public()` explicitly allows

---

*Architecture analysis: 2026-02-01*
