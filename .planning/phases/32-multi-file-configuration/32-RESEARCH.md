# Phase 32: Multi-File & Configuration - Research

**Researched:** 2026-02-03
**Domain:** Multi-file upload fields, per-field upload configuration, code generation
**Confidence:** HIGH

## Summary

Phase 32 extends the existing single-file upload system (Phases 29-31) with two capabilities: (1) multi-file fields that store a JSON array of storage keys, and (2) per-field configuration for accept types, max size, and upload directory in `gux.config.json`.

The existing infrastructure is well-suited for extension. The upload endpoint (`core/upload.go`) already processes multiple files per request and returns an array of `UploadResult`. The `ui.FileUpload` component already accepts `Accept` and `MaxSize` props. The code generation pipeline (`cmd/gux/modelgen.go`) already has `IsFile` detection and file-specific rendering for forms, detail views, and list views. The primary work is: (a) a new `MultiFileUpload` UI component, (b) extending `ModelField` with file configuration fields, (c) teaching code generation to emit the new component and pass config-driven props, and (d) extending CRUD cleanup to handle JSON arrays of keys.

**Primary recommendation:** Use `"input": "file[]"` in gux.config.json for multi-file fields. Model field type remains `string` (stores JSON array). DTO type becomes `[]*core.FileInfo`. Per-field config uses `"accept"`, `"maxSize"`, and `"directory"` keys on the field definition in gux.config.json.

## Standard Stack

### Core (Already in Codebase)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `core/storage.go` | N/A | Storage interface + FileInfo | Already handles Put/Delete/URL/Serve |
| `core/storage_local.go` | N/A | Local filesystem storage | Content-addressed with SHA-256 |
| `core/upload.go` | N/A | HTTP upload handler | Already returns `[]*UploadResult` array |
| `ui/file_upload.go` | N/A | Single file upload component | Base for multi-file component |
| `cmd/gux/modelgen.go` | N/A | Code generation pipeline | Existing `IsFile` detection, `generateFormFieldCode`, `generateDetailFieldCode` |

### No New Dependencies Required

All functionality can be built with the existing Go standard library and the codebase's established patterns. JSON marshaling/unmarshaling for the array of keys uses `encoding/json`. No new external libraries needed.

## Architecture Patterns

### Pattern 1: Multi-File Field Storage as JSON Array

**What:** Multi-file fields store a JSON-encoded string array of storage keys in a single `string` database column.
**When to use:** Always for `"input": "file[]"` fields.

The model field remains `type string` (consistent with single file fields). The value is a JSON array like `["ab/abc123.jpg","cd/cde456.pdf"]`. This avoids needing a join table or separate model for file associations.

```go
// Model field (unchanged type)
type Product struct {
    gorm.Model
    Gallery string // stores: `["ab/abc123.jpg","cd/cde456.pdf"]`
}

// DTO field becomes a slice of FileInfo
type ProductDetail struct {
    ID      uint             `json:"id"`
    Gallery []*core.FileInfo `json:"gallery"` // Resolved from JSON array of keys
}
```

### Pattern 2: MultiFileUpload Component

**What:** A new `ui.MultiFileUpload` component that manages a list of uploaded files, allowing add/remove of individual items.
**When to use:** Generated for `"input": "file[]"` fields in admin forms.

The component maintains a visual list of uploaded files. Each file has a remove button. New files are uploaded individually (reusing `fetch.Upload`). State is tracked as a JSON string of keys via `r.StateString`.

```go
// MultiFileUpload renders a multi-file upload area with file list
ui.MultiFileUpload(ui.MultiFileUploadProps{
    Accept:           []string{"image/*"},
    MaxSize:          5 * 1024 * 1024,
    Values:           []string{"/__gux_api/files/ab/abc.jpg"}, // Current file URLs
    OnUploadComplete: func(result ui.UploadResult) { /* append key to state */ },
    OnRemove:         func(index int) { /* remove key at index from state */ },
})
```

### Pattern 3: Per-Field Config in gux.config.json

**What:** File configuration fields (`accept`, `maxSize`, `directory`) are added to the `ModelField` struct and read from `gux.config.json`.
**When to use:** Any file or file[] field that needs constraints.

```json
{
    "name": "Avatar",
    "type": "string",
    "input": "file",
    "accept": "image/*",
    "maxSize": "5MB",
    "directory": "avatars"
}
```

These values flow through the code generation pipeline into the `FileUploadProps` / `MultiFileUploadProps` as literal values in generated code.

### Pattern 4: Upload Directory Routing

**What:** The `"directory"` config controls a subdirectory prefix within the storage base directory, passed as a query parameter or form field to the upload endpoint.
**When to use:** When different file fields should be stored in different subdirectories.

