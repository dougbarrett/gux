# Technology Stack: File Upload System

**Project:** Gux File Upload System
**Researched:** 2026-02-01
**Overall Confidence:** HIGH

## Executive Summary

The file upload system requires only **two new direct dependencies** (AWS SDK v2 for S3, rs/xid for file IDs). Everything else uses Go standard library or extends existing Gux infrastructure. The key architectural decision is to build a thin storage abstraction interface owned by Gux (not a third-party abstraction) with two backends: local filesystem (default, zero deps) and S3 (opt-in). On the WASM side, the upload component uses `XMLHttpRequest` via `syscall/js` rather than extending the existing `fetch` package, because the Fetch API does not support upload progress events.

## Recommended Stack

### Storage Abstraction (Server-Side)

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go `io` / `os` stdlib | Go 1.24.3 | Local filesystem storage | Zero dependencies. `io.Reader`/`io.Writer` interfaces are the foundation for any storage abstraction. Local FS is the default backend. |
| `aws-sdk-go-v2/service/s3` | v1.96.0+ | S3 storage backend | Official AWS SDK. V1 reached EOL July 2025 -- must use V2. Works with any S3-compatible service (MinIO, DigitalOcean Spaces, Backblaze B2, Cloudflare R2). |
| `aws-sdk-go-v2/config` | latest | AWS credential loading | Auto-discovers credentials from env, config files, IAM roles. Required companion to s3 service package. |
| `aws-sdk-go-v2/feature/s3/manager` | v1.21.0+ | Multipart upload/download | Handles chunked uploads for large files (>5MB), concurrent part uploads, automatic retry. v1.21.0 optimized memory for single uploads. |

**Why NOT MinIO Go SDK (`minio-go` v7.0.98):** While minio-go provides a simpler API, the AWS SDK v2 is the canonical S3 implementation. Using it means zero translation layer for AWS-native features (presigned URLs, SSE, lifecycle policies). Any S3-compatible provider works with the AWS SDK via custom endpoint configuration. Adding minio-go would be a second S3 client with no benefit.

**Why NOT third-party storage abstractions (`stow`, `go-storage`, `dstore`):** These add multi-cloud abstraction layers Gux does not need. The storage interface should be defined by Gux itself (a simple `Store` interface with `Put`/`Get`/`Delete`/`URL` methods). Two backends (local FS + S3) are sufficient. Adding a third-party abstraction constrains the interface design and adds dependency weight for no gain.

### File Naming / ID Generation

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `github.com/rs/xid` | v1.6.0 | Unique storage path generation | 20-char, URL-safe, sortable (time-ordered), no hyphens, all lowercase alphanumeric. Better than UUID for file paths: shorter, sortable, filename-safe. |

**Why NOT `google/uuid` (v1.6.0):** UUIDs are 36 chars with hyphens, not time-sortable. For storage paths, xid produces cleaner results (e.g., `cv1k7p6d7hs000a3q8cg.jpg` vs `550e8400-e29b-41d4-a716-446655440000.jpg`). Sortability aids debugging and log correlation.

**Why NOT `crypto/rand` + encoding:** Reimplementing ID generation is unnecessary when xid is battle-tested and purpose-built.

### Multipart Form Handling (Server-Side)

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go `mime/multipart` stdlib | Go 1.24.3 | Parse multipart file uploads | `http.Request.ParseMultipartForm()` and `http.Request.FormFile()` are the canonical approach. No library needed. |
| Go `net/http` stdlib | Go 1.24.3 | Request size limits, file serving | `http.MaxBytesReader` for upload size limits. `http.ServeFile` / `http.FileServer` for serving stored files. |
| Go `mime` stdlib | Go 1.24.3 | MIME type detection | `mime.TypeByExtension()` and `http.DetectContentType()` for file type validation (content sniffing). |
| Go `path/filepath` stdlib | Go 1.24.3 | Safe path construction | Path sanitization for local filesystem storage. Prevents directory traversal attacks. |

### Image Validation (Server-Side)

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go `image` stdlib | Go 1.24.3 | Image format detection, dimension reading | Already an indirect dependency (`golang.org/x/image` v0.24.0 in go.mod). Decode image headers for validation without loading full image into memory. |

