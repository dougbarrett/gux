package server

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// SPAHandler serves static files and falls back to index.html for client-side routing.
// This enables single-page applications to handle their own routing.
//
// It supports two modes:
//   - Filesystem mode: serves files from a directory on disk (for development)
//   - Embedded mode: serves files from an embed.FS (for single-binary deployment)
//
// When a WASM hash is configured, the handler automatically:
//   - Injects the hash into index.html (replacing main.wasm with main.<hash>.wasm)
//   - Routes requests for main.<hash>.wasm back to the embedded main.wasm
type SPAHandler struct {
	// fs is the filesystem to serve from (can be os.DirFS or embed.FS)
	fs fs.FS

	// wasmHash is the content hash of main.wasm (empty = no hash replacement)
	wasmHash string

	// cachedIndex is the pre-processed index.html with hash injected
	cachedIndex []byte

	// legacyDir is for backwards compatibility with NewSPAHandler(dir string)
	legacyDir string
}

// NewSPAHandler creates a new SPA handler for the given directory.
// This is the legacy constructor for filesystem-based serving (development mode).
func NewSPAHandler(dir string) *SPAHandler {
	return &SPAHandler{
		legacyDir: dir,
	}
}

// NewEmbeddedSPAHandler creates an SPA handler that serves from an embedded filesystem.
// This enables single-binary deployment with all static assets bundled in.
//
// The fsys should be the embedded filesystem (e.g., from //go:embed public/*).
// The subdir is the subdirectory within the embed.FS (e.g., "public").
//
// If main.wasm exists in the filesystem, a content hash is computed and:
//   - index.html is served with main.wasm replaced by main.<hash>.wasm
//   - Requests for main.<hash>.wasm are mapped back to main.wasm
//
// This provides cache-busting without modifying source files.
func NewEmbeddedSPAHandler(fsys fs.FS, subdir string) *SPAHandler {
	// Get the subdirectory as the root
	var rootFS fs.FS
	var err error
	if subdir != "" {
		rootFS, err = fs.Sub(fsys, subdir)
		if err != nil {
			// Fall back to root if subdir doesn't exist
			rootFS = fsys
		}
	} else {
		rootFS = fsys
	}

	h := &SPAHandler{
		fs: rootFS,
	}

	// Compute WASM hash if main.wasm exists
	if wasmFile, err := rootFS.Open("main.wasm"); err == nil {
		defer wasmFile.Close()
		hash := sha256.New()
		if _, err := io.Copy(hash, wasmFile); err == nil {
			h.wasmHash = fmt.Sprintf("%x", hash.Sum(nil))[:8]
		}
	}

	// Pre-process index.html with hash if we have one
	if h.wasmHash != "" {
		if indexData, err := fs.ReadFile(rootFS, "index.html"); err == nil {
			// Replace main.wasm with main.<hash>.wasm
			content := string(indexData)
			content = strings.ReplaceAll(content, `"main.wasm"`, fmt.Sprintf(`"main.%s.wasm"`, h.wasmHash))
			content = strings.ReplaceAll(content, `"/main.wasm"`, fmt.Sprintf(`"/main.%s.wasm"`, h.wasmHash))
			h.cachedIndex = []byte(content)
		}
	}

	return h
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Use legacy filesystem mode if configured
	if h.legacyDir != "" {
		h.serveLegacy(w, r)
		return
	}

	// Embedded filesystem mode
	h.serveEmbedded(w, r)
}

// serveLegacy handles requests using the legacy filesystem-based approach
func (h *SPAHandler) serveLegacy(w http.ResponseWriter, r *http.Request) {
	urlPath := filepath.Clean(r.URL.Path)
	if urlPath == "/" {
		urlPath = "/index.html"
	}

	fullPath := filepath.Join(h.legacyDir, urlPath)

	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		// File doesn't exist or is a directory - serve index.html for SPA routing
		h.serveFileLegacy(w, r, filepath.Join(h.legacyDir, "index.html"))
		return
	}

	h.serveFileLegacy(w, r, fullPath)
}

func (h *SPAHandler) serveFileLegacy(w http.ResponseWriter, r *http.Request, filePath string) {
	setContentType(w, filePath)
	http.ServeFile(w, r, filePath)
}

// serveEmbedded handles requests using the embedded filesystem
func (h *SPAHandler) serveEmbedded(w http.ResponseWriter, r *http.Request) {
	urlPath := path.Clean(r.URL.Path)
	if urlPath == "/" || urlPath == "" {
		urlPath = "index.html"
	} else {
		urlPath = strings.TrimPrefix(urlPath, "/")
	}

	// Handle index.html with hash injection
	if urlPath == "index.html" && h.cachedIndex != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(h.cachedIndex)
		return
	}

	// Handle hashed WASM request (main.<hash>.wasm -> main.wasm)
	if h.wasmHash != "" && urlPath == fmt.Sprintf("main.%s.wasm", h.wasmHash) {
		urlPath = "main.wasm"
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}

	// Try to open the file
	file, err := h.fs.Open(urlPath)
	if err != nil {
		// File not found - serve index.html for SPA routing
		if h.cachedIndex != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache")
			w.Write(h.cachedIndex)
			return
		}
		// Fallback to raw index.html
		if indexData, err := fs.ReadFile(h.fs, "index.html"); err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(indexData)
			return
		}
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	// Get file info for ServeContent
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// If it's a directory, serve index.html
	if stat.IsDir() {
		if h.cachedIndex != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache")
			w.Write(h.cachedIndex)
			return
		}
		http.NotFound(w, r)
		return
	}

	// Set content type
	setContentType(w, urlPath)

	// Serve the file
	if seeker, ok := file.(io.ReadSeeker); ok {
		http.ServeContent(w, r, urlPath, stat.ModTime(), seeker)
	} else {
		// Fallback for non-seekable files
		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Write(data)
	}
}

// setContentType sets the Content-Type header based on file extension
func setContentType(w http.ResponseWriter, filePath string) {
	ext := strings.ToLower(path.Ext(filePath))
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".mjs":
		w.Header().Set("Content-Type", "application/javascript")
	case ".wasm":
		w.Header().Set("Content-Type", "application/wasm")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	case ".pdf":
		w.Header().Set("Content-Type", "application/pdf")
	case ".woff":
		w.Header().Set("Content-Type", "font/woff")
	case ".woff2":
		w.Header().Set("Content-Type", "font/woff2")
	case ".ttf":
		w.Header().Set("Content-Type", "font/ttf")
	case ".eot":
		w.Header().Set("Content-Type", "application/vnd.ms-fontobject")
	case ".xml":
		w.Header().Set("Content-Type", "application/xml")
	case ".txt":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	case ".md":
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	case ".mp4":
		w.Header().Set("Content-Type", "video/mp4")
	case ".webm":
		w.Header().Set("Content-Type", "video/webm")
	case ".mp3":
		w.Header().Set("Content-Type", "audio/mpeg")
	case ".wav":
		w.Header().Set("Content-Type", "audio/wav")
	case ".ogg":
		w.Header().Set("Content-Type", "audio/ogg")
	}
}
