# Architecture Research: Component Library for Gux

**Project:** Component library building on core framework
**Researched:** 2026-01-22
**Confidence:** HIGH (based on existing codebase analysis + established patterns)

## Executive Summary

This research analyzes how to structure a component library that builds on Gux's existing `core.Node` system. The core framework already provides the foundational patterns: components are functions returning `core.Node`, rendered to HTML on server or DOM via WASM. The component library should follow these established patterns while introducing higher-level abstractions for Cards, Modals, Tables, Forms, etc.

**Key insight:** The existing codebase already demonstrates component patterns (Nav, tableCell, formField). The component library is about standardizing these patterns, not inventing new ones.

---

## Component Structure Patterns

### Pattern 1: Simple Component Function (Stateless)

For presentational components without internal state:

```go
// components/card.go
package components

import "github.com/dougbarrett/gux/core"

// CardProps configures a Card component
type CardProps struct {
    Title    string
    Subtitle string
    Class    string
}

// Card renders a card container
func Card(props CardProps, children ...core.Node) core.Node {
    return core.Div(core.Attrs{Class: mergeClasses("bg-white rounded-lg shadow-lg p-6", props.Class)},
        core.If(props.Title != "",
            core.H2(core.Class("text-xl font-bold mb-2"), core.Text(props.Title)),
        ),
        core.If(props.Subtitle != "",
            core.P(core.Class("text-gray-600 mb-4"), core.Text(props.Subtitle)),
        ),
        core.Frag(children...),
    )
}
```

**Rationale:** Mirrors existing patterns like `tableHeader()` and `tableCell()` in the codebase. Props struct makes configuration explicit. Children passed as variadic argument maintains composition.

### Pattern 2: Component with Variants (Builder Pattern)

For components with multiple variants, use Go's functional options pattern:

```go
// components/button.go
package components

import "github.com/dougbarrett/gux/core"

type ButtonVariant string
const (
    ButtonPrimary   ButtonVariant = "primary"
    ButtonSecondary ButtonVariant = "secondary"
    ButtonDanger    ButtonVariant = "danger"
)

type ButtonSize string
const (
    ButtonSm ButtonSize = "sm"
    ButtonMd ButtonSize = "md"
    ButtonLg ButtonSize = "lg"
)

type ButtonProps struct {
    Variant  ButtonVariant
    Size     ButtonSize
    Disabled bool
    OnClick  func()
    Class    string
}

func Button(props ButtonProps, children ...core.Node) core.Node {
    baseClass := "font-medium rounded-lg transition"

    // Variant classes
    variantClass := map[ButtonVariant]string{
        ButtonPrimary:   "bg-blue-500 text-white hover:bg-blue-600",
        ButtonSecondary: "bg-gray-200 text-gray-800 hover:bg-gray-300",
        ButtonDanger:    "bg-red-500 text-white hover:bg-red-600",
    }[props.Variant]
    if variantClass == "" {
        variantClass = variantClass[ButtonPrimary]
    }

    // Size classes
    sizeClass := map[ButtonSize]string{
        ButtonSm: "px-3 py-1.5 text-sm",
        ButtonMd: "px-4 py-2",
        ButtonLg: "px-6 py-3 text-lg",
    }[props.Size]
    if sizeClass == "" {
        sizeClass = sizeClass[ButtonMd]
    }

    return core.Button(
        core.Attrs{
            Class:   mergeClasses(baseClass, variantClass, sizeClass, props.Class),
            OnClick: props.OnClick,
            Extra:   conditionalAttr(props.Disabled, "disabled", "disabled"),
        },
        children...,
    )
}
```

**Rationale:** Type-safe variants via string constants. Props struct keeps all configuration together. Defaults handled in function body.

### Pattern 3: Stateful Component (Requires Router)

For components with internal state, follow the existing page pattern:

