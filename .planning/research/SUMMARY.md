# Project Research Summary

**Project:** Gux File Upload System
**Domain:** File upload integration for a full-stack Go/WASM framework with code generation, SSR, and storage abstraction
**Researched:** 2026-02-01
**Confidence:** HIGH

## Executive Summary

The file upload system is a cross-cutting feature that touches every layer of the Gux framework: storage backends, HTTP endpoints, WASM client, UI components, code generation, and admin scaffolding. Expert consensus across Django, Laravel, Rails, and Go ecosystems is that the right approach is a thin storage abstraction interface (owned by the framework, not a third-party library) with pluggable backends, a separate upload endpoint (not mixed into CRUD JSON handlers), and a declarative config-driven model (`"input": "file"` in `gux.config.json`) that generates all the wiring automatically. The killer feature is zero-config file fields: add one config line, get upload UI, storage, API, and admin integration with no manual code.

The recommended stack is minimal: only 2 new direct dependencies (AWS SDK v2 for S3, rs/xid for file IDs), with everything else using Go stdlib or extending existing Gux infrastructure. The critical architectural decision is to keep file uploads as a separate endpoint flow (`/__gux_api/upload/{model}/{field}`) rather than modifying CRUD to accept multipart. This preserves the existing JSON-based CRUD contract and enables progress tracking. On the WASM side, the upload component must use `XMLHttpRequest` via `syscall/js` because the browser Fetch API does not support upload progress events, and file bytes must stay in JavaScript land (never copied into Go WASM memory) to avoid crashing on large files.

The top risks are: (1) the existing `fetch` package's string-only Body field cannot send FormData, requiring a new upload path that must still integrate CSRF token injection; (2) SSR/WASM build-tag separation must be designed from the start to avoid server panics; and (3) orphaned files from failed model saves need a cleanup strategy from day one. All risks have well-understood mitigations documented in the research.

## Key Findings

### Recommended Stack

Only 2 new direct dependencies are needed. The S3 dependency is opt-in at runtime -- apps using local storage do not pull in AWS SDK.

**Core technologies:**
- **AWS SDK Go v2** (`aws-sdk-go-v2/service/s3`): S3-compatible cloud storage -- official SDK, works with MinIO/R2/Spaces, v1 reached EOL July 2025
- **rs/xid**: Unique file ID generation -- 20-char, URL-safe, time-sortable, cleaner than UUID for file paths
- **Go stdlib** (`mime/multipart`, `net/http`, `image`, `io`): Multipart parsing, size limits, MIME detection, image validation -- no external deps needed
- **XMLHttpRequest via syscall/js**: WASM upload with progress tracking -- Fetch API does not support upload progress events

**Not adding:** MinIO Go SDK (redundant with AWS SDK), third-party storage abstractions (stow/go-storage), image thumbnails (deferred), tus resumable uploads (over-engineered for v1), presigned URL uploads (optimization for later).

See [STACK.md](STACK.md) for full rationale and alternatives.

### Expected Features

**Must have (table stakes) -- 20 features:**
- Click-to-browse and drag-and-drop file selection
- Upload progress indicator (via XMLHttpRequest)
- Client-side AND server-side file type/size validation
- Single and multiple file upload per field
- Image preview before upload (using `URL.createObjectURL`)
- Storage abstraction with local filesystem + S3 backends
- File serving endpoint with auth awareness
- CRUD integration in admin forms, detail views, and list views
- Delete/replace existing files with old file cleanup
- Unique filename generation (xid-based), configurable upload paths
- Automatic CSRF protection on upload endpoints

**Should have (differentiators) -- 10 features:**
- Zero-config file fields via `gux.config.json` (the killer feature)
- Lifecycle hooks: `BeforeUpload`, `AfterUpload`, `BeforeDelete`
- Automatic orphan cleanup on record delete (`WithFileField` option)
- File metadata storage (original filename, MIME type, size)
- Storage config in `gux.config.json` with env var interpolation
- Accept filter presets (`"accept": "images"` expands to common MIME types)

