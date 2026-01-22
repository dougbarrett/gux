# Phase 16: Component Foundation - Research

**Researched:** 2026-01-22
**Domain:** Go/WASM Component Library with Tailwind CSS
**Confidence:** HIGH

## Summary

This research establishes the patterns for building a new component library on top of Gux's `core.Node` system. The codebase already has two distinct component systems: the **old `components/` package** (WASM-only, uses `js.Value` directly) and the **new `core/` package** (universal, uses `Node` interface for both SSR and WASM). Phase 16 will build the new component library using `core.Node` patterns.

The key insight is that components in the new system are **pure functions returning `core.Node`**, not imperative DOM manipulations. This enables SSR compatibility and eliminates hydration mismatches. The existing codebase demonstrates the pattern in `examples/minimal/` - components like `formField()` and `activityItem()` show how to create reusable UI elements as functions.

**Primary recommendation:** Create a new `ui/` package with components built on `core.Node`, using explicit Props structs and Tailwind CSS utility classes. Start with utilities (class merging), then Button, Card, and layout primitives (VStack/HStack/Container).

## Standard Stack

The established libraries/tools for this domain:

### Core (Already in codebase)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| core.Node | - | Abstract UI tree | Enables SSR + WASM rendering |
| core.Element | - | HTML element wrapper | Composes children and attributes |
| core.Attrs | - | Element attributes | Type-safe attribute handling |
| core.Class() | - | Class-only attrs shorthand | Reduces boilerplate |
| Tailwind CSS | CDN | Utility-first styling | Already in use, consistent |

### Supporting (New for Phase 16)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| strings.Builder | stdlib | Class string building | mergeClasses utility |
| strings.Join | stdlib | Class concatenation | Joining non-empty classes |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom class merging | tailwind-merge (JS) | JS interop overhead, not needed for Go |
| Variant maps | cva (class-variance-authority) | JS library, recreate pattern in Go |

**Installation:**
No new dependencies required. All components will use existing `core` package.

## Architecture Patterns

### Recommended Project Structure
```
ui/
├── utils.go           # Class merging, conditional attrs
├── button.go          # Button, ButtonGroup, IconButton
├── card.go            # Card, CardHeader, CardContent, CardFooter
├── layout.go          # Container, VStack, HStack, Grid, Divider
└── README.md          # Component API documentation
```

### Pattern 1: Simple Stateless Component
**What:** Component as function returning `core.Node`
**When to use:** Components without internal state (Card, Badge, Container)
**Example:**
```go
// Source: Based on examples/minimal pattern
type CardProps struct {
    Class    string
    Children []core.Node
}

func Card(props CardProps) core.Node {
    class := mergeClasses("bg-white dark:bg-gray-800 rounded-lg shadow p-6", props.Class)
    return core.Div(core.Class(class), props.Children...)
}
```

### Pattern 2: Variant Component with Props Struct
**What:** Component with predefined style variants via typed enums
**When to use:** Components with multiple visual states (Button variants, Badge colors)
**Example:**
```go
// Source: Based on existing components/button.go pattern, adapted for core.Node
type ButtonVariant string

const (
    ButtonPrimary     ButtonVariant = "primary"
    ButtonSecondary   ButtonVariant = "secondary"
    ButtonOutline     ButtonVariant = "outline"
    ButtonGhost       ButtonVariant = "ghost"
    ButtonDestructive ButtonVariant = "destructive"
)

type ButtonSize string

const (
    ButtonSM ButtonSize = "sm"
    ButtonMD ButtonSize = "md"
    ButtonLG ButtonSize = "lg"
)

type ButtonProps struct {
    Variant  ButtonVariant
    Size     ButtonSize
    Class    string
    Disabled bool
    OnClick  func()
    Children []core.Node
}

func Button(props ButtonProps) core.Node {
    variant := props.Variant
    if variant == "" {
        variant = ButtonPrimary
    }
    size := props.Size
    if size == "" {
        size = ButtonMD
    }

    class := mergeClasses(
        "font-medium rounded transition-colors cursor-pointer",
        buttonVariantClasses[variant],
        buttonSizeClasses[size],
        props.Class,
    )

    attrs := core.Attrs{
        Class:   class,
        OnClick: props.OnClick,
    }
    if props.Disabled {
        attrs.Extra = map[string]string{"disabled": "disabled"}
    }

    return core.El("button", attrs, props.Children...)
}
```

