# Phase 29: Storage Foundation - Research

**Researched:** 2026-02-01
**Domain:** File upload, storage abstraction, multipart parsing, content hashing, file serving
**Confidence:** HIGH

## Summary

Storage Foundation implements server-side file upload and serving with a clean storage abstraction. Research focused on Go standard library patterns for multipart parsing, magic bytes validation, hash-based filenames, and aggressive caching. The phase leverages stdlib capabilities (mime/multipart, crypto/sha256, image.DecodeConfig) with minimal third-party dependencies.

The locked decisions from CONTEXT.md define a content-addressed storage system: hash-based filenames with 2-char prefix subdirectories, magic bytes validation, rich upload response metadata including image dimensions, and aggressive caching (1 year max-age with immutable directive). Implementation follows established gux patterns (functional options, interface abstractions, secure-by-default auth).

**Primary recommendation:** Use Go stdlib for core functionality (multipart parsing, SHA-256 hashing, image dimension detection), validate file types with magic bytes libraries (gabriel-vasile/mimetype), and follow the storage interface pattern for clean abstraction between app code and filesystem operations.

## Standard Stack

Go standard library provides robust primitives for file upload handling. Third-party libraries fill specific gaps (magic bytes validation, image dimensions).

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| mime/multipart | stdlib | Parse multipart form data | Official Go HTTP standard, memory-efficient streaming |
| crypto/sha256 | stdlib | Content hashing for filenames | Fast, secure, widely supported (no external deps) |
| net/http | stdlib | File serving, MaxBytesReader | Built-in HTTP primitives with range request support |
| image | stdlib | Image dimension detection | DecodeConfig reads dimensions without full decode |
| mime | stdlib | Extension to MIME type mapping | Built-in type registry for Content-Type headers |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/gabriel-vasile/mimetype | latest | Magic bytes validation | Required for secure MIME type detection (user-uploaded files) |
| golang.org/x/crypto/blake2b | latest | Optional fast hashing | Alternative to SHA-256 if performance critical |
| github.com/google/renameio | latest | Atomic file writes (POSIX) | Production deployments on Linux/Unix |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| SHA-256 | BLAKE2b | 50% faster hashing, but requires golang.org/x/crypto import |
| SHA-256 | BLAKE3 | 3x faster, but third-party dependency |
| gabriel-vasile/mimetype | h2non/filetype | Faster but less comprehensive format coverage |
| Standard mime package | http.DetectContentType | Only checks first 512 bytes, limited format support |

**Installation:**
```bash
go get github.com/gabriel-vasile/mimetype
```

BLAKE2b is optional (only if SHA-256 hashing becomes a bottleneck):
```bash
go get golang.org/x/crypto/blake2b
```

## Architecture Patterns

### Recommended Project Structure
```
core/
├── storage.go          # Storage interface definition
├── storage_local.go    # Local filesystem implementation
├── storage_test.go     # Storage backend tests
└── upload.go           # Upload endpoint handler

storage/                # Package-level convenience (optional)
└── local.go           # Re-exports core.NewLocalStorage
```

### Pattern 1: Storage Interface Abstraction
**What:** Interface-based abstraction decouples storage location from application code
**When to use:** Always - enables future S3/cloud backends without app code changes
**Example:**
```go
// Source: Go filesystem abstraction patterns (afero, sajari/storage)
package core

type Storage interface {
    // Put stores file bytes and returns provider-agnostic key (not full path/URL)
    Put(ctx context.Context, filename string, data io.Reader, size int64) (key string, url string, err error)

    // Delete removes file by key
    Delete(ctx context.Context, key string) error

    // URL returns the serving URL for a key
    URL(key string) string
}

// LocalStorage implements Storage for filesystem
type LocalStorage struct {
    baseDir string
    baseURL string  // e.g., "/__gux_api/files"
    maxSize int64
    allowedTypes []string
}

func NewLocalStorage(dir string, opts ...StorageOption) *LocalStorage {
    s := &LocalStorage{
        baseDir: dir,
        baseURL: "/__gux_api/files",
        maxSize: 0, // No limit by default
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

type StorageOption func(*LocalStorage)

func WithMaxSize(bytes int64) StorageOption {
    return func(s *LocalStorage) { s.maxSize = bytes }
}

func WithAllowedTypes(patterns ...string) StorageOption {
    return func(s *LocalStorage) { s.allowedTypes = patterns }
}
```

