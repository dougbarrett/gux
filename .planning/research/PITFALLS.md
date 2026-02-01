# Domain Pitfalls: File Upload System for Gux Framework

**Domain:** File upload integration into Go/WASM framework with code generation, SSR, and storage abstraction
**Researched:** 2026-02-01
**Confidence:** HIGH (based on codebase analysis of fetch/, core/crud.go, core/csrf.go, core/endpoint.go, cmd/gux/modelgen.go + verified web research)

## Critical Pitfalls

Mistakes that cause rewrites or major breakage if not addressed upfront.

### P1: fetch/ Package Cannot Send FormData (Body Is string-Only)

**What goes wrong:** The existing `fetch/fetch.go` package has `Body string` in its `Options` struct (line 24). File uploads require sending `FormData` objects via the browser's Fetch API, which are not string-serializable. If you try to base64-encode files and send them as JSON strings, you double memory usage and hit WASM memory limits on large files.

**Why it happens:** The fetch package was designed for JSON API calls. It converts Go HTTP options into browser `fetch()` calls, but the entire abstraction assumes text bodies. `FormData` is a browser-native object that must be passed as a `js.Value` directly to `fetch()`, not as a string.

**Consequences:**
- File uploads silently fail or corrupt data
- Large files (>10MB) crash the WASM runtime from memory pressure
- Base64 encoding inflates file size by ~33%, wasting bandwidth and memory

**Warning signs:**
- Attempting to `json.Marshal` file bytes into the request body
- Creating workaround functions that bypass the fetch package entirely
- Inconsistent CSRF token handling between the workaround and the standard fetch path

**Prevention:**
- Extend the fetch package to accept `js.Value` as body (for FormData), not just `string`
- Add a dedicated `FetchFormData(url string, formData js.Value)` function or modify `Options` to support both string and js.Value bodies
- Ensure CSRF token injection still works when Content-Type is not manually set (browser sets multipart boundary automatically)
- Do NOT manually set `Content-Type: multipart/form-data` -- the browser must set it with the correct boundary parameter

**Detection:** If any code path sets `headers["Content-Type"] = "multipart/form-data"` manually, it is broken. The browser must auto-set the boundary.

**Phase:** Must be addressed in the first implementation phase (storage/upload infrastructure), before any UI work.

---

### P2: CSRF Token Injection Breaks with FormData Requests

**What goes wrong:** The current CSRF system works because `fetch/fetch.go` sets `Content-Type: application/json` by default (lines 177-178) and includes the CSRF token as the `X-CSRF-Token` header (line 106). When switching to `FormData` for file uploads, the `Content-Type` must NOT be set manually (the browser auto-sets it with the multipart boundary). If the fetch extension inadvertently sets or strips the Content-Type, uploads break with garbled multipart boundaries.

Additionally, the CSRF middleware (`core/csrf.go`) validates by comparing the `X-CSRF-Token` header against the cookie (line 138-156). This header-based approach works for FormData uploads without changes, BUT only if the new fetch function still reads the CSRF meta tag and includes the header. If a developer creates a new upload path that bypasses the existing `getCSRFToken()` function, uploads will get 403 Forbidden.

**Why it happens:** Developers often treat file upload as a special case and write separate HTTP code instead of extending the existing fetch abstraction. The CSRF token gets forgotten in the new code path.

**Consequences:**
- 403 Forbidden on all file uploads in production (CSRF rejection)
- Uploads work in dev (if CSRF is disabled for testing) but fail when deployed
- Debugging is confusing because the error message says "CSRF token invalid" not "missing header"

**Warning signs:**
- File uploads work with `app.DisableCSRF()` but fail without it
- New upload functions that don't call `getCSRFToken()`
- Tests that disable CSRF to make upload tests pass

