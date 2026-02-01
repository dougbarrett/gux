# Codebase Structure

**Analysis Date:** 2026-02-01

## Directory Layout

```
goquery/
├── core/                      # Universal rendering framework
│   ├── node.go               # Node interface, element building blocks
│   ├── renderer.go           # Renderer interface (abstract)
│   ├── html_renderer.go      # Server-side HTML rendering
│   ├── dom_renderer.go       # WASM DOM rendering
│   ├── elements.go           # HTML element helpers (Div, Button, etc)
│   ├── app.go                # App, Router, route registration
│   ├── router_server.go      # Server-specific Router methods (no-ops for timers)
│   ├── router_wasm.go        # WASM-specific Router methods (timers, navigation)
│   ├── endpoint.go           # Typed API endpoints (APIContext, API, APIGet, APIDelete)
│   ├── crud.go               # CRUD API generation
│   ├── auth.go               # Authentication & sessions
│   ├── csrf.go               # CSRF protection
│   ├── audit.go              # Audit logging
│   ├── env.go                # Environment variable loading
│   ├── script_loader_wasm.go # Dynamic JS/CSS loading (WASM)
│   ├── script_loader_stub.go # Stub for non-WASM
│   ├── unmount_stub.go       # Lifecycle hook support
│   ├── *_test.go             # Unit tests (endpoint, csrf, crud, auth, etc)
│   └── *_gen.go              # Generated files (not in source, created by build)
│
├── ui/                        # Component library (UI elements, widgets)
│   ├── alert.go              # Alert component
│   ├── button.go             # Button with variants
│   ├── card.go               # Card layout container
│   ├── modal.go              # Modal dialog
│   ├── datatable.go          # Data table widget
│   ├── dropdown.go           # Dropdown menu
│   ├── form.go               # Form builder
│   ├── icon.go               # Icon wrapper
│   ├── input.go              # Input field
│   ├── checkbox.go           # Checkbox control
│   ├── select.go             # Select/dropdown
│   ├── video_player.go       # Video player with optional VideoJS
│   ├── breadcrumb.go         # Breadcrumb navigation
│   ├── badge.go              # Badge/tag component
│   ├── chart.go              # Chart components
│   ├── *_test.go             # Component tests
│   └── ...                   # Additional UI components
│
├── fetch/                     # WASM HTTP client
│   ├── fetch.go              # Browser fetch wrapper with auto CSRF
│
├── api/                       # API utilities
│   ├── errors.go             # Error types and constructors
│   └── query.go              # Query parameter parsing
│
├── server/                    # Server utilities
│   ├── middleware.go         # Middleware: Logger, CORS, Gzip, Recover, RequestID
│   ├── jwt.go                # JWT authentication (optional)
│   └── spa.go                # SPA static file serving
│
├── cmd/gux/                   # CLI tool for gux commands
│   ├── main.go               # Command routing
│   ├── build.go              # gux build (compiles WASM, generates code)
│   ├── dev.go                # gux dev (dev mode with hot reload)
│   ├── generate.go           # gux gen (code generation)
│   ├── model.go              # gux model (CRUD scaffolding)
│   ├── modelgen.go           # Model code generation logic
│   ├── pagegen.go            # Page template generation
│   ├── apigen.go             # API client generation
│   ├── scaffold.go           # Project scaffolding
│   ├── interactive.go        # Interactive CLI
│   ├── update.go             # Auto-update functionality
│   ├── help.go               # Help text
│   ├── claude.go             # Claude Code integration
│   ├── *_test.go             # CLI tests
│   ├── templates/            # Code generation templates
│   │   ├── auth/             # Auth preset templates
│   │   ├── admin/            # Admin panel templates
│   │   ├── models/           # Model scaffolding
│   │   ├── pages/            # Page scaffolding
│   │   └── claude/           # Claude skill templates
│   └── defaults/             # Default configurations
│
├── examples/
│   ├── test-admin/           # Full example with auth & admin panel
│   │   ├── app.go            # Main app setup
│   │   ├── main.go           # Entry point
│   │   ├── models/           # GORM models (User, etc)
│   │   ├── dto/              # Data transfer objects
│   │   ├── pages/            # Page functions (dashboard, users, etc)
│   │   ├── admin/            # Custom admin page hooks
│   │   ├── guxgen/           # Generated files (not checked in)
│   │   │   ├── api/          # Generated API client
│   │   │   ├── wasm/         # Generated WASM main()
│   │   │   ├── admin/        # Generated admin pages
│   │   │   ├── styles/       # Generated Tailwind CSS
│   │   │   └── dist/         # Compiled assets
│   │   └── go.mod, go.sum
│   │
│   ├── auth/                 # Auth-only example
│   │   └── [similar structure to test-admin]
│   │
│   └── ...                   # Additional examples
│
├── docs/                      # Documentation
│   └── ...
│
├── .planning/                 # GSD planning documents
│   └── codebase/             # Analysis docs (ARCHITECTURE.md, STRUCTURE.md, etc)
│
├── .claude/                   # Claude Code skills
│   └── skills/               # Skill definitions
│
├── gux.go                     # Package root (package gux)
├── go.mod, go.sum            # Module definition
├── fly.toml                   # Fly.io deployment config
└── CLAUDE.md                  # Project instructions (checked into repo)
```

