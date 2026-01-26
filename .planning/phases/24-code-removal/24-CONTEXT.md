# Phase 24: Code Removal - Context

**Gathered:** 2026-01-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Remove dead packages from pre-v2.0 architecture: `components/`, `auth/`, `storage/`, `state/`, `ws/`. Verify no remaining references exist. Documentation updates are Phase 25; dependency cleanup is Phase 26.

</domain>

<decisions>
## Implementation Decisions

### Verification depth
- Use static analysis tools (go vet, staticcheck) for thorough verification
- Full verification scope: main module AND all example apps (auth, marketing, saas, admin)
- Standard analysis only: check direct imports, no reflection/string-based scanning needed
- If references to dead packages are found: fix automatically (remove stale imports/references)

### Claude's Discretion
- Specific static analysis tool selection
- Order of package deletion
- Commit message formatting

</decisions>

<specifics>
## Specific Ideas

No specific requirements — standard dead code removal approach.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 24-code-removal*
*Context gathered: 2026-01-25*