**Prevention:**
- Extend the existing `fetch/fetch.go` rather than creating a parallel upload function
- Ensure the FormData fetch path calls `getCSRFToken()` and sets `X-CSRF-Token` header identically to the existing JSON path
- Write an integration test that uploads a file WITH CSRF enabled
- The header-based CSRF approach (Double Submit Cookie) works naturally with FormData -- do not switch to embedding tokens in form fields

**Phase:** Must be validated in the storage infrastructure phase with an end-to-end upload test.

---

### P3: SSR Renders File Upload Component Without Browser APIs

**What goes wrong:** Gux renders pages on the server first (SSR), then hydrates with WASM. A file upload component needs browser APIs (`File`, `FileReader`, `FormData`, drag-and-drop events) that do not exist during SSR. If the component tries to access `js.Global()` during SSR, the server panics.

**Why it happens:** Gux's existing components handle this correctly (e.g., `video_player_stub.go` for non-WASM builds, `script_loader_wasm.go` with build tags). But file upload is more complex because it has both a meaningful SSR representation (styled drop zone, file list) AND interactive behavior (drag events, file reading, upload progress). Developers may forget to separate the two layers.

**Consequences:**
- Server crashes on pages containing file upload components
- Hydration mismatch if SSR output differs from WASM render (different HTML structure)
- Event handlers attached to wrong elements after hydration

**Warning signs:**
- Build tag `//go:build js && wasm` on the file upload component file but no corresponding stub
- Direct `js.Global()` calls in the main component render function (not guarded by build tags)
- SSR test failures on pages that include file upload

**Prevention:**
- Follow the existing `video_player.go` / `video_player_wasm.go` / `video_player_stub.go` pattern:
  - `fileupload.go`: Props struct, SSR-safe render (static HTML for drop zone, no JS interop)
  - `fileupload_wasm.go`: Browser interactivity (File API, drag-and-drop event listeners, FormData construction)
  - `fileupload_stub.go` (if needed): No-op for non-WASM builds
- SSR render should output the visual drop zone with placeholder text ("Drag files here or click to upload") but no event handlers
- WASM hydration adds the interactive behavior via `OnMount`
- Test: render the component in a non-WASM test to verify no panics

**Phase:** UI component phase. Must be designed with build-tag separation from the start.

---

### P4: CRUD Endpoints Only Accept JSON -- No Multipart Support

**What goes wrong:** The existing CRUD `handleCreate` and `handleUpdate` in `core/crud.go` (lines 551-633, 635-760) use `json.NewDecoder(r.Body).Decode()` exclusively. They cannot handle `multipart/form-data` requests. If you try to add a file field to a CRUD model, the create/update endpoints will return "Invalid JSON" because the body is multipart, not JSON.

**Why it happens:** CRUD was designed for structured data (models with scalar fields). File upload introduces binary data that cannot be JSON-encoded efficiently. The two concerns (model field updates + file storage) need different transport mechanisms.

**Consequences:**
- Adding `"type": "file"` to a model in `gux.config.json` breaks the existing create/update flow
- Generated admin forms send multipart but CRUD expects JSON
- Workaround of separate upload endpoint creates two-request flow (upload file, then update model) with atomicity issues

**Warning signs:**
- "Invalid JSON" errors when submitting forms with file fields
- Hacks that separate file upload from model save into two API calls
- Race conditions where the model references a file that hasn't finished uploading

**Prevention:**
- Design a clear strategy upfront. Two viable approaches:

  **Option A (Recommended): Separate upload endpoint + JSON CRUD**
  - File upload goes to a dedicated `/__gux_api/upload` endpoint that returns a file URL/path
  - CRUD create/update still receives JSON, with the file field containing the URL string
  - The admin form uploads the file first, gets the URL, then submits the JSON form
  - Pro: Minimal changes to existing CRUD. Con: Requires handling orphaned uploads (see P6)

  **Option B: Multipart-aware CRUD**
  - Modify `handleCreate`/`handleUpdate` to detect `Content-Type: multipart/form-data` and parse accordingly
  - Extract file parts, upload to storage, replace with URLs, then process remaining fields as before
  - Pro: Single atomic request. Con: Significant changes to core CRUD code, harder to test

