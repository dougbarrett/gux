//go:build js && wasm

package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"syscall/js"

	"github.com/dougbarrett/gux/fetch"
)

var (
	dragPreventionOnce sync.Once
	objectURLs         = make(map[string]js.Func) // Store cleanup functions for object URLs
	objectURLsMu       sync.Mutex
)

// createFileUploadInit returns an OnMount callback that wires up all WASM interactivity
func createFileUploadInit(id string, props FileUploadProps) func(el any) {
	return func(el any) {
		// Prevent browser from opening dropped files
		dragPreventionOnce.Do(func() {
			window := js.Global()
			preventDefaultFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
				if len(args) > 0 {
					args[0].Call("preventDefault")
				}
				return nil
			})
			window.Call("addEventListener", "dragover", preventDefaultFunc)
			window.Call("addEventListener", "drop", preventDefaultFunc)
		})

		// Cast element to js.Value
		element := el.(js.Value)
		doc := js.Global().Get("document")

		// Get the file input element
		input := doc.Call("getElementById", id+"-input")
		if input.IsNull() || input.IsUndefined() {
			return
		}

		// File input change handler
		changeHandler := js.FuncOf(func(this js.Value, args []js.Value) any {
			files := input.Get("files")
			if files.Get("length").Int() > 0 {
				file := files.Index(0)
				handleFileSelected(id, props, file)
			}
			return nil
		})
		input.Call("addEventListener", "change", changeHandler)

		// Drop zone drag-and-drop handlers
		setupDragAndDrop(element, id, props)

		// If there's an existing remove button, wire it up
		removeBtn := doc.Call("getElementById", id+"-remove-btn")
		if !removeBtn.IsNull() && !removeBtn.IsUndefined() {
			removeBtnHandler := js.FuncOf(func(this js.Value, args []js.Value) any {
				// Clear the input
				input.Set("value", "")

				// Clear preview, progress, error
				clearElement(id + "-preview")
				clearElement(id + "-progress")
				clearElement(id + "-error")

				// Call OnRemove callback
				if props.OnRemove != nil {
					props.OnRemove()
				}

				return nil
			})
			removeBtn.Call("addEventListener", "click", removeBtnHandler)
		}
	}
}

// setupDragAndDrop wires up drag-and-drop event handlers on the element
func setupDragAndDrop(element js.Value, id string, props FileUploadProps) {
	// dragenter - add visual feedback
	dragEnterHandler := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := args[0]
			event.Call("preventDefault")
			element.Get("classList").Call("add", "gux-upload-dragover")
		}
		return nil
	})
	element.Call("addEventListener", "dragenter", dragEnterHandler)

	// dragover - required for drop to work
	dragOverHandler := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := args[0]
			event.Call("preventDefault")
			dataTransfer := event.Get("dataTransfer")
			if !dataTransfer.IsNull() && !dataTransfer.IsUndefined() {
				dataTransfer.Set("dropEffect", "copy")
			}
		}
		return nil
	})
	element.Call("addEventListener", "dragover", dragOverHandler)

	// dragleave - remove visual feedback
	dragLeaveHandler := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := args[0]
			relatedTarget := event.Get("relatedTarget")
			// Only remove class if leaving the element entirely
			if relatedTarget.IsNull() || relatedTarget.IsUndefined() ||
				!element.Call("contains", relatedTarget).Bool() {
				element.Get("classList").Call("remove", "gux-upload-dragover")
			}
		}
		return nil
	})
	element.Call("addEventListener", "dragleave", dragLeaveHandler)

	// drop - handle the dropped file
	dropHandler := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := args[0]
			event.Call("preventDefault")
			element.Get("classList").Call("remove", "gux-upload-dragover")

			dataTransfer := event.Get("dataTransfer")
			if !dataTransfer.IsNull() && !dataTransfer.IsUndefined() {
				files := dataTransfer.Get("files")
				if files.Get("length").Int() > 0 {
					file := files.Index(0)
					handleFileSelected(id, props, file)
				}
			}
		}
		return nil
	})
	element.Call("addEventListener", "drop", dropHandler)
}

