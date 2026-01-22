# Stack Research: Gux Component Library + Example Apps

**Project:** Gux - Go Full-Stack Framework
**Dimension:** Stack additions for component library and example applications
**Researched:** 2026-01-22
**Overall Confidence:** HIGH

## Executive Summary

Gux already has a solid foundation with comprehensive components in `components/` and core rendering in `core/`. The existing stack is well-suited for the component library and example apps - **no major additions are required**. The key findings:

1. **Keep the current approach** - The existing component pattern (returning `js.Value` directly) and core Node system work well together
2. **No external chart libraries needed** - SVG-based charts already exist in `components/charts.go` and avoid JS interop overhead
3. **Form validation is already implemented** - `components/form.go` has validation rules; no need for `go-playground/validator`
4. **Tailwind via CDN is acceptable for development** - Consider build-time CSS generation for production optimization

## Current Stack (No Changes Needed)

| Technology | Version | Purpose | Status |
|------------|---------|---------|--------|
| Go | 1.24.3 | Core language | Keep |
| GORM | 1.31.1 | Database ORM | Keep |
| SQLite | via go-sqlite3 | Development database | Keep |
| Gorilla WebSocket | 1.5.3 | Hot reload | Keep |
| Tailwind CSS | CDN (v3) | Styling | Keep, see optimization notes |

## Recommended Additions

### 1. Production CSS Build (Optional, Post-MVP)

**What:** Tailwind CSS CLI for production builds
**Version:** v4.0+ (latest as of Jan 2026)
**Why:** Current CDN approach works but has drawbacks:
- Requires internet connection
- Loads full Tailwind (~300KB)
- No tree-shaking

**When to add:** After example apps are feature-complete, before production deployment.

**Installation:**
```bash
# Node-based (recommended)
npm install -D tailwindcss@latest

# Or standalone CLI (no Node required)
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
chmod +x tailwindcss-macos-arm64
```