- Either way, codegen must generate the correct client-side logic for the chosen approach

**Phase:** Architecture decision required before any CRUD integration work. Should be settled in the storage infrastructure phase.

---

### P5: Code Generator Assumes All Fields Are Scalar Types

**What goes wrong:** The `generateFormFieldCode` function in `cmd/gux/modelgen.go` (line 2295) handles input types like "text", "textarea", "select", "checkbox", "number", "email", "password". It has no concept of a "file" input type. The generated admin forms will either skip file fields entirely or render them as text inputs.

Similarly, the DTO generation, detail page generation, and list page generation all assume fields are scalar values that can be JSON-serialized as strings, numbers, or booleans.

**Why it happens:** The code generator was built for database-backed scalar fields. File fields are fundamentally different: the model stores a URL/path string, but the form needs a file picker, the detail page needs a preview/download link, and the list page might need a thumbnail.

**Consequences:**
- Generated admin forms have text inputs where file pickers should be
- Detail pages show raw file URLs instead of previews or download links
- List pages show long URL strings in table cells
- Developers must manually override every generated page for models with file fields

**Warning signs:**
- `gux model regen` produces forms with `<input type="text">` for file fields
- Detail pages display `/uploads/abc123.pdf` as plain text
- Image fields show URLs instead of thumbnails

**Prevention:**
- Add `"input": "file"` and `"input": "image"` as recognized types in `generateFormFieldCode`
- For form generation: emit a `ui.FileUpload` component instead of `core.Input`
- For detail generation: emit an image preview for image fields, a download link for other file types
- For list generation: emit a thumbnail for images, a file icon for other types
- For DTO generation: file fields are `string` type (URL/path) -- no special handling needed
- The model field in GORM is also `string` -- it stores the file path/URL

**Phase:** Code generation phase, after the UI component and storage layer exist.

---

## Moderate Pitfalls

Mistakes that cause delays, tech debt, or poor UX if not addressed.

### P6: Orphaned Files When Upload Succeeds but Model Save Fails

**What goes wrong:** With the separate-upload-then-save approach (Option A from P4), a file gets uploaded to storage, then the model save fails (validation error, network issue, user navigates away). The file remains in storage with no database record pointing to it.

**Why it happens:** File storage and database writes are not transactional. You cannot atomically write to S3 and SQLite in one operation.

**Consequences:**
- Storage fills with orphaned files over time
- Storage costs increase for S3 deployments
- No way to distinguish "active" files from orphaned ones without scanning the database

**Prevention:**
- Implement a cleanup strategy from the start:
  1. Store uploaded files with a "pending" status or in a temporary path
  2. When the model save succeeds, move/mark the file as "active"
  3. Run a periodic cleanup job (e.g., delete pending files older than 24 hours)
- Alternative simpler approach: store the file reference in the model field only after successful upload, and run a daily background scan that removes files not referenced by any model
- The `BeforeDelete` lifecycle hook should delete associated files when the model is deleted
- Document this behavior so developers know to implement cleanup for custom upload flows

**Phase:** Storage infrastructure phase. The cleanup mechanism should be designed alongside the upload flow.

---

### P7: Memory Pressure from Large File Reads in WASM

**What goes wrong:** Reading a file in WASM via the JavaScript File API (`FileReader.readAsArrayBuffer`) loads the entire file contents into the WASM linear memory. For a 100MB video file, this means 100MB+ allocated in the WASM heap. Combined with the FormData copy, you effectively need 2x the file size in memory. WASM in browsers typically has a memory limit of 2-4GB, but practical limits are lower due to browser memory pressure.

**Why it happens:** The JavaScript File API is designed for browser-native code that manages memory efficiently. WASM has a single linear memory that grows but never shrinks. Go's garbage collector running in WASM is less efficient than native GC.