**Thumbnail generation is DEFERRED.** Rationale: Thumbnails (resize, crop) are a significant feature that should be a separate milestone. For the initial upload system, image validation (format, dimensions) using stdlib is sufficient. When thumbnails are needed later, `disintegration/imaging` (v1.6.2, pure Go, no CGO) is the recommended choice.

**Why NOT `bimg` or `imagor`:** Both require libvips (CGO). Conflicts with Gux's simplicity goals and complicates cross-compilation.

### WASM File Upload (Client-Side)

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `syscall/js` (Go stdlib) | Go 1.24.3 | Browser File API access, drag-drop events, XMLHttpRequest for progress | The existing pattern in `fetch/fetch.go` and `ui/video_player_wasm.go` proves this approach. |

**Critical design decision: XMLHttpRequest, not Fetch API.** The browser Fetch API does not support upload progress events. `XMLHttpRequest.upload.onprogress` is the only way to track upload progress in the browser. The upload component uses XHR directly via `syscall/js` for progress tracking.

**The existing `fetch` package does NOT need modification.** Its `Body` field is `string`-typed (JSON-focused). File uploads use a separate code path (XHR via syscall/js in the upload component). This avoids breaking existing API contracts.

### CRUD Integration (Code Generation)

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go `text/template` stdlib | Go 1.24.3 | Generate upload endpoint code | Already used by `cmd/gux/` for code generation. File field support extends existing templates. |
| `gux.config.json` (existing) | N/A | Model field configuration | Add `"input": "file"` and `"input": "image"` field types. Consistent with existing `"input": "email"`, `"input": "select"`. |

No new dependencies for code generation.

## Complete New Dependencies

```bash
# Storage - S3 backend (opt-in, only needed if using S3)
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/aws/aws-sdk-go-v2/feature/s3/manager

# File naming (always needed)
go get github.com/rs/xid
```

**Total new direct dependencies: 2** (AWS SDK v2, xid).

**S3 dependencies are opt-in at runtime.** The storage interface is designed so local filesystem is the default with zero additional imports. Users who configure S3 pull in AWS SDK; those who use local storage do not.

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| S3 Client | AWS SDK Go v2 | MinIO Go SDK (minio-go v7) | Second S3 client adds no value; AWS SDK works with all S3-compatible services |
| Storage Abstraction | Custom `Store` interface | stow / go-storage / dstore | Unnecessary third-party layer; 2 backends don't justify it |
| File IDs | rs/xid | google/uuid | UUID is longer (36 vs 20 chars), not sortable, has hyphens |
| File IDs | rs/xid | crypto/rand + custom encoding | Reinventing what xid does well |
| Image Thumbnails | Defer to future milestone | disintegration/imaging now | Separate concern; complexity can wait |
| Image Thumbnails | N/A (deferred) | bimg / imagor (libvips) | CGO dependency conflicts with simplicity goals |
| WASM Upload | XMLHttpRequest via syscall/js | Extend fetch package with FormData | XHR provides upload progress events; Fetch API does not |
| Multipart Parsing | Go stdlib | third-party form parsers | Stdlib is complete for this use case |
| MIME Detection | `http.DetectContentType` + `mime.TypeByExtension` | filetype library | Stdlib content sniffing reads first 512 bytes, sufficient for common types |

## Integration Points with Existing Stack

### core/crud.go

The `CRUDModel` struct gains file field metadata:

```go
type CRUDModel struct {
    // ... existing fields ...
    FileFields []FileFieldConfig // NEW: which fields are file uploads
}

type FileFieldConfig struct {
    FieldName    string   // e.g., "Image", "Document"
    MaxSize      int64    // bytes
    AllowedTypes []string // e.g., ["image/jpeg", "image/png"]
    StoragePath  string   // prefix path in storage
}
```

CRUD create/update handlers switch from JSON-only to multipart/form-data when file fields are present. File deletion hook fires when models with file fields are deleted.

### core/app.go

The `App` struct gains a `store` field:

```go
type App struct {
    // ... existing fields ...
    store Storage // Storage backend (local FS default)
}
```

New configuration: `app.SetStorage(store)` or option in `core.New()`. Static file serving route auto-registered for local storage (e.g., `/uploads/` maps to upload directory).

### fetch/fetch.go

**No changes needed.** The upload component handles its own HTTP via XMLHttpRequest for progress support. The existing fetch package remains JSON-focused.

### cmd/gux/ (Code Generation)

