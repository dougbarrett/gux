---
phase: 29-storage-foundation
plan: 01
subsystem: storage
tags: [storage, filesystem, content-addressing, validation, image-processing]
requires: []
provides:
  - Storage interface for pluggable backends
  - LocalStorage with content-hash filenames
  - MIME type validation with magic bytes
  - Image dimension detection
  - Functional options pattern
affects:
  - 29-02 (file serving endpoint uses Storage.Serve)
  - 30-* (upload UI uses Storage.Put)
  - 31-* (codegen uses Storage interface)
tech-stack:
  added:
    - github.com/gabriel-vasile/mimetype
  patterns:
    - Functional options (WithMaxSize, WithAllowedTypes, etc.)
    - Content-addressed storage (SHA-256 hash filenames)
    - Idempotent operations (Delete returns nil if not found)
key-files:
  created:
    - core/storage.go
    - core/storage_local.go
    - core/storage_test.go
  modified: []
key-decisions:
  - decision: Content-addressed storage with hash-prefix subdirectories
    rationale: Deduplicates identical files, distributes I/O, prevents filename collisions
    date: 2026-02-01
  - decision: Magic bytes MIME detection instead of trusting file extensions
    rationale: Security - prevents malicious files disguised with wrong extensions
    date: 2026-02-01
  - decision: Separate Serve method on Storage interface
    rationale: Enables efficient file serving (io.ReadSeekCloser for Range requests) without loading entire file into memory
    date: 2026-02-01
duration: 2m 14s
completed: 2026-02-01
---

# Phase 29 Plan 01: Storage Interface Summary

**One-liner:** Content-addressed storage abstraction with local filesystem backend, magic bytes validation, and image dimension detection

## Performance

- **Duration:** 2 minutes 14 seconds
- **Tasks:** 3/3 completed
- **Files:** 3 created (storage.go, storage_local.go, storage_test.go)
- **Tests:** 12 test functions, 100% pass rate
- **Lines Added:** ~616 total (102 interface, 188 impl, 326 tests)

## Accomplishments

✅ **Storage Interface** - Defined pluggable Storage interface with Put/Delete/URL/Serve methods, enabling future cloud backends (S3, GCS) without application code changes

✅ **UploadResult Metadata** - Rich upload metadata includes key, URL, filename, size, content type, and optional image dimensions (width/height)

✅ **Functional Options** - Implemented options pattern (WithMaxSize, WithAllowedTypes, WithServeAuth, WithBaseURL) for flexible storage configuration

✅ **LocalStorage Backend** - Complete local filesystem implementation with:
  - Content-addressed filenames (SHA-256 hash)
  - Hash-prefix subdirectories (e.g., ab/abc123def.jpg) for I/O distribution
  - Atomic writes (temp file + rename)
  - Magic bytes MIME detection via mimetype library
  - Image dimension detection (image.DecodeConfig)
  - Idempotent delete operations

✅ **Validation** - Size limits (WithMaxSize) and MIME type restrictions (WithAllowedTypes with wildcard support like "image/*")

✅ **Comprehensive Tests** - 12 test functions covering:
  - Constructor and options
  - Basic file upload with metadata verification
  - Image dimension detection
  - Content deduplication (same content = same key)
  - Size limit enforcement
  - MIME type filtering (allowed/rejected)
  - Delete operations (success and not found)
  - URL generation
  - MIME pattern matching

## Task Commits

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Storage interface, types, and functional options | 65746ff | core/storage.go |
| 2 | LocalStorage implementation with content hashing and validation | 0d944f2 | core/storage_local.go, go.mod, go.sum |
| 3 | Storage unit tests | a93c831 | core/storage_test.go |

## Files Created

1. **core/storage.go** (102 lines)
   - Storage interface: Put, Delete, URL, Serve
   - UploadResult struct with metadata fields
   - StorageOption functional options
   - StorageError type
   - matchMimePattern helper

2. **core/storage_local.go** (188 lines)
   - LocalStorage struct with config fields
   - NewLocalStorage constructor
   - Put: content hashing, prefix subdirs, validation, atomic writes
   - Delete: idempotent removal with parent cleanup
   - URL: serving URL generation
   - Serve: open file for serving with ReadSeekCloser

3. **core/storage_test.go** (326 lines)
   - 12 table-driven test functions
   - Minimal 1x1 PNG test fixture
   - Tests for constructor, upload, validation, delete, URL, pattern matching

## Files Modified

None - all new files.

## Decisions Made

### Content-Addressed Storage
**Decision:** Use SHA-256 hash of file content as filename, with first 2 characters as subdirectory prefix

**Rationale:**
- Automatic deduplication: same content uploaded twice = single stored file
- Prevents filename collisions (cryptographic hash uniqueness)
- Distributes I/O across subdirectories (avoids large directory performance issues)
- Enables verification of content integrity

**Implementation:** `Put` computes SHA-256, creates key as `prefix/hash+ext` where prefix = hash[:2]

**Tradeoffs:** Original filename preserved only in metadata, not on disk

### Magic Bytes MIME Detection
**Decision:** Use github.com/gabriel-vasile/mimetype for content-based MIME detection instead of trusting file extensions

**Rationale:**
- Security: prevents malicious files disguised with wrong extensions (e.g., PHP script named .jpg)
- Accuracy: detects actual content type regardless of user-provided filename
- Validation: ensures WithAllowedTypes restrictions can't be bypassed

**Implementation:** `Put` calls mimetype.Detect(buf) before validation

**Tradeoffs:** Requires reading entire file into memory (already needed for hashing)

### Serve Method with ReadSeekCloser
**Decision:** Add Serve(key) method to Storage interface returning io.ReadSeekCloser

**Rationale:**
- Enables efficient HTTP Range request support (video seeking, download resume)
- Avoids loading entire file into memory for serving
- Provides os.FileInfo for Content-Length and Last-Modified headers
- Allows streaming large files

**Implementation:** LocalStorage.Serve opens os.File (implements ReadSeekCloser) and returns Stat

**Next Phase:** Plan 29-02 file serving handler will use this method

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## Next Phase Readiness

**Blockers:** None

**Concerns:** None

**Ready for Plan 29-02:** ✅ Yes
- Storage interface complete and tested
- LocalStorage.Serve method ready for file serving endpoint
- All verification criteria met

**Next Steps:**
1. Plan 29-02: File serving endpoint with Range request support
2. Plan 30-*: Upload UI components using Storage.Put
3. Plan 31-*: Codegen for model file field scaffolding