**Defer to v2+:**
- Presigned URL uploads (direct-to-S3)
- Image thumbnail generation (use `disintegration/imaging` later)
- Image cropping UI (Cropper.js integration)
- Drag-and-drop reordering for multi-file fields
- Chunked/resumable uploads (TUS protocol)

**Anti-features (do NOT build):** Media library/asset manager, on-demand image transformation CDN, virus scanning, in-browser video transcoding, universal file preview (PDF/Office), per-file ACL system.

See [FEATURES.md](FEATURES.md) for full feature landscape and dependency graph.

### Architecture Approach

The system adds a new `storage/` package at the framework level (alongside `core/`, `ui/`, `fetch/`), a separate upload endpoint that returns file paths to be included in JSON CRUD payloads, and a `ui.FileUpload` component with SSR/WASM build-tag separation. The model field is always a `string` in the database (storing the file path); the "file" nature is a UI and upload concern, not a schema concern.

**Major components:**
1. **`storage/` package** -- `Storage` interface with `Upload`/`Delete`/`URL` methods; `LocalStorage` (disk) and `S3Storage` (AWS SDK) implementations; kept separate from `core/` to avoid pulling AWS SDK into all apps
2. **Upload HTTP handler** -- `POST /__gux_api/upload/{model}/{field}` parses multipart, validates, stores via Storage, returns JSON with path/URL; separate from CRUD to preserve JSON contract
3. **`ui.FileUpload` component** -- Props-based component following VideoPlayer pattern (`file_upload.go` + `file_upload_wasm.go` + `file_upload_stub.go`); SSR renders static drop zone, WASM adds interactivity via `OnMount`
4. **`fetch/upload.go`** -- New WASM upload function using FormData via `js.Value`; reuses existing CSRF token injection; does NOT modify existing `fetch.Fetch()`
5. **Code generation extensions** -- `"input": "file"` in `gux.config.json` triggers `ui.FileUpload` in admin forms, image preview in detail views, thumbnails in list views, and `WithFileField` in CRUD registration
6. **File serving handler** -- Auth-aware handler at `/__gux_api/files/` for local storage; S3 uses direct/presigned URLs

**Data flow:** User selects file -> WASM uploads via XHR to upload endpoint -> server validates and stores -> returns path -> component stores path in state -> form submit includes path as string in JSON -> CRUD saves string to DB.

See [ARCHITECTURE.md](ARCHITECTURE.md) for full component boundaries and build order.

### Critical Pitfalls

1. **fetch/ Body is string-only (P1)** -- Cannot send FormData for file uploads. Must add a new `fetch.Upload()` function that passes `js.Value` (FormData) directly to browser fetch, keeping file bytes in JS land. Do NOT base64-encode files into JSON strings.

2. **CSRF breaks with FormData (P2)** -- New upload path must call existing `getCSRFToken()` and set `X-CSRF-Token` header. Must NOT manually set `Content-Type` (browser must set multipart boundary). Test with CSRF enabled from day one.

3. **SSR panics on browser APIs (P3)** -- File upload component must use build-tag separation (`_wasm.go` / `_stub.go`). SSR renders static HTML drop zone; WASM adds interactivity via `OnMount`. Follow existing VideoPlayer pattern exactly.

4. **CRUD only accepts JSON (P4)** -- Do NOT modify CRUD to handle multipart. Use separate upload endpoint that returns file path, then include path in JSON CRUD payload. This is the clean architectural boundary.

5. **WASM memory pressure on large files (P7)** -- Never copy file bytes into Go WASM memory. Build FormData in JavaScript, pass as `js.Value`. Enforce configurable max file size in UI (default 10MB). Use `URL.createObjectURL` for previews, not `FileReader.readAsDataURL`.

See [PITFALLS.md](PITFALLS.md) for all 15 pitfalls with prevention strategies and phase-specific warnings.

## Implications for Roadmap

Based on dependency analysis across all research, the system should be built in 5 phases. Phases 1-4 deliver a complete MVP. Phase 5 is enhancement.