### Pattern 3: Compound Component Set
**What:** Related components that compose together
**When to use:** Card with header/content/footer, Table with rows/cells
**Example:**
```go
// Source: shadcn/ui Card pattern, adapted for core.Node
type CardHeaderProps struct {
    Class    string
    Children []core.Node
}

func CardHeader(props CardHeaderProps) core.Node {
    class := mergeClasses("pb-4", props.Class)
    return core.Div(core.Class(class), props.Children...)
}

type CardContentProps struct {
    Class    string
    Children []core.Node
}

func CardContent(props CardContentProps) core.Node {
    class := mergeClasses("py-4", props.Class)
    return core.Div(core.Class(class), props.Children...)
}

type CardFooterProps struct {
    Class    string
    Children []core.Node
}

func CardFooter(props CardFooterProps) core.Node {
    class := mergeClasses("pt-4", props.Class)
    return core.Div(core.Class(class), props.Children...)
}

// Usage:
Card(CardProps{Children: []core.Node{
    CardHeader(CardHeaderProps{Children: []core.Node{
        core.H3(core.Class("text-lg font-semibold"), core.Text("Title")),
    }}),
    CardContent(CardContentProps{Children: []core.Node{
        core.P(core.Attrs{}, core.Text("Content here")),
    }}),
    CardFooter(CardFooterProps{Children: []core.Node{
        Button(ButtonProps{Children: []core.Node{core.Text("Action")}}),
    }}),
}})
```

### Pattern 4: Layout Primitive with Gap/Alignment
**What:** Flexbox/grid wrappers with typed spacing and alignment
**When to use:** VStack, HStack, Grid components
**Example:**
```go
// Source: Common design system patterns
type StackProps struct {
    Gap      string // Tailwind gap class value: "2", "4", "6"
    Align    string // items-start, items-center, items-end
    Justify  string // justify-start, justify-center, justify-between
    Class    string
    Children []core.Node
}

func VStack(props StackProps) core.Node {
    gap := props.Gap
    if gap == "" {
        gap = "4"
    }
    class := mergeClasses(
        "flex flex-col",
        "gap-"+gap,
        props.Align,
        props.Justify,
        props.Class,
    )
    return core.Div(core.Class(class), props.Children...)
}

func HStack(props StackProps) core.Node {
    gap := props.Gap
    if gap == "" {
        gap = "4"
    }
    class := mergeClasses(
        "flex flex-row",
        "gap-"+gap,
        props.Align,
        props.Justify,
        props.Class,
    )
    return core.Div(core.Class(class), props.Children...)
}
```

### Anti-Patterns to Avoid
- **Using `interface{}` for props:** Always use explicit Props structs with typed fields
- **Direct `js.Value` in components:** Use `core.Node` interface for SSR compatibility
- **Hardcoded colors without variants:** Use variant enums for consistent theming
- **Inline class strings everywhere:** Use mergeClasses utility for maintainability
- **Returning `*Element` or mutable types:** Components should return `Node` interface

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Class string concatenation | String + " " + String | `mergeClasses()` utility | Handles empty strings, duplicates |
| Conditional attributes | If statements in each component | `conditionalClass()` helper | Reduces boilerplate |
| Button styling variants | Switch statements | Variant map lookup | Consistent pattern, easier to extend |
| Component children | Manually building slices | Variadic `...core.Node` | Go idiom, clean API |

**Key insight:** Build a small set of utilities (mergeClasses, conditional helpers) once in Phase 16, then all subsequent components benefit. This is the foundation pattern from shadcn/ui: utilities enable composable components.

## Common Pitfalls

### Pitfall 1: Empty Class Fragments
**What goes wrong:** `mergeClasses("", "text-red-500", "")` produces `" text-red-500 "` with leading/trailing spaces
**Why it happens:** Naive string concatenation doesn't filter empty strings
**How to avoid:** Filter empty strings before joining
```go
func mergeClasses(classes ...string) string {
    var result []string
    for _, c := range classes {
        if c != "" {
            result = append(result, c)
        }
    }
    return strings.Join(result, " ")
}
```
**Warning signs:** Double spaces in rendered HTML, Tailwind classes not applying

