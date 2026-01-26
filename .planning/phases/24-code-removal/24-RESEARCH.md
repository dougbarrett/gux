# Phase 24: Code Removal - Research

**Researched:** 2026-01-25
**Domain:** Go dead code removal and verification
**Confidence:** HIGH

## Summary

This phase removes five dead packages (`components/`, `auth/`, `storage/`, `state/`, `ws/`) that are no longer used in the v2.0 architecture. Research confirms these packages are already orphaned - no Go files in the main module or example apps import them. The only internal reference is `components/connection_status.go` importing `state/`, which is moot since both packages are being deleted together.

The standard approach uses Go's built-in tools (`go build`, `go vet`) combined with `staticcheck` for comprehensive verification. The official `deadcode` tool from `golang.org/x/tools` provides additional confidence but is optional since simple grep-based import verification is sufficient for this scope.

**Primary recommendation:** Delete packages in any order, then verify with `go build ./...` and `staticcheck ./...`. No complex dependency analysis needed since all packages are fully orphaned.

## Standard Stack

The established tools for Go dead code removal and verification:

### Core
| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| `go build` | 1.24.3 | Compilation verification | Fails on missing imports, official toolchain |
| `go vet` | 1.24.3 | Static analysis | Detects unused variables, official toolchain |
| `staticcheck` | 0.6.1 (2025.1.1) | Extended static analysis | Detects unused code (U1000), industry standard |

### Supporting
| Tool | Version | Purpose | When to Use |
|------|---------|---------|-------------|
| `deadcode` | latest | Unreachable function detection | Post-removal confidence check (optional) |
| `golangci-lint` | 2.8.0 | Meta-linter | Alternative to separate staticcheck (includes U1000) |
| `rm -rf` | system | Directory removal | Standard UNIX deletion |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| staticcheck | golangci-lint | golangci-lint is heavier but includes more checks |
| deadcode | grep | grep is simpler for import verification; deadcode adds function-level analysis |

**Verification commands:**
```bash
# Primary verification (REQUIRED)
go build ./...
staticcheck ./...

# Secondary verification (OPTIONAL, for confidence)
deadcode -test ./...
```

## Architecture Patterns

### Deletion Order

Any order is valid since packages are fully orphaned. Recommended order by size (largest first) for clear progress:

```
1. components/  (59 .go files, 61 total) - largest, provides biggest impact
2. state/       (4 .go files)           - referenced only by components
3. ws/          (1 .go file)            - standalone
4. auth/        (1 .go file)            - standalone
5. storage/     (1 .go file)            - standalone
```

### Verification Pattern
**What:** Build-then-lint verification after each major deletion
**When to use:** After deleting each package or batch

```bash
# After each deletion
go build ./... && staticcheck ./...

# Final comprehensive check
go build ./... && go vet ./... && staticcheck ./...
```

### Safe Deletion Pattern
**What:** Verify no imports, delete, verify build
**Example:**
```bash
# Step 1: Verify no imports exist
grep -r "github.com/dougbarrett/gux/components" --include="*.go" .

# Step 2: Delete directory
rm -rf components/

# Step 3: Verify build still works
go build ./...
```

### Anti-Patterns to Avoid
- **Deleting without verification:** Always run `go build ./...` after removal
- **Ignoring example apps:** Verify examples still build, not just main module
- **Skipping staticcheck:** `go build` won't catch all dead code issues

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Import detection | Custom AST parser | `grep` or `staticcheck` | Standard tools are reliable and fast |
| Dead code detection | Custom analysis | `deadcode` tool | Official Go team tool, handles edge cases |
| Multi-package verification | Manual file inspection | `go build ./...` | Recursive build catches all import errors |

**Key insight:** Go's toolchain provides everything needed. The compiler itself is the ultimate verification - if it builds, imports are resolved.

## Common Pitfalls

### Pitfall 1: Forgetting Test Files
**What goes wrong:** Delete package, tests still import it, CI fails
**Why it happens:** Test files may import packages differently than main code
**How to avoid:** Use `go build ./...` which includes test compilation
**Warning signs:** Build passes locally but CI fails

### Pitfall 2: Missing Example Apps
**What goes wrong:** Main module builds but example apps fail
**Why it happens:** Example apps are separate binaries with different entry points
**How to avoid:** Verify from project root: `go build ./examples/...`
**Warning signs:** Partial build commands, not full `./...`

### Pitfall 3: Internal Cross-References
**What goes wrong:** Package A imports Package B, both being deleted, deletion order matters
**Why it happens:** Order-dependent removal when packages reference each other
**How to avoid:** Delete together in single operation, or delete importer first
**Warning signs:** `components/` imports `state/` - but both are dead, so no issue