- `gux.config.json` parser recognizes `"input": "file"` and `"input": "image"` field types
- Model generation: file fields become `string` in Go model (storing file path/URL)
- DTO generation: file fields include the resolved URL for reads
- Admin page generation: file fields render `ui.FileUpload` component
- API generation: endpoints with file fields use multipart instead of JSON

### ui/ (Component Library)

New `ui.FileUpload` component following existing patterns:

- Props struct pattern (like `ui.VideoPlayerProps`)
- SSR build: renders static `<input type="file">` for progressive enhancement
- WASM build: drag-drop via `syscall/js`, progress via XHR, file preview
- Platform split files: `file_upload.go`, `file_upload_wasm.go`, `file_upload_stub.go`

### GORM Schema

File metadata stored as simple string columns (path or URL), not separate tables:

```go
type Product struct {
    gorm.Model
    Name  string
    Image string // "uploads/cv1k7p6d7hs000a3q8cg.jpg" or full S3 URL
}
```

For multiple files per field, use a child model with existing CRUD parent-child pattern. Start with single-file fields only.

## What NOT to Add

| Do Not Add | Why |
|------------|-----|
| `tus` (resumable upload protocol) | Over-engineered for initial implementation. Standard multipart is sufficient. Add later if large file uploads become a requirement. |
| `blurhash` / placeholder generation | Nice-to-have, not table stakes. Defer to thumbnail milestone. |
| Virus scanning / file analysis | Enterprise feature. Document as a hook point (BeforeUpload) but do not implement. |
| Client-side image resizing | Adds WASM binary size. Server-side processing (deferred) is the right approach. |
| Presigned URL uploads (direct-to-S3) | Optimization for high traffic. Start with server-proxied uploads for simplicity. Can be added as enhancement later. |
| CDN integration | Deployment concern, not framework concern. Document URL rewriting pattern. |

## Sources

- [AWS SDK for Go v2 - S3 package](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3) - v1.96.0+, published Jan 28, 2026
- [AWS SDK for Go v2 - S3 Transfer Manager](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/feature/s3/manager) - v1.21.0, memory optimization for single uploads
- [AWS SDK for Go v2 - S3 Code Examples](https://docs.aws.amazon.com/code-library/latest/ug/go_2_s3_code_examples.html)
- [AWS SDK Go v1 EOL](https://github.com/aws/aws-sdk-go/releases) - V1 end-of-life July 31, 2025
- [MinIO Go SDK](https://github.com/minio/minio-go) - v7.0.98, Jan 2026 (considered, not recommended)
- [rs/xid](https://pkg.go.dev/github.com/rs/xid) - v1.6.0, Aug 2024
- [google/uuid](https://pkg.go.dev/github.com/google/uuid) - v1.6.0 (considered, not recommended)
- [disintegration/imaging](https://pkg.go.dev/github.com/disintegration/imaging) - v1.6.2 (deferred for future thumbnail milestone)
- [Go WASM File Input](https://donatstudios.com/Read-User-Files-With-Go-WASM) - Pattern for browser File API via syscall/js
- [Go Multipart Upload](https://www.slingacademy.com/article/using-multipart-requests-for-file-uploads-in-go/) - stdlib multipart handling patterns
- [S3 Upload Guide with SDK v2](https://www.buanacoding.com/2025/10/how-to-upload-files-to-aws-s3-in-go-with-sdk-v2.html) - Presigned URLs, multipart, production patterns

## Confidence Assessment

| Area | Confidence | Reason |
|------|------------|--------|
| AWS SDK v2 for S3 | HIGH | Official docs verified, pkg.go.dev version confirmed, actively maintained (Jan 2026 release) |
| rs/xid for file IDs | HIGH | pkg.go.dev verified, well-established, API is stable |
| Go stdlib for multipart/MIME | HIGH | Standard library, battle-tested, no version concerns |
| XMLHttpRequest for WASM upload progress | MEDIUM | Standard browser API, proven pattern, but Go/WASM XHR progress handling is less documented than Fetch. Needs validation during implementation. |
| Existing fetch package -- no changes | HIGH | Read source code directly; string-only Body field confirms separate upload path needed |
| GORM string columns for file paths | HIGH | Standard pattern, consistent with existing model/DTO architecture |
| Code generation integration | HIGH | Read existing cmd/gux/ tooling; extending config schema and templates is straightforward |
