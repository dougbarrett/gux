# Roadmap: GoQuery

## Overview

Transform GoQuery from a functional database query tool into a polished, production-ready application with comprehensive UI/UX enhancements including a command palette, enhanced tables, notification system, and progressive web app capabilities.

## Domain Expertise

- ~/.claude/skills/expertise/templ/SKILL.md
- ~/.claude/skills/expertise/go/SKILL.md

## Milestones

- 🚧 **v1.0 UX Polish** - Phases 1-6 (in progress)

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

- [x] **Phase 1: Header Components** - User Menu + Notification Center
- [x] **Phase 2: Layout & Navigation** - Collapsible Sidebar + Command Palette
- [x] **Phase 3: Table Enhancements** - Sorting, filtering, pagination, bulk actions
- [x] **Phase 4: UX Polish** - Persistent Preferences + Keyboard Navigation + Confirmation Dialog
- [ ] **Phase 5: Data & States** - Data Export + Empty States
- [ ] **Phase 6: Progressive Enhancement** - Skeleton Loaders + Connection Status + Breadcrumbs + PWA

## Phase Details

### 🚧 v1.0 UX Polish (In Progress)

**Milestone Goal:** Comprehensive UI/UX enhancements to bring the application to production-ready quality

#### Phase 1: Header Components

**Goal**: Add User Menu dropdown and Notification Center with real-time updates
**Depends on**: Nothing (first phase)
**Research**: Skipped (Level 0 - all patterns exist in codebase)
**Plans**: 2 plans, 4 tasks

Plans:
- [x] 01-01: Core Header Components (UserMenu + NotificationCenter components)
- [x] 01-02: Header Integration (extend Header, update example app)

#### Phase 2: Layout & Navigation (Complete)

**Goal**: Implement collapsible sidebar and Cmd+K command palette
**Depends on**: Phase 1
**Research**: Skipped (patterns existed in Modal + Combobox)
**Plans**: 2 plans, 5 tasks

Plans:
- [x] 02-01: Collapsible Sidebar with Cmd/Ctrl+B shortcut
- [x] 02-02: Command Palette with Cmd/Ctrl+K shortcut

#### Phase 3: Table Enhancements

**Goal**: Add sorting, filtering, pagination, and bulk selection with actions to tables
**Depends on**: Phase 2
**Research**: Skipped (Level 0 - all patterns exist: Checkbox, Pagination, Input)
**Plans**: 4 plans, 9 tasks

Plans:
- [x] 03-01: Table Sorting (sortable columns, sort icons, client-side sort)
- [x] 03-02: Table Filtering (search input, real-time filter, debounce)
- [x] 03-03: Table Pagination (integrate Pagination component, page-aware rendering)
- [x] 03-04: Bulk Selection & Actions (checkbox column, select-all, action bar)

#### Phase 4: UX Polish (Complete)

**Goal**: Implement persistent preferences, keyboard navigation, and confirmation dialogs
**Depends on**: Phase 3
**Research**: Skipped (localStorage patterns, focus management)
**Plans**: 3 plans, 7 tasks

Plans:
- [x] 04-01: Sidebar localStorage Persistence
- [x] 04-02: Dropdown Keyboard Navigation
- [x] 04-03: ConfirmDialog Component

#### Phase 5: Data & States

**Goal**: Add data export (CSV/JSON/PDF) and empty state illustrations
**Depends on**: Phase 4
**Research**: Likely (PDF generation)
**Research topics**: PDF generation in Go/JS, file download patterns, CSV export
**Plans**: TBD

Plans:
- [ ] 05-01: TBD

#### Phase 6: Progressive Enhancement

**Goal**: Add skeleton loaders, connection status indicator, breadcrumbs, and PWA support
**Depends on**: Phase 5
**Research**: Likely (PWA architecture)
**Research topics**: Service worker setup, offline capability, PWA manifest, cache strategies
**Plans**: TBD

Plans:
- [ ] 06-01: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6

| Phase | Milestone | Plans | Status | Completed |
|-------|-----------|-------|--------|-----------|
| 1. Header Components | v1.0 | 2/2 | Complete | 2026-01-15 |
| 2. Layout & Navigation | v1.0 | 2/2 | Complete | 2026-01-15 |
| 3. Table Enhancements | v1.0 | 4/4 | Complete | 2026-01-15 |
| 4. UX Polish | v1.0 | 3/3 | Complete | 2026-01-15 |
| 5. Data & States | v1.0 | 0/? | Not started | - |
| 6. Progressive Enhancement | v1.0 | 0/? | Not started | - |
