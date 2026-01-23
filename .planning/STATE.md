# Project State: Gux Framework

## Current Position

Phase: 23 — Admin Panel (COMPLETE)
Plan: 3/3 complete
Status: Phase complete
Last activity: 2026-01-23 — Completed 23-03-PLAN.md (Activity Logs and Settings)

Progress: ██████████ 100% of Phase 23

## Project Reference

See: .planning/PROJECT.md (updated 2026-01-22)

**Core value:** Developers can build complete web apps in Go with SSR + WASM
**Current focus:** v2.0 Core Components — component library + four examples

## Current Milestone: v2.0 Core Components

**Goal:** Rebuild component library on core's Node system with four example applications

**Phases:**
- [x] Phase 16: Component Foundation — utilities, Button, Card, layout primitives (Complete 2026-01-22)
- [x] Phase 17: Form Components — Input, Textarea, Select, Checkbox, Radio (Complete 2026-01-22)
- [x] Phase 18: Data Display — Table, DataTable[T], Badge, Avatar, Pagination (Complete 2026-01-22)
- [x] Phase 19: Interactive — Modal, Dropdown, Tabs, Toast, Tooltip (Complete 2026-01-22)
- [x] Phase 20: Auth Example — Login, Register, Password Reset (Complete 2026-01-22)
- [x] Phase 21: Marketing Example — Home, Features, Pricing, Contact (Complete 2026-01-22)
- [x] Phase 22: SaaS Dashboard — Dashboard, CRUD, Settings (Complete 2026-01-23)
- [x] Phase 23: Admin Panel — User Management, Activity Logs (Complete 2026-01-23)

## Accumulated Context

### Key Decisions (Summary)

Full decision log preserved in milestone archives:
- [v1.0-ROADMAP.md](milestones/v1.0-ROADMAP.md) - UX/UI patterns
- [v1.1-ROADMAP.md](milestones/v1.1-ROADMAP.md) - Accessibility patterns
- [v1.2-ROADMAP.md](milestones/v1.2-ROADMAP.md) - Documentation patterns

**v2.0 decisions:**
- Fresh start on components (not porting old library)
- Shared foundation across examples
- Tailwind CSS styling
- Core auth flows only (no OAuth)
- No new stack additions needed
- 8 phases: 4 component phases + 4 example apps

**Phase 16 decisions:**
- Exported function names (MergeClasses, ConditionalClass) for public API
- Whitespace trimming in MergeClasses for robustness
- Table-driven tests with t.Run for comprehensive coverage
- ButtonVariant/ButtonSize as string types with named constants
- Default variant=primary, size=md, type=button for Button
- Disabled adds both attribute and visual styling
- Card uses compound pattern (Card wraps CardHeader/CardContent/CardFooter)
- Shared StackProps struct for VStack and HStack consistency
- Container defaults to max-w-7xl, Gap defaults to "4", Cols defaults to "3"

**Phase 17 decisions:**
- InputSize and input styling constants defined in input.go as canonical location
- Error takes precedence over description in FormField
- Required asterisk only shown when label is present
- Textarea value renders as child text node (not value attr) for SSR
- Select placeholder is disabled option at index 0, selected when Value is empty
- Select uses appearance-none + pr-10 for custom dropdown styling
- Boolean attributes use presence-based pattern (checked="checked" only when true)
- Switch uses hidden checkbox + visual spans for form submission and accessibility
- RadioGroup uses role="radiogroup" wrapper with individual radio inputs
- All boolean controls share disabled styling pattern (opacity-50, cursor-not-allowed)

**Phase 18 decisions (18-01):**
- Badge uses span element for inline display semantics
- Avatar fallback decided at render time (props.Src empty = initials), no JS onerror
- getInitials returns "?" for empty/whitespace-only names
- Avatar image takes precedence when both Src and Name provided

**Phase 18 decisions (18-02):**
- Table wraps in overflow-x-auto div for wide content safety
- Thead uses bg-gray-50 background for visual header distinction
- Tbody uses divide-y for row separation matching List pattern
- Th uses uppercase tracking-wider for standard header cell styling
- Td uses whitespace-nowrap for compact cell display
- List uses divide-y for item separation (same pattern as Tbody)
- ListItem includes hover:bg-gray-50 for interactivity feedback

**Phase 18 decisions (18-03):**
- Pagination uses 1-indexed pages (user-facing convention)
- Pagination returns empty fragment when total <= 1 pages
- Pagination shows up to 5 page numbers around current page
- DataTable[T] uses capturedItem pattern to avoid closure capture bug
- DataTable striped applies to odd-indexed rows (i%2 == 1)
- DataTable uses Table compound internally for consistent styling

**Phase 19 decisions (19-01):**
- Modal always renders full structure, visibility via CSS hidden class
- Header shows when Title OR OnClose provided (allows close-only modals with spacer)
- Backdrop OnClick calls OnClose (standard click-outside-to-close)
- ModalFull uses max-w-4xl (not 100vw) for generous but not excessive width

**Phase 19 decisions (19-02):**
- CSS-only visibility via group-hover and group-focus-within (no JS state needed)
- TooltipTop as default position
- Tooltip element renders after children (trigger first in DOM)
- pointer-events-none on tooltip to prevent hover interference

**Phase 19 decisions (19-03):**
- TabList uses role=tablist with aria-label=Tabs
- Tab uses roving tabindex (0 for active, -1 for inactive)
- TabPanel uses hidden class for inactive panels (CSS visibility)
- Disabled tabs have both disabled attribute and aria-disabled=true

