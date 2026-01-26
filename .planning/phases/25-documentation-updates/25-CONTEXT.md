# Phase 25: Documentation Updates - Context

**Gathered:** 2026-01-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Update documentation to reflect packages removed in Phase 24 (components/, auth/, storage/, state/, ws/). This is documentation cleanup only - no code changes, no new documentation, just removal of stale content.

</domain>

<decisions>
## Implementation Decisions

### Removal Scope
- Delete doc files for dead packages entirely - no tombstones or redirects
- Remove sections about removed packages from CLAUDE.md (don't update to v2)
- Comprehensive hunt: find and remove all mentions across all docs
- Delete entire doc files that only cover dead packages

### Reference Cleanup
- Remove dead entries from project structure sections (don't reorganize)
- Delete examples that import/reference removed packages entirely
- Remove entire sentences containing dead file references
- Clean comments in Go files that mention dead packages (not just markdown)

### Verification
- Run both grep search AND link verification
- Create reusable verification script in scripts/
- Script should exit non-zero if stale references found (CI-friendly)
- Report findings before failing

### Claude's Discretion
- Exact grep patterns for package detection
- Script implementation details
- Order of operations for cleanup

</decisions>

<specifics>
## Specific Ideas

- Removed packages: components/, auth/, storage/, state/, ws/
- Clean removal approach - no migration notes or deprecation warnings
- CI-friendly verification for future use

</specifics>

<deferred>
## Deferred Ideas

None - discussion stayed within phase scope

</deferred>

---

*Phase: 25-documentation-updates*
*Context gathered: 2026-01-25*
