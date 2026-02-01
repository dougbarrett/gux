# Requirements: Gux v2.4 File Upload System

**Defined:** 2026-02-01
**Core Value:** Developers can add `"input": "file"` to a model field in gux.config.json and get complete file upload -- UI, storage, API, admin integration -- with zero manual wiring.

## v1 Requirements

Requirements for v2.4 milestone. Each maps to roadmap phases.

### Storage

- [x] **STOR-01**: Developer can configure a storage backend via `app.SetStorage()` in code
- [ ] **STOR-02**: Developer can configure storage in gux.config.json with code overrides
- [x] **STOR-03**: Local filesystem storage backend stores files to a configurable directory
- [x] **STOR-04**: Storage interface provides Put, Delete, and URL methods
- [x] **STOR-05**: Uploaded files are accessible via HTTP URL (served by Go HTTP server for local storage)
- [x] **STOR-06**: File paths stored in database are provider-agnostic keys (not absolute paths or full URLs)

### Upload Endpoint

- [x] **UPLD-01**: Server accepts multipart file uploads at a dedicated upload endpoint (separate from CRUD)
- [x] **UPLD-02**: Server validates file size against configurable maximum before fully reading the file
- [x] **UPLD-03**: Server validates file type via magic bytes (content sniffing), not just Content-Type header
- [x] **UPLD-04**: Server generates unique filenames (UUID/xid-based) to prevent collisions and path traversal
- [x] **UPLD-05**: Upload endpoint respects existing CSRF protection automatically
- [x] **UPLD-06**: Upload endpoint respects authentication when auth is configured on the app

### Upload UI Component

- [x] **UICM-01**: User can click to browse and select a file for upload
- [x] **UICM-02**: User can drag and drop a file onto the upload zone
- [x] **UICM-03**: User sees a progress indicator during file upload
- [x] **UICM-04**: User sees client-side validation errors for disallowed file types before upload starts
- [x] **UICM-05**: User sees client-side validation errors for oversized files before upload starts
- [x] **UICM-06**: User sees an image preview after selecting an image file (before upload)
- [x] **UICM-07**: User can remove or replace an existing uploaded file
- [x] **UICM-08**: Component renders a static upload zone in SSR (no browser API calls)
- [x] **UICM-09**: Component hydrates with full interactivity in WASM (drag-drop, progress, preview)
- [x] **UICM-10**: File bytes stay in JavaScript land during upload (not copied into Go WASM memory)

### Multi-File Support

- [ ] **MULT-01**: Developer can define a multi-file field using `"input": "file[]"` or equivalent in gux.config.json
- [ ] **MULT-02**: User can upload multiple files to a multi-file field
- [ ] **MULT-03**: User can remove individual files from a multi-file field
- [ ] **MULT-04**: Multi-file values are stored as a JSON array of file paths in the database

### CRUD & Code Generation Integration

- [ ] **CRUD-01**: Developer adds `"input": "file"` to a field in gux.config.json and `gux gen` produces working upload integration
- [ ] **CRUD-02**: Generated model uses `string` type for single-file fields (stores file path)
- [ ] **CRUD-03**: Generated DTOs include the file path/URL for file fields
- [ ] **CRUD-04**: Generated admin create forms render `ui.FileUpload` for file fields
- [ ] **CRUD-05**: Generated admin edit forms render `ui.FileUpload` with current file displayed
- [ ] **CRUD-06**: Generated admin detail views show image preview for image file fields
- [ ] **CRUD-07**: Generated admin detail views show filename + download link for non-image file fields
- [ ] **CRUD-08**: Generated admin list views show thumbnail for image file fields
- [ ] **CRUD-09**: Generated admin list views show file icon or filename for non-image file fields
- [ ] **CRUD-10**: Files are automatically deleted from storage when the parent record is deleted via CRUD

### Lifecycle Hooks

- [ ] **HOOK-01**: Developer can register a BeforeUpload hook that validates or rejects a file before it is stored
- [ ] **HOOK-02**: Developer can register an AfterUpload hook that processes a file after it is stored (e.g., document conversion, metadata extraction)
- [ ] **HOOK-03**: Developer can register a BeforeDelete hook that runs before a file is removed from storage
- [ ] **HOOK-04**: BeforeUpload hook receives file metadata (name, size, content type)
- [ ] **HOOK-05**: AfterUpload hook receives stored file info (path, URL, original name, size, content type)
- [ ] **HOOK-06**: Hooks follow the existing `WithCreateHook` / CRUD option pattern

### Configuration