### Pitfall 4: Build Tags / Conditional Compilation
**What goes wrong:** Package appears unused but is conditionally compiled
**Why it happens:** `//go:build` tags hide imports from standard grep
**How to avoid:** Check for build tags in dead packages
**Warning signs:** WASM-specific code (js.Value usage) - `components/` is WASM-only

### Pitfall 5: Documentation References (Out of Scope)
**What goes wrong:** Docs reference deleted packages
**Why it happens:** Code deletion without doc updates
**How to avoid:** Phase 25 handles this - don't worry about docs in Phase 24
**Warning signs:** N/A - explicitly deferred to next phase

## Code Examples

### Verifying No Imports Exist
```bash
# Check all packages at once
grep -rE "github\.com/dougbarrett/gux/(components|auth|storage|state|ws)" \
  --include="*.go" .

# Expected output: Only components/connection_status.go importing state/
# (This internal reference is moot - both packages are being deleted)
```

### Safe Batch Deletion
```bash
# Delete all dead packages in one operation
rm -rf components/ auth/ storage/ state/ ws/

# Verify everything still builds
go build ./...
```

### Full Verification Suite
```bash
# Complete verification after deletion
go build ./... && \
go vet ./... && \
staticcheck ./...

# Optional: Check for any remaining dead code
deadcode -test ./...
```

### Verifying Example Apps
```bash
# Each example app should build
go build ./examples/admin/...
go build ./examples/auth/...
go build ./examples/marketing/...
go build ./examples/minimal/...
go build ./examples/saas/...

# Or all at once
go build ./examples/...
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual grep for imports | `deadcode` tool | Go 1.22 (2024) | Automated dead code detection |
| `golint` (deprecated) | `staticcheck` | 2020 | More comprehensive analysis |
| Individual tool runs | `golangci-lint` | 2018 | Single meta-linter option |

**Current in this codebase:**
- Go 1.24.3 (latest stable)
- staticcheck 0.6.1 (2025.1.1)
- golangci-lint 2.8.0
- deadcode (latest from golang.org/x/tools)

All tools are current and available on this system.

## Codebase Analysis

### Packages to Remove

| Package | Files | Description | Status |
|---------|-------|-------------|--------|
| `components/` | 59 .go | Old js.Value component library | Orphaned, no imports |
| `auth/` | 1 .go | JWT/localStorage auth | Orphaned, no imports |
| `storage/` | 1 .go | Local storage wrapper | Orphaned, no imports |
| `state/` | 4 .go | Old state management | Orphaned (only imported by components/) |
| `ws/` | 1 .go | WebSocket package | Orphaned, no imports |

**Total:** 66 Go files to remove

### Current Import Analysis

```
grep results for dead package imports:
- components/: No external imports found
- auth/: No imports found
- storage/: No imports found
- state/: Only imported by components/connection_status.go (internal)
- ws/: No imports found
```

All packages are fully orphaned from the active codebase.

### Build Verification Baseline

Current build status:
```
go build ./...     -> Warning: examples/minimal/assets_gen.go missing .wasm file (acceptable)
go vet ./...       -> Same warning (acceptable)
staticcheck ./...  -> Various U1000 warnings in other packages (pre-existing, out of scope)
```

Post-deletion, same warnings should persist (no new errors).

## Open Questions

None. Research is complete with high confidence:

1. All dead packages are verified orphaned
2. Tools are installed and working
3. Verification strategy is clear
4. No complex dependencies to resolve

## Sources

### Primary (HIGH confidence)
- Local codebase analysis via `grep`, `go build`, `staticcheck`, `deadcode`
- [Go Blog: Finding unreachable functions with deadcode](https://go.dev/blog/deadcode) - Official Go team documentation
- [Staticcheck Documentation](https://staticcheck.dev/docs/) - Official staticcheck docs
- [go vet Documentation](https://pkg.go.dev/cmd/vet) - Official Go documentation

### Secondary (MEDIUM confidence)
- [DEV.to: Eliminating dead code in Go projects](https://dev.to/mfbmina/eliminating-dead-code-in-go-projects-1glc) - Community patterns
- [Medium: Banishing the Ghosts in Your Go Code](https://medium.com/@0xgotznit/banishing-the-ghosts-in-your-go-code-a-friendly-guide-to-eliminating-dead-code-bfd8f3b0984f) - Best practices

### Tertiary (LOW confidence)
- None needed - all claims verified with local tooling

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Tools verified installed and working on this codebase
- Architecture: HIGH - Deletion order verified through import analysis
- Pitfalls: HIGH - Based on Go's well-documented build system behavior

**Research date:** 2026-01-25
**Valid until:** Indefinite - Go's build system is stable; tools may update but patterns remain
