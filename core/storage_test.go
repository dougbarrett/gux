package core

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Minimal valid 1x1 PNG
var minPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
	0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
	0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
	0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
	0x00, 0x00, 0x02, 0x00, 0x01, 0xE2, 0x21, 0xBC,
	0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
	0x44, 0xAE, 0x42, 0x60, 0x82,
}

func TestNewLocalStorage(t *testing.T) {
	t.Run("constructor creates directory", func(t *testing.T) {
		dir := t.TempDir()
		storageDir := filepath.Join(dir, "storage")

		storage := NewLocalStorage(storageDir)

		if storage == nil {
			t.Fatal("NewLocalStorage returned nil")
		}

		if _, err := os.Stat(storageDir); os.IsNotExist(err) {
			t.Error("storage directory not created")
		}
	})

	t.Run("applies options", func(t *testing.T) {
		dir := t.TempDir()
		storage := NewLocalStorage(dir,
			WithMaxSize(1024),
			WithAllowedTypes("image/*"),
			WithServeAuth(),
			WithBaseURL("/custom"),
		)

		if storage.maxSize != 1024 {
			t.Errorf("maxSize = %d, want 1024", storage.maxSize)
		}
		if len(storage.allowedTypes) != 1 || storage.allowedTypes[0] != "image/*" {
			t.Errorf("allowedTypes = %v, want [image/*]", storage.allowedTypes)
		}
		if !storage.serveAuth {
			t.Error("serveAuth = false, want true")
		}
		if storage.baseURL != "/custom" {
			t.Errorf("baseURL = %s, want /custom", storage.baseURL)
		}
	})
}

func TestLocalStorage_Put_BasicFile(t *testing.T) {
	dir := t.TempDir()
	storage := NewLocalStorage(dir)

	content := []byte("hello world")
	result, err := storage.Put(context.Background(), "test.txt", bytes.NewReader(content), int64(len(content)))

	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Verify result fields
	if result.Key == "" {
		t.Error("result.Key is empty")
	}
	if result.URL == "" {
		t.Error("result.URL is empty")
	}
	if result.Filename != "test.txt" {
		t.Errorf("result.Filename = %s, want test.txt", result.Filename)
	}
	if result.Size != int64(len(content)) {
		t.Errorf("result.Size = %d, want %d", result.Size, len(content))
	}
	if result.ContentType != "text/plain; charset=utf-8" {
		t.Errorf("result.ContentType = %s, want text/plain; charset=utf-8", result.ContentType)
	}

	// Verify file on disk
	filePath := filepath.Join(dir, result.Key)
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Errorf("file content = %s, want %s", data, content)
	}
}

func TestLocalStorage_Put_ImageDimensions(t *testing.T) {
	dir := t.TempDir()
	storage := NewLocalStorage(dir)

	result, err := storage.Put(context.Background(), "test.png", bytes.NewReader(minPNG), int64(len(minPNG)))

	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	if result.Width != 1 {
		t.Errorf("result.Width = %d, want 1", result.Width)
	}
	if result.Height != 1 {
		t.Errorf("result.Height = %d, want 1", result.Height)
	}
	if !strings.HasPrefix(result.ContentType, "image/png") {
		t.Errorf("result.ContentType = %s, want image/png", result.ContentType)
	}
}

func TestLocalStorage_Put_ContentAddressed(t *testing.T) {
	dir := t.TempDir()
	storage := NewLocalStorage(dir)

	content := []byte("identical content")

	// Upload same content twice
	result1, err := storage.Put(context.Background(), "file1.txt", bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("first Put failed: %v", err)
	}

	result2, err := storage.Put(context.Background(), "file2.txt", bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("second Put failed: %v", err)
	}

	// Same content should produce same key
	if result1.Key != result2.Key {
		t.Errorf("keys differ: %s != %s", result1.Key, result2.Key)
	}

	// Original filenames should be preserved separately
	if result1.Filename != "file1.txt" {
		t.Errorf("result1.Filename = %s, want file1.txt", result1.Filename)
	}
	if result2.Filename != "file2.txt" {
		t.Errorf("result2.Filename = %s, want file2.txt", result2.Filename)
	}
}

