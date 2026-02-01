---
phase: 30-upload-client-ui-component
plan: 01
subsystem: api
tags: [wasm, xhr, upload, csrf, javascript]

# Dependency graph
requires:
  - phase: 29-storage-foundation
    provides: Upload endpoint at /__gux_api/upload
provides:
  - WASM file upload client using XMLHttpRequest with progress tracking
  - UploadOptions API for configuring uploads with callbacks
  - UploadResult parsing from JSON response
  - Automatic CSRF token inclusion via getCSRFToken()
affects: [30-02-upload-ui-component, upload-ui]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "XHR-based upload with progress callbacks (Fetch API lacks upload progress)"
    - "File bytes stay in JavaScript heap (js.Value passed to FormData, no CopyBytesToGo)"
    - "WASM-only package pattern with build tags (no stub file needed)"

key-files:
  created:
    - fetch/upload.go
  modified: []

key-decisions:
  - "XMLHttpRequest chosen over Fetch API because Fetch lacks upload.onprogress events"
  - "No stub file created - fetch package is WASM-only (same pattern as fetch.go)"
  - "File bytes never copied to Go WASM memory - js.Value File passed directly to FormData"
  - "Progress listener attached before xhr.open() to avoid browser event firing bugs"

patterns-established:
  - "Upload pattern: UploadOptions with js.Value File, OnProgress callback, ExtraFields map"
  - "Response reuse: Upload returns fetch.Response type for consistency"
  - "Memory safety: All js.FuncOf callbacks released after completion"

# Metrics
duration: 8min
completed: 2026-02-01
---

# Phase 30 Plan 01: Upload Client Summary

**XMLHttpRequest-based WASM upload client with progress callbacks, automatic CSRF inclusion, and zero-copy file handling**

## Performance

- **Duration:** 8 min
- **Started:** 2026-02-01T23:23:00Z
- **Completed:** 2026-02-01T23:31:00Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments
- WASM upload client using XMLHttpRequest for progress tracking (Fetch API lacks this capability)
- File bytes remain in JavaScript memory - no CopyBytesToGo calls, js.Value passed to FormData
- Automatic CSRF token inclusion via existing getCSRFToken() from fetch package
- All js.FuncOf callbacks properly released to prevent memory leaks
- ParseUploadResult helper for JSON response parsing

## Task Commits

Each task was committed atomically:

1. **Task 1: Create fetch/upload.go (WASM upload with XHR progress)** - `9b5ba90` (feat)
2. **Task 2: Create fetch/upload_stub.go (non-WASM stub)** - `2507fe6` (docs)

## Files Created/Modified
- `fetch/upload.go` - WASM-only upload client with UploadOptions, Upload function, UploadResult parsing, XHR progress tracking, CSRF integration

## Decisions Made

**XMLHttpRequest over Fetch API**
- Fetch API does not support upload progress events (no upload.onprogress equivalent)
- XMLHttpRequest is the only browser API that provides upload progress tracking
- This is a web platform limitation, not a design choice

**No stub file needed**
- fetch package follows WASM-only pattern (fetch.go has `//go:build js && wasm` with no stub)
- ui/ component stubs in phase 30-02 will handle SSR case by not calling fetch.Upload
- Non-WASM builds correctly exclude entire fetch package

**File bytes stay in JavaScript**
- js.Value File object passed directly to FormData.append()
- No js.CopyBytesToGo calls anywhere in upload.go
- Prevents unnecessary memory allocation in WASM heap
- Files can be large (videos, images) - keeping them in JS avoids WASM memory pressure

**Progress listener timing**
- Attached before xhr.open() to avoid browser bugs where events don't fire
- Some browsers skip events if listeners registered after open()

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

**js.Func comparison bug**
- Initial code: `if progressCb.Value != js.Undefined()` failed (struct cannot be compared)
- Fix: Changed to `if opts.OnProgress != nil` to check if callback should be released
- Committed in 9b5ba90 (Task 1 commit)

## Next Phase Readiness

**Ready for Phase 30 Plan 02 (Upload UI Component):**
- fetch.Upload function available for WASM builds
- UploadOptions configures URL, file, field name, progress callback, extra fields
- UploadResult parses server JSON response (key, url, filename, size, content_type)
- CSRF automatically handled
- Response type matches fetch.Response for consistency

**No blockers or concerns.**

---
*Phase: 30-upload-client-ui-component*
*Completed: 2026-02-01*
