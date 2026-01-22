# Phase 18: Data Display Components - Research

**Researched:** 2026-01-22
**Domain:** Data visualization and display components
**Confidence:** HIGH (based on existing codebase patterns)

## Summary

Phase 18 implements data display components: Table (compound pattern), DataTable[T] (generics), Badge (variants), Avatar (fallback), Pagination, and List. These components build directly on patterns established in Phases 16-17.

The existing codebase provides clear templates:
- **Card compound pattern** (Phase 16): Template for Table compound structure
- **Button variants** (Phase 16): Template for Badge variants
- **Input props pattern** (Phase 17): Template for all component props

The `components/` directory contains WASM-only versions (using `js.Value`) of Table, Badge, Avatar, and Pagination that demonstrate desired behavior but need reimplementation using `core.Node` for SSR/WASM compatibility.

**Primary recommendation:** Follow established patterns exactly. Table uses compound components (Table/Thead/Tbody/Tr/Th/Td), Badge uses variant maps, Avatar handles image error via conditional rendering (not JS events), DataTable[T] uses Go generics following existing `UseState[T]` pattern.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/dougbarrett/gux/core | local | Node interface, elements | Foundation of all components |
| github.com/dougbarrett/gux/ui | local | Component library | Where new components go |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| MergeClasses | ui/utils.go | CSS class combination | Every component |
| ConditionalClass | ui/utils.go | Conditional CSS | State-dependent styling |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Table compound | Single Table component | Compound offers more flexibility for custom layouts |
| DataTable[T] generics | map[string]any | Generics provide type safety, compile-time checks |
| Manual pagination | DataTable built-in | Separate Pagination allows use with any data source |

**Installation:**
No additional packages needed. All components build on existing `core` and `ui` packages.

## Architecture Patterns

### Recommended Project Structure
```
ui/
  table.go           # Table, Thead, Tbody, Tr, Th, Td (compound)
  table_test.go
  datatable.go       # DataTable[T] (generic)
  datatable_test.go
  badge.go           # Badge with variants
  badge_test.go
  avatar.go          # Avatar with fallback
  avatar_test.go
  pagination.go      # Pagination
  pagination_test.go
  list.go            # List, ListItem
  list_test.go
```

### Pattern 1: Compound Components (Table)

**What:** Family of related components that work together
**When to use:** Complex structures with multiple parts (headers, rows, cells)
**Example:**
```go
// Source: Established pattern from ui/card.go
// Table compound pattern - each part is independent but designed to compose

func Table(props TableProps) core.Node {
    class := MergeClasses(
        "min-w-full divide-y divide-gray-200 dark:divide-gray-700",
        props.Class,
    )
    return core.Table(core.Class(class), props.Children...)
}

func Thead(props TheadProps) core.Node {
    class := MergeClasses(
        "bg-gray-50 dark:bg-gray-800",
        props.Class,
    )
    return core.Thead(core.Class(class), props.Children...)
}

func Tr(props TrProps) core.Node {
    return core.Tr(core.Attrs{
        Class:   MergeClasses(props.Class),
        OnClick: props.OnClick,
    }, props.Children...)
}

// Usage mirrors HTML structure
Table(TableProps{},
    Thead(TheadProps{},
        Tr(TrProps{},
            Th(ThProps{}, core.Text("Name")),
            Th(ThProps{}, core.Text("Email")),
        ),
    ),
    Tbody(TbodyProps{},
        Tr(TrProps{},
            Td(TdProps{}, core.Text("John")),
            Td(TdProps{}, core.Text("john@example.com")),
        ),
    ),
)
```

### Pattern 2: Variant Maps (Badge)

**What:** Type-safe variants using string constants and maps
**When to use:** Components with multiple visual styles
**Example:**
```go
// Source: Established pattern from ui/button.go

type BadgeVariant string

const (
    BadgeDefault BadgeVariant = "default"
    BadgeSuccess BadgeVariant = "success"
    BadgeWarning BadgeVariant = "warning"
    BadgeError   BadgeVariant = "error"
)

var badgeVariantClasses = map[BadgeVariant]string{
    BadgeDefault: "bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200",
    BadgeSuccess: "bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200",
    BadgeWarning: "bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200",
    BadgeError:   "bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200",
}
```