### Phase 1: Storage Foundation
**Rationale:** Everything depends on having a working storage backend and upload endpoint. This phase establishes the core abstraction that all other phases build on.
**Delivers:** Storage interface, local filesystem backend, upload HTTP handler, file serving handler, `App.SetStorage()` integration
**Addresses:** Storage abstraction (table stakes), file serving endpoint, unique filename generation, server-side validation
**Avoids:** P8 (leaky abstraction) by designing URL generation into the interface from the start; P10 (client-only validation) by building server-side validation first; P14 (route conflicts) by using `/__gux_api/` prefix convention
**Stack:** Go stdlib (`io`, `os`, `mime`, `net/http`), rs/xid

### Phase 2: WASM Upload Client
**Rationale:** The UI component (Phase 3) needs this to function. Must be built before the component but after the server endpoint exists to test against.
**Delivers:** `fetch/upload.go` with FormData support, CSRF token integration, upload progress via XHR callbacks
**Avoids:** P1 (string-only Body) by creating a new function, not modifying existing fetch; P2 (CSRF breaks) by reusing `getCSRFToken()` and NOT setting Content-Type manually; P7 (memory pressure) by keeping file bytes in JS land
**Stack:** `syscall/js` (Go stdlib)

### Phase 3: UI Component
**Rationale:** Depends on both server upload endpoint (Phase 1) and WASM upload client (Phase 2). This is the user-facing piece.
**Delivers:** `ui.FileUpload` component with click-to-browse, drag-and-drop, progress bar, image preview, current file display with remove
**Addresses:** All table stakes UI features (click-to-browse, drag-drop, progress, client validation, preview)
**Avoids:** P3 (SSR panics) by using build-tag separation from the start; P12 (preview memory) by using `createObjectURL`
**Stack:** `syscall/js`, existing `core.Node` system, `ui/` component patterns

### Phase 4: Code Generation and CRUD Integration
**Rationale:** Depends on the UI component existing. This is the integration layer that delivers the zero-config experience.
**Delivers:** `"input": "file"` and `"input": "image"` support in `gux.config.json`, generated admin forms with `ui.FileUpload`, generated detail views with image preview/download links, generated list views with thumbnails/icons, `WithFileField` CRUD option for auto-cleanup, `WithDeleteHook` for general-purpose cleanup
**Addresses:** CRUD integration (table stakes), zero-config file fields (key differentiator), lifecycle hooks, automatic orphan cleanup
**Avoids:** P5 (scalar-only codegen) by extending `generateFormFieldCode`; P9 (state explosion) by creating composite `FileFieldState`; P15 (delete orphans) by generating cleanup hooks; P4 (JSON-only CRUD) by using separate upload endpoint approach
**Stack:** Go `text/template` (existing), `gux.config.json` schema extension

### Phase 5: S3 Backend and Advanced Features
**Rationale:** Optional enhancement. Local filesystem is sufficient for development and simple deployments. S3 is needed for production scale.
**Delivers:** S3 storage backend, presigned URL support, multi-file field support, storage config in `gux.config.json`
**Addresses:** S3 storage (table stakes for production), multi-file fields, accept filter presets, storage config
**Avoids:** P6 (orphaned files) by implementing cleanup job alongside multi-file support; P13 (ordering) by using JSON array storage with explicit ordering
**Stack:** AWS SDK Go v2 (`aws-sdk-go-v2/service/s3`, `aws-sdk-go-v2/config`, `aws-sdk-go-v2/feature/s3/manager`)

### Phase Ordering Rationale

- **Phases 1-2-3 are a strict dependency chain:** Server endpoint -> WASM client -> UI component. Cannot be reordered.
- **Phase 4 depends on Phase 3** because code generation emits `ui.FileUpload` component calls. The component must exist first.
- **Phase 5 is independent of Phase 4** and could theoretically be built after Phase 1, but S3 is not needed for the MVP admin experience.
- **Phases 2 and 3 could be merged** if the team prefers fewer, larger phases. They are separated here because the WASM upload client needs careful testing with CSRF before the UI component adds complexity.
- **Critical architecture decisions must be made before Phase 1:** Separate upload endpoint (not multipart CRUD), storage interface design with URL generation, `/__gux_api/upload/` path convention.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 2 (WASM Upload Client):** XHR progress handling from Go/WASM is less documented than standard Fetch. Needs a spike/prototype to validate the `syscall/js` -> `XMLHttpRequest.upload.onprogress` -> Go callback chain works reliably. STACK.md rates this MEDIUM confidence.
- **Phase 5 (S3 Backend):** Presigned URL upload flow (client-side direct-to-S3) is a significant departure from the proxy upload pattern. Needs research on CORS configuration, multipart upload coordination, and how it interacts with the existing CSRF model.

