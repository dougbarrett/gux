---
phase: 30
plan: 02
subsystem: ui
status: complete
completed: 2026-02-01

requires:
  - phase: 30
    plan: 01
    what: fetch.Upload with XHR progress

provides:
  - ui.FileUpload component
  - SSR/WASM split pattern for uploads
  - Accessible file upload UX
  - Client-side validation before upload
  - Image preview with object URLs

affects:
  - phase: 31
    what: Form field generator needs FileUpload option

tech-stack:
  added:
    - URL.createObjectURL for image previews
  patterns:
    - SSR/WASM split (main + stub + wasm)
    - Zero-copy file handling via js.Value
    - Window-level drag prevention (sync.Once)

key-files:
  created:
    - ui/file_upload.go
    - ui/file_upload_stub.go
    - ui/file_upload_wasm.go
  modified: []

decisions:
  - what: "Use Extra map for accept, disabled, for, aria-* attributes"
    why: "core.Attrs only has common fields, Extra handles HTML attributes"
    alternatives: ["Add fields to core.Attrs"]
    impact: "Consistent with existing UI components (select, input, etc.)"

  - what: "Store object URL cleanup functions in package-level map"
    why: "Need to revoke old URLs when replacing previews to prevent memory leaks"
    alternatives: ["Let browser handle cleanup on navigation"]
    impact: "Better memory management for long-lived SPAs"

  - what: "Window-level drag prevention uses sync.Once"
    why: "Multiple FileUpload instances on same page need shared prevention"
    alternatives: ["Per-instance handlers"]
    impact: "Prevents duplicate event listeners"

metrics:
  duration: 4m
  tasks: 2/2
  commits: 2
  files: 3
  lines_added: 746
---

# Phase 30 Plan 02: Upload UI Component Summary

**One-liner:** Accessible FileUpload component with SSR-safe drop zone, WASM drag-drop, client-side validation, image preview via object URLs, and XHR progress tracking.

## What Was Built

Created the ui.FileUpload component following the VideoPlayer SSR/WASM split pattern:

1. **ui/file_upload.go (main component, no build tag):**
   - FileUploadProps struct with Accept, MaxSize, UploadURL, callbacks
   - UploadResult struct (re-declared from fetch without js dependency)
   - FileUpload function renders static HTML: hidden input + label drop zone
   - Existing file display with preview (image thumbnail or filename)
   - Empty containers for preview, progress, error, status (WASM hydration points)
   - Helper functions: formatFileSize, isImageURL, acceptString
   - ARIA attributes via Extra map (aria-live, aria-atomic, aria-label)

2. **ui/file_upload_stub.go (non-WASM builds):**
   - createFileUploadInit returns nil (SSR no-op)