### Pattern 2: Multipart Parsing with Memory Limits
**What:** ParseMultipartForm with MaxBytesReader prevents memory exhaustion
**When to use:** All file upload endpoints
**Example:**
```go
// Source: https://freshman.tech/file-upload-golang/
// ParseMultipartForm limits in-memory storage, spills to temp files

// Limit request body size BEFORE parsing
r.Body = http.MaxBytesReader(w, r.Body, maxSize)

// Parse with memory limit (32MB in-memory, rest goes to temp files)
if err := r.ParseMultipartForm(32 << 20); err != nil {
    // Handle size limit exceeded
    return
}

// Access uploaded files
file, header, err := r.FormFile("upload")
if err != nil {
    return
}
defer file.Close()
```

### Pattern 3: Content Hash-Based Filenames with Prefix Sharding
**What:** Hash file content to generate collision-free filenames, use first 2 chars as subdirectory
**When to use:** All stored files (prevents filesystem bottleneck at 10,000+ files)
**Example:**
```go
// Source: gux existing pattern (server/spa.go), Go crypto/sha256
import "crypto/sha256"

// Hash file content
hash := sha256.New()
io.Copy(hash, fileReader)
hashHex := fmt.Sprintf("%x", hash.Sum(nil))

// Extract extension from validated MIME type
ext := mime.ExtensionsByType(detectedMimeType)[0]

// Build key: first 2 chars as subdirectory
prefix := hashHex[:2]
filename := hashHex + ext
key := filepath.Join(prefix, filename) // "ab/abc123def456.jpg"

// Create directory and write file
fullPath := filepath.Join(baseDir, key)
os.MkdirAll(filepath.Dir(fullPath), 0755)
```

### Pattern 4: Magic Bytes Validation
**What:** Validate file type by reading magic bytes (first few bytes), reject mismatches
**When to use:** All user uploads (prevents malicious file upload attacks)
**Example:**
```go
// Source: https://github.com/gabriel-vasile/mimetype
import "github.com/gabriel-vasile/mimetype"

// Detect MIME type from file content
mtype, err := mimetype.DetectReader(file)
if err != nil {
    return err
}

// Check against allowed types (from storage config or per-field rules)
if !isAllowed(mtype.String(), allowedTypes) {
    return fmt.Errorf("file type %s not allowed", mtype.String())
}

// Extract extension for hash-based filename
ext := mtype.Extension() // Returns ".jpg", ".png", etc.
```

### Pattern 5: Image Dimension Detection
**What:** Use image.DecodeConfig to read width/height without decoding full image
**When to use:** All uploaded images (for rich response metadata)
**Example:**
```go
// Source: https://pkg.go.dev/image
import (
    "image"
    _ "image/jpeg" // Register JPEG decoder
    _ "image/png"  // Register PNG decoder
    _ "image/gif"  // Register GIF decoder
)

// Detect dimensions efficiently
config, format, err := image.DecodeConfig(file)
if err != nil {
    // Not an image or unsupported format
    return 0, 0, err
}

// config.Width, config.Height available without full decode
return config.Width, config.Height, nil
```

### Pattern 6: Content-Addressed File Serving with Aggressive Caching
**What:** Serve files with immutable cache headers since hash = content identity
**When to use:** All file serving endpoints
**Example:**
```go
// Source: gux server/spa.go, https://www.debugbear.com/docs/http-cache-control-header

// Hash-based URLs never change content, cache forever
w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

// Content-Disposition based on MIME type
if isInlineType(mimeType) {
    w.Header().Set("Content-Disposition", "inline")
} else {
    w.Header().Set("Content-Disposition", "attachment")
}

func isInlineType(mimeType string) bool {
    // Safe types to display in browser
    safe := []string{
        "image/png", "image/jpeg", "image/gif", "image/webp",
        "video/mp4", "video/webm", "application/pdf",
    }
    for _, t := range safe {
        if strings.HasPrefix(mimeType, t) {
            return true
        }
    }
    return false
}
```

