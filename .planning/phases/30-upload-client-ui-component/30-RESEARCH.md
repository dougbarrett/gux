# Phase 30: Upload Client & UI Component - Research

**Researched:** 2026-02-01
**Domain:** Go WASM file upload with XMLHttpRequest progress, drag-and-drop UI, client-side validation, image preview
**Confidence:** MEDIUM

## Summary

Phase 30 implements the client-side file upload experience: a WASM upload client that tracks progress using XMLHttpRequest, and a reusable ui.FileUpload component that handles file selection, drag-and-drop, client-side validation, image preview, and SSR/WASM split rendering. Research focused on Go/WASM integration with JavaScript FormData and XMLHttpRequest upload events, file validation patterns, and accessibility best practices.

The key technical challenge is that Go/WASM must use JavaScript interop (syscall/js) to create FormData objects and XMLHttpRequest instances, as the Fetch API doesn't support upload progress events. Files must stay in JavaScript memory during upload (not copied into WASM heap) for performance. The UI component follows gux's established SSR/WASM split pattern (static HTML during SSR, hydrated interactivity in WASM) seen in VideoPlayer.

**Primary recommendation:** Create fetch/upload.go using XMLHttpRequest via syscall/js for progress tracking, keeping File objects in JavaScript land. Build ui.FileUpload following the VideoPlayer pattern (single component function with build tag stub functions for WASM-only initialization). Use URL.createObjectURL for efficient image previews, and implement comprehensive keyboard navigation and ARIA attributes for accessibility.

## Standard Stack

Go WASM file upload relies entirely on browser APIs accessed via syscall/js. No external libraries required.

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| syscall/js | stdlib | JavaScript interop in WASM | Official Go WASM bridge for browser APIs |
| net/url | stdlib | URL encoding for form fields | Standard query parameter handling |
| fetch (existing) | gux | HTTP client with CSRF | Already implements CSRF token handling |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| N/A | - | No third-party libraries needed | JavaScript APIs provide all functionality |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| XMLHttpRequest | Fetch API | Fetch API doesn't support upload progress events (deal-breaker) |
| syscall/js | github.com/sternix/wasm | Wrapper adds complexity, syscall/js is official and sufficient |
| URL.createObjectURL | FileReader.readAsDataURL | FileReader loads entire file into memory, createObjectURL is synchronous and efficient |

**Installation:**
No external dependencies - uses Go stdlib and existing gux packages.

## Architecture Patterns

### Recommended Project Structure
```
fetch/
├── fetch.go              # Existing fetch client (GET/POST/PUT/DELETE with CSRF)
└── upload.go             # New: XMLHttpRequest upload with progress tracking

ui/
├── file_upload.go        # Main component (SSR-safe)
├── file_upload_stub.go   # No-op stubs for non-WASM builds
└── file_upload_wasm.go   # WASM-only interactivity (drag-drop, preview, validation)
```

### Pattern 1: XMLHttpRequest via syscall/js for Upload Progress
**What:** JavaScript's XMLHttpRequest.upload provides progress events that Fetch API lacks
**When to use:** All file uploads requiring progress feedback
**Example:**
```go
// Source: MDN XMLHttpRequest.upload + Go syscall/js patterns
//go:build js && wasm
package fetch

import (
    "errors"
    "syscall/js"
)

// UploadOptions configures file upload with progress tracking
type UploadOptions struct {
    URL      string
    File     js.Value              // JavaScript File object (stays in JS memory)
    OnProgress func(loaded, total int64)  // Progress callback
}

// Upload performs multipart file upload with progress tracking
func Upload(opts UploadOptions) (*Response, error) {
    // Create FormData and append file (file stays in JavaScript heap)
    formData := js.Global().Get("FormData").New()
    formData.Call("append", "file", opts.File)

    // Create XMLHttpRequest
    xhr := js.Global().Get("XMLHttpRequest").New()

    // Register upload progress listener BEFORE open()
    if opts.OnProgress != nil {
        upload := xhr.Get("upload")
        upload.Call("addEventListener", "progress", js.FuncOf(func(this js.Value, args []js.Value) any {
            e := args[0]
            if e.Get("lengthComputable").Bool() {
                loaded := int64(e.Get("loaded").Int())
                total := int64(e.Get("total").Int())
                opts.OnProgress(loaded, total)
            }
            return nil
        }))
    }

    // Open and configure
    xhr.Call("open", "POST", opts.URL)

    // Add CSRF token (reuse fetch package logic)
    csrfToken := getCSRFToken()
    if csrfToken != "" {
        xhr.Call("setRequestHeader", "X-CSRF-Token", csrfToken)
    }

    // Send (FormData automatically sets multipart Content-Type)
    xhr.Call("send", formData)

    // Wait for completion via channel
    // ... error handling and response parsing
}
```