## Directory Purposes

**core/:**
- Purpose: Universal rendering system and framework foundation
- Contains: Node abstraction, renderers, routing, state, CRUD, auth, CSRF
- Key files: `node.go`, `app.go`, `endpoint.go`, `crud.go`, `auth.go`
- Entry point for applications: Import `github.com/dougbarrett/gux/core`

**ui/:**
- Purpose: Pre-built UI components (buttons, cards, modals, tables, etc)
- Contains: Component functions that return `core.Node`, styled with Tailwind
- Key files: `button.go`, `card.go`, `modal.go`, `datatable.go`
- Usage: `ui.Button(props)`, `ui.Card(props)`, etc in pages

**fetch/:**
- Purpose: WASM HTTP client with automatic CSRF token injection
- Contains: Single `Fetch()` function; auto-extracts CSRF from meta/cookie
- Key files: `fetch.go` (build-tagged `//go:build js && wasm`)
- Usage: Called by generated API clients for HTTP requests

**api/:**
- Purpose: API utilities and error types
- Contains: `Error` struct, status code helpers (`NotFound()`, `BadRequest()`), error response writers
- Key files: `errors.go`, `query.go`
- Usage: Endpoint handlers return `api.BadRequest("message")` for HTTP status mapping

**server/:**
- Purpose: HTTP server utilities and middleware
- Contains: Logger, CORS, Gzip compression, panic recovery, request IDs
- Key files: `middleware.go` (Chain, Logger, CORS, Gzip, Recover)
- Usage: `server.Chain()` to compose middleware, e.g., `server.Chain(server.Logger(), server.Gzip())`

**cmd/gux/:**
- Purpose: CLI tool for project scaffolding, building, code generation
- Contains: Command implementations (init, build, dev, gen), code generators
- Key files: `main.go`, `build.go`, `generate.go`, `modelgen.go`
- Usage: `gux init`, `gux build`, `gux dev`, `gux gen`, `gux model`

**examples/:**
- Purpose: Reference implementations showing framework patterns
- Contains: Working projects with auth, CRUD, admin panels
- Key files: Each example has `app.go` (setup), `models/`, `dto/`, `pages/`
- Usage: Copy as starting point for new projects

**guxgen/ (in app projects, not framework):**
- Purpose: Generated code (not in framework repo, created in each app)
- Contains: Compiled WASM, generated API clients, admin pages, styles
- Key files: `wasm/main_gen.go`, `api/*_gen.go`, `admin/*_gen.go`, `styles.css`
- Note: Never edit; regenerated on every `gux build` or `gux gen`

## Key File Locations

**Entry Points:**
- `core/app.go`: `App.Handler()` - HTTP request entry point
- `core/app.go`: `App.Run(addr)` - Start HTTP server
- Generated `guxgen/wasm/main_gen.go`: WASM entry point (generated, not in repo)
- `cmd/gux/main.go`: CLI entry point

**Core Abstractions:**
- `core/node.go`: `Node` interface, `Element`, `TextNode`
- `core/renderer.go`: `Renderer` interface
- `core/html_renderer.go`: HTML string rendering
- `core/dom_renderer.go`: WASM DOM element rendering

**Routing & Pages:**
- `core/app.go`: `Route`, `RouteBuilder`, `RouteGroup`, matching logic
- User pages: `examples/test-admin/pages/*.go` (Dashboard, Users, Login, etc)

**API & CRUD:**
- `core/endpoint.go`: `APIContext`, `API()`, `APIGet()`, `APIDelete()` helpers
- `core/crud.go`: `App.CRUD()`, `CRUDModel`, DTOMapper interface
- Generated `guxgen/api/*_gen.go`: Generated WASM client functions

**Security:**
- `core/auth.go`: `SessionUser`, `AuthConfig`, `SessionStore`, session management
- `core/csrf.go`: CSRF token generation/validation, double-submit cookie pattern
- `fetch/fetch.go`: Automatic CSRF header injection for WASM requests

**Database:**
- `core/crud.go`: CRUD handler logic
- `core/audit.go`: Audit logging for model changes
- User models: `examples/test-admin/models/user.go`
- User DTOs: `examples/test-admin/dto/user.go`

