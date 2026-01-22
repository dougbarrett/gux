# Phase 19: Interactive Components - Research

**Researched:** 2026-01-22
**Domain:** Interactive UI components (Modal, Dropdown, Tabs, Toast, Tooltip) for SSR+WASM
**Confidence:** HIGH

## Summary

Phase 19 implements interactive components that require state management and dynamic behavior: Modal/Dialog, Dropdown menu, Tabs with panels, Toast notifications, and Tooltip. These components build on the established `core.Node` and `ui/` package patterns from Phases 16-18.

The key challenge is that interactive components need to work in both SSR (static HTML) and WASM (interactive). The existing codebase has WASM-only implementations in `components/` that use `js.Value` directly - these must be reimplemented using `core.Node` for SSR compatibility. The pattern established in Phases 16-18 shows how: **components are pure functions returning `core.Node`**, state lives at the page level via `r.StateBool()` etc., and interactivity works via `OnClick` handlers that toggle state.

Critical insight: For overlays (Modal, Dropdown, Tooltip), SSR renders them in a closed/hidden state. The WASM hydration then makes them interactive. The `hidden` CSS class or conditional rendering (`core.If`) controls visibility, and `OnClick` handlers toggle state.

**Primary recommendation:** Follow the established Props struct pattern. Each interactive component renders its full DOM structure with appropriate ARIA attributes. Visibility is controlled via CSS (`hidden` class) or conditional rendering, driven by boolean props. Page-level state drives the props, and OnClick handlers toggle that state.

## Standard Stack

### Core (From Phases 16-18)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/dougbarrett/gux/core | local | Node interface, elements, Attrs | Foundation of all components |
| github.com/dougbarrett/gux/ui | local | Component library | Where interactive components go |
| MergeClasses | ui/utils.go | CSS class combination | Every component |
| ConditionalClass | ui/utils.go | Conditional CSS | State-dependent styling |

### Supporting Patterns
| Pattern | Source | Purpose |
|---------|--------|---------|
| Props struct | ui/button.go | Type-safe component configuration |
| Variant maps | ui/badge.go | Type-safe style variants |
| Compound components | ui/card.go | Related components that compose together |
| boolToAttr() | ui/switch.go | Boolean to "true"/"false" for ARIA |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| CSS visibility | Conditional rendering | CSS visibility preserves DOM for accessibility, conditional is simpler |
| core.Attrs.Extra | New Attrs fields | Extra is flexible but less discoverable |
| Page-level state | Component-internal state | Page-level required for SSR+WASM compatibility |

**Installation:**
No additional packages needed. All components build on existing `core` and `ui` packages.

## Architecture Patterns

### Recommended Project Structure
```
ui/
  modal.go           # Modal, ModalHeader, ModalContent, ModalFooter
  modal_test.go
  dropdown.go        # Dropdown, DropdownTrigger, DropdownMenu, DropdownItem
  dropdown_test.go
  tabs.go            # Tabs, TabList, Tab, TabPanel
  tabs_test.go
  toast.go           # ToastContainer, Toast
  toast_test.go
  tooltip.go         # Tooltip
  tooltip_test.go
```

### Pattern 1: Controlled Visibility Component (Modal)

**What:** Component renders full structure, visibility controlled by boolean prop
**When to use:** Overlays that need to be toggled (Modal, Dropdown)

