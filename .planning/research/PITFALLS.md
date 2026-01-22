# Pitfalls Research: Go/WASM Component Library with SSR + Hydration

**Domain:** Go/WASM component library, SSR + hydration, example applications
**Researched:** 2026-01-22
**Confidence:** MEDIUM-HIGH (based on codebase analysis + verified web research)

## Component Library Pitfalls

### CL-1: Props API That Fights Go's Type System

**What goes wrong:** Porting React/Vue component patterns directly to Go creates awkward APIs. React components accept `any` props freely; Go requires explicit types. Developers try to use `map[string]interface{}` or `interface{}` everywhere, losing type safety.

**Why it happens:** Go's type system is fundamentally different from JavaScript. React's `props` object pattern doesn't translate directly.

**Consequences:**
- Runtime errors instead of compile-time errors
- IDE autocomplete breaks
- Documentation becomes essential (not optional)
- Reflection overhead for dynamic prop handling

**Warning signs:**
- Frequent use of `interface{}` in component signatures
- Type assertions scattered throughout component code
- Components accepting `map[string]interface{}` for configuration

**Prevention:**
- Use explicit Props structs for each component (like `ButtonProps`, `InputProps`)
- Leverage Go generics where appropriate (e.g., `AsyncStore[T]`)
- Design APIs around Go idioms, not React idioms
- Your existing `core.Attrs` pattern is correct - keep it explicit

**Phase to address:** Phase 1 (Component Architecture) - establish patterns before building components

**Codebase evidence:** The existing `ButtonProps`, `FormProps` patterns are good. Maintain this approach.

---

### CL-2: Variant Explosion Without Composition

**What goes wrong:** Creating separate functions for every variant combination: `PrimaryButton`, `SecondaryButton`, `PrimarySmallButton`, `SecondaryLargeRoundedButton`. This doesn't scale.

**Why it happens:** In React, variants often use string props. Developers port this as separate functions instead of composable options.

**Consequences:**
- Combinatorial explosion of component variants
- Inconsistent naming across components
- Difficult to extend or customize

**Warning signs:**
- More than 3-4 convenience constructors per component
- Variant combinations need new functions
- Users can't combine options

**Prevention:**
- Use option pattern: `Button(ButtonProps{Variant: ButtonPrimary, Size: ButtonLG})`
- Provide convenience constructors only for the most common cases
- Keep variant enums (like `ButtonVariant`, `ButtonSize`) consistent across components

**Phase to address:** Phase 2 (Core Components) - design component APIs with composition in mind

**Codebase evidence:** Current `Button` implementation already follows this pattern with `ButtonVariant` and `ButtonSize` enums.

---

### CL-3: Build Tag Splitting Creates API Divergence

**What goes wrong:** Components have `//go:build js && wasm` versions that diverge from server-side implementations. Over time, the two implementations drift, causing hydration mismatches.

**Why it happens:** Different implementations for different build targets. Go's build tag system makes this easy but doesn't enforce consistency.

**Consequences:**
- Components behave differently server vs client
- Hydration mismatches
- Hard-to-debug production issues
- Double maintenance burden

**Warning signs:**
- `_wasm.go` files with significantly different logic from base files
- Feature parity tracking needed between builds
- Manual testing required for both environments

**Prevention:**
- Minimize build-tag-split code; keep core logic in shared files
- Use interfaces to abstract platform differences (like `Renderer`)
- Your current `core.Node` -> `HTMLRenderer`/`DOMRenderer` pattern is correct
- Only split at the true platform boundary (DOM access)

**Phase to address:** Phase 1 (Component Architecture) - establish dual-render patterns

**Codebase evidence:** `core/html_renderer.go` and `core/dom_renderer.go` show the correct split - only the rendering mechanism differs, not the component logic.

---

### CL-4: Event Handler Leaks in WASM

**What goes wrong:** Each `js.FuncOf` callback creates a Go function wrapper that must be explicitly released. Components that don't call `Func.Release()` leak memory.

