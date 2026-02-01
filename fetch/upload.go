//go:build js && wasm

package fetch

import (
	"encoding/json"
	"errors"
	"syscall/js"
)

// UploadOptions configures a file upload request
type UploadOptions struct {
	// URL is the upload endpoint (default: "/__gux_api/upload")
	URL string

	// File is the JavaScript File object to upload (stays in JS heap)
	File js.Value

	// FieldName is the form field name for the file (default: "file")
	FieldName string

	// OnProgress is called periodically with upload progress
	OnProgress func(loaded, total int64)

	// ExtraFields are additional form fields to include
	ExtraFields map[string]string
}

// UploadResult represents the server's response to a successful upload
type UploadResult struct {
	Key         string `json:"key"`
	URL         string `json:"url"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

// Upload sends a file to the server using XMLHttpRequest with progress tracking.
// Returns a Response with status, body, and OK fields matching fetch.Response.
// File bytes remain in JavaScript memory (js.Value passed to FormData, no CopyBytesToGo).
func Upload(opts UploadOptions) (*Response, error) {
	// Apply defaults
	if opts.URL == "" {
		opts.URL = "/__gux_api/upload"
	}
	if opts.FieldName == "" {
		opts.FieldName = "file"
	}

	// Validate File
	if opts.File.IsNull() || opts.File.IsUndefined() {
		return nil, errors.New("upload: File is required")
	}

	// Create FormData and append file (keeps bytes in JS land)
	formData := js.Global().Get("FormData").New()
	formData.Call("append", opts.FieldName, opts.File)

	// Append extra fields
	for key, value := range opts.ExtraFields {
		formData.Call("append", key, value)
	}

	// Create XMLHttpRequest
	xhr := js.Global().Get("XMLHttpRequest").New()

	// Channel to synchronize async XHR with blocking Go call
	done := make(chan struct{})
	var response *Response
	var uploadErr error

	// Callback storage for cleanup
	var progressCb, loadCb, errorCb js.Func

	// Attach progress listener BEFORE xhr.open() to avoid browser bugs
	if opts.OnProgress != nil {
		upload := xhr.Get("upload")
		progressCb = js.FuncOf(func(this js.Value, args []js.Value) any {
			if len(args) > 0 {
				event := args[0]
				if event.Get("lengthComputable").Bool() {
					loaded := event.Get("loaded").Int()
					total := event.Get("total").Int()
					opts.OnProgress(int64(loaded), int64(total))
				}
			}
			return nil
		})
		upload.Call("addEventListener", "progress", progressCb)
	}

	// Register load event (success)
	loadCb = js.FuncOf(func(this js.Value, args []js.Value) any {
		status := xhr.Get("status").Int()
		statusText := xhr.Get("statusText").String()
		responseText := xhr.Get("responseText").String()

		response = &Response{
			Status:     status,
			StatusText: statusText,
			OK:         status >= 200 && status < 300,
			Body:       responseText,
			Headers:    make(map[string]string),
		}

		close(done)
		return nil
	})
	xhr.Call("addEventListener", "load", loadCb)

	// Register error event
	errorCb = js.FuncOf(func(this js.Value, args []js.Value) any {
		uploadErr = errors.New("upload network error")
		close(done)
		return nil
	})
	xhr.Call("addEventListener", "error", errorCb)

	// Open connection
	xhr.Call("open", "POST", opts.URL)

	// Add CSRF token (reuse existing getCSRFToken from fetch.go)
	csrfToken := getCSRFToken()
	if csrfToken != "" {
		xhr.Call("setRequestHeader", "X-CSRF-Token", csrfToken)
	}

	// Send the FormData
	xhr.Call("send", formData)

	// Block until upload completes
	<-done

	// Release all callbacks to prevent memory leaks
	if opts.OnProgress != nil {
		progressCb.Release()
	}
	loadCb.Release()
	errorCb.Release()

	if uploadErr != nil {
		return nil, uploadErr
	}

	return response, nil
}

// ParseUploadResult parses the JSON response body into an UploadResult
func ParseUploadResult(resp *Response) (*UploadResult, error) {
	if !resp.OK {
		return nil, errors.New("upload failed: " + resp.StatusText)
	}

	var result UploadResult
	if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
		return nil, err
	}

	return &result, nil
}