### Pattern 3: Generic DataTable[T]

**What:** Type-safe data table using Go generics
**When to use:** Rendering typed collections with column definitions
**Example:**
```go
// Source: Follows pattern from core/app.go UseState[T]

type ColumnDef[T any] struct {
    Header string
    Render func(T) core.Node
    Width  string // Optional width (e.g., "w-48")
}

type DataTableProps[T any] struct {
    Data       []T
    Columns    []ColumnDef[T]
    RowKey     func(T) string      // For keys/identification
    OnRowClick func(T)             // Optional row click handler
    Striped    bool
    Hoverable  bool
    Class      string
}

func DataTable[T any](props DataTableProps[T]) core.Node {
    // Build headers
    var headers []core.Node
    for _, col := range props.Columns {
        headers = append(headers, Th(ThProps{Class: col.Width}, core.Text(col.Header)))
    }

    // Build rows
    var rows []core.Node
    for i, item := range props.Data {
        var cells []core.Node
        for _, col := range props.Columns {
            cells = append(cells, Td(TdProps{}, col.Render(item)))
        }

        // Capture for closure
        capturedItem := item
        rowClass := ""
        if props.Striped && i%2 == 1 {
            rowClass = "bg-gray-50 dark:bg-gray-800"
        }
        if props.Hoverable {
            rowClass = MergeClasses(rowClass, "hover:bg-gray-100 dark:hover:bg-gray-700")
        }

        var onClick func()
        if props.OnRowClick != nil {
            onClick = func() { props.OnRowClick(capturedItem) }
        }

        rows = append(rows, Tr(TrProps{Class: rowClass, OnClick: onClick}, cells...))
    }

    return Table(TableProps{Class: props.Class},
        Thead(TheadProps{}, Tr(TrProps{}, headers...)),
        Tbody(TbodyProps{}, rows...),
    )
}

// Usage
DataTable(DataTableProps[User]{
    Data: users,
    Columns: []ColumnDef[User]{
        {Header: "Name", Render: func(u User) core.Node { return core.Text(u.Name) }},
        {Header: "Email", Render: func(u User) core.Node { return core.Text(u.Email) }},
        {Header: "Status", Render: func(u User) core.Node {
            return Badge(BadgeProps{
                Text: u.Status,
                Variant: BadgeSuccess,
            })
        }},
    },
    OnRowClick: func(u User) { /* navigate */ },
})
```

### Pattern 4: Conditional Fallback (Avatar)

**What:** Show fallback content when primary content unavailable
**When to use:** Images that may fail to load, optional data
**Example:**
```go
// Note: SSR cannot detect image load errors, so Avatar must decide at render time
// based on whether Src is provided

func Avatar(props AvatarProps) core.Node {
    sizeClass := avatarSizeClasses[props.Size]
    if sizeClass == "" {
        sizeClass = avatarSizeClasses[AvatarMD]
    }

    baseClass := MergeClasses(
        "rounded-full overflow-hidden flex items-center justify-center",
        sizeClass,
        props.Class,
    )

    // Show image if Src provided, otherwise show initials
    if props.Src != "" {
        return core.Div(core.Class(baseClass),
            core.Img(core.Attrs{
                Src:   props.Src,
                Alt:   props.Alt,
                Class: "w-full h-full object-cover",
            }),
        )
    }

    // Fallback: initials
    initials := getInitials(props.Name)
    return core.Div(core.Attrs{
        Class: MergeClasses(baseClass, "bg-gray-200 dark:bg-gray-600 text-gray-600 dark:text-gray-200 font-medium"),
    }, core.Text(initials))
}
```

### Anti-Patterns to Avoid

