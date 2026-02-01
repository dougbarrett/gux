# External Integrations

**Analysis Date:** 2026-02-01

## APIs & External Services

**Not detected** - Gux framework itself does not ship with pre-built integrations to external APIs. Applications built with Gux can add integrations via custom typed API endpoints in `core/endpoint.go`.

**Custom API Pattern:**
- Use `core.API()` and `core.APIGet()` for typed endpoints
- Applications can integrate:
  - Third-party HTTP APIs (via standard `net/http`)
  - OAuth/OIDC providers (via custom auth handlers)
  - Payment processors, analytics, etc. (application-specific)

## Data Storage

**Databases:**
- SQLite (default) - File-based, zero-config database
  - Connection: `gorm.io/driver/sqlite` (included)
  - Client: GORM v1.31.1
  - Connection string example: `guxgen/data.db` (file path)
- PostgreSQL - Supported via optional GORM driver (not included by default)
  - Connection: `gorm.io/driver/postgres` (user-added)
  - Client: GORM v1.31.1
  - Connection string: Standard PostgreSQL DSN

**File Storage:**
- Local filesystem only - Assets served from `cmd/server/public/` or embedded binary
- User files - Application-specific storage (no built-in S3, GCS, etc.)
- Media handling - HTML5 video/audio via `core.Video()`, `core.Audio()` with local or external URLs

**Caching:**
- None built-in - Applications can add Redis/Memcached via custom code
- Session store interface (`core.SessionStore`) allows pluggable backends:
  - In-memory (stub implementation)
  - Redis (custom implementation)
  - Database-backed (custom implementation)

## Authentication & Identity

**Auth Provider:**
- Custom session-based (built-in)
  - No OAuth/OIDC integration included by default
  - `core.AuthConfig` configures session store, cookie settings, login redirect path
  - `server/jwt.go` provides JWT claim parsing (for JWT-based auth if needed)

**Implementation:**
- Session cookie-based (`__gux_session` by default, configurable)
- Password hashing via `golang.org/x/crypto` - bcrypt support included
- Auth preset models auto-generate `SetPassword()` and `CheckPassword()` methods
- Session propagation in SSR via `core.SSRSessionSetter` and `core.SSRSessionClearer`

**Password Security:**
- Field: `PasswordHash string` (hidden from API via DTOs)
- Hashing: bcrypt (`golang.org/x/crypto/bcrypt`)
- Verification: Custom login page validates via `CheckPassword()`
- Password field only appears in create/edit forms via auth preset

## Monitoring & Observability

**Error Tracking:**
- None built-in
- Applications can integrate Sentry, Rollbar, etc. via custom middleware in `server/middleware.go`

**Logs:**
- Standard library `log` package for request/panic logging
- Request logger: `server.Logger()` middleware (logs method, path, duration)
- Panic recovery: `server.Recover()` middleware (logs panics, returns 500)
- No structured logging framework included (applications can add)

**Request Tracing:**
- Request ID via `server.RequestID()` middleware (sets `X-Request-ID` header)
- No distributed tracing integration

## CI/CD & Deployment

**Hosting:**
- Self-hosted (recommended) - Run as single binary on any platform
- Fly.io (example config in `fly.toml`) - Dockerfile-based deployment
- Docker (supported) - Gux generates Dockerfile template for apps
- Kubernetes (possible) - Requires container image, no special integration

**CI Pipeline:**
- None built-in
- Example Dockerfile in generated projects supports standard CI systems (GitHub Actions, GitLab CI, etc.)

**Build Process:**
- `gux gen` - Generate API clients, CRUD code, admin pages from models
- `gux build` - Compile WASM (TinyGo/Go), generate Tailwind CSS, embed assets
- `gux dev` - Build + run server with hot reload on file changes
- All commands are self-contained in `cmd/gux/main.go`

## Environment Configuration

**Required env vars (example):**
- `PORT` - Server port (default: 8080)
- `DATABASE_URL` - Database connection (SQLite path or PostgreSQL DSN)
- Custom application secrets (not defined by framework)

**Secrets location:**
- `.env` file (loaded via `core.LoadEnv()`, optional, not versioned)
- Environment variables at deployment time
- No secrets manager integration (applications can add Vault, AWS Secrets Manager, etc.)

## Webhooks & Callbacks

**Incoming:**
- Not built-in - Applications define custom webhook endpoints via typed API

**Outgoing:**
- No built-in outgoing webhooks
- Applications can make HTTP calls to external services using standard `net/http`

## CSRF Protection

**Automatic protection:**
- Double Submit Cookie pattern (default, enabled)
- Meta tag: `<meta name="csrf-token">` (rendered server-side)
- Cookie: `__gux_csrf` (auto-set, configurable)
- Header: `X-CSRF-Token` (auto-added by `fetch` package on POST/PUT/PATCH/DELETE)

**Implementation:**
- `core/csrf.go` - Server-side generation and validation
- `fetch/fetch.go` - Client-side auto-inclusion (WASM only)
- Configuration: `app.EnableCSRF()` or `app.DisableCSRF()`
- No external CSRF service needed

## Media & Rich Content

**Video/Audio:**
- HTML5 `<video>` and `<audio>` elements via `core.Video()`, `core.Audio()`
- VideoJS integration (optional) - Enhanced player from CDN
- Google IMA (optional) - Pre-roll/mid-roll ad insertion via `ui.VideoPlayer`

**Charts:**
- Chart rendering via `github.com/go-analyze/charts` (internal)
- Chart component: `ui.Chart()` - Line, bar, pie charts for dashboards

## Security Headers

**Server Middleware:**
- CORS: `server.CORS()` middleware (configurable allow origin/methods/headers)
- Gzip compression: `server.Gzip()` middleware (transparent)
- Request ID tracking: `server.RequestID()` middleware
- No built-in security headers (CSP, X-Frame-Options, etc.) - applications should add

---

*Integration audit: 2026-02-01*
