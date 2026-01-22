# Project Research Summary

**Project:** Gux v2.0 Core Components
**Domain:** Go/WASM Component Library with SSR + Hydration + Example Applications
**Researched:** 2026-01-22
**Confidence:** HIGH

## Executive Summary

Gux is building a component library on top of its existing core Node system, targeting four example applications (marketing, SaaS dashboard, admin panel, auth flows). The research confirms this is a well-scoped effort: **the existing stack requires no major additions**, the core rendering architecture is sound, and established patterns from shadcn/ui and Radix UI translate well to Go. The component library should use the `core.Node` abstraction (not direct `js.Value`) to maintain SSR compatibility.

The recommended approach is to build components in dependency order: utilities first, then static display components, then form components, then data display, then stateful components (modals, tabs, dropdowns), and finally the four example apps. Each component should follow the existing props struct pattern (`ButtonProps`, `CardProps`) with Tailwind utility classes for styling. The existing CRUD generation, routing, and state management systems provide the foundation; the component library adds the UI layer.

The key risks are (1) WASM memory leaks from unreleased `js.FuncOf` callbacks, (2) WASM binary size growth beyond 5MB, and (3) hydration mismatches from conditional rendering. All three have mitigation strategies: the existing `SetInChangeEvent` pattern addresses focus loss during re-renders, the multi-bundle architecture with TinyGo support addresses size, and the `core.Node` abstraction isolates browser APIs from components.

## Key Findings

### Recommended Stack

The existing Gux stack is complete for this effort. **No new dependencies are required for development.** The stack is Go 1.24.3, GORM 1.31.1, SQLite (via go-sqlite3), Gorilla WebSocket for hot reload, and Tailwind CSS via CDN.

**Core technologies (keep):**
- **Go 1.24.3**: Core language, latest stable version
- **GORM 1.31.1**: Database ORM, works well with CRUD generation
- **Tailwind CSS (CDN)**: Styling framework, acceptable for development

**Production optimizations (defer):**
- **Tailwind CLI v4.0**: CSS tree-shaking for production (after components complete)
- **TinyGo 0.40.1**: 60-80% WASM size reduction (after core components stable)
- **wasm-opt (Binaryen)**: Further WASM binary optimization

**Explicitly not recommended:**
- External chart libraries (existing SVG charts in `components/charts.go` are sufficient, avoid JS interop overhead)
- go-playground/validator (existing form validation is WASM-compatible, external validation is server-side focused)
- DaisyUI/Headless UI (would create inconsistency with existing styled components)

### Expected Features

**Component Library - Must Have (table stakes):**
- Button with variants (primary, secondary, outline, ghost, destructive)
- Form inputs (Input, Textarea, Select, Checkbox, Radio, Switch)
- Card, Badge, Alert, Modal/Dialog
- Dropdown Menu with keyboard navigation
- Tabs, Accordion (collapsible sections)
- Table with sorting and selection
- Pagination, Tooltip, Toast/Sonner
- Skeleton loaders, Spinner/Progress, Avatar, Breadcrumb

**Component Library - Differentiators (already exist in codebase):**
- Command Palette (Cmd+K) - rare in component libraries, major DX win
- Data Table with sort/filter/select
- Date Picker, Combobox
- Charts (bar, line, pie, donut, sparkline)
- Rich Text Editor, Virtual List
- Theme Provider with dark mode

**Example Apps - Table Stakes:**
- Marketing: Home, Features, Pricing, About, Contact pages with Hero, Feature Grid, Testimonials, FAQ, CTA
- SaaS Dashboard: Dashboard overview, Resource CRUD, Settings, Profile with Stats Cards, Charts, Data Tables
- Admin Panel: User Management, Activity Logs, Roles/Permissions, Settings
- Auth Flows: Login, Register, Forgot/Reset Password, Email Verification

**Defer (v2+):**
- Magic link login, Passkey support
- User impersonation in admin
- Real-time updates via WebSocket push (existing SSE/WebSocket infra exists but not required for MVP)

### Architecture Approach