```go
// components/modal.go
package components

import "github.com/dougbarrett/gux/core"

type ModalProps struct {
    ID       string        // Unique ID for state key
    Title    string
    OnClose  func()
}

// Modal creates a modal that manages its own open/close state
// Usage: Modal(r, ModalProps{ID: "confirm-modal", Title: "Confirm"}, content...)
func Modal(r *core.Router, props ModalProps, children ...core.Node) core.Node {
    isOpen := r.StateBool(props.ID+"_open", false)

    if !isOpen.Get() {
        return core.Frag() // Return nothing when closed
    }

    return core.Div(core.Class("fixed inset-0 z-50 flex items-center justify-center"),
        // Backdrop
        core.Div(core.Attrs{
            Class:   "absolute inset-0 bg-black bg-opacity-50",
            OnClick: func() {
                isOpen.Set(false)
                if props.OnClose != nil {
                    props.OnClose()
                }
            },
        }),
        // Modal content
        core.Div(core.Class("relative bg-white rounded-lg shadow-xl max-w-md w-full mx-4 p-6"),
            core.Div(core.Class("flex justify-between items-center mb-4"),
                core.H3(core.Class("text-lg font-semibold"), core.Text(props.Title)),
                core.Button(core.Attrs{
                    Class:   "text-gray-400 hover:text-gray-600",
                    OnClick: func() {
                        isOpen.Set(false)
                        if props.OnClose != nil {
                            props.OnClose()
                        }
                    },
                }, core.Text("x")),
            ),
            core.Frag(children...),
        ),
    )
}

// ModalTrigger returns a function to open the modal
func ModalTrigger(r *core.Router, modalID string) func() {
    return func() {
        isOpen := r.StateBool(modalID+"_open", false)
        isOpen.Set(true)
    }
}
```

**Rationale:** Follows existing pattern where stateful pages take `*core.Router`. State keyed by unique ID allows multiple modals per page.

---

## Props and Configuration

### Recommended: Explicit Props Struct

```go
type InputProps struct {
    Type        string           // text, email, password, etc.
    Name        string
    Value       string
    Placeholder string
    Label       string
    Error       string
    Required    bool
    Disabled    bool
    OnChange    func(string)
    OnEnter     func()
    Class       string
}

func Input(props InputProps) core.Node {
    // Implementation
}
```

**Why props structs over functional options:**
1. **Explicit:** All options visible at call site via struct literal
2. **IDE support:** Autocomplete works with struct fields
3. **Simple:** No function indirection
4. **Matches existing code:** Mirrors `core.Attrs` pattern

### Alternative: Functional Options (For Library Extension Points)

```go
type TableOption func(*TableConfig)

func WithPagination(pageSize int) TableOption {
    return func(c *TableConfig) {
        c.PageSize = pageSize
        c.Paginated = true
    }
}

func Table(data []Row, opts ...TableOption) core.Node {
    config := &TableConfig{PageSize: 10}
    for _, opt := range opts {
        opt(config)
    }
    // ...
}
```

**When to use:** Only for complex components where users might want to extend behavior (Tables, Forms with validation).

### Props Merging Pattern

```go
// Helper to merge Tailwind classes
func mergeClasses(classes ...string) string {
    var result []string
    for _, c := range classes {
        if c != "" {
            result = append(result, c)
        }
    }
    return strings.Join(result, " ")
}

// Usage: allows overriding default styles
func Card(props CardProps, children ...core.Node) core.Node {
    return core.Div(core.Attrs{
        Class: mergeClasses("bg-white rounded-lg shadow p-4", props.Class),
    }, children...)
}
```

---

## Composition Patterns

### Pattern 1: Children (Default Composition)

Most components accept children as variadic `...core.Node`:

```go
func Card(props CardProps, children ...core.Node) core.Node {
    return core.Div(core.Class("card"),
        // Card header
        core.Div(core.Class("card-header"), core.Text(props.Title)),
        // User-provided content
        core.Div(core.Class("card-body"), children...),
    )
}

// Usage
Card(CardProps{Title: "Settings"},
    core.P(core.Attrs{}, core.Text("Content here")),
    core.Button(core.Attrs{OnClick: save}, core.Text("Save")),
)
```

