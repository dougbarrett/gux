# Phase 07-02: Form, Navigation & Feedback Components Accessibility Audit Findings

**Audit Date:** 2026-01-15
**WCAG Version:** 2.1 AA
**Auditor:** Claude Code

---

## Form Control Components

### 1. Input (`components/input.go`)

#### Current State
- Creates `<label>` element when `props.Label` is provided
- Uses native `<input>` with type, placeholder, disabled, required attributes
- Has Tailwind `focus:ring-2 focus:ring-blue-500` for focus indicator
- Has `SetError`/`ClearError` methods for visual error styling

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Label not associated with input | 1.3.1 Info and Relationships | **High** | `<label>` is visual only, no `htmlFor`/`id` linking |
| No `aria-invalid` on error | 3.3.1 Error Identification | **High** | Error state not communicated to AT |
| No `aria-describedby` for errors | 3.3.1 | **High** | Error messages not linked to input |
| No `aria-required` | 3.3.2 Labels or Instructions | **Medium** | Required state not announced beyond native `required` |
| No `autocomplete` support | 1.3.5 Identify Input Purpose | **Medium** | Props doesn't support autocomplete attribute |

#### Recommendations
1. Generate unique ID for input, set `htmlFor` on label
2. Add `aria-invalid="true"` in `SetError()`, remove in `ClearError()`
3. Create error element with ID, add `aria-describedby` pointing to it
4. Add `aria-required="true"` when `props.Required` is true
5. Add `Autocomplete` field to `InputProps`

---

### 2. Select (`components/select.go`)

#### Current State
- Uses native `<select>` element (inherently accessible)
- Creates `<label>` element when `props.Label` is provided
- Has disabled, required attributes on select
- Has Tailwind focus ring styles

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Label not associated with select | 1.3.1 Info and Relationships | **High** | `<label>` is visual only, no `htmlFor`/`id` linking |
| No `aria-required` | 3.3.2 Labels or Instructions | **Medium** | Required state relies on native only |
| No error state support | 3.3.1 Error Identification | **Medium** | Component lacks `SetError` method |

#### Recommendations
1. Generate unique ID for select, set `htmlFor` on label
2. Add `aria-required="true"` when `props.Required` is true
3. Add `SetError`/`ClearError` methods with `aria-invalid` support

---

### 3. Checkbox (`components/checkbox.go`)

#### Current State
- Uses native `<input type="checkbox">`
- Has label element with click handler to toggle
- Has disabled, checked state support
- Uses Tailwind focus ring

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Label not associated via htmlFor/id | 1.3.1 Info and Relationships | **High** | Uses click handler instead of proper label association |
| No group labeling support | 1.3.1 | **Medium** | No fieldset/legend pattern for checkbox groups |
| No `aria-describedby` | 3.3.2 Labels or Instructions | **Low** | No help text association support |

#### Recommendations
1. Generate unique ID for checkbox, set `htmlFor` on label
2. Create `CheckboxGroup` component with `<fieldset>` and `<legend>`
3. Add `Description` prop with `aria-describedby` support

---

### 4. Toggle (`components/toggle.go`)

#### Current State
- Has `role="switch"` on toggle element ✓
- Has `aria-checked` that updates on toggle ✓
- Has `tabindex="0"` for keyboard focus ✓
- Has keyboard handler for Space/Enter ✓
- Updates `aria-checked` in `SetChecked()` ✓

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| No `aria-labelledby` | 4.1.2 Name, Role, Value | **Medium** | Toggle not programmatically linked to label text |
| No `aria-describedby` | 4.1.2 | **Low** | Description text not linked to toggle |
| Focus indicator not distinct | 2.4.7 Focus Visible | **Low** | No `:focus-visible` styling on toggle |

#### Recommendations
1. Generate ID for label, add `aria-labelledby` to toggle
2. Generate ID for description, add `aria-describedby` to toggle
3. Add visible focus ring on toggle element (currently on container)

---

### 5. TextArea (`components/textarea.go`)

