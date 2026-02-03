# Phase 31: Code Generation & CRUD Integration - Context

**Gathered:** 2026-02-02
**Status:** Ready for planning

<domain>
## Phase Boundary

Developer adds `"input": "file"` to a field in gux.config.json and `gux gen` produces complete upload integration across admin forms, detail views, list views, and CRUD lifecycle. Multi-file fields and per-field configuration (allowed types, max size, upload directory) are Phase 32.

</domain>

<decisions>
## Implementation Decisions

### Generated admin forms (create/edit)
- File fields render `ui.FileUpload` component in both create and edit forms
- On edit: inline preview above the upload zone — current image thumbnail (or filename for non-images) with a "Remove" button; upload zone text changes to "Replace file" instead of "Upload file"
- File fields are optional by default; only required if `"required": true` is set in gux.config.json
- "Remove" button clears the field value to empty string on save; old file is cleaned up from storage
- Validation errors (wrong type, too large, etc.) display inline below the FileUpload component, matching existing form field validation patterns

### Generated list views
- File fields show a small thumbnail + filename in the table cell
- Non-image files show a generic file-type icon + filename
- Thumbnail size consistent with table row height (~40px)

### Generated detail views
- Image file fields render a medium preview (~200-300px wide), clickable to open full-size in new tab
- Non-image file fields show a file-type icon + filename + download button
- DTO includes file metadata struct (URL, filename, size, content type) — not just a URL string

### File auto-cleanup
- Automatic by default — any CRUD model with file fields gets auto-cleanup on record delete and file replacement
- When a record is hard-deleted, all associated file fields are cleaned up from storage automatically
- When a file field is updated (replaced), the old file is immediately deleted
- If a file upload succeeds but the database save fails, the uploaded file is rolled back (deleted) to prevent orphans
- Developer can disable auto-cleanup if needed via a config option

### Lifecycle hooks
- Two upload hooks: `WithBeforeUpload()` runs before file is written to storage (receives file metadata — name, size, type — can reject); `WithAfterUpload()` runs after file is saved (receives file path and metadata, read-only access, can return extra metadata to store)
- Both hooks follow the existing CRUD option pattern (`WithCreateHook`, etc.) — added as CRUD options
- `WithBeforeFileDelete()` hook fires before file removal on delete/replace; returning an error prevents the file from being deleted (useful for archival/compliance)
- AfterUpload hook can generate derivatives (thumbnails, transcodes) but cannot modify the original file

### Claude's Discretion
- File metadata struct shape (exact fields beyond URL, filename, size, content type)
- How file-type icons are determined (extension-based mapping)
- Exact thumbnail generation approach in list views (CSS resize vs server-side)
- Implementation details of rollback on DB save failure

</decisions>

<specifics>
## Specific Ideas

No specific references — open to standard approaches that follow existing gux code generation patterns. The key constraint is consistency with existing CRUD option patterns and admin scaffolding conventions.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope. Multi-file support and per-field configuration (allowed types, max size, upload directory) are explicitly Phase 32.

</deferred>

---

*Phase: 31-code-generation-crud-integration*
*Context gathered: 2026-02-02*