### Pitfall 2: Props Struct Zero Values
**What goes wrong:** Component renders with unexpected defaults when props are omitted
**Why it happens:** Go zero values for strings are "", for enums are first constant
**How to avoid:** Explicit default handling in component functions
```go
variant := props.Variant
if variant == "" {
    variant = ButtonPrimary // Explicit default
}
```
**Warning signs:** All buttons look the same regardless of Variant prop

### Pitfall 3: Children Ordering
**What goes wrong:** Children render in wrong order or get lost
**Why it happens:** Mixing variadic args with slice appends
**How to avoid:** Consistent pattern: Props.Children for known slots, variadic for simple cases
```go
// GOOD: Props struct for compound components
type CardProps struct {
    Header   core.Node   // Named slot
    Footer   core.Node   // Named slot
    Children []core.Node // Main content
}

// GOOD: Variadic for simple wrappers
func Container(children ...core.Node) core.Node {
    return core.Div(core.Class("max-w-7xl mx-auto px-4"), children...)
}
```
**Warning signs:** Components render empty, children appear twice

### Pitfall 4: SSR vs WASM Attribute Handling
**What goes wrong:** `disabled` attribute not working in WASM, or `onclick` events firing in SSR
**Why it happens:** Attrs.Extra works differently than built-in Attrs fields
**How to avoid:** Use Attrs.Extra for boolean HTML attributes, OnClick only fires in WASM (by design)
```go
attrs := core.Attrs{Class: class}
if props.Disabled {
    attrs.Extra = map[string]string{"disabled": "disabled"}
}
```
**Warning signs:** Disabled buttons are clickable, form submissions fail

### Pitfall 5: Tailwind Class Conflicts
**What goes wrong:** `bg-blue-500` doesn't override when `bg-red-500` is also present
**Why it happens:** CSS specificity with utility classes depends on source order
**How to avoid:** Put custom classes last in mergeClasses, they'll override
```go
class := mergeClasses(
    buttonVariantClasses[variant], // Base styles
    props.Class,                    // Custom overrides (wins)
)
```
**Warning signs:** Component styles don't change despite custom Class prop

## Code Examples

Verified patterns from codebase analysis:

### Class Merging Utility
```go
// Source: Pattern derived from shadcn/ui cn() function, adapted for Go
package ui

import "strings"

// mergeClasses combines CSS class strings, filtering empty values.
// Custom classes should be passed last to allow overrides.
func mergeClasses(classes ...string) string {
    var result []string
    for _, c := range classes {
        c = strings.TrimSpace(c)
        if c != "" {
            result = append(result, c)
        }
    }
    return strings.Join(result, " ")
}

// conditionalClass returns the class if condition is true, empty string otherwise.
func conditionalClass(condition bool, class string) string {
    if condition {
        return class
    }
    return ""
}
```

