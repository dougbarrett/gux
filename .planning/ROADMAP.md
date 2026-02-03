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
- [v2.0 Core Components](milestones/v2.0-ROADMAP.md) (Phases 16-23) - SHIPPED 2026-01-25
- [v2.1 Dead Code Cleanup](milestones/v2.1-ROADMAP.md) (Phases 24-26) - SHIPPED 2026-01-26
- [v2.2 gux init Modernization](milestones/v2.2-ROADMAP.md) (Phase 27) - SHIPPED 2026-01-26
- [v2.3 gux help Patterns](milestones/v2.3-ROADMAP.md) (Phase 28) - SHIPPED 2026-01-26
- **v2.4 File Upload System** (Phases 29-32) - IN PROGRESS

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

<details>
<summary>v2.0 Core Components (Phases 16-23) - SHIPPED 2026-01-25</summary>

**Milestone Goal:** Build a fresh component library on core's Node system with four example applications

- [x] Phase 16: Component Foundation (3/3 plans) - Utilities, Button, Card, layout primitives
- [x] Phase 17: Form Components (3/3 plans) - Input, Textarea, Select, Checkbox, Radio, Switch, FormField
- [x] Phase 18: Data Display Components (3/3 plans) - Badge, Avatar, Table, List, Pagination, DataTable[T]
- [x] Phase 19: Interactive Components (5/5 plans) - Modal, Dropdown, Tabs, Toast, Tooltip
- [x] Phase 20: Auth Example (4/4 plans) - Login, Register, Password Reset, Verification
- [x] Phase 21: Marketing Example (3/3 plans) - Home, Features, Pricing, About, Contact
- [x] Phase 22: SaaS Dashboard Example (3/3 plans) - Dashboard, CRUD, Settings, Profile
- [x] Phase 23: Admin Panel Example (3/3 plans) - User Management, Activity Logs, Settings

Full details: [milestones/v2.0-ROADMAP.md](milestones/v2.0-ROADMAP.md)

</details>

<details>
<summary>v2.1 Dead Code Cleanup (Phases 24-26) - SHIPPED 2026-01-26</summary>

- [x] Phase 24: Code Removal (1/1 plans) - completed 2026-01-25
- [x] Phase 25: Documentation Updates (3/3 plans) - completed 2026-01-26
- [x] Phase 26: Dependency Cleanup (1/1 plans) - completed 2026-01-26

Full details: [milestones/v2.1-ROADMAP.md](milestones/v2.1-ROADMAP.md)

</details>

<details>
<summary>v2.2 gux init Modernization (Phase 27) - SHIPPED 2026-01-26</summary>

**Milestone Goal:** Update `gux init` to scaffold working apps using the modern `core/` framework instead of the deleted `components/` package.

- [x] Phase 27: gux init Templates (3/3 plans) - completed 2026-01-26

Full details: [milestones/v2.2-ROADMAP.md](milestones/v2.2-ROADMAP.md)

</details>

<details>
<summary>v2.3 gux help Patterns (Phase 28) - SHIPPED 2026-01-26</summary>

**Milestone Goal:** Add `gux help <pattern>` commands that output boilerplate code for developers and LLMs working in established projects.

- [x] Phase 28: Help Pattern Commands (2/2 plans) - completed 2026-01-26

Full details: [milestones/v2.3-ROADMAP.md](milestones/v2.3-ROADMAP.md)

</details>

### v2.4 File Upload System (In Progress)

**Milestone Goal:** Developers can add `"input": "file"` to a model field in gux.config.json and get complete file upload -- UI, storage, API, admin integration -- with zero manual wiring.

- [x] **Phase 29: Storage Foundation** - Storage interface, local backend, upload endpoint, file serving
- [x] **Phase 30: Upload Client & UI Component** - WASM upload with progress, drag-drop UI, image preview
- [ ] **Phase 31: Code Generation & CRUD Integration** - Config-driven file fields, admin scaffolding, lifecycle hooks, auto-cleanup
- [ ] **Phase 32: Multi-File & Configuration** - Multi-file fields, per-field config for types/size/directory

#### Phase 29: Storage Foundation
**Goal**: Developer can upload files through a server endpoint and serve them via HTTP, with a clean storage abstraction that decouples file location from application code
**Depends on**: Nothing (first phase of v2.4)
**Requirements**: STOR-01, STOR-02, STOR-03, STOR-04, STOR-05, STOR-06, UPLD-01, UPLD-02, UPLD-03, UPLD-04, UPLD-05, UPLD-06
**Success Criteria** (what must be TRUE):
  1. Developer calls `app.SetStorage(storage.NewLocalStorage("./uploads"))` and files are stored to that directory
  2. A POST request with a multipart file to the upload endpoint returns a JSON response containing the stored file path
  3. Files are accessible via browser at their served URL (e.g., `/__gux_api/files/abc123.jpg` returns the image)
  4. Upload endpoint rejects files exceeding the size limit and files with disallowed content types (validated by magic bytes, not just headers)
  5. Upload endpoint enforces CSRF protection and authentication when configured on the app
**Plans**: 2 plans