**Components & UI:**
- `ui/button.go`: Button component
- `ui/card.go`: Card container
- `ui/datatable.go`: Data table widget
- `core/elements.go`: HTML helpers (Div, Button, Input, Video, Canvas, etc)

**Testing:**
- `core/*_test.go`: Unit tests for core functionality
- `cmd/gux/*_test.go`: CLI and code generation tests
- `ui/*_test.go`: Component tests

## Naming Conventions

**Files:**
- Lowercase with underscores: `html_renderer.go`, `router_server.go`, `script_loader_wasm.go`
- Platform-specific suffix: `_wasm.go` (WASM only, build tag `//go:build js && wasm`), `_test.go` (tests)
- Generated files: `*_gen.go` suffix (e.g., `user_gen.go`, `dashboard_gen.go`)

**Directories:**
- Lowercase: `core/`, `ui/`, `cmd/`, `examples/`
- Functional grouping: `cmd/gux/` (all CLI code), `examples/` (all examples)

**Functions & Types:**
- PascalCase for exported: `Node`, `Element`, `Renderer`, `HTMLRenderer`, `DOMRenderer`
- camelCase for unexported: `matchRoute()`, `pluralize()`, `csrfTokenStore`
- Handler suffixes: `*Func` for function types (`PageFunc`), `*Handler` for HTTP handlers
- Option/builder suffixes: `*Option` for option functions (`CRUDOption`), `*Builder` for builders (`RouteBuilder`)

**Constants:**
- UPPERCASE: `CSRFTokenLength`, `DefaultSessionCookieName`, `SessionIDLength`, `CSRFCookieName`

## Where to Add New Code

**New Feature:**
- Primary code: `core/{feature}.go` if low-level (routing, rendering), or user project
- Tests: `core/{feature}_test.go`
- Documentation: CLAUDE.md or docs/

**New Component/UI Widget:**
- Implementation: `ui/{component_name}.go` (returns `core.Node`)
- Props struct: In same file, e.g., `type ButtonProps struct { ... }`
- Tests: `ui/{component_name}_test.go`
- Export from `ui/` package

**New Page in Example:**
- Implementation: `examples/{app}/pages/{page_name}.go`
- Function signature: `func PageName(r *core.Router) func() core.Node { ... }`
- Register in `app.go`: `app.Routes().Hybrid("/path", pages.PageName)`

**New API Endpoint:**
- DTO types: `examples/{app}/dto/{name}.go`
- Handler: In `app.go` or separate `api.go`
- Registration: `core.API(app, "POST", "/api/path", handler).Public()` or `.WithRoles()`

**New Middleware:**
- Implementation: `server/{name}.go`
- Type: `func(http.Handler) http.Handler`
- Usage: `server.Chain(middleware1, middleware2)`

**New CLI Command:**
- Implementation: `cmd/gux/{command}.go`
- Router: Add case in `cmd/gux/main.go` switch statement

**Database Model (via scaffolding):**
- Run: `gux model add User --auth` or edit `gux.config.json`
- Generated: `guxgen/models/user_gen.go`, DTOs, admin pages
- Custom: Copy generated to project, customize, update `app.go` imports

## Special Directories

**guxgen/ (Generated, in app projects only):**
- Purpose: Contains all generated code
- Generated: On every `gux build` or `gux gen`
- Committed: No (add to .gitignore)
- Contains: WASM binaries, API clients, admin pages, Tailwind CSS

**admin/ (Custom admin pages, in app projects):**
- Purpose: Custom overrides for generated admin pages
- Generated: Hook setter functions (`admin/hooks_gen.go` is auto-generated)
- Committed: Yes (your custom code)
- Pattern: Copy page from `guxgen/admin/{name}_gen.go`, customize, register in `app.go`

**.planning/codebase/ (GSD Documentation):**
- Purpose: Architecture and implementation guides
- Generated: By `/gsd:map-codebase` command
- Contains: ARCHITECTURE.md, STRUCTURE.md, CONVENTIONS.md, TESTING.md, CONCERNS.md
- Usage: Referenced by `/gsd:plan-phase` and `/gsd:execute-phase` commands

**.claude/skills/ (Claude Code Skills):**
- Purpose: Claude Code skill definitions for project-specific automation
- Format: Skill metadata in JSON
- Usage: Loaded by Claude Code for context-aware code generation

**cmd/gux/templates/:**
- Purpose: Code generation templates for scaffolding
- Types: `auth/`, `admin/`, `models/`, `pages/`, `claude/`
- Files: Go template syntax (`.go` files with `{{.VarName}}` placeholders)
- Usage: `gux model` and `gux init` render these templates

---

*Structure analysis: 2026-02-01*