**Consequences:**
- Browser tab crashes on large file uploads (>50-100MB)
- Mobile browsers crash at even smaller sizes
- Memory never reclaims after upload completes (WASM linear memory does not shrink)

**Warning signs:**
- Tests only use small files (<1MB)
- No file size validation in the UI before starting the upload
- `FileReader.readAsArrayBuffer` used instead of streaming approaches

**Prevention:**
- Enforce a maximum file size in the UI component (configurable, default 10MB)
- For the JS interop layer, use `XMLHttpRequest` with `upload.onprogress` or `fetch` with a `ReadableStream` instead of reading the entire file into Go memory
- Keep file data in JavaScript land as much as possible. Build the FormData object in JavaScript, pass it to fetch as a js.Value, and never copy the file bytes into Go
- The Go WASM code should orchestrate the upload (construct FormData, call fetch) but not hold the file content in Go variables
- For the upload progress UI, use JavaScript event callbacks that send progress updates back to Go

**Phase:** UI component phase. The js.Value-based approach must be designed before implementing the file upload component.

---

### P8: Storage Interface Leaks Provider-Specific Behavior

**What goes wrong:** The storage abstraction (Local vs S3) works fine for simple Put/Get/Delete operations but breaks when provider differences surface: S3 requires pre-signed URLs for direct browser access, local storage uses file paths served by the Go HTTP server, S3 has eventual consistency for overwrites, local filesystem has immediate consistency.

**Why it happens:** Storage abstractions often target the lowest common denominator (upload/download/delete) but real applications need provider-specific features: pre-signed URLs, CDN integration, thumbnail generation hooks, content-type detection.

**Consequences:**
- Image preview works locally (direct file path) but breaks on S3 (needs signed URL or public bucket)
- File URLs stored in the database are provider-specific, making migration between providers impossible without a data migration
- Tests pass with local storage but fail in production with S3

**Warning signs:**
- File URLs in the database contain absolute filesystem paths (`/uploads/abc.jpg`)
- No URL generation method on the storage interface
- Tests only exercise the local storage adapter

**Prevention:**
- The storage interface must include a `URL(key string) string` method that returns an access URL appropriate for the provider
- Store only the storage key (relative path) in the database, never the full URL
- The URL method generates the appropriate access URL at runtime:
  - Local: `/uploads/{key}` (served by the Go HTTP server)
  - S3: pre-signed URL or CDN URL
- Include both Local and S3 adapters in tests (S3 can use a mock or MinIO in tests)
- Design the interface: `Upload(key string, reader io.Reader, contentType string) error`, `Delete(key string) error`, `URL(key string) string`

**Phase:** Storage infrastructure phase. The interface design is a critical early decision.

---

### P9: Generated Admin Form State Management for File Fields

**What goes wrong:** Existing admin form generation uses `r.StateString()` for text fields and `r.StateBool()` for booleans. A file field has multiple states: the selected file object (js.Value, WASM-only), the file name (string, for display), upload progress (float, 0-100), upload status (idle/uploading/done/error), and the resulting URL (string, for the model). Managing all of these with separate `r.StateString` calls creates verbose, error-prone generated code.

**Why it happens:** File upload has inherently more complex state than scalar form fields. The code generator creates per-field state variables, and for file fields this multiplies by 4-5x.

**Consequences:**
- Generated form code for models with file fields becomes extremely verbose
- State synchronization bugs (progress shows 100% but URL is still empty)
- Edit forms must handle "existing file" vs "new file" vs "remove file" states

**Warning signs:**
- 20+ state variables for a form with 3 file fields
- Inconsistent state during upload (file name shows but progress is at 0)
- Edit form does not show the existing file, only the upload control

