package core

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Minimal valid 1x1 PNG (shared with storage_test.go)
var testMinPNG = []byte{
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

// setupTestApp creates an app with storage for testing
func setupTestApp(t *testing.T, opts ...StorageOption) (*App, *httptest.Server) {
	t.Helper()
	dir := t.TempDir()

	app := New()
	app.DisableCSRF() // Simplify most tests
	storage := NewLocalStorage(dir, opts...)
	app.SetStorage(storage)

	server := httptest.NewServer(app.Handler())
	t.Cleanup(server.Close)

	return app, server
}

// uploadFile creates a multipart form and POSTs to the upload endpoint
func uploadFile(t *testing.T, serverURL, fieldName, filename string, content []byte) *http.Response {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", serverURL+"/__gux_api/upload", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	return resp
}

func TestUpload_SingleFile(t *testing.T) {
	_, server := setupTestApp(t)

	resp := uploadFile(t, server.URL, "file", "test.txt", []byte("hello world"))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var results []UploadResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Filename != "test.txt" {
		t.Errorf("expected filename test.txt, got %s", result.Filename)
	}
	if result.Size != 11 {
		t.Errorf("expected size 11, got %d", result.Size)
	}
	if result.ContentType != "text/plain; charset=utf-8" {
		t.Errorf("expected text/plain, got %s", result.ContentType)
	}
	if result.Key == "" {
		t.Error("expected non-empty key")
	}
	if result.URL == "" {
		t.Error("expected non-empty URL")
	}
}

func TestUpload_MultipleFiles(t *testing.T) {
	_, server := setupTestApp(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// File 1
	part1, _ := writer.CreateFormFile("file1", "first.txt")
	part1.Write([]byte("first"))

	// File 2
	part2, _ := writer.CreateFormFile("file2", "second.txt")
	part2.Write([]byte("second"))

	writer.Close()

	req, _ := http.NewRequest("POST", server.URL+"/__gux_api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
	}

	var results []UploadResult
	json.NewDecoder(resp.Body).Decode(&results)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Check both filenames exist (order is non-deterministic due to map iteration)
	filenames := map[string]bool{}
	for _, r := range results {
		filenames[r.Filename] = true
	}
	if !filenames["first.txt"] {
		t.Error("missing first.txt in results")
	}
	if !filenames["second.txt"] {
		t.Error("missing second.txt in results")
	}
}

func TestUpload_ImageDimensions(t *testing.T) {
	_, server := setupTestApp(t)

	resp := uploadFile(t, server.URL, "file", "test.png", testMinPNG)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var results []UploadResult
	json.NewDecoder(resp.Body).Decode(&results)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Width != 1 {
		t.Errorf("expected width 1, got %d", result.Width)
	}
	if result.Height != 1 {
		t.Errorf("expected height 1, got %d", result.Height)
	}
}

func TestUpload_ContentAddressed(t *testing.T) {
	_, server := setupTestApp(t)

	// Upload same content twice
	resp1 := uploadFile(t, server.URL, "file", "test1.txt", []byte("same content"))
	defer resp1.Body.Close()

	var results1 []UploadResult
	json.NewDecoder(resp1.Body).Decode(&results1)

	resp2 := uploadFile(t, server.URL, "file", "test2.txt", []byte("same content"))
	defer resp2.Body.Close()

	var results2 []UploadResult
	json.NewDecoder(resp2.Body).Decode(&results2)

	// Same content should produce same key (content-addressed)
	if results1[0].Key != results2[0].Key {
		t.Errorf("expected same key for same content, got %s and %s", results1[0].Key, results2[0].Key)
	}
}

func TestUpload_FileTooLarge(t *testing.T) {
	_, server := setupTestApp(t, WithMaxSize(10))

	// 100 bytes exceeds 10 byte limit
	largeContent := bytes.Repeat([]byte("x"), 100)
	resp := uploadFile(t, server.URL, "file", "large.txt", largeContent)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	var errResp StorageError
	json.NewDecoder(resp.Body).Decode(&errResp)

	if errResp.Code != "file_too_large" {
		t.Errorf("expected file_too_large error, got %s", errResp.Code)
	}
	if errResp.MaxSize != 10 {
		t.Errorf("expected max_size 10, got %d", errResp.MaxSize)
	}
}

func TestUpload_TypeNotAllowed(t *testing.T) {
	_, server := setupTestApp(t, WithAllowedTypes("image/*"))

	// Text file not allowed
	resp := uploadFile(t, server.URL, "file", "test.txt", []byte("hello"))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	var errResp StorageError
	json.NewDecoder(resp.Body).Decode(&errResp)

	if errResp.Code != "invalid_file_type" {
		t.Errorf("expected invalid_file_type error, got %s", errResp.Code)
	}
}

func TestUpload_NoFiles(t *testing.T) {
	_, server := setupTestApp(t)

	// Create empty multipart form (valid form, but no files)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req, _ := http.NewRequest("POST", server.URL+"/__gux_api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	var errResp map[string]string
	json.NewDecoder(resp.Body).Decode(&errResp)

	if errResp["error"] != "no_files" {
		t.Errorf("expected no_files error, got %s", errResp["error"])
	}
}

func TestServeFile(t *testing.T) {
	_, server := setupTestApp(t)

	// Upload a file first
	uploadResp := uploadFile(t, server.URL, "file", "test.txt", []byte("hello world"))
	defer uploadResp.Body.Close()

	var results []UploadResult
	json.NewDecoder(uploadResp.Body).Decode(&results)
	url := results[0].URL

	// Fetch the file
	resp, err := http.Get(server.URL + url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Check Content-Type
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		t.Errorf("expected text/plain, got %s", contentType)
	}

	// Check Cache-Control
	cacheControl := resp.Header.Get("Cache-Control")
	if cacheControl != "public, max-age=31536000, immutable" {
		t.Errorf("expected aggressive caching, got %s", cacheControl)
	}

	// Check body content
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "hello world" {
		t.Errorf("expected 'hello world', got %s", body)
	}
}

func TestServeFile_InlineDisposition(t *testing.T) {
	_, server := setupTestApp(t)

	// Upload PNG
	uploadResp := uploadFile(t, server.URL, "file", "test.png", testMinPNG)
	defer uploadResp.Body.Close()

	var results []UploadResult
	json.NewDecoder(uploadResp.Body).Decode(&results)
	url := results[0].URL

	// Fetch the file
	resp, err := http.Get(server.URL + url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	disposition := resp.Header.Get("Content-Disposition")
	if disposition != "inline" {
		t.Errorf("expected inline disposition for image, got %s", disposition)
	}
}

func TestServeFile_AttachmentDisposition(t *testing.T) {
	_, server := setupTestApp(t)

	// Upload text file
	uploadResp := uploadFile(t, server.URL, "file", "test.txt", []byte("hello"))
	defer uploadResp.Body.Close()

	var results []UploadResult
	json.NewDecoder(uploadResp.Body).Decode(&results)
	url := results[0].URL

	// Fetch the file
	resp, err := http.Get(server.URL + url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	disposition := resp.Header.Get("Content-Disposition")
	if disposition != "attachment" {
		t.Errorf("expected attachment disposition for text, got %s", disposition)
	}
}

func TestServeFile_NotFound(t *testing.T) {
	_, server := setupTestApp(t)

	resp, err := http.Get(server.URL + "/__gux_api/files/nonexistent.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestServeFile_PathTraversal(t *testing.T) {
	_, server := setupTestApp(t)

	// Test with .. in the key itself (not in URL path)
	// The ServeMux normalizes paths, so we test the actual key check
	resp, err := http.Get(server.URL + "/__gux_api/files/../../etc/passwd")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// ServeMux normalizes the path, so this won't match our handler pattern
	// and will return 404. The path traversal check works for keys containing ".."
	// when the pattern matches (e.g., "aa/../bb.txt")
	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 404 or 400 for path traversal, got %d", resp.StatusCode)
	}

	// Test a key that would pass ServeMux but has .. in it
	resp2, err := http.Get(server.URL + "/__gux_api/files/aa/..%2F..%2Fetc%2Fpasswd")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	// This should be caught by our path traversal check
	if resp2.StatusCode != http.StatusBadRequest && resp2.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 400 for encoded path traversal, got %d", resp2.StatusCode)
	}
}

func TestUpload_WithAuth(t *testing.T) {
	dir := t.TempDir()
	app := New()
	app.DisableCSRF()
	storage := NewLocalStorage(dir)
	app.SetStorage(storage)

	// Enable auth
	sessionStore := NewMemorySessionStore()
	app.EnableAuth(AuthConfig{
		SessionStore: sessionStore,
	})

	server := httptest.NewServer(app.Handler())
	defer server.Close()

	// Upload without session - should fail
	resp := uploadFile(t, server.URL, "file", "test.txt", []byte("hello"))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth, got %d", resp.StatusCode)
	}

	// Create session
	user := &SessionUser{ID: "1", Email: "test@example.com"}
	sessionID := "test-session"
	sessionStore.Set(sessionID, user, 24*time.Hour)

	// Upload with session - should succeed
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("hello"))
	writer.Close()

	req, _ := http.NewRequest("POST", server.URL+"/__gux_api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{
		Name:  DefaultSessionCookieName,
		Value: sessionID,
	})

	client := &http.Client{}
	resp2, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp2.Body)
		t.Fatalf("expected 200 with auth, got %d: %s", resp2.StatusCode, bodyBytes)
	}
}

