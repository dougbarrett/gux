---
phase: 19
plan: 04
subsystem: components
tags: [dropdown, menu, aria, compound-components]

dependency-graph:
  requires: [19-01]
  provides: [Dropdown, DropdownTrigger, DropdownMenu, DropdownItem]
  affects: [19-05]

tech-stack:
  added: []
  patterns: [compound-components, backdrop-click-outside, aria-menu]

key-files:
  created:
    - ui/dropdown.go
    - ui/dropdown_test.go
  modified: []

decisions:
  - id: "dropdown-backdrop"
    choice: "Invisible backdrop div for outside click handling"
    why: "Avoids global event listeners, matches Modal pattern"
  - id: "dropdown-position-default"
    choice: "DropdownBottomLeft as default position"
    why: "Most common dropdown pattern opens below and left-aligned"
  - id: "dropdown-item-destructive"
    choice: "Destructive prop for red styling on dangerous actions"
    why: "Visual distinction for delete/remove actions"

metrics:
  duration: "2m 8s"
  completed: "2026-01-22"
---

# Phase 19 Plan 04: Dropdown Component Summary

Dropdown compound components providing menu semantics with backdrop for outside click handling.

## What Was Built

### DropdownPosition Type
- `DropdownBottomLeft` - opens below, left-aligned (default)
- `DropdownBottomRight` - opens below, right-aligned
- `DropdownTopLeft` - opens above, left-aligned
- `DropdownTopRight` - opens above, right-aligned

### Dropdown Wrapper
- Relative positioning context for absolute menu
- Invisible backdrop when open for click-outside-to-close
- OnClose callback for backdrop clicks

### DropdownTrigger
- Wrapper for toggle button
- `aria-haspopup="menu"` for menu semantics
- `aria-expanded` reflects open state

### DropdownMenu
- `role="menu"` for accessibility
- Position variants for all four corners
- Hidden via CSS when closed
- Base styling: shadow, border, rounded corners

### DropdownItem
- `role="menuitem"` for accessibility
- `type="button"` for button semantics
- Default styling: gray text, hover background
- Destructive styling: red text, red hover background
- Disabled styling: gray text, cursor-not-allowed, disabled attribute

## Decisions Made

1. **Invisible backdrop for click outside** - Uses a fixed full-screen div when open that captures clicks outside the menu. Calls OnClose on click. This avoids needing global event listeners in WASM.

2. **DropdownBottomLeft as default** - Most dropdown menus open below the trigger and align to the left edge. This matches user expectations for action menus.

3. **Destructive prop for danger actions** - Rather than variants like Button, DropdownItem uses a boolean Destructive prop for delete/remove actions. Red styling provides visual warning.

4. **Reuse boolToAttr from switch.go** - The helper function for converting bool to ARIA string was already defined in switch.go, so we reuse it rather than duplicating.

## Commits

| Hash | Type | Description |
|------|------|-------------|
| 77d8d88 | feat | Add Dropdown compound components |
| 1f0e0bb | test | Add Dropdown component tests |

## Test Coverage

22 tests covering:
- Dropdown wrapper (4 tests): rendering, backdrop open/closed, custom class
- DropdownTrigger (5 tests): aria-haspopup, aria-expanded true/false, children, custom class
- DropdownMenu (7 tests): role, visibility, all 4 positions, default position, children, custom class
- DropdownItem (6 tests): role, rendering, destructive, disabled, children, custom class

## Files

### Created
- `ui/dropdown.go` (221 lines) - Dropdown compound components
- `ui/dropdown_test.go` (378 lines) - Component tests

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Removed duplicate boolToAttr function**
- **Found during:** Task 1
- **Issue:** boolToAttr was already defined in switch.go
- **Fix:** Removed the duplicate definition from dropdown.go
- **Files modified:** ui/dropdown.go
- **Commit:** 77d8d88

**2. [Rule 3 - Blocking] Fixed disabled attribute handling**
- **Found during:** Task 1
- **Issue:** core.Attrs has no Disabled field, must use Extra map
- **Fix:** Added disabled attribute via Extra map instead of Attrs.Disabled
- **Files modified:** ui/dropdown.go
- **Commit:** 77d8d88

## Next Phase Readiness

Plan 19-05 (Toast component) can proceed. Dropdown is independent and complete.
