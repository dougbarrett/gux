# Phase 26: Dependency Cleanup - Research

**Researched:** 2026-01-26
**Domain:** Go module dependency management and build verification
**Confidence:** HIGH

## Summary

Dependency cleanup in Go involves running `go mod tidy` to synchronize go.mod with actual code imports, removing unused dependencies, and verifying that all example applications build successfully. The research revealed that the main module currently has one unused dependency (gorilla/websocket) that needs to be removed, and that the project follows a monorepo structure where examples/ are subdirectories that share the parent module's dependencies rather than having their own go.mod files.

Key findings:
- `go mod tidy` removes unused dependencies and promotes indirect dependencies to direct when actually used
- The gorilla/websocket package is currently declared as a direct dependency but is not imported anywhere in the main module
- All four example applications (minimal, auth, marketing, admin, saas) build successfully via `gux build`
- Examples rely on parent module dependencies (gorm, sqlite driver, crypto, fsnotify)

**Primary recommendation:** Run `go mod tidy` to clean dependencies and verify all examples build via `gux build` (not `go build` directly, as examples need WASM compilation).

## Standard Stack

The established tools for Go dependency management:

### Core

| Tool | Version | Purpose | Why Standard |
|------|---------|---------|--------------|
| go mod tidy | Built-in (Go 1.24.3) | Synchronize go.mod with source code | Official Go toolchain command for dependency cleanup |
| go mod verify | Built-in (Go 1.24.3) | Verify dependency checksums | Official tool to ensure dependencies haven't been tampered with |
| go list | Built-in (Go 1.24.3) | Query module information | Built-in tool for inspecting dependencies and build lists |
| gux build | Project CLI | Build WASM + server binary | Project-specific build tool that handles WASM compilation and asset generation |

### Supporting

| Tool | Purpose | When to Use |
|------|---------|-------------|
| go mod why | Explain why a package is needed | Debugging unexpected dependencies |
| go mod graph | Show dependency graph | Understanding transitive dependencies |
| go list -m all | List all dependencies | Auditing dependency versions |
| go list -m -u all | Check for updates | Finding available dependency updates |

**No installation needed:** All commands are built into Go 1.24.3 except `gux build` which is already available in the project.

## Architecture Patterns

### Go Module Structure in Gux

The project follows a **single-module monorepo** pattern:

```
goquery/
├── go.mod                    # Single module file
├── go.sum                    # Checksum database
├── core/                     # Universal rendering framework
├── ui/                       # Component library
├── fetch/                    # WASM HTTP client
├── api/                      # API utilities
├── server/                   # Server middleware
├── cmd/gux/                  # CLI tool (uses fsnotify)
└── examples/                 # Example apps (no separate go.mod files)
    ├── minimal/
    ├── auth/
    ├── marketing/
    ├── admin/
    └── saas/
```

**Key characteristic:** Examples are **subdirectories**, not separate modules. They import from the parent module:
```go
import (
    "github.com/dougbarrett/gux/core"
    "github.com/dougbarrett/gux/ui"
    "gorm.io/gorm"  // From parent's dependencies
)
```

### Pattern 1: Dependency Cleanup Workflow

**What:** Standard process for cleaning up unused dependencies and verifying builds

**When to use:** After removing packages/code, before releases, as part of milestone completion

**Steps:**
```bash
# 1. Clean dependencies
go mod tidy -v

# 2. Verify checksums
go mod verify

# 3. Build examples via gux (not go build directly)
cd examples/minimal && gux build
cd examples/auth && gux build
cd examples/marketing && gux build
cd examples/admin && gux build
cd examples/saas && gux build

# 4. Commit changes
git add go.mod go.sum
git commit -m "chore: clean up dependencies"
```

**Why this sequence:**
- `go mod tidy` first to clean up go.mod
- `go mod verify` to ensure cache integrity
- `gux build` (not `go build`) because examples need WASM compilation + Tailwind CSS generation
- Test all examples to ensure no hidden breakage

### Pattern 2: Current Dependency Usage

**Direct dependencies (after cleanup):**
- `github.com/fsnotify/fsnotify v1.9.0` - Used by cmd/gux for file watching in dev mode
- `golang.org/x/crypto v0.47.0` - Used by examples for password hashing
- `gorm.io/driver/sqlite v1.6.0` - Used by examples for database connectivity
- `gorm.io/gorm v1.31.1` - Used by examples for ORM