**Implementation approach:** Pass directory as a query parameter to the upload endpoint: `POST /__gux_api/upload?dir=avatars`. The upload handler creates the subdirectory within the storage base. The storage key then includes the directory prefix: `avatars/ab/abc123.jpg`. File serving continues to work because keys are path-relative.

### Recommended Project Structure (Changes)

```
core/
    storage.go          # Add MultiFileInfo helper functions
    crud.go             # Extend populateFileInfoFields for []*FileInfo, extend cleanup for JSON arrays
    upload.go           # Add directory query parameter support
ui/
    file_upload.go          # Unchanged (single file)
    multi_file_upload.go    # NEW: Multi-file upload component (shared SSR)
    multi_file_upload_wasm.go  # NEW: WASM interactivity for multi-file
    multi_file_upload_stub.go  # NEW: Stub for non-WASM builds
cmd/gux/
    model.go            # Add Accept, MaxSize, Directory to ModelField
    modelgen.go         # Extend generateFormFieldCode, generateDetailFieldCode for file[]
```

### Anti-Patterns to Avoid

- **Join table for multi-file:** Do not create a separate model/table for file associations. A JSON array in a string column is the established pattern (consistent with `[]string` handling already in the codebase).
- **Multiple upload endpoints:** Do not create per-model or per-field upload endpoints. The single `/__gux_api/upload` endpoint with optional `?dir=` parameter is sufficient.
- **Client-side size parsing:** Do not parse `"5MB"` strings in WASM. Parse at code generation time and emit the byte value as a literal in generated code.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JSON array serialization | Custom string splitting/joining | `encoding/json` Marshal/Unmarshal of `[]string` | Edge cases with special characters in filenames |
| Size string parsing | Inline parsing in generated code | Parse at generation time, emit int64 literal | `"5MB"` -> `5242880` at gen time, simpler generated code |
| File type validation | New validation system | Existing `matchMimePattern` in both `core/storage.go` and `ui/file_upload_wasm.go` | Already battle-tested, supports wildcards |
| Upload progress | New upload mechanism | Existing `fetch.Upload` with XHR progress | Already implemented in Phase 30 |
| File cleanup on delete | New cleanup logic | Extend existing `getFileFieldValues` to parse JSON arrays | Just needs to handle both string and JSON array formats |

## Common Pitfalls

### Pitfall 1: State Management for Multi-File Fields
**What goes wrong:** Using a Go slice for state instead of a JSON string, which doesn't work with `r.StateString`.
**Why it happens:** Tempting to use `core.UseState[[]string]` but the state system works best with primitives.
**How to avoid:** Store the JSON array as a string via `r.StateString`. Parse/serialize on every read/write. Helper functions `parseFileKeys(jsonStr) []string` and `serializeFileKeys(keys []string) string` should be generated or provided.
**Warning signs:** State not persisting across re-renders, hydration failures.

### Pitfall 2: Cleanup Logic Must Handle Both Formats
**What goes wrong:** `getFileFieldValues` returns a single string, but multi-file fields contain a JSON array. Cleanup deletes the JSON string as if it were a storage key.
**Why it happens:** Existing code assumes `fileFields` values are always single storage keys (strings).
**How to avoid:** Introduce a field type indicator (single vs multi) in `CRUDModel`, or have `getFileFieldValues` detect JSON arrays (starts with `[`) and return all keys. Better: add `MultiFileFields []string` alongside `FileFields []string` in `CRUDModel`.
**Warning signs:** Files not cleaned up on delete, orphaned files in storage.

### Pitfall 3: Directory Config Must Not Break Content-Addressed Storage
**What goes wrong:** Adding a directory prefix interferes with the SHA-256 content-addressing scheme, causing duplicate files or serving failures.
**Why it happens:** Storage keys are `prefix/hash+ext` (e.g., `ab/abc123.jpg`). Adding a directory would make it `avatars/ab/abc123.jpg`.
**How to avoid:** The directory prefix is prepended to the existing key structure. `LocalStorage.Put` needs to accept an optional directory parameter or a new method. The key returned includes the directory prefix, so URL/Serve/Delete all work transparently.
**Warning signs:** 404s when serving files, wrong paths in database.

### Pitfall 4: Edit Form Initialization for Multi-File
**What goes wrong:** Edit forms fail to display existing files because the state initialization doesn't parse the JSON array.
**Why it happens:** Single file fields initialize state from `FileInfo.Key` (via nil-safe pattern). Multi-file needs to parse a JSON array of keys into URLs for display.
**How to avoid:** The edit form state init for multi-file fields should store the raw JSON string. The `MultiFileUpload` component's `Values` prop should be computed by parsing the JSON array and converting each key to a URL.
**Warning signs:** Edit form shows empty file list even when files exist.