### Pattern 2: Named Slots via Props

For components needing multiple insertion points:

```go
type CardProps struct {
    Title   string
    Header  core.Node  // Optional header slot
    Footer  core.Node  // Optional footer slot
    Class   string
}

func Card(props CardProps, children ...core.Node) core.Node {
    return core.Div(core.Attrs{Class: mergeClasses("card", props.Class)},
        core.If(props.Header != nil, props.Header),
        core.Div(core.Class("card-body"), children...),
        core.If(props.Footer != nil, props.Footer),
    )
}

// Usage
Card(CardProps{
    Title: "User Profile",
    Footer: core.Div(core.Class("flex justify-end gap-2"),
        Button(ButtonProps{Variant: ButtonSecondary}, core.Text("Cancel")),
        Button(ButtonProps{Variant: ButtonPrimary}, core.Text("Save")),
    ),
},
    core.Text("Card content..."),
)
```

### Pattern 3: Compound Components

For tightly related component groups (like Table with Head/Body/Row):

```go
// table.go - Compound components share a namespace

// Table is the container
func Table(props TableProps, children ...core.Node) core.Node {
    return core.Table(core.Attrs{Class: mergeClasses("w-full", props.Class)},
        children...,
    )
}

// TableHead renders the header section
func TableHead(children ...core.Node) core.Node {
    return core.Thead(core.Class("bg-gray-100"),
        core.Tr(core.Attrs{}, children...),
    )
}

// TableHeadCell renders a header cell
func TableHeadCell(text string) core.Node {
    return core.Th(core.Class("px-4 py-2 text-left text-sm font-medium text-gray-700"),
        core.Text(text),
    )
}

// TableBody renders the body section
func TableBody(children ...core.Node) core.Node {
    return core.Tbody(core.Attrs{}, children...)
}

// TableRow renders a data row
func TableRow(props TableRowProps, children ...core.Node) core.Node {
    return core.Tr(core.Attrs{
        Class:   mergeClasses("border-t hover:bg-gray-50", props.Class),
        OnClick: props.OnClick,
    }, children...)
}

// TableCell renders a data cell
func TableCell(children ...core.Node) core.Node {
    return core.Td(core.Class("px-4 py-2 text-sm"), children...)
}

// Usage - reads like declarative markup
Table(TableProps{},
    TableHead(
        TableHeadCell("Name"),
        TableHeadCell("Email"),
        TableHeadCell("Role"),
    ),
    TableBody(
        TableRow(TableRowProps{},
            TableCell(core.Text("John")),
            TableCell(core.Text("john@example.com")),
            TableCell(core.Text("Admin")),
        ),
    ),
)
```

### Pattern 4: Render Props (For Dynamic Children)

For data-driven rendering:

```go
type DataTableProps[T any] struct {
    Data       []T
    Columns    []ColumnDef[T]
    RowKey     func(T) string
    OnRowClick func(T)
}

type ColumnDef[T any] struct {
    Header string
    Render func(T) core.Node
}

func DataTable[T any](props DataTableProps[T]) core.Node {
    var rows []core.Node
    for _, item := range props.Data {
        var cells []core.Node
        for _, col := range props.Columns {
            cells = append(cells, TableCell(col.Render(item)))
        }
        rows = append(rows, TableRow(TableRowProps{
            OnClick: func() {
                if props.OnRowClick != nil {
                    props.OnRowClick(item)
                }
            },
        }, cells...))
    }

    var headers []core.Node
    for _, col := range props.Columns {
        headers = append(headers, TableHeadCell(col.Header))
    }

    return Table(TableProps{},
        TableHead(headers...),
        TableBody(rows...),
    )
}

// Usage
DataTable(DataTableProps[User]{
    Data: users,
    Columns: []ColumnDef[User]{
        {Header: "Name", Render: func(u User) core.Node { return core.Text(u.Name) }},
        {Header: "Email", Render: func(u User) core.Node { return core.Text(u.Email) }},
    },
    OnRowClick: func(u User) { r.Navigate("/users/" + u.ID) },
})
```