func TestUpload_CSRFProtection(t *testing.T) {
	dir := t.TempDir()
	app := New()
	// CSRF enabled by default
	storage := NewLocalStorage(dir)
	app.SetStorage(storage)

	server := httptest.NewServer(app.Handler())
	defer server.Close()

	// Upload without CSRF token - should fail
	resp := uploadFile(t, server.URL, "file", "test.txt", []byte("hello"))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 without CSRF token, got %d", resp.StatusCode)
	}

	// Get CSRF token by making a GET request first
	getResp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	getResp.Body.Close()

	// Extract CSRF cookie
	var csrfCookie *http.Cookie
	for _, cookie := range getResp.Cookies() {
		if cookie.Name == CSRFCookieName {
			csrfCookie = cookie
			break
		}
	}

	if csrfCookie == nil {
		t.Fatal("CSRF cookie not set")
	}

	// Upload with CSRF token - should succeed
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("hello"))
	writer.Close()

	req, _ := http.NewRequest("POST", server.URL+"/__gux_api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-CSRF-Token", csrfCookie.Value)
	req.AddCookie(csrfCookie)

	client := &http.Client{}
	resp2, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp2.Body)
		t.Fatalf("expected 200 with CSRF token, got %d: %s", resp2.StatusCode, bodyBytes)
	}
}