### Button with Variants (Complete)
```go
// Source: Adapted from components/button.go for core.Node
package ui

import "github.com/dougbarrett/gux/core"

type ButtonVariant string

const (
    ButtonPrimary     ButtonVariant = "primary"
    ButtonSecondary   ButtonVariant = "secondary"
    ButtonOutline     ButtonVariant = "outline"
    ButtonGhost       ButtonVariant = "ghost"
    ButtonDestructive ButtonVariant = "destructive"
)

var buttonVariantClasses = map[ButtonVariant]string{
    ButtonPrimary:     "bg-blue-600 text-white hover:bg-blue-700",
    ButtonSecondary:   "bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200 hover:bg-gray-300 dark:hover:bg-gray-600",
    ButtonOutline:     "border border-gray-300 dark:border-gray-600 bg-transparent hover:bg-gray-100 dark:hover:bg-gray-800",
    ButtonGhost:       "bg-transparent hover:bg-gray-100 dark:hover:bg-gray-800",
    ButtonDestructive: "bg-red-600 text-white hover:bg-red-700",
}

type ButtonSize string

const (
    ButtonSM ButtonSize = "sm"
    ButtonMD ButtonSize = "md"
    ButtonLG ButtonSize = "lg"
)

var buttonSizeClasses = map[ButtonSize]string{
    ButtonSM: "px-3 py-1.5 text-sm",
    ButtonMD: "px-4 py-2 text-base",
    ButtonLG: "px-6 py-3 text-lg",
}

type ButtonProps struct {
    Variant  ButtonVariant
    Size     ButtonSize
    Class    string
    Disabled bool
    Type     string // "button", "submit", "reset"
    OnClick  func()
    Children []core.Node
}

func Button(props ButtonProps) core.Node {
    variant := props.Variant
    if variant == "" {
        variant = ButtonPrimary
    }
    size := props.Size
    if size == "" {
        size = ButtonMD
    }
    buttonType := props.Type
    if buttonType == "" {
        buttonType = "button"
    }

    class := mergeClasses(
        "font-medium rounded-lg transition-colors cursor-pointer inline-flex items-center justify-center",
        buttonVariantClasses[variant],
        buttonSizeClasses[size],
        conditionalClass(props.Disabled, "opacity-50 cursor-not-allowed"),
        props.Class,
    )

    attrs := core.Attrs{
        Class:   class,
        Type:    buttonType,
        OnClick: props.OnClick,
    }
    if props.Disabled {
        attrs.Extra = map[string]string{"disabled": "disabled"}
    }

    return core.El("button", attrs, props.Children...)
}
```

### Card Compound Components
```go
// Source: Pattern from existing examples/minimal/admin/dashboard.go
package ui

import "github.com/dougbarrett/gux/core"

type CardProps struct {
    Class    string
    Children []core.Node
}

func Card(props CardProps) core.Node {
    class := mergeClasses(
        "bg-white dark:bg-gray-800 rounded-lg shadow dark:shadow-gray-900",
        props.Class,
    )
    return core.Div(core.Class(class), props.Children...)
}

type CardHeaderProps struct {
    Class    string
    Children []core.Node
}

func CardHeader(props CardHeaderProps) core.Node {
    class := mergeClasses("px-6 py-4 border-b border-gray-200 dark:border-gray-700", props.Class)
    return core.Div(core.Class(class), props.Children...)
}

type CardContentProps struct {
    Class    string
    Children []core.Node
}

func CardContent(props CardContentProps) core.Node {
    class := mergeClasses("px-6 py-4", props.Class)
    return core.Div(core.Class(class), props.Children...)
}

type CardFooterProps struct {
    Class    string
    Children []core.Node
}

func CardFooter(props CardFooterProps) core.Node {
    class := mergeClasses("px-6 py-4 border-t border-gray-200 dark:border-gray-700", props.Class)
    return core.Div(core.Class(class), props.Children...)
}
```