```go
// Source: Established pattern from ui/checkbox.go (checked prop), adapted for visibility

type ModalProps struct {
    Open        bool        // Controls visibility
    OnClose     func()      // Called when close is requested (ESC, overlay click, X button)
    Title       string      // Optional title (sets aria-labelledby)
    Size        ModalSize   // sm, md, lg, xl
    Class       string      // Additional wrapper classes
    Children    []core.Node // Content
}

func Modal(props ModalProps) core.Node {
    // Always render the structure - visibility controlled by CSS
    overlayClass := MergeClasses(
        "fixed inset-0 bg-black/50 flex items-center justify-center z-50",
        ConditionalClass(!props.Open, "hidden"),
    )

    modalClass := MergeClasses(
        "bg-white dark:bg-gray-800 rounded-lg shadow-xl max-h-[90vh] overflow-auto",
        modalSizeClasses[props.Size],
        props.Class,
    )

    // ARIA attributes for accessibility
    modalAttrs := core.Attrs{
        Class: modalClass,
        Extra: map[string]string{
            "role":       "dialog",
            "aria-modal": "true",
        },
    }
    if props.Title != "" {
        modalAttrs.Extra["aria-labelledby"] = "modal-title"
    }

    return core.Div(core.Class(overlayClass),
        core.Div(modalAttrs, props.Children...),
    )
}
```

### Pattern 2: Tab Selection Pattern

**What:** Multiple panels, one visible at a time, controlled by index
**When to use:** Tabs, accordions, wizards

```go
// Source: WAI-ARIA Tabs pattern

type TabsProps struct {
    ActiveIndex int         // Which tab is active (0-indexed)
    OnChange    func(int)   // Called when tab selection changes
    Class       string      // Additional wrapper classes
    Children    []core.Node // TabList and TabPanels
}

type TabProps struct {
    Index    int         // Tab index
    Active   bool        // Whether this tab is active
    OnSelect func()      // Called when tab is clicked
    Children []core.Node // Tab label content
}

type TabPanelProps struct {
    Index    int         // Panel index
    Active   bool        // Whether this panel is visible
    Children []core.Node // Panel content
}

func Tab(props TabProps) core.Node {
    class := MergeClasses(
        "px-4 py-2 font-medium cursor-pointer border-b-2 transition-colors",
        ConditionalClass(props.Active, "border-blue-500 text-blue-600"),
        ConditionalClass(!props.Active, "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"),
    )

    return core.Button(core.Attrs{
        Class:   class,
        OnClick: props.OnSelect,
        Extra: map[string]string{
            "role":          "tab",
            "aria-selected": boolToAttr(props.Active),
            "tabindex":      tabindexValue(props.Active), // "0" if active, "-1" otherwise
        },
    }, props.Children...)
}

func TabPanel(props TabPanelProps) core.Node {
    return core.Div(core.Attrs{
        Class: ConditionalClass(!props.Active, "hidden"),
        Extra: map[string]string{
            "role":     "tabpanel",
            "tabindex": "0",
        },
    }, props.Children...)
}
```

### Pattern 3: Portal-Free Toast Container

**What:** Fixed-position container for notifications, no portal needed
**When to use:** Toast notifications, alerts

```go
// Source: Existing components/toast.go behavior, adapted for core.Node

// ToastContainerProps configures the toast container.
// The container should be rendered once at the app root level.
type ToastContainerProps struct {
    Position ToastPosition // top-right (default), top-left, bottom-right, bottom-left
    Class    string
}

// ToastContainer creates a fixed-position container for toast notifications.
// Render this once at the top level of your app layout.
//
// Note: In SSR, this renders an empty container. In WASM, toasts are added
// dynamically. The container uses aria-live="polite" for accessibility.
func ToastContainer(props ToastContainerProps) core.Node {
    position := props.Position
    if position == "" {
        position = ToastTopRight
    }

    class := MergeClasses(
        "fixed z-50 flex flex-col gap-2",
        toastPositionClasses[position],
        props.Class,
    )

    return core.Div(core.Attrs{
        ID:    "toast-container",
        Class: class,
        Extra: map[string]string{
            "role":        "status",
            "aria-live":   "polite",
            "aria-atomic": "true",
        },
    })
}

// ToastProps configures a single toast notification.
type ToastProps struct {
    Message string
    Variant ToastVariant // info, success, warning, error
    OnClose func()       // Called when close button clicked
}

// Toast creates a single toast notification.
// Typically used with page-level state to manage visible toasts.
func Toast(props ToastProps) core.Node {
    variant := props.Variant
    if variant == "" {
        variant = ToastInfo
    }

    style := toastVariantStyles[variant]

    return core.Div(core.Attrs{
        Class: MergeClasses(style.bg, style.text, "px-4 py-3 rounded-lg shadow-lg flex items-center gap-3 min-w-64"),
    },
        core.Span(core.Attrs{
            Class: "text-lg",
            Extra: map[string]string{"aria-hidden": "true"},
        }, core.Text(style.icon)),
        core.Span(core.Class("flex-1"), core.Text(props.Message)),
        core.If(props.OnClose != nil,
            core.Button(core.Attrs{
                Class:   "opacity-70 hover:opacity-100 cursor-pointer text-lg",
                OnClick: props.OnClose,
                Extra:   map[string]string{"aria-label": "Dismiss notification"},
            }, core.Text("x")),
        ),
    )
}
```

