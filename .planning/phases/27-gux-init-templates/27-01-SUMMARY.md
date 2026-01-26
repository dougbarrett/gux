---
phase: 27-gux-init-templates
plan: 01
subsystem: scaffolding
tags: [cli, templates, gux-init, core-framework]
requires: []
provides: [app.go.tmpl, go.mod.tmpl, models/item.go.tmpl]
affects: [27-02, 27-03]
tech-stack:
  added: []
  patterns: [gux-init-templates, core-framework-scaffold]
key-files:
  created:
    - cmd/gux/templates/app.go.tmpl
    - cmd/gux/templates/models/item.go.tmpl
  modified:
    - cmd/gux/templates/go.mod.tmpl
decisions:
  - id: 27-01-D1
    title: Use core.New() pattern instead of cmd/app
    rationale: Modernize scaffold to use core/ framework
    impact: Generated apps follow examples/minimal pattern
metrics:
  duration: 1 minute
  completed: 2026-01-26
---

# Phase 27 Plan 01: Core Scaffold Templates Summary

**One-liner:** Created app.go, go.mod, and models/item.go templates using core/ framework pattern

## What Was Built

Created three core scaffold templates for `gux init`:

1. **app.go.tmpl** - Main application template
   - Uses core.New() for app initialization
   - Sets up SQLite database with GORM
   - Registers Item model CRUD
   - Defines three Hybrid routes: /, /items, /items/new
   - Uses template placeholders: {{.ModulePath}}, {{.AppName}}

2. **go.mod.tmpl** - Module dependencies template
   - Go 1.24
   - Gux framework with {{.GuxModule}} {{.GuxVersion}} placeholders
   - GORM dependencies (v1.31.1) and SQLite driver (v1.6.0)

3. **models/item.go.tmpl** - Example GORM model
   - Embeds gorm.Model (ID, CreatedAt, UpdatedAt, DeletedAt)
   - Includes Name and Description fields with JSON tags
   - Demonstrates basic model structure

## Technical Implementation

**Pattern used:** Modern core/ framework architecture
- Single app.go file (not cmd/app + cmd/server split)
- Hybrid SSR+WASM routing with app.Routes().Hybrid()
- Generated API client via api.Init(db)
- CRUD endpoint auto-generation via app.CRUD()

**Template placeholders:**
- {{.ModulePath}} - Replaced with module path during scaffold
- {{.AppName}} - Replaced with application name
- {{.GuxModule}} - Gux module path
- {{.GuxVersion}} - Gux version string

## Decisions Made

### D1: Use core.New() pattern instead of cmd/app

**Context:** Previous gux init used outdated cmd/app + cmd/server architecture

**Decision:** Scaffold templates use core.New() and hybrid routing

**Alternatives considered:**
1. Keep cmd/app pattern - rejected (outdated)
2. Create both patterns - rejected (confusing)

**Rationale:**
- examples/minimal demonstrates the modern pattern
- Single app.go is simpler for new users
- Hybrid routing enables SSR + WASM seamlessly

**Impact:**
- Generated apps work out of the box with gux dev
- Users start with best practices immediately
- Consistency across examples and scaffolded apps

## Tasks Completed

| Task | Description | Commit | Files |
|------|-------------|--------|-------|
| 1 | Create app.go.tmpl | 89184fa | cmd/gux/templates/app.go.tmpl |
| 2 | Update go.mod.tmpl | 73db369 | cmd/gux/templates/go.mod.tmpl |
| 3 | Create models/item.go.tmpl | c60dc96 | cmd/gux/templates/models/item.go.tmpl |

## Deviations from Plan

None - plan executed exactly as written.

## Known Issues / Tech Debt

None

## Next Phase Readiness

**Ready for:** 27-02 (page templates), 27-03 (pages/ directory templates)

**Provides:**
- Core app structure template
- Model example template
- Dependency management template

**Dependencies satisfied:** All templates use correct placeholders for scaffold.go substitution

## Testing Notes

Templates are substituted by cmd/gux/scaffold.go during `gux init`.

Will be validated in integration testing when 27-03 completes.

## Performance Characteristics

N/A - Templates are static files

## Security Considerations

None - Templates don't handle sensitive data

## Documentation Impact

None - Templates are internal to CLI

## Future Iterations

Potential enhancements (not in current scope):
- Add DTOs template example
- Add authentication hooks template
- Support PostgreSQL/MySQL options