- **JS-dependent fallbacks in SSR:** Don't use `onerror` for image fallback - decide at render time
- **Children as variadic + props.Children:** Pick one approach per component (props.Children preferred for compound components)
- **State in stateless components:** Table, Badge, Avatar, List are all stateless - state belongs in the page using them
- **map[string]any for data:** Use generics for type safety in DataTable

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| CSS class merging | String concatenation | MergeClasses() | Handles empty strings, whitespace |
| Conditional classes | Ternary in template | ConditionalClass() | Cleaner, tested |
| Initials from name | String indexing | getInitials() helper | Handles edge cases (single name, empty) |
| Page calculations | Manual math | (total + pageSize - 1) / pageSize | Standard ceiling division |

**Key insight:** All utility functions already exist in `ui/utils.go`. Component patterns established in Phases 16-17 should be followed exactly.

## Common Pitfalls

### Pitfall 1: Image Fallback in SSR
**What goes wrong:** Using JavaScript `onerror` event for image fallback
**Why it happens:** WASM components often rely on JS events
**How to avoid:** Decide fallback at render time based on props - if no `Src`, show initials
**Warning signs:** Components rendering differently in SSR vs WASM

### Pitfall 2: Closure Capture in Loops
**What goes wrong:** Row click handlers all refer to last item
**Why it happens:** Go closure captures variable reference, not value
**How to avoid:** `capturedItem := item` inside loop before using in closure
**Warning signs:** All rows trigger same action

### Pitfall 3: Generic Type Inference
**What goes wrong:** `DataTable(props)` without type parameter
**Why it happens:** Go needs type parameter when inferring from struct
**How to avoid:** Always specify: `DataTable[User](props)` or let compiler infer from data
**Warning signs:** Compile errors about type inference

### Pitfall 4: Table Without Overflow
**What goes wrong:** Wide tables break layout
**Why it happens:** Forgetting overflow-x-auto on container
**How to avoid:** Table component should include overflow handling or document wrapper requirement
**Warning signs:** Horizontal scrollbar on page instead of table

### Pitfall 5: Pagination Off-by-One
**What goes wrong:** Wrong page calculations (0-indexed vs 1-indexed)
**Why it happens:** Mixing page numbering conventions
**How to avoid:** Document and consistently use 1-indexed pages (user-facing)
**Warning signs:** "Page 0", empty last page, wrong "Showing X-Y of Z"

## Code Examples

Verified patterns from official sources (existing codebase):

### Props Struct Pattern
```go
// Source: ui/button.go - standard props struct format

type BadgeProps struct {
    Text    string       // Required: display text
    Variant BadgeVariant // Visual style (default: default)
    Size    BadgeSize    // Size (default: md)
    Rounded bool         // Pill style vs rectangular
    Class   string       // Additional classes
}
```

### Variant with Default
```go
// Source: ui/button.go - handling defaults

func Badge(props BadgeProps) core.Node {
    variant := props.Variant
    if variant == "" {
        variant = BadgeDefault
    }
    // ...
}
```

### Testing Pattern
```go
// Source: ui/button_test.go - table-driven tests

func TestBadge_AllVariants(t *testing.T) {
    tests := []struct {
        variant       BadgeVariant
        expectedClass string
    }{
        {BadgeDefault, "bg-gray-100"},
        {BadgeSuccess, "bg-green-100"},
        {BadgeWarning, "bg-yellow-100"},
        {BadgeError, "bg-red-100"},
    }

    for _, tt := range tests {
        t.Run(string(tt.variant), func(t *testing.T) {
            badge := Badge(BadgeProps{Text: "Test", Variant: tt.variant})
            html := renderHTML(badge)

            if !strings.Contains(html, tt.expectedClass) {
                t.Errorf("variant %s: expected %q in: %s", tt.variant, tt.expectedClass, html)
            }
        })
    }
}
```

### Compound Component Composition
```go
// Source: ui/card.go - compound pattern with children

// Usage in practice
Card(CardProps{},
    CardHeader(CardHeaderProps{}, core.H2(core.Class("font-bold"), core.Text("Title"))),
    CardContent(CardContentProps{}, core.Text("Body content")),
    CardFooter(CardFooterProps{}, Button(ButtonProps{}, core.Text("Save"))),
)
```

