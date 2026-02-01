# Codebase Concerns

**Analysis Date:** 2026-02-01

## Resource Leaks & Memory Management

### Ticker Resource Leak in CSRF and Auth Middleware

**Issue:** Background cleanup goroutines created in `CSRFMiddleware()` and `AuthMiddleware()` initialize tickers that are never stopped.

**Files:**
- `core/csrf.go` (lines 184-190)
- `core/auth.go` (lines 155-161)

**Problem:**
```go
// In CSRFMiddleware
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        tokenStore.Cleanup()
    }
}() // Ticker never stopped
```

**Impact:**
- Each call to `CSRFMiddleware()` spawns a goroutine with a new ticker
- Middleware is typically instantiated once, but the pattern is fragile
- In testing environments with multiple app instances, this compounds to wasted resources
- Long-running servers accumulate stopped goroutines

**Mitigation needed:**
- Store ticker reference and provide explicit cleanup mechanism
- Call `ticker.Stop()` on shutdown
- Consider using context cancellation pattern instead

---

### CSRF Token Store Memory Growth

**Issue:** In-memory CSRF token store (`csrfTokenStore`) grows monotonically if cleanup doesn't run or if tokens aren't accessed frequently.

**Files:** `core/csrf.go` (lines 70-107)

**Problem:**
- Tokens map grows as new tokens are generated
- Cleanup runs on 5-minute interval but only when middleware is active
- In low-traffic periods, cleanup may not run frequently enough
- No maximum size limit on token map

**Impact:**
- Memory usage grows unbounded in long-running servers
- Tokens never expire if cleanup goroutine isn't running
- No eviction policy if token storage exhausts memory

**Mitigation needed:**
- Add maximum token store size with LRU eviction
- Track cleanup success metrics
- Consider storing tokens in external session store (Redis) for production

---

### Session Store Memory Growth

**Issue:** `MemorySessionStore` (referenced in `core/auth.go`) accumulates expired sessions until cleanup runs.

**Files:** `core/auth.go` (lines 140-165)

**Problem:**
- Sessions are only deleted during periodic cleanup (5-minute interval)
- High-traffic servers can accumulate many expired sessions between cleanups
- No bounds on concurrent sessions

**Impact:**
- Memory usage grows with session count
- Cleanup delay means stale data in memory
- Production deployments need external session store

**Recommendations:**
- `MemorySessionStore` is appropriate for development/testing only
- Document that production deployments must provide custom `SessionStore` implementation (e.g., Redis-backed)
- Add warning in documentation about memory implications

---

## Input Validation & Security

### Unsafe JSON Unmarshaling in CRUD Operations

**Issue:** CRUD endpoints decode untrusted JSON without size limits or validation.

**Files:** `core/crud.go` (lines 561-578, 669-705)

**Problem:**
```go
var data map[string]interface{}
if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
    http.Error(w, "Invalid JSON", http.StatusBadRequest)
    return
}
```

**Impact:**
- No request body size limit: attackers can send multi-GB payloads
- No field validation: arbitrary fields can be injected
- No type coercion validation: type assertions in hooks could panic

**Missing mitigations:**
- `http.MaxBytesReader` wrapping request body
- Field whitelist validation before hook execution
- Type validation for critical fields

---

### Query Parameter Filter Type Coercion Without Bounds

**Issue:** Query parameter filtering applies type coercion without validating input ranges.

**Files:** `core/crud.go` (lines 872-907)

**Problem:**
```go
case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
    if v, err := strconv.ParseUint(rawVal, 10, 64); err == nil {
        val = v
    }
```

**Impact:**
- Large uint64 values could bypass validation for smaller uint types
- Float32/64 parsing allows precision loss attacks
- No validation that coerced values match field constraints

**Recommendations:**
- Validate coerced values fit in target type's bit width
- Document type coercion behavior
- Consider rejecting parameters with conflicting types

---

## Complexity & Maintainability

### Monolithic Build Command (4850 lines)

**Issue:** `cmd/gux/build.go` is extremely large and handles too many responsibilities.

**Files:** `cmd/gux/build.go`

