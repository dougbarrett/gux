# Coding Conventions

**Analysis Date:** 2026-02-01

## Naming Patterns

**Files:**
- Lowercase with underscores for multi-word names: `html_renderer.go`, `dom_renderer.go`, `router_server.go`
- Test files: `*_test.go` suffix (e.g., `button_test.go`, `endpoint_test.go`)
- Component files match component names: `button.go` contains `Button()`, `form.go` contains `FormField()`

**Functions:**
- Exported functions: PascalCase (e.g., `Text()`, `Button()`, `Input()`, `FormField()`)
- Unexported helper functions: camelCase (e.g., `renderHTML()`, `writeTestFile()`, `matchRoute()`)
- Constructor functions: `New` prefix for main types (e.g., `New()` returns `*App`)
- Predicate functions use present tense (e.g., `isNumericParam()`, `isExported()`)

**Variables:**
- Local variables: camelCase (e.g., `variant`, `size`, `buttonType`)
- Constants: PascalCase for exported, UPPER_SNAKE_CASE for style maps: `buttonVariantClasses`, `buttonSizeClasses`, `buttonBaseClasses`
- Package-level vars: camelCase or PascalCase based on export (e.g., `defaultWasmBinary`, `defaultWasmBundles`)

**Types:**
- Struct types: PascalCase (e.g., `ButtonProps`, `FormFieldProps`, `InputProps`)
- Enum-like constants: PascalCase (e.g., `ButtonPrimary`, `ButtonSecondary`, `InputEmail`, `ButtonSM`)
- Type aliases for enums: PascalCase suffix `string` (e.g., `ButtonVariant`, `ButtonSize`, `InputType`, `InputSize`)

**Interfaces:**
- Exported interfaces: PascalCase with `r` prefix when receiver parameter is standard (e.g., `Node` interface with `Render()` method)

## Code Style

**Formatting:**
- Go standard formatting (enforced by `gofmt`)
- Imports organized in groups: stdlib, third-party, local
- No single-letter variable names except in loop indices and short scopes

**Linting:**
- No explicit linter config detected, but code follows standard Go conventions
- Unused imports will be caught by compiler

**Comments:**
- Package-level documentation above `package` declaration
- Function documentation comments start with function name
- Inline comments for complex logic or non-obvious decisions
- Comments explain "why" not "what" (code shows "what")

**Example comment style from `core/node.go`:**
```go
// Package core provides the fundamental abstractions for gux's universal rendering.
//
// The key insight: components return Nodes, not DOM elements or HTML strings.
// Nodes are an intermediate representation that can be rendered to either target.
package core

// Node represents a renderable element in the UI tree.
// This is the universal currency - components produce Nodes,
// renderers consume them.
type Node interface {
    // Render converts this node to its final form using the given renderer.
    // For HTML: returns a string node containing the markup.
    // For DOM: returns a node wrapping the js.Value element.
    Render(r Renderer) RenderResult
}
```

## Import Organization

**Order:**
1. Standard library (e.g., `encoding/json`, `net/http`, `testing`)
2. Third-party packages (e.g., `gorm.io/gorm`, `github.com/...`)
3. Local imports (e.g., `"myapp/guxgen/models"`)

**Path Aliases:**
- Used sparingly when name conflicts occur
- Example from `build_test.go`: `usermodels "myapp/models"` when disambiguating between `models` and `guxgen/models`

**Pattern:** Bare imports for standard packages:
```go
import (
    "encoding/json"
    "net/http"
    "testing"

    "gorm.io/gorm"

    "myapp/guxgen/models"
)
```

## Error Handling

**Pattern: Constructor functions return errors**
```go
func (a *App) Run(addr string) error {
    // Implementation
}
```

**Pattern: API errors use typed *Error with HTTP status**
- Location: `api/errors.go`
- All API errors implement standard `error` interface
- Use error constructors: `api.NotFound()`, `api.BadRequest()`, `api.Unauthorized()`, `api.InternalError()`
- Error constructors have both plain and formatted variants: `NotFound(msg)` and `NotFoundf(format, args...)`

**Example from `api/errors.go`:**
```go
func NotFound(message string) *Error {
    return &Error{
        Status: http.StatusNotFound,
        Code:   "not_found",
        Message: message,
    }
}

func NotFoundf(format string, args ...any) *Error {
    return NotFound(fmt.Sprintf(format, args...))
}
```

**Pattern: Test errors use `t.Errorf()` or `t.Fatalf()`**
- For assertion failures: `t.Errorf()` to continue test
- For setup failures: `t.Fatalf()` to stop test
- Error messages include expected vs. actual values