---

## State Patterns for Interactive Components

### Pattern 1: External State (Preferred)

Component receives state, doesn't manage it:

```go
type DropdownProps struct {
    IsOpen    bool
    OnToggle  func()
    Trigger   core.Node
}

func Dropdown(props DropdownProps, children ...core.Node) core.Node {
    return core.Div(core.Class("relative"),
        core.Div(core.Attrs{OnClick: props.OnToggle}, props.Trigger),
        core.If(props.IsOpen,
            core.Div(core.Class("absolute mt-2 bg-white shadow-lg rounded"),
                children...,
            ),
        ),
    )
}

// Usage - state managed by parent
func MyPage(r *core.Router) func() core.Node {
    return func() core.Node {
        menuOpen := r.StateBool("menu_open", false)

        return Dropdown(DropdownProps{
            IsOpen:   menuOpen.Get(),
            OnToggle: func() { menuOpen.Set(!menuOpen.Get()) },
            Trigger:  Button(ButtonProps{}, core.Text("Menu")),
        },
            DropdownItem("Settings", func() { /* ... */ }),
            DropdownItem("Logout", func() { /* ... */ }),
        )
    }
}
```

**Why external state:**
- Predictable behavior
- Parent controls open/close logic
- Easy to coordinate multiple components
- Follows React "lifting state up" principle

### Pattern 2: Internal State (For Self-Contained Components)

When component truly owns its state:

```go
type AccordionProps struct {
    ID    string  // Required for state key
    Items []AccordionItem
}

type AccordionItem struct {
    Title   string
    Content core.Node
}

func Accordion(r *core.Router, props AccordionProps) core.Node {
    activeIndex := r.StateInt(props.ID+"_active", -1)

    var items []core.Node
    for i, item := range props.Items {
        idx := i // Capture for closure
        isActive := activeIndex.Get() == idx

        items = append(items, core.Div(core.Class("border-b"),
            core.Button(core.Attrs{
                Class: "w-full text-left p-4 hover:bg-gray-50",
                OnClick: func() {
                    if activeIndex.Get() == idx {
                        activeIndex.Set(-1)
                    } else {
                        activeIndex.Set(idx)
                    }
                },
            }, core.Text(item.Title)),
            core.If(isActive,
                core.Div(core.Class("p-4"), item.Content),
            ),
        ))
    }

    return core.Div(core.Class("border rounded"), items...)
}
```

### Pattern 3: Controlled vs Uncontrolled

Support both patterns:

```go
type TabsProps struct {
    ID           string          // For uncontrolled state
    DefaultIndex int             // Initial index (uncontrolled)
    ActiveIndex  *int            // External control (controlled)
    OnChange     func(int)       // Callback when tab changes
    Items        []TabItem
}

func Tabs(r *core.Router, props TabsProps) core.Node {
    // Determine active index
    var activeIdx int
    var setActive func(int)

    if props.ActiveIndex != nil {
        // Controlled: use provided value
        activeIdx = *props.ActiveIndex
        setActive = func(i int) {
            if props.OnChange != nil {
                props.OnChange(i)
            }
        }
    } else {
        // Uncontrolled: use internal state
        state := r.StateInt(props.ID+"_tab", props.DefaultIndex)
        activeIdx = state.Get()
        setActive = func(i int) {
            state.Set(i)
            if props.OnChange != nil {
                props.OnChange(i)
            }
        }
    }

    // Render tabs...
}
```

---

## Directory Structure

### Recommended Layout

