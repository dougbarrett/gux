# Architecture Research: File Upload System for Gux

**Project:** File upload system integrated with existing Gux framework
**Researched:** 2026-02-01
**Confidence:** HIGH (based on thorough codebase analysis of existing patterns)

## Executive Summary

This research analyzes how a file upload system integrates with Gux's existing architecture: the CRUD system (`core/crud.go`), code generation (`cmd/gux/modelgen.go`), UI component library (`ui/`), config-driven scaffolding (`gux.config.json`), and the WASM fetch client (`fetch/`). The upload system touches every layer of the framework and must maintain consistency with established patterns.

**Key insight:** The existing CRUD system uses JSON request/response throughout (`json.NewDecoder(r.Body).Decode`). File uploads require multipart form handling, which is a fundamentally different content type. The architecture must bridge this gap cleanly -- either by adding a parallel upload endpoint alongside CRUD, or by extending CRUD to detect file fields and switch content types. The parallel endpoint approach is cleaner because it avoids breaking the existing JSON-based CRUD contract.

**Second key insight:** The code generation pipeline (`gux gen` / `gux model regen`) already handles field types like `string`, `*uint`, `bool`, `[]string`, and input types like `email`, `textarea`, `select`. Adding `"input": "file"` (or a new `"type": "file"`) to `gux.config.json` and teaching the code generator to produce upload-aware forms, DTOs, and API clients is the natural integration path.

---

## Existing Architecture Analysis

### Layer Map

```
gux.config.json          Config layer: model definitions with field types
        |
cmd/gux/modelgen.go      Code generation: models, DTOs, admin pages, hooks
        |
   guxgen/                Generated output: api/, dto/, models/, admin/
        |
   core/                  Runtime: App, CRUD handlers, Router, Node system
        |
   ui/                    Components: Input, Select, FormField, etc.
        |
   fetch/                 WASM HTTP client with CSRF
```

### How CRUD Currently Works (Integration Points)

1. **Config** (`gux.config.json`): Defines model fields with `type`, `input`, `required`, `table`, etc.
2. **Code gen** (`modelgen.go`): Reads config, generates GORM models, DTOs, admin CRUD pages, API clients
3. **Runtime** (`crud.go`): `app.CRUD(model{})` registers REST endpoints at `/__gux_api/crud/{plural}`
4. **Create flow**: Client sends JSON body -> `json.NewDecoder(r.Body).Decode(&data)` -> `db.Create()`
5. **Update flow**: Client sends JSON body -> decode -> merge with existing -> `db.Save()`
6. **Generated admin forms**: Use `ui.Input`, `ui.Select`, `ui.Textarea` etc., submit via `api.Models.Create(data, callback)`
7. **Generated API client** (WASM): Uses `fetch.Fetch()` with JSON body and auto CSRF headers

### Key Patterns to Follow

| Pattern | Where Used | How Upload Adapts |
|---------|-----------|-------------------|
| `Props` struct per component | `ui/input.go`, `ui/select.go` | `ui.FileUpload` gets `FileUploadProps` |
| `core.Attrs` for HTML attributes | All elements | Upload needs `type="file"`, `accept`, `multiple` attrs |
| State binding via `StringState` | `ui/input.go` Bind field | File state needs different type (not string) |
| `CRUDOption` functional options | `core/crud.go` | `WithFileField("Avatar", storage.Config{})` |
| Generated admin form fields | `modelgen.go` AdminNew/AdminEdit templates | Template detects `"input": "file"` and emits `ui.FileUpload` |
| `ModelField.Input` string | `model.go` | New value: `"file"` or `"image"` |
| `fetch.Fetch()` for WASM HTTP | `fetch/fetch.go` | Need `fetch.Upload()` for multipart from WASM |
| `core.API()` typed endpoints | `endpoint.go` | Upload endpoint uses `*http.Request` directly for multipart |

---

## Recommended Architecture

### New Package: `storage/`

Lives at the framework level alongside `core/`, `ui/`, `fetch/`, `api/`.

```
goquery/
  storage/
    storage.go       # Storage interface, Config, FileInfo types
    local.go         # LocalStorage implementation (disk)
    s3.go            # S3Storage implementation (AWS/MinIO)
    middleware.go     # HTTP handler for serving files (with auth)
```

**Why a separate package (not in core/):** The storage package has external dependencies (AWS SDK for S3). Keeping it separate means `core/` stays dependency-light. Apps that only use local storage don't pull in AWS SDK. This follows the same pattern as `fetch/` being separate from `core/`.

