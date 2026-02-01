package ui

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/dougbarrett/gux/core"
)

// UploadResult represents the server's response to a successful upload.
// This is a re-declaration of fetch.UploadResult without js dependency.
type UploadResult struct {
	Key         string `json:"key"`
	URL         string `json:"url"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

// FileUploadProps configures a FileUpload component.
type FileUploadProps struct {
	// Accept is a list of allowed MIME patterns (e.g., "image/*", "application/pdf").
	// Empty means any file type is accepted.
	Accept []string

	// MaxSize is the maximum file size in bytes. 0 means no limit.
	MaxSize int64

	// UploadURL is the endpoint URL for file upload (default: "/__gux_api/upload")
	UploadURL string

	// Placeholder is the text shown in the empty state (default: "Click to upload or drag and drop")
	Placeholder string

	// Class contains additional CSS classes for the outer container
	Class string

	// Disabled disables all interaction
	Disabled bool

	// Value is the current file URL (for edit forms showing existing files)
	Value string

	// OnUploadComplete is called after successful upload with server response (WASM only)
	OnUploadComplete func(result UploadResult)

	// OnError is called on validation or upload errors (WASM only)
	OnError func(err string)

	// OnRemove is called when user removes a file (WASM only)
	OnRemove func()
}

var fileUploadIDCounter atomic.Int64

// FileUpload creates a file upload component with drag-and-drop support.
//
// Basic usage:
//
//	ui.FileUpload(ui.FileUploadProps{
//	    Accept: []string{"image/*"},
//	    MaxSize: 5 * 1024 * 1024, // 5MB
//	    OnUploadComplete: func(result ui.UploadResult) {
//	        println("Uploaded:", result.URL)
//	    },
//	})
//
// The component renders static HTML in SSR and gains full interactivity
// (drag-drop, validation, preview, progress) after WASM hydration.
func FileUpload(props FileUploadProps) core.Node {
	if props.Placeholder == "" {
		props.Placeholder = "Click to upload or drag and drop"
	}
	if props.UploadURL == "" {
		props.UploadURL = "/__gux_api/upload"
	}

	id := fmt.Sprintf("gux-upload-%d", fileUploadIDCounter.Add(1))

	// Build outer container
	outerClass := MergeClasses(
		"gux-file-upload",
		"relative",
		props.Class,
	)

	var children []core.Node

	// Hidden file input (accessible but visually hidden)
	inputExtra := make(map[string]string)
	if len(props.Accept) > 0 {
		inputExtra["accept"] = acceptString(props.Accept)
	}
	if props.Disabled {
		inputExtra["disabled"] = "disabled"
	}

	inputAttrs := core.Attrs{
		Type:  "file",
		ID:    id + "-input",
		Class: "sr-only", // screen-reader-only class (visually hidden)
		Extra: inputExtra,
	}

	children = append(children, core.Input(inputAttrs))

	// If there's an existing file value, show it
	if props.Value != "" {
		children = append(children, buildExistingFileDisplay(id, props.Value))
	} else {
		// Label wrapping the visible drop zone
		dropZoneClass := MergeClasses(
			"border-2 border-dashed rounded-lg p-8 transition-colors",
			"border-gray-300 dark:border-gray-600",
			"bg-gray-50 dark:bg-gray-800",
			"hover:border-blue-400 dark:hover:border-blue-500",
			"flex flex-col items-center justify-center",
			"cursor-pointer",
		)
		if props.Disabled {
			dropZoneClass = MergeClasses(
				"border-2 border-dashed rounded-lg p-8",
				"border-gray-200 dark:border-gray-700",
				"bg-gray-100 dark:bg-gray-900",
				"flex flex-col items-center justify-center",
				"cursor-not-allowed opacity-60",
			)
		}

		labelChildren := []core.Node{
			// Upload icon
			core.Div(core.Class("mb-4 text-gray-400 dark:text-gray-500"),
				buildUploadIcon(),
			),

			// Placeholder text
			core.P(core.Class("mb-2 text-sm text-gray-700 dark:text-gray-300 font-medium"),
				core.Text(props.Placeholder),
			),
		}

		// Accepted types hint
		if len(props.Accept) > 0 {
			labelChildren = append(labelChildren,
				core.P(core.Class("text-xs text-gray-500 dark:text-gray-400"),
					core.Text("Accepted types: "+strings.Join(props.Accept, ", ")),
				),
			)
		}

		// Max size hint
		if props.MaxSize > 0 {
			labelChildren = append(labelChildren,
				core.P(core.Class("text-xs text-gray-500 dark:text-gray-400 mt-1"),
					core.Text("Max size: "+formatFileSize(props.MaxSize)),
				),
			)
		}

		children = append(children,
			core.Label(
				core.Attrs{
					Class: dropZoneClass,
					Extra: map[string]string{
						"for": id + "-input",
					},
				},
				labelChildren...,
			),
		)
	}

	// Preview container (populated by WASM)
	children = append(children,
		core.Div(core.Attrs{
			ID:    id + "-preview",
			Class: "mt-4",
		}),
	)

	// Progress container (populated by WASM)
	children = append(children,
		core.Div(core.Attrs{
			ID:    id + "-progress",
			Class: "mt-4",
			Extra: map[string]string{
				"aria-live":   "polite",
				"aria-atomic": "true",
			},
		}),
	)

	// Error container (populated by WASM)
	children = append(children,
		core.Div(core.Attrs{
			ID:    id + "-error",
			Class: "mt-4",
			Extra: map[string]string{
				"aria-live":   "assertive",
				"aria-atomic": "true",
			},
		}),
	)

	// Screen reader status announcements (visually hidden)
	children = append(children,
		core.Div(core.Attrs{
			ID:    id + "-status",
			Class: "sr-only",
			Extra: map[string]string{
				"aria-live":   "polite",
				"aria-atomic": "true",
			},
		}),
	)

	// Wire up WASM interactivity via OnMount
	attrs := core.Attrs{
		ID:      id,
		Class:   outerClass,
		OnMount: createFileUploadInit(id, props),
		Extra: map[string]string{
			"role":       "group",
			"aria-label": "File upload",
			"tabindex":   "0",
		},
	}

	return core.Div(attrs, children...)
}

// buildExistingFileDisplay creates the display for an existing uploaded file
func buildExistingFileDisplay(id, url string) core.Node {
	containerClass := "border-2 border-solid rounded-lg p-4 bg-white dark:bg-gray-800 border-gray-300 dark:border-gray-600"

	var preview core.Node
	if isImageURL(url) {
		preview = core.Img(core.Attrs{
			Src:   url,
			Class: "max-w-48 max-h-48 rounded object-cover mb-2",
			Alt:   "Uploaded file preview",
		})
	} else {
		// Extract filename from URL
		parts := strings.Split(url, "/")
		filename := parts[len(parts)-1]
		preview = core.P(core.Class("text-sm text-gray-700 dark:text-gray-300 mb-2"),
			core.Text("File: "+filename),
		)
	}

	return core.Div(core.Class(containerClass),
		preview,
		core.Button(core.Attrs{
			Type:  "button",
			Class: "text-sm text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300",
			ID:    id + "-remove-btn",
		},
			core.Text("Remove file"),
		),
	)
}

// buildUploadIcon creates an SVG upload icon
func buildUploadIcon() core.Node {
	return core.RawHTML(`<svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"></path>
	</svg>`)
}

// formatFileSize converts bytes to human-readable format
func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	if bytes >= GB {
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	} else if bytes >= MB {
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	} else if bytes >= KB {
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	}
	return fmt.Sprintf("%d B", bytes)
}

// isImageURL checks if a URL points to a common image format
func isImageURL(url string) bool {
	lower := strings.ToLower(url)
	return strings.HasSuffix(lower, ".jpg") ||
		strings.HasSuffix(lower, ".jpeg") ||
		strings.HasSuffix(lower, ".png") ||
		strings.HasSuffix(lower, ".gif") ||
		strings.HasSuffix(lower, ".webp") ||
		strings.HasSuffix(lower, ".svg")
}

// acceptString joins accept patterns with commas for the input accept attribute
func acceptString(accept []string) string {
	return strings.Join(accept, ",")
}
