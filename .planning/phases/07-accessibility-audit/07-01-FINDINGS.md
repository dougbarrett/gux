# Phase 07-01: Interactive Components Accessibility Audit Findings

**Audit Date:** 2026-01-15
**WCAG Version:** 2.1 AA
**Auditor:** Claude Code

---

## Overlay Components

### 1. Modal (`components/modal.go`)

#### Current State
- Has visible close button (X) in header
- Closes on Escape key (when `CloseOnEsc: true`)
- Closes on overlay click
- Prevents body scroll when open

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Missing `role="dialog"` | 4.1.2 Name, Role, Value | **High** | Modal container has no ARIA role |
| Missing `aria-modal="true"` | 4.1.2 | **High** | Screen readers don't know this is a modal |
| No focus trap | 2.1.2 No Keyboard Trap / 2.4.3 Focus Order | **High** | Focus can escape modal with Tab key |
| No focus restoration | 2.4.3 Focus Order | **High** | Focus not returned to trigger on close |
| No `aria-labelledby` | 4.1.2 | **Medium** | Modal not associated with its title |
| No initial focus management | 2.4.3 | **Medium** | Focus not moved to modal on open |
| Close button lacks `aria-label` | 4.1.2 | **Medium** | Button only has visual X character |

#### Recommendations
1. Add `role="dialog"` and `aria-modal="true"` to modal container
2. Implement FocusTrap (already exists in codebase - `focustrap.go`)
3. Store and restore focus on open/close
4. Add `aria-labelledby` pointing to title element
5. Add `aria-label="Close"` to close button

---

### 2. Dropdown (`components/dropdown.go`)

#### Current State
- Has `role="menu"` on dropdown container
- Has `role="menuitem"` on items
- Has `aria-orientation="vertical"`
- Has keyboard navigation (ArrowUp/ArrowDown/Enter/Escape)
- Has `aria-activedescendant` for active item tracking
- Has unique IDs on menu items
- Closes on outside click
- Closes on Escape
- Skips disabled items during navigation
- Focus management with `focusout` handler

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Trigger lacks `aria-expanded` | 4.1.2 Name, Role, Value | **High** | Trigger doesn't announce expanded state |
| Trigger lacks `aria-haspopup="menu"` | 4.1.2 | **High** | Trigger doesn't announce popup type |
| No `aria-controls` linking trigger to menu | 4.1.2 | **Medium** | Relationship not programmatically defined |
| No focus restoration to trigger | 2.4.3 Focus Order | **Medium** | On Escape, focus stays on menu |
| Visible focus indicator not verified | 2.4.7 Focus Visible | **Low** | CSS `focus:` styles should be audited |

#### Recommendations
1. Add `aria-expanded` to trigger wrapper, toggle on open/close
2. Add `aria-haspopup="menu"` to trigger
3. Generate ID for menu and add `aria-controls` to trigger
4. Return focus to trigger element on close

---

### 3. CommandPalette (`components/command_palette.go`)

#### Current State
- Has FocusTrap (uses `focustrap.go`) - focus trapped correctly
- Has focus restoration (via FocusTrap.Deactivate)
- Has keyboard navigation (ArrowUp/ArrowDown/Enter/Escape)
- Has keyboard shortcut (Cmd+K / Ctrl+K)
- Prevents body scroll when open
- Auto-highlights first result on filter
- Visual keyboard hints in footer

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Container lacks `role="dialog"` | 4.1.2 Name, Role, Value | **High** | Not identified as dialog to screen readers |
| Container lacks `aria-modal="true"` | 4.1.2 | **High** | Modal nature not communicated |
| Results list lacks `role="listbox"` | 4.1.2 | **High** | List not identified as selection widget |
| Result items lack `role="option"` | 4.1.2 | **High** | Options not identified |
| No `aria-selected` on highlighted item | 4.1.2 | **High** | Current selection not announced |
| No `aria-activedescendant` | 4.1.2 | **Medium** | Active descendant pattern not used |
| Input lacks `aria-label` or `aria-labelledby` | 4.1.2 | **Medium** | Input has placeholder but no accessible name |
| Input lacks `aria-controls` for results | 4.1.2 | **Low** | Relationship not defined |
| No live region for result count | 4.1.3 Status Messages | **Low** | Filter result count not announced |