### Storage Interface

```go
package storage

type Storage interface {
    Upload(ctx context.Context, path string, reader io.Reader, opts UploadOptions) (*FileInfo, error)
    Delete(ctx context.Context, path string) error
    URL(path string) string           // Public URL or serving path
    SignedURL(path string, ttl time.Duration) (string, error)  // For private files
}

type FileInfo struct {
    Path        string    // Storage path: "uploads/users/123/avatar.jpg"
    Filename    string    // Original filename: "photo.jpg"
    Size        int64     // Bytes
    ContentType string    // MIME type
    URL         string    // Serving URL
}

type UploadOptions struct {
    MaxSize      int64    // Max file size in bytes
    AllowedTypes []string // Allowed MIME types: ["image/jpeg", "image/png"]
    Directory    string   // Upload subdirectory: "avatars"
}

type Config struct {
    Provider    string            // "local" or "s3"
    Local       *LocalConfig
    S3          *S3Config
}
```

### Integration with core.App

```go
// In core/app.go - add storage field
type App struct {
    // ... existing fields
    storage    interface{}  // storage.Storage, kept as interface{} to avoid import cycle
}

// New method
func (a *App) SetStorage(s interface{}) { a.storage = s }
func (a *App) GetStorage() interface{} { return a.storage }
```

**Why `interface{}` not `storage.Storage`:** Same pattern as `db` field. Avoids `core` depending on `storage`. The code generator and CRUD system use type assertions at runtime.

### Upload Endpoint Architecture

**Separate endpoint, not inside CRUD.** The CRUD endpoints stay JSON-only.

```
POST /__gux_api/upload/{model}/{field}        # Upload a file
DELETE /__gux_api/upload/{model}/{field}/{id}  # Delete a file
GET /__gux_api/files/{path...}                # Serve uploaded files
```

**Why separate from CRUD:**
1. CRUD uses `json.NewDecoder` everywhere -- multipart would require conditional content-type detection
2. File uploads are often done before the record is saved (upload, get URL, then submit form with URL)
3. Separate endpoint allows progress tracking, chunked uploads, direct-to-S3 presigned URLs later
4. Admin forms can upload the file first, receive a path/URL, then include that path in the JSON create/update payload

### Data Flow: Upload in Admin Form

```
1. User selects file in FileUpload component
2. WASM: fetch.Upload("/__gux_api/upload/users/avatar", file)
   -> multipart POST with CSRF token
3. Server: parse multipart, validate, store via storage.Storage
4. Server returns: { "path": "uploads/users/avatar_abc123.jpg", "url": "/files/uploads/..." }
5. Component stores returned path in state
6. User clicks Save
7. Form submits JSON including { "avatar": "uploads/users/avatar_abc123.jpg" }
8. Normal CRUD create/update flow stores the path string in the DB
```

This means the model field is just a `string` in the database (storing the file path), and the "file" nature is a UI/upload concern, not a schema concern.

### Model Field in gux.config.json

```json
{
  "name": "Avatar",
  "type": "string",
  "input": "file",
  "fileConfig": {
    "accept": "image/*",
    "maxSize": "5MB",
    "directory": "avatars"
  }
}
```

**Why `"type": "string"` not `"type": "file"`:** The database column stores a string (the file path). The `"input": "file"` tells the code generator to render a `ui.FileUpload` component instead of `ui.Input`, and to generate upload endpoint wiring. This is consistent with how `"type": "string", "input": "textarea"` already works -- the Go type is string, but the UI representation differs.

---

## Component Architecture

### New UI Component: `ui.FileUpload`

```
ui/
  file_upload.go       # FileUpload component (SSR + shared logic)
  file_upload_wasm.go  # WASM-specific: actual upload via fetch, progress
  file_upload_stub.go  # Non-WASM stub (SSR placeholder)
```

Follows the same platform-split pattern as `ui/video_player.go` / `ui/video_player_wasm.go` / `ui/video_player_stub.go`.

**Props pattern (consistent with existing components):**

```go
type FileUploadProps struct {
    ID          string
    Name        string
    Accept      string       // MIME types: "image/*", ".pdf,.doc"
    MaxSize     int64        // Max bytes
    Multiple    bool         // Allow multiple files
    Value       string       // Current file path (for edit forms)
    PreviewURL  string       // Current file preview URL (for images)
    Disabled    bool
    Required    bool
    Error       string
    Class       string
    OnUpload    func(FileInfo)       // Called after successful upload
    OnRemove    func()               // Called when user removes file
    UploadURL   string               // Override upload endpoint
}
```

