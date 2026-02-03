package core

import (
	"context"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

// Storage defines the interface for file storage backends.
type Storage interface {
	// Put stores a file and returns metadata about the upload.
	Put(ctx context.Context, filename string, data io.Reader, size int64) (*UploadResult, error)

	// Delete removes a file by its storage key. Returns nil if file doesn't exist (idempotent).
	Delete(ctx context.Context, key string) error

	// URL returns the serving URL for a stored file.
	URL(key string) string

	// Serve opens a file for serving. Returns the file, its info, and any error.
	Serve(key string) (io.ReadSeekCloser, os.FileInfo, error)
}

// UploadResult contains metadata about an uploaded file.
type UploadResult struct {
	Key         string `json:"key"`          // Storage key (provider-specific)
	URL         string `json:"url"`          // Serving URL
	Filename    string `json:"filename"`     // Original filename
	Size        int64  `json:"size"`         // File size in bytes
	ContentType string `json:"content_type"` // MIME type
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

// StorageOption configures storage behavior.
type StorageOption func(*storageConfig)

type storageConfig struct {
	maxSize      int64
	allowedTypes []string
	serveAuth    bool
	baseURL      string
}

// WithMaxSize sets the maximum file size in bytes.
func WithMaxSize(bytes int64) StorageOption {
	return func(c *storageConfig) {
		c.maxSize = bytes
	}
}

// WithAllowedTypes restricts uploads to specific MIME type patterns.
// Patterns support wildcards (e.g., "image/*", "application/pdf").
func WithAllowedTypes(patterns ...string) StorageOption {
	return func(c *storageConfig) {
		c.allowedTypes = patterns
	}
}

// WithServeAuth requires authentication for the file serving endpoint.
func WithServeAuth() StorageOption {
	return func(c *storageConfig) {
		c.serveAuth = true
	}
}

// WithBaseURL overrides the default serving base URL.
// Default: "/__gux_api/files"
func WithBaseURL(url string) StorageOption {
	return func(c *storageConfig) {
		c.baseURL = url
	}
}

// StorageError represents a storage validation error.
type StorageError struct {
	Code    string `json:"error"`
	Message string `json:"message"`
	MaxSize int64  `json:"max_size,omitempty"`
}

func (e *StorageError) Error() string {
	return e.Message
}

// FileInfo represents file metadata exposed in DTOs for file fields.
// Models store a storage key (string), but DTOs expose FileInfo with
// resolved URL and metadata for rendering in admin views.
type FileInfo struct {
	URL         string `json:"url"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

// FileInfoFromKey builds a FileInfo from a storage key using the Storage interface.
// Returns nil if key is empty. Uses Serve() to get file metadata (size, content type).
// If Serve fails (file missing), returns a FileInfo with just the URL and filename derived from key.
func FileInfoFromKey(s Storage, key string) *FileInfo {
	if key == "" || s == nil {
		return nil
	}
	info := &FileInfo{
		URL:      s.URL(key),
		Filename: extractFilename(key),
	}
	// Try to get full metadata from storage
	reader, fileInfo, err := s.Serve(key)
	if err == nil {
		reader.Close()
		info.Size = fileInfo.Size()
		// Content type detection from filename
		info.ContentType = detectContentType(key)
	}
	return info
}

// extractFilename extracts the filename from a storage key.
// Returns the last path segment (the hash-based filename).
func extractFilename(key string) string {
	if idx := strings.LastIndex(key, "/"); idx >= 0 {
		return key[idx+1:]
	}
	return key
}

// detectContentType detects MIME type from file extension.
func detectContentType(key string) string {
	ext := filepath.Ext(key)
	if ext == "" {
		return "application/octet-stream"
	}
	ct := mime.TypeByExtension(ext)
	if ct == "" {
		return "application/octet-stream"
	}
	return ct
}

// matchMimePattern checks if a MIME type matches a pattern.
// Supports wildcards like "image/*".
func matchMimePattern(mimeType, pattern string) bool {
	// Exact match
	if mimeType == pattern {
		return true
	}

	// Wildcard match
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(mimeType, prefix+"/")
	}

	return false
}