#### Recommendations
1. Add `role="dialog"` and `aria-modal="true"` to container
2. Add `role="listbox"` to results list
3. Add `role="option"` and `aria-selected` to result items
4. Add `aria-activedescendant` to results list
5. Add `aria-label="Search commands"` to input
6. Consider aria-live region for announcing result count changes

---

### 4. ConfirmDialog (`components/confirm_dialog.go`)

#### Current State
- Wraps Modal component
- Has Cancel and Confirm buttons in footer
- Closes on Escape (via Modal's `CloseOnEsc: true`)

#### Gaps

*Inherits all Modal gaps plus:*

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Missing `role="alertdialog"` | 4.1.2 Name, Role, Value | **High** | Should use alertdialog for confirmations |
| Missing `aria-describedby` | 4.1.2 | **Medium** | Message content not linked |
| No focus on destructive action | 2.4.3 Focus Order | **Medium** | Should focus Cancel for danger variant |
| Buttons lack full keyboard access testing | 2.1.1 Keyboard | **Low** | Needs Tab order verification |

#### Recommendations
1. Override Modal's role with `role="alertdialog"` for confirmation pattern
2. Add `aria-describedby` pointing to message content
3. Focus Cancel button (not Confirm) when variant is "danger"
4. Ensure Tab order is Cancel -> Confirm for safest default

---

### 5. Combobox (`components/combobox.go`)

#### Current State
- Has text input with filtering
- Has keyboard navigation (ArrowUp/ArrowDown/Enter/Escape)
- Closes on outside click
- Supports disabled state
- Supports required attribute
- Has placeholder text

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Missing `role="combobox"` | 4.1.2 Name, Role, Value | **High** | Input not identified as combobox |
| Missing `aria-expanded` | 4.1.2 | **High** | Dropdown state not communicated |
| Missing `aria-autocomplete="list"` | 4.1.2 | **High** | Autocomplete behavior not announced |
| Dropdown lacks `role="listbox"` | 4.1.2 | **High** | List not identified |
| Options lack `role="option"` | 4.1.2 | **High** | Options not identified |
| Missing `aria-activedescendant` | 4.1.2 | **High** | Highlighted option not announced |
| Missing `aria-controls` | 4.1.2 | **Medium** | Input-listbox relationship not defined |
| Label not associated with input | 1.3.1 Info and Relationships | **Medium** | Uses text label not `<label for>` pattern |
| Missing `aria-selected` on options | 4.1.2 | **Medium** | Selected state not communicated |
| No live region for results | 4.1.3 Status Messages | **Low** | Filter results not announced |

#### Recommendations
1. Add `role="combobox"` to input
2. Add `aria-expanded`, toggle on open/close
3. Add `aria-autocomplete="list"`
4. Add `role="listbox"` to dropdown, `role="option"` to items
5. Implement `aria-activedescendant` for highlighted option
6. Generate IDs and add `aria-controls`
7. Use proper `<label for="">` pattern for label-input association
8. Add `aria-selected` to options

---

## Structural Components

### 6. Tabs (`components/tabs.go`)

#### Current State
- Has clickable tab buttons
- Shows/hides corresponding panels
- Updates visual styling for active tab
- Has horizontal scrolling on mobile

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Tab list lacks `role="tablist"` | 4.1.2 Name, Role, Value | **High** | Container not identified |
| Tabs lack `role="tab"` | 4.1.2 | **High** | Buttons not identified as tabs |
| Panels lack `role="tabpanel"` | 4.1.2 | **High** | Content regions not identified |
| Missing `aria-selected` | 4.1.2 | **High** | Active tab not announced |
| Missing `aria-controls` | 4.1.2 | **High** | Tab-panel relationship not defined |
| Missing `aria-labelledby` on panels | 4.1.2 | **High** | Panels not labeled by tabs |
| No arrow key navigation | 2.1.1 Keyboard | **High** | ARIA tabs pattern requires arrow keys |
| No Home/End key support | 2.1.1 Keyboard | **Medium** | Pattern recommends these |
| Missing `tabindex` management | 2.1.1 | **Medium** | Only active tab should be tabbable |

#### Recommendations
1. Add `role="tablist"` to nav container
2. Add `role="tab"` to tab buttons
3. Add `role="tabpanel"` to panels
4. Add `aria-selected="true/false"` to tabs
5. Add unique IDs to tabs/panels and link with `aria-controls`/`aria-labelledby`
6. Implement arrow key navigation (Left/Right for horizontal tabs)
7. Manage `tabindex`: active tab = 0, others = -1
8. Tab key should skip inactive tabs and go directly to panel

---

### 7. Accordion (`components/accordion.go`)

#### Current State
- Uses button elements for headers
- Has visual chevron indicator that rotates
- Has animated expand/collapse
- Supports single or multiple open panels
- Has focus outline on header button

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Missing `aria-expanded` | 4.1.2 Name, Role, Value | **High** | Expanded state not announced |
| Content regions lack proper role | 4.1.2 | **Medium** | Consider `role="region"` or implicit section |
| Missing `aria-controls` | 4.1.2 | **Medium** | Header-content relationship not defined |
| Missing `id` attributes | 4.1.2 | **Medium** | Needed for ARIA relationships |
| No Enter/Space verification | 2.1.1 Keyboard | **Low** | Buttons should respond to both (native, likely works) |

#### Recommendations
1. Add `aria-expanded="true/false"` to header buttons, toggle on click
2. Generate unique IDs for content regions
3. Add `aria-controls="[content-id]"` to header buttons
4. Consider `role="region"` with `aria-labelledby` for content panels

---

### 8. Table (`components/table.go`)

#### Current State
- Uses semantic `<table>`, `<thead>`, `<tbody>` elements
- Has header cells in `<thead>`
- Has sortable column headers with visual indicators
- Has clickable rows (optional)
- Has checkbox selection (optional)
- Has pagination component
- Has filter input
- Has bulk action bar
- Has empty state handling
- Has export dropdown

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Missing `<th scope="col">` | 1.3.1 Info and Relationships | **High** | Headers not properly scoped |
| Missing `aria-sort` | 4.1.2 Name, Role, Value | **High** | Sort state not announced |
| Sort icons are emoji text | 4.1.2 | **Medium** | Consider `aria-hidden` + sr-only text |
| Filter input lacks `aria-label` | 4.1.2 | **Medium** | Has placeholder but no accessible name |
| Select-all checkbox lacks label | 4.1.2 | **Medium** | No accessible name |
| Row checkboxes lack labels | 4.1.2 | **Medium** | No accessible name per row |
| Missing `aria-selected` on selected rows | 4.1.2 | **Medium** | Selection state not announced |
| Bulk action count not live | 4.1.3 Status Messages | **Low** | Selection count changes not announced |
| Pagination accessibility | Various | **Low** | Needs separate audit |

#### Recommendations
1. Add `scope="col"` to all `<th>` elements
2. Add `aria-sort="ascending/descending/none"` to sortable headers
3. Add `aria-label="Sort by [column]"` or visually hidden text for sort indicators
4. Add `aria-label="Search table"` to filter input
5. Add `aria-label="Select all rows"` to select-all checkbox
6. Add `aria-label="Select row for [identifier]"` to row checkboxes
7. Add `aria-selected="true/false"` to selectable rows
8. Consider `aria-live="polite"` for bulk action count

---

## Summary

### Gap Count by Priority

| Priority | Count | Components Affected |
|----------|-------|---------------------|
| **High** | 28 | All 8 components |
| **Medium** | 18 | All 8 components |
| **Low** | 8 | Various |

### Most Critical Gaps (Immediate Remediation Needed)

1. **Modal**: Missing role, aria-modal, focus trap, focus restoration
2. **Tabs**: Missing all ARIA roles and keyboard navigation
3. **Combobox**: Missing combobox ARIA pattern entirely
4. **CommandPalette**: Missing dialog role and listbox pattern
5. **Table**: Missing th scope and aria-sort

### Components with Best Current State

1. **Dropdown**: Has role="menu", menuitem, keyboard nav, aria-activedescendant - needs trigger ARIA additions
2. **CommandPalette**: Has focus trap and restoration - needs ARIA roles
3. **Accordion**: Uses semantic buttons - needs aria-expanded

### Next Steps

1. Consolidate findings in 07-03 (Combined Findings)
2. Create remediation plan in Phase 08-11
3. Prioritize High gaps for Phase 08 (Core Components)
4. Address Medium gaps in Phase 09 (Forms & Navigation)
5. Address Low gaps and live regions in Phase 10 (Live & Focus)