**Prevention:**
- Create a `FileFieldState` composite state type that encapsulates all file-related state in one object
- The code generator should emit a single `fileState := UseFileState(r, "avatar")` instead of multiple state variables
- The `FileUpload` UI component should accept this composite state and manage its internal sub-states
- For edit forms, initialize `FileFieldState` with the existing file URL from the model data
- The composite state should expose: `FileName() string`, `URL() string`, `IsUploading() bool`, `Progress() float64`, `Error() string`

**Phase:** Code generation phase, designed alongside the UI component.

---

### P10: Content-Type Validation Only on Client Side

**What goes wrong:** The file upload UI validates file types (accept="image/*") but the server does not validate. An attacker can bypass the WASM UI entirely and send any file type to the upload endpoint, including executable files, HTML files (XSS vector), or SVG files with embedded JavaScript.

**Why it happens:** Client-side validation is for UX (immediate feedback). Developers assume it is also security. It is not.

**Consequences:**
- Stored XSS via uploaded SVG/HTML files served from the same origin
- Malicious files stored on the server
- If files are served with incorrect Content-Type or without Content-Disposition, browsers may execute them

**Warning signs:**
- File type validation only in the UI component props
- Server endpoint accepts any file without checking Content-Type or file magic bytes
- Uploaded files served without `Content-Disposition: attachment` header

**Prevention:**
- Server-side validation: check file extension, Content-Type header, AND file magic bytes (first few bytes of the file)
- Serve uploaded files with `Content-Disposition: attachment` for non-image types
- Serve images with explicit `Content-Type` (do not trust the uploaded Content-Type header)
- For SVG files specifically: either strip JavaScript/event handlers, or serve with `Content-Type: image/svg+xml` and `Content-Security-Policy: sandbox`
- The `BeforeUpload` lifecycle hook should be the place where server-side validation runs
- Configure allowed file types per model field in `gux.config.json` and enforce on the server

**Phase:** Storage infrastructure phase. Server-side validation must exist before any upload endpoint is exposed.

---

## Minor Pitfalls

Mistakes that cause annoyance but are fixable without major rework.

### P11: File Upload Progress Not Working During SSR Page Load

**What goes wrong:** If a user navigates to a page while a file is uploading in the background, the SSR page load replaces the DOM, and the upload progress indicator disappears. The upload continues in the background but the user has no visibility.

**Prevention:**
- File uploads should be page-independent (not tied to a specific component instance)
- Consider a global upload manager that persists across page navigations
- Or, simpler: disable navigation while uploads are in progress (show a warning dialog)

**Phase:** UI component phase, polish iteration.

---

### P12: Image Preview Generates Full-Size Thumbnails in Browser

**What goes wrong:** To show image previews before upload, developers read the full image with `FileReader.readAsDataURL` and display it in an `<img>` tag. For a 5MB photo, this creates a 5MB+ data URL string in memory, which combined with the original file data means ~15MB per image preview.

**Prevention:**
- Use `URL.createObjectURL(file)` instead of `FileReader.readAsDataURL` for previews
- `URL.createObjectURL` creates a reference to the file without copying its contents
- Call `URL.revokeObjectURL` in the `OnUnmount` callback to free the blob reference
- For server-side thumbnails (in list views), generate actual thumbnails on upload rather than serving full-size images

**Phase:** UI component phase.

---

### P13: Multi-File Field Ordering Not Preserved

**What goes wrong:** When a model has a multi-file field (e.g., "Product Images"), the order of files matters to the user but is not preserved. Files are stored as a JSON array of URLs, but upload order may differ from display order, and there is no drag-to-reorder in the admin UI.

**Prevention:**
- Store multi-file fields as a JSON array with explicit ordering (`["img1.jpg", "img2.jpg"]`)
- The admin UI should display files in array order and support reordering
- Do not use a separate join table for multi-file fields unless there is metadata per file (caption, alt text)
- Use GORM's JSON column type (the existing `StringList` type in modelgen.go is already suitable)

**Phase:** Multi-file support phase, after single file upload works.

---

