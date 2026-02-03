//go:build js && wasm

package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/dougbarrett/gux/fetch"
)

// createMultiFileUploadInit returns an OnMount callback that wires up all WASM interactivity
func createMultiFileUploadInit(id string, props MultiFileUploadProps) func(el any) {
	return func(el any) {
		// Prevent browser from opening dropped files (use shared global handler)
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

		// File input change handler (supports multiple file selection)
		changeHandler := js.FuncOf(func(this js.Value, args []js.Value) any {
			files := input.Get("files")
			length := files.Get("length").Int()
			if length > 0 {
				// Upload each file individually
				for i := 0; i < length; i++ {
					file := files.Index(i)
					go handleMultiFileSelected(id, props, file, i, length)
				}
			}
			return nil
		})
		input.Call("addEventListener", "change", changeHandler)

		// Drop zone drag-and-drop handlers
		setupMultiFileDragAndDrop(element, id, props)

		// Wire up all existing remove buttons
		setupRemoveButtons(id, props)
	}
}

// setupMultiFileDragAndDrop wires up drag-and-drop event handlers on the element
func setupMultiFileDragAndDrop(element js.Value, id string, props MultiFileUploadProps) {
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

	// drop - handle the dropped files
	dropHandler := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := args[0]
			event.Call("preventDefault")
			element.Get("classList").Call("remove", "gux-upload-dragover")

			dataTransfer := event.Get("dataTransfer")
			if !dataTransfer.IsNull() && !dataTransfer.IsUndefined() {
				files := dataTransfer.Get("files")
				length := files.Get("length").Int()
				if length > 0 {
					// Upload each file individually
					for i := 0; i < length; i++ {
						file := files.Index(i)
						go handleMultiFileSelected(id, props, file, i, length)
					}
				}
			}
		}
		return nil
	})
	element.Call("addEventListener", "drop", dropHandler)
}

// setupRemoveButtons attaches click handlers to all remove buttons
func setupRemoveButtons(id string, props MultiFileUploadProps) {
	doc := js.Global().Get("document")
	// Find all buttons with data-mfu-remove attribute
	buttons := doc.Call("querySelectorAll", fmt.Sprintf("[data-mfu-remove]"))
	length := buttons.Get("length").Int()
	for i := 0; i < length; i++ {
		btn := buttons.Index(i)
		indexStr := btn.Call("getAttribute", "data-mfu-remove").String()
		index := 0
		fmt.Sscanf(indexStr, "%d", &index)

		handler := js.FuncOf(func(this js.Value, args []js.Value) any {
			if props.OnRemove != nil {
				props.OnRemove(index)
			}
			return nil
		})
		btn.Call("addEventListener", "click", handler)
	}
}

// handleMultiFileSelected processes a selected/dropped file
func handleMultiFileSelected(id string, props MultiFileUploadProps, file js.Value, fileIndex, totalFiles int) {
	// Clear previous errors
	clearElement(id + "-error")

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

	// Show per-file progress indicator
	progressID := fmt.Sprintf("%s-progress-%d", id, fileIndex)
	showFileProgress(id, progressID, fileName, 0)

	// Update status for screen readers
	updateStatus(id, fmt.Sprintf("Uploading %s: 0%%", fileName))

	// Upload the file
	resp, err := fetch.Upload(fetch.UploadOptions{
		URL:       props.UploadURL,
		File:      file,
		FieldName: "file",
		OnProgress: func(loaded, total int64) {
			if total > 0 {
				percent := float64(loaded) / float64(total) * 100
				updateFileProgress(progressID, int(percent))
				updateStatus(id, fmt.Sprintf("Uploading %s: %d%%", fileName, int(percent)))
			}
		},
	})

	if err != nil {
		removeFileProgress(progressID)
		showError(id, "Upload failed: "+err.Error())
		if props.OnError != nil {
			props.OnError(err.Error())
		}
		return
	}

	if !resp.OK {
		removeFileProgress(progressID)
		errMsg := fmt.Sprintf("Upload failed: %s", resp.StatusText)
		showError(id, errMsg)
		if props.OnError != nil {
			props.OnError(errMsg)
		}
		return
	}

	// Parse response (expect array of results)
	var results []UploadResult
	if err := json.Unmarshal([]byte(resp.Body), &results); err != nil {
		// Try parsing as single result (backwards compatibility)
		var singleResult UploadResult
		if err := json.Unmarshal([]byte(resp.Body), &singleResult); err != nil {
			removeFileProgress(progressID)
			showError(id, "Failed to parse response: "+err.Error())
			if props.OnError != nil {
				props.OnError(err.Error())
			}
			return
		}
		results = []UploadResult{singleResult}
	}

	// Success! Remove progress indicator
	removeFileProgress(progressID)
	updateStatus(id, fmt.Sprintf("Upload complete: %s", fileName))

	// Call callback with first result (single file upload)
	if len(results) > 0 && props.OnUploadComplete != nil {
		props.OnUploadComplete(results[0])
	}
}

// showFileProgress displays a progress indicator for a specific file
func showFileProgress(containerID, progressID, fileName string, percent int) {
	doc := js.Global().Get("document")
	container := doc.Call("getElementById", containerID+"-progress")
	if container.IsNull() || container.IsUndefined() {
		return
	}

	// Create progress row
	html := fmt.Sprintf(`
		<div id="%s" class="flex items-center gap-2 mb-2">
			<div class="flex-1">
				<div class="text-xs text-gray-600 dark:text-gray-400 mb-1 truncate">%s</div>
				<div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
					<div class="bg-blue-600 h-2 rounded-full transition-all duration-300"
					     style="width: %d%%"
					     role="progressbar"
					     aria-valuenow="%d"
					     aria-valuemin="0"
					     aria-valuemax="100"></div>
				</div>
			</div>
		</div>
	`, progressID, fileName, percent, percent)

	// Append to container
	div := doc.Call("createElement", "div")
	div.Set("innerHTML", html)
	container.Call("appendChild", div.Get("firstElementChild"))
}

// updateFileProgress updates the progress bar for a specific file
func updateFileProgress(progressID string, percent int) {
	doc := js.Global().Get("document")
	progressEl := doc.Call("getElementById", progressID)
	if progressEl.IsNull() || progressEl.IsUndefined() {
		return
	}

	progressBar := progressEl.Call("querySelector", "[role='progressbar']")
	if !progressBar.IsNull() && !progressBar.IsUndefined() {
		progressBar.Get("style").Set("width", fmt.Sprintf("%d%%", percent))
		progressBar.Set("ariaValueNow", percent)
	}
}

// removeFileProgress removes a progress indicator
func removeFileProgress(progressID string) {
	doc := js.Global().Get("document")
	progressEl := doc.Call("getElementById", progressID)
	if !progressEl.IsNull() && !progressEl.IsUndefined() {
		progressEl.Call("remove")
	}
}