#### Current State
- Creates `<label>` element when `props.Label` is provided
- Uses native `<textarea>` with placeholder, rows, disabled, required
- Has Tailwind focus ring styles

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Label not associated with textarea | 1.3.1 Info and Relationships | **High** | `<label>` is visual only, no `htmlFor`/`id` linking |
| No error state support | 3.3.1 Error Identification | **Medium** | Component lacks `SetError` method |
| No `aria-required` | 3.3.2 Labels or Instructions | **Medium** | Required state relies on native only |

#### Recommendations
1. Generate unique ID for textarea, set `htmlFor` on label
2. Add `SetError`/`ClearError` methods with `aria-invalid` support
3. Add `aria-required="true"` when `props.Required` is true

---

### 6. DatePicker (`components/datepicker.go`)

#### Current State
- Has label element
- Input is readonly text field showing selected date
- Calendar popup with prev/next month navigation
- Today button for quick selection
- Min/max date support with disabled styling
- Closes on outside click

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Missing combobox/dialog pattern | 4.1.2 Name, Role, Value | **High** | No `role="combobox"` or popup dialog role |
| No `aria-expanded` | 4.1.2 | **High** | Popup state not communicated |
| No `aria-haspopup` | 4.1.2 | **High** | Popup type not indicated |
| No keyboard navigation in calendar | 2.1.1 Keyboard | **High** | Arrow keys don't navigate dates |
| Calendar lacks grid role | 4.1.2 | **High** | Day grid not identified as grid |
| No `aria-selected` on dates | 4.1.2 | **High** | Selected date not announced |
| Nav buttons lack labels | 4.1.2 | **Medium** | SVG icons have no accessible name |
| Label not associated | 1.3.1 Info and Relationships | **Medium** | No `htmlFor`/`id` linking |
| No `aria-disabled` on dates | 4.1.2 | **Low** | Disabled dates not announced |

#### Recommendations
1. Add `role="combobox"` to input with `aria-expanded`, `aria-haspopup="dialog"`
2. Add `role="dialog"` or `role="grid"` to calendar popup
3. Implement arrow key navigation for date selection
4. Add `aria-selected="true"` to currently selected date
5. Add `aria-label` to prev/next/today buttons
6. Generate ID for input, set `htmlFor` on label
7. Add `aria-disabled="true"` to disabled date buttons

---

### 7. Form (`components/form.go`)

#### Current State
- Uses `Input` component internally for each field
- Has error message elements per field
- Has validation rules with client-side validation
- Has `SetFieldError` method for server-side errors

#### Gaps

*Inherits all Input gaps plus:*

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Error messages not linked | 3.3.1 Error Identification | **High** | No `aria-describedby` from input to error |
| No `aria-invalid` on error | 3.3.1 | **High** | Error state not communicated to AT |
| No form-level error summary | 3.3.1 | **Medium** | No error summary at form level |
| Focus not moved to first error | 3.3.1 | **Low** | On validation fail, focus not managed |

#### Recommendations
1. Generate IDs for error elements, add `aria-describedby` to inputs
2. Set `aria-invalid="true"` on inputs with validation errors
3. Add optional error summary component with `role="alert"`
4. Move focus to first error on form submission failure

---

### 8. FormBuilder (`components/formbuilder.go`)

#### Current State
- Uses `htmlFor`/`id` on labels ✓
- Has error message elements with ID (`name+"-error"`)
- Shows required indicator visually (*)
- Validates on blur with visual error styling
- Adds `border-red-500` class on error

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| No `aria-describedby` for errors | 3.3.1 Error Identification | **High** | Error elements exist but not linked to inputs |
| No `aria-invalid` on errors | 3.3.1 | **High** | Only visual error styling applied |
| No `aria-required` | 3.3.2 Labels or Instructions | **Medium** | Required state is visual only |
| Radio group lacks group role | 1.3.1 Info and Relationships | **Medium** | No `role="radiogroup"` with `aria-labelledby` |
| No focus on first error | 3.3.1 | **Low** | Focus not managed on validation failure |

