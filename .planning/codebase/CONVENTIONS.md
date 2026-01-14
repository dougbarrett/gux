# Coding Conventions

**Analysis Date:** 2026-01-14

## Naming Patterns

**Files:**
- `snake_case.go` for multi-word names (`data_display.go`, `file_upload.go`, `virtual_list.go`)
- `lowercase.go` for single-word names (`button.go`, `input.go`, `modal.go`)
- `*_gen.go` for generated files (`posts_client_gen.go`, `posts_server_gen.go`)
- `*_ws.go` for WebSocket handlers (`posts_ws.go`)

**Functions:**
- `PascalCase` for exported functions (`NewTable`, `PrimaryButton`)
- `camelCase` for unexported functions
- `New` prefix for constructors returning complex types (`NewPostsService`, `NewPostsWSHandler`)
- Direct name for simple constructors (`Button`, `Input`, `Card`)
- `With` prefix for option functions (`WithOnOpen`, `WithOnClose`, `WithOnError`)

**Variables:**
- `camelCase` for local variables
- Short names in tight scopes (`v`, `err`, `fn`, `sub`, `id`, `idx`)
- Descriptive names for struct fields (`ClassName`, `OnClick`, `Disabled`)

**Types:**
- `PascalCase` for all types (`ButtonProps`, `Store`, `PostsAPI`)
- `[Type][Variant]` for grouped constants (`ButtonPrimary`, `InputText`, `TextMuted`)
- No `I` prefix for interfaces (`PostsAPI`, not `IPostsAPI`)

## Code Style

**Formatting:**
- Standard `gofmt` formatting
- Tab indentation (Go default)
- No explicit line length limit (follows Go conventions)
- Double quotes for strings

**Linting:**
- No explicit linting configuration (`.golangci.yml` not present)
- Assumes IDE integration for `gofmt` and `goimports`

## Import Organization

**Order:**
1. Standard library (`fmt`, `context`, `encoding/json`)
2. External packages (`github.com/gorilla/websocket`)
3. Internal packages (`gux/api`, `gux/components`)

**Grouping:**
- Blank line between standard library and external
- No explicit sorting enforced

**Path Aliases:**
- None used (direct imports)

## Error Handling

**Patterns:**
- Return errors, check at call site
- Wrap errors with context: `fmt.Errorf("marshal payload: %w", err)`
- Helper constructors for API errors: `api.NotFound()`, `api.BadRequest()`

**Error Types:**
- `api.Error` struct with Status, Code, Message (`api/errors.go`)
- Standard `error` interface throughout
- Formatted errors with `%f` variants: `NotFoundf(format, args...)`

## Logging

**Framework:**
- `log.Printf` for server-side logging
- `console.log` equivalent via `js.Global()` for WASM
- No structured logging library

**Patterns:**
- Log errors with context: `log.Printf("Failed to marshal: %v", err)`
- Log connection events: `log.Printf("WebSocket client connected")`

## Comments

**When to Comment:**
- All exported types and functions (GoDoc convention)
- Comments are complete sentences ending with periods
- Package-level documentation at top of primary file

**GoDoc Style:**
```go
// ButtonVariant defines button color variants
type ButtonVariant string

// NewTable creates a data table component with sortable columns
func NewTable(props TableProps) *Table {
```

**Inline Comments:**
- Sparse, used for non-obvious logic
- Example: `// Return unsubscribe function`

## Function Design

**Size:**
- Most functions under 50 lines
- Larger functions in complex components (formbuilder, charts)

**Parameters:**
- Props struct pattern for component configuration
- Options pattern for configurable clients (`ws.Option`)
- Max 3-4 positional parameters, then use struct

**Return Values:**
- Explicit returns (no naked returns)
- Multiple returns: `(result, error)`
- Single returns for simple getters

## Module Design

**Exports:**
- Named exports for all public APIs
- No default exports (Go standard)
- One primary type per file typically

**Build Tags:**
- `//go:build js && wasm` for WASM-only code
- Tag on first line of file
- Blank line after tag

## Patterns

**Props Pattern:**
```go
type ButtonProps struct {
    Text      string
    ClassName string
    Variant   ButtonVariant
    OnClick   func()
}

func Button(props ButtonProps) js.Value {
```

**Variant Maps:**
```go
var buttonVariantClasses = map[ButtonVariant]string{
    ButtonPrimary:   "bg-blue-500 text-white hover:bg-blue-600",
    ButtonSecondary: "bg-gray-200 dark:bg-gray-700...",
}
```

**Functional Options:**
```go
type Option func(*Client)

func WithOnOpen(fn func()) Option {
    return func(c *Client) { c.onOpen = fn }
}

client := ws.NewClient(url, WithOnOpen(handler))
```

**Middleware Chain:**
```go
func Chain(middlewares ...Middleware) Middleware {
    return func(next http.Handler) http.Handler {
        for i := len(middlewares) - 1; i >= 0; i-- {
            next = middlewares[i](next)
        }
        return next
    }
}
```

**API Annotations:**
```go
// @client PostsClient
// @basepath /api/posts
type PostsAPI interface {
    // @route GET /
    GetAll(ctx context.Context) ([]Post, error)
}
```

---

*Convention analysis: 2026-01-14*
*Update when patterns change*
