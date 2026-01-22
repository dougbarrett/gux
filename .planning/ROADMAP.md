# Roadmap: Gux Framework

## Overview

A full-stack Go framework for building modern web applications with WebAssembly. The `core` package provides universal rendering (SSR + WASM hydration), and the component library provides pre-built UI components built on that foundation.

## Domain Expertise

- ~/.claude/skills/expertise/templ/SKILL.md
- ~/.claude/skills/expertise/go/SKILL.md

## Milestones

- [v1.0 UX Polish](milestones/v1.0-ROADMAP.md) (Phases 1-6) - SHIPPED 2026-01-15
- [v1.1 Accessibility](milestones/v1.1-ROADMAP.md) (Phases 7-11) - SHIPPED 2026-01-16
- [v1.2 Documentation](milestones/v1.2-ROADMAP.md) (Phases 12-15) - SHIPPED 2026-01-15
- [v2.0 Core Components](milestones/v2.0-ROADMAP.md) (Phases 16-23) - IN PROGRESS

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

## Phase Details

<details>
<summary>v1.0 UX Polish (Phases 1-6) - SHIPPED 2026-01-15</summary>

**Milestone Goal:** Comprehensive UI/UX enhancements to bring the application to production-ready quality

- [x] Phase 1: Header Components (2/2 plans) - UserMenu + NotificationCenter
- [x] Phase 2: Layout & Navigation (2/2 plans) - Collapsible Sidebar + Command Palette
- [x] Phase 3: Table Enhancements (4/4 plans) - Sorting, filtering, pagination, bulk actions
- [x] Phase 4: UX Polish (3/3 plans) - Persistent Preferences + Keyboard Navigation
- [x] Phase 5: Data & States (3/3 plans) - Data Export + Empty States
- [x] Phase 6: Progressive Enhancement (3/3 plans) - Connection Status + PWA

Full details: [milestones/v1.0-ROADMAP.md](milestones/v1.0-ROADMAP.md)

</details>

<details>
<summary>v1.1 Accessibility (Phases 7-11) - SHIPPED 2026-01-16</summary>

**Milestone Goal:** Enterprise-ready accessibility compliance with WCAG 2.1 AA standards

- [x] Phase 7: Accessibility Audit (3/3 plans) - 114 gaps documented, prioritized, mapped
- [x] Phase 8: ARIA & Semantic Markup (6/6 plans) - Screen reader support
- [x] Phase 9: Keyboard Navigation (4/4 plans) - Comprehensive keyboard support
- [x] Phase 10: Visual Accessibility (2/2 plans) - Focus indicators, contrast, motion
- [x] Phase 11: A11y Testing Infrastructure (1/1 plans) - axe-core integration

Full details: [milestones/v1.1-ROADMAP.md](milestones/v1.1-ROADMAP.md)

</details>

<details>
<summary>v1.2 Documentation (Phases 12-15) - SHIPPED 2026-01-15</summary>

**Milestone Goal:** Bring documentation up to date with v1.0 and v1.1 features

- [x] Phase 12: README Update (1/1 plans) - Comprehensive README with v1.0/v1.1 features
- [x] Phase 13: Component Docs (1/1 plans) - 7 new component APIs documented
- [x] Phase 14: Keyboard Shortcuts (1/1 plans) - 279 lines, all shortcuts by feature area
- [x] Phase 15: Accessibility Guide (1/1 plans) - 570 lines, ARIA patterns, testing guide

Full details: [milestones/v1.2-ROADMAP.md](milestones/v1.2-ROADMAP.md)

</details>

<details open>
<summary>v2.0 Core Components (Phases 16-23) - IN PROGRESS</summary>

**Milestone Goal:** Build a fresh component library on core's Node system with four example applications

- [x] Phase 16: Component Foundation (3/3 plans) - Utilities, Button, Card, layout primitives
  Plans:
  - [x] 16-01-PLAN.md - Utility functions (MergeClasses, ConditionalClass)
  - [x] 16-02-PLAN.md - Button component with variants and sizes
  - [x] 16-03-PLAN.md - Card compound components and layout primitives
- [x] Phase 17: Form Components (3/3 plans) - Input, Textarea, Select, Checkbox, Radio, Switch, FormField
  Plans:
  - [x] 17-01-PLAN.md - Input and FormField components
  - [x] 17-02-PLAN.md - Textarea and Select components
  - [x] 17-03-PLAN.md - Checkbox, RadioGroup, and Switch components
- [ ] Phase 18: Data Display Components (3 plans) - Badge, Avatar, Table, List, Pagination, DataTable[T]
  Plans:
  - [ ] 18-01-PLAN.md - Badge and Avatar components
  - [ ] 18-02-PLAN.md - Table compound and List components
  - [ ] 18-03-PLAN.md - Pagination and DataTable[T] generic component
- [ ] Phase 19: Interactive Components - Modal, Dropdown, Tabs, Toast, Tooltip
- [ ] Phase 20: Auth Example - Login, Register, Password Reset, Verification
- [ ] Phase 21: Marketing Example - Home, Features, Pricing, About, Contact
- [ ] Phase 22: SaaS Dashboard Example - Dashboard, CRUD, Settings, Profile
- [ ] Phase 23: Admin Panel Example - User Management, Activity Logs, Settings

Full details: [milestones/v2.0-ROADMAP.md](milestones/v2.0-ROADMAP.md)

</details>

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> ... -> 15 -> 16 -> ... -> 23

| Phase | Milestone | Plans | Status | Completed |
|-------|-----------|-------|--------|-----------|
| 1. Header Components | v1.0 | 2/2 | Complete | 2026-01-15 |
| 2. Layout & Navigation | v1.0 | 2/2 | Complete | 2026-01-15 |
| 3. Table Enhancements | v1.0 | 4/4 | Complete | 2026-01-15 |
| 4. UX Polish | v1.0 | 3/3 | Complete | 2026-01-15 |
| 5. Data & States | v1.0 | 3/3 | Complete | 2026-01-15 |
| 6. Progressive Enhancement | v1.0 | 3/3 | Complete | 2026-01-15 |
| 7. Accessibility Audit | v1.1 | 3/3 | Complete | 2026-01-15 |
| 8. ARIA & Semantic Markup | v1.1 | 6/6 | Complete | 2026-01-15 |
| 9. Keyboard Navigation | v1.1 | 4/4 | Complete | 2026-01-15 |
| 10. Visual Accessibility | v1.1 | 2/2 | Complete | 2026-01-15 |
| 11. A11y Testing Infrastructure | v1.1 | 1/1 | Complete | 2026-01-16 |
| 12. README Update | v1.2 | 1/1 | Complete | 2026-01-15 |
| 13. Component Docs | v1.2 | 1/1 | Complete | 2026-01-15 |
| 14. Keyboard Shortcuts | v1.2 | 1/1 | Complete | 2026-01-15 |
| 15. Accessibility Guide | v1.2 | 1/1 | Complete | 2026-01-15 |
| 16. Component Foundation | v2.0 | 3/3 | Complete | 2026-01-22 |
| 17. Form Components | v2.0 | 3/3 | Complete | 2026-01-22 |
| 18. Data Display Components | v2.0 | 3 | Planned | - |
| 19. Interactive Components | v2.0 | - | Pending | - |
| 20. Auth Example | v2.0 | - | Pending | - |
| 21. Marketing Example | v2.0 | - | Pending | - |
| 22. SaaS Dashboard Example | v2.0 | - | Pending | - |
| 23. Admin Panel Example | v2.0 | - | Pending | - |
