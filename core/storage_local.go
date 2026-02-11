//go:build !js || !wasm

package core

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
)

// LocalStorage implements the Storage interface using the local filesystem.
type LocalStorage struct {
	baseDir      string
	baseURL      string
	maxSize      int64
	allowedTypes []string
	serveAuth    bool
}

// NewLocalStorage creates a new local filesystem storage backend.
// The directory is created if it doesn't exist.
func NewLocalStorage(dir string, opts ...StorageOption) *LocalStorage {
	// Apply defaults
	cfg := storageConfig{
		baseURL: "/__gux_api/files",
	}

	// Apply options
	for _, opt := range opts {
		opt(&cfg)
	}

	// Create directory if it doesn't exist
	_ = os.MkdirAll(dir, 0755)

	return &LocalStorage{
		baseDir:      dir,
		baseURL:      cfg.baseURL,
		maxSize:      cfg.maxSize,
		allowedTypes: cfg.allowedTypes,
		serveAuth:    cfg.serveAuth,
	}
}

// Put stores a file and returns metadata.
func (s *LocalStorage) Put(ctx context.Context, filename string, data io.Reader, size int64) (*UploadResult, error) {
	// Validate size
	if s.maxSize > 0 && size > s.maxSize {
		return nil, &StorageError{
			Code:    "file_too_large",
			Message: fmt.Sprintf("File size %d bytes exceeds maximum %d bytes", size, s.maxSize),
			MaxSize: s.maxSize,
		}
	}

	// Read file content (need for hashing and MIME detection)
	reader := data
	if s.maxSize > 0 {
		reader = io.LimitReader(data, s.maxSize)
	}
	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Detect MIME type
	mtype := mimetype.Detect(buf)
	contentType := mtype.String()

	// Validate allowed types
	if len(s.allowedTypes) > 0 {
		allowed := false
		for _, pattern := range s.allowedTypes {
			if matchMimePattern(contentType, pattern) {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, &StorageError{
				Code:    "invalid_file_type",
				Message: fmt.Sprintf("File type %s not allowed", contentType),
			}
		}
	}

	// Compute SHA-256 hash
	hash := sha256.Sum256(buf)
	hashHex := hex.EncodeToString(hash[:])

	// Get extension from detected MIME type
	ext := mtype.Extension()

	// Build key: prefix/hashHex+ext (prefix = first 2 chars of hash)
	prefix := hashHex[:2]
	key := filepath.Join(prefix, hashHex+ext)

	// Create subdirectory
	subdir := filepath.Join(s.baseDir, prefix)
	if err := os.MkdirAll(subdir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create subdirectory: %w", err)
	}

	// Write file atomically (temp file + rename)
	destPath := filepath.Join(s.baseDir, key)
	tempPath := destPath + ".tmp"
	if err := os.WriteFile(tempPath, buf, 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := os.Rename(tempPath, destPath); err != nil {
		_ = os.Remove(tempPath)
		return nil, fmt.Errorf("failed to rename temp file: %w", err)
	}

	// Detect image dimensions if content type is image
	var width, height int
	if len(contentType) >= 6 && contentType[:6] == "image/" {
		config, _, err := image.DecodeConfig(bytes.NewReader(buf))
		if err == nil {
			width = config.Width
			height = config.Height
		}
		// Ignore errors - not all images are decodable
	}

	return &UploadResult{
		Key:         key,
		URL:         s.URL(key),
		Filename:    filename,
		Size:        int64(len(buf)),
		ContentType: contentType,
		Width:       width,
		Height:      height,
	}, nil
}

// Delete removes a file by its storage key.
// Returns nil if the file doesn't exist (idempotent).
func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	path := filepath.Join(s.baseDir, key)
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	// Clean up empty parent directory
	parent := filepath.Dir(path)
	_ = os.Remove(parent) // Ignore error - dir might not be empty

	return nil
}

// URL returns the serving URL for a stored file.
func (s *LocalStorage) URL(key string) string {
	return s.baseURL + "/" + key
}

// Serve opens a file for serving.
func (s *LocalStorage) Serve(key string) (io.ReadSeekCloser, os.FileInfo, error) {
	path := filepath.Join(s.baseDir, key)
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	return file, info, nil
}

// PutInDir stores a file in a specific directory subdirectory.
// The returned key includes the directory prefix (e.g., "gallery/ab/abc123.jpg").
func (s *LocalStorage) PutInDir(dir, filename string, data io.Reader, size int64) (*UploadResult, error) {
	// Create a temporary LocalStorage with the directory as a subdirectory of baseDir
	dirStorage := &LocalStorage{
		baseDir:      filepath.Join(s.baseDir, dir),
		baseURL:      s.baseURL,
		maxSize:      s.maxSize,
		allowedTypes: s.allowedTypes,
		serveAuth:    s.serveAuth,
	}

	// Use the existing Put logic
	result, err := dirStorage.Put(context.Background(), filename, data, size)
	if err != nil {
		return nil, err
	}

	// Prepend the directory to the key
	result.Key = filepath.Join(dir, result.Key)
	result.URL = s.URL(result.Key)

	return result, nil
}