```
components/
  README.md                 # Usage documentation

  # Core utilities
  utils.go                  # mergeClasses, conditionalAttr, etc.

  # Layout components (no state)
  layout/
    card.go                 # Card, CardHeader, CardBody, CardFooter
    container.go            # Container, Section
    stack.go                # VStack, HStack (flex layouts)
    grid.go                 # Grid, GridItem
    divider.go              # Divider, Spacer

  # Form components
  form/
    input.go                # Input (text, email, password, etc.)
    textarea.go             # Textarea
    select.go               # Select, Option
    checkbox.go             # Checkbox, CheckboxGroup
    radio.go                # Radio, RadioGroup
    form.go                 # Form, FormField, FormError
    label.go                # Label

  # Data display
  data/
    table.go                # Table, TableHead, TableBody, TableRow, TableCell
    datatable.go            # DataTable (generic, with pagination)
    badge.go                # Badge (status indicators)
    avatar.go               # Avatar
    list.go                 # List, ListItem

  # Feedback
  feedback/
    alert.go                # Alert (success, warning, error, info)
    toast.go                # Toast (requires state)
    spinner.go              # Spinner, Loading
    progress.go             # ProgressBar
    skeleton.go             # Skeleton loaders

  # Navigation
  nav/
    breadcrumb.go           # Breadcrumb, BreadcrumbItem
    pagination.go           # Pagination
    tabs.go                 # Tabs, TabList, TabPanel (requires state)
    menu.go                 # Menu, MenuItem

  # Overlay
  overlay/
    modal.go                # Modal, ModalHeader, ModalBody, ModalFooter
    drawer.go               # Drawer (slide-in panel)
    dropdown.go             # Dropdown, DropdownItem
    tooltip.go              # Tooltip
    popover.go              # Popover

  # Action
  action/
    button.go               # Button with variants
    buttongroup.go          # ButtonGroup
    link.go                 # Link (styled anchor)
    iconbutton.go           # IconButton
```

### Package Structure Option A: Single Package

```go
import "github.com/dougbarrett/gux/components"

components.Card(components.CardProps{Title: "Hello"},
    components.Button(components.ButtonProps{Variant: components.ButtonPrimary},
        core.Text("Click"),
    ),
)
```

**Pros:** Simple imports
**Cons:** Long prefixes, potential name collisions

### Package Structure Option B: Sub-packages

```go
import (
    "github.com/dougbarrett/gux/components/layout"
    "github.com/dougbarrett/gux/components/action"
)

layout.Card(layout.CardProps{Title: "Hello"},
    action.Button(action.ButtonProps{Variant: action.Primary},
        core.Text("Click"),
    ),
)
```

**Pros:** Clear organization, shorter names within package
**Cons:** More imports

### Recommended: Single Package with Type Prefixes

```go
import "github.com/dougbarrett/gux/ui"

ui.Card(ui.CardProps{Title: "Hello"},
    ui.Button(ui.ButtonProps{Variant: ui.ButtonPrimary},
        core.Text("Click"),
    ),
)
```

**Why:** Short package name (`ui`), clear prefixes on types prevent collisions, single import for all components.

---

## Build Order

### Phase 1: Foundation (No Dependencies)

Build utilities and primitives first:

1. **utils.go** - `mergeClasses()`, `conditionalAttr()`
2. **layout/stack.go** - `VStack`, `HStack` (flex containers)
3. **layout/container.go** - `Container`, `Section`
4. **action/button.go** - `Button` with variants

**Rationale:** These have no dependencies on other components. Button is needed by almost everything else.

### Phase 2: Static Display Components

Components that display data without internal state:

5. **layout/card.go** - `Card`, `CardHeader`, `CardBody`
6. **data/badge.go** - `Badge`
7. **data/avatar.go** - `Avatar`
8. **feedback/alert.go** - `Alert`
9. **feedback/spinner.go** - `Spinner`
10. **layout/divider.go** - `Divider`, `Spacer`

**Rationale:** Pure display, easy to test, no router needed.

### Phase 3: Form Components

Form inputs follow the existing `formField` pattern:

11. **form/label.go** - `Label`
12. **form/input.go** - `Input` (mirrors existing pattern)
13. **form/textarea.go** - `Textarea`
14. **form/select.go** - `Select`, `Option`
15. **form/checkbox.go** - `Checkbox`
16. **form/radio.go** - `Radio`, `RadioGroup`
17. **form/form.go** - `Form`, `FormField` (wraps label + input + error)