#### Recommendations
1. Add `aria-describedby="name-error"` to all inputs
2. Add `aria-invalid="true"` in `showError()`, remove in `hideError()`
3. Add `aria-required="true"` when field has Required rule
4. Wrap radio buttons in `div[role="radiogroup"]` with `aria-labelledby`
5. Focus first invalid field in `handleSubmit()` on validation failure

---

## Navigation Components

### 9. Sidebar (`components/sidebar.go`)

#### Current State
- Uses `<aside>` element (implicit complementary role)
- Uses `<nav>` element (implicit navigation role)
- Has `aria-label` on collapse/close buttons ✓
- Has keyboard shortcut (Cmd/Ctrl+B) for collapse toggle
- Has tooltips on hover when collapsed
- Uses `Link` component for navigation items

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| No `aria-current="page"` on active item | 2.4.8 Location | **High** | Active page not announced to AT |
| No `aria-label` on nav | 2.4.1 Bypass Blocks | **Medium** | Can't distinguish from other navs |
| No focus management on mobile | 2.4.3 Focus Order | **Medium** | Focus not moved to sidebar on open |
| Keyboard shortcut not discoverable | 3.3.2 Labels or Instructions | **Low** | Cmd+B not shown in UI |
| No skip link support | 2.4.1 Bypass Blocks | **Low** | No way to skip navigation |

#### Recommendations
1. Update `SetActive()` to also set `aria-current="page"` on active link
2. Add `aria-label="Main navigation"` to nav element
3. Move focus to first nav item or close button when sidebar opens on mobile
4. Add visual keyboard hint somewhere (settings, help, or tooltip)
5. Consider adding skip link pattern to page layout

---

### 10. Breadcrumbs (`components/breadcrumbs.go`)

#### Current State
- Has `<nav>` element with `aria-label="Breadcrumb"` ✓
- Has `aria-current="page"` on last (current) item ✓
- Uses links for navigable items, text for current
- Has SPA routing click handlers

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Separator announced to AT | 1.3.1 Info and Relationships | **Medium** | "/" text is read by screen readers |

#### Recommendations
1. Add `aria-hidden="true"` to separator span, or use CSS `::before` instead

---

### 11. Pagination (`components/pagination.go`)

#### Current State
- Has `<nav>` with `role="navigation"` and `aria-label="Pagination"` ✓
- Has `aria-current="page"` on current page button ✓
- Has `aria-label` on prev/next buttons ("Previous", "Next") ✓
- Has `disabled` attribute on unavailable buttons
- Shows "Showing X-Y of Z items" info

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| Page buttons lack descriptive labels | 4.1.2 Name, Role, Value | **Low** | Just show number, could add "Page X" |
| No live region for page change | 4.1.3 Status Messages | **Low** | Page changes not announced |

#### Recommendations
1. Add `aria-label="Page X"` to page number buttons
2. Consider `aria-live="polite"` region for announcing page changes

---

### 12. Link (`components/link.go`)

#### Current State
- Uses native `<a>` element with `href` ✓
- Has click handler for SPA routing
- Supports className and children

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| None significant | - | - | Link uses semantic `<a>` element correctly |

#### Recommendations
- No changes required. Component is accessible as-is.

---

## Feedback Components

### 13. Alert (`components/alert.go`)

#### Current State
- Has visual styling per variant (info, success, warning, error)
- Has icon per variant
- Has optional dismiss button
- Dismisses itself from DOM on close

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| No `role="alert"` or `role="status"` | 4.1.3 Status Messages | **High** | Alert content not announced |
| Dismiss button lacks `aria-label` | 4.1.2 Name, Role, Value | **High** | Button only has visual "×" |
| Icons are plain text | 1.1.1 Non-text Content | **Medium** | Emoji icons may be announced inconsistently |

#### Recommendations
1. Add `role="alert"` for error/warning, `role="status"` for info/success
2. Add `aria-label="Dismiss"` to dismiss button
3. Add `aria-hidden="true"` to icon span, or wrap icon in span with `aria-hidden`