**SSR rendering:** Shows a styled file input or a link to the existing file. No upload capability (forms submit via WASM).

**WASM rendering:** Full upload UI with drag-and-drop, progress bar, preview. Uses `fetch.Upload()` to POST multipart data.

### New Fetch Function: `fetch.Upload`

```
fetch/
  fetch.go           # Existing: Fetch() for JSON requests
  upload.go          # New: Upload() for multipart/FormData
  upload_stub.go     # Non-WASM stub
```

The existing `fetch.Fetch()` sets `Content-Type: application/json` and sends string bodies. File uploads need `FormData` via the browser's `fetch()` API. This requires a new function, not an extension of the existing one.

```go
// In WASM, uses JavaScript FormData and fetch()
func Upload(url string, file js.Value, callback func(*UploadResponse, error)) {
    // Creates FormData, appends file
    // Reads CSRF token (same as existing getCSRFToken())
    // Calls fetch() with FormData body (no explicit Content-Type - browser sets boundary)
}
```

---

## Code Generation Changes

### What `gux gen` / `gux model regen` Must Generate

When a field has `"input": "file"`:

| Generated File | What Changes |
|---------------|-------------|
| `guxgen/models/{model}_gen.go` | Field is `string` (no change from regular string) |
| `guxgen/dto/{model}_gen.go` | Field is `string` with `json:"avatar"` (no change) |
| `guxgen/admin/{model}_new_gen.go` | Emits `ui.FileUpload` instead of `ui.Input` for this field |
| `guxgen/admin/{model}_edit_gen.go` | Emits `ui.FileUpload` with current value/preview |
| `guxgen/admin/{model}_detail_gen.go` | Renders file preview (image) or download link |
| `guxgen/api/{model}_gen.go` | No change (API client sends string path) |
| App setup (`app.go`) | Needs `app.SetStorage(...)` and upload route registration |

### Template Changes in `modelgen.go`

The `AdminNew` and `AdminEdit` templates currently emit field code based on `ModelField.Input`:

```go
// Existing pattern in templates:
{{if eq .Input "email"}}
    ui.Input(ui.InputProps{Type: ui.InputEmail, ...})
{{else if eq .Input "textarea"}}
    ui.Textarea(ui.TextareaProps{...})
{{else if eq .Input "select"}}
    ui.Select(ui.SelectProps{...})
```

Adding file:

```go
{{else if eq .Input "file"}}
    ui.FileUpload(ui.FileUploadProps{
        Name:      "{{.JSONName}}",
        Accept:    "{{.FileAccept}}",
        Value:     {{.FieldValueExpr}},
        UploadURL: "/__gux_api/upload/{{$.NameLower}}/{{.JSONName}}",
        OnUpload:  func(info ui.FileInfo) { {{.StateSetterExpr}}(info.Path) },
    })
```

### New Config Fields on ModelField

```go
// In model.go - extend ModelField struct
type FileConfig struct {
    Accept    string `json:"accept,omitempty"`    // MIME types
    MaxSize   string `json:"maxSize,omitempty"`   // "5MB", "10MB"
    Directory string `json:"directory,omitempty"` // Upload subdirectory
}

type ModelField struct {
    // ... existing fields
    FileConfig *FileConfig `json:"fileConfig,omitempty"` // File upload config
}
```

---

## Upload Endpoint Implementation

### Registration (in core or via code generation)

Two approaches, recommend approach A:

**Approach A: Explicit registration in generated app setup**

```go
// Generated in guxgen/ or user writes in app.go:
storage.RegisterUploadEndpoints(app, storage.EndpointConfig{
    Models: map[string][]storage.FieldConfig{
        "user":   {{Field: "avatar", Accept: "image/*", MaxSize: 5<<20, Dir: "avatars"}},
        "post":   {{Field: "cover", Accept: "image/*", MaxSize: 10<<20, Dir: "covers"}},
    },
})
```

**Approach B: Auto-register from CRUD model metadata**

Add file field metadata to `CRUDModel` struct and auto-register upload endpoints in `registerCRUDHandlers`. Tighter coupling but less boilerplate.

**Recommendation: Approach A.** It keeps `core/crud.go` clean and makes file upload opt-in at the app level. The code generator can emit the registration call based on `gux.config.json`.

### Handler Implementation