// handleFileSelected processes a selected/dropped file
func handleFileSelected(id string, props FileUploadProps, file js.Value) {
	// Clear previous state
	clearElement(id + "-error")
	clearElement(id + "-preview")
	clearElement(id + "-progress")

	// Get file metadata
	fileName := file.Get("name").String()
	fileType := file.Get("type").String()
	fileSize := int64(file.Get("size").Int())

	// Validate type if Accept is specified
	if len(props.Accept) > 0 {
		valid := false
		for _, pattern := range props.Accept {
			if matchMimePattern(fileType, pattern) {
				valid = true
				break
			}
		}
		if !valid {
			errMsg := fmt.Sprintf("File type '%s' is not allowed. Accepted types: %s",
				fileType, strings.Join(props.Accept, ", "))
			showError(id, errMsg)
			if props.OnError != nil {
				props.OnError(errMsg)
			}
			return
		}
	}

	// Validate size if MaxSize is specified
	if props.MaxSize > 0 && fileSize > props.MaxSize {
		errMsg := fmt.Sprintf("File size %s exceeds maximum allowed size %s",
			formatFileSize(fileSize), formatFileSize(props.MaxSize))
		showError(id, errMsg)
		if props.OnError != nil {
			props.OnError(errMsg)
		}
		return
	}

	// Show image preview for image files
	if strings.HasPrefix(fileType, "image/") {
		showImagePreview(id, file)
	}

	// Upload the file (in goroutine to not block)
	go uploadFile(id, props, file, fileName)
}

// showImagePreview creates a preview for image files using URL.createObjectURL
func showImagePreview(id string, file js.Value) {
	doc := js.Global().Get("document")
	previewContainer := doc.Call("getElementById", id+"-preview")
	if previewContainer.IsNull() || previewContainer.IsUndefined() {
		return
	}

	// Create object URL
	url := js.Global().Get("URL")
	objURL := url.Call("createObjectURL", file)
	objURLString := objURL.String()

	// Create img element
	img := doc.Call("createElement", "img")
	img.Set("src", objURLString)
	img.Set("className", "max-w-48 max-h-48 rounded object-cover")
	img.Set("alt", "File preview")

	// Clear and append
	previewContainer.Set("innerHTML", "")
	previewContainer.Call("appendChild", img)

	// Store cleanup function
	objectURLsMu.Lock()
	defer objectURLsMu.Unlock()

	// Revoke any previous URL for this ID
	if oldCleanup, exists := objectURLs[id]; exists {
		oldCleanup.Invoke()
	}

	// Store new cleanup function
	objectURLs[id] = js.FuncOf(func(this js.Value, args []js.Value) any {
		url.Call("revokeObjectURL", objURLString)
		return nil
	})
}

// uploadFile uploads the file to the server with progress tracking
func uploadFile(id string, props FileUploadProps, file js.Value, fileName string) {
	// Update status for screen readers
	updateStatus(id, "Uploading: 0%")

	// Show progress bar
	showProgressBar(id, 0)

	// Call fetch.Upload
	resp, err := fetch.Upload(fetch.UploadOptions{
		URL:       props.UploadURL,
		File:      file,
		FieldName: "file",
		OnProgress: func(loaded, total int64) {
			if total > 0 {
				percent := float64(loaded) / float64(total) * 100
				showProgressBar(id, int(percent))
				updateStatus(id, fmt.Sprintf("Uploading: %d%%", int(percent)))
			}
		},
	})

	if err != nil {
		showError(id, "Upload failed: "+err.Error())
		if props.OnError != nil {
			props.OnError(err.Error())
		}
		return
	}

	if !resp.OK {
		errMsg := fmt.Sprintf("Upload failed: %s", resp.StatusText)
		showError(id, errMsg)
		if props.OnError != nil {
			props.OnError(errMsg)
		}
		return
	}

	// Parse response
	var result UploadResult
	if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
		showError(id, "Failed to parse response: "+err.Error())
		if props.OnError != nil {
			props.OnError(err.Error())
		}
		return
	}

	// Success! Clear progress, show success state
	clearElement(id + "-progress")
	updateStatus(id, "Upload complete")
	showSuccess(id, result)

	// Call callback
	if props.OnUploadComplete != nil {
		props.OnUploadComplete(result)
	}
}