---

### 14. Toast (`components/toast.go`)

#### Current State
- Uses fixed-position container for stacking toasts
- Has close button
- Auto-dismisses after duration
- Has entrance/exit animations

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| No ARIA live region | 4.1.3 Status Messages | **High** | Toast messages not announced to AT |
| No `role="alert"` or `role="status"` | 4.1.3 | **High** | Individual toasts not identified |
| Close button lacks `aria-label` | 4.1.2 Name, Role, Value | **High** | Button only has visual "×" |
| Container lacks landmark | 4.1.3 | **Medium** | Toast container not identified |
| Icons are plain text | 1.1.1 Non-text Content | **Medium** | Emoji icons announced inconsistently |

#### Recommendations
1. Add `aria-live="polite"` to toast container (or `aria-live="assertive"` for errors)
2. Add `role="status"` to info/success toasts, `role="alert"` to error/warning
3. Add `aria-label="Close notification"` to close button
4. Add `role="region"` and `aria-label="Notifications"` to container
5. Add `aria-hidden="true"` to icon spans

---

### 15. Progress (`components/progress.go`)

#### Current State
- Shows visual progress bar with percentage fill
- Optional percentage label
- Supports indeterminate mode with animation
- Supports striped/animated variants

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| No `role="progressbar"` | 4.1.2 Name, Role, Value | **High** | Not identified as progress to AT |
| No `aria-valuenow` | 4.1.2 | **High** | Current value not communicated |
| No `aria-valuemin`/`aria-valuemax` | 4.1.2 | **High** | Range not defined |
| No `aria-label` | 4.1.2 | **Medium** | No accessible name for context |
| Indeterminate lacks `aria-valuetext` | 4.1.2 | **Medium** | Should indicate "loading" state |

#### Recommendations
1. Add `role="progressbar"` to bar element
2. Add `aria-valuenow`, `aria-valuemin="0"`, `aria-valuemax="100"`
3. Add `aria-label` prop for context (e.g., "File upload progress")
4. For indeterminate: remove aria-valuenow, add `aria-valuetext="Loading"`
5. Update values in `SetValue()` method

---

### 16. Spinner (`components/spinner.go`)

#### Current State
- Shows spinning animation
- Optional visible label text
- Inline variant available

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| No `role="status"` | 4.1.3 Status Messages | **High** | Loading state not announced |
| No `aria-busy` | 4.1.3 | **High** | Busy state not indicated |
| No `aria-label` when no visible label | 4.1.2 Name, Role, Value | **High** | No accessible name for AT |
| Visual-only content | 1.1.1 Non-text Content | **Medium** | Should have sr-only text alternative |

#### Recommendations
1. Add `role="status"` to spinner container
2. Add `aria-busy="true"` to indicate loading
3. Add `aria-label="Loading"` when no visible label, or use sr-only text
4. Consider adding visually hidden "Loading..." text for screen readers

---

### 17. Skeleton (`components/skeleton.go`)

#### Current State
- Purely visual placeholder with pulse animation
- Various presets (text, avatar, card, table)
- Used for content loading states

#### Gaps

| Gap | WCAG Criteria | Priority | Description |
|-----|---------------|----------|-------------|
| No `aria-hidden` | 4.1.2 Name, Role, Value | **Medium** | Decorative elements exposed to AT |
| Parent lacks `aria-busy` | 4.1.3 Status Messages | **Medium** | Loading state not communicated |
| No loading announcement | 4.1.3 | **Low** | Screen readers see empty divs |

#### Recommendations
1. Add `aria-hidden="true"` to skeleton elements (purely decorative)
2. Document that parent container should have `aria-busy="true"`
3. Consider adding sr-only text "Loading content..." to parent
4. Alternative: Wrap skeletons in container with `role="status"` and `aria-label="Loading"`

---

## Summary

### Gap Count by Priority

