# Gux Feature Parity Checklist

Keep this list updated when adding new features to ensure Claude Code has accurate knowledge.

## Core Features

- [x] Node system (Div, Span, P, H1-H6, etc.)
- [x] Attrs (Class, OnClick, OnSubmit, OnChange, etc.)
- [x] Page functions pattern (loader + component)
- [x] State management (StateInt, StateString, StateBool)
- [x] Generic state (core.UseState[T])
- [x] Router methods (Param, Path, Query, User, Navigate)
- [x] OnLoad for data fetching
- [x] Hydration flow (SSR + WASM)
- [x] Fragment (core.Frag)
- [x] Conditional rendering (core.If)

## Routing

- [x] Simple routes (GET, POST)
- [x] Hybrid routes (SSR + WASM)
- [x] Route groups with bundles
- [x] Route parameters (:id)
- [x] URL query parameters (r.Query("name"))
- [x] External links (cross-bundle navigation)
- [x] Protected routes (.Protected())

## CRUD & API

- [x] CRUD generation (app.CRUD)
- [x] DTO mapping (gux tags)
- [x] List/Detail DTOs with preloading
- [x] Brief DTOs (auto-generated for relations in briefs_gen.go)
- [x] Create/Update hooks
- [x] Typed API endpoints (core.API, APIGet, APIDelete)
- [x] APIContext methods
- [x] Error responses (api.NotFound, etc.)
- [x] CRUD query parameter filtering (?order_id=2)
- [x] ListFiltered API client method

## Audit Logging

- [x] WithAuditLog() CRUD option
- [x] AuditEntry GORM model
- [x] Field-level change diffs for updates
- [x] Auto-migration of audit_entries table
- [x] Async write (non-blocking)
- [x] IP address capture (X-Forwarded-For aware)
- [x] Configurable field exclusion

## Authentication

- [x] Session management
- [x] Protected routes
- [x] Login/Logout endpoints
- [x] SessionUser struct
- [x] Role-based access (HasRole, HasAnyRole)
- [x] Auth configuration (EnableAuth)

## Security

- [x] CSRF protection
- [x] CSRF configuration
- [x] Double Submit Cookie pattern

## UI Components (ui/ package)

- [x] Button (variants, sizes)
- [x] Input (types, sizes, error state)
- [x] Select (options, placeholder)
- [x] Checkbox
- [x] Radio
- [x] Textarea
- [x] Switch
- [x] Card (Header, Content, Footer)
- [x] Alert (variants)
- [x] Badge (variants)
- [x] Modal
- [x] Toast
- [x] Tooltip
- [x] Dropdown
- [x] Tabs
- [x] DataTable
- [x] Pagination
- [x] List
- [x] Avatar
- [x] Icon
- [x] Sidebar / SidebarLayout
- [x] Breadcrumb
- [x] Form

## Model Scaffolding

- [x] gux.config.json model definitions
- [x] Field types: string, int, uint, float64, bool, *uint (FK), time.Time, []string
- [x] display config option (controls Brief DTO display field)
- [x] Parent-child entity management (parent, parentField, sidebar)
- [x] Admin page hooks (Set{Model}ListActions, DetailActions, etc.)
- [x] Hook regeneration safety (user hook files preserved)
- [x] Auth preset (password hashing, verification)
- [x] Customizable roles

## CLI Commands

- [x] gux init (interactive and non-interactive)
- [x] gux init --auth, --auth-public, --admin
- [x] gux init --claude
- [x] gux dev (with --go, --watch, --server)
- [x] gux build (with --go, --server) — always runs gux gen first
- [x] gux gen (with --watch, --server)
- [x] gux clean
- [x] gux claude
- [x] gux update (with --check)
- [x] gux help [pattern]
- [x] gux version

## Configuration

- [x] gux.config.json format
- [x] Config loading and saving
- [x] Re-init from config (gux init --config .)

## Generated Code (guxgen/)

- [x] API client generation (List, ListFiltered, Get, Create, Update, Delete)
- [x] WASM entry points
- [x] Bundle separation (app, admin)
- [x] Brief DTOs (briefs_gen.go)
- [x] Build regeneration behavior (gux build always runs gux gen)

## Templates

- [x] Basic app template
- [x] Auth templates (login, register, dashboard)
- [x] Admin templates (layout, dashboard, users)
- [x] DTO templates
- [x] Model templates

---

## How to Update

When adding a new feature:

1. Implement the feature in `core/` or relevant package
2. Add documentation to `gux-framework.md` skill file
3. Update this checklist
4. Consider if CLAUDE.md.tmpl needs updating for scaffolded projects
