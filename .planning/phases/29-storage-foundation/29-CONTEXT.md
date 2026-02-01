# Phase 29: Storage Foundation - Context

**Gathered:** 2026-02-01
**Status:** Ready for planning

<domain>
## Phase Boundary

Storage interface, local filesystem backend, upload endpoint, and file serving handler. Developer can upload files through a server endpoint and serve them via HTTP, with a clean storage abstraction that decouples file location from application code. Upload UI (Phase 30), code generation (Phase 31), and per-field configuration (Phase 32) are separate phases.

</domain>

<decisions>
## Implementation Decisions

### Upload response shape
- Rich metadata: path, serving URL, original filename, size in bytes, content-type, plus image dimensions (width/height) for images
- Include both storage path (for DB storage) and ready-to-use serving URL (for immediate display)
- Accept multiple files per request in a single multipart POST; return array of results
- Structured JSON errors with error code, human message, and relevant limits (e.g. `{"error": "file_too_large", "message": "File exceeds 10MB limit", "max_size": 10485760}`)

### File naming & organization
- Content hash-based filenames (e.g. sha256hex.jpg) — predictable, collision-free
- Hash-prefix subdirectories: first 2 chars of hash as subdirectory (e.g. `uploads/ab/abc123def.jpg`) — prevents filesystem bottleneck
- Same content uploaded twice overwrites (identical bytes) — no deduplication logic or reference counting needed
- Original filename preserved in upload response metadata, not on disk

### Validation defaults
- No default maximum file size — developer must explicitly set limits via functional options
- All content types allowed by default — developer configures restrictions per field in Phase 32
- Magic bytes validation: block (reject upload) if extension doesn't match actual file content — strict security
- Configuration via functional options: `storage.NewLocalStorage("./uploads", storage.WithMaxSize(10*MB), storage.WithAllowedTypes("image/*"))`

### File serving behavior
- Images/PDFs display inline in browser; other types trigger download (standard Content-Disposition behavior)
- Aggressive caching: `Cache-Control: public, max-age=31536000` (1 year) since filenames are content hashes — content never changes for a given hash
- Hash-only URLs: `/__gux_api/files/ab/abc123def456.jpg` — clean, no encoding issues, matches disk structure
- Public serving by default; functional option to require authentication on the serve endpoint

### Claude's Discretion
- Exact hash algorithm choice (SHA-256, BLAKE2, etc.)
- Storage interface method signatures
- Internal error handling and logging
- Multipart parsing implementation details
- Image dimension detection approach

</decisions>

<specifics>
## Specific Ideas

- Multiple uploader configurations with different auth rules per path (some public, some restricted) — user wants this to be possible, but full per-field config is Phase 32
- Functional options pattern for storage configuration matches existing gux patterns (e.g. CRUD options like `core.WithRoles()`)

</specifics>

<deferred>
## Deferred Ideas

- Multiple storage configurations per path with different auth rules — Phase 32 (per-field configuration)
- Per-field upload constraints (accept, maxSize, directory) — Phase 32
- S3/cloud storage backends — future milestone

</deferred>

---

*Phase: 29-storage-foundation*
*Context gathered: 2026-02-01*