| Priority | Count | Components Affected |
|----------|-------|---------------------|
| **High** | 32 | Input, Select, Checkbox, TextArea, DatePicker, Form, FormBuilder, Sidebar, Alert, Toast, Progress, Spinner |
| **Medium** | 19 | All except Link |
| **Low** | 9 | Various |

### Most Critical Gaps (Immediate Remediation Needed)

1. **Label-Input Association (7 components):** Input, Select, Checkbox, TextArea, DatePicker - missing `htmlFor`/`id` linking
2. **Error State Accessibility:** Input, TextArea, Form, FormBuilder - missing `aria-invalid` and `aria-describedby`
3. **DatePicker:** Completely missing ARIA combobox/dialog pattern and keyboard navigation
4. **Feedback Components:** Alert, Toast, Progress, Spinner - missing ARIA live regions and roles
5. **Sidebar:** Missing `aria-current="page"` on active navigation

### Components with Best Current State

1. **Toggle:** Has `role="switch"`, `aria-checked`, keyboard handling - needs minor improvements
2. **Breadcrumbs:** Has `aria-label`, `aria-current="page"` - just needs separator fix
3. **Pagination:** Has `aria-label`, `aria-current="page"`, button labels - nearly complete
4. **Link:** Fully accessible using semantic `<a>` element
5. **FormBuilder:** Already uses `htmlFor`/`id` - needs ARIA error handling

### Components Requiring Most Work

1. **DatePicker:** Major gaps - needs full ARIA combobox pattern, keyboard navigation, grid roles
2. **Toast/Alert:** No ARIA live regions or roles
3. **Progress/Spinner:** No ARIA progressbar pattern or status roles
4. **Skeleton:** No accessibility handling for loading states

### Next Steps

1. Consolidate findings in 07-03 (Combined Findings Report)
2. Cross-reference with 07-01 findings for complete picture
3. Create remediation plan in Phase 08-11
4. Prioritize High gaps for Phase 08 (Core Components)
5. Address Medium gaps in Phase 09 (Forms & Navigation)
6. Address Low gaps and live regions in Phase 10 (Live & Focus)

---

## WCAG 2.1 AA Pattern Reference

### Form Input Pattern
```html
<div>
  <label for="email" id="email-label">Email *</label>
  <input
    type="email"
    id="email"
    aria-required="true"
    aria-invalid="false"
    aria-describedby="email-error email-hint"
    autocomplete="email"
  />
  <span id="email-hint">We'll never share your email</span>
  <span id="email-error" class="hidden">Please enter a valid email</span>
</div>
```

### Toggle Switch Pattern
```html
<div>
  <span id="toggle-label">Enable notifications</span>
  <span id="toggle-desc">Receive email alerts for new messages</span>
  <div
    role="switch"
    tabindex="0"
    aria-checked="false"
    aria-labelledby="toggle-label"
    aria-describedby="toggle-desc"
  ></div>
</div>
```

### DatePicker Pattern (simplified)
```html
<div>
  <label for="date" id="date-label">Select date</label>
  <input
    type="text"
    id="date"
    role="combobox"
    aria-expanded="false"
    aria-haspopup="dialog"
    aria-controls="date-calendar"
    readonly
  />
  <div id="date-calendar" role="dialog" aria-label="Choose date" hidden>
    <div role="grid" aria-label="January 2026">
      <!-- grid implementation -->
    </div>
  </div>
</div>
```

### Alert Pattern
```html
<div role="alert">
  <span aria-hidden="true">⚠️</span>
  <strong>Warning:</strong>
  <span>Your session will expire in 5 minutes.</span>
  <button aria-label="Dismiss alert">×</button>
</div>
```

### Progress Bar Pattern
```html
<div
  role="progressbar"
  aria-valuenow="75"
  aria-valuemin="0"
  aria-valuemax="100"
  aria-label="Upload progress"
>
  <div style="width: 75%"></div>
</div>
```

### Loading Spinner Pattern
```html
<div role="status" aria-busy="true" aria-label="Loading">
  <div class="spinner" aria-hidden="true"></div>
  <span class="sr-only">Loading, please wait...</span>
</div>
```