```go
func handleUpload(storage Storage, w http.ResponseWriter, r *http.Request) {
    // 1. Parse multipart form (r.ParseMultipartForm)
    // 2. Get file from form ("file" field)
    // 3. Validate: size, MIME type
    // 4. Generate unique path: "{dir}/{model}_{field}_{uuid}{ext}"
    // 5. Call storage.Upload()
    // 6. Return JSON: { "path": "...", "url": "...", "filename": "...", "size": 123 }
}
```

### File Serving

For local storage, files need to be served via HTTP. Two options:

1. **Static file server**: `http.FileServer` on the uploads directory
2. **Auth-aware handler**: Custom handler that checks authentication before serving

**Recommendation:** Auth-aware handler registered at `/__gux_api/files/` that checks the same auth config as CRUD endpoints. For S3, this is unnecessary (use presigned URLs).

### File Lifecycle: Deletion

When a record is deleted or a file field is updated, the old file should be cleaned up. This integrates with existing hooks:

```go
// Option 1: CRUD delete hook (new)
app.CRUD(models.User{}, core.WithDeleteHook(func(existing interface{}) error {
    // Clean up file
    storage.Delete(ctx, existing.Avatar)
    return nil
}))

// Option 2: Auto-cleanup via WithFileField option
app.CRUD(models.User{}, core.WithFileField("Avatar", fileConfig))
// Auto-registers: on delete, delete file; on update, if field changed, delete old file
```

**Recommendation:** `WithFileField` because it's declarative and the code generator can emit it. But `WithDeleteHook` should also exist as the general-purpose escape hatch.

---

## SSR vs WASM Considerations

### SSR (Server-Side Rendering)

- `ui.FileUpload` renders as a styled `<input type="file">` with optional preview of existing file
- No upload functionality in SSR (file uploads require JavaScript)
- Edit forms show current file with preview/download link
- The WASM bundle handles actual upload interaction after hydration

### WASM (Client-Side)

- Full upload UI: drag-and-drop zone, progress indicator, preview
- Uses `fetch.Upload()` with FormData API
- Progress tracking via `XMLHttpRequest` or fetch with `ReadableStream`
- CSRF token included automatically (same pattern as existing `fetch.Fetch`)

### Hydration

File upload state follows the same hydration pattern as other components:

```go
avatarPath := r.StateString("avatar", existingUser.Avatar)
// On SSR: renders with current value
// On WASM hydration: picks up value from state, enables upload
```

---

## Component Boundaries

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| `storage.Storage` interface | Store/retrieve/delete files | Used by upload handlers |
| `storage.LocalStorage` | Disk-based file storage | Filesystem |
| `storage.S3Storage` | S3-compatible cloud storage | AWS SDK |
| Upload HTTP handler | Parse multipart, validate, delegate to Storage | Storage, core.App (auth) |
| File serving handler | Serve stored files with auth | Storage, core.App (auth) |
| `ui.FileUpload` component | UI for selecting/uploading files | `fetch.Upload()`, state |
| `fetch.Upload()` | WASM multipart HTTP client | Browser fetch API |
| Code generator changes | Emit file-aware admin forms | gux.config.json |
| `ModelField.FileConfig` | Config schema for file fields | gux.config.json |

---

## Anti-Patterns to Avoid

### Anti-Pattern 1: Storing Files in the Database
**What:** Using BLOB columns for file content.
**Why bad:** Database bloat, slow queries, backup complexity, no CDN integration.
**Instead:** Store file path as string in DB, actual file in storage backend.

### Anti-Pattern 2: Synchronous Upload in CRUD Create
**What:** Making file upload part of the CRUD POST JSON body (base64 encoded).
**Why bad:** Large request bodies, no progress indication, timeout issues, 413 errors.
**Instead:** Upload file first (separate endpoint), get path, include path in CRUD JSON.

### Anti-Pattern 3: No File Validation Server-Side
**What:** Trusting client-side MIME type and size validation only.
**Why bad:** Client validation is bypassable. Malicious files can be uploaded.
**Instead:** Always validate on server: check magic bytes for MIME type, enforce size limits, sanitize filenames.

### Anti-Pattern 4: Predictable File Paths
**What:** Storing files at `uploads/{model}/{id}/avatar.jpg` with predictable names.
**Why bad:** Enumerable paths enable scraping of private files.
**Instead:** Include UUID or hash in filename: `uploads/avatars/{uuid}_{sanitized_name}.jpg`.

### Anti-Pattern 5: Coupling Storage Implementation to Core
**What:** Putting S3 client directly in `core/` package.
**Why bad:** Adds heavy AWS SDK dependency to everyone, even those using local storage.
**Instead:** Interface in `storage/`, implementations as separate files with build tags or lazy loading.

