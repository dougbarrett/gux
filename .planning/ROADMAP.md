# Roadmap: GoQuery

## Overview

Transform GoQuery from a functional database query tool into a polished, production-ready application with comprehensive UI/UX enhancements including a command palette, enhanced tables, notification system, and progressive web app capabilities.

## Domain Expertise

- ~/.claude/skills/expertise/templ/SKILL.md
- ~/.claude/skills/expertise/go/SKILL.md

## Milestones

- ✅ **v1.0 UX Polish** - Phases 1-6 (shipped 2026-01-15)
- 🚧 **v1.1 Accessibility** - Phases 7-11 (in progress)

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

- [x] **Phase 1: Header Components** - User Menu + Notification Center
- [x] **Phase 2: Layout & Navigation** - Collapsible Sidebar + Command Palette
- [x] **Phase 3: Table Enhancements** - Sorting, filtering, pagination, bulk actions
- [x] **Phase 4: UX Polish** - Persistent Preferences + Keyboard Navigation + Confirmation Dialog
- [x] **Phase 5: Data & States** - Data Export + Empty States
- [x] **Phase 6: Progressive Enhancement** - Skeleton Loaders + Connection Status + Breadcrumbs + PWA
- [x] **Phase 7: Accessibility Audit** - Review all components, document gaps, establish baseline
- [ ] **Phase 8: ARIA & Semantic Markup** - Screen reader support, labels, roles, live regions
- [ ] **Phase 9: Keyboard Navigation** - Comprehensive keyboard support for all components
- [ ] **Phase 10: Visual Accessibility** - Focus indicators, contrast, reduced motion
- [ ] **Phase 11: A11y Testing Infrastructure** - axe-core integration, testing patterns

## Phase Details

<details>
<summary>✅ v1.0 UX Polish (Phases 1-6) - SHIPPED 2026-01-15</summary>

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

#### Phase 5: Data & States (Complete)

**Goal**: Add data export (CSV/JSON/PDF) and empty state illustrations
**Depends on**: Phase 4
**Research**: Likely (PDF generation)
**Research topics**: PDF generation in Go/JS, file download patterns, CSV export
**Plans**: 3 plans

Plans:
- [x] 05-01: Data Export (CSV/JSON with Table integration)
- [x] 05-02: PDF Export (jsPDF with autoTable)
- [x] 05-03: Empty States

#### Phase 6: Progressive Enhancement

**Goal**: Add connection status indicator and PWA support (skeleton loaders and breadcrumbs already exist)
**Depends on**: Phase 5
**Research**: Skipped (Level 0 - existing patterns for UI, standard PWA patterns)
**Plans**: 3 plans, 7 tasks

Plans:
- [x] 06-01: Connection Status Component (visual indicator for WebSocket state)
- [x] 06-02: PWA Foundation (manifest, service worker, asset caching)
- [x] 06-03: PWA Install Experience (install prompt, offline fallback)

</details>

### 🚧 v1.1 Accessibility (In Progress)

**Milestone Goal:** Enterprise-ready accessibility compliance with WCAG 2.1 AA standards

#### Phase 7: Accessibility Audit (Complete)

**Goal**: Review all components, document gaps, establish baseline
**Depends on**: v1.0 complete
**Research**: Completed during planning
**Output**: 114 gaps documented across 25 components, prioritized P0-P3, mapped to Phases 8-11
**Plans**: 3 plans, 3 tasks

Plans:
- [x] 07-01: Interactive Components Audit (Modal, Dropdown, CommandPalette, ConfirmDialog, Combobox, Tabs, Accordion, Table)
- [x] 07-02: Form & Navigation Audit (Input, Select, Checkbox, Toggle, TextArea, DatePicker, Form, FormBuilder, Sidebar, Breadcrumbs, Pagination, Link, Alert, Toast, Progress, Spinner, Skeleton)
- [x] 07-03: Combined Findings & Remediation Plan

#### Phase 8: ARIA & Semantic Markup (In Progress)

**Goal**: Add screen reader support with proper labels, roles, and live regions
**Depends on**: Phase 7
**Research**: Completed during planning
**Plans**: 6 plans

Plans:
- [x] 08-01: ARIA Dialog Patterns (Modal, ConfirmDialog, CommandPalette)
- [x] 08-02: ARIA Live Regions (Alert, Toast, Progress, Spinner)
- [ ] 08-03: Table Accessibility
- [ ] 08-04: Navigation & Landmark Roles
- [ ] 08-05: Live Regions & Status Updates
- [ ] 08-06: Interactive Widget ARIA

#### Phase 9: Keyboard Navigation

**Goal**: Ensure comprehensive keyboard support for all interactive components
**Depends on**: Phase 8
**Research**: Unlikely (patterns exist from v1.0: Dropdown 04-02, Command Palette 02-02)
**Plans**: TBD

Plans:
- [ ] 09-01: TBD (run /gsd:plan-phase 9 to break down)

#### Phase 10: Visual Accessibility

**Goal**: Implement focus indicators, color contrast compliance, and reduced motion support
**Depends on**: Phase 9
**Research**: Likely (WCAG contrast ratios, focus-visible patterns)
**Research topics**: Contrast ratio requirements, focus-visible patterns, prefers-reduced-motion
**Plans**: TBD

Plans:
- [ ] 10-01: TBD (run /gsd:plan-phase 10 to break down)

#### Phase 11: A11y Testing Infrastructure

**Goal**: Integrate axe-core for automated accessibility regression testing
**Depends on**: Phase 10
**Research**: Likely (axe-core API, Go/WASM testing integration)
**Research topics**: axe-core API, automated testing integration, Go testing patterns
**Plans**: TBD

Plans:
- [ ] 11-01: TBD (run /gsd:plan-phase 11 to break down)

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9 → 10 → 11

| Phase | Milestone | Plans | Status | Completed |
|-------|-----------|-------|--------|-----------|
| 1. Header Components | v1.0 | 2/2 | Complete | 2026-01-15 |
| 2. Layout & Navigation | v1.0 | 2/2 | Complete | 2026-01-15 |
| 3. Table Enhancements | v1.0 | 4/4 | Complete | 2026-01-15 |
| 4. UX Polish | v1.0 | 3/3 | Complete | 2026-01-15 |
| 5. Data & States | v1.0 | 3/3 | Complete | 2026-01-15 |
| 6. Progressive Enhancement | v1.0 | 3/3 | Complete | 2026-01-15 |
| 7. Accessibility Audit | v1.1 | 3/3 | Complete | 2026-01-15 |
| 8. ARIA & Semantic Markup | v1.1 | 2/6 | In progress | - |
| 9. Keyboard Navigation | v1.1 | 0/? | Not started | - |
| 10. Visual Accessibility | v1.1 | 0/? | Not started | - |
| 11. A11y Testing Infrastructure | v1.1 | 0/? | Not started | - |
