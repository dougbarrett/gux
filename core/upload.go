package core

import (
	"encoding/json"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/dougbarrett/gux/api"
)

// registerUploadHandlers registers upload and file serving endpoints.
func (a *App) registerUploadHandlers(mux *http.ServeMux) {
	// Upload endpoint: POST /__gux_api/upload
	mux.HandleFunc("POST "+a.apiPrefix+"/upload", a.handleUpload)

	// File serving endpoint: GET /__gux_api/files/
	mux.HandleFunc(a.apiPrefix+"/files/", a.handleServeFile)
}

// handleUpload processes multipart file uploads.
func (a *App) handleUpload(w http.ResponseWriter, r *http.Request) {
	// Auth check: if auth is configured, require authentication
	if a.authConfig != nil {
		user := a.getUserFromRequest(r)
		if user == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			return
		}
	}

	// CSRF is already handled by CSRFMiddleware wrapping the mux

	// Parse multipart form (32MB max in memory)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_request",
			"message": "Failed to parse multipart form",
		})
		return
	}
	defer r.MultipartForm.RemoveAll()

	// Get uploaded files
	if r.MultipartForm == nil || r.MultipartForm.File == nil || len(r.MultipartForm.File) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "no_files",
			"message": "No files in upload request",
		})
		return
	}

	// Collect results from all files
	var results []*UploadResult

	// Process each form field (can have multiple files per field)
	for _, headers := range r.MultipartForm.File {
		for _, header := range headers {
			// Open uploaded file
			file, err := header.Open()
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "file_error",
					"message": "Failed to open uploaded file",
				})
				return
			}

			// Upload to storage
			result, err := a.storage.Put(r.Context(), header.Filename, file, header.Size)
			file.Close()

			// Check for storage errors
			if err != nil {
				// If it's a StorageError, return it directly
				if storageErr, ok := err.(*StorageError); ok {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(storageErr)
					return
				}
				// Other errors
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "upload_failed",
					"message": err.Error(),
				})
				return
			}

			results = append(results, result)
		}
	}

	// Return array of upload results
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

// handleServeFile serves files from storage.
func (a *App) handleServeFile(w http.ResponseWriter, r *http.Request) {
	// Auth check (optional): if storage requires auth, check session
	if ls, ok := a.storage.(*LocalStorage); ok && ls.serveAuth {
		if a.authConfig != nil {
			user := a.getUserFromRequest(r)
			if user == nil {
				api.WriteError(w, api.Unauthorized("authentication required"))
				return
			}
		}
	}

	// Extract key from URL path
	key := strings.TrimPrefix(r.URL.Path, a.apiPrefix+"/files/")

	// Path traversal protection
	if strings.Contains(key, "..") || strings.HasPrefix(key, "/") {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Serve file from storage
	file, info, err := a.storage.Serve(key)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	// Detect MIME type from file extension
	ext := filepath.Ext(key)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Set headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	// Set Content-Disposition based on type
	if isInlineType(contentType) {
		w.Header().Set("Content-Disposition", "inline")
	} else {
		w.Header().Set("Content-Disposition", "attachment")
	}

	// Serve the file with Range request support
	http.ServeContent(w, r, key, info.ModTime(), file)
}

// isInlineType returns true if the MIME type should be displayed inline.
func isInlineType(mimeType string) bool {
	// Images
	if strings.HasPrefix(mimeType, "image/") {
		return true
	}
	// Videos
	if mimeType == "video/mp4" || mimeType == "video/webm" {
		return true
	}
	// PDFs
	if mimeType == "application/pdf" {
		return true
	}
	return false
}