Plans:
- [x] 29-01-PLAN.md -- Storage interface, types, functional options, LocalStorage with content-hash filenames, magic bytes validation, image dimensions
- [x] 29-02-PLAN.md -- Upload endpoint (multipart POST), file serving handler (GET with caching), app.SetStorage(), auth/CSRF integration

#### Phase 30: Upload Client & UI Component
**Goal**: User can select, preview, and upload files through an interactive UI component that works across SSR and WASM with upload progress feedback
**Depends on**: Phase 29
**Requirements**: UICM-01, UICM-02, UICM-03, UICM-04, UICM-05, UICM-06, UICM-07, UICM-08, UICM-09, UICM-10
**Success Criteria** (what must be TRUE):
  1. User can click to browse or drag-and-drop a file onto the upload zone, and sees a progress bar during upload
  2. User sees an immediate validation error when selecting a file that is too large or has a disallowed type (before upload starts)
  3. User sees an image thumbnail preview after selecting an image file, and can remove or replace the uploaded file
  4. The upload zone renders as static HTML during SSR (no JavaScript errors) and gains full interactivity after WASM hydration
  5. File bytes remain in JavaScript memory during upload (never copied into Go WASM heap)
**Plans**: 2 plans

Plans:
- [x] 30-01-PLAN.md -- WASM upload client (fetch/upload.go with XHR progress and CSRF)
- [x] 30-02-PLAN.md -- ui.FileUpload component with SSR/WASM split (drag-drop, validation, preview, progress)

#### Phase 31: Code Generation & CRUD Integration
**Goal**: Developer adds `"input": "file"` to a field in gux.config.json and `gux gen` produces complete upload integration across admin forms, detail views, list views, and CRUD lifecycle
**Depends on**: Phase 30
**Requirements**: CRUD-01, CRUD-02, CRUD-03, CRUD-04, CRUD-05, CRUD-06, CRUD-07, CRUD-08, CRUD-09, CRUD-10, HOOK-01, HOOK-02, HOOK-03, HOOK-04, HOOK-05, HOOK-06
**Success Criteria** (what must be TRUE):
  1. After adding `"input": "file"` to a field and running `gux gen`, the admin create and edit forms render a working `ui.FileUpload` component (with current file shown on edit)
  2. Generated admin detail views show an image preview for image fields and a filename with download link for non-image fields
  3. Generated admin list views show a thumbnail for image fields and a file icon or filename for non-image fields
  4. When a record with file fields is deleted via CRUD, the associated files are automatically removed from storage
  5. Developer can register BeforeUpload, AfterUpload, and BeforeDelete hooks that receive file metadata and follow the existing CRUD option pattern
**Plans**: 2 plans

Plans:
- [ ] 31-01-PLAN.md -- CRUD file lifecycle hooks, auto-cleanup on delete/update, rollback on DB failure
- [ ] 31-02-PLAN.md -- Code gen for file input: form fields (FileUpload), detail views (preview/download), list cells (thumbnail/icon), CRUD WithFileFields

#### Phase 32: Multi-File & Configuration
**Goal**: Developer can define multi-file fields and configure per-field upload constraints (allowed types, max size, upload directory) in gux.config.json
**Depends on**: Phase 31
**Requirements**: MULT-01, MULT-02, MULT-03, MULT-04, CONF-01, CONF-02, CONF-03
**Success Criteria** (what must be TRUE):
  1. Developer adds `"input": "file[]"` to a field in gux.config.json and the admin form allows uploading multiple files
  2. User can add and remove individual files from a multi-file field, with values stored as a JSON array of file paths
  3. Per-field `"accept"`, `"maxSize"`, and `"directory"` configuration in gux.config.json controls upload validation and storage location
**Plans**: TBD

Plans:
- [ ] 32-01: Multi-file field support (config, model, UI, storage)
- [ ] 32-02: Per-field configuration (accept, maxSize, directory)

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> ... -> 28 -> 29 -> 30 -> 31 -> 32

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
| 18. Data Display Components | v2.0 | 3/3 | Complete | 2026-01-22 |
| 19. Interactive Components | v2.0 | 5/5 | Complete | 2026-01-22 |
| 20. Auth Example | v2.0 | 4/4 | Complete | 2026-01-22 |
| 21. Marketing Example | v2.0 | 3/3 | Complete | 2026-01-22 |
| 22. SaaS Dashboard Example | v2.0 | 3/3 | Complete | 2026-01-23 |
| 23. Admin Panel Example | v2.0 | 3/3 | Complete | 2026-01-23 |
| 24. Code Removal | v2.1 | 1/1 | Complete | 2026-01-25 |
| 25. Documentation Updates | v2.1 | 3/3 | Complete | 2026-01-26 |
| 26. Dependency Cleanup | v2.1 | 1/1 | Complete | 2026-01-26 |
| 27. gux init Templates | v2.2 | 3/3 | Complete | 2026-01-26 |
| 28. Help Pattern Commands | v2.3 | 2/2 | Complete | 2026-01-26 |
| 29. Storage Foundation | v2.4 | 2/2 | Complete | 2026-02-01 |
| 30. Upload Client & UI Component | v2.4 | 2/2 | Complete | 2026-02-01 |
| 31. Code Generation & CRUD Integration | v2.4 | 0/2 | Not started | - |
| 32. Multi-File & Configuration | v2.4 | 0/2 | Not started | - |