### Layout Primitives
```go
// Source: Common design system patterns
package ui

import "github.com/dougbarrett/gux/core"

type ContainerProps struct {
    MaxWidth string // "sm", "md", "lg", "xl", "2xl", "7xl" (default)
    Class    string
    Children []core.Node
}

var containerMaxWidths = map[string]string{
    "sm":  "max-w-sm",
    "md":  "max-w-md",
    "lg":  "max-w-lg",
    "xl":  "max-w-xl",
    "2xl": "max-w-2xl",
    "7xl": "max-w-7xl",
}

func Container(props ContainerProps) core.Node {
    maxWidth := props.MaxWidth
    if maxWidth == "" {
        maxWidth = "7xl"
    }
    class := mergeClasses(
        containerMaxWidths[maxWidth],
        "mx-auto px-4 sm:px-6 lg:px-8",
        props.Class,
    )
    return core.Div(core.Class(class), props.Children...)
}

type StackProps struct {
    Gap      string // "0", "1", "2", "4", "6", "8" (default: "4")
    Align    string // "start", "center", "end", "stretch" (default: "stretch")
    Justify  string // "start", "center", "end", "between", "around"
    Class    string
    Children []core.Node
}

var alignClasses = map[string]string{
    "start":   "items-start",
    "center":  "items-center",
    "end":     "items-end",
    "stretch": "items-stretch",
}

var justifyClasses = map[string]string{
    "start":   "justify-start",
    "center":  "justify-center",
    "end":     "justify-end",
    "between": "justify-between",
    "around":  "justify-around",
}

func VStack(props StackProps) core.Node {
    gap := props.Gap
    if gap == "" {
        gap = "4"
    }
    class := mergeClasses(
        "flex flex-col",
        "gap-"+gap,
        alignClasses[props.Align],
        justifyClasses[props.Justify],
        props.Class,
    )
    return core.Div(core.Class(class), props.Children...)
}

func HStack(props StackProps) core.Node {
    gap := props.Gap
    if gap == "" {
        gap = "4"
    }
    class := mergeClasses(
        "flex flex-row",
        "gap-"+gap,
        alignClasses[props.Align],
        justifyClasses[props.Justify],
        props.Class,
    )
    return core.Div(core.Class(class), props.Children...)
}

type DividerProps struct {
    Orientation string // "horizontal" (default), "vertical"
    Class       string
}

func Divider(props DividerProps) core.Node {
    orientation := props.Orientation
    if orientation == "" {
        orientation = "horizontal"
    }
    var baseClass string
    if orientation == "vertical" {
        baseClass = "w-px h-full bg-gray-200 dark:bg-gray-700"
    } else {
        baseClass = "h-px w-full bg-gray-200 dark:bg-gray-700"
    }
    class := mergeClasses(baseClass, props.Class)
    return core.Div(core.Class(class))
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `components/` with `js.Value` | `core.Node` interface | Gux v2.0 | SSR compatibility |
| Imperative DOM manipulation | Declarative node trees | Gux v2.0 | No hydration mismatches |
| Single WASM bundle | Multi-bundle with WithBundle | v1.x | Smaller initial loads |
| Tailwind config.js | Tailwind v4 @theme (optional) | 2025 | CSS-first tokens |

**Deprecated/outdated:**
- `components/button.go` pattern using `js.Value` directly - replaced by `core.Node`
- `El(tag, className, children...)` from old components - use `core.El(tag, Attrs{}, children...)`

## Open Questions

Things that couldn't be fully resolved:

1. **Grid Component Specifics**
   - What we know: Grid needs responsive columns, gap control
   - What's unclear: Exact API for column counts at breakpoints
   - Recommendation: Start simple with fixed column props, iterate based on usage

2. **IconButton Implementation**
   - What we know: Need Button variant for icon-only buttons
   - What's unclear: How icons will be rendered (SVG strings? Components?)
   - Recommendation: Accept `core.Node` children, defer icon system to Phase 19

3. **ButtonGroup Layout**
   - What we know: Groups of buttons with shared borders/spacing
   - What's unclear: Whether to use HStack or custom component
   - Recommendation: Create ButtonGroup as specialized HStack with rounded corner handling

## Sources

### Primary (HIGH confidence)
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/node.go` - Node interface definition
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/elements.go` - Element helpers and Class() shorthand
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/html_renderer.go` - SSR rendering behavior
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/dom_renderer.go` - WASM rendering behavior
- `/Users/dougbarrett/projects/dbb1dev/goquery/examples/minimal/admin/dashboard.go` - Card-like patterns
- `/Users/dougbarrett/projects/dbb1dev/goquery/examples/minimal/admin/user_new.go` - formField pattern
- `/Users/dougbarrett/projects/dbb1dev/goquery/.planning/research/SUMMARY.md` - Prior v2.0 research

### Secondary (MEDIUM confidence)
- [shadcn/ui Button](https://ui.shadcn.com/docs/components/button) - Variant and size patterns
- [Tailwind CSS v4 @theme](https://tailwindcss.com/blog/tailwindcss-v4) - Design token patterns

### Tertiary (LOW confidence)
- [Tailwind CSS Best Practices](https://www.frontendtools.tech/blog/tailwind-css-best-practices-design-system-patterns) - Design system patterns

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Based on existing codebase analysis
- Architecture: HIGH - Patterns verified in examples/minimal
- Pitfalls: HIGH - Derived from existing code review
- Code examples: HIGH - Tested against core package API

**Research date:** 2026-01-22
**Valid until:** 2026-03-22 (90 days - stable domain)