### Pitfall 5: Size String Parsing Inconsistency
**What goes wrong:** `"5MB"` vs `"5mb"` vs `"5 MB"` parsed differently.
**Why it happens:** No canonical format specified.
**How to avoid:** Define canonical format in docs: `"<number><unit>"` where unit is `KB`, `MB`, `GB` (case-insensitive, no space). Parse at code generation time using a helper function in `cmd/gux/model.go`. Emit the int64 byte value in generated code.
**Warning signs:** Validation passing/failing unexpectedly.

### Pitfall 6: populateFileInfoFields Reflection for Slice Type
**What goes wrong:** The existing `populateFileInfoFields` checks for `*FileInfo` type on DTO fields. Multi-file fields have `[]*FileInfo` type, which won't match.
**Why it happens:** The reflection code uses `reflect.TypeOf((*FileInfo)(nil))` for type comparison.
**How to avoid:** Add a second check for `reflect.SliceOf(reflect.TypeOf((*FileInfo)(nil)))` or better, introduce a separate loop for multi-file fields using `MultiFileFields` list.
**Warning signs:** Multi-file DTOs returning nil/empty arrays even when files exist.

## Code Examples

### gux.config.json Multi-File Field Definition
```json
{
    "name": "Gallery",
    "type": "string",
    "input": "file[]",
    "accept": "image/*",
    "maxSize": "10MB",
    "directory": "gallery"
}
```

### ModelField Extensions (cmd/gux/model.go)
```go
type ModelField struct {
    // ... existing fields ...
    Accept    string `json:"accept,omitempty"`    // Allowed MIME patterns (e.g., "image/*")
    MaxSize   string `json:"maxSize,omitempty"`   // Max file size (e.g., "5MB")
    Directory string `json:"directory,omitempty"` // Upload subdirectory (e.g., "avatars")
}
```

### Size String Parser (cmd/gux/model.go or modelgen.go)
```go
func parseSizeString(s string) (int64, error) {
    s = strings.TrimSpace(strings.ToUpper(s))
    multipliers := map[string]int64{
        "B": 1, "KB": 1024, "MB": 1024 * 1024, "GB": 1024 * 1024 * 1024,
    }
    for suffix, mult := range multipliers {
        if strings.HasSuffix(s, suffix) {
            numStr := strings.TrimSpace(strings.TrimSuffix(s, suffix))
            num, err := strconv.ParseFloat(numStr, 64)
            if err != nil { return 0, err }
            return int64(num * float64(mult)), nil
        }
    }
    // Try plain number (bytes)
    return strconv.ParseInt(s, 10, 64)
}
```

### Generated Form Code for file[] Field
```go
// State: JSON string of storage keys
galleryState := r.StateString("gallery", "")

// In form:
ui.MultiFileUpload(ui.MultiFileUploadProps{
    Accept:  []string{"image/*"},
    MaxSize: 10485760, // 10MB parsed at gen time
    Values: func() []string {
        var keys []string
        if galleryState.Get() != "" {
            json.Unmarshal([]byte(galleryState.Get()), &keys)
        }
        urls := make([]string, len(keys))
        for i, k := range keys {
            urls[i] = "/__gux_api/files/" + k
        }
        return urls
    }(),
    OnUploadComplete: func(result ui.UploadResult) {
        var keys []string
        if galleryState.Get() != "" {
            json.Unmarshal([]byte(galleryState.Get()), &keys)
        }
        keys = append(keys, result.Key)
        data, _ := json.Marshal(keys)
        galleryState.Set(string(data))
    },
    OnRemove: func(index int) {
        var keys []string
        if galleryState.Get() != "" {
            json.Unmarshal([]byte(galleryState.Get()), &keys)
        }
        keys = append(keys[:index], keys[index+1:]...)
        data, _ := json.Marshal(keys)
        galleryState.Set(string(data))
    },
    UploadURL: "/__gux_api/upload?dir=gallery",
})
```

### MultiFileUploadProps
```go
type MultiFileUploadProps struct {
    Accept           []string
    MaxSize          int64
    UploadURL        string
    Placeholder      string
    Class            string
    Disabled         bool
    Values           []string // Current file URLs for display
    OnUploadComplete func(result UploadResult)
    OnRemove         func(index int) // Index of file to remove
    OnError          func(err string)
}
```

### CRUD Multi-File Cleanup Extension
```go
// In CRUDModel:
type CRUDModel struct {
    // ... existing ...
    MultiFileFields []string // Fields that store JSON arrays of storage keys
}

// WithMultiFileFields registers multi-file field names
func WithMultiFileFields(fields ...string) CRUDOption {
    return func(m *CRUDModel) {
        m.MultiFileFields = fields
    }
}

// Extended cleanup: parse JSON array and delete each key
func getMultiFileFieldValues(item interface{}, multiFileFields []string) map[string][]string {
    values := make(map[string][]string)
    v := reflect.ValueOf(item)
    if v.Kind() == reflect.Ptr { v = v.Elem() }
    for _, fieldName := range multiFileFields {
        f := v.FieldByName(fieldName)
        if f.IsValid() && f.Kind() == reflect.String {
            var keys []string
            json.Unmarshal([]byte(f.String()), &keys)
            values[fieldName] = keys
        }
    }
    return values
}
```

