# Gux Framework

## What This Is

A full-stack Go framework for building modern web applications with WebAssembly. The `core` package provides universal rendering (SSR + WASM hydration), and the component library provides pre-built UI components built on that foundation.

## Core Value

Developers can build complete web applications in Go — from marketing sites to SaaS products — with server-side rendering for SEO and instant loads, plus client-side interactivity without writing JavaScript.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. -->

- ✓ Universal rendering system (SSR + WASM hydration) — core
- ✓ Node-based component architecture — core
- ✓ Hybrid routing with bundle support — core
- ✓ CRUD API generation with DTOs and hooks — core
- ✓ Automatic CSRF protection — core
- ✓ Hot reload development server — cmd/gux

### Active

<!-- Current scope: v2.0 Core Components -->

**Component Library:**
- [ ] Fresh component library built on core's Node system
- [ ] Tailwind CSS styling integration
- [ ] Common layout components (navigation, sidebars, headers, footers)
- [ ] Form components (inputs, selects, checkboxes, buttons)
- [ ] Feedback components (alerts, toasts, modals)
- [ ] Data display components (tables, cards, badges)

**Marketing Example:**
- [ ] Multi-page marketing site (home, about, features, pricing, contact)
- [ ] Responsive navigation with mobile menu
- [ ] Hero sections, feature grids, testimonials, CTAs
- [ ] SEO-optimized SSR pages

**SaaS Example:**
- [ ] Dashboard with charts and metrics
- [ ] CRUD resource management
- [ ] Authenticated routes
- [ ] Settings and profile pages

**Admin Example:**
- [ ] Full admin panel layout
- [ ] User management (list, detail, edit, roles)
- [ ] Content management with CRUD
- [ ] Settings and configuration
- [ ] Activity logs and dashboard

**Auth Example:**
- [ ] Login flow with validation
- [ ] Registration with email verification
- [ ] Forgot password / reset flow
- [ ] Session management patterns

### Out of Scope

<!-- Explicit boundaries. Includes reasoning to prevent re-adding. -->

- Porting old components library — fresh start, architectural mismatch with core
- OAuth providers — adds complexity, defer to future milestone
- 2FA — defer to auth v2
- Real-time features (WebSocket) — separate concern, defer
- Mobile app — web-first focus

## Context

The project has shipped v1.0-v1.2 milestones focused on the old components-based approach. The new `core` package introduces universal rendering (SSR + WASM), fundamentally changing the architecture. This milestone rebuilds the component layer on top of `core`, with four example applications demonstrating real-world use cases.

**Existing foundation:**
- `core/` — universal rendering, routing, state, CRUD, CSRF
- `fetch/` — WASM HTTP client with auto CSRF
- `api/` — error types, query utilities
- `server/` — middleware (CORS, JWT, logging)
- `examples/minimal/` — reference implementation

**Technical context:**
- Go 1.21+ with WASM compilation
- Tailwind CSS for styling
- GORM for database (via CRUD)
- SSR for initial load, WASM hydration for interactivity

## Constraints

- **Architecture**: Components must work with core's Node system (not raw js.Value)
- **Rendering**: All components must render correctly in both SSR and WASM
- **Styling**: Tailwind CSS, no additional CSS frameworks
- **Examples**: Each example should be 5-10 pages, focused demos not complete products

## Key Decisions

<!-- Decisions that constrain future work. Add throughout project lifecycle. -->

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Fresh component library | Core's Node architecture is fundamentally different from old js.Value approach | — Pending |
| Shared foundation across examples | Reduces duplication, shows realistic patterns | — Pending |
| Tailwind CSS only | Consistency, no framework mixing | — Pending |
| Core flows for auth (no OAuth) | Keep scope manageable for v2.0 | — Pending |

---
*Last updated: 2026-01-22 after v2.0 milestone start*
