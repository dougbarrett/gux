# Requirements: Gux Framework v2.3

**Defined:** 2026-01-26
**Core Value:** Developers can quickly see correct patterns for common gux tasks without leaving the terminal

## v2.3 Requirements

Requirements for the gux help patterns milestone.

### Help Commands

- [ ] **HELP-01**: `gux help` lists all available pattern names with brief descriptions
- [ ] **HELP-02**: `gux help page` outputs basic page template with OnLoad pattern
- [ ] **HELP-03**: `gux help page:list` outputs list page with API call and table display
- [ ] **HELP-04**: `gux help page:form` outputs form page with state, validation, API submission
- [ ] **HELP-05**: `gux help model` outputs GORM model with ID, common fields, timestamps
- [ ] **HELP-06**: `gux help dto` outputs DTO struct with gux tags for field mapping
- [ ] **HELP-07**: `gux help routes` outputs route registration with hybrid routes and CRUD
- [ ] **HELP-08**: `gux help app` outputs complete app.go with core.New(), DB, routes, run

### go.mod Integration

- [ ] **MOD-01**: Read module path from go.mod in current directory
- [ ] **MOD-02**: Replace import path placeholders with actual module path in output
- [ ] **MOD-03**: Gracefully handle missing go.mod (use placeholder, warn user)

### Output Format

- [ ] **OUT-01**: Output includes file path header comment (e.g., `// pages/items.go`)
- [ ] **OUT-02**: Output uses correct package name for file location
- [ ] **OUT-03**: Output is valid, copy-pasteable Go code

## Out of Scope

| Feature | Reason |
|---------|--------|
| Interactive prompts for names | Keep simple, fixed templates with placeholders |
| File generation (writing files) | This is for viewing patterns, gux init handles generation |
| Context-aware detection | Adds complexity, fixed templates are sufficient |
| Custom template support | Defer to future, built-in templates cover common cases |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| HELP-01 | Phase 28 | Pending |
| HELP-02 | Phase 28 | Pending |
| HELP-03 | Phase 28 | Pending |
| HELP-04 | Phase 28 | Pending |
| HELP-05 | Phase 28 | Pending |
| HELP-06 | Phase 28 | Pending |
| HELP-07 | Phase 28 | Pending |
| HELP-08 | Phase 28 | Pending |
| MOD-01 | Phase 28 | Pending |
| MOD-02 | Phase 28 | Pending |
| MOD-03 | Phase 28 | Pending |
| OUT-01 | Phase 28 | Pending |
| OUT-02 | Phase 28 | Pending |
| OUT-03 | Phase 28 | Pending |

**Coverage:**
- v2.3 requirements: 14 total
- Mapped to phases: 14
- Unmapped: 0 ✓

---
*Requirements defined: 2026-01-26*
*Last updated: 2026-01-26 after initial definition*
