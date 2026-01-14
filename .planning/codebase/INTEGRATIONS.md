# External Integrations

**Analysis Date:** 2026-01-14

## APIs & External Services

**Payment Processing:**
- Not detected

**Email/SMS:**
- Not detected

**External APIs:**
- WebSocket Echo Service - Demo/testing only (`example/app/main.go`)
  - URL: `wss://echo.websocket.org`
  - Purpose: WebSocket functionality demonstration
  - No persistent data or authentication

## Data Storage

**Databases:**
- Not detected (in-memory storage in example)
- Example uses Go maps: `example/server/posts.go`

**File Storage:**
- Not detected

**Caching:**
- Browser localStorage for state persistence (`storage/local.go`)
  - Local only, no cloud sync
  - Used for auth state and persistent stores

## Authentication & Identity

**Auth Provider:**
- Custom implementation (`auth/auth.go`)
- Token-based with access/refresh tokens
- No third-party auth provider integration

**OAuth Integrations:**
- Not detected

## Monitoring & Observability

**Error Tracking:**
- Not detected (no Sentry, Bugsnag, etc.)

**Analytics:**
- Not detected

**Logs:**
- Server-side: `log.Printf` to stdout
- No external logging service

## CI/CD & Deployment

**Hosting:**
- Fly.io configured (`fly.toml`)
  - App name: `gux-example`
  - Region: LAX
  - Memory: 256MB
  - HTTP on port 8080 with HTTPS enforcement

**CI Pipeline:**
- Not detected (no GitHub Actions, etc.)

**Container Registry:**
- Docker support via `example/Dockerfile`
- No configured registry push

## Environment Configuration

**Development:**
- No required environment variables
- Build flags only: `GOOS=js GOARCH=wasm`
- Local development via `make dev-tinygo`

**Staging:**
- Not configured

**Production:**
- Fly.io deployment
- Auto-scaling: min_machines_running = 0
- Health checks configured

## Browser APIs (Internal Integrations)

**JavaScript Interop:**
- `syscall/js` for all browser communication

**Storage APIs:**
- localStorage - `storage/local.go`
- sessionStorage - `storage/local.go`

**Network APIs:**
- Fetch API - `fetch/fetch.go`
- WebSocket API - `ws/ws.go`, `state/websocket.go`

**DOM APIs:**
- Document manipulation via `js.Global().Get("document")`
- Event handling via callbacks

## Webhooks & Callbacks

**Incoming:**
- Not detected

**Outgoing:**
- Not detected

## Third-Party Dependencies

**Runtime:**
- `github.com/gorilla/websocket` v1.5.3 - Server-side WebSocket (`go.sum`)

**Build-time:**
- TinyGo compiler (optional, for optimized builds)
- Standard Go toolchain

## Summary

This is a **self-contained framework** with minimal external dependencies:

- **Zero third-party services** - No payment, analytics, auth providers
- **Single external dependency** - Gorilla WebSocket for server-side
- **Browser-native** - Uses standard browser APIs directly
- **Deployment-ready** - Fly.io configuration included

The framework is designed for users to integrate their own services rather than prescribing specific providers.

---

*Integration audit: 2026-01-14*
*Update when adding/removing external services*
