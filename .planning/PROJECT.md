# Gux Framework

## What This Is

A full-stack Go framework for building modern web applications with WebAssembly. The `core` package provides universal rendering (SSR + WASM hydration), and the `ui` package provides a complete component library built on that foundation.

## Core Value

Developers can build complete web applications in Go — from marketing sites to SaaS products to admin panels — with server-side rendering for SEO and instant loads, plus client-side interactivity without writing JavaScript.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. -->

- Universal rendering system (SSR + WASM hydration) — core
- Node-based component architecture — core
- Hybrid routing with bundle support — core
- CRUD API generation with DTOs and hooks — core
- Automatic CSRF protection — core
- Hot reload development server — cmd/gux
- Component library with 26 components — v2.0
- Layout primitives (Container, VStack, HStack, Grid, Divider) — v2.0
- Form components (Input, Textarea, Select, Checkbox, Radio, Switch, FormField) — v2.0
- Data display (Table, DataTable[T], Badge, Avatar, List, Pagination) — v2.0
- Interactive components (Modal, Dropdown, Tabs, Toast, Tooltip) — v2.0
- Feedback components (Alert, Button with variants/sizes) — v2.0
- Auth example with login, register, password reset flows — v2.0
- Marketing example with responsive nav, hero, pricing, contact — v2.0
- SaaS Dashboard example with CRUD, DataTable, settings — v2.0
- Admin Panel example with user management, activity logs, bulk actions — v2.0

### Active

<!-- Current scope: Planning next milestone -->

(No active requirements — planning next milestone)

### Out of Scope

<!-- Explicit boundaries. Includes reasoning to prevent re-adding. -->

- Porting old components library — completed fresh rebuild in v2.0
- OAuth providers — adds complexity, defer to future milestone
- 2FA — defer to auth v2
- Real-time features (WebSocket) — separate concern, defer
- Mobile app — web-first focus
- Spinner/Progress/Skeleton — v2.0 tech debt, not needed by examples
- Breadcrumb/ButtonGroup/IconButton — v2.0 tech debt, not needed by examples

## Context

### Current State (v2.0)

Shipped v2.0 with complete component library and four example applications:
- **ui/**: 49 component files, ~9,700 LOC
- **examples/**: 4 apps (auth, marketing, saas, admin), ~6,400 LOC
- **core/**: Universal rendering, routing, CRUD, CSRF

### Tech Stack

- Go 1.24.3 with WASM compilation
- Tailwind CSS for styling (CDN)
- GORM for database (via CRUD)
- SQLite for examples
- SSR for initial load, WASM hydration for interactivity

### Known Issues

- 6 P0 components not implemented (Spinner, Progress, Skeleton, Breadcrumb, ButtonGroup, IconButton)
- No OAuth/2FA support yet
- No real-time features yet

## Constraints

- **Architecture**: Components must work with core's Node system (not raw js.Value)
- **Rendering**: All components must render correctly in both SSR and WASM
- **Styling**: Tailwind CSS, no additional CSS frameworks
- **Examples**: Each example should be 5-10 pages, focused demos not complete products

## Key Decisions

<!-- Decisions that constrain future work. Add throughout project lifecycle. -->

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Fresh component library | Core's Node architecture is fundamentally different from old js.Value approach | Good - clean separation |
| Shared foundation across examples | Reduces duplication, shows realistic patterns | Good - consistent patterns |
| Tailwind CSS only | Consistency, no framework mixing | Good - simple styling |
| Core flows for auth (no OAuth) | Keep scope manageable for v2.0 | Good - focused scope |
| Props structs for all components | Type safety, no interface{} | Good - clean API |
| CSS-driven visibility for overlays | SSR compatibility without JS state | Good - hydration-safe |

---

*Last updated: 2026-01-25 after v2.0 milestone*