Components should be functions returning `core.Node` with explicit props structs. Three patterns emerge: simple stateless components (Card), variant components with enums (Button with ButtonVariant/ButtonSize), and stateful components requiring `*core.Router` (Modal, Tabs). The recommended directory structure is a single `ui` package with subdirectories by category (layout, form, data, feedback, nav, overlay, action).

**Major components:**
1. **Utilities** (`utils.go`) - `mergeClasses()`, `conditionalAttr()` helpers
2. **Layout Components** (`layout/`) - Card, Container, VStack/HStack, Grid, Divider
3. **Form Components** (`form/`) - Input, Textarea, Select, Checkbox, Radio, Form wrapper
4. **Data Components** (`data/`) - Table compound components, DataTable, Badge, Avatar, List
5. **Feedback Components** (`feedback/`) - Alert, Toast, Spinner, Progress, Skeleton
6. **Navigation Components** (`nav/`) - Breadcrumb, Pagination, Tabs, Menu
7. **Overlay Components** (`overlay/`) - Modal, Drawer, Dropdown, Tooltip, Popover
8. **Action Components** (`action/`) - Button with variants, ButtonGroup, Link, IconButton

**Composition patterns:**
- Children via variadic `...core.Node` (default)
- Named slots via Props (Header, Footer fields)
- Compound components for related sets (Table + TableHead + TableRow + TableCell)
- Render props via generics for data-driven rendering (DataTable[T])

### Critical Pitfalls

1. **Event Handler Memory Leaks (CL-4)** - Each `js.FuncOf` must be released. Current `dom_renderer.go` creates callbacks without cleanup tracking. Implement component cleanup or callback reuse. Phase 2 priority.

2. **WASM Binary Size Explosion (GW-1)** - Standard Go WASM is 5-15MB. Use multi-bundle architecture (already supported via `WithBundle`), implement size budgets in CI, defer TinyGo optimization to production phase.

3. **Props API Fighting Go's Type System (CL-1)** - Avoid `interface{}` and `map[string]any` for props. Use explicit props structs (ButtonProps, InputProps). The existing pattern is correct; maintain it.

4. **Hydration Mismatches (HY-3)** - Server and client must render identical initial content. Avoid `time.Now()` or random values in render. Use state hydration for values that might differ.

5. **Goroutine Blocking (GW-3)** - Blocking operations freeze browser. Use callback patterns for all async work. The existing `AsyncStore[T]` and API callback patterns are correct.

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 1: Component Foundation
**Rationale:** Establish patterns before building components. Wrong patterns baked in early are expensive to fix.
**Delivers:** Utility functions, base primitives (Button, Card, VStack/HStack), props patterns
**Addresses:** Layout primitives, Button variants (table stakes)
**Avoids:** CL-1 (props API), CL-3 (build tag splitting)

### Phase 2: Form Components
**Rationale:** Forms are critical for admin/auth examples. Build on existing `formField` pattern.
**Delivers:** Input, Textarea, Select, Checkbox, Radio, Form wrapper with validation
**Addresses:** Form handling (table stakes), validation patterns
**Avoids:** HY-5 (focus loss during re-render - already solved in core)

### Phase 3: Data Display Components
**Rationale:** Tables and lists needed for SaaS/Admin examples. Requires Phase 1 components.
**Delivers:** Table compound components, DataTable[T], Badge, Avatar, Pagination, Breadcrumb
**Addresses:** Resource listing (SaaS table stakes), data tables (differentiator)
**Avoids:** GW-2 (memory growth with large datasets), GW-5 (reflection in hot paths)

### Phase 4: Interactive Components
**Rationale:** Modals, dropdowns, tabs require state management patterns established in Phase 1.
**Delivers:** Modal, Dropdown, Tabs, Accordion, Toast, Tooltip
**Addresses:** Overlay components (table stakes), feedback components
**Avoids:** CL-4 (event handler leaks - implement cleanup)

### Phase 5: Auth Example App
**Rationale:** Auth needed by SaaS and Admin examples. Simplest stateful example.
**Delivers:** Login, Register, Password Reset, Email Verification pages
**Addresses:** Auth flows (table stakes)
**Avoids:** EX-1 (over-engineered) - keep focused on auth patterns only