### P14: Local Storage File Serving Conflicts with Route Patterns

**What goes wrong:** The local storage adapter serves files from a directory (e.g., `/uploads/`). If this path is registered as a static file handler, it may conflict with Gux's routing patterns, especially with the catch-all WASM routing that handles all paths for client-side navigation.

**Prevention:**
- Use a dedicated prefix for file serving (e.g., `/__gux_uploads/`) that matches the existing `/__gux_api/` convention
- Register the file serving handler before the catch-all WASM handler
- Ensure the file handler sets correct cache headers for uploaded files
- For S3, this is not an issue (files are served from a different domain)

**Phase:** Storage infrastructure phase.

---

### P15: Delete Model Does Not Delete Associated Files

**What goes wrong:** The existing `handleDelete` in `core/crud.go` (line 762) only calls `db.Delete(item)`. It does not know about file fields and will not delete the associated files from storage. Over time, storage fills with files from deleted records.

**Prevention:**
- The `BeforeDelete` or `AfterDelete` lifecycle hook must handle file cleanup
- The code generator should auto-generate delete hooks for models with file fields
- The hook should read the model's file field values before deletion and call `storage.Delete(key)` for each
- Handle errors gracefully: if file deletion fails, log the error but do not block the model deletion (prefer database consistency over storage consistency, then clean up orphans later)

**Phase:** CRUD integration phase.

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| Storage infrastructure | P1 (fetch Body is string-only), P2 (CSRF breaks), P8 (leaky abstraction) | Extend fetch package first; design storage interface with URL generation; test with CSRF enabled |
| Security | P10 (client-only validation) | Server-side file type validation from day one; BeforeUpload hook runs before storage |
| UI component | P3 (SSR panics), P7 (memory pressure), P12 (preview memory) | Build-tag separation; keep file bytes in JS land; use createObjectURL for previews |
| CRUD integration | P4 (JSON-only CRUD), P5 (scalar-only codegen), P15 (delete orphans) | Decide upload-then-save vs multipart-aware CRUD early; extend codegen for file input types |
| Code generation | P5 (no file input type), P9 (state explosion) | Composite FileFieldState; codegen emits file-specific form fields, detail views, list thumbnails |
| Multi-file support | P6 (orphaned files), P13 (ordering) | Cleanup job for orphans; JSON array storage with ordering |
| Polish | P11 (progress across navigation), P14 (route conflicts) | Dedicated upload URL prefix; navigation warnings during upload |

## Sources

- Codebase analysis: `fetch/fetch.go` (Body string limitation, CSRF injection), `core/csrf.go` (Double Submit Cookie pattern), `core/crud.go` (JSON-only create/update), `core/endpoint.go` (APIContext), `cmd/gux/modelgen.go` (form field code generation)
- [Uploading files using fetch and FormData](https://muffinman.io/blog/uploading-files-using-fetch-multipart-form-data/) - Do not manually set Content-Type
- [Fix uploading files using fetch and multipart/form-data](https://thevalleyofcode.com/fix-formdata-multipart-fetch/) - Browser must set boundary
- [Spring Security CSRF documentation](https://docs.spring.io/spring-security/reference/servlet/exploits/csrf.html) - Multipart + CSRF chicken-and-egg problem
- [Handling Large File Uploads in Go with AWS S3](https://dev.to/neelp03/handling-large-file-uploads-in-go-with-aws-s3-stream-like-a-pro-3dle) - Streaming uploads, memory management
- [Django Forum: Orphaned uploaded files](https://forum.djangoproject.com/t/correct-way-to-handle-orphaned-image-files-and-other-files/11903) - Cleanup strategies
- [Caspio Orphan File Cleanup](https://howto.caspio.com/files-and-images/orphan-file-cleanup/) - Background cleanup approach
- [Go Wiki: WebAssembly](https://go.dev/wiki/WebAssembly) - WASM limitations and fetch mapping
