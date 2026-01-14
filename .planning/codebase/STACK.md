# Technology Stack

**Analysis Date:** 2026-01-14

## Languages

**Primary:**
- Go 1.24.3 - All application code (`go.mod`)

**Secondary:**
- WebAssembly - Frontend compilation target via Go/TinyGo
- JavaScript - Build support files (`example/wasm_exec.js`)

## Runtime

**Environment:**
- Go WASM runtime (browser-based WebAssembly execution)
- Standard Go compiler for ~5MB WASM builds
- TinyGo 0.30+ for optimized ~500KB WASM builds (`README.md`)
- Node.js not required (pure browser runtime)

**Package Manager:**
- Go modules
- Lockfile: `go.sum` present

## Frameworks

**Core:**
- Custom component framework - 45+ UI components (`components/*.go`)
- Reactive state management - Generic stores with subscriptions (`state/store.go`)
- Type-safe API generation - Custom code generator (`cmd/apigen/main.go`)

**Testing:**
- No test framework configured (0 test files found)
- Go's built-in `testing` package available but unused

**Build/Dev:**
- Go compiler - Standard WASM builds
- TinyGo compiler - Optimized WASM builds
- Make - Build automation (`example/Makefile`)
- Docker - Multi-stage containerization (`example/Dockerfile`)

## Key Dependencies

**Critical:**
- `github.com/gorilla/websocket` v1.5.3 - Server-side WebSocket handling (`go.sum`)

**Infrastructure:**
- `syscall/js` - Browser JavaScript interop (Go stdlib)
- `encoding/json` - JSON serialization (Go stdlib)
- `net/http` - HTTP server (Go stdlib)

## Configuration

**Environment:**
- No environment variables required for core framework
- Configuration via code (no .env files)
- Build flags for WASM: `GOOS=js GOARCH=wasm`

**Build:**
- `go.mod` - Module definition
- `example/Makefile` - Build targets (dev, build-tinygo, docker)
- `example/Dockerfile` - Multi-stage build (TinyGo + Go + Alpine)

## Platform Requirements

**Development:**
- Any platform with Go 1.24+
- Optional: TinyGo 0.30+ for smaller builds
- No external dependencies (self-contained)

**Production:**
- Docker container (Alpine-based)
- Fly.io deployment configured (`fly.toml`)
- Memory: 256MB, CPUs: 1
- HTTP service on port 8080

---

*Stack analysis: 2026-01-14*
*Update after major dependency changes*