### Pattern 2: File Input with Drag-and-Drop (SSR/WASM Split)
**What:** Component renders static HTML in SSR, hydrates with drag-drop handlers in WASM
**When to use:** File upload UI components
**Example:**
```go
// Source: ui/video_player.go pattern, MDN File Drag and Drop
// file_upload.go (works in both SSR and WASM)
func FileUpload(props FileUploadProps) core.Node {
    uploadID := fmt.Sprintf("gux-upload-%d", idCounter.Add(1))

    attrs := core.Attrs{
        ID:    uploadID,
        Class: buildUploadClasses(props),
    }

    // Set OnMount for WASM initialization (no-op in SSR)
    attrs.OnMount = createFileUploadHandlers(uploadID, props)

    return core.Div(attrs,
        core.Input(core.Attrs{
            Type:  "file",
            ID:    uploadID + "-input",
            Class: "hidden",
        }),
        core.Label(core.Attrs{For: uploadID + "-input"},
            core.Text("Click to upload or drag and drop"),
        ),
    )
}

// file_upload_stub.go (!js || !wasm)
//go:build !js || !wasm
func createFileUploadHandlers(id string, props FileUploadProps) func(el any) {
    return nil  // SSR no-op
}

// file_upload_wasm.go (js && wasm)
//go:build js && wasm
func createFileUploadHandlers(id string, props FileUploadProps) func(el any) {
    return func(el any) {
        jsEl := el.(js.Value)

        // Prevent default browser file handling on window
        js.Global().Get("window").Call("addEventListener", "drop", js.FuncOf(func(this js.Value, args []js.Value) any {
            args[0].Call("preventDefault")
            return nil
        }))

        // Drop zone handlers
        jsEl.Call("addEventListener", "dragover", js.FuncOf(func(this js.Value, args []js.Value) any {
            e := args[0]
            e.Call("preventDefault")
            e.Get("dataTransfer").Set("dropEffect", "copy")
            return nil
        }))

        jsEl.Call("addEventListener", "drop", js.FuncOf(func(this js.Value, args []js.Value) any {
            e := args[0]
            e.Call("preventDefault")
            files := e.Get("dataTransfer").Get("files")
            handleFiles(files, props)
            return nil
        }))
    }
}
```

### Pattern 3: Client-Side File Validation (Type and Size)
**What:** Validate file type and size before upload to provide instant feedback
**When to use:** All file upload components (complements server-side validation)
**Example:**
```go
// Source: MDN File API, web validation best practices
//go:build js && wasm

func validateFile(file js.Value, props FileUploadProps) error {
    // Check size
    size := file.Get("size").Int()
    if props.MaxSize > 0 && int64(size) > props.MaxSize {
        return fmt.Errorf("File too large: %d bytes (max %d bytes)", size, props.MaxSize)
    }

    // Check type (client-side only - server validates magic bytes)
    fileType := file.Get("type").String()
    if len(props.AllowedTypes) > 0 {
        allowed := false
        for _, pattern := range props.AllowedTypes {
            if matchMimePattern(fileType, pattern) {
                allowed = true
                break
            }
        }
        if !allowed {
            return fmt.Errorf("File type %s not allowed", fileType)
        }
    }

    return nil
}

func matchMimePattern(mimeType, pattern string) bool {
    // "image/*" matches "image/jpeg", etc.
    if strings.HasSuffix(pattern, "/*") {
        prefix := strings.TrimSuffix(pattern, "/*")
        return strings.HasPrefix(mimeType, prefix+"/")
    }
    return mimeType == pattern
}
```

