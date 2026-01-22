# Project State: Gux Framework

## Current Position

Phase: 17 — Form Components
Plan: 1/? in progress
Status: In progress
Last activity: 2026-01-22 — Completed 17-01-PLAN.md (Input and FormField)

Progress: ████████░░ ~80% of Phase 17

## Project Reference

See: .planning/PROJECT.md (updated 2026-01-22)

**Core value:** Developers can build complete web apps in Go with SSR + WASM
**Current focus:** v2.0 Core Components — component library + four examples

## Current Milestone: v2.0 Core Components

**Goal:** Rebuild component library on core's Node system with four example applications

**Phases:**
- [x] Phase 16: Component Foundation — utilities, Button, Card, layout primitives (Complete 2026-01-22)
- [ ] Phase 17: Form Components — Input, Textarea, Select, Checkbox, Radio
- [ ] Phase 18: Data Display — Table, DataTable[T], Badge, Avatar, Pagination
- [ ] Phase 19: Interactive — Modal, Dropdown, Tabs, Toast, Tooltip
- [ ] Phase 20: Auth Example — Login, Register, Password Reset
- [ ] Phase 21: Marketing Example — Home, Features, Pricing, Contact
- [ ] Phase 22: SaaS Dashboard — Dashboard, CRUD, Settings
- [ ] Phase 23: Admin Panel — User Management, Activity Logs

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
- **v2.0 Core Components** - 8 phases, ~20 plans (in progress)

## Session Continuity

Last session: 2026-01-22
Stopped at: Completed 17-01-PLAN.md (Input and FormField components)
Resume file: None

---

*Last updated: 2026-01-22 after 17-01 execution*