**Unused dependency (to be removed):**
- `github.com/gorilla/websocket v1.5.3` - Not imported anywhere, likely left over from previous WebSocket implementation

**Why examples don't have their own go.mod:**
According to Go monorepo best practices, subdirectories can share the parent module's dependencies when they're part of the same logical project. Separate go.mod files are needed when services should have independent versioning or when you want to prevent dependency conflicts between services.

### Anti-Patterns to Avoid

- **Running `go build` on examples directly:** Examples require `gux build` to generate WASM, API clients, and Tailwind CSS
- **Building without assets:** Running `go build` will fail with "pattern .gux/dist/*.wasm: no matching files found" because assets aren't generated
- **Assuming go mod tidy is safe in CI:** Always verify builds after tidy in case indirect dependencies were holding transitive requirements
- **Ignoring -v flag on tidy:** The verbose output shows what was removed, which is valuable for understanding changes

## Don't Hand-Roll

Problems that have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Dependency verification | Custom checksum checking | `go mod verify` | Uses go.sum and checksum database, handles all edge cases |
| Finding unused deps | Manual grep/analysis | `go mod tidy` | Analyzes entire build list including build tags/platforms |
| Example builds | Shell scripts calling go build | `gux build` | Handles WASM compilation, Tailwind CSS, API generation |
| Dependency updates | Manual go.mod editing | `go get package@version` | Updates go.sum correctly, handles transitive dependencies |

**Key insight:** Go's module system is comprehensive and handles cross-platform build requirements automatically. Custom solutions miss edge cases around build tags, GOOS/GOARCH combinations, and transitive dependencies.

## Common Pitfalls

### Pitfall 1: Forgetting to Test Builds After Tidy

**What goes wrong:** `go mod tidy` removes an indirect dependency that was actually needed through a transitive chain

**Why it happens:** You removed code that imported package A, which required package B. Tidy removes A, but B might have been used elsewhere indirectly.

**How to avoid:**
- Always run builds after `go mod tidy`
- Test all example applications, not just the main module
- Use `gux build` for examples (not `go build`) to catch WASM/asset issues

**Warning signs:**
- Build errors mentioning missing packages after tidy
- Tests passing but runtime errors about missing dependencies

### Pitfall 2: Running go build on Examples

**What goes wrong:** Example apps fail to build with "pattern .gux/dist/*.wasm: no matching files found"

**Why it happens:** Examples use `//go:embed` to embed WASM files that must be generated by `gux build` first

**How to avoid:**
- Always use `gux build` for example apps
- Never use `go build .` or `go run .` directly in examples/
- CI/CD should call `gux build` not `go build`

**Warning signs:**
- Embed pattern errors
- Missing assets_gen.go file
- WASM files not found

### Pitfall 3: Treating Examples as Separate Modules

**What goes wrong:** Attempting to create go.mod files in examples/ subdirectories or running `go mod tidy` there

**Why it happens:** Misunderstanding the monorepo structure - examples share the parent module

**How to avoid:**
- Run `go mod tidy` only at the repository root
- Examples import from the parent module path
- All dependencies go in the root go.mod

**Warning signs:**
- Multiple go.mod files in examples/
- Import path confusion between parent and child modules
- Duplicate dependency declarations

### Pitfall 4: Assuming All Dependencies Are Needed

**What goes wrong:** Keeping dependencies in go.mod "just in case" or because they might be used by examples

**Why it happens:** Fear of breaking things or not understanding that `go mod tidy` is safe

**How to avoid:**
- Trust `go mod tidy` - it analyzes all packages including examples
- Use `go mod why <package>` to verify if a package is actually needed
- Test builds after removal to verify nothing breaks

**Warning signs:**
- go.mod growing over time without corresponding code additions
- Dependencies marked as direct when unused
- Running `go mod tidy -v` shows "unused" messages

## Code Examples

Verified patterns from official sources and project structure:

### Clean Dependencies and Verify Builds

```bash
# Source: https://go.dev/ref/mod (go mod tidy section)
# Run at repository root
cd /path/to/goquery

# Step 1: Clean dependencies with verbose output
go mod tidy -v
# Output: "unused github.com/gorilla/websocket" (if unused)

# Step 2: Verify checksums in cache
go mod verify
# Output: "all modules verified." (or reports tampered modules)

# Step 3: Build all examples via gux
for example in examples/*/; do
    echo "Building $example"
    (cd "$example" && gux build)
done
```

