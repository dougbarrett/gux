---
phase: 32
plan: 02
subsystem: code-generation
tags: [code-gen, multi-file, file-upload, admin-ui, config]
requires: [32-01]
provides:
  - Multi-file admin form generation
  - Per-field file configuration system
  - File gallery rendering in admin
depends_on:
  - runtime: 32-01 (multi-file runtime infrastructure)
  - ui: 32-01 (MultiFileUpload component)
affects:
  - developers: `gux model regen` now supports file[] fields
  - developers: Per-field accept, maxSize, directory config flows to generated UI
tech-stack:
  patterns:
    - Code generation with template fields
    - Size string parsing at generation time
    - JSON state management for file arrays
key-files:
  created: []
  modified:
    - cmd/gux/model.go
    - cmd/gux/modelgen.go
    - cmd/gux/modelgen_test.go
decisions:
  - title: Per-field config parsed at generation time
    context: MaxSize strings like "5MB" must be converted to bytes for UI props
    decision: parseSizeString converts at generation time, emits literal int64 in generated code
    rationale: Avoids runtime parsing overhead and string manipulation in generated admin pages
  - title: Multi-file fields use JSON arrays in string model field
    context: Database schema uses string type for multi-file fields
    decision: Store JSON array of storage keys in string field, DTOs use []*FileInfo
    rationale: Matches 32-01 design - backwards compatible with existing string fields
  - title: Early return from convertToTemplateField for file fields
    context: EditStateInit was being overwritten by type switch after file-specific setup
    decision: Generate form/detail code and return early for file/file[] fields
    rationale: Prevents type-based switch from overwriting file-specific initialization
metrics:
  duration: 6 minutes
  tasks: 2
  commits: 2
  tests_added: 6
completed: 2026-02-03
---

# Phase 32 Plan 02: Multi-File and Configuration Code Generation

**Complete admin scaffolding for multi-file fields and per-field upload constraints**

## One-Liner

Code generation for `"input": "file[]"` fields produces admin forms with MultiFileUpload, galleries in detail views, and per-field accept/maxSize/directory configuration.

## Objective

Extend the code generation pipeline to support `"input": "file[]"` fields and per-field file configuration (`accept`, `maxSize`, `directory`) in gux.config.json, generating admin forms with MultiFileUpload, detail views with file galleries, and list views with file counts.

## Accomplishments

### Code Generation Infrastructure
1. **ModelField Extensions**
   - Added `Accept`, `MaxSize`, `Directory` fields to ModelField struct
   - Implemented `parseSizeString()` utility to convert "5MB" -> 5242880 bytes at generation time
   - Comprehensive tests for all size formats (B, KB, MB, GB, case-insensitive, with spaces)

2. **TemplateField Extensions**
   - Added `IsMultiFile`, `FileAccept`, `FileMaxSize`, `FileDir` fields
   - Extended `convertToTemplateField()` to parse file config from ModelField
   - Set `DTOType = "[]*core.FileInfo"` and `GoType = "string"` for file[] fields
   - Generate `EditStateInit` that reconstructs JSON key array from []*FileInfo

3. **Form Code Generation**
   - Multi-file fields emit `ui.MultiFileUpload` with JSON state management
   - Per-field config (Accept, MaxSize, Directory) emits as literal props in generated code
   - OnUploadComplete appends to JSON key array
   - OnRemove removes from JSON key array by index
   - Single file fields emit per-field UploadURL with ?dir= parameter when Directory is set

4. **Detail View Code Generation**
   - Multi-file fields render image galleries with thumbnails (h-20 w-20 rounded)
   - Non-image files render as download links with paperclip icons
   - Mixed content (images + files) separated with mt-4 spacing
   - Empty/nil file arrays show "-"

5. **List View Code Generation**
   - Multi-file fields show first thumbnail + "+N more" indicator for images
   - Non-image files show "{N} files" count
   - Empty arrays show "-"

6. **Template Data Management**
   - Added `HasMultiFileFields` flag to ModelTemplateData
   - Import `encoding/json` in AdminNew template when HasMultiFileFields is true
   - AdminEdit already imports encoding/json unconditionally

## Task Commits

| Task | Description | Commit | Files |
|------|-------------|--------|-------|
| 1 | ModelField extensions and parseSizeString | ac41f5c | cmd/gux/model.go, cmd/gux/modelgen_test.go |
| 2 | Code generation for file[] and per-field config | 7825396 | cmd/gux/modelgen.go, cmd/gux/modelgen_test.go |

## Files Created

None - extended existing code generation system.

## Files Modified