### Initials Helper
```go
// Source: components/avatar.go (WASM version) - reimplement without js

func getInitials(name string) string {
    if name == "" {
        return "?"
    }
    parts := strings.Fields(name)
    if len(parts) == 0 {
        return "?"
    }
    if len(parts) == 1 {
        if len(parts[0]) > 0 {
            return strings.ToUpper(string(parts[0][0]))
        }
        return "?"
    }
    first := string(parts[0][0])
    last := string(parts[len(parts)-1][0])
    return strings.ToUpper(first + last)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| WASM-only components (js.Value) | core.Node universal | v2.0 | Components work in SSR + WASM |
| interface{} for data | Go generics | Go 1.18+ | Type-safe DataTable[T] |
| js.Value event handlers | core.Attrs callbacks | v2.0 | OnClick works universally |

**Deprecated/outdated:**
- `components/table.go` (WASM-only): Being replaced by `ui/table.go` using core.Node
- `components/badge.go` (WASM-only): Being replaced by `ui/badge.go` using core.Node
- `components/avatar.go` (WASM-only): Being replaced by `ui/avatar.go` using core.Node
- `components/pagination.go` (WASM-only): Being replaced by `ui/pagination.go` using core.Node

## Open Questions

Things that couldn't be fully resolved:

1. **Avatar WASM-side image error handling**
   - What we know: SSR cannot handle image load errors, must decide at render
   - What's unclear: Whether to add WASM-side error handling for runtime fallback
   - Recommendation: Start with SSR-compatible approach (no Src = show initials). Can enhance later if needed.

2. **DataTable Pagination Integration**
   - What we know: DataTable and Pagination are separate components
   - What's unclear: Should DataTable optionally include pagination, or always external?
   - Recommendation: Keep separate for flexibility. DataTable renders data, Pagination is used alongside.

3. **Table Overflow Wrapper**
   - What we know: Wide tables need overflow-x-auto
   - What's unclear: Should Table include wrapper or document requirement?
   - Recommendation: Include overflow wrapper in Table component for safety.

## Sources

### Primary (HIGH confidence)
- `/Users/dougbarrett/projects/dbb1dev/goquery/ui/card.go` - Compound component pattern
- `/Users/dougbarrett/projects/dbb1dev/goquery/ui/button.go` - Variant pattern
- `/Users/dougbarrett/projects/dbb1dev/goquery/ui/input.go` - Props struct pattern
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/elements.go` - Table/Thead/Tbody/Tr/Th/Td helpers exist
- `/Users/dougbarrett/projects/dbb1dev/goquery/.planning/research/ARCHITECTURE.md` - DataTable[T] design

### Secondary (MEDIUM confidence)
- `/Users/dougbarrett/projects/dbb1dev/goquery/components/table.go` - WASM reference (behavior, not structure)
- `/Users/dougbarrett/projects/dbb1dev/goquery/components/badge.go` - WASM reference
- `/Users/dougbarrett/projects/dbb1dev/goquery/components/avatar.go` - WASM reference
- `/Users/dougbarrett/projects/dbb1dev/goquery/components/pagination.go` - WASM reference

### Tertiary (LOW confidence)
None - all patterns verified from existing codebase.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Uses existing core and ui packages
- Architecture: HIGH - Follows established patterns from Phases 16-17
- Pitfalls: HIGH - Based on direct codebase analysis

**Research date:** 2026-01-22
**Valid until:** Indefinite (based on stable internal patterns)

## Component Implementation Order

Recommended build order based on dependencies:

1. **Badge** - No dependencies, simple variant pattern
2. **Avatar** - No dependencies, demonstrates fallback pattern
3. **Table compound** (Table, Thead, Tbody, Tr, Th, Td) - No dependencies, needed by DataTable
4. **List/ListItem** - No dependencies, simple compound
5. **Pagination** - No dependencies, needed alongside DataTable
6. **DataTable[T]** - Depends on Table compound, uses generics

Each component: implementation file + test file, following Phase 16-17 patterns.
