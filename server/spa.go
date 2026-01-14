package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// SPAHandler serves static files and falls back to index.html for client-side routing.
// This enables single-page applications to handle their own routing.
type SPAHandler struct {
	// StaticDir is the directory containing static files
	StaticDir string
}

// NewSPAHandler creates a new SPA handler for the given directory
func NewSPAHandler(dir string) *SPAHandler {
	return &SPAHandler{StaticDir: dir}
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean(r.URL.Path)
	if path == "/" {
		path = "/index.html"
	}

	fullPath := filepath.Join(h.StaticDir, path)

	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		// File doesn't exist or is a directory - serve index.html for SPA routing
		h.serveFile(w, r, filepath.Join(h.StaticDir, "index.html"))
		return
	}

	h.serveFile(w, r, fullPath)
}

func (h *SPAHandler) serveFile(w http.ResponseWriter, r *http.Request, path string) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".js":
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
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	}

	http.ServeFile(w, r, path)
}