**Why it happens:** Go WASM's `js.FuncOf` returns a function that pins Go memory until released. This is documented but easy to forget.

**Consequences:**
- Memory grows continuously with user interaction
- Long-running apps eventually crash
- Affects user experience in SPAs

**Warning signs:**
- `js.FuncOf` without corresponding cleanup
- No component lifecycle/cleanup mechanism
- Memory grows in browser DevTools over time

**Prevention:**
- Track all `js.Func` values created by components
- Implement component cleanup that releases functions
- Document that event handlers should be reused when possible
- Consider a component lifecycle hook for cleanup

**Phase to address:** Phase 2 (Core Components) - implement cleanup patterns

**Codebase evidence:** Current `dom_renderer.go` creates `js.FuncOf` callbacks in `RenderElement` without tracking them for cleanup. This needs attention.

**Verified source:** [Go WASM Memory Leak Issue #74342](https://github.com/golang/go/issues/74342) - recent (June 2025) confirmed memory leak when passing JS pointers to Go WASM.

---

## SSR + Hydration Pitfalls

### HY-1: State Serialization Type Mismatch

**What goes wrong:** Server serializes state as JSON, client deserializes. JSON numbers become `float64`, not `int`. Component expects `int`, gets `float64`, breaks.

**Why it happens:** JSON has no integer type. All numbers are float64 when unmarshaled to `interface{}` or `map[string]any`.

**Consequences:**
- Type assertions fail: `val.(int)` panics on `float64`
- State appears corrupted after hydration
- Intermittent failures based on state content

**Warning signs:**
- Type assertions without proper handling
- `panic: interface conversion` errors in WASM console
- State works on client-only navigation, fails on page refresh

**Prevention:**
- Your existing `State[T].Get()` handles this with float64 -> int conversion
- Ensure all state accessors handle JSON number types
- Test hydration with various data types
- Consider explicit serialization types for complex state

**Phase to address:** Phase 1 (Component Architecture) - validate state serialization handling

**Codebase evidence:** `core/app.go` lines 587-607 already handle this conversion. Good pattern to maintain.

---

### HY-2: Server-Only Data in Hydrated State

**What goes wrong:** Server includes data in state that shouldn't reach the client (database connections, internal IDs, session tokens). This data serializes into the HTML.

**Why it happens:** Server state is convenient for data loading. Easy to accidentally include sensitive data.

**Consequences:**
- Security vulnerabilities (exposed tokens, internal data)
- Hydration failures if data isn't JSON-serializable
- Bloated page HTML size

**Warning signs:**
- Large `__gux_state` script tags in HTML
- Database model objects in state instead of DTOs
- `r.DB()` access results stored directly in state

**Prevention:**
- Only store serializable, client-safe data in state
- Use DTOs for data passed to client (like current `UserList`, `UserDetail`)
- Never store `r.DB()` results directly in state
- Review state contents in rendered HTML during development

**Phase to address:** All phases - ongoing discipline

**Codebase evidence:** Current DTO pattern in `dto/user.go` correctly strips sensitive fields like `PasswordHash`. Maintain this pattern.

---

### HY-3: Conditional Rendering Hydration Mismatch

**What goes wrong:** Server renders based on condition A, client hydrates with condition B. DOM doesn't match, hydration fails or causes visual glitches.

**Why it happens:**
- Server-only code (e.g., auth checks) produces different results
- Time-based rendering (timestamps, "posted 5 minutes ago")
- Random values or UUIDs generated during render

**Consequences:**
- Visual flicker as client replaces server-rendered content
- Console warnings/errors about mismatched content
- React/Vue throw explicit errors; Go WASM silently produces wrong DOM

**Warning signs:**
- Content "flashes" on page load
- Server HTML differs from client first render
- Components that use `time.Now()` or random values

**Prevention:**
- Ensure server and client render identical initial content
- Use state hydration for any values that might differ
- Don't use `time.Now()` or random values during render
- Use the `r.OnLoad()` pattern to fetch data consistently

**Phase to address:** Phase 3 (Advanced Components) - add client-only rendering capability

**Verified sources:**
- [Hydration Mismatch | Vike](https://vike.dev/hydration-mismatch)
- [Next.js React Hydration Error](https://nextjs.org/docs/messages/react-hydration-error)

---

### HY-4: Browser API Access During SSR

**What goes wrong:** Component code accesses `window`, `document`, `localStorage` on server where they don't exist. Server panics or renders wrong content.

**Why it happens:** Go WASM components use `syscall/js` which only works in browser. Server-side rendering runs in regular Go.

**Consequences:**
- Server crashes during render
- Build failures if WASM-only code imported
- Different behavior SSR vs client

**Warning signs:**
- `js.Global()` calls in shared code
- `syscall/js` imports in non-WASM-tagged files
- Components that need browser state for rendering

**Prevention:**
- All `syscall/js` code behind `//go:build js && wasm` tags
- Components return `core.Node`, not `js.Value`
- Use `r.OnLoad()` for browser-dependent initialization
- Your existing architecture handles this correctly

**Phase to address:** Phase 1 (Component Architecture) - enforce build tag discipline

**Codebase evidence:** The `core.Node` abstraction correctly isolates browser APIs. Components like `Button` in `components/button.go` should probably use this pattern instead of direct `js.Value`.

---

### HY-5: Re-render During Input Events Destroys Focus

**What goes wrong:** User types in input, state updates, component re-renders, input loses focus. User must click back into input to continue typing.

**Why it happens:** State change triggers re-render, which replaces DOM including the focused input element.

**Consequences:**
- Terrible UX - impossible to type continuously
- Users blame the app for being "broken"
- Forms become unusable

**Warning signs:**
- Input losing focus on every keystroke
- `OnChange` handlers that call `Set()` immediately
- Forms requiring "submit" instead of live updates

**Prevention:**
- Your `SetInChangeEvent` pattern is correct - suppress re-renders during change events
- Use `SetQuiet()` for input values that don't need re-render
- Batch state updates
- Restore focus after necessary re-renders (your code does this)

**Phase to address:** Already addressed in core - document pattern for component authors

**Codebase evidence:** `core/dom_renderer.go` lines 80-92 and `core/router_wasm.go` implement this correctly with `SetInChangeEvent`.

---

## Go/WASM Specific Pitfalls

### GW-1: WASM Binary Size Explosion

**What goes wrong:** Standard Go WASM binaries are 5-15+ MB. Users on slow connections experience long load times. Mobile users may abandon the app.

**Why it happens:** Go runtime, garbage collector, and standard library all compile into WASM. Even "Hello World" is several MB.

**Consequences:**
- Slow initial page loads (especially mobile)
- High bandwidth costs
- Poor Core Web Vitals scores
- User abandonment

**Warning signs:**
- `app.wasm` grows beyond 5MB
- Page load times > 3 seconds on average connections
- Adding features causes disproportionate size increases

**Prevention:**
- Use TinyGo for significantly smaller binaries (often 10x reduction)
- Split into multiple bundles (you already support this with `WithBundle`)
- Implement lazy loading for non-critical routes
- Use Brotli compression (4.6MB -> 32MB typical ratio)
- Monitor WASM size in CI

**Phase to address:** Phase 1 (Component Architecture) - establish size budgets

**Codebase evidence:** `buildWasmBundle` already supports TinyGo. Multiple WASM bundles via `RouteGroup` with `WithBundle` is implemented.

**Verified source:** [Dagger: Replaced React with Go](https://dagger.io/blog/replaced-react-with-go) - discusses their 32MB -> 4.6MB compression journey.

---

### GW-2: WASM Memory Never Released to OS

**What goes wrong:** Go WASM allocates memory but never releases it back to the browser. Memory usage only ever increases.

**Why it happens:** Known Go WASM limitation - memory blocks freed by GC aren't returned to the OS/browser.

**Consequences:**
- Long-running SPAs accumulate memory
- Browser eventually crashes or kills tab
- Users must refresh to reclaim memory

**Warning signs:**
- Memory in DevTools grows over time
- Tab marked as "high memory" by browser
- App becomes sluggish after extended use

**Prevention:**
- Be aware this is a known Go limitation
- Consider page reloads for very long sessions
- Minimize object allocation in render loops
- Reuse allocations where possible
- Monitor memory usage during development

**Phase to address:** Phase 4 (Data Display) - be careful with large data sets

**Verified source:** [Go Issue #59061](https://github.com/golang/go/issues/59061) - WASM does not return memory to the OS.

---

### GW-3: Goroutine Blocking Freezes Browser

**What goes wrong:** A goroutine blocks (channel wait, network call without callback pattern), the browser's event loop freezes, and the page becomes unresponsive.

**Why it happens:** Go WASM runs on the browser's main thread. Blocking operations block the entire thread.

**Consequences:**
- Page freezes during data loading
- Browser "unresponsive page" warnings
- Poor user experience

**Warning signs:**
- Page freezes during network requests
- Synchronous patterns from server-side Go code
- Missing callback patterns in async operations

**Prevention:**
- All async operations must use callback patterns
- Never block waiting for channels from user code
- Your `AsyncStore[T]` and `api.X.List(callback)` patterns are correct
- Always use goroutines for I/O: `go func() { ... callback() }()`

**Phase to address:** Phase 2 (Core Components) - enforce async patterns

**Codebase evidence:** `state/async.go` and generated API client use correct callback patterns.

**Verified source:** [syscall/js documentation](https://pkg.go.dev/syscall/js) - "if one wrapped function blocks, JavaScript's event loop is blocked."

---

### GW-4: wasm_exec.js Version Mismatch

**What goes wrong:** The `wasm_exec.js` file doesn't match the Go version used to compile WASM. WASM fails to load with cryptic errors.

**Why it happens:** Go's WASM ABI changes between versions. The runtime shim must match.

**Consequences:**
- WASM fails to instantiate
- Errors like "expected magic word 00 61 73 6d"
- "Imports argument must be present" errors

**Warning signs:**
- WASM works locally but fails in production
- Different Go versions in dev vs CI
- Manual wasm_exec.js copies

**Prevention:**
- Always copy wasm_exec.js from the same Go version used to compile
- Your `copyWasmExec` function does this correctly
- Pin Go version in CI/CD
- Include Go version check in build process

**Phase to address:** Already addressed in build tooling

**Codebase evidence:** `copyWasmExec` in `build.go` correctly sources from `go env GOROOT` or TinyGo.

---

### GW-5: Reflection Performance in WASM

**What goes wrong:** Heavy use of `reflect` package in WASM is significantly slower than native Go. Components using reflection become sluggish.

**Why it happens:** WASM has different performance characteristics than native code. Reflection is already slow; WASM makes it worse.

**Consequences:**
- Slow component rendering
- Laggy UI interactions
- Poor perceived performance

**Warning signs:**
- Components using `reflect.ValueOf`, `reflect.TypeOf` in render paths
- Generic components with heavy type inspection
- DTO conversion in hot paths

**Prevention:**
- Avoid reflection in render-critical code
- Use code generation instead of runtime reflection where possible
- Your DTO conversion in `crud.go` uses reflection - consider generating the mapping code
- Cache reflect results where appropriate

**Phase to address:** Phase 4 (Data Display) - optimize if performance issues arise

**Codebase evidence:** `core/crud.go` uses reflection for DTO mapping. The generated API code in `build.go` partially addresses this.

---

## Example App Pitfalls

### EX-1: Over-Engineered Starter Template

**What goes wrong:** Starter app includes every feature: auth, admin panel, multi-tenant, i18n, payments, dark mode. Users spend more time removing code than building.

**Why it happens:** Desire to showcase framework capabilities. "Look how much you can do!"

**Consequences:**
- Intimidating for beginners
- Hard to understand what's essential
- Features they don't need create maintenance burden
- Difficult to strip down to basics

**Warning signs:**
- Starter has 50+ files
- Multiple "optional" features that are actually required
- README says "just delete what you don't need"

**Prevention:**
- Keep starter apps minimal: one page, one component, one data fetch
- Show the framework, not a finished app
- Put advanced features in separate example repos
- "Remove code" should never be the first step

**Phase to address:** Phase 5 (Example Apps) - resist scope creep

---

### EX-2: Examples That Don't Build

**What goes wrong:** Example code in documentation/starters doesn't compile against current framework version. Missing imports, deprecated APIs, wrong versions.

**Why it happens:** Examples are written once and not updated with framework changes.

**Consequences:**
- New users can't get started
- Loss of trust in framework quality
- Support burden from "it doesn't work" issues

**Warning signs:**
- No CI for example apps
- Version numbers in examples don't match releases
- Copy-paste from examples fails

**Prevention:**
- Include example apps in CI
- Test examples against framework changes
- Use monorepo or submodules to keep examples in sync
- Pin versions explicitly in example go.mod

**Phase to address:** Phase 5 (Example Apps) - implement from start

---

### EX-3: Unrealistic Data Patterns

**What goes wrong:** Examples use patterns that don't scale: in-memory arrays, hardcoded data, no pagination. Users copy these patterns into production.

**Why it happens:** Real database setup adds friction. Simpler to show concepts without infrastructure.

**Consequences:**
- Users learn bad patterns
- "Works in example, breaks at scale" issues
- Must unlearn then relearn proper patterns

**Warning signs:**
- `var users = []User{...}` as fake database
- List endpoints returning all records
- No pagination, filtering, or error handling

**Prevention:**
- Use real database (SQLite is fine for examples)
- Show pagination even for small datasets
- Include error handling in all examples
- Document why patterns matter, not just how

**Phase to address:** Phase 5 (Example Apps) - use real patterns from start

**Codebase evidence:** Current minimal example uses real SQLite + GORM. Good pattern to continue.

---

### EX-4: Missing "Why" in Examples

**What goes wrong:** Examples show *what* to do but not *why*. Users copy code without understanding, then can't adapt it.

**Why it happens:** Tutorial writers focus on steps, assume context is obvious.

**Consequences:**
- Cargo culting of patterns
- Can't debug when things go wrong
- Can't adapt examples to real needs

**Warning signs:**
- Code comments describe *what*, not *why*
- No explanation of architectural decisions
- Users ask "why do I need this?" in issues

**Prevention:**
- Add comments explaining intent, not just syntax
- Include "why" sections in READMEs
- Show alternative approaches and tradeoffs
- Link to deeper documentation for concepts

**Phase to address:** Phase 5 (Example Apps) - documentation quality

---

### EX-5: Framework Lock-in in Examples

**What goes wrong:** Examples demonstrate framework-specific patterns that don't represent good Go patterns. Users learn "Gux Go" not "Go."

**Why it happens:** Framework authors focus on their APIs, forget idiomatic Go.

**Consequences:**
- Skills don't transfer to other projects
- Code becomes coupled to framework internals
- Framework changes break user code

**Warning signs:**
- Core Go patterns replaced by framework abstractions
- "You must use X, Y, Z" instead of showing options
- No escape hatches to standard library

**Prevention:**
- Show idiomatic Go patterns, enhanced by framework
- Don't hide standard library behind abstractions
- Allow opt-out of framework features
- Examples should teach Go, not just the framework

**Phase to address:** Phase 5 (Example Apps) - design principle

---

## Prevention Strategies

### Strategy 1: Build Tag Discipline

**Applies to:** CL-3, HY-4

Create clear boundaries:
```
// Components return core.Node - no build tags needed
func MyComponent(props Props) core.Node { ... }

// Only platform code needs build tags
//go:build js && wasm
func attachEventHandlers(el js.Value) { ... }
```

### Strategy 2: State Type Safety

**Applies to:** HY-1, HY-2

Validate state roundtrips:
```go
// Test that state survives JSON serialization
func TestStateRoundtrip(t *testing.T) {
    original := MyState{Count: 42, Name: "test"}
    json, _ := json.Marshal(original)
    var restored MyState
    json.Unmarshal(json, &restored)
    assert.Equal(t, original, restored)
}
```

### Strategy 3: WASM Resource Cleanup

**Applies to:** CL-4, GW-2

Track and release resources:
```go
type Component struct {
    callbacks []js.Func
}

func (c *Component) OnClick(fn func()) {
    cb := js.FuncOf(func(this js.Value, args []js.Value) any {
        fn()
        return nil
    })
    c.callbacks = append(c.callbacks, cb)
    // ... attach to element
}

func (c *Component) Cleanup() {
    for _, cb := range c.callbacks {
        cb.Release()
    }
}
```

### Strategy 4: WASM Size Budget

**Applies to:** GW-1

Add to CI:
```bash
# Check WASM size doesn't exceed budget
max_size=5000000  # 5MB
actual_size=$(stat -f%z .gux/dist/app.wasm)
if [ $actual_size -gt $max_size ]; then
    echo "WASM size $actual_size exceeds budget $max_size"
    exit 1
fi
```

### Strategy 5: Example CI Validation

**Applies to:** EX-2

Include examples in CI:
```yaml
- name: Build example apps
  run: |
    cd examples/minimal && gux build
    cd examples/marketing && gux build
    cd examples/admin && gux build
```

---

## Phase-Specific Warnings

| Phase | Pitfalls to Watch | Primary Risk |
|-------|-------------------|--------------|
| Component Architecture | CL-1, CL-3, HY-4, GW-1 | Wrong patterns get baked in |
| Core Components | CL-2, CL-4, GW-3 | Inconsistent APIs, memory leaks |
| Advanced Components | HY-3, GW-5 | Performance issues, complexity |
| Data Display | GW-2, GW-5 | Memory growth, slow rendering |
| Example Apps | EX-1 through EX-5 | Poor first impressions |

---

## Sources

**Verified (HIGH confidence):**
- [Go WASM Memory Leak Issue #74342](https://github.com/golang/go/issues/74342)
- [Go Issue #59061 - WASM memory not returned](https://github.com/golang/go/issues/59061)
- [syscall/js Package Documentation](https://pkg.go.dev/syscall/js)
- [Dagger: Replaced React with Go](https://dagger.io/blog/replaced-react-with-go)
- [Vike Hydration Mismatch Documentation](https://vike.dev/hydration-mismatch)
- [Next.js Hydration Error Documentation](https://nextjs.org/docs/messages/react-hydration-error)

**Codebase analysis (HIGH confidence):**
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/node.go` - Node abstraction pattern
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/dom_renderer.go` - Event handling, change event suppression
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/app.go` - State management, hydration
- `/Users/dougbarrett/projects/dbb1dev/goquery/cmd/gux/build.go` - Build tooling, bundle support

**Web research (MEDIUM confidence):**
- [Go's Type System: Interfaces vs Generics](https://medium.com/@speedcraft21/gos-type-system-for-humans-interfaces-vs-generics-98f033a1f7ad)
- [Five Challenges with WebAssembly, Go, TypeScript](https://doray.me/articles/five-challenges-when-using-webassembly-golang-typescript-h5lA2/)
- [Component Library Mistakes](https://www.sencha.com/blog/top-mistakes-developers-make-when-using-react-ui-component-library-and-how-to-avoid-them/)
