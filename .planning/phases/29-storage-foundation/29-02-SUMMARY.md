---
phase: 29
plan: 02
subsystem: http
tags: [upload, file-serving, storage, http, auth, csrf]
requires: [29-01-storage-interface]
provides: [upload-endpoint, file-serving-endpoint, storage-app-integration]
affects: [30-upload-ui, 31-codegen-integration]
tech-stack:
  added: []
  patterns: [multipart-form-handling, content-disposition, range-requests]
key-files:
  created:
    - core/upload.go
    - core/upload_test.go
  modified:
    - core/app.go
key-decisions:
  - upload-auth-protected-by-default
  - file-serving-public-by-default
  - csrf-handled-by-middleware
  - path-traversal-protection
metrics:
  duration: 4m 34s
  completed: 2026-02-01
---

# Phase 29 Plan 02: Upload Endpoint and File Serving Summary

**One-liner:** POST multipart upload at `/__gux_api/upload` with auth/CSRF protection, GET serving at `/__gux_api/files/{key}` with Range support and aggressive caching.

## Performance

**Duration:** 4m 34s
**Tasks:** 3/3 completed

## Accomplishments

Wired storage into the HTTP layer with secure-by-default patterns:

1. **App integration** - `SetStorage(s Storage)` method configures storage, `Handler()` registers endpoints when storage is set
2. **Upload endpoint** - `POST /__gux_api/upload` processes multipart forms, enforces auth (when configured), returns JSON array of `UploadResult` metadata
3. **File serving** - `GET /__gux_api/files/{key}` serves files with proper Content-Type, Cache-Control (1 year immutable), Content-Disposition (inline for images/videos/PDFs, attachment for others)
4. **Security** - Auth check when `authConfig` set, CSRF handled by existing middleware, path traversal protection on serving
5. **Integration tests** - 14 test cases covering upload/serving/auth/CSRF/path-traversal with no regressions

## Task Commits

| Task | Type | Hash | Description |
|------|------|------|-------------|
| 1 | feat | 45bb435 | Add storage field to App and SetStorage method |
| 2 | feat | 238d79d | Implement upload endpoint and file serving handler |
| 3 | test | b66d000 | Add upload and serving integration tests |

## Files Created

- **core/upload.go** (181 lines) - `registerUploadHandlers`, `handleUpload`, `handleServeFile`, `isInlineType`
- **core/upload_test.go** (531 lines) - 14 integration tests using httptest

## Files Modified

- **core/app.go** (+23 lines) - Added `storage Storage` field, `SetStorage`/`Storage` methods, `registerUploadHandlers` call in `Handler()`

## Decisions Made

### Upload endpoint is protected by default (when auth configured)

**Context:** Upload is a mutating operation that creates files.

**Decision:** Check `a.authConfig != nil` and require authenticated session via `getUserFromRequest(r)`. Return 401 if no session.

**Rationale:** Secure-by-default pattern matches CRUD/API endpoints. Prevents anonymous uploads in auth-enabled apps. Developers can disable auth app-wide if needed (public file uploads).

**Alternative rejected:** Always public - would allow spam uploads in auth-enabled apps.

### File serving is public by default

**Context:** Serving files at `/__gux_api/files/{key}` could expose private content.

**Decision:** Public by default. Developers opt into auth via `WithServeAuth()` option when constructing `LocalStorage`.

**Rationale:** Most file storage use cases are public assets (images, PDFs). Content-addressed keys are unguessable (SHA-256 hashes), providing security through obscurity for sensitive files. Opt-in auth for explicit privacy requirements.

**Implementation:** Type assertion to `*LocalStorage` to check `serveAuth` field. If true AND `authConfig` set, check session.

**Alternative rejected:** Protected by default - would break common use case of public CDN-style file serving.

### CSRF handled by existing middleware

**Context:** Upload is a POST endpoint and needs CSRF protection.

**Decision:** CSRF validation is already applied via `CSRFMiddleware` wrapping the mux in `Handler()`. Upload handler does not need custom CSRF logic.

**Rationale:** Middleware approach ensures all POST/PUT/PATCH/DELETE endpoints are protected consistently. Upload endpoint is registered after CRUD handlers, so it's inside the middleware chain.

**Verified:** `TestUpload_CSRFProtection` confirms upload fails without token, succeeds with token.

### Path traversal protection on serving endpoint

**Context:** User-provided `key` could contain `..` or `/` to access files outside storage directory.

**Decision:** Reject keys containing `..` or starting with `/` with 400 Bad Request before calling `storage.Serve()`.

**Rationale:** Defense-in-depth. Even though `filepath.Join` in `LocalStorage.Serve` prevents traversal, the HTTP handler should validate input. Go's ServeMux normalizes URL paths, so `/__gux_api/files/../../etc/passwd` becomes `/__gux_api/etc/passwd` and won't match the handler pattern (returns 404). The check catches keys like `aa/../bb.txt` that pass ServeMux but contain traversal.

**Test coverage:** `TestServeFile_PathTraversal` validates both cases.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

### Test fixes required for edge cases

**Issue 1:** `TestUpload_NoFiles` initially sent an empty POST body, which triggered `ParseMultipartForm` error (invalid_request) instead of "no_files" error.

**Fix:** Create valid multipart form with no files. `ParseMultipartForm` succeeds, then "no_files" check fires.

**Issue 2:** `TestServeFile_PathTraversal` expected 400 for `/__gux_api/files/../../../etc/passwd`, but Go's ServeMux normalizes the path before routing, so it became `/__gux_api/etc/passwd` and returned 404 (no match).

**Fix:** Updated test to accept both 404 (normalized by ServeMux) and 400 (caught by our check). Added second test case with URL encoding to verify the check works when the pattern matches.

**Issue 3:** `TestUpload_WithAuth` created session with `maxAge: 0`, causing immediate expiration. `getUserFromRequest` found expired session and returned nil.

**Fix:** Changed `sessionStore.Set(sessionID, user, 0)` to `sessionStore.Set(sessionID, user, 24*time.Hour)`.

All test fixes were standard debugging - no design changes required.

## Next Phase Readiness

### Ready for Phase 30 (Upload UI)

- [x] Upload endpoint accepts multipart forms and returns `UploadResult[]` JSON
- [x] File serving works with correct headers for client-side rendering
- [x] Auth/CSRF integration tested and working
- [x] Storage interface proven via integration tests

### Blockers/Concerns

None. All must-have truths validated:

- ✅ Developer calls `app.SetStorage()` to configure storage
- ✅ POST to `/__gux_api/upload` returns JSON array of `UploadResult` metadata
- ✅ Multiple files in single request each get their own `UploadResult`
- ✅ Upload endpoint rejects oversized files with `StorageError` JSON
- ✅ Upload endpoint rejects disallowed file types with `StorageError` JSON
- ✅ Files served at `/__gux_api/files/{key}` with correct Content-Type and caching
- ✅ Images/PDFs serve inline; other types trigger download
- ✅ Upload endpoint enforces CSRF when enabled
- ✅ Upload endpoint requires authentication when auth enabled
- ✅ File serving public by default; requires auth when `WithServeAuth()` configured

### Open Questions for Phase 30

- **XHR progress events:** Plan references XHR for upload progress (not Fetch API). Need to validate Go/WASM syscall/js can bind to `upload.onprogress` events.
- **Multi-file UI pattern:** Single file picker vs multiple file picker vs drag-and-drop. Research needed for common UX patterns.

---

**Phase Status:** 29 Storage Foundation complete (2/2 plans). Ready to proceed to Phase 30 Upload UI.
