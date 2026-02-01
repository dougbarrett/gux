# Testing Patterns

**Analysis Date:** 2026-02-01

## Test Framework

**Runner:**
- Go standard `testing` package (stdlib)
- No external test framework detected

**Run Commands:**
```bash
go test ./...              # Run all tests
go test ./... -count=1     # Run without test caching
go test ./... -timeout 120s  # With timeout
```

**Note:** Full test suite command from CLAUDE.md:
```bash
go test ./... -count=1 -timeout 120s
```

## Test File Organization

**Location:** Co-located with source code
- Test file next to implementation: `button.go` → `button_test.go`
- Pattern: `*_test.go` suffix
- Same package as implementation: tests in `package ui`, `package core`, etc.

**Files tested (37 test files total):**
- `ui/` directory: 24 test files (button_test.go, input_test.go, form_test.go, table_test.go, etc.)
- `core/` directory: 3 test files (endpoint_test.go, crud_test.go, audit_test.go)
- `cmd/gux/` directory: 3 test files (build_test.go, modelgen_test.go, generate_test.go)

## Test Structure

**Pattern: Table-driven tests with `t.Run()` subtests**

From `ui/button_test.go`:
```go
func TestButton_AllVariants(t *testing.T) {
    tests := []struct {
        variant       ButtonVariant
        expectedClass string
    }{
        {ButtonPrimary, "bg-blue-600"},
        {ButtonSecondary, "bg-gray-200"},
        {ButtonOutline, "border border-gray-300"},
        {ButtonGhost, "bg-transparent"},
        {ButtonDestructive, "bg-red-600"},
    }

    for _, tt := range tests {
        t.Run(string(tt.variant), func(t *testing.T) {
            btn := Button(ButtonProps{
                Variant:  tt.variant,
                Children: []core.Node{core.Text("Button")},
            })
            html := renderHTML(btn)

            if !strings.Contains(html, tt.expectedClass) {
                t.Errorf("variant %s: expected class %q in: %s", tt.variant, tt.expectedClass, html)
            }
        })
    }
}
```

**Key patterns:**
- `tests := []struct{ ... }` - anonymous struct slice for test cases
- `for _, tt := range tests` - iterate with underscore for index
- `t.Run(name, func(t *testing.T) { ... })` - subtest with descriptive name
- Each test case is self-contained

**Naming convention:**
- Test function: `TestFunctionName` or `TestType_Scenario`
- Subtest names: lowercase, descriptive (e.g., "no params", "single param", "disabled", "all variants")
- Examples: `TestButton_DefaultVariant`, `TestButton_Disabled`, `TestConvertColonParams`, `TestExtractParamNames`

## Helper Functions

**Shared test helpers defined at package level:**

From `ui/button_test.go`:
```go
// renderHTML renders a node to HTML string for testing.
func renderHTML(node core.Node) string {
    renderer := core.HTML()
    result := node.Render(renderer)
    return result.HTML()
}
```

**Helper characteristics:**
- Receive `*testing.T` parameter if they call `t.Fatalf()`
- Receive `t *testing.T` and call `t.Helper()` at function start to report errors at call site
- Reused across multiple tests in same package

From `cmd/gux/build_test.go`:
```go
// writeTestFile creates a temporary Go source file for testing parseCRUDModels.
func writeTestFile(t *testing.T, content string) string {
    t.Helper()  // Report errors at call site
    dir := t.TempDir()
    path := filepath.Join(dir, "app.go")
    if err := os.WriteFile(path, []byte(content), 0644); err != nil {
        t.Fatalf("write test file: %v", err)
    }
    return path
}
```

## Assertions

**Pattern: Direct value comparison with if statements**
- No assertion library (AssertJ, testify, etc.)
- Compare values directly using `if` and error reporting

From `ui/button_test.go`:
```go
// Should render as button element
if !strings.HasPrefix(html, "<button") {
    t.Errorf("expected <button> element, got: %s", html)
}

// Should have primary variant classes (default)
if !strings.Contains(html, "bg-blue-600") {
    t.Errorf("expected primary variant classes, got: %s", html)
}
```

**Assertion patterns:**
- `strings.Contains()` for substring checks in HTML
- `strings.HasPrefix()` / `strings.HasSuffix()` for HTML tags
- Direct equality: `if got != tt.want`
- Slice length: `if len(got) != len(tt.want)`

**Formatting in error messages:**
```go
t.Errorf("FunctionName(%q) = %q, want %q", input, got, want)
t.Errorf("variant %s: expected class %q in: %s", variant, expectedClass, html)
```

## Mocking

**Approach: Minimal mocking**
- Prefer real objects over mocks
- Use `reflect.Value` with mock values for testing parameter handling
- Use `httptest.NewRequest()` for HTTP request testing
- Use `httptest.NewRecorder()` for HTTP response testing

**Example from `core/crud_test.go` (reflect-based mock):**
```go
mockDB := reflect.ValueOf(struct{}{})
result := applyQueryFilters(mockDB, req, modelType)
```

**Example from `core/audit_test.go` (httptest):**
```go
r, _ := http.NewRequest("GET", "/", nil)
r.Header.Set("X-Forwarded-For", "203.0.113.50")
r.RemoteAddr = "10.0.0.1:12345"

ip := getClientIP(r)
```

**What to mock:**
- Requests/responses: Use `httptest` package
- Complex external dependencies: Use `reflect.Value` or minimal stubs