### Pattern 7: Functional Options for Configuration
**What:** Variadic options pattern for flexible, backward-compatible APIs
**When to use:** Storage configuration (gux convention)
**Example:**
```go
// Source: gux core/crud.go, https://golang.cafe/blog/golang-functional-options-pattern

// Following gux patterns (core.WithRoles, core.WithPublic, etc.)
app.SetStorage(storage.NewLocalStorage("./uploads",
    storage.WithMaxSize(10*1024*1024), // 10MB
    storage.WithAllowedTypes("image/*", "application/pdf"),
    storage.WithPublic(), // Allow unauthenticated serving
))

// Option type
type StorageOption func(*LocalStorage)

// Required args explicit, optional args via options
func NewLocalStorage(dir string, opts ...StorageOption) *LocalStorage {
    s := &LocalStorage{
        baseDir: dir,
        // Sensible defaults
        maxSize: 0, // No limit
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

### Anti-Patterns to Avoid
- **Storing full file paths in database:** Use provider-agnostic keys (e.g., "ab/abc123.jpg") not "/var/uploads/file.jpg" - enables storage backend migration
- **Trusting Content-Type header:** User can set arbitrary Content-Type - always validate with magic bytes
- **No size limits by default:** Forces developers to think about DoS protection (gux philosophy: no surprising defaults)
- **Blocking I/O during hash computation:** Use io.TeeReader to hash while writing to disk, not separate read pass
- **Overwriting files without atomic write:** Use temp file + rename on POSIX systems to prevent corruption

## Don't Hand-Roll

File handling has security and correctness edge cases. Use established solutions.

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| MIME type detection | String matching on extensions | github.com/gabriel-vasile/mimetype | Magic bytes prevent `.jpg.exe` attacks, handles 120+ formats |
| Image dimensions | Custom parsers for JPEG/PNG | image.DecodeConfig (stdlib) | Supports JPEG, PNG, GIF, handles EXIF rotation |
| Atomic file writes | `os.Create` + write | Temp file + os.Rename (or github.com/google/renameio) | Prevents partial writes on crash/power loss |
| Content hashing | Custom hash streaming | crypto/sha256.New() with io.Copy | Optimized assembly implementations (AES-NI on x86) |
| Multipart parsing | Manual boundary detection | http.Request.ParseMultipartForm | Handles edge cases, memory limits, temp file spillover |
| File serving | Custom byte streaming | http.ServeContent (stdlib) | Range requests, If-Modified-Since, proper headers |

**Key insight:** File upload security is non-trivial. Go stdlib provides production-grade primitives; mimetype library is industry-standard for content validation. Rolling custom solutions introduces vulnerabilities (path traversal, zip bombs, SSRF via SVG, EXIF exploits).

## Common Pitfalls

### Pitfall 1: Memory Exhaustion from Large Uploads
**What goes wrong:** Reading entire file into memory before validation causes OOM
**Why it happens:** ParseMultipartForm without memory limit or MaxBytesReader
**How to avoid:** Always use MaxBytesReader BEFORE ParseMultipartForm, set reasonable memory limit
**Warning signs:** Server crashes under load testing with large files

### Pitfall 2: Path Traversal via Filename
**What goes wrong:** Using user-supplied filename directly allows `../../etc/passwd` attacks
**Why it happens:** Not sanitizing header.Filename before filepath.Join
**How to avoid:** Use content hash for storage filename, preserve original name only in response metadata
**Warning signs:** Security scanner flags directory traversal vulnerability

### Pitfall 3: Extension Spoofing (MIME Confusion)
**What goes wrong:** Uploading `malware.exe` renamed to `image.jpg` bypasses extension checks
**Why it happens:** Trusting file extension or Content-Type header instead of magic bytes
**How to avoid:** Always validate with mimetype.DetectReader, reject mismatches
**Warning signs:** Executables stored with image extensions

### Pitfall 4: Filesystem Bottleneck (Too Many Files in One Directory)
**What goes wrong:** ext4/XFS slow down with 10,000+ files in single directory
**Why it happens:** Storing all uploads flat in one directory
**How to avoid:** Use 2-char hash prefix as subdirectory (256 buckets)
**Warning signs:** Upload/serve performance degrades as file count grows

### Pitfall 5: Non-Atomic File Writes
**What goes wrong:** Server crash mid-write leaves corrupted zero-byte or partial files
**Why it happens:** Writing directly to final path with os.Create
**How to avoid:** Write to temp file, then os.Rename (atomic on POSIX)
**Warning signs:** Random zero-byte files after crashes

### Pitfall 6: Missing CSRF Protection on Upload
**What goes wrong:** Malicious site uploads files to user's account
**Why it happens:** Upload endpoint not integrated with gux CSRF middleware
**How to avoid:** Upload endpoint respects app.csrf automatically (same as CRUD)
**Warning signs:** CSRF test suite fails for upload endpoint

### Pitfall 7: Caching Non-Immutable Content
**What goes wrong:** File updates don't reflect in browser due to aggressive caching
**Why it happens:** Using long max-age without content hashing
**How to avoid:** Only use immutable + long max-age for hash-based URLs
**Warning signs:** Users report stale files after updates

## Code Examples

Verified patterns from official sources and gux conventions.

### Multipart Upload Handler
```go
// Source: https://freshman.tech/file-upload-golang/
func handleUpload(w http.ResponseWriter, r *http.Request) {
    // Limit request body size (10MB example)
    maxSize := int64(10 << 20)
    r.Body = http.MaxBytesReader(w, r.Body, maxSize)

    // Parse multipart form (32MB in-memory limit)
    if err := r.ParseMultipartForm(32 << 20); err != nil {
        http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
        return
    }
    defer r.MultipartForm.RemoveAll() // Clean up temp files

    // Get uploaded file
    file, header, err := r.FormFile("upload")
    if err != nil {
        http.Error(w, "Missing file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Validate, hash, store...
}
```

### Magic Bytes Validation
```go
// Source: https://github.com/gabriel-vasile/mimetype
import "github.com/gabriel-vasile/mimetype"

func validateFileType(file io.Reader, allowedTypes []string) (string, string, error) {
    // Detect MIME type from content
    mtype, err := mimetype.DetectReader(file)
    if err != nil {
        return "", "", err
    }

    // Check against whitelist
    allowed := false
    for _, pattern := range allowedTypes {
        if matchMimePattern(mtype.String(), pattern) {
            allowed = true
            break
        }
    }

    if !allowed {
        return "", "", fmt.Errorf("file type %s not allowed", mtype.String())
    }

    return mtype.String(), mtype.Extension(), nil
}

func matchMimePattern(mimeType, pattern string) bool {
    // "image/*" matches "image/jpeg", "image/png", etc.
    if strings.HasSuffix(pattern, "/*") {
        prefix := strings.TrimSuffix(pattern, "/*")
        return strings.HasPrefix(mimeType, prefix+"/")
    }
    return mimeType == pattern
}
```

### Content Hash Storage
```go
// Source: crypto/sha256, gux server/spa.go patterns
func storeFile(baseDir string, file io.Reader, ext string) (key string, err error) {
    // Hash content while writing to temp file
    tempFile, err := os.CreateTemp(baseDir, "upload-*")
    if err != nil {
        return "", err
    }
    defer os.Remove(tempFile.Name()) // Clean up temp

    hash := sha256.New()
    tee := io.TeeReader(file, hash)

    if _, err := io.Copy(tempFile, tee); err != nil {
        return "", err
    }
    tempFile.Close()

    // Generate hash-based filename
    hashHex := fmt.Sprintf("%x", hash.Sum(nil))
    prefix := hashHex[:2]
    filename := hashHex + ext
    key = filepath.Join(prefix, filename)

    // Create subdirectory
    fullPath := filepath.Join(baseDir, key)
    if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
        return "", err
    }

    // Atomic move (rename is atomic on POSIX)
    if err := os.Rename(tempFile.Name(), fullPath); err != nil {
        return "", err
    }

    return key, nil
}
```

### File Serving Handler
```go
// Source: gux server/spa.go, net/http.ServeContent
func serveFile(w http.ResponseWriter, r *http.Request, key string, baseDir string) {
    fullPath := filepath.Join(baseDir, key)

    file, err := os.Open(fullPath)
    if err != nil {
        http.NotFound(w, r)
        return
    }
    defer file.Close()

    stat, err := file.Stat()
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Detect MIME type from extension
    ext := filepath.Ext(key)
    mimeType := mime.TypeByExtension(ext)
    if mimeType == "" {
        mimeType = "application/octet-stream"
    }
    w.Header().Set("Content-Type", mimeType)

    // Aggressive caching (content hash = immutable)
    w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

    // Inline vs attachment
    if isInlineType(mimeType) {
        w.Header().Set("Content-Disposition", "inline")
    } else {
        w.Header().Set("Content-Disposition", "attachment")
    }

    // Serve with range request support
    http.ServeContent(w, r, key, stat.ModTime(), file)
}
```

### Image Dimension Detection
```go
// Source: https://pkg.go.dev/image
import (
    "image"
    _ "image/gif"
    _ "image/jpeg"
    _ "image/png"
)

func getImageDimensions(file io.Reader) (width, height int, err error) {
    config, _, err := image.DecodeConfig(file)
    if err != nil {
        return 0, 0, err
    }
    return config.Width, config.Height, nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| UUID filenames | Content hash filenames | 2015+ (CDN era) | Automatic deduplication, immutable URLs enable aggressive caching |
| Extension validation | Magic bytes validation | 2010+ (post web security boom) | Prevents file type spoofing attacks |
| Single directory storage | Hash-prefix sharding | 2010+ (filesystem limits) | Prevents directory listing slowdown at 10k+ files |
| SHA-1 hashing | SHA-256 hashing | 2017+ (SHAttered attack) | SHA-1 collision attacks make it unsuitable for content addressing |
| Manual multipart parsing | ParseMultipartForm | Always (stdlib) | Memory limits, temp file spillover, edge case handling |

**Deprecated/outdated:**
- **MD5 for content hashing**: Collision attacks make it unsuitable for security-sensitive scenarios
- **SHA-1 for content hashing**: SHAttered attack (2017) demonstrates practical collisions
- **Trusting file extensions**: Trivially bypassed, must use magic bytes
- **http.DetectContentType alone**: Only checks first 512 bytes, limited format support

## Open Questions

Things that couldn't be fully resolved:

1. **Optimal hash prefix length (2 chars vs 4 chars)**
   - What we know: 2 chars = 256 buckets (common pattern), 4 chars = 65,536 buckets
   - What's unclear: Performance tradeoff point - when does 256 buckets become insufficient?
   - Recommendation: Start with 2 chars (proven pattern), make configurable for future tuning

2. **SHA-256 vs BLAKE2b performance impact**
   - What we know: BLAKE2b is 50% faster, SHA-256 has hardware acceleration (AES-NI)
   - What's unclear: Real-world difference with modern CPUs - is hashing the bottleneck?
   - Recommendation: Use SHA-256 (stdlib, no deps), add BLAKE2b option if profiling shows hashing bottleneck

3. **Atomic write handling on Windows**
   - What we know: os.Rename is NOT atomic on Windows (unlike POSIX)
   - What's unclear: Does gux need Windows server support? Is renameio's complexity justified?
   - Recommendation: Use stdlib os.Rename for initial implementation (works for Linux/Mac), add github.com/google/renameio if Windows server support needed

## Sources

### Primary (HIGH confidence)
- Go standard library documentation:
  - [mime/multipart](https://pkg.go.dev/mime/multipart) - Multipart form parsing
  - [crypto/sha256](https://pkg.go.dev/crypto/sha256) - SHA-256 hashing
  - [image](https://pkg.go.dev/image) - Image dimension detection
  - [mime](https://pkg.go.dev/mime) - MIME type to extension mapping
- [gabriel-vasile/mimetype](https://github.com/gabriel-vasile/mimetype) - Magic bytes MIME detection
- [golang.org/x/crypto/blake2b](https://pkg.go.dev/golang.org/x/crypto/blake2b) - BLAKE2 hashing
- Gux codebase patterns:
  - server/spa.go - Content hashing, aggressive caching, file serving
  - core/crud.go - Functional options pattern
  - core/app.go - App configuration

### Secondary (MEDIUM confidence)
- [How to process file uploads in Go (freshman.tech)](https://freshman.tech/file-upload-golang/) - Multipart best practices
- [Golang multipart upload guide (Sling Academy)](https://www.slingacademy.com/article/using-multipart-requests-for-file-uploads-in-go/) - Memory management patterns
- [Cache-Control header guide (DebugBear)](https://www.debugbear.com/docs/http-cache-control-header) - Immutable directive usage
- [Functional Options Pattern guide (golang.cafe)](https://golang.cafe/blog/golang-functional-options-pattern) - Options pattern best practices
- [10 years of functional options (BytesizeGo)](https://www.bytesizego.com/blog/10-years-functional-options-golang) - Lessons learned
- [Atomically writing files in Go (Michael Stapelberg)](https://michael.stapelberg.ch/posts/2017-01-28-golang_atomically_writing/) - Atomic write patterns

### Tertiary (LOW confidence)
- [BLAKE2 vs SHA-256 comparison (hash.tools)](https://www.hash.tools/115/cryptographic-algorithms/13927/comparative-analysis-of-hash-functions-sha-256-vs-blake2) - Performance claims
- [Content-Disposition security discussion (GitHub pocketbase)](https://github.com/pocketbase/pocketbase/discussions/164) - Inline vs attachment decisions
- WebSearch results for hash performance - No single authoritative source, marked for validation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All recommendations based on Go stdlib documentation and established libraries
- Architecture: HIGH - Patterns verified in gux codebase and official Go documentation
- Pitfalls: HIGH - Well-documented security issues with official sources

**Research date:** 2026-02-01
**Valid until:** 2026-03-01 (30 days - stable domain, stdlib rarely changes)