**Problem:**
- CRUD parsing, API code generation, WASM building, asset handling all in one file
- Deep nesting and complex control flow make it difficult to modify
- Testing specific functionality requires understanding entire build pipeline

**Impact:**
- Bug fixes risk breaking unrelated features
- Hard for new contributors to understand
- Difficult to parallelize build steps
- Code generation changes require navigating complex logic

**Refactoring approach:**
- Extract CRUD parsing to separate module
- Separate API code generation into dedicated generator
- Create clear interfaces between build phases

---

### ModelGen Template Field Evolution Incomplete

**Issue:** Many-to-many relationship initialization marked incomplete with TODO.

**Files:** `cmd/gux/modelgen.go` (line 2230)

**Problem:**
```go
tf.EditStateInit = `"[]"`    // TODO: Initialize from displayItem for M2M
```

**Impact:**
- M2M edit forms initialize with empty array, not existing relationships
- Users must manually populate M2M selections in edit forms
- Could lead to unintended data loss if form is submitted without checking values

**Recommendations:**
- Complete M2M initialization from `displayItem`
- Add test coverage for M2M edit forms
- Document workaround for current behavior

---

### Client-Side Hook Model Coverage Gaps

**Issue:** Admin form hooks in `admin/hooks_gen.go` have placeholder implementations in example.

**Files:** `examples/minimal/admin/client_hooks.go` (lines 22, 50)

**Problem:**
```go
// TODO: Implement export logic
// TODO: Open email dialog
```

**Impact:**
- Hook pattern not fully demonstrated
- Users unclear how to implement custom actions
- Example may suggest hooks are unfinished feature

**Recommendations:**
- Complete example implementations or remove TODO markers
- Add comprehensive hook documentation with real use cases

---

## Memory Leak Risks

### Event Handler Reference Tracking (Potential Issue)

**Issue:** DOM renderer creates `js.FuncOf` callbacks without explicit cleanup tracking.

**Files:** `core/dom_renderer.go` (referenced in PITFALLS.md)

**Concern:** Each render creates `js.Func` values that must be released to avoid memory leaks in WASM.

**Current Status:** Architecture uses `OnMount`/`OnUnmount` pattern which can mitigate this, but needs verification that all event handlers are properly cleaned up on unmount.

**Verification needed:**
- Audit `dom_renderer.go` to ensure all `js.FuncOf` callbacks are released
- Add memory leak tests for long-running SPA navigation
- Document WASM memory management patterns for component authors

---

### Large Data Set Reflection Performance

**Issue:** DTO conversion uses reflection in render-critical paths.

**Files:** `core/crud.go` (lines 439-475)

**Problem:**
```go
func (a *App) convertToDTO(models interface{}, dtoType reflect.Type) interface{} {
    // Reflection-heavy conversion
}
```

**Impact:**
- Rendering large lists (100+ items) becomes slow due to reflection overhead
- WASM reflection is 10-100x slower than native code
- Cumulative effect on page load performance

**Recommendations:**
- Generate DTO conversion code instead of using reflection
- Cache `reflect.Type` lookups
- Consider code generation for hot paths

---

## Known Limitations & Workarounds

### WASM Memory Not Returned to OS