// showProgressBar displays/updates the progress bar
func showProgressBar(id string, percent int) {
	doc := js.Global().Get("document")
	container := doc.Call("getElementById", id+"-progress")
	if container.IsNull() || container.IsUndefined() {
		return
	}

	// Create/update progress bar HTML
	html := fmt.Sprintf(`
		<div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2.5">
			<div class="bg-blue-600 h-2.5 rounded-full transition-all duration-300"
			     style="width: %d%%"
			     role="progressbar"
			     aria-valuenow="%d"
			     aria-valuemin="0"
			     aria-valuemax="100"></div>
		</div>
	`, percent, percent)

	container.Set("innerHTML", html)
}

// showSuccess displays the success state with filename and remove button
func showSuccess(id string, result UploadResult) {
	doc := js.Global().Get("document")
	previewContainer := doc.Call("getElementById", id+"-preview")
	if previewContainer.IsNull() || previewContainer.IsUndefined() {
		return
	}

	// Build success HTML
	var html string
	if strings.HasPrefix(result.ContentType, "image/") {
		html = fmt.Sprintf(`
			<div class="border-2 border-solid rounded-lg p-4 bg-white dark:bg-gray-800 border-gray-300 dark:border-gray-600">
				<img src="%s" class="max-w-48 max-h-48 rounded object-cover mb-2" alt="Uploaded file" />
				<p class="text-sm text-green-600 dark:text-green-400 mb-2">✓ Upload complete</p>
				<button type="button" id="%s-remove-success" class="text-sm text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300">
					Remove file
				</button>
			</div>
		`, result.URL, id)
	} else {
		html = fmt.Sprintf(`
			<div class="border-2 border-solid rounded-lg p-4 bg-white dark:bg-gray-800 border-gray-300 dark:border-gray-600">
				<p class="text-sm text-gray-700 dark:text-gray-300 mb-2">File: %s</p>
				<p class="text-sm text-green-600 dark:text-green-400 mb-2">✓ Upload complete</p>
				<button type="button" id="%s-remove-success" class="text-sm text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300">
					Remove file
				</button>
			</div>
		`, result.Filename, id)
	}

	previewContainer.Set("innerHTML", html)

	// Wire up the remove button
	removeBtn := doc.Call("getElementById", id+"-remove-success")
	if !removeBtn.IsNull() && !removeBtn.IsUndefined() {
		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			// Get input and clear it
			input := doc.Call("getElementById", id+"-input")
			if !input.IsNull() && !input.IsUndefined() {
				input.Set("value", "")
			}

			// Clear all containers
			clearElement(id + "-preview")
			clearElement(id + "-progress")
			clearElement(id + "-error")
			updateStatus(id, "File removed")

			// Call OnRemove callback
			// Note: We don't have access to props here in this closure,
			// but the user can track state via OnUploadComplete and implement
			// their own remove logic
			return nil
		})
		removeBtn.Call("addEventListener", "click", handler)
	}
}

// showError displays an error message
func showError(id string, message string) {
	doc := js.Global().Get("document")
	container := doc.Call("getElementById", id+"-error")
	if container.IsNull() || container.IsUndefined() {
		return
	}

	html := fmt.Sprintf(`
		<div class="text-sm text-red-600 dark:text-red-400 p-3 bg-red-50 dark:bg-red-900/20 rounded border border-red-200 dark:border-red-800">
			%s
		</div>
	`, message)

	container.Set("innerHTML", html)
	updateStatus(id, "Error: "+message)
}

// updateStatus updates the screen reader status announcement
func updateStatus(id string, message string) {
	doc := js.Global().Get("document")
	container := doc.Call("getElementById", id+"-status")
	if container.IsNull() || container.IsUndefined() {
		return
	}
	container.Set("textContent", message)
}

// clearElement clears an element's content
func clearElement(elementID string) {
	doc := js.Global().Get("document")
	el := doc.Call("getElementById", elementID)
	if !el.IsNull() && !el.IsUndefined() {
		el.Set("innerHTML", "")
	}
}

// matchMimePattern checks if a MIME type matches a pattern (e.g., "image/*")
func matchMimePattern(mimeType, pattern string) bool {
	if pattern == "*/*" || pattern == "*" {
		return true
	}

	if strings.HasSuffix(pattern, "/*") {
		// Wildcard pattern like "image/*"
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(mimeType, prefix+"/")
	}

	// Exact match
	return mimeType == pattern
}
