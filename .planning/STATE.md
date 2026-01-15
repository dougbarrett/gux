# Project State: GoQuery

## Current Position

Phase: 9 of 11 (Keyboard Navigation)
Plan: 4 of 4 in current phase
Status: Phase complete
Last activity: 2026-01-15 - Completed 09-04-PLAN.md

Progress: █████████░ 90% (v1.1: Phase 9 complete, ready for Phase 10)

## Accumulated Context

### Key Decisions
- Phase 1 discovery Level 0 (skip) - all patterns exist in codebase (Dropdown, Badge, Avatar, WebSocket)
- 01-01: Extended Dropdown with custom content for UserMenu and NotificationCenter
- 01-01: Used emoji icons for menu items for simplicity
- 01-02: Display order in header: bell icon, user avatar, then action buttons
- 01-02: Used nested div for action buttons to maintain tighter gap
- 02-01: Hide title completely when collapsed (cleaner than truncating)
- 02-01: Use w-16 for collapsed width (fits icons with padding)
- 02-01: Keyboard shortcut registration pattern with js.Func cleanup
- 02-02: Group commands by category with sticky headers
- 02-02: Use updateHighlightStyles() for hover to preserve click handlers
- 03-01: Emoji sort indicators (▲/▼/⇅) for simplicity
- 03-01: Case-insensitive string sorting, nil values sort to end
- 03-02: 150ms debounce for filter input
- 03-02: Case-insensitive substring matching
- 03-02: Filter → sort → render pipeline order
- 03-03: Default PageSize of 10 items per page
- 03-03: Reset to page 1 when filter changes
- 03-03: Filter → sort → paginate → render pipeline order
- 03-04: Clear selection when SetData is called
- 03-04: Selection persists across pages
- 03-04: Bulk action bar positioned between filter and table
- 04-01: Follow theme.go localStorage pattern for sidebar persistence
- 04-01: Use applyCollapsedState() helper to avoid callback during init
- 04-02: Follow CommandPalette pattern for keyboard handling
- 04-02: Use crypto.randomUUID() for unique menuitem IDs
- 04-02: Skip disabled items during keyboard navigation
- 04-03: Use ConfirmVariant* prefix for constants to avoid name collision with convenience functions
- 04-03: Wrap Modal internally rather than exposing Modal configuration
- 05-01: Manual CSV building instead of encoding/csv (cleaner for WASM)
- 05-01: Export dropdown in toolbar next to filter input
- 05-01: Export selected rows when selection exists, otherwise filtered data
- 05-02: Load jsPDF from CDN (no bundler in project)
- 05-02: Use positional arguments for jsPDF constructor (orientation, unit, format)
- 05-02: Use autoTable plugin for professional table formatting
- 05-03: Default icon 📭 for no-data, 🔍 for no-results
- 05-03: EmptyState hides table wrapper and pagination when active
- 05-03: Clear filter action built into no-results state
- 06-01: Dot variant as default for header ConnectionStatus
- 06-01: BindToWebSocket() for reactive subscription to store
- 06-01: SetState() for manual state control in demos
- 06-02: Gux branding for PWA (user preference)
- 06-02: Cache-first for same-origin, network-first for CDN
- 06-02: gux-v1 cache name for versioned cache management
- 06-03: InstallPromptManager pattern separates event lifecycle from UI
- 06-03: 7-day dismissal cooldown stored in localStorage
- 06-03: 503 response for failed CDN resources instead of throwing
- 07-03: P0 Critical = 8 blocking gaps (focus trap, live regions, labels)
- 07-03: P1 High = 52 major barriers mapped to WCAG criteria
- 07-03: Phase 8 gets ARIA roles/labels (60 issues)
- 07-03: Phase 9 gets keyboard navigation (15 issues)
- 07-03: Phase 10 gets visual accessibility (10 issues)
- 08-01: crypto.randomUUID() for unique ARIA IDs
- 08-01: ModalElement() getter for ConfirmDialog role override
- 08-01: role=combobox on CommandPalette input
- 08-02: Alert error/warning → role="alert", info/success → role="status"
- 08-02: Toast container gets live region, not individual toasts
- 08-02: Progress aria-valuenow omitted for indeterminate state
- 08-02: Spinner default aria-label="Loading", custom via AriaLabel prop
- 08-03: crypto.randomUUID() for form control IDs (consistent with 08-01)
- 08-03: Error elements use role="alert" for immediate announcement
- 08-03: htmlFor attribute for label-input association
- 08-04: strconv.Itoa + UUID for indexed widget IDs
- 08-04: Tabs roving tabindex (active=0, inactive=-1)
- 08-04: Combobox aria-activedescendant for virtual listbox focus
- 08-04: Accordion role=region on panels for screen reader context
- 08-05: Dropdown trigger stores reference for ARIA state updates
- 08-05: Sidebar nav uses role=navigation with aria-label
- 08-05: Breadcrumbs use semantic ol/li per WAI-ARIA practices
- 08-05: Separators hidden from AT with aria-hidden=true
- 08-06: DatePicker uses semantic table for calendar grid (not divs)
- 08-06: aria-live=polite on month/year for navigation announcements
- 08-06: Table aria-sort tracks none/ascending/descending state
- 09-01: Reuse existing FocusTrap component for Modal (consistent with CommandPalette)
- 09-01: FocusTrap.Activate() stores trigger, FocusTrap.Deactivate() restores focus
- 09-02: Automatic activation on arrow press (WAI-ARIA Tabs pattern)
- 09-02: Horizontal arrows only (Left/Right) matching tablist orientation
- 09-03: moveFocusBy() handles month boundary transitions automatically
- 09-03: Roving tabindex pattern for DatePicker grid (focused=0, others=-1)
- 09-03: Focus initialization priority: selected date > today > day 1
- 09-04: Dropdown focus restoration in Close() handles all close paths
- 09-04: Sidebar stores lastFocusedElement, focuses close button on mobile open
- 09-04: SkipLink already exists in skiplinks.go with MainSkipLink() function

### Blockers/Concerns Carried Forward
- None

## Deferred Issues

None yet.

## Roadmap Evolution

- Milestone v1.0 UX Polish created: UI/UX enhancements for production readiness, 6 phases (Phase 1-6)
- Phase 3 (Table Enhancements) complete with all 4 plans executed
- Phase 4 (UX Polish) complete with all 3 plans executed
- Phase 5 (Data & States) complete with all 3 plans executed
- Phase 6 (Progressive Enhancement) complete with all 3 plans executed
- **v1.0 UX Polish milestone complete** - All 6 phases, 17 plans executed (shipped 2026-01-15)
- Milestone v1.1 Accessibility created: Enterprise-ready a11y compliance, 5 phases (Phase 7-11)
- Phase 7 (Accessibility Audit) complete with all 3 plans executed - 114 gaps documented, prioritized, and mapped
- Phase 8 (ARIA & Semantic Markup) complete - 6/6 plans executed
- Phase 9 (Keyboard Navigation) complete - 4/4 plans executed

## Session Continuity

Last session: 2026-01-15
Stopped at: Completed 09-04-PLAN.md (Focus Management Polish)
Resume file: None
