# Technology Stack

**Analysis Date:** 2026-02-01

## Languages

**Primary:**
- Go 1.24.3 - Full-stack framework (backend, CLI, code generation)

**Secondary:**
- Go (WASM compilation targets) - Client-side rendering via WebAssembly
- JavaScript (WASM glue) - DOM manipulation in browser via `syscall/js` bindings

## Runtime

**Environment:**
- Go 1.24.3 runtime (server-side)
- TinyGo 0.33+ or standard Go (WebAssembly compilation)
- WebAssembly (client-side via browsers)

**Package Manager:**
- Go modules
- Lockfile: `go.mod`, `go.sum` (present)

## Frameworks

**Core:**
- Gux framework (custom full-stack framework)
  - `core/` - Universal rendering system (server + WASM)
  - `server/` - HTTP middleware and utilities
  - `ui/` - Component library (Tailwind-styled UI components)
  - `fetch/` - Browser fetch API wrapper for WASM

**Build/Dev:**
- gux CLI (`cmd/gux/`) - Scaffolding, code generation, build orchestration
  - Model generation from `gux.config.json`
  - WASM bundle compilation (TinyGo or standard Go)
  - Code generation for API clients and CRUD endpoints
  - Development server with hot reload support

**CSS:**
- Tailwind CSS v4+ - Utility-first CSS framework
  - Integrated into build pipeline via `@tailwindcss/cli` or `tailwindcss-ruby`
  - Safelist defined in `ui/tailwind-safelist.txt`
  - Generated to `guxgen/dist/styles.css` via build process

## Key Dependencies

**Critical:**
- `gorm.io/gorm` v1.31.1 - Object-relational mapping for database access
- `gorm.io/driver/sqlite` v1.6.0 - SQLite driver for GORM
- `golang.org/x/crypto` v0.47.0 - Cryptographic operations (bcrypt for passwords, HMAC for CSRF)
- `github.com/charmbracelet/huh` v0.8.0 - Interactive CLI form builder for `gux init`

**Infrastructure:**
- `golang.org/x/mod` v0.32.0 - Go module parsing and manipulation for code generation
- `golang.org/x/text` v0.33.0 - Unicode and text handling
- `github.com/fsnotify/fsnotify` v1.9.0 - File system monitoring (hot reload detection)
- `github.com/go-analyze/charts` v0.5.24 - Chart rendering for data visualization (admin dashboards)

**Charmbracelet TUI stack (CLI):**
- `github.com/charmbracelet/bubbletea` v1.3.6 - Terminal UI framework
- `github.com/charmbracelet/bubbles` v0.21.1 - Reusable TUI components
- `github.com/charmbracelet/lipgloss` v1.1.0 - Terminal styling

**Indirect (transitive):**
- `github.com/mattn/go-sqlite3` v1.14.22 - cgo SQLite bindings (required by GORM SQLite driver)
- `github.com/jinzhu/gorm` v1.9.x - ORM core dependency

## Configuration

**Environment:**
- `.env` file support via `core.LoadEnv(filename string)` - optional, non-blocking if missing
- Environment variables read with `os.Getenv(key)`
- Common patterns:
  - `PORT` - Server listen port (default: 8080)
  - `DATABASE_URL` - Database connection string
  - Database credentials (driver-specific)
  - Session store configuration (if using Redis/custom backends)

**Build:**
- `gux.config.json` - Application configuration
  - Module name, authentication mode, roles, model definitions
  - Generated during `gux init` and updated by `gux model add`
- `go.mod` - Go module definition
- No `go.sum` verification required at runtime (present in repo)

## Platform Requirements

**Development:**
- Go 1.24.3+
- TinyGo 0.33+ OR standard Go (for WASM compilation)
- Node.js or gem (for Tailwind CSS CLI - one of):
  - Global `tailwindcss` binary (v4+)
  - `npm install -D @tailwindcss/cli` (v4)
  - `npm install -D tailwindcss` (v3 compat)
  - `gem install tailwindcss-ruby` (Ruby alternative)
- Make or bash (for scripts if present)

**Production:**
- Linux (Alpine/musl or glibc), macOS, or Windows
- Single Go binary (statically linked with `CGO_ENABLED=0`)
- Port binding (default 8080, configurable)
- Optional: SQLite database (file-based) or external PostgreSQL/MySQL via GORM drivers
- No runtime dependencies on Node, Python, or external services (self-contained)

## Build Artifacts

**WASM Modules:**
- Default bundle: `main.wasm` (~500KB with TinyGo, ~5MB with Go)
- Named bundles: `{name}.wasm` (e.g., `admin.wasm`)
- Location: Embedded in server binary or served from `cmd/server/public/`
- Platform: `GOOS=js GOARCH=wasm`

**JavaScript Runtime:**
- `wasm_exec.js` - Browser WebAssembly glue code
  - Copied from TinyGo root (`targets/wasm_exec.js`) or Go root (`lib/wasm/wasm_exec.js`)
  - Enables syscall/js bindings in WASM

**Styles:**
- `styles.css` - Combined Tailwind output
- Location: `guxgen/dist/styles.css` (generated), embedded in server binary

**Server Binary:**
- Single statically-linked executable
- All assets (WASM, CSS, HTML templates) embedded via `//go:embed`
- Built with: `go build -ldflags="-s -w" -o server ./cmd/server`

---

*Stack analysis: 2026-02-01*