**Phase 19 decisions (19-04):**
- Invisible backdrop div for outside click handling (avoids global event listeners)
- DropdownBottomLeft as default position (most common pattern)
- Destructive prop for red styling on dangerous actions (delete/remove)
- Reuse boolToAttr from switch.go (avoid duplication)

**Phase 19 decisions (19-05):**
- ToastContainer uses aria-live=polite, role=status, aria-atomic=true for accessibility
- Default position is ToastTopRight (most common notification placement)
- Close button rendered conditionally via props.OnClose != nil pattern
- Icon uses aria-hidden=true (decorative only)
- Unicode icons for cross-platform compatibility

**Phase 20 decisions (20-01):**
- Used core.El("strong", ...) for title since core.Strong helper doesn't exist
- Circled i unicode (U+24D8) for info icon to differentiate from Toast
- Alert uses 3-part variant styling (bg + border + text) vs Toast's 2-part

**Phase 20 decisions (20-02):**
- User.PasswordHash uses json:"-" to never expose in JSON serialization
- PasswordReset tokens have 1-hour expiry for security
- All 7 routes share single WASM bundle (no separate admin bundle)
- Stub pages ready for replacement in Plans 03/04
- AuthLayout: centered full-height container with max-w-md card area
- PageLayout: Nav header with max-w-7xl content area

**Phase 20 decisions (20-03):**
- Checkbox toggle uses wrapper div OnClick since core.Attrs.OnChange is string-based
- Field-level error state for each form field for inline validation display
- featureCard helper function for DRY feature grid in Home page

**Phase 20 decisions (20-04):**
- Security message: "If an account exists for..." to not reveal if email exists
- Demo mode with direct link to reset page for testing
- Token validation in OnLoad before render
- 8 character minimum password length
- Demo tokens "expired" and "invalid" for testing error states

**Phase 21 decisions (21-01):**
- Used core.El("h4", ...) for H4 elements since no core.H4 helper exists
- Used core.Frag() for fragment creation (not core.Fragment{})
- Used core.Attrs{} for empty attributes (nil not allowed)
- Mobile menu toggle via r.StateBool("menuOpen", false)
- Footer uses ui.Grid for 4-column responsive layout

**Phase 21 decisions (21-02):**
- Unicode emoji icons for feature cards (cross-platform compatible)
- featureDetailSection helper with visualRight parameter for alternating layouts
- Testimonial cards use italic quote with quoted text for visual distinction

**Phase 21 decisions (21-03):**
- PricingCardProps struct for tier card data with highlight and badge support
- Enterprise 'Contact Sales' uses href link instead of OnClick navigation
- Avatar initials from Name prop in teamMemberCard helper
- Form validation uses strings.TrimSpace and @ check for email

**Phase 22 decisions (22-01):**
- Horizontal nav pattern (matching examples/minimal/admin) for consistency
- Single WASM bundle (no separate admin bundle) for simplicity
- Port 8082 to avoid conflicts with other examples
- Mock user "John Doe" for avatar and profile display
- DashboardLayout wraps all pages with Nav + max-w-6xl content area

**Phase 22 decisions (22-02):**
- statCard helper with title, value, colorClass params for reusability
- statusBadge maps status to BadgeVariant (active->Success, completed->Default, archived->Warning)
- Delete confirmation uses ui.Modal with ModalSM size
- formField and selectOption helpers shared between create/edit forms
- DataTable with Striped=true, Hoverable=true, OnRowClick for navigation

**Phase 22 decisions (22-03):**
- settingRow helper wraps label/description with clickable toggle for better UX
- Profile email is read-only with 'Contact admin' hint
- Separate success state per tab to avoid cross-tab message confusion
- Danger zone uses ButtonDestructive for consistent destructive action styling

**Phase 23 decisions (23-01):**
- Port 8084 for admin example (following port pattern)
- User model extended with Role, Status, LastLoginAt fields
- PasswordHash uses json:"-" tag to never expose in JSON
- Dark theme with bg-gray-900 and gray-800 nav
- ActivityLog model with UserID, Action, Entity, Description, IPAddress, Metadata

**Phase 23 decisions (23-03):**
- Client-side filtering with JSON state for activity logs
- Action badge colors: created/login=Success, updated=Default, deleted=Error, logout=Warning
- Settings tabs: General, Security, Email (admin system settings focus)
- Role-based UI: Danger Zone only visible for admin role

### Research Findings

Research completed 2026-01-22. Key findings:
- Stack: Go 1.24.3, GORM, Tailwind CDN — no changes needed
- Architecture: Props structs, core.Node interface, compound patterns
- Pitfalls: Callback cleanup, binary size, hydration mismatches

See: [.planning/research/SUMMARY.md](research/SUMMARY.md)

### Blockers/Concerns Carried Forward

None

## Deferred Issues

None

## Roadmap Evolution

- **v1.0 UX Polish** - 6 phases, 17 plans (shipped 2026-01-15)
- **v1.1 Accessibility** - 5 phases, 16 plans (shipped 2026-01-16)
- **v1.2 Documentation** - 4 phases, 4 plans (shipped 2026-01-15)
- **v2.0 Core Components** - 8 phases, ~20 plans (COMPLETE 2026-01-23)

## Session Continuity

Last session: 2026-01-23
Stopped at: Completed 23-03-PLAN.md, Phase 23 complete
Resume file: None

---

*Last updated: 2026-01-23 after 23-03 plan execution*