- [ ] **CONF-01**: File field accepts configurable allowed file types per field (e.g., `"accept": "image/*"`)
- [ ] **CONF-02**: File field accepts configurable max file size per field (e.g., `"maxSize": "5MB"`)
- [ ] **CONF-03**: File field accepts configurable upload directory per field (e.g., `"directory": "avatars"`)

## v2 Requirements

Deferred to future milestone. Tracked but not in current roadmap.

### S3 Storage Backend

- **S3-01**: Developer can configure S3-compatible storage (AWS, MinIO, DigitalOcean Spaces, etc.)
- **S3-02**: S3 backend supports presigned URLs for direct browser-to-S3 uploads
- **S3-03**: S3 backend uses AWS SDK Go v2 with multipart upload for large files

### Image Processing

- **IMG-01**: Server generates thumbnails on upload (configurable sizes)
- **IMG-02**: Thumbnails stored alongside originals in same storage backend
- **IMG-03**: Admin list views use generated thumbnails instead of full images

### Multi-File Enhancements

- **MFIL-01**: User can drag-and-drop to reorder files in a multi-file field
- **MFIL-02**: Image cropping UI for image file fields

## Out of Scope

| Feature | Reason |
|---------|--------|
| Built-in media library / asset manager | Massive scope; becomes its own product. Users build one with CRUD + file fields if needed. |
| Virus/malware scanning | Infrastructure concern. Document as BeforeUpload hook use case. |
| Image transformation CDN (resize on request) | Use AfterUpload hook + external service. Server-side on-demand resizing invites abuse. |
| Chunked/resumable uploads (TUS protocol) | Standard multipart handles files <100MB. Add later if users need large file uploads. |
| Universal file preview (PDF, Office docs) | Heavy JS dependencies. Preview images only; show download link for everything else. |
| Per-file ACL / permission system | Files inherit parent model's access control. |
| Client-side image resizing | Adds WASM binary size. Defer to server-side processing. |
| Presigned URL uploads (direct-to-S3) | Deferred with S3 backend to v2. |
| CDN cache invalidation | Infrastructure concern. Use content-hash filenames for cache busting. |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| STOR-01 | Phase 29 | Complete |
| STOR-02 | Phase 29 | Deferred (config-driven setup is Phase 31) |
| STOR-03 | Phase 29 | Complete |
| STOR-04 | Phase 29 | Complete |
| STOR-05 | Phase 29 | Complete |
| STOR-06 | Phase 29 | Complete |
| UPLD-01 | Phase 29 | Complete |
| UPLD-02 | Phase 29 | Complete |
| UPLD-03 | Phase 29 | Complete |
| UPLD-04 | Phase 29 | Complete |
| UPLD-05 | Phase 29 | Complete |
| UPLD-06 | Phase 29 | Complete |
| UICM-01 | Phase 30 | Complete |
| UICM-02 | Phase 30 | Complete |
| UICM-03 | Phase 30 | Complete |
| UICM-04 | Phase 30 | Complete |
| UICM-05 | Phase 30 | Complete |
| UICM-06 | Phase 30 | Complete |
| UICM-07 | Phase 30 | Complete |
| UICM-08 | Phase 30 | Complete |
| UICM-09 | Phase 30 | Complete |
| UICM-10 | Phase 30 | Complete |
| CRUD-01 | Phase 31 | Pending |
| CRUD-02 | Phase 31 | Pending |
| CRUD-03 | Phase 31 | Pending |
| CRUD-04 | Phase 31 | Pending |
| CRUD-05 | Phase 31 | Pending |
| CRUD-06 | Phase 31 | Pending |
| CRUD-07 | Phase 31 | Pending |
| CRUD-08 | Phase 31 | Pending |
| CRUD-09 | Phase 31 | Pending |
| CRUD-10 | Phase 31 | Pending |
| HOOK-01 | Phase 31 | Pending |
| HOOK-02 | Phase 31 | Pending |
| HOOK-03 | Phase 31 | Pending |
| HOOK-04 | Phase 31 | Pending |
| HOOK-05 | Phase 31 | Pending |
| HOOK-06 | Phase 31 | Pending |
| MULT-01 | Phase 32 | Pending |
| MULT-02 | Phase 32 | Pending |
| MULT-03 | Phase 32 | Pending |
| MULT-04 | Phase 32 | Pending |
| CONF-01 | Phase 32 | Pending |
| CONF-02 | Phase 32 | Pending |
| CONF-03 | Phase 32 | Pending |

**Coverage:**
- v1 requirements: 45 total
- Mapped to phases: 45
- Unmapped: 0

---
*Requirements defined: 2026-02-01*
*Last updated: 2026-02-01 after phase 30 completion*
