# Codebase Concerns

**Analysis Date:** 2026-01-14

## Tech Debt

**Ignored JSON Marshaling Errors:**
- Issue: JSON marshal error explicitly ignored, could corrupt auth state
- File: `auth/auth.go` (line ~181)
- Code: `data, _ := json.Marshal(a.state)`
- Impact: If marshaling fails, empty data stored in localStorage
- Fix approach: Handle error gracefully or log it

**Silent WebSocket Send Failures:**
- Issue: Returns nil when WebSocket not connected instead of error
- File: `state/websocket.go` (lines 185-189)
- Impact: Callers don't know if message was sent
- Fix approach: Return `ErrNotConnected` error

**Large Files Needing Refactoring:**
- Issue: Several component files exceed 400 lines
- Files:
  - `components/formbuilder.go` (734 lines)
  - `components/theme.go` (589 lines)
  - `components/charts.go` (502 lines)
  - `components/inspector.go` (483 lines)
  - `components/fileupload.go` (428 lines)
- Impact: Harder to maintain and understand
- Fix approach: Split by concern (e.g., separate chart types)

## Known Bugs

**Silent Auth State Corruption:**
- Symptoms: Auth state may not load correctly after browser restart
- Trigger: Corrupted JSON in localStorage
- File: `auth/auth.go` (lines 192-194)
- Workaround: Clear localStorage
- Root cause: Silent failure on JSON unmarshal error

## Security Considerations

**CORS Allows All Origins by Default:**
- Risk: Cross-origin requests from any domain accepted
- File: `server/middleware.go` (lines 36-38)
- Current mitigation: None (default is `*`)
- Recommendations: Require explicit origin configuration

**No JWT Signature Validation:**
- Risk: Token could be forged
- File: `auth/auth.go` (lines 205-228)
- Current mitigation: None (explicitly noted as "without verification")
- Recommendations: Add signature validation before trusting token claims

**Unvalidated localStorage Access:**
- Risk: Could panic in private browsing mode
- File: `auth/auth.go` (lines 96, 161-174)
- Current mitigation: Null check on value only
- Recommendations: Check if localStorage API exists before calling

## Performance Bottlenecks

**Blocking Fetch Implementation:**
- Problem: `fetch.Fetch()` blocks goroutine waiting on channel
- File: `fetch/fetch.go` (lines 35-106)
- Impact: Can freeze UI in WASM if misused
- Improvement path: Document async patterns, consider callback API

**Sequential Message Handler Dispatch:**
- Problem: WebSocket handlers called sequentially, not concurrently
- File: `state/websocket.go` (lines 144-152)
- Impact: Slow handlers block message processing
- Improvement path: Option for concurrent dispatch

## Fragile Areas

**WebSocket JS Function Cleanup:**
- Why fragile: JS callbacks only released in `Close()` method
- File: `ws/ws.go` (lines 244-247)
- Common failures: Memory leak if Close() not called
- Safe modification: Use defer or sync.Once for cleanup
- Test coverage: None

**Store Subscription Index Tracking:**
- Why fragile: Index updating logic assumes stable indices
- File: `state/store.go` (lines 68-87)
- Common failures: Concurrent unsubscribes could corrupt list
- Safe modification: Add integration tests before changes
- Test coverage: None

## Missing Critical Features

**No Test Suite:**
- Problem: Zero test files in codebase
- Files affected: All packages
- Current workaround: Manual testing only
- Blocks: Safe refactoring, regression detection
- Implementation complexity: Medium (need WASM test harness)

**No CI/CD Pipeline:**
- Problem: No automated testing or deployment
- Current workaround: Manual builds and deploys
- Blocks: Reliable releases, contributor confidence
- Implementation complexity: Low (GitHub Actions)

## Test Coverage Gaps

**All Modules Untested:**
- What's not tested: Entire codebase (0 test files)
- Risk: Any change could break functionality undetected
- Priority: Critical
- Key areas needing tests:
  - `state/store.go` - Reactive subscriptions
  - `auth/auth.go` - Token handling
  - `ws/ws.go` - WebSocket connection lifecycle
  - `cmd/apigen/main.go` - Code generation

## Dependencies at Risk

**go.mod Version Warning:**
- Risk: Indirect dependency noted in IDE warning
- File: `go.mod` (line 5)
- Warning: `github.com/gorilla/websocket should be direct`
- Impact: Minor (go mod tidy recommended)
- Migration plan: Run `go mod tidy`

## Documentation Gaps

**Complex Components Lack Internal Docs:**
- Files:
  - `components/combobox.go` - State management not explained
  - `components/animation.go` - Timing logic not documented
  - `components/virtuallist.go` - Virtualization strategy unclear
- Impact: Harder for contributors to understand
- Fix: Add implementation comments

---

*Concerns audit: 2026-01-14*
*Update as issues are fixed or new ones discovered*