### Check Why a Dependency Exists

```bash
# Source: https://go.dev/doc/modules/managing-dependencies
# Query why a package is in the build list
go mod why github.com/fsnotify/fsnotify

# Output shows import chain:
# github.com/dougbarrett/gux/cmd/gux
# github.com/fsnotify/fsnotify

# Check if unused:
go mod why github.com/gorilla/websocket
# Output: "(main module does not need package ...)"
```

### Verify go.mod Changes

```bash
# Source: Project git workflow
# After running go mod tidy, check what changed
git diff go.mod

# Example output showing gorilla/websocket removal:
# -require github.com/gorilla/websocket v1.5.3
#
# +require (
# +    github.com/fsnotify/fsnotify v1.9.0
# +    ...
# +)
```

### Build Pattern for Examples

```bash
# Source: Project CLAUDE.md (testing examples section)
# DO use gux build:
cd examples/minimal
gux build
# Generates: .gux/dist/app.wasm, .gux/dist/admin.wasm, assets_gen.go, bin/app

# DON'T use go build directly:
cd examples/minimal
go build .
# ERROR: "pattern .gux/dist/admin.wasm: no matching files found"
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual dependency management | go mod tidy automatic cleanup | Go 1.11 (2018) | Removes error-prone manual editing |
| Vendor directories | Go module proxy + checksums | Go 1.13 (2019) | Better security, reproducible builds |
| GOPATH workspace | Go modules | Go 1.16 (2021) | Modules mode default, better isolation |
| go get for adding deps | go get with module-aware mode | Go 1.24 (2024) | Added tool directive support |

**No deprecated patterns in this phase:** Standard Go module commands (tidy, verify) remain the current best practice as of 2026.

## Open Questions

1. **Should examples eventually become separate modules?**
   - What we know: Current structure works, examples share dependencies cleanly
   - What's unclear: If the project scales to many more examples, would independent go.mod files be better?
   - Recommendation: Keep current structure. Only split if examples start having conflicting dependency requirements or need independent versioning

2. **Is there a CI/CD pipeline to automate this?**
   - What we know: No CI/CD configuration files found (no .github, .gitlab-ci.yml, etc.)
   - What's unclear: Whether dependency cleanup should be gated in PR workflows
   - Recommendation: Out of scope for this phase. Current manual verification via `gux build` is sufficient for milestone cleanup

3. **Should go mod verify run automatically?**
   - What we know: `go mod verify` checks cache integrity against go.sum
   - What's unclear: When it provides value vs. being redundant (Go tools verify automatically)
   - Recommendation: Include in manual cleanup workflow for documentation purposes, but it's largely redundant since `go build` verifies automatically

## Sources

### Primary (HIGH confidence)

- [Go Modules Reference](https://go.dev/ref/mod) - Official Go module documentation, go mod tidy specification
- [Managing Dependencies](https://go.dev/doc/modules/managing-dependencies) - Official dependency management guide
- Local project analysis via `go mod why`, `go list`, `go build` - Direct inspection of current module state
- Project CLAUDE.md - Project-specific build instructions and example structure

### Secondary (MEDIUM confidence)

- [Mastering go.mod: Dependency Management the Right Way in Go](https://medium.com/@moksh.9/mastering-go-mod-dependency-management-the-right-way-in-go-918226a69d58) - Best practices guide
- [Go mod tidy - A Quick Introduction](https://matthewsetter.com/go-mod-tidy-quick-intro/) - Practical tidy usage
- [Mastering Go Modules: A Practical Guide to Dependency Management](https://leapcell.medium.com/mastering-go-modules-a-practical-guide-to-dependency-management-e18eed09939c) - Comprehensive guide
- [Building a Monorepo in Golang](https://earthly.dev/blog/golang-monorepo/) - Monorepo patterns

### Tertiary (LOW confidence)

None - all findings verified against official documentation or direct project inspection.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Official Go toolchain commands, well-documented
- Architecture: HIGH - Direct inspection of project structure and working builds
- Pitfalls: HIGH - Verified through actual build attempts and go mod commands

**Research date:** 2026-01-26
**Valid until:** 2026-06-26 (6 months - Go module system is stable, only minor changes in point releases)
