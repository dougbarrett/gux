# Requirements: Gux Framework v2.2

**Defined:** 2026-01-26
**Core Value:** Developers can scaffold working apps with `gux init` that use the modern core/ framework

## v2.2 Requirements

Requirements for the gux init modernization milestone.

### Templates

- [ ] **TMPL-01**: app.go template creates app with core.New(), SetDB(), SetTitle(), CRUD(), Routes(), Run()
- [ ] **TMPL-02**: go.mod template includes gux, gorm, and sqlite dependencies
- [ ] **TMPL-03**: models/item.go template provides example GORM model with ID, Name, Description, timestamps
- [ ] **TMPL-04**: pages/home.go template renders landing page with ui.Card, links to /items
- [ ] **TMPL-05**: pages/items.go template shows list with OnLoad(), API call, ui.Table
- [ ] **TMPL-06**: pages/item_new.go template shows form with state, validation, API submission, navigation

### Scaffold Process

- [ ] **PROC-01**: `gux init` creates all required files from templates
- [ ] **PROC-02**: `gux init` runs `gux gen` automatically after file creation
- [ ] **PROC-03**: `gux init` runs `go mod tidy` to download dependencies
- [ ] **PROC-04**: Scaffolded app runs with `gux dev` immediately after init

### Cleanup

- [ ] **CLNP-01**: Remove old cmd/app/main.go.tmpl template (uses deleted components/)
- [ ] **CLNP-02**: Remove old cmd/server/main.go.tmpl template (not needed with core/)
- [ ] **CLNP-03**: Remove internal/api/*.tmpl templates (replaced by gux gen)
- [ ] **CLNP-04**: Update scaffold.go to use new file list

## Out of Scope

| Feature | Reason |
|---------|--------|
| Admin routes in scaffold | Keep simple, users add themselves |
| DTOs in scaffold | Adds complexity, basic CRUD is cleaner for starter |
| Multiple models | One model demonstrates the pattern |
| Authentication | Covered by auth example, not starter scaffold |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| TMPL-01 | Phase 27 | Pending |
| TMPL-02 | Phase 27 | Pending |
| TMPL-03 | Phase 27 | Pending |
| TMPL-04 | Phase 27 | Pending |
| TMPL-05 | Phase 27 | Pending |
| TMPL-06 | Phase 27 | Pending |
| PROC-01 | Phase 27 | Pending |
| PROC-02 | Phase 27 | Pending |
| PROC-03 | Phase 27 | Pending |
| PROC-04 | Phase 27 | Pending |
| CLNP-01 | Phase 27 | Pending |
| CLNP-02 | Phase 27 | Pending |
| CLNP-03 | Phase 27 | Pending |
| CLNP-04 | Phase 27 | Pending |

**Coverage:**
- v2.2 requirements: 14 total
- Mapped to phases: 14
- Unmapped: 0 ✓

---
*Requirements defined: 2026-01-26*
*Last updated: 2026-01-26 after initial definition*