**What NOT to mock:**
- Standard library types (use real `*http.Request`, real `testing.T`, etc.)
- Model structs (use real types with test data)
- Rendered HTML (test against actual HTML output)

## Test-Specific Patterns

**Temporary directories:** Tests that need file system use `t.TempDir()`
```go
dir := t.TempDir()
path := filepath.Join(dir, "app.go")
os.WriteFile(path, []byte(content), 0644)
```

**Field access validation:** Tests verify struct field mapping
From `core/crud_test.go`:
```go
type TestModel struct {
    ID        uint   `json:"id"`
    Name      string `json:"name"`
    OrderID   uint   `json:"order_id"`
    Active    bool   `json:"active"`
    Price     float64
    Secret    string `json:"-"`
}

modelType := reflect.TypeOf(TestModel{})
```

**JSON parsing in tests:** Tests verify JSON output
From `core/audit_test.go`:
```go
var diff map[string]interface{}
if err := json.Unmarshal([]byte(result), &diff); err != nil {
    t.Fatalf("failed to parse diff JSON: %v", err)
}

if _, ok := diff["name"]; ok {
    t.Error("name should not appear in diff (unchanged)")
}
```

## Coverage

**Requirements:** Not enforced (no coverage flags in build config)

**Test count:** 37 test files across UI, core, and cmd packages

**Key areas tested:**
- `ui/` components: Button, Input, Form, Table, DataTable, Select, Checkbox, Radio, Switch, etc.
- `core/` endpoints: Path parameter conversion, extraction, HTTP API handling
- `core/` CRUD: Query filtering, snake_case conversion, pluralization
- `core/` audit: Field diffing, snapshots, entity ID extraction, IP detection, audit logging
- `cmd/gux/` build: CRUD model parsing, DTO imports, model import path resolution
- `cmd/gux/` generation: Dockerfile regeneration, model generation, template field conversion

## Integration Tests

**Pattern: Tests use real types but limited integration**
- UI tests use `core.HTML()` renderer to generate real HTML output
- Tests verify HTML structure and content rather than mocking rendering
- CRUD tests use `reflect.ValueOf()` to test field resolution logic

From `ui/button_test.go`:
```go
btn := Button(ButtonProps{
    Variant:  ButtonPrimary,
    Size:     ButtonMD,
    Children: []core.Node{core.Text("Button")},
})
html := renderHTML(btn)  // Render to real HTML
if !strings.Contains(html, "bg-blue-600") { ... }
```

**Full integration test pattern:** Not present in current tests
- No `httptest.Server` integration tests
- No database integration tests (GORM mocked with reflect)

## Important Test Requirements

**From CLAUDE.md documentation:**
- All bug fixes and new features must include unit tests
- Pointer type handling: Every code generation path must test with pointer types (`*string`, `*float64`, `*uint`, `*time.Time`)
- Path parameters: Test both `:param` and `{param}` syntax; test numeric (`id`, `user_id`) and string (`slug`, `code`) params
- Table-driven tests: Use `t.Run()` subtests
- Temp directories: Use `t.TempDir()` with `os.Chdir()` + defer to restore
- Cache clearing: Tests using `getModelFieldTypes()` must reset `modelFieldTypesCache`

**Test file locations (from CLAUDE.md):**

| File | Coverage |
|------|----------|
| `cmd/gux/build_test.go` | CRUD parsing, endpoint generation, pointer type mapping, `isNumericParam`, `generateFieldMapping`, `generateServerAPICode` |
| `cmd/gux/modelgen_test.go` | Template field conversion, form field generation, detail field generation, pointer types, display field detection, DTO type overrides, GORM tags |
| `cmd/gux/generate_test.go` | Dockerfile regeneration |
| `core/endpoint_test.go` | Path parameter extraction, `:param` to `{param}` conversion, param name extraction |
| `core/crud_test.go` | CRUD API operations |
| `core/audit_test.go` | Audit logging |

## Example Complete Test

From `ui/button_test.go` - comprehensive test combining multiple patterns:
```go
func TestButton_CombinedProps(t *testing.T) {
    btn := Button(ButtonProps{
        Variant:  ButtonDestructive,
        Size:     ButtonLG,
        Class:    "extra-class",
        Disabled: true,
        Type:     "submit",
        Children: []core.Node{core.Text("Delete")},
    })
    html := renderHTML(btn)

    // Check variant
    if !strings.Contains(html, "bg-red-600") {
        t.Errorf("expected destructive variant, got: %s", html)
    }

    // Check size
    if !strings.Contains(html, "px-6 py-3") {
        t.Errorf("expected large size, got: %s", html)
    }

    // Check custom class
    if !strings.Contains(html, "extra-class") {
        t.Errorf("expected extra-class, got: %s", html)
    }

    // Check disabled
    if !strings.Contains(html, "opacity-50") {
        t.Errorf("expected disabled styling, got: %s", html)
    }
    if !strings.Contains(html, `disabled="disabled"`) {
        t.Errorf("expected disabled attribute, got: %s", html)
    }

    // Check type
    if !strings.Contains(html, `type="submit"`) {
        t.Errorf("expected type=submit, got: %s", html)
    }

    // Check content
    if !strings.Contains(html, "Delete") {
        t.Errorf("expected Delete content, got: %s", html)
    }
}
```

---

*Testing analysis: 2026-02-01*