3. **ui/file_upload_wasm.go (WASM interactivity):**
   - Window-level drag prevention (sync.Once for multiple instances)
   - Drop zone drag-and-drop handlers (enter, over, leave, drop)
   - File input change handler
   - handleFileSelected: type/size validation before upload
   - showImagePreview: URL.createObjectURL with cleanup
   - uploadFile: calls fetch.Upload with OnProgress callback
   - showProgressBar: animated progress bar with ARIA attributes
   - showSuccess: filename/thumbnail + remove button
   - showError: validation/upload error display
   - updateStatus: screen reader announcements (aria-live="polite")
   - matchMimePattern: supports wildcard MIME patterns (image/*)

## User Experience

**Click to upload:**
1. User clicks drop zone (label wrapping hidden input)
2. Browser file picker opens
3. User selects file
4. File input change event fires → handleFileSelected

**Drag and drop:**
1. User drags file over drop zone
2. Visual feedback: border changes to blue (gux-upload-dragover class)
3. User drops file
4. Drop handler gets file from dataTransfer → handleFileSelected

**Validation:**
- Type mismatch: shows error immediately, no upload started
- Size exceeded: shows error immediately, no upload started
- Validation errors announced to screen readers (aria-live="assertive")

**Upload flow:**
1. Image files: thumbnail preview shown via URL.createObjectURL
2. Progress bar animates 0-100% as fetch.Upload reports progress
3. Screen reader: "Uploading: X%" (aria-live="polite")
4. Success: filename/thumbnail + remove button shown
5. Error: error message in red box + screen reader announcement

**Remove/replace:**
- Click remove button → clears input, preview, progress, error
- OnRemove callback fires (optional)
- Drop zone returns to empty state

## Technical Implementation

**SSR safety:**
- file_upload.go has NO build tag (compiles everywhere)
- Only static HTML rendered in SSR: div + input + label + empty containers
- OnMount = createFileUploadInit(id, props)
- Stub provides nil for createFileUploadInit on non-WASM builds

**WASM interactivity:**
- createFileUploadInit returns OnMount closure that:
  - Prevents window drag-drop (sync.Once)
  - Wires up input change listener
  - Wires up drop zone drag/drop listeners
  - Wires up remove button if existing file shown

**Zero-copy file handling:**
- File js.Value passed directly to fetch.Upload
- NO js.CopyBytesToGo anywhere
- File bytes stay in JavaScript heap
- FormData.append(field, jsFile) → XHR sends directly

**Memory management:**
- Object URLs stored in map with cleanup functions
- When replacing preview, old URL revoked via url.revokeObjectURL
- Prevents memory leaks in long-lived SPAs

**Accessibility:**
- Hidden file input (sr-only) with label association for keyboard access
- Drop zone has aria-label="File upload"
- Progress container: aria-live="polite" + aria-atomic="true"
- Error container: aria-live="assertive" + aria-atomic="true"
- Status container: sr-only + aria-live="polite" for announcements
- Progress bar: role="progressbar" + aria-valuenow

## Integration Points

**Uses from 30-01:**
- fetch.Upload(UploadOptions) with OnProgress callback
- fetch.Response with OK, Status, Body fields

**Provides for 31-XX (codegen):**
- ui.FileUpload component for generated forms
- Props: Accept, MaxSize, UploadURL, OnUploadComplete
- UploadResult with Key, URL, Filename, Size, ContentType

**Pattern established:**
- Same SSR/WASM split as VideoPlayer
- main file (no tag) + stub (non-WASM) + wasm (js && wasm)
- OnMount as split point
- Zero dependencies on WASM in main file

## Code Quality

**Testing:**
- go vet ./ui/ ✓
- go build ./ui/ ✓
- GOOS=js GOARCH=wasm go vet ./ui/ ✓
- GOOS=js GOARCH=wasm go build ./ui/ ✓
- No js.CopyBytesToGo ✓

**Accessibility:**
- Hidden input with label association ✓
- ARIA live regions for status changes ✓
- Progress bar with proper ARIA attributes ✓
- Screen reader announcements for all state changes ✓

**Error handling:**
- Type validation before upload
- Size validation before upload
- Network error display
- Server error display (non-2xx response)
- JSON parse error handling

## Deviations from Plan

None - plan executed exactly as written.

## Next Phase Readiness

**For 31-XX (Upload Codegen):**
- ✓ FileUpload component API stable
- ✓ UploadResult shape defined
- ✓ Callback pattern established (OnUploadComplete)
- ✓ Client-side validation pattern (Accept, MaxSize)

**Blockers:** None

**Concerns:** None

## Performance Notes

- Object URLs created/revoked efficiently (no memory leaks)
- Progress updates happen outside render cycle (direct DOM manipulation)
- Window-level drag prevention registered once (sync.Once)
- File upload runs in goroutine (non-blocking)

## Architecture Decisions

**Why Extra map instead of adding Attrs fields?**
- core.Attrs is minimal by design (common HTML attrs only)
- Extra map is established pattern (used in select, input, checkbox, etc.)
- Keeps Attrs struct from growing unbounded
- HTML spec has 100+ attributes - can't add them all

**Why package-level objectURLs map?**
- Multiple FileUpload instances might exist on same page
- Each needs cleanup when replacing preview
- Map allows cleanup of previous URL when showing new preview
- Prevents memory leaks from unreleased object URLs

**Why sync.Once for drag prevention?**
- Window-level listeners should only be registered once
- Multiple FileUpload instances share same window
- sync.Once prevents duplicate listeners
- Cleaner than tracking global state

## Files Changed

**Created (3):**
1. ui/file_upload.go (313 lines) - main component + types
2. ui/file_upload_stub.go (7 lines) - non-WASM stub
3. ui/file_upload_wasm.go (433 lines) - WASM interactivity

**Modified:** None

**Total:** 746 lines added

## Commits

1. ea0baf3 - feat(30-02): create FileUpload component (main + stub)
2. 0677023 - feat(30-02): add WASM interactivity for FileUpload

## Dependencies Met

✓ 30-01 complete (fetch.Upload with progress)
✓ core.Attrs Extra map pattern
✓ VideoPlayer SSR/WASM split pattern reference

## Testing Notes

**Manual testing needed:**
1. Click drop zone → file picker opens
2. Drag file over zone → border turns blue
3. Drop file → upload starts, progress bar animates
4. Select image → thumbnail preview shown
5. Select oversized file → error shown, no upload
6. Select wrong type → error shown, no upload
7. Successful upload → filename + remove button shown
8. Click remove → all cleared, back to empty state
9. Screen reader: verify all status changes announced

**Browser compatibility:**
- URL.createObjectURL: all modern browsers
- XMLHttpRequest upload.onprogress: all modern browsers
- FormData: all modern browsers
- Drag and drop API: all modern browsers

## Future Enhancements

Not in scope for v2.4, but possible future additions:

- Multiple file selection (Accept multiple files)
- Drag-drop file list preview
- Upload queue with cancel
- Resumable uploads
- Chunked upload for large files
- Image cropping before upload
- Webcam capture integration

---

**Status:** Complete
**Duration:** 4 minutes
**Commits:** 2
**Files:** 3 (all created)
**Lines:** 746 added

Next: Phase 31 - Upload Codegen Integration