### Pattern 4: Tooltip with ARIA

**What:** Popup text on hover/focus using aria-describedby
**When to use:** Additional context for UI elements

```go
// Source: WAI-ARIA Tooltip pattern

type TooltipProps struct {
    Content   string          // Tooltip text
    Position  TooltipPosition // top, bottom, left, right
    Class     string          // Additional classes for trigger wrapper
    Children  []core.Node     // The trigger element(s)
}

// Tooltip wraps content with a tooltip that appears on hover/focus.
//
// Note: The tooltip visibility is CSS-driven (group-hover), not state-driven.
// This works in SSR because the tooltip is always in the DOM but hidden via CSS.
func Tooltip(props TooltipProps) core.Node {
    position := props.Position
    if position == "" {
        position = TooltipTop
    }

    // Wrapper with group for hover state
    wrapperClass := MergeClasses("relative inline-block group", props.Class)

    // Tooltip element - hidden until hover
    tooltipClass := MergeClasses(
        "absolute z-50 px-2 py-1 text-sm text-white bg-gray-900 rounded shadow-lg whitespace-nowrap",
        "opacity-0 invisible group-hover:opacity-100 group-hover:visible",
        "transition-all duration-200 pointer-events-none",
        tooltipPositionClasses[position],
    )

    return core.Div(core.Class(wrapperClass),
        // Trigger content
        core.Frag(props.Children...),
        // Tooltip element
        core.Div(core.Attrs{
            Class: tooltipClass,
            Extra: map[string]string{"role": "tooltip"},
        }, core.Text(props.Content)),
    )
}
```

### Anti-Patterns to Avoid

- **js.Value in UI components:** Use `core.Node` for SSR compatibility
- **Component-internal state:** State must live at page level for SSR hydration
- **Portal rendering:** Not supported in core.Node - use fixed positioning instead
- **Focus trapping in SSR:** Focus trap only works in WASM - document this limitation
- **Auto-dismiss without state:** Toast auto-dismiss requires page-level timer state

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Boolean ARIA values | "true"/"false" strings | boolToAttr() helper | Consistency, avoids errors |
| CSS class toggling | Ternary in templates | ConditionalClass() | Cleaner, tested |
| Visibility toggling | Custom logic | `hidden` class + ConditionalClass | Standard Tailwind pattern |
| ARIA role/state | Inline strings | Centralized constants | Maintainability |
| Unique IDs | Random generation | Auto-increment or crypto.randomUUID in WASM | SSR determinism |

**Key insight:** Interactive components are simpler in this architecture than in React because there's no virtual DOM diffing. The component renders, state changes trigger full re-render. CSS handles transitions.

## Common Pitfalls

### Pitfall 1: SSR/WASM State Mismatch

**What goes wrong:** Modal renders open in SSR but closed after hydration
**Why it happens:** Initial state differs between SSR and WASM
**How to avoid:** Initial state props must be consistent. Always default overlays to closed.

```go
// Page-level state - same initial value for SSR and WASM
modalOpen := r.StateBool("modalOpen", false) // Always start closed
```

**Warning signs:** Flash of content on hydration, hydration mismatch warnings