**Example from `core/endpoint_test.go`:**
```go
if got != tt.want {
    t.Errorf("convertColonParams(%q) = %q, want %q", tt.input, got, tt.want)
}
```

## Logging

**Framework:** `println()` for debug/demo code

**Patterns:**
- Simple debugging: `println("message")`
- No structured logging library detected
- Errors returned via error values, not logged

## Comments & Documentation

**JSDoc/TSDoc:** Not applicable (Go codebase)

**Go documentation style:**
- All exported types and functions have documentation comments
- Comments start with the name being documented
- First sentence is a summary (appears in `godoc`)
- Multi-line comments provide more detail
- Package comments describe the overall purpose

**Example from `ui/button.go`:**
```go
// ButtonVariant defines the visual style of a button.
type ButtonVariant string

// Button creates a button component with variants and sizes.
//
// Example:
//
//    Button(ButtonProps{
//        Variant: ButtonPrimary,
//        Size: ButtonMD,
//        Children: []core.Node{core.Text("Click me")},
//    })
func Button(props ButtonProps) core.Node {
```

## Function Design

**Size:**
- Helper functions like `renderHTML()` are 3-5 lines
- Component rendering functions like `Button()` are 40-50 lines with setup + class building + return
- No explicit line limit, but functions stay focused on single responsibility

**Parameters:**
- Props structs instead of many positional parameters: `Button(ButtonProps{...})`
- Props structs hold optional fields (variant, size, class, disabled, etc.)
- Children passed as slice: `Children []core.Node`
- Event handlers passed as function pointers in props: `OnClick func()`

**Return Values:**
- Single return for success path (e.g., `func Button(props ButtonProps) core.Node`)
- Functions returning interfaces return the interface type, not concrete type
- Error-returning functions return `error` as last return value

**Example props struct pattern from `ui/button.go`:**
```go
type ButtonProps struct {
    Variant  ButtonVariant // Visual style (default: primary)
    Size     ButtonSize    // Size (default: md)
    Class    string        // Additional classes (override defaults)
    Disabled bool          // Disabled state
    Type     string        // Button type: "button", "submit", "reset" (default: "button")
    OnClick  func()        // Click handler (WASM only)
    Children []core.Node   // Button content
}
```

**Default application pattern:**
```go
// Apply defaults at function start
variant := props.Variant
if variant == "" {
    variant = ButtonPrimary  // Default
}

size := props.Size
if size == "" {
    size = ButtonMD  // Default
}
```

## Module Design

**Exports:**
- Only public API exported (PascalCase)
- All components are functions returning `core.Node`
- All props are exported structs
- Helper functions (lowercase) remain internal to package

**Pattern from `ui/button.go`:**
- Exported: `Button()`, `ButtonProps`, `ButtonVariant`, `ButtonSize`, and constants like `ButtonPrimary`
- Unexported: `buttonVariantClasses`, `buttonSizeClasses`, `buttonBaseClasses`, `buttonDisabledClasses`

**Barrel Files:** Not used; each component in its own file

**Type/Enum constants:**
- Type alias: `type ButtonVariant string`
- Constants: `const (ButtonPrimary ButtonVariant = "primary")`
- Maps for metadata: `var buttonVariantClasses = map[ButtonVariant]string{...}`

## Class Building Pattern

**UI components use `MergeClasses()` utility** (from `ui/button.go`):
```go
class := MergeClasses(
    buttonBaseClasses,
    buttonVariantClasses[variant],
    buttonSizeClasses[size],
    ConditionalClass(props.Disabled, buttonDisabledClasses),
    props.Class,
)
```

**Helper functions:**
- `MergeClasses(...string) string` - joins non-empty, trimmed strings
- `ConditionalClass(condition bool, class string) string` - returns class if true, empty if false

**Tailwind conventions:**
- Base classes for all instances
- Variant-specific classes (primary, secondary, outline, etc.)
- Size-specific classes (sm, md, lg)
- State classes (disabled, focus, hover)
- Extra/custom classes appended last for override capability

## Attribute Pattern

**Attributes with structured props:**
```go
attrs := core.Attrs{
    Class:   class,
    Type:    buttonType,
    OnClick: props.OnClick,
}

if props.Disabled {
    attrs.Extra = map[string]string{
        "disabled": "disabled",
    }
}
```

**Data attributes and extras:**
- Standard HTML attributes: `ID`, `Class`, `Href`, `Src`, `Alt`, `Type`, `Name`, `Value`
- Non-standard attributes in `Extra map[string]string`
- Example: `attrs.Extra = map[string]string{"for": labelID, "role": "alert"}`

---

*Convention analysis: 2026-02-01*