**Status:** Known Go limitation (Issue #59061)

**Impact:**
- Long-running SPAs accumulate memory
- Memory freed by GC isn't returned to browser
- Very long sessions may require page reload

**Workaround:** Document that users may need periodic page reloads for extended sessions.

---

### Go WASM Binary Size

**Status:** Known but mitigated

**Files:** Build tooling supports TinyGo and WASM bundling

**Mitigation:**
- TinyGo reduces binaries 10x
- Multiple bundles via `WithBundle()` already implemented
- Consider documenting size optimization best practices

---

## Test Coverage Gaps

### M2M Relationship Edit Forms

**Issue:** Many-to-many edit form initialization is incomplete (TODO at line 2230).

**Impact:** No test coverage for M2M edit forms serializing existing relationships.

**Recommendation:** Add test case for M2M relationship display and modification in edit forms.

---

### CSRF Token Cleanup Reliability

**Issue:** Token cleanup runs on timer but isn't guaranteed in test environments.

**Impact:** Tests may be unreliable if cleanup is expected but doesn't run.

**Recommendation:** Add test coverage for token store cleanup behavior and max size limits.

---

### Query Filter Coercion Edge Cases

**Issue:** Type coercion for query parameters lacks comprehensive tests for boundary cases.

**Impact:** Unexpected behavior with extreme values (max int/uint, special floats like Inf/NaN).

**Recommendation:** Add table-driven tests for all coercion paths with boundary values.

---

## Dependencies & Versions

### Build Tool Fragility

**Issue:** Build system depends on external tools (Go, TinyGo, npm/bun for CSS).

**Files:** `cmd/gux/build.go` (toolchain detection and invocation)

**Risk:**
- Tool version mismatches cause cryptic failures
- No lockfile for tool versions
- Development environment setup requires multiple tools

**Recommendations:**
- Document exact tool versions in README
- Consider Docker-based build environment
- Add tool version validation to build process

---

## Performance Concerns

### CRUD List Endpoint No Pagination

**Issue:** CRUD list endpoint returns all records without pagination support.

**Files:** `core/crud.go` (list handler around line 300-410)

**Impact:**
- Serializing 10,000+ item lists is slow
- Network bandwidth wasted sending all results to client
- Client-side rendering of large lists causes frame drops

**Recommendations:**
- Add built-in pagination support (offset/limit or cursor-based)
- Document pagination patterns for large datasets
- Add example showing pagination usage

---

### Build Command Sequential Processing

**Issue:** WASM bundle building processes bundles sequentially.

**Impact:** Multi-bundle apps have slow build times.

**Recommendation:** Parallelize WASM bundle compilation.

---

## Security Hardening Opportunities

### CSRF Configuration Default to Insecure

**Issue:** Default CSRF config sets `Secure: false`.

**Files:** `core/csrf.go` (line 54)

**Impact:** Development default is insecure; must remember to enable in production.

**Recommendation:**
- Auto-detect `Secure: true` when environment is production
- Add build-time warning if `Secure: false` in production binary
- Provide environment variable override

---

### Session Cookie SameSite Default

**Issue:** Session cookies default to `SameSiteStrictMode` which may be too restrictive for cross-origin requests.

**Files:** `core/auth.go` (defaults)

**Consider:** Document tradeoffs and provide examples for relaxing SameSite when needed.

---

### API Error Messages Leak Implementation Details

**Issue:** Some error messages return raw database or parsing errors.

**Files:** `core/crud.go` (various error responses)

**Risk:** Error messages could reveal schema structure or internal implementation.

**Recommendation:** Sanitize error messages before returning to client. Return generic errors for database/parsing failures.

---

## Scalability Concerns

### Single Server Architecture Limitations

**Issue:** Session and CSRF token stores are in-memory only (for default MemorySessionStore).

**Impact:**
- Not suitable for multi-server deployments
- No session persistence across restarts
- Horizontal scaling requires sticky sessions or external store

**Recommendation:** Document that production requires external session store (Redis, database) via custom `SessionStore` implementation.

---

## Documentation & Future Considerations

### M2M Relationship Handling Unclear

**Status:** M2M edit form state initialization is incomplete with TODO marker.

**Recommendation:** Either complete the implementation or clearly document the current limitation and workaround.

---

### Audit Logging Performance Impact Undocumented

**Issue:** Audit logging includes async snapshot diffing but performance impact not quantified.

**Files:** `core/audit_test.go` and implementation in `crud.go`

**Recommendation:** Add performance benchmarks and document when to disable audit logging.

---

## Summary of Priorities

**HIGH Priority (Blocking production use):**
- Request body size limits for CRUD endpoints (security)
- CSRF/Session token store bounds (memory safety)
- Session store external implementation requirement (scalability)

**MEDIUM Priority (Quality/maintainability):**
- Ticker cleanup and graceful shutdown (resource management)
- Build command refactoring (maintainability)
- Query filter type coercion validation (robustness)
- M2M edit form initialization (completeness)

**LOW Priority (Nice-to-have improvements):**
- Reflect performance optimization (performance)
- Build parallelization (build speed)
- Error message sanitization (security hardening)
- Pagination support in CRUD (usability)

---

*Concerns audit: 2026-02-01*