### Pattern 4: Image Preview with URL.createObjectURL
**What:** Generate blob URL for instant preview without loading file into memory
**When to use:** Image uploads requiring preview (more efficient than FileReader)
**Example:**
```go
// Source: MDN Using files from web applications
//go:build js && wasm

func showImagePreview(file js.Value, containerID string) func() {
    // Create object URL (synchronous, efficient)
    objURL := js.Global().Get("URL").Call("createObjectURL", file).String()

    // Create img element
    doc := js.Global().Get("document")
    img := doc.Call("createElement", "img")
    img.Set("src", objURL)
    img.Get("classList").Call("add", "upload-preview")

    // Append to container
    container := doc.Call("getElementById", containerID)
    container.Call("appendChild", img)

    // Return cleanup function
    return func() {
        js.Global().Get("URL").Call("revokeObjectURL", objURL)
        container.Call("removeChild", img)
    }
}
```

### Pattern 5: State Management for Upload Progress
**What:** Use Router.StateInt for upload progress percentage tracking
**When to use:** Components displaying upload progress bars
**Example:**
```go
// Source: gux core/app.go State patterns
func UploadPage(r *core.Router) func() core.Node {
    uploadProgress := r.StateInt("uploadProgress", 0)
    uploading := r.StateBool("uploading", false)

    handleUpload := func(file js.Value) {
        uploading.Set(true)
        uploadProgress.Set(0)

        fetch.Upload(fetch.UploadOptions{
            URL:  "/__gux_api/upload",
            File: file,
            OnProgress: func(loaded, total int64) {
                pct := int((loaded * 100) / total)
                uploadProgress.Set(pct)
            },
        })

        uploading.Set(false)
    }

    return func() core.Node {
        if uploading.Get() {
            return core.Div(core.Class("progress-bar"),
                core.Text(fmt.Sprintf("Uploading: %d%%", uploadProgress.Get())),
            )
        }
        return ui.FileUpload(ui.FileUploadProps{
            OnFileSelected: handleUpload,
        })
    }
}
```

### Anti-Patterns to Avoid
- **Copying file bytes into Go WASM memory:** Use JavaScript File objects directly - copying large files exhausts WASM memory
- **Using Fetch API for upload progress:** Fetch doesn't support upload.progress events - must use XMLHttpRequest
- **Attaching event listeners after send():** Upload progress listeners must be registered BEFORE xhr.open() due to browser bugs
- **Not canceling dragover event:** Browser will open/download files if dragover isn't prevented
- **FileReader for image preview:** Less efficient than URL.createObjectURL, loads entire file into memory
- **Missing URL.revokeObjectURL:** Creates memory leaks - always clean up blob URLs

## Don't Hand-Roll

Browser APIs and established patterns handle complex file upload scenarios.

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Upload progress tracking | Custom fetch wrapper | XMLHttpRequest.upload.progress | Only API supporting upload progress events |
| File type validation | String extension checks | File.type property + pattern matching | Browser provides accurate MIME type from file metadata |
| Image preview | Custom image loading | URL.createObjectURL | Synchronous, efficient, no memory copy |
| Drag-and-drop | Custom event handling | HTML5 Drag and Drop API | Handles DataTransfer, file extraction, cursor changes |
| FormData construction | Manual multipart encoding | FormData.append | Browser handles boundary generation, encoding, headers |
| CSRF token management | Manual token extraction | Reuse fetch.getCSRFToken() | Already implemented with meta tag + cookie fallback |
| Keyboard navigation | Custom focus management | Native tabindex + focusable elements | Screen readers and keyboard users expect standard behavior |

**Key insight:** Browser File API keeps file bytes in JavaScript heap until upload. Go WASM can reference File objects via js.Value without copying data, enabling efficient multi-GB file uploads that would exceed WASM memory limits if copied.

## Common Pitfalls

### Pitfall 1: Upload Progress Listeners Registered Too Late
**What goes wrong:** Upload progress events never fire, progress callback never called
**Why it happens:** Browsers have bugs - listeners registered after open() may not receive events
**How to avoid:** Always attach upload.progress listeners BEFORE calling xhr.open()
**Warning signs:** Progress always shows 0% or jumps immediately to 100%

### Pitfall 2: Fetch API Used Instead of XMLHttpRequest
**What goes wrong:** No progress feedback during upload, users don't know if upload is working
**Why it happens:** Fetch API is modern but lacks upload.progress events
**How to avoid:** Use XMLHttpRequest for file uploads, reserve Fetch API for regular requests
**Warning signs:** Upload progress feature request appears impossible to implement

