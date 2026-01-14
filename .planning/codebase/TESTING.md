# Testing Patterns

**Analysis Date:** 2026-01-14

## Test Framework

**Runner:**
- No test framework configured
- Go's built-in `testing` package available but unused

**Assertion Library:**
- Not applicable (no tests)

**Run Commands:**
```bash
go test ./...              # Would run all tests (none exist)
go test -v ./...           # Verbose mode
go test -cover ./...       # Coverage report
```

## Test File Organization

**Location:**
- No `*_test.go` files found in codebase
- Expected pattern: Co-located with source (`button_test.go` alongside `button.go`)

**Naming:**
- Standard Go convention: `{file}_test.go`

**Current Structure:**
```
components/
  button.go           # Source
  (no button_test.go) # Missing
state/
  store.go            # Source
  (no store_test.go)  # Missing
```

## Test Structure

**Expected Pattern (not implemented):**
```go
package components

import "testing"

func TestButton(t *testing.T) {
    t.Run("renders with props", func(t *testing.T) {
        // arrange
        props := ButtonProps{Text: "Click me"}

        // act
        result := Button(props)

        // assert
        if result.IsUndefined() {
            t.Error("expected button element")
        }
    })
}
```

## Mocking

**Framework:**
- Not applicable (no tests)
- Would use Go's interface-based mocking

**What Would Need Mocking:**
- Browser APIs (`syscall/js` calls)
- HTTP responses (for fetch tests)
- WebSocket connections

## Coverage

**Requirements:**
- No coverage requirements defined
- No CI/CD pipeline enforcing coverage

**Configuration:**
- Would use `go test -cover`

## Test Types

**Unit Tests:**
- Not implemented
- Would test: Store operations, validation rules, error handling

**Integration Tests:**
- Not implemented
- Would test: API client/server roundtrip, middleware chain

**E2E Tests:**
- Not implemented
- Could use Playwright for browser automation

## Testability Design

The code is structured for testability despite lacking tests:

**Dependency Injection:**
- Props structs allow controlled component creation
- Callbacks (`OnClick`, `OnChange`) are injectable
- Options pattern for WebSocket client configuration

**Interface-Based Design:**
- `PostsAPI` interface enables mock implementations
- Service implementations separate from handlers

**Pure Functions:**
- Validation rules are pure: `func(string) bool`
- Store transformations are pure: `func(T) U`

## Testing Gaps

**Critical Paths Without Tests:**
- Auth token extraction and expiration (`auth/auth.go`)
- State store subscriptions (`state/store.go`)
- WebSocket reconnection logic (`ws/ws.go`)
- Form validation rules (`components/validation.go`)
- API code generation (`cmd/apigen/main.go`)

**High Risk Areas:**
- JSON marshaling/unmarshaling
- localStorage persistence
- Error handling paths

## Build & CI

**Makefile Targets:**
```bash
make dev-tinygo     # Build and run
make docker         # Build container
# No test target defined
```

**Recommended Additions:**
```makefile
test:
    go test ./...

test-cover:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out
```

## Documentation Testing

**README Examples:**
- Code examples in README.md are not tested
- Could use `testable examples` in Go

**Component Documentation:**
- COMPONENTS.md has usage examples (untested)

---

*Testing analysis: 2026-01-14*
*Update when test infrastructure is added*
