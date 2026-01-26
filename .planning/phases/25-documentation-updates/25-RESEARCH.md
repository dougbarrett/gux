# Phase 25: Documentation Updates - Research

**Researched:** 2026-01-25
**Domain:** Documentation cleanup for removed packages
**Confidence:** HIGH

## Summary

This phase removes documentation and references for five packages deleted in Phase 24: `components/`, `auth/`, `storage/`, `state/`, and `ws/`. The audit identified extensive stale documentation spread across 15+ files including dedicated doc files, README.md, CLAUDE.md, LLM.txt, COMPONENTS.md, and template files.

The cleanup approach is straightforward: delete entire files that only cover dead packages, remove sections about dead packages from multi-topic files, and create a verification script to ensure no stale references remain. The decision to delete rather than tombstone means no migration notes or deprecation warnings.

**Primary recommendation:** Delete docs/websocket.md, docs/auth.md, docs/components.md entirely, then surgically remove dead package references from remaining files using grep patterns, finishing with a verification script.

## Audit Findings

### Files to DELETE Entirely

These files document only dead packages and should be deleted in full:

| File | Documents | Action |
|------|-----------|--------|
| `docs/websocket.md` | ws/ package (575 lines) | DELETE |
| `docs/auth.md` | auth/ package (474 lines) | DELETE |
| `docs/components.md` | components/ package (1364 lines) | DELETE |
| `docs/state-management.md` | state/ package (568 lines) | DELETE |
| `COMPONENTS.md` | components/, state/ packages (1027 lines) | DELETE |

### Files Requiring Surgical Edits

These files contain sections about dead packages mixed with valid content:

| File | Dead References | Action |
|------|-----------------|--------|
| `docs/_sidebar.md` | Links to deleted docs | Remove 4 navigation entries |
| `docs/templates.md` | imports components/ | Remove import statements, update examples |
| `docs/api-generation.md` | References auth.GetToken() | Remove auth examples |
| `docs/getting-started.md` | imports components/, state/ | Major rewrite needed |
| `README.md` | Project structure, imports, examples | Remove dead packages from structure, examples |
| `CLAUDE.md` | Project structure lists components/, state/ | Remove from structure section |
| `LLM.txt` | Repository structure lists all dead packages | Remove from structure section |
| `.claude/skills/gux-framework.md` | imports components/, state/ | Remove dead references |
| `cmd/gux/templates/claude/skills/gux-framework.md` | imports components/, state/ | Remove dead references |

### Files Confirmed Clean

| File | Status |
|------|--------|
| `core/*.go` | No references to dead packages |
| `docs/server.md` | No dead package references |
| `docs/cli.md` | Clean |
| `docs/deployment.md` | Clean |
| `docs/keyboard-shortcuts.md` | Clean |
| `docs/accessibility.md` | Clean |

## Grep Patterns for Detection

Verified patterns that reliably detect stale references:

### Import Statements
```bash
# Full import path pattern (HIGH confidence)
grep -rn "github\.com/dougbarrett/gux/\(components\|auth\|storage\|state\|ws\)" --include="*.go" --include="*.md"

# Captures:
# - import "github.com/dougbarrett/gux/components"
# - import "github.com/dougbarrett/gux/auth"
# - import "github.com/dougbarrett/gux/state"
# - import "github.com/dougbarrett/gux/ws"
# - import "github.com/dougbarrett/gux/storage"
```

### Package References in Code Examples
```bash
# Package prefix pattern in code examples (HIGH confidence)
grep -rn "\(components\|auth\|state\|ws\)\.\(New\|Init\|Toast\|Button\|Get\|Set\|Store\|Client\)" --include="*.md"

# Captures:
# - components.Button(
# - components.Toast(
# - auth.GetToken()
# - state.New(
# - ws.NewClient(
```

### Documentation Links
```bash
# Internal doc links pattern (HIGH confidence)
grep -rn "\(websocket\.md\|auth\.md\|components\.md\|state-management\.md\)" --include="*.md"

# Captures sidebar links and cross-references
```

### Project Structure References
```bash
# Directory listing pattern (MEDIUM confidence - may have false positives)
grep -rn "├── \(components/\|auth/\|storage/\|state/\|ws/\)" --include="*.md" --include="*.txt"
grep -rn "│.*\(components/\|auth/\|storage/\|state/\|ws/\)" --include="*.md" --include="*.txt"
```