| File | Changes |
|------|---------|
| `cmd/gux/model.go` | Added Accept, MaxSize, Directory to ModelField; added parseSizeString utility |
| `cmd/gux/modelgen.go` | Extended TemplateField, convertToTemplateField, generateFormFieldCode, generateDetailFieldCode, AdminList template for multi-file support and per-field config |
| `cmd/gux/modelgen_test.go` | Added 7 tests: parseSizeString (all formats), multi-file conversion, file config, form/detail code generation |

## Decisions Made

### Per-field config parsed at generation time
**Context**: MaxSize strings like "5MB" must be converted to bytes for UI props.

**Decision**: `parseSizeString` converts at generation time, emits literal int64 in generated code.

**Rationale**: Avoids runtime parsing overhead and string manipulation in generated admin pages. Config validation happens once at generation time, not on every page render.

### Multi-file fields use JSON arrays in string model field
**Context**: Database schema uses string type for multi-file fields.

**Decision**: Store JSON array of storage keys in string field, DTOs use `[]*FileInfo`.

**Rationale**: Matches 32-01 design. Backwards compatible with existing string fields. DTO transformation handles JSON marshaling/unmarshaling transparently.

### Early return from convertToTemplateField for file fields
**Context**: EditStateInit was being overwritten by type switch after file-specific setup.

**Decision**: Generate form/detail code and return early for file/file[] fields.

**Rationale**: Prevents type-based switch (case "string") from overwriting file-specific initialization. File fields have custom state management that differs from plain strings.

## Deviations from Plan

None - plan executed exactly as written.

## Technical Notes

### parseSizeString Implementation
- Supports B, KB, MB, GB suffixes (case-insensitive)
- Handles spaces between number and suffix ("5 MB" works)
- Validates entire input string (rejects "5XB")
- Returns error for invalid formats
- Plain numbers (no suffix) treated as bytes

### EditStateInit for Multi-File
Generated code reconstructs JSON key array from DTO's []*FileInfo:
```go
func() string {
    if displayItem.Gallery != nil {
        keys := make([]string, len(displayItem.Gallery))
        for i, fi := range displayItem.Gallery {
            keys[i] = fi.Key
        }
        data, _ := json.Marshal(keys)
        return string(data)
    }
    return ""
}()
```

### Form Code for Multi-File
Generated MultiFileUpload manages JSON state:
```go
func() core.Node {
    var keys []string
    if galleryState.Get() != "" {
        json.Unmarshal([]byte(galleryState.Get()), &keys)
    }
    urls := make([]string, len(keys))
    for i, k := range keys {
        urls[i] = "/__gux_api/files/" + k
    }
    return ui.MultiFileUpload(ui.MultiFileUploadProps{
        Values: urls,
        Accept: []string{"image/*"},
        MaxSize: 5242880,
        UploadURL: "/__gux_api/upload?dir=gallery",
        OnUploadComplete: func(result ui.UploadResult) {
            // Append to key array
        },
        OnRemove: func(index int) {
            // Remove from key array
        },
    })
}()
```

## Test Coverage

| Test | Purpose |
|------|---------|
| `TestParseSizeString` | All size formats, edge cases, error handling |
| `TestConvertToTemplateField_MultiFile` | IsMultiFile flag, DTOType, EditStateInit |
| `TestConvertToTemplateField_FileWithConfig` | FileAccept, FileMaxSize, FileDir population |
| `TestGenerateFormFieldCode_MultiFile` | ui.MultiFileUpload, JSON state management |
| `TestGenerateFormFieldCode_FileWithConfig` | Per-field Accept/MaxSize/Directory props |
| `TestGenerateDetailFieldCode_MultiFile` | Gallery rendering, image detection |

All tests pass. No regressions in core test suite.

## Issues Encountered

None - straightforward code generation extension.

## Performance

**Duration**: 6 minutes (2 tasks, 2 commits)

**Verification**: All core tests pass (`./cmd/gux/... ./core/... ./ui/...`)

## Next Phase Readiness

**Phase complete**: This was the final plan of phase 32 (Multi-file Support and Per-field Configuration).

**What's now possible**:
- Developers add `"input": "file[]"` to gux.config.json
- `gux model regen` generates full admin scaffolding for multi-file fields
- Per-field `accept`, `maxSize`, `directory` config flows to generated UI props
- Admin forms show MultiFileUpload with JSON state management
- Detail views show file galleries with image previews
- List views show file counts or thumbnails

**Integration**:
- Works with 32-01 runtime infrastructure (MultiFileUpload component, ParseFileKeys helpers)
- Works with 31-02 file upload system (FileInfo, storage interface)
- Works with existing CRUD code generation system

**No blockers** - phase 32 complete.