### Pitfall 3: File Bytes Copied to WASM Memory
**What goes wrong:** WASM memory exhaustion when uploading files > 100MB
**Why it happens:** Using js.CopyBytesToGo to read file bytes into Go []byte
**How to avoid:** Pass File objects as js.Value directly to FormData, keep bytes in JavaScript
**Warning signs:** Out of memory errors or extreme slowness with large files

### Pitfall 4: Missing dragover preventDefault
**What goes wrong:** Browser navigates to file or downloads it instead of uploading
**Why it happens:** Default dragover behavior opens files
**How to avoid:** Always call preventDefault() in both dragover and drop handlers
**Warning signs:** File opens in new tab instead of triggering upload

### Pitfall 5: CORS Preflight Triggered by Upload Listeners
**What goes wrong:** Simple POST becomes preflight OPTIONS request, adding latency
**Why it happens:** Attaching XMLHttpRequestUpload event listeners makes request "non-simple"
**How to avoid:** Accept preflight cost (negligible compared to upload time), or use Fetch for small uploads without progress
**Warning signs:** Unnecessary OPTIONS requests visible in network inspector

### Pitfall 6: No Keyboard Access to File Input
**What goes wrong:** Keyboard-only users cannot trigger file selection
**Why it happens:** Custom styled div without proper label/input association
**How to avoid:** Use <label for="inputID"> wrapping visible UI, with hidden <input type="file" id="inputID">
**Warning signs:** Failed accessibility audit, keyboard navigation broken

### Pitfall 7: Missing ARIA Live Regions for Upload Status
**What goes wrong:** Screen reader users unaware of upload progress or completion
**Why it happens:** Visual progress indicators without aria-live announcements
**How to avoid:** Wrap status messages in <div aria-live="polite"> or aria-live="assertive"
**Warning signs:** Screen reader users report no upload feedback

### Pitfall 8: Network Interruption Shows False 100% Progress
**What goes wrong:** Progress bar jumps to 100% when network disconnects
**Why it happens:** XHR spec behavior - loadend event fires regardless of outcome
**How to avoid:** Check xhr.status in load/error handlers, distinguish success from failure
**Warning signs:** Users report "upload succeeded" but file missing on server

## Code Examples

Verified patterns from official sources.

### XMLHttpRequest Upload with Progress (WASM)
```go
// Source: MDN XMLHttpRequest.upload property
//go:build js && wasm
package fetch

import (
    "errors"
    "syscall/js"
)

type UploadOptions struct {
    URL        string
    File       js.Value  // JavaScript File object
    OnProgress func(loaded, total int64)
}

func Upload(opts UploadOptions) (*Response, error) {
    done := make(chan struct{})
    var response *Response
    var uploadErr error

    // Create FormData
    formData := js.Global().Get("FormData").New()
    formData.Call("append", "file", opts.File)

    // Create XHR
    xhr := js.Global().Get("XMLHttpRequest").New()

    // Upload progress (BEFORE open)
    if opts.OnProgress != nil {
        upload := xhr.Get("upload")
        progressCb := js.FuncOf(func(this js.Value, args []js.Value) any {
            e := args[0]
            if e.Get("lengthComputable").Bool() {
                loaded := int64(e.Get("loaded").Int())
                total := int64(e.Get("total").Int())
                opts.OnProgress(loaded, total)
            }
            return nil
        })
        upload.Call("addEventListener", "progress", progressCb)
    }

    // Load handler (success)
    loadCb := js.FuncOf(func(this js.Value, args []js.Value) any {
        response = &Response{
            Status:     xhr.Get("status").Int(),
            StatusText: xhr.Get("statusText").String(),
            OK:         xhr.Get("status").Int() >= 200 && xhr.Get("status").Int() < 300,
            Body:       xhr.Get("responseText").String(),
        }
        close(done)
        return nil
    })
    xhr.Call("addEventListener", "load", loadCb)

    // Error handler
    errorCb := js.FuncOf(func(this js.Value, args []js.Value) any {
        uploadErr = errors.New("upload failed")
        close(done)
        return nil
    })
    xhr.Call("addEventListener", "error", errorCb)

    // Open and send
    xhr.Call("open", "POST", opts.URL)

    // Add CSRF token
    csrfToken := getCSRFToken()
    if csrfToken != "" {
        xhr.Call("setRequestHeader", "X-CSRF-Token", csrfToken)
    }

    xhr.Call("send", formData)

    // Wait for completion
    <-done

    return response, uploadErr
}
```