### Prose References
```bash
# Package names in sentences (MEDIUM confidence - manual review needed)
grep -rn "the \(components\|auth\|storage\|state\|ws\) package" --include="*.md"
grep -rn "\`\(components\|auth\|storage\|state\|ws\)\`" --include="*.md"
```

## Verification Script Design

### Requirements (from CONTEXT.md)
- Run both grep search AND link verification
- Create reusable script in scripts/
- Exit non-zero if stale references found (CI-friendly)
- Report findings before failing

### Recommended Implementation: Shell Script

**Rationale:** Shell script is simpler, has no dependencies, and aligns with existing project patterns. A Go implementation would require building before running.

### Script Structure
```bash
#!/bin/bash
# scripts/verify-no-stale-refs.sh
# Verifies no references to removed packages remain

set -e

DEAD_PACKAGES="components|auth|storage|state|ws"
EXIT_CODE=0
FOUND_REFS=""

# Check import statements
echo "Checking import statements..."
if grep -rn "github\.com/dougbarrett/gux/\(${DEAD_PACKAGES}\)" \
    --include="*.go" --include="*.md" \
    --exclude-dir=".planning" \
    --exclude-dir=".git" 2>/dev/null; then
    FOUND_REFS="imports"
    EXIT_CODE=1
fi

# Check package prefix usage
echo "Checking package usage..."
if grep -rn "\(components\|auth\|state\|ws\)\.\(New\|Init\|Get\|Set\|Toast\|Button\|Store\|Client\)" \
    --include="*.md" \
    --exclude-dir=".planning" \
    --exclude-dir=".git" 2>/dev/null; then
    FOUND_REFS="${FOUND_REFS} usage"
    EXIT_CODE=1
fi

# Check documentation links
echo "Checking documentation links..."
if grep -rn "\(websocket\.md\|auth\.md\|components\.md\|state-management\.md\)" \
    --include="*.md" \
    --exclude-dir=".planning" \
    --exclude-dir=".git" 2>/dev/null; then
    FOUND_REFS="${FOUND_REFS} links"
    EXIT_CODE=1
fi

# Check project structure listings
echo "Checking project structure..."
if grep -rn "├── \(components/\|auth/\|storage/\|state/\|ws/\)" \
    --include="*.md" --include="*.txt" \
    --exclude-dir=".planning" \
    --exclude-dir=".git" 2>/dev/null; then
    FOUND_REFS="${FOUND_REFS} structure"
    EXIT_CODE=1
fi

# Summary
echo ""
if [ $EXIT_CODE -eq 0 ]; then
    echo "SUCCESS: No stale references found"
else
    echo "FAILURE: Found stale references in: ${FOUND_REFS}"
    echo "Please remove all references to: ${DEAD_PACKAGES}"
fi

exit $EXIT_CODE
```

## Order of Operations

Recommended cleanup sequence to minimize conflicts and enable incremental verification:

### Phase 1: Delete Dead Documentation Files
1. Delete `docs/websocket.md`
2. Delete `docs/auth.md`
3. Delete `docs/components.md`
4. Delete `docs/state-management.md`
5. Delete `COMPONENTS.md`

**Rationale:** Removing entire files first reduces grep noise for subsequent edits.

### Phase 2: Update Navigation
1. Edit `docs/_sidebar.md` to remove links to deleted files

**Rationale:** Prevents broken links in docsify navigation.

### Phase 3: Update Core Documentation
1. Edit `CLAUDE.md` - remove dead packages from Project Structure
2. Edit `LLM.txt` - remove dead packages from Repository Structure
3. Edit `README.md` - remove project structure entries, dead imports, dead examples

**Rationale:** These are the most visible files and define project structure.

### Phase 4: Update Remaining Docs
1. Edit `docs/templates.md` - remove dead imports from examples
2. Edit `docs/api-generation.md` - remove auth.GetToken() references
3. Edit `docs/getting-started.md` - remove dead imports/examples

**Rationale:** Less critical files that may have significant edits needed.

### Phase 5: Update Claude Skills
1. Edit `.claude/skills/gux-framework.md`
2. Edit `cmd/gux/templates/claude/skills/gux-framework.md`