---

## Suggested Build Order

Based on dependency analysis, build in this order:

### Phase 1: Foundation (storage interface + local backend)
1. `storage/storage.go` - Interface, types, Config
2. `storage/local.go` - Local filesystem implementation
3. `core/app.go` - Add `SetStorage`/`GetStorage` methods
4. Upload HTTP handler (in `storage/` or `core/`)
5. File serving handler

**Rationale:** Everything else depends on having a working storage backend and upload endpoint.

### Phase 2: WASM Upload Client
1. `fetch/upload.go` - FormData-based upload from WASM
2. `fetch/upload_stub.go` - Non-WASM stub

**Rationale:** The UI component needs this to function in WASM.

### Phase 3: UI Component
1. `ui/file_upload.go` - Shared types and SSR rendering
2. `ui/file_upload_wasm.go` - WASM upload logic, progress, preview
3. `ui/file_upload_stub.go` - Stub for non-WASM builds

**Rationale:** Depends on fetch.Upload for WASM functionality.

### Phase 4: Code Generation
1. Extend `ModelField` with `FileConfig`
2. Update `modelgen.go` templates: AdminNew, AdminEdit, AdminDetail
3. Update `build.go` if needed for upload endpoint registration
4. Test with example app

**Rationale:** Depends on ui.FileUpload component existing. This is the integration layer.

### Phase 5: S3 Backend + Advanced Features
1. `storage/s3.go` - S3 implementation
2. Presigned URL support
3. Image processing (thumbnails, resize)
4. File lifecycle hooks (delete on record delete)
5. `WithFileField` CRUD option for auto-cleanup

**Rationale:** Optional enhancement, not needed for core functionality.

---

## File Lifecycle Hook Integration

The existing hook system in `core/crud.go`:

```go
OnCreate CreateHook  // func(data map[string]interface{}) (interface{}, error)
OnUpdate UpdateHook  // func(existing interface{}, data map[string]interface{}) (interface{}, error)
```

And the admin page hooks in `admin/hooks_gen.go`:

```go
Set{Model}BeforeSave  // func(ctx HookContext, data map[string]any, isEdit bool) error
Set{Model}AfterSave   // func(ctx HookContext, id uint, isEdit bool)
```

File cleanup needs a **delete hook** (currently missing from CRUD):

```go
// New addition to CRUDModel:
OnDelete DeleteHook  // func(existing interface{}) error

// New CRUDOption:
func WithDeleteHook(hook DeleteHook) CRUDOption
```

And a file-specific convenience:

```go
// Registers both update and delete hooks for file cleanup
func WithFileField(fieldName string, config FileFieldConfig) CRUDOption
```

This `WithFileField` option:
- On **update**: if field value changed, delete old file from storage
- On **delete**: delete the file from storage
- On **create**: no-op (file already uploaded to storage)

---

## Security Considerations

### Upload Authentication
Upload endpoints follow the same auth pattern as CRUD: protected by default, `Public()` to opt out, `WithRoles()` for role restriction. Uses existing `checkCRUDAuth` or `APIRegistration.checkAuth` patterns.

### CSRF Protection
The existing CSRF middleware in `core/csrf.go` protects POST requests. Upload POSTs go through the same middleware. The `fetch.Upload()` function must include the CSRF token (same `getCSRFToken()` from `fetch/fetch.go`).

### File Validation
Server-side validation is mandatory:
- Check file size before fully reading (use `http.MaxBytesReader`)
- Verify MIME type via magic bytes (not just the `Content-Type` header)
- Sanitize filename (strip path traversal, limit characters)
- Store with generated name (UUID prefix), not user-provided name

### File Serving Authorization
For private files, the file serving handler should check authentication before serving. Configuration option: `storage.WithPublicServing()` for truly public files (product images) vs default authenticated serving.

---

## Sources

- Codebase analysis: `core/crud.go`, `core/app.go`, `core/endpoint.go`, `core/node.go`, `core/elements.go`
- Component patterns: `ui/input.go`, `ui/form.go`, `ui/video_player.go` (platform split pattern)
- Code generation: `cmd/gux/modelgen.go`, `cmd/gux/model.go`, `cmd/gux/build.go`
- Config schema: `examples/minimal/gux.config.json`
- WASM fetch: `fetch/fetch.go`
- All findings HIGH confidence (direct codebase analysis, no external sources needed for architecture patterns)