### Phase 6: Marketing Example App
**Rationale:** SSR showcase, minimal WASM. Demonstrates static content patterns.
**Delivers:** Marketing site with Hero, Features, Pricing, Contact pages
**Addresses:** Marketing site patterns, SSR value proposition
**Avoids:** EX-3 (unrealistic patterns) - use real page structure

### Phase 7: SaaS Dashboard Example App
**Rationale:** Core framework use case. Requires Phases 1-4 components.
**Delivers:** Dashboard, Resource CRUD, Settings, Profile pages
**Addresses:** SaaS dashboard (primary use case)
**Avoids:** EX-4 (missing "why") - document architectural decisions

### Phase 8: Admin Panel Example App
**Rationale:** Builds on SaaS patterns with admin-specific additions.
**Delivers:** User Management, Activity Logs, Settings pages
**Addresses:** Admin panel patterns
**Avoids:** EX-5 (framework lock-in) - show idiomatic Go patterns

### Phase Ordering Rationale

- **Dependency chain:** Foundation -> Forms -> Data -> Interactive mirrors component complexity and dependencies
- **Example app order:** Auth first (simplest, needed by others) -> Marketing (SSR showcase) -> SaaS (primary use case) -> Admin (most complex)
- **Pitfall mitigation:** Early phases establish cleanup patterns before building stateful components

### Research Flags

**Phases likely needing deeper research during planning:**
- **Phase 4 (Interactive Components):** Focus trap implementation, scroll locking for modals may need core changes or workarounds
- **Phase 7 (SaaS Dashboard):** Real-time updates pattern if pursued, chart interactions

**Phases with standard patterns (skip research-phase):**
- **Phase 1 (Foundation):** Well-established patterns from existing codebase
- **Phase 2 (Forms):** Existing `formField` pattern provides blueprint
- **Phase 5 (Auth):** Standard auth flow patterns, well-documented
- **Phase 6 (Marketing):** Static content, SSR-focused, standard patterns

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Verified against existing codebase, no changes needed |
| Features | HIGH | Verified against shadcn/ui, Radix UI, Tailwind UI patterns |
| Architecture | HIGH | Based on existing codebase patterns, established Go idioms |
| Pitfalls | MEDIUM-HIGH | Verified against Go issues, WASM docs; some require production validation |

**Overall confidence:** HIGH

### Gaps to Address

- **Callback cleanup pattern:** The exact implementation for releasing `js.FuncOf` callbacks needs design during Phase 4 planning. Current code creates callbacks without tracking.
- **Portal-like behavior:** Modals rendered at document body root would need core changes. Workaround (fixed positioning with z-index) is acceptable for v2.0.
- **TinyGo compatibility:** TinyGo optimization deferred, but verify critical code paths compile with TinyGo before committing to production use.
- **Focus trap implementation:** May need `el.Call("focus")` via js interop. Test during Phase 4.

## Sources

### Primary (HIGH confidence)
- Existing Gux codebase analysis - `core/`, `components/`, `examples/minimal/`
- [Go WASM Documentation](https://pkg.go.dev/syscall/js) - blocking behavior, callback patterns
- [shadcn/ui Components](https://ui.shadcn.com/docs/components) - component inventory and API patterns
- [Radix UI Primitives](https://www.radix-ui.com/primitives) - accessibility patterns

### Secondary (MEDIUM confidence)
- [Go Issue #59061](https://github.com/golang/go/issues/59061) - WASM memory not returned to OS
- [Go Issue #74342](https://github.com/golang/go/issues/74342) - WASM memory leak with JS pointers
- [Dagger: Replaced React with Go](https://dagger.io/blog/replaced-react-with-go) - WASM size benchmarks
- [TinyGo WASM Guide](https://tinygo.org/docs/guides/webassembly/wasm/)

### Tertiary (LOW confidence)
- Community patterns for Go component libraries (sparse, this is pioneering work)

---
*Research completed: 2026-01-22*
*Ready for roadmap: yes*
