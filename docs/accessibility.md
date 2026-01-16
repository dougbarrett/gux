# Accessibility Guide

This document provides comprehensive accessibility guidelines for contributors working with Gux components. All components are built to meet WCAG 2.1 AA compliance, with proper ARIA attributes, keyboard navigation, and focus management.

> **Compliance Target:** WCAG 2.1 Level AA with 25+ accessible components and automated testing via axe-core.

## Table of Contents

- [ARIA Patterns Reference](#aria-patterns-reference)
- [Unique ID Generation](#unique-id-generation)
- [Keyboard Navigation Patterns](#keyboard-navigation-patterns)
- [Focus Management](#focus-management)
- [Visual Accessibility](#visual-accessibility)
- [Testing Accessibility](#testing-accessibility)
- [Contributing Guidelines](#contributing-guidelines)
- [WCAG Reference](#wcag-reference)

## ARIA Patterns Reference

### Dialog Pattern (Modal)

The Modal component implements the WAI-ARIA dialog pattern for accessible modal dialogs.

| Attribute | Value | WCAG Criterion |
|-----------|-------|----------------|
| `role` | `dialog` | 4.1.2 Name, Role, Value |
| `aria-modal` | `true` | Prevents screen reader from reading background |
| `aria-labelledby` | `{title-id}` | 4.1.2 Name, Role, Value |

```go
// Modal creates accessible dialog with proper ARIA
modal := document.Call("createElement", "div")
modal.Call("setAttribute", "role", "dialog")
modal.Call("setAttribute", "aria-modal", "true")
modal.Call("setAttribute", "aria-labelledby", titleID)

// Title with matching ID
title := document.Call("createElement", "h3")
title.Set("id", titleID)

// Close button with accessible name
closeBtn := document.Call("createElement", "button")
closeBtn.Call("setAttribute", "aria-label", "Close")
```

### Alertdialog Pattern (ConfirmDialog)

For destructive confirmations, use `role="alertdialog"` which announces immediately to screen readers.

| Attribute | Value | Purpose |
|-----------|-------|---------|
| `role` | `alertdialog` | Assertive announcement |
| `aria-describedby` | `{message-id}` | Links to confirmation message |

```go
// Override modal role for destructive actions
m.modal.ModalElement().Call("setAttribute", "role", "alertdialog")
m.modal.ModalElement().Call("setAttribute", "aria-describedby", messageID)
```

### Combobox/Listbox Pattern (CommandPalette, Combobox)

Searchable dropdowns use the combobox pattern with listbox results.

| Element | Role | Additional Attributes |
|---------|------|----------------------|
| Input | `combobox` | `aria-expanded`, `aria-haspopup="listbox"`, `aria-controls`, `aria-activedescendant` |
| Results container | `listbox` | — |
| Result items | `option` | `aria-selected`, `id` for activedescendant |
| Category headers | `presentation` | Not announced as options |

```go
// Input with combobox role
input.Call("setAttribute", "role", "combobox")
input.Call("setAttribute", "aria-expanded", "true")
input.Call("setAttribute", "aria-haspopup", "listbox")
input.Call("setAttribute", "aria-controls", resultsID)
input.Call("setAttribute", "aria-activedescendant", activeItemID)

// Results with listbox role
results.Call("setAttribute", "role", "listbox")

// Each option
option.Call("setAttribute", "role", "option")
option.Call("setAttribute", "aria-selected", isActive)
```

### Tablist Pattern (Tabs)

Tabbed interfaces use the tablist pattern with roving tabindex for keyboard navigation.

| Element | Role | Additional Attributes |
|---------|------|----------------------|
| Tab container | `tablist` | `aria-label` |
| Tab buttons | `tab` | `aria-selected`, `aria-controls`, `tabindex` |
| Tab panels | `tabpanel` | `aria-labelledby`, `tabindex="0"` |

**Roving Tabindex:** Only the active tab has `tabindex="0"`, inactive tabs have `tabindex="-1"`.

```go
// Tablist container
tabNav.Call("setAttribute", "role", "tablist")
tabNav.Call("setAttribute", "aria-label", "Tabs")

// Generate unique IDs
uuid := crypto.Call("randomUUID").String()
tabID := "tabs-tab-" + strconv.Itoa(i) + "-" + uuid
panelID := "tabs-panel-" + strconv.Itoa(i) + "-" + uuid

// Tab button
btn.Call("setAttribute", "role", "tab")
btn.Set("id", tabID)
btn.Call("setAttribute", "aria-controls", panelID)
btn.Call("setAttribute", "aria-selected", isActive)
btn.Call("setAttribute", "tabindex", tabindex) // 0 for active, -1 for inactive

// Tab panel
panel.Call("setAttribute", "role", "tabpanel")
panel.Set("id", panelID)
panel.Call("setAttribute", "aria-labelledby", tabID)
panel.Call("setAttribute", "tabindex", "0")
```

### Grid Pattern (DatePicker)

Calendar grids use the grid pattern for spatial navigation.

| Element | Role | Additional Attributes |
|---------|------|----------------------|
| Calendar | `grid` | `aria-label` with month/year |
| Day cells | `gridcell` | `aria-selected` for selected date |

**Navigation:** Arrow keys move through days, Enter/Space selects.

### Live Regions (Alert, Toast, Progress, Spinner)

Dynamic content uses ARIA live regions to announce changes.

| Component | Role | aria-live | When to Use |
|-----------|------|-----------|-------------|
| Alert (error/warning) | `alert` | `assertive` | Interrupts immediately |
| Alert (info/success) | `status` | `polite` | Waits for pause |
| Toast container | `status` | `polite` | Non-critical updates |
| Progress | `progressbar` | — | With `aria-valuenow/min/max` |
| Spinner | `status` | `polite` | With `aria-label` (default: "Loading") |

```go
// Alert with urgency-based role
if variant == AlertError || variant == AlertWarning {
    container.Call("setAttribute", "role", "alert")
    container.Call("setAttribute", "aria-live", "assertive")
} else {
    container.Call("setAttribute", "role", "status")
    container.Call("setAttribute", "aria-live", "polite")
}

// Progress with value range
progress.Call("setAttribute", "role", "progressbar")
progress.Call("setAttribute", "aria-valuemin", "0")
progress.Call("setAttribute", "aria-valuemax", "100")
progress.Call("setAttribute", "aria-valuenow", value) // Omit for indeterminate

// Spinner with label
spinner.Call("setAttribute", "role", "status")
spinner.Call("setAttribute", "aria-live", "polite")
spinner.Call("setAttribute", "aria-label", "Loading") // Customizable via AriaLabel prop
```

### Form Controls

Form inputs require proper label association and error messaging.

| Pattern | WCAG Criterion | Implementation |
|---------|----------------|----------------|
| Label association | 1.3.1, 4.1.2 | `htmlFor`/`id` or wrapping `<label>` |
| Error messages | 3.3.1 | `aria-invalid="true"` + `aria-describedby` |
| Required fields | 3.3.2 | `aria-required="true"` |

```go
// Label association
label := document.Call("createElement", "label")
label.Call("setAttribute", "for", inputID)
input.Set("id", inputID)

// Error state
input.Call("setAttribute", "aria-invalid", "true")
input.Call("setAttribute", "aria-describedby", errorID)

errorMsg := document.Call("createElement", "div")
errorMsg.Set("id", errorID)
errorMsg.Call("setAttribute", "role", "alert")
```

### Navigation

Navigation landmarks help screen reader users orient within the page.

```go
nav := document.Call("createElement", "nav")
nav.Call("setAttribute", "role", "navigation")
nav.Call("setAttribute", "aria-label", "Main navigation")

// Active item
link.Call("setAttribute", "aria-current", "page")
```

## Unique ID Generation

Use `crypto.randomUUID()` for unique ARIA IDs to link related elements.

```go
// Generate unique ID for aria-labelledby/describedby
titleID := "modal-title-" + js.Global().Get("crypto").Call("randomUUID").String()

// For indexed elements (tabs, accordion items)
tabID := "tabs-tab-" + strconv.Itoa(index) + "-" + crypto.Call("randomUUID").String()
panelID := "tabs-panel-" + strconv.Itoa(index) + "-" + crypto.Call("randomUUID").String()
```

**Why UUIDs?** Multiple instances of the same component on a page must have unique IDs for ARIA references to work correctly.

## Keyboard Navigation Patterns

### FocusTrap Component

The FocusTrap component manages focus within modal content, preventing focus from escaping to background elements.

**API:**

| Method | Description |
|--------|-------------|
| `NewFocusTrap(container)` | Create trap for container element |
| `Activate()` | Store current focus, add Tab handler, focus first element |
| `Deactivate()` | Remove handler, restore previous focus |
| `Destroy()` | Clean up js.Func resources |
| `FocusFirst()` | Focus first focusable element |
| `FocusLast()` | Focus last focusable element |
| `IsActive()` | Check if trap is active |

**Usage:**

```go
// Create focus trap
focusTrap := NewFocusTrap(modalContent)

// On modal open
focusTrap.Activate() // Stores trigger, focuses first focusable

// On modal close
focusTrap.Deactivate() // Restores focus to trigger

// Cleanup when component unmounts
focusTrap.Destroy()
```

**Focusable Elements Selector:**

```js
a[href]:not([disabled]):not([tabindex="-1"]),
button:not([disabled]):not([tabindex="-1"]),
textarea:not([disabled]):not([tabindex="-1"]),
input:not([disabled]):not([tabindex="-1"]):not([type="hidden"]),
select:not([disabled]):not([tabindex="-1"]),
[tabindex]:not([tabindex="-1"]):not([disabled])
```

### Focus Restoration Pattern

When overlays close, focus must return to the trigger element.

```go
// On open - store trigger
previousFocus := document.Get("activeElement")

// On close - restore focus
if !previousFocus.IsUndefined() && !previousFocus.IsNull() {
    previousFocus.Call("focus")
}
```

### Arrow Key Navigation (Roving Tabindex)

For Tabs, Dropdown, and DatePicker, use roving tabindex with arrow keys.

```go
// Keyboard handler
keyHandler := js.FuncOf(func(this js.Value, args []js.Value) any {
    event := args[0]
    key := event.Get("key").String()

    switch key {
    case "ArrowRight":
        event.Call("preventDefault")
        nextIdx := (activeIndex + 1) % len(items)
        setActive(nextIdx)
        items[nextIdx].Call("focus")
    case "ArrowLeft":
        event.Call("preventDefault")
        prevIdx := activeIndex - 1
        if prevIdx < 0 {
            prevIdx = len(items) - 1
        }
        setActive(prevIdx)
        items[prevIdx].Call("focus")
    case "Home":
        event.Call("preventDefault")
        setActive(0)
        items[0].Call("focus")
    case "End":
        event.Call("preventDefault")
        lastIdx := len(items) - 1
        setActive(lastIdx)
        items[lastIdx].Call("focus")
    }
    return nil
})
```

### Escape Key Handling

Close overlays on Escape and restore focus.

```go
document.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) any {
    if isOpen && args[0].Get("key").String() == "Escape" {
        Close() // Calls focusTrap.Deactivate() internally
    }
    return nil
}))
```

### Global Shortcuts

Command Palette registers `Cmd/Ctrl+K` globally.

```go
func (c *CommandPalette) RegisterKeyboardShortcut() {
    c.keyHandler = js.FuncOf(func(this js.Value, args []js.Value) any {
        event := args[0]
        key := event.Get("key").String()
        metaKey := event.Get("metaKey").Bool()
        ctrlKey := event.Get("ctrlKey").Bool()

        if (metaKey || ctrlKey) && (key == "k" || key == "K") {
            event.Call("preventDefault")
            c.Toggle()
        }
        return nil
    })
    js.Global().Get("document").Call("addEventListener", "keydown", c.keyHandler)
}
```

## Focus Management

### When to Use FocusTrap

| Component | FocusTrap Required |
|-----------|-------------------|
| Modal | ✅ Yes |
| ConfirmDialog | ✅ Yes (inherits from Modal) |
| CommandPalette | ✅ Yes |
| Dropdown | ❌ No (closes on focus outside) |
| Drawer | ✅ Yes |

### Focus Order (WCAG 2.4.3)

- Tab order follows visual order (left-to-right, top-to-bottom)
- Don't use positive tabindex values
- Hidden elements should have `tabindex="-1"` or be removed from DOM

### Skip Links

The SkipLink component allows keyboard users to bypass navigation.

```go
skipLinks := components.SkipLinks()
document.Get("body").Call("prepend", skipLinks)

// Targets #main-content landmark
main := document.Call("createElement", "main")
main.Set("id", "main-content")
```

## Visual Accessibility

### Focus Indicator Pattern

All interactive elements must have visible focus indicators (WCAG 2.4.7).

**Standard pattern:**

```css
focus:outline-none focus:ring-2 focus:ring-blue-500
```

**For full-width elements:**

```css
focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500
```

**With offset (standalone buttons):**

```css
focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
```

### Color Contrast (WCAG 1.4.3)

Minimum contrast ratios:
- **Normal text:** 4.5:1
- **Large text (18px+):** 3:1
- **UI components:** 3:1

**Established patterns:**
- Primary buttons: `bg-blue-600` (not `bg-blue-500`)
- Success buttons: `bg-green-700` (not `bg-green-600`)
- Secondary text: `text-gray-500` (not `text-gray-400`)

### Reduced Motion (WCAG 2.3.3)

Respect `prefers-reduced-motion` for users with vestibular disorders.

```go
// Check user preference
func PrefersReducedMotion() bool {
    window := js.Global()
    matchMedia := window.Call("matchMedia", "(prefers-reduced-motion: reduce)")
    return matchMedia.Get("matches").Bool()
}

// Usage in Animate()
func Animate(props AnimateProps) {
    if PrefersReducedMotion() {
        // Skip animation, fire callback immediately
        if props.OnComplete != nil {
            props.OnComplete()
        }
        return
    }
    // ... apply animation
}
```

**CSS fallback:**

```css
@media (prefers-reduced-motion: reduce) {
    *, *::before, *::after {
        animation-duration: 0.01ms !important;
        animation-iteration-count: 1 !important;
        transition-duration: 0.01ms !important;
    }
}
```

## Testing Accessibility

### Automated Testing with axe-core

Gux uses Playwright with axe-core for automated WCAG testing.

**Run tests:**

```bash
cd example
make test-a11y        # Run accessibility tests
make test-a11y-debug  # Run with visible browser
make test-report      # View HTML report
```

**WCAG tags tested:** `wcag2a`, `wcag2aa`, `wcag21a`, `wcag21aa`

**Test structure:**

```typescript
import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test('should have no accessibility violations', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('#app:not(:has-text("Loading"))');

    const results = await new AxeBuilder({ page })
        .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
        .analyze();

    expect(results.violations).toEqual([]);
});
```

### Manual Testing Checklist

Before submitting a component:

- [ ] **Keyboard-only:** Navigate using Tab, Arrow keys, Enter, Escape, Space
- [ ] **Screen reader:** Test with VoiceOver (macOS), NVDA (Windows), or JAWS
- [ ] **Zoom:** Content remains usable at 200% browser zoom
- [ ] **Focus visible:** Focus indicator visible on all interactive elements
- [ ] **Color contrast:** Verify with axe DevTools or Lighthouse
- [ ] **Reduced motion:** Test with system setting enabled

### Testing Tools

| Tool | Purpose |
|------|---------|
| axe DevTools | Browser extension for quick audits |
| Lighthouse | Chrome DevTools accessibility audit |
| WAVE | Browser extension for visual feedback |
| VoiceOver | macOS built-in screen reader |
| NVDA | Free Windows screen reader |

## Contributing Guidelines

### Checklist for New Components

When adding a new interactive component:

1. **ARIA Roles:** Add appropriate `role` attribute
2. **Labels:** Ensure accessible name via `aria-label` or `aria-labelledby`
3. **States:** Communicate state with `aria-expanded`, `aria-selected`, `aria-pressed`, etc.
4. **Keyboard:** All functionality accessible via keyboard
5. **Focus:** Visible focus indicator, proper focus management
6. **Landmarks:** Use semantic HTML or ARIA landmarks

### Use Existing Patterns

| Need | Use |
|------|-----|
| Focus trapping | `NewFocusTrap(container)` |
| Unique IDs | `crypto.randomUUID()` |
| Focus indicators | `focus:ring-2 focus:ring-blue-500` |
| Motion preference | `PrefersReducedMotion()` |

### Common Mistakes to Avoid

- ❌ Using `div` or `span` for interactive elements (use `button` or `a`)
- ❌ Positive `tabindex` values (disrupts natural tab order)
- ❌ Missing accessible names on icon-only buttons
- ❌ Color as the only way to convey information
- ❌ Auto-playing animations without reduced motion check
- ❌ Focus trapping without focus restoration

## WCAG Reference

### Key Success Criteria

| Criterion | Level | Description | Component Impact |
|-----------|-------|-------------|------------------|
| 1.3.1 Info and Relationships | A | Structure conveyed programmatically | Semantic HTML, ARIA |
| 1.4.3 Contrast (Minimum) | AA | 4.5:1 for text | Color choices |
| 2.1.1 Keyboard | A | All functionality via keyboard | Focus, handlers |
| 2.1.2 No Keyboard Trap | A | Focus can move away | FocusTrap + Escape |
| 2.4.3 Focus Order | A | Logical focus sequence | Tab order, roving |
| 2.4.7 Focus Visible | AA | Visible focus indicator | focus:ring classes |
| 2.3.3 Animation | AAA | Respect motion preference | PrefersReducedMotion |
| 3.3.1 Error Identification | A | Describe errors | aria-invalid, aria-describedby |
| 4.1.2 Name, Role, Value | A | Components have accessible names | ARIA attributes |

### Resources

- [WAI-ARIA Authoring Practices](https://www.w3.org/WAI/ARIA/apg/)
- [WCAG 2.1 Quick Reference](https://www.w3.org/WAI/WCAG21/quickref/)
- [MDN Accessibility Guide](https://developer.mozilla.org/en-US/docs/Web/Accessibility)
- [axe-core Rules](https://dequeuniversity.com/rules/axe/)

---

See also: [Keyboard Shortcuts](keyboard-shortcuts.md) | [Components Reference](components.md)