func TestLocalStorage_Put_MaxSizeExceeded(t *testing.T) {
	dir := t.TempDir()
	storage := NewLocalStorage(dir, WithMaxSize(10))

	content := make([]byte, 100)
	_, err := storage.Put(context.Background(), "large.bin", bytes.NewReader(content), int64(len(content)))

	if err == nil {
		t.Fatal("Put should have failed with max size exceeded")
	}

	storageErr, ok := err.(*StorageError)
	if !ok {
		t.Fatalf("expected StorageError, got %T", err)
	}
	if storageErr.Code != "file_too_large" {
		t.Errorf("error code = %s, want file_too_large", storageErr.Code)
	}
	if storageErr.MaxSize != 10 {
		t.Errorf("error MaxSize = %d, want 10", storageErr.MaxSize)
	}
}

func TestLocalStorage_Put_TypeNotAllowed(t *testing.T) {
	dir := t.TempDir()
	storage := NewLocalStorage(dir, WithAllowedTypes("image/*"))

	content := []byte("plain text")
	_, err := storage.Put(context.Background(), "test.txt", bytes.NewReader(content), int64(len(content)))

	if err == nil {
		t.Fatal("Put should have failed with invalid file type")
	}

	storageErr, ok := err.(*StorageError)
	if !ok {
		t.Fatalf("expected StorageError, got %T", err)
	}
	if storageErr.Code != "invalid_file_type" {
		t.Errorf("error code = %s, want invalid_file_type", storageErr.Code)
	}
}

func TestLocalStorage_Put_TypeAllowed(t *testing.T) {
	dir := t.TempDir()
	storage := NewLocalStorage(dir, WithAllowedTypes("image/*"))

	result, err := storage.Put(context.Background(), "test.png", bytes.NewReader(minPNG), int64(len(minPNG)))

	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected upload result, got nil")
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	dir := t.TempDir()
	storage := NewLocalStorage(dir)

	content := []byte("to be deleted")
	result, err := storage.Put(context.Background(), "delete.txt", bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(dir, result.Key)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("file not found after upload")
	}

	// Delete file
	if err := storage.Delete(context.Background(), result.Key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify file removed
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("file still exists after delete")
	}
}

func TestLocalStorage_Delete_NotFound(t *testing.T) {
	dir := t.TempDir()
	storage := NewLocalStorage(dir)

	// Delete non-existent file should not error
	err := storage.Delete(context.Background(), "nonexistent/file.txt")
	if err != nil {
		t.Errorf("Delete of non-existent file returned error: %v", err)
	}
}

func TestLocalStorage_URL(t *testing.T) {
	dir := t.TempDir()
	storage := NewLocalStorage(dir)

	url := storage.URL("ab/abc123.jpg")
	expected := "/__gux_api/files/ab/abc123.jpg"
	if url != expected {
		t.Errorf("URL = %s, want %s", url, expected)
	}
}

func TestLocalStorage_URL_CustomBase(t *testing.T) {
	dir := t.TempDir()
	storage := NewLocalStorage(dir, WithBaseURL("/custom"))

	url := storage.URL("ab/abc123.jpg")
	expected := "/custom/ab/abc123.jpg"
	if url != expected {
		t.Errorf("URL = %s, want %s", url, expected)
	}
}

func TestMatchMimePattern(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		pattern  string
		want     bool
	}{
		{
			name:     "exact match",
			mimeType: "image/jpeg",
			pattern:  "image/jpeg",
			want:     true,
		},
		{
			name:     "wildcard match",
			mimeType: "image/jpeg",
			pattern:  "image/*",
			want:     true,
		},
		{
			name:     "wildcard no match",
			mimeType: "text/plain",
			pattern:  "image/*",
			want:     false,
		},
		{
			name:     "exact no match",
			mimeType: "image/png",
			pattern:  "image/jpeg",
			want:     false,
		},
		{
			name:     "different category",
			mimeType: "video/mp4",
			pattern:  "image/*",
			want:     false,
		},
		{
			name:     "application wildcard",
			mimeType: "application/pdf",
			pattern:  "application/*",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchMimePattern(tt.mimeType, tt.pattern)
			if got != tt.want {
				t.Errorf("matchMimePattern(%q, %q) = %v, want %v", tt.mimeType, tt.pattern, got, tt.want)
			}
		})
	}
}
