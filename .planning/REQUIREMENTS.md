# Requirements: Gux Framework

**Defined:** 2026-01-25
**Core Value:** Clean, focused codebase with only active code

## v2.1 Requirements

Requirements for dead code cleanup. Each maps to roadmap phases.

### Code Removal

- [x] **CLEAN-01**: Remove components/ directory (61 files, old js.Value component library)
- [x] **CLEAN-02**: Remove auth/ directory (unused JWT/localStorage auth package)
- [x] **CLEAN-03**: Remove storage/ directory (only used by old components)
- [x] **CLEAN-04**: Remove state/ directory (superseded by core.Router state management)
- [x] **CLEAN-05**: Remove ws/ directory (WebSocket package never integrated)

### Documentation Updates

- [x] **DOCS-01**: Remove docs/websocket.md (documents deleted ws/ package)
- [x] **DOCS-02**: Update or remove docs/auth.md (documents deleted auth/ package)
- [x] **DOCS-03**: Update docs/components.md (may reference old components/)
- [x] **DOCS-04**: Update docs/_sidebar.md navigation
- [x] **DOCS-05**: Update README.md if it references removed packages
- [x] **DOCS-06**: Update LLM.txt framework description

### Dependency Cleanup

- [x] **DEPS-01**: Remove unused dependencies from go.mod (gorilla/websocket, etc.)
- [x] **DEPS-02**: Run go mod tidy to clean up go.sum
- [x] **DEPS-03**: Verify all examples still build after cleanup

## Out of Scope

Explicitly excluded from this milestone.

| Feature | Reason |
|---------|--------|
| Porting old component features to ui/ | Already completed in v2.0 |
| Adding new features | Cleanup only |
| Refactoring core/ | Already working correctly |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| CLEAN-01 | Phase 24 | Complete |
| CLEAN-02 | Phase 24 | Complete |
| CLEAN-03 | Phase 24 | Complete |
| CLEAN-04 | Phase 24 | Complete |
| CLEAN-05 | Phase 24 | Complete |
| DOCS-01 | Phase 25 | Complete |
| DOCS-02 | Phase 25 | Complete |
| DOCS-03 | Phase 25 | Complete |
| DOCS-04 | Phase 25 | Complete |
| DOCS-05 | Phase 25 | Complete |
| DOCS-06 | Phase 25 | Complete |
| DEPS-01 | Phase 26 | Complete |
| DEPS-02 | Phase 26 | Complete |
| DEPS-03 | Phase 26 | Complete |

**Coverage:**
- v2.1 requirements: 14 total
- Mapped to phases: 14
- Unmapped: 0 ✓

---
*Requirements defined: 2026-01-25*
*Last updated: 2026-01-26 after phase 26 complete — v2.1 milestone shipped*