### Drag-and-Drop File Input (WASM)
```go
// Source: MDN File drag and drop
//go:build js && wasm

func initDropZone(el js.Value, props FileUploadProps) {
    doc := js.Global().Get("document")

    // Prevent default on window
    js.Global().Get("window").Call("addEventListener", "drop", js.FuncOf(func(this js.Value, args []js.Value) any {
        args[0].Call("preventDefault")
        return nil
    }))

    // Dragover (required for drop to fire)
    el.Call("addEventListener", "dragover", js.FuncOf(func(this js.Value, args []js.Value) any {
        e := args[0]
        e.Call("preventDefault")
        e.Get("dataTransfer").Set("dropEffect", "copy")
        el.Get("classList").Call("add", "drag-over")
        return nil
    }))

    // Dragleave (visual feedback)
    el.Call("addEventListener", "dragleave", js.FuncOf(func(this js.Value, args []js.Value) any {
        el.Get("classList").Call("remove", "drag-over")
        return nil
    }))

    // Drop
    el.Call("addEventListener", "drop", js.FuncOf(func(this js.Value, args []js.Value) any {
        e := args[0]
        e.Call("preventDefault")
        el.Get("classList").Call("remove", "drag-over")

        files := e.Get("dataTransfer").Get("files")
        if files.Get("length").Int() > 0 {
            handleFiles(files, props)
        }
        return nil
    }))
}
```

### Client-Side File Validation
```go
// Source: MDN File API
//go:build js && wasm

func validateFile(file js.Value, maxSize int64, allowedTypes []string) error {
    // Size check
    size := int64(file.Get("size").Int())
    if maxSize > 0 && size > maxSize {
        return fmt.Errorf("File too large: %s (max %s)",
            formatBytes(size), formatBytes(maxSize))
    }

    // Type check
    fileType := file.Get("type").String()
    if len(allowedTypes) > 0 {
        allowed := false
        for _, pattern := range allowedTypes {
            if matchMimePattern(fileType, pattern) {
                allowed = true
                break
            }
        }
        if !allowed {
            return fmt.Errorf("File type %s not allowed", fileType)
        }
    }

    return nil
}

func formatBytes(bytes int64) string {
    if bytes < 1024 {
        return fmt.Sprintf("%d B", bytes)
    } else if bytes < 1024*1024 {
        return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
    } else {
        return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
    }
}
```