### Pitfall 2: Focus Management in SSR

**What goes wrong:** Focus trap code runs in SSR (where there's no DOM)
**Why it happens:** Focus trap uses js.Value which only works in WASM
**How to avoid:** Focus trap must be WASM-only. Document that keyboard navigation for modals/dropdowns requires WASM.

```go
// Focus trap is a WASM-only feature
// In SSR, the modal renders but focus management is not available
// This is acceptable because SSR is for initial load - interaction requires hydration
```

**Warning signs:** Compile errors, runtime panics in SSR

### Pitfall 3: ID Collisions

**What goes wrong:** Multiple modals/tooltips with same ARIA IDs
**Why it happens:** Hard-coded IDs like "modal-title"
**How to avoid:** Accept ID as prop, or generate unique IDs

```go
// Option 1: Accept ID as prop
type ModalProps struct {
    ID string // Used for aria-labelledby
}

// Option 2: Use prop-based ID
titleID := "modal-title-" + props.ID
```

**Warning signs:** ARIA relationships broken, accessibility audit failures

### Pitfall 4: Dropdown Outside Click

**What goes wrong:** Dropdown doesn't close when clicking outside
**Why it happens:** OnClick only fires on the element, not "outside"
**How to avoid:** Two approaches:
1. Render backdrop overlay when open (catches outside clicks)
2. Use `onBlur` event (loses focus = close)

```go
// Approach 1: Backdrop overlay
core.If(props.Open,
    core.Div(core.Attrs{
        Class:   "fixed inset-0 z-40",
        OnClick: props.OnClose, // Clicking backdrop closes dropdown
    }),
)
```

**Warning signs:** Dropdown stays open when clicking elsewhere

### Pitfall 5: Toast ARIA Live Region Timing

**What goes wrong:** Screen reader doesn't announce toast
**Why it happens:** Toast container must exist before content is added
**How to avoid:** Toast container must be in SSR output (empty). Toasts are added after hydration.

```go
// ToastContainer renders in SSR (empty)
// Toasts are rendered conditionally based on state AFTER hydration
// The aria-live region needs to exist before content changes
```

**Warning signs:** Toasts appear but aren't announced by screen readers

### Pitfall 6: Tooltip Keyboard Accessibility

**What goes wrong:** Tooltip only shows on hover, not keyboard focus
**Why it happens:** CSS group-hover doesn't handle focus
**How to avoid:** Include focus-within in tooltip visibility

```go
// CSS classes that work for both hover and focus
"group-hover:visible group-focus-within:visible"
"group-hover:opacity-100 group-focus-within:opacity-100"
```

**Warning signs:** Keyboard users can't see tooltips

## Accessibility Requirements

### Modal/Dialog
| Attribute | Value | Purpose |
|-----------|-------|---------|
| role | "dialog" | Identifies as dialog |
| aria-modal | "true" | Indicates modal behavior |
| aria-labelledby | title ID | Associates with title |
| aria-describedby | content ID | Associates with description (optional) |

**Keyboard:** Tab cycles within modal, Escape closes, focus trapped (WASM only)

### Dropdown Menu
| Attribute | Value | Purpose |
|-----------|-------|---------|
| role | "menu" | On menu container |
| role | "menuitem" | On each item |
| aria-haspopup | "menu" | On trigger button |
| aria-expanded | "true"/"false" | On trigger button |
| aria-controls | menu ID | On trigger button |

**Keyboard:** Arrow keys navigate, Enter selects, Escape closes

### Tabs
| Attribute | Value | Purpose |
|-----------|-------|---------|
| role | "tablist" | On tab container |
| role | "tab" | On each tab |
| role | "tabpanel" | On each panel |
| aria-selected | "true"/"false" | On tabs |
| aria-controls | panel ID | On tabs |
| aria-labelledby | tab ID | On panels |

**Keyboard:** Left/Right arrows switch tabs, Home/End to first/last

### Toast
| Attribute | Value | Purpose |
|-----------|-------|---------|
| role | "status" | On container |
| aria-live | "polite" | Announces changes |
| aria-atomic | "true" | Announces full content |

**Keyboard:** Toasts should not steal focus, dismiss button is focusable

### Tooltip
| Attribute | Value | Purpose |
|-----------|-------|---------|
| role | "tooltip" | On tooltip element |
| aria-describedby | tooltip ID | On trigger element |

**Keyboard:** Escape dismisses (optional), focus shows tooltip

## Code Examples

### Complete Modal Implementation

```go
package ui

import "github.com/dougbarrett/gux/core"

// ModalSize defines modal width variants.
type ModalSize string

const (
    ModalSM   ModalSize = "sm"
    ModalMD   ModalSize = "md"
    ModalLG   ModalSize = "lg"
    ModalXL   ModalSize = "xl"
    ModalFull ModalSize = "full"
)

var modalSizeClasses = map[ModalSize]string{
    ModalSM:   "max-w-sm w-full mx-4",
    ModalMD:   "max-w-md w-full mx-4",
    ModalLG:   "max-w-lg w-full mx-4",
    ModalXL:   "max-w-xl w-full mx-4",
    ModalFull: "max-w-4xl w-full mx-4",
}

// ModalProps configures the Modal component.
type ModalProps struct {
    Open     bool        // Controls visibility
    OnClose  func()      // Close handler (X button, backdrop click, Escape key in WASM)
    Title    string      // Modal title (optional, sets aria-labelledby)
    Size     ModalSize   // Width size (default: md)
    Class    string      // Additional content wrapper classes
    Children []core.Node // Modal content
}

// Modal creates an accessible modal dialog.
//
// The modal is always rendered but hidden via CSS when not open.
// This ensures consistent SSR output and smooth WASM hydration.
//
// Example:
//
//     modalOpen := r.StateBool("modal", false)
//     Modal(ModalProps{
//         Open:    modalOpen.Get(),
//         OnClose: func() { modalOpen.Set(false) },
//         Title:   "Confirm Action",
//         Children: []core.Node{
//             ModalContent(ModalContentProps{
//                 Children: []core.Node{core.Text("Are you sure?")},
//             }),
//             ModalFooter(ModalFooterProps{
//                 Children: []core.Node{
//                     Button(ButtonProps{OnClick: func() { modalOpen.Set(false) }},
//                         core.Text("Cancel")),
//                     Button(ButtonProps{Variant: ButtonPrimary, OnClick: confirm},
//                         core.Text("Confirm")),
//                 },
//             }),
//         },
//     })
func Modal(props ModalProps) core.Node {
    size := props.Size
    if size == "" {
        size = ModalMD
    }

    // Overlay - covers entire screen, hidden when closed
    overlayClass := MergeClasses(
        "fixed inset-0 bg-black/50 flex items-center justify-center z-50 transition-opacity",
        ConditionalClass(!props.Open, "hidden"),
    )

    // Modal container
    modalClass := MergeClasses(
        "bg-white dark:bg-gray-800 rounded-lg shadow-xl max-h-[90vh] flex flex-col",
        modalSizeClasses[size],
    )

    // Build ARIA attributes
    modalAttrs := core.Attrs{
        Class: modalClass,
        Extra: map[string]string{
            "role":       "dialog",
            "aria-modal": "true",
        },
    }
    if props.Title != "" {
        modalAttrs.Extra["aria-labelledby"] = "modal-title"
    }

    // Build content
    var content []core.Node

    // Header with title and close button
    if props.Title != "" || props.OnClose != nil {
        headerChildren := []core.Node{}

        if props.Title != "" {
            headerChildren = append(headerChildren,
                core.H3(core.Attrs{
                    ID:    "modal-title",
                    Class: "text-lg font-semibold text-gray-900 dark:text-gray-100",
                }, core.Text(props.Title)),
            )
        } else {
            // Spacer when no title
            headerChildren = append(headerChildren, core.Div(core.Attrs{}))
        }

        if props.OnClose != nil {
            headerChildren = append(headerChildren,
                core.Button(core.Attrs{
                    Class:   "text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-2xl leading-none cursor-pointer",
                    OnClick: props.OnClose,
                    Extra:   map[string]string{"aria-label": "Close"},
                }, core.Text("x")),
            )
        }

        content = append(content,
            core.Div(core.Class("flex justify-between items-center px-6 py-4 border-b border-gray-200 dark:border-gray-700"),
                headerChildren...),
        )
    }

    // Children
    content = append(content, props.Children...)

    return core.Div(core.Attrs{
        Class:   overlayClass,
        OnClick: props.OnClose, // Clicking backdrop closes
    },
        core.Div(modalAttrs, content...),
    )
}

// ModalContentProps configures modal body content.
type ModalContentProps struct {
    Class    string
    Children []core.Node
}

// ModalContent wraps modal body content with appropriate padding.
func ModalContent(props ModalContentProps) core.Node {
    class := MergeClasses("px-6 py-4 overflow-y-auto flex-1", props.Class)
    return core.Div(core.Class(class), props.Children...)
}

// ModalFooterProps configures modal footer (typically action buttons).
type ModalFooterProps struct {
    Class    string
    Children []core.Node
}

// ModalFooter wraps modal actions with appropriate styling.
func ModalFooter(props ModalFooterProps) core.Node {
    class := MergeClasses("px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-2", props.Class)
    return core.Div(core.Class(class), props.Children...)
}
```

### Complete Tabs Implementation

```go
package ui

import "github.com/dougbarrett/gux/core"

// TabsProps configures the Tabs container.
type TabsProps struct {
    Class    string
    Children []core.Node // TabList and TabPanels
}

// Tabs creates a tabs container.
func Tabs(props TabsProps) core.Node {
    class := MergeClasses("w-full", props.Class)
    return core.Div(core.Class(class), props.Children...)
}

// TabListProps configures the tab button container.
type TabListProps struct {
    Class    string
    Children []core.Node // Tab elements
}

// TabList creates a container for tab buttons.
func TabList(props TabListProps) core.Node {
    class := MergeClasses(
        "flex border-b border-gray-200 dark:border-gray-700 overflow-x-auto",
        props.Class,
    )
    return core.Div(core.Attrs{
        Class: class,
        Extra: map[string]string{
            "role":       "tablist",
            "aria-label": "Tabs",
        },
    }, props.Children...)
}

// TabProps configures a single tab button.
type TabProps struct {
    Active   bool        // Whether this tab is active
    OnSelect func()      // Called when tab is clicked
    Disabled bool        // Disabled state
    Class    string
    Children []core.Node // Tab label content
}

// Tab creates a tab button.
func Tab(props TabProps) core.Node {
    class := MergeClasses(
        "px-4 py-2 font-medium text-sm cursor-pointer border-b-2 transition-colors whitespace-nowrap",
        "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset",
        ConditionalClass(props.Active, "border-blue-500 text-blue-600 dark:text-blue-400"),
        ConditionalClass(!props.Active && !props.Disabled, "border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300"),
        ConditionalClass(props.Disabled, "border-transparent text-gray-300 dark:text-gray-600 cursor-not-allowed"),
        props.Class,
    )

    attrs := core.Attrs{
        Class:   class,
        OnClick: props.OnSelect,
        Extra: map[string]string{
            "role":          "tab",
            "aria-selected": boolToAttr(props.Active),
        },
    }

    // Active tab gets tabindex 0, others get -1 (roving tabindex)
    if props.Active {
        attrs.Extra["tabindex"] = "0"
    } else {
        attrs.Extra["tabindex"] = "-1"
    }

    if props.Disabled {
        attrs.Extra["disabled"] = "disabled"
        attrs.Extra["aria-disabled"] = "true"
    }

    return core.Button(attrs, props.Children...)
}

// TabPanelProps configures a tab panel.
type TabPanelProps struct {
    Active   bool        // Whether this panel is visible
    Class    string
    Children []core.Node // Panel content
}

// TabPanel creates a tab content panel.
func TabPanel(props TabPanelProps) core.Node {
    class := MergeClasses(
        "mt-4",
        ConditionalClass(!props.Active, "hidden"),
        props.Class,
    )

    return core.Div(core.Attrs{
        Class: class,
        Extra: map[string]string{
            "role":     "tabpanel",
            "tabindex": "0",
        },
    }, props.Children...)
}
```

### Usage Example (Page Level)

```go
func MyPage(r *core.Router) func() core.Node {
    return func() core.Node {
        // State for modal
        modalOpen := r.StateBool("modalOpen", false)

        // State for tabs
        activeTab := r.StateInt("activeTab", 0)

        // State for toasts (simplified - real impl might use slice)
        showToast := r.StateBool("showToast", false)

        return ui.VStack(ui.StackProps{
            Children: []core.Node{
                // Trigger button
                ui.Button(ui.ButtonProps{
                    OnClick: func() { modalOpen.Set(true) },
                    Children: []core.Node{core.Text("Open Modal")},
                }),

                // Modal (always rendered, visibility controlled by prop)
                ui.Modal(ui.ModalProps{
                    Open:    modalOpen.Get(),
                    OnClose: func() { modalOpen.Set(false) },
                    Title:   "My Modal",
                    Children: []core.Node{
                        ui.ModalContent(ui.ModalContentProps{
                            Children: []core.Node{core.Text("Modal content here")},
                        }),
                        ui.ModalFooter(ui.ModalFooterProps{
                            Children: []core.Node{
                                ui.Button(ui.ButtonProps{
                                    OnClick: func() {
                                        modalOpen.Set(false)
                                        showToast.Set(true)
                                    },
                                    Children: []core.Node{core.Text("Save")},
                                }),
                            },
                        }),
                    },
                }),

                // Tabs
                ui.Tabs(ui.TabsProps{
                    Children: []core.Node{
                        ui.TabList(ui.TabListProps{
                            Children: []core.Node{
                                ui.Tab(ui.TabProps{
                                    Active:   activeTab.Get() == 0,
                                    OnSelect: func() { activeTab.Set(0) },
                                    Children: []core.Node{core.Text("Tab 1")},
                                }),
                                ui.Tab(ui.TabProps{
                                    Active:   activeTab.Get() == 1,
                                    OnSelect: func() { activeTab.Set(1) },
                                    Children: []core.Node{core.Text("Tab 2")},
                                }),
                            },
                        }),
                        ui.TabPanel(ui.TabPanelProps{
                            Active:   activeTab.Get() == 0,
                            Children: []core.Node{core.Text("Content 1")},
                        }),
                        ui.TabPanel(ui.TabPanelProps{
                            Active:   activeTab.Get() == 1,
                            Children: []core.Node{core.Text("Content 2")},
                        }),
                    },
                }),

                // Toast container (always rendered, toasts added conditionally)
                ui.ToastContainer(ui.ToastContainerProps{}),

                // Conditional toast
                core.If(showToast.Get(),
                    ui.Toast(ui.ToastProps{
                        Message: "Saved successfully!",
                        Variant: ui.ToastSuccess,
                        OnClose: func() { showToast.Set(false) },
                    }),
                ),
            },
        })
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| WASM-only components (js.Value) | core.Node universal | Phase 16-18 | Components work in SSR + WASM |
| Component-internal state | Page-level state via Router | v2.0 | SSR hydration works correctly |
| Portal rendering | Fixed positioning | core.Node design | No portal API needed |
| js.FuncOf for events | core.Attrs.OnClick | core v2.0 | Abstracted event handling |

**Deprecated/outdated:**
- `components/modal.go` (WASM-only): Being replaced by `ui/modal.go` using core.Node
- `components/dropdown.go` (WASM-only): Being replaced by `ui/dropdown.go` using core.Node
- `components/tabs.go` (WASM-only): Being replaced by `ui/tabs.go` using core.Node
- `components/toast.go` (WASM-only): Being replaced by `ui/toast.go` using core.Node
- `components/tooltip.go` (WASM-only): Being replaced by `ui/tooltip.go` using core.Node
- `components/focustrap.go` (WASM-only): Focus trapping remains WASM-only, cannot be ported to core.Node

## Open Questions

1. **Focus Trapping for Modals**
   - What we know: `components/focustrap.go` exists but uses js.Value directly
   - What's unclear: Whether to include focus trap logic in Modal or leave as WASM-only enhancement
   - Recommendation: Document that full keyboard accessibility for modals requires WASM. Consider a WASM-only enhancement hook.

2. **Toast Auto-Dismiss**
   - What we know: Current WASM toast uses `time.Sleep` in goroutine
   - What's unclear: How to implement auto-dismiss with page-level state
   - Recommendation: Use SetTimeout callback pattern - page stores toast ID, removes after delay

3. **Dropdown Position Detection**
   - What we know: Need to flip dropdown if near viewport edge
   - What's unclear: Viewport detection requires WASM/JS
   - Recommendation: Provide position prop (top/bottom), let page handle dynamic positioning if needed

4. **Keyboard Navigation for Dropdown**
   - What we know: WAI-ARIA requires arrow key navigation
   - What's unclear: How to implement roving tabindex with page-level state
   - Recommendation: Start with basic open/close, document keyboard navigation as WASM enhancement

## Sources

### Primary (HIGH confidence)
- [WAI-ARIA Dialog Pattern](https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/) - Modal accessibility requirements
- [WAI-ARIA Tabs Pattern](https://www.w3.org/WAI/ARIA/apg/patterns/tabs/) - Tabs accessibility requirements
- [WAI-ARIA Menu Pattern](https://www.w3.org/WAI/ARIA/apg/patterns/menubar/) - Dropdown accessibility requirements
- [WAI-ARIA Tooltip Pattern](https://www.w3.org/WAI/ARIA/apg/patterns/tooltip/) - Tooltip accessibility requirements
- [MDN ARIA Live Regions](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Guides/Live_regions) - Toast accessibility
- `/Users/dougbarrett/projects/dbb1dev/goquery/ui/` - Established Phase 16-18 patterns
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/` - core.Node architecture

### Secondary (MEDIUM confidence)
- `/Users/dougbarrett/projects/dbb1dev/goquery/components/modal.go` - WASM reference for behavior
- `/Users/dougbarrett/projects/dbb1dev/goquery/components/dropdown.go` - WASM reference for behavior
- `/Users/dougbarrett/projects/dbb1dev/goquery/components/tabs.go` - WASM reference for behavior
- `/Users/dougbarrett/projects/dbb1dev/goquery/components/toast.go` - WASM reference for behavior
- `/Users/dougbarrett/projects/dbb1dev/goquery/components/tooltip.go` - WASM reference for behavior

### Tertiary (LOW confidence)
- Sara Soueidan's ARIA Live Regions articles - Toast implementation best practices (not verified with official spec)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Uses existing core and ui packages
- Architecture: HIGH - Follows established patterns from Phases 16-18
- Accessibility: HIGH - Based on official WAI-ARIA patterns
- Pitfalls: HIGH - Based on codebase analysis and SSR/WASM constraints

**Research date:** 2026-01-22
**Valid until:** 2026-03-22 (60 days - stable domain)

## Component Implementation Order

Recommended build order based on dependencies:

1. **Modal** - No dependencies on other Phase 19 components, foundational overlay pattern
2. **Tooltip** - Simplest interactive component, CSS-driven visibility
3. **Tabs** - No dependencies, demonstrates index-based selection pattern
4. **Dropdown** - Similar to Modal but with menu semantics
5. **Toast** - Requires understanding of ARIA live regions

Each component: implementation file + test file, following Phase 16-18 patterns.