**Confidence:** HIGH - [Tailwind v4 documentation](https://tailwindcss.com/docs/upgrade-guide) confirms CSS-first configuration and standalone CLI.

### 2. WASM Size Optimization (Production)

**What:** TinyGo compiler + wasm-opt
**Version:** TinyGo 0.40.1, wasm-opt (Binaryen) latest
**Why:** Standard Go WASM binaries are large (~2-5MB). TinyGo can reduce to ~100-500KB.

**Trade-offs:**
- TinyGo has limited standard library support
- May require code adjustments (avoid `fmt`, `reflect` in WASM)
- Current architecture uses `syscall/js` which TinyGo supports

**When to add:** After core components are stable, as optimization pass.

**Installation:**
```bash
# TinyGo
brew install tinygo

# wasm-opt (Binaryen)
brew install binaryen
```

**Build command:**
```bash
tinygo build -o app.wasm -target=wasm -no-debug ./main.go
wasm-opt -Oz -o app-opt.wasm app.wasm
```

**Confidence:** HIGH - [TinyGo WASM guide](https://tinygo.org/docs/guides/webassembly/wasm/) and [Binaryen optimization](https://github.com/WebAssembly/binaryen)

## Integration Points with Existing Stack

### Component Library + Core Node System

The existing patterns work well together. Two approaches coexist:

1. **`core/` Node-based components** - Return `core.Node`, render to HTML (SSR) or DOM (WASM)
   ```go
   func MyComponent(attrs core.Attrs) core.Node {
       return core.Div(attrs, core.Text("Hello"))
   }
   ```

2. **`components/` WASM-only components** - Return `js.Value` directly
   ```go
   func Button(props ButtonProps) js.Value {
       btn := document.Call("createElement", "button")
       // ...
       return btn
   }
   ```

**Recommendation:** The component library should use **approach #1** (core Node system) for SSR support. WASM-only components like modals, tooltips can continue with approach #2.

### Tailwind Integration

Current approach in `components/tailwind.go`:
```go
script.Set("src", "https://cdn.tailwindcss.com")
```

This works for development. For production:
1. Extract used classes from Go source
2. Generate minimal CSS with Tailwind CLI
3. Embed in Go binary or serve as static file

**No changes needed for component library development.**

### Chart/Visualization (Already Complete)

The existing `components/charts.go` provides:
- `BarChart` (horizontal/vertical)
- `LineChart` (SVG-based)
- `PieChart` / `DonutChart`
- `Sparkline` (in `sparkline.go`)

**Why this is better than external libraries:**
- No JavaScript dependency
- No WASM<->JS interop overhead (syscall/js calls are slow)
- Pure SVG renders on both SSR and WASM
- Smaller bundle size

**Confidence:** HIGH - Verified by reading existing implementation.

### Form Handling (Already Complete)

The existing `components/form.go` provides:
- `ValidationRule` type with built-in rules (Required, Email, MinLength, MaxLength, Pattern)
- `Form` component with field validation
- Error display with accessibility (ARIA attributes)
- Server-side error injection via `SetFieldError`

**Why NOT add go-playground/validator:**
- It's designed for server-side struct validation
- Would add dependency without benefit for WASM client-side forms
- Current regex-based validation is WASM-compatible
- Server validation should happen in CRUD hooks, not client

**Confidence:** HIGH - Verified by reading existing implementation.

## NOT Recommended

### External Chart Libraries

| Library | Why Not |
|---------|---------|
| Chart.js | Requires JS interop, adds ~200KB, syscall/js overhead |
| go-echarts | Server-side only, generates static HTML, no interactivity |
| Plotters (Rust) | Wrong language, would need embedding |

**Existing SVG charts are sufficient** for dashboards. They're lightweight, SSR-compatible, and avoid JS interop performance issues.

### Heavy Validation Libraries

| Library | Why Not |
|---------|---------|
| go-playground/validator v10 | Server-side focused, reflection-heavy, unnecessary in WASM |
| ozzo-validation | Same issues |

**Client-side validation needs:**
- Simple string checks (length, pattern, required)
- Already implemented in `components/form.go`

**Server-side validation needs:**
- Should happen in CRUD hooks (already supported)
- Can use validator library there if needed (server code, not WASM)

### DaisyUI / Headless UI

| Library | Why Not |
|---------|---------|
| DaisyUI | CSS-only, but requires Tailwind plugin config; adds complexity |
| Headless UI | React/Vue focused, no Go support |

**Gux already has styled components.** Adding another component library creates inconsistency and maintenance burden.

### gotailwindcss (Pure Go Tailwind)

| Library | Why Not |
|---------|---------|
| gotailwindcss | Incomplete implementation, not actively maintained |

**Use official Tailwind CLI** for production builds when needed.

## WASM Size Considerations

Current WASM binary considerations:

| Factor | Impact | Mitigation |
|--------|--------|------------|
| `fmt` package | +400KB | Use `strconv` for number formatting |
| `reflect` package | +200KB | Unavoidable for routing state |
| `regexp` package | +100KB | Used for validation; acceptable |
| Standard Go runtime | +1-2MB | TinyGo for production |

**For development:** Standard Go compiler is fine, fast iteration is more valuable.

**For production:** Consider TinyGo + wasm-opt for 60-80% size reduction.

## Example App Technology Choices

For the four example apps, use the existing stack:

| App | Database | Auth | Special Needs |
|-----|----------|------|---------------|
| Marketing | None | None | Static pages, SEO (SSR handles this) |
| SaaS Dashboard | SQLite | Session-based | Charts (existing), tables (existing) |
| Admin Panel | SQLite | Session-based | CRUD (existing), forms (existing) |
| Auth Flows | SQLite | Sessions + CSRF | Login/register forms (existing) |

**No additional libraries needed.**

## Versions Summary

| Component | Current | Recommended | Notes |
|-----------|---------|-------------|-------|
| Go | 1.24.3 | Keep | Latest stable |
| GORM | 1.31.1 | Keep | Works well |
| Tailwind | CDN v3 | v4.0 CLI (production only) | CSS-first config in v4 |
| TinyGo | - | 0.40.1 (production only) | WASM optimization |
| wasm-opt | - | Latest (production only) | Binary size reduction |

## Installation Commands (When Needed)

```bash
# Production CSS (when ready)
npm init -y
npm install -D tailwindcss@latest

# WASM optimization (when ready)
brew install tinygo binaryen
```

## Sources

- [TinyGo Releases](https://github.com/tinygo-org/tinygo/releases) - v0.40.1 latest
- [TinyGo WASM Guide](https://tinygo.org/docs/guides/webassembly/wasm/)
- [Tailwind CSS v4](https://tailwindcss.com/blog/tailwindcss-v4) - CSS-first configuration
- [syscall/js Performance](https://github.com/golang/go/issues/32591) - Known overhead issues
- [Binaryen wasm-opt](https://github.com/WebAssembly/binaryen) - WASM optimization
- [go-playground/validator](https://github.com/go-playground/validator) - v10.30.1 latest (not recommended for WASM)

## Confidence Assessment

| Area | Confidence | Reason |
|------|------------|--------|
| Keep existing stack | HIGH | Verified by code review of existing components |
| No chart library needed | HIGH | Existing SVG charts verified functional |
| No validation library needed | HIGH | Existing form validation verified complete |
| TinyGo for production | MEDIUM | Requires testing with actual codebase |
| Tailwind CLI for production | HIGH | Well-documented, straightforward migration |
