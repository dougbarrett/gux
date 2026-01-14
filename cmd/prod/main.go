package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed dist/*
var staticFiles embed.FS

func main() {
	port := flag.Int("port", 8080, "Port to serve on")
	flag.Parse()

	// Get the dist subdirectory
	distFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		log.Fatal(err)
	}

	handler := &SPAHandler{FS: distFS}

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Production server running at http://localhost%s\n", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

// SPAHandler serves from an embedded filesystem with SPA fallback
type SPAHandler struct {
	FS fs.FS
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	// Try to open the file
	f, err := h.FS.Open(path)
	if err != nil {
		// Fall back to index.html for SPA routing
		path = "index.html"
	} else {
		f.Close()
	}

	// Set content type
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
	}

	http.ServeFileFS(w, r, h.FS, path)
}