**Rationale:** Forms are critical for admin UIs. Build on existing `formField` helper pattern.

### Phase 4: Data Display

Tables and lists build on Phase 1-2:

18. **data/table.go** - Compound table components
19. **data/list.go** - `List`, `ListItem`
20. **nav/breadcrumb.go** - `Breadcrumb`
21. **nav/pagination.go** - `Pagination`

**Rationale:** Admin apps need tables. Pagination pairs naturally with tables.

### Phase 5: Stateful Components

Components requiring `*core.Router`:

22. **nav/tabs.go** - `Tabs` (needs state)
23. **overlay/dropdown.go** - `Dropdown` (needs state)
24. **overlay/modal.go** - `Modal` (needs state)
25. **feedback/toast.go** - `Toast` (needs state)

**Rationale:** More complex, requires understanding state patterns. Modal is commonly needed for confirmations.

### Phase 6: Advanced Components

Build on everything:

26. **data/datatable.go** - Generic data table with sorting/pagination
27. **overlay/drawer.go** - Slide-in panel
28. **overlay/popover.go** - Positioned overlay
29. **layout/grid.go** - Grid layout

**Rationale:** These are "nice to have" and build on simpler components.

---

## Integration Points with Core

### What Exists (Use As-Is)

| Core Element | Purpose | Usage in Components |
|--------------|---------|---------------------|
| `core.Node` | Return type for all components | Every component returns `core.Node` |
| `core.Div`, `core.Button`, etc. | HTML element constructors | Base building blocks |
| `core.Attrs` | Attribute configuration | Passed to elements |
| `core.Frag()` | Fragment for multiple nodes | Grouping children |
| `core.If()` | Conditional rendering | Show/hide logic |
| `core.Text()` | Text content | Labels, content |
| `*core.Router` | State management | Stateful components |
| `r.StateBool/Int/String()` | State hooks | Internal component state |

### What Needs Adding to Core (None Required)

The component library can be built entirely on top of existing core primitives. No core changes needed.

### Styling Integration

Components use Tailwind classes directly. The existing Tailwind build process will include component styles automatically as long as the components package is included in Tailwind's content configuration:

```js
// tailwind.config.js
module.exports = {
  content: [
    './pages/**/*.go',
    './components/**/*.go', // Add this
  ],
  // ...
}
```

---

## SSR + WASM Considerations

### What Works Automatically

1. **Static components** - Render identically on server and client
2. **Event handlers** - Ignored in SSR, active in WASM (handled by core)
3. **State** - Hydrated from server via `__gux_state` script (handled by core)

### Patterns That Need Care

**Portal-like behavior (modals at body root):**
The current architecture renders everything inside `#app`. For true portals (modals rendered at document body), this would require core changes. Workaround: Use fixed positioning with high z-index.

**Focus management:**
WASM can call DOM methods. For modals that need to trap focus:
```go
// In dom_renderer.go or component
// el.Call("focus") is possible via js interop
```

**Scroll locking:**
For modals that lock scroll:
```go
// Need to add/remove class on body - requires core support
// Workaround: Use overflow-hidden on app div
```

### Recommendation: Start Simple

Build components without portal/focus management first. These can be added later if needed. The existing modal pattern in admin examples works without them.

---

## Sources

- Existing codebase analysis (PRIMARY)
- [Go Functional Options Pattern](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis) - Dave Cheney
- [Vugu Component Structure](https://www.vugu.org/doc/components) - Component patterns in Go WASM
- [Compound Components Pattern](https://www.smashingmagazine.com/2021/08/compound-components-react/) - React compound components (adapted for Go)
- [Headless UI Patterns](https://martinfowler.com/articles/headless-component.html) - Martin Fowler on headless components
- [Tailwind Component Library Architecture](https://hexshift.medium.com/how-to-design-and-maintain-a-scalable-tailwind-component-library-be41e87cbf1a) - Design system structure