Phases with standard patterns (skip research-phase):
- **Phase 1 (Storage Foundation):** Well-documented Go patterns for file upload handlers, multipart parsing, and storage interfaces. Direct codebase analysis confirms integration points.
- **Phase 3 (UI Component):** Follows established VideoPlayer component pattern exactly. Build-tag separation and Props struct are proven patterns in the codebase.
- **Phase 4 (Code Generation):** Extends existing `modelgen.go` templates with a new input type. Same pattern as adding `"input": "email"` or `"input": "select"`.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Only 2 new deps, both verified on pkg.go.dev with recent releases. Go stdlib covers everything else. |
| Features | HIGH | Cross-referenced against Django, Laravel, Rails, Payload CMS. Feature list validated by multiple framework precedents. |
| Architecture | HIGH | Based on direct codebase analysis of existing patterns (CRUD, VideoPlayer, fetch, codegen). Integration points verified in source. |
| Pitfalls | HIGH | 15 pitfalls identified from codebase analysis and domain research. Critical pitfalls (P1-P5) confirmed by reading actual source code. |

**Overall confidence:** HIGH

### Gaps to Address

- **XHR upload progress in Go/WASM:** MEDIUM confidence area. The `XMLHttpRequest.upload.onprogress` callback pattern from Go WASM is less documented than other syscall/js patterns. Build a minimal prototype in Phase 2 before committing to the full component.
- **Multi-file field UX:** The research covers single-file thoroughly but multi-file interaction patterns (add/remove/reorder in admin forms) need design work during Phase 5 planning.
- **Image thumbnail generation:** Deliberately deferred. When picked up, research `disintegration/imaging` (pure Go, no CGO) for resize/crop. This is a well-understood problem.
- **Orphan cleanup scheduling:** The research identifies the need for periodic cleanup of orphaned files but does not specify the scheduling mechanism (goroutine timer, cron job, on-demand). Decide during Phase 4 planning.

## Sources

### Primary (HIGH confidence)
- Gux codebase analysis: `core/crud.go`, `core/app.go`, `core/endpoint.go`, `core/csrf.go`, `fetch/fetch.go`, `ui/video_player.go`, `cmd/gux/modelgen.go`
- [AWS SDK Go v2 - S3](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3) -- v1.96.0+, Jan 2026
- [rs/xid](https://pkg.go.dev/github.com/rs/xid) -- v1.6.0, stable API
- [Payload CMS Upload Docs](https://payloadcms.com/docs/upload/overview) -- media collection pattern
- [Strapi Upload Lifecycle Issue #17170](https://github.com/strapi/strapi/issues/17170) -- BeforeDelete ordering bug

### Secondary (MEDIUM confidence)
- [Salt Design System - File Upload Pattern](https://www.saltdesignsystem.com/salt/patterns/file-upload) -- upload area + file list UX
- [S3 Upload Guide with SDK v2](https://www.buanacoding.com/2025/10/how-to-upload-files-to-aws-s3-in-go-with-sdk-v2.html) -- production patterns
- [Go WASM File Input](https://donatstudios.com/Read-User-Files-With-Go-WASM) -- browser File API via syscall/js
- [FormData + fetch multipart](https://muffinman.io/blog/uploading-files-using-fetch-multipart-form-data/) -- do not manually set Content-Type

### Tertiary (LOW confidence)
- XHR upload progress from Go/WASM -- limited documentation, needs prototype validation

---
*Research completed: 2026-02-01*
*Ready for roadmap: yes*