**Rationale:** Template files are scaffolded into new projects.

### Phase 6: Create Verification Script
1. Create `scripts/verify-no-stale-refs.sh`
2. Make it executable
3. Run to verify cleanup complete

### Phase 7: Final Verification
1. Run verification script
2. Manual review of any flagged items
3. Fix any remaining issues

## Common Pitfalls

### Pitfall 1: Over-Matching with Grep
**What goes wrong:** Grep patterns match legitimate content (e.g., "state" matches "state machine")
**Why it happens:** Package names are common English words
**How to avoid:** Use compound patterns that include package prefix (e.g., `state.New` not just `state`)
**Warning signs:** False positives in grep output

### Pitfall 2: Breaking Cross-References
**What goes wrong:** Removing a doc file breaks links from other files
**Why it happens:** Files reference each other with "See [Authentication](auth.md)"
**How to avoid:** Update `_sidebar.md` first, then grep for all `auth.md` links
**Warning signs:** Docsify 404 errors

### Pitfall 3: Incomplete Example Removal
**What goes wrong:** Removing an import but leaving code that uses the import
**Why it happens:** Examples span multiple code blocks
**How to avoid:** Read full context around any import removal
**Warning signs:** Code examples that reference undefined packages

### Pitfall 4: Planning Directory False Positives
**What goes wrong:** Verification script flags .planning/ files
**Why it happens:** Planning docs reference packages by design
**How to avoid:** Exclude .planning/ from grep searches
**Warning signs:** Script fails but only in planning files

## Code Examples

### _sidebar.md After Cleanup
```markdown
- **Getting Started**
  - [Introduction](/)
  - [Getting Started](getting-started.md)
  - [CLI Reference](cli.md)

- **Core Concepts**
  - [API Generation](api-generation.md)
  - [Templates](templates.md)

- **Features**
  - [Server Utilities](server.md)

- **Reference**
  - [Keyboard Shortcuts](keyboard-shortcuts.md)
  - [Accessibility](accessibility.md)
  - [Deployment](deployment.md)
```

### CLAUDE.md Project Structure After Cleanup
```markdown
## Project Structure

\`\`\`
goquery/
├── core/                    # Universal rendering framework
│   ├── node.go             # Node interface & types
│   ├── elements.go         # Element helpers
│   ├── app.go              # App, Router, State
│   ├── crud.go             # CRUD API generation
│   ├── csrf.go             # CSRF protection
│   ├── html_renderer.go    # Server-side rendering
│   ├── dom_renderer.go     # WASM DOM rendering
│   └── router_*.go         # Platform-specific routing
├── fetch/                   # WASM HTTP client with auto CSRF
├── api/                    # API utilities & error types
├── server/                 # Server middleware
├── cmd/gux/               # CLI tool
└── examples/
    └── minimal/           # Reference implementation
\`\`\`
```

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Dead link detection | Custom link checker | grep for `.md` links | Simple, already have patterns |
| Finding all references | Manual search | grep patterns above | Comprehensive, repeatable |
| Verification | Manual review | Shell script | CI-friendly, automated |

## Open Questions

### Question 1: docs/getting-started.md Rewrite Scope
- **What we know:** File heavily references components/, state/
- **What's unclear:** How much rewrite is needed vs. just deletion
- **Recommendation:** Review during task execution; may need to defer significant rewrites to separate phase

### Question 2: Examples in README.md
- **What we know:** README has component examples that will be deleted
- **What's unclear:** What replaces them (core/ examples?)
- **Recommendation:** Remove dead examples; adding new core/ examples is out of scope per CONTEXT.md decisions

## Sources

### Primary (HIGH confidence)
- Direct file audit of repository (this research)
- CONTEXT.md decisions from discussion phase

### Verification
- All grep patterns tested against current repository state
- File counts and line counts verified with `wc -l`

## Metadata

**Confidence breakdown:**
- Audit findings: HIGH - direct file inspection
- Grep patterns: HIGH - tested against codebase
- Pitfalls: HIGH - based on file cross-reference analysis
- Verification script: MEDIUM - design only, not tested

**Research date:** 2026-01-25
**Valid until:** Until next code change (static documentation)