### Image Preview with URL.createObjectURL
```go
// Source: MDN Using files from web applications
//go:build js && wasm

func createImagePreview(file js.Value, containerID string) func() {
    // Check if image
    fileType := file.Get("type").String()
    if !strings.HasPrefix(fileType, "image/") {
        return func() {} // Not an image
    }

    doc := js.Global().Get("document")

    // Create blob URL
    objURL := js.Global().Get("URL").Call("createObjectURL", file).String()

    // Create img element
    img := doc.Call("createElement", "img")
    img.Set("src", objURL)
    img.Get("classList").Call("add", "upload-preview")
    img.Get("style").Set("maxWidth", "200px")
    img.Get("style").Set("maxHeight", "200px")

    // Append to container
    container := doc.Call("getElementById", containerID)
    container.Call("appendChild", img)

    // Return cleanup function
    return func() {
        js.Global().Get("URL").Call("revokeObjectURL", objURL)
        if !img.IsNull() && !img.IsUndefined() {
            parent := img.Get("parentNode")
            if !parent.IsNull() && !parent.IsUndefined() {
                parent.Call("removeChild", img)
            }
        }
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Flash-based uploaders | Native File API + XHR | 2010+ (HTML5) | No plugins, better security, mobile support |
| Synchronous file reading | Asynchronous FileReader | 2010+ (File API) | Non-blocking UI during large file reads |
| FileReader for preview | URL.createObjectURL | 2012+ (Blob URLs) | Synchronous, memory efficient, no base64 bloat |
| Fetch API for uploads | XMLHttpRequest | Always (Fetch limitation) | Only API with upload.progress events |
| Client-side validation only | Client + server validation | Always (security) | Client UX, server security - both required |

**Deprecated/outdated:**
- **Flash-based uploaders (Uploadify, SWFUpload):** Removed from browsers, security vulnerabilities
- **jQuery File Upload plugin:** Modern vanilla JS is simpler, fewer dependencies
- **FileReader.readAsDataURL for upload:** Encodes to base64 (33% size increase), use FormData instead
- **Synchronous XHR:** Deprecated, blocks main thread, terrible UX

## Open Questions

Things that couldn't be fully resolved:

1. **XHR upload progress reliability in Go/WASM**
   - What we know: MDN documents XHR.upload.progress as standard, webkit/gecko bugs with listener timing
   - What's unclear: TinyGo WASM syscall/js compatibility - need prototype validation in planning
   - Recommendation: Implement in 30-01 as prototype, validate progress events fire correctly before finalizing 30-02 component

2. **Multi-file upload handling**
   - What we know: File input supports `multiple` attribute, DataTransfer.files is a FileList
   - What's unclear: Whether phase 30 scope includes multi-file (requirements focus on "a file" singular)
   - Recommendation: Design API to support multiple files (OnFilesSelected vs OnFileSelected), implement single-file first, multi-file in future phase if needed

3. **Chunk upload for large files**
   - What we know: Files > 1GB may timeout, standard pattern is chunked upload with resumability
   - What's unclear: Does gux need chunk upload for v1.0, or acceptable limitation?
   - Recommendation: Defer to future phase - document "files < 500MB recommended" in initial release, add chunking if users request it

4. **Accessibility keyboard shortcuts**
   - What we know: Enter/Space should trigger file selection, Escape should cancel
   - What's unclear: Standard keyboard patterns for drag-drop zones (no ARIA spec)
   - Recommendation: Implement basic keyboard access (label+input pattern, focusable drop zone), defer custom shortcuts

## Sources

### Primary (HIGH confidence)
- MDN Web Docs:
  - [XMLHttpRequest.upload property](https://developer.mozilla.org/en-US/docs/Web/API/XMLHttpRequest/upload) - Upload progress events
  - [FormData.append()](https://developer.mozilla.org/en-US/docs/Web/API/FormData/append) - FormData construction
  - [Using files from web applications](https://developer.mozilla.org/en-US/docs/Web/API/File_API/Using_files_from_web_applications) - File API, URL.createObjectURL
  - [File drag and drop](https://developer.mozilla.org/en-US/docs/Web/API/HTML_Drag_and_Drop_API/File_drag_and_drop) - Drag-and-drop patterns
- Go standard library:
  - [syscall/js](https://pkg.go.dev/syscall/js) - JavaScript interop
  - [net/url](https://pkg.go.dev/net/url) - URL encoding
- Gux codebase patterns:
  - fetch/fetch.go - CSRF token handling
  - ui/video_player.go - SSR/WASM split pattern
  - core/app.go - State management

### Secondary (MEDIUM confidence)
- [Go and WebAssembly interacting with browser JS API](https://macias.info/entry/202003151900_go_wasm_js.md) - syscall/js patterns
- [A Guide to Go's syscall/js Package](https://reintech.io/blog/a-guide-to-gos-syscall-js-package-go-and-webassembly) - WASM JavaScript integration
- [HTML File Upload Accessibility](https://blog.filestack.com/html-file-upload-accessibility/) - WCAG compliance patterns
- [File drag and drop UX guidelines](https://smart-interface-design-patterns.com/articles/drag-and-drop-ux/) - Visual feedback best practices
- [Read Browser File Inputs with Go WebAssembly](https://donatstudios.com/Read-User-Files-With-Go-WASM) - File reading patterns

### Tertiary (LOW confidence)
- WebSearch results on WASM file upload - No comprehensive Go/WASM + XHR examples found, patterns must be synthesized
- [XHR upload progress bugs](https://bugzilla.mozilla.org/show_bug.cgi?id=908375) - Browser implementation inconsistencies (historical)
- [WASM memory copying discussions](https://news.ycombinator.com/item?id=28577739) - Community discussions on File API patterns

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - syscall/js is official, XMLHttpRequest is standard browser API
- Architecture: MEDIUM - Patterns based on gux VideoPlayer and MDN docs, but no existing Go/WASM XHR upload examples to reference
- Pitfalls: HIGH - Well-documented browser API gotchas from MDN and community experience

**Research date:** 2026-02-01
**Valid until:** 2026-03-01 (30 days - browser APIs stable, WASM ecosystem maturing)
