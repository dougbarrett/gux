package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	apiserver "goquery/internal/api/server"
)

func main() {
	port := flag.Int("port", 8080, "Port to serve on")
	dir := flag.String("dir", ".", "Directory to serve")
	flag.Parse()

	// Create mux for routing
	mux := http.NewServeMux()

	// Register API handlers
	postsHandler := apiserver.NewPostsHandler()
	postsHandler.RegisterRoutes(mux)

	// SPA handler for everything else
	spaHandler := &SPAHandler{StaticDir: *dir}
	mux.HandleFunc("/", spaHandler.ServeHTTP)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Starting server at http://localhost%s\n", addr)
	fmt.Printf("Serving files from: %s\n", *dir)
	fmt.Println("API endpoints:")
	fmt.Println("  GET    /api/posts      - List all posts")
	fmt.Println("  GET    /api/posts/:id  - Get post by ID")
	fmt.Println("  POST   /api/posts      - Create post")
	fmt.Println("  PUT    /api/posts/:id  - Update post")
	fmt.Println("  DELETE /api/posts/:id  - Delete post")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// SPAHandler serves static files and falls back to index.html for SPA routing
type SPAHandler struct {
	StaticDir string
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clean the path
	path := filepath.Clean(r.URL.Path)
	if path == "/" {
		path = "/index.html"
	}

	// Build full file path
	fullPath := filepath.Join(h.StaticDir, path)

	// Check if file exists
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		// File doesn't exist or is a directory - serve index.html for SPA
		h.serveFile(w, r, filepath.Join(h.StaticDir, "index.html"))
		return
	}

	// Serve the actual file
	h.serveFile(w, r, fullPath)
}

func (h *SPAHandler) serveFile(w http.ResponseWriter, r *http.Request, path string) {
	// Set content type based on extension
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

	http.ServeFile(w, r, path)
}