### Upload Endpoint Directory Support
```go
// In handleUpload, read directory from query parameter:
dir := r.URL.Query().Get("dir")
if dir != "" {
    // Validate: no path traversal
    if strings.Contains(dir, "..") || strings.Contains(dir, "/") {
        // reject
    }
}
// Pass to storage.Put - the key returned will include the directory prefix
```

### Detail View for Multi-File (Generated Code Pattern)
```go
// Generated detail field for Gallery []*core.FileInfo
func() core.Node {
    files := displayItem.Gallery
    if len(files) == 0 {
        return productDetailRow("Gallery", "-")
    }
    var nodes []core.Node
    for _, fi := range files {
        if isImageFile(fi.Filename) {
            nodes = append(nodes, core.A(core.Attrs{Href: fi.URL, Extra: map[string]string{"target": "_blank"}},
                core.Img(core.Attrs{Src: fi.URL, Class: "h-20 w-20 rounded object-cover", Alt: fi.Filename}),
            ))
        } else {
            nodes = append(nodes, core.A(core.Attrs{Href: fi.URL, Class: "text-blue-600 hover:underline"},
                core.Text(fi.Filename),
            ))
        }
    }
    return core.Div(core.Class("py-3 px-6"),
        core.Div(core.Class("text-sm font-medium text-gray-500 dark:text-gray-400 mb-2"), core.Text("Gallery")),
        core.Div(core.Class("flex flex-wrap gap-2"), nodes...),
    )
}(),
```

### List View for Multi-File (Generated Code Pattern)
```go
// Show count + first thumbnail
fi := item.Gallery
if len(fi) == 0 {
    return tableCell("0 files")
}
// Show first image thumbnail + count
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Single file only | Extending to multi-file | Phase 32 | Enables galleries, document collections |
| Global storage config | Per-field config | Phase 32 | Different constraints per field |
| Fixed upload directory | Per-field directories | Phase 32 | Better file organization |

## Open Questions

1. **Upload endpoint directory parameter: query param vs form field?**
   - What we know: Query parameter (`?dir=avatars`) is simpler and doesn't conflict with multipart form parsing.
   - What's unclear: Should `LocalStorage.Put` accept the directory, or should a wrapper handle it?
   - Recommendation: Pass directory as query param, handle in `handleUpload` by prepending to the key path before calling `storage.Put`. Alternatively, create a `PutWithDir` method or pass directory through context.

2. **Should `"directory"` affect server-side validation or just storage location?**
   - What we know: The requirement says "upload directory" suggesting storage location only.
   - What's unclear: Whether per-field accept/maxSize should be validated server-side (in addition to client-side).
   - Recommendation: Server-side validation via per-field config is out of scope for Phase 32. The global storage config (`WithMaxSize`, `WithAllowedTypes`) provides server-side validation. Per-field config drives client-side validation in the UI component. This is consistent with Phase 31's approach where `FileUploadProps.Accept` and `FileUploadProps.MaxSize` are client-side only.

3. **TemplateField changes: new `IsMultiFile` flag or extend `IsFile`?**
   - What we know: `IsFile` is currently `field.Input == "file"`.
   - Recommendation: Add `IsMultiFile bool` (`field.Input == "file[]"`), keep `IsFile` for single files. Code generation switches on both.

## Sources

### Primary (HIGH confidence)
- Codebase analysis: `core/storage.go`, `core/storage_local.go`, `core/upload.go`, `core/crud.go`
- Codebase analysis: `ui/file_upload.go`, `ui/file_upload_wasm.go`
- Codebase analysis: `cmd/gux/model.go` (ModelField struct), `cmd/gux/modelgen.go` (TemplateField, generateFormFieldCode, generateDetailFieldCode)
- Prior phase decisions from `.planning/phases/31-code-generation-crud-integration/`

### Secondary (MEDIUM confidence)
- Requirements from `.planning/REQUIREMENTS.md` (MULT-01 through MULT-04, CONF-01 through CONF-03)
- Architecture decisions from `.planning/research/ARCHITECTURE.md`

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All code paths analyzed directly from codebase
- Architecture: HIGH - Patterns follow established conventions from Phases 29-31
- Pitfalls: HIGH - Based on direct analysis of existing code (reflection, state management, cleanup logic)
- Code examples: MEDIUM - Generated code patterns extrapolated from existing `generateFormFieldCode` patterns

**Research date:** 2026-02-03
**Valid until:** 2026-03-03 (stable internal codebase, no external dependency changes)
