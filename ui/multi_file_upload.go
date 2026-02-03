package ui

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/dougbarrett/gux/core"
)

// MultiFileUploadProps configures a MultiFileUpload component.
type MultiFileUploadProps struct {
	// Accept is a list of allowed MIME patterns (e.g., "image/*", "application/pdf").
	// Empty means any file type is accepted.
	Accept []string

	// MaxSize is the maximum file size in bytes. 0 means no limit.
	MaxSize int64

	// UploadURL is the endpoint URL for file upload (default: "/__gux_api/upload")
	UploadURL string

	// Placeholder is the text shown in the add zone (default: "Click to add or drag and drop")
	Placeholder string

	// Class contains additional CSS classes for the outer container
	Class string

	// Disabled disables all interaction
	Disabled bool

	// Values is the current list of file URLs (for displaying existing files)
	Values []string

	// OnUploadComplete is called after each successful upload with server response (WASM only)
	OnUploadComplete func(result UploadResult)

	// OnError is called on validation or upload errors (WASM only)
	OnError func(err string)

	// OnRemove is called when user removes a file at the given index (WASM only)
	OnRemove func(index int)
}

var multiFileUploadIDCounter atomic.Int64

// MultiFileUpload creates a multi-file upload component with drag-and-drop support.
//
// Basic usage:
//
//	ui.MultiFileUpload(ui.MultiFileUploadProps{
//	    Accept: []string{"image/*"},
//	    MaxSize: 5 * 1024 * 1024, // 5MB
//	    Values: []string{"/files/abc.jpg", "/files/def.png"},
//	    OnUploadComplete: func(result ui.UploadResult) {
//	        println("Uploaded:", result.URL)
//	    },
//	    OnRemove: func(index int) {
//	        println("Removed file at index:", index)
//	    },
//	})
//
// The component renders static HTML in SSR and gains full interactivity
// (drag-drop, validation, preview, progress) after WASM hydration.
func MultiFileUpload(props MultiFileUploadProps) core.Node {
	if props.Placeholder == "" {
		props.Placeholder = "Click to add or drag and drop"
	}
	if props.UploadURL == "" {
		props.UploadURL = "/__gux_api/upload"
	}

	id := fmt.Sprintf("gux-multi-upload-%d", multiFileUploadIDCounter.Add(1))

	// Build outer container
	outerClass := MergeClasses(
		"gux-multi-file-upload",
		"relative",
		props.Class,
	)

	var children []core.Node

	// Container: border, rounded corners
	containerClass := "border border-gray-300 dark:border-gray-600 rounded-lg overflow-hidden"

	var containerChildren []core.Node

	// File list: for each URL in Values, render a flex row
	if len(props.Values) > 0 {
		for i, url := range props.Values {
			containerChildren = append(containerChildren, buildFileRow(id, i, url))
		}
	}

	// Add zone at bottom
	addZoneClass := MergeClasses(
		"flex items-center justify-center gap-2 px-3 py-4",
		"border-t border-dashed border-gray-300 dark:border-gray-600",
		"cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800",
		"text-sm text-gray-500",
	)
	if props.Disabled {
		addZoneClass = MergeClasses(
			"flex items-center justify-center gap-2 px-3 py-4",
			"border-t border-dashed border-gray-200 dark:border-gray-700",
			"cursor-not-allowed opacity-60",
			"text-sm text-gray-400",
		)
	}

	// Hidden file input (accessible but visually hidden)
	inputExtra := make(map[string]string)
	inputExtra["multiple"] = "multiple"
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

	addZoneChildren := []core.Node{
		core.Input(inputAttrs),
		buildPlusIcon(),
		core.Span(core.Class(""),
			core.Text(props.Placeholder),
		),
	}

	// Accepted types hint
	if len(props.Accept) > 0 {
		addZoneChildren = append(addZoneChildren,
			core.Span(core.Class("ml-2 text-xs"),
				core.Text("("+strings.Join(props.Accept, ", ")+")"),
			),
		)
	}

	// Max size hint
	if props.MaxSize > 0 {
		addZoneChildren = append(addZoneChildren,
			core.Span(core.Class("ml-2 text-xs"),
				core.Text("Max: "+formatFileSize(props.MaxSize)),
			),
		)
	}

	addZone := core.Label(
		core.Attrs{
			Class: addZoneClass,
			Extra: map[string]string{
				"for": id + "-input",
			},
		},
		addZoneChildren...,
	)

	containerChildren = append(containerChildren, addZone)

	children = append(children, core.Div(core.Class(containerClass), containerChildren...))

	// Progress container (populated by WASM during upload)
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
		OnMount: createMultiFileUploadInit(id, props),
		Extra: map[string]string{
			"role":       "group",
			"aria-label": "Multi-file upload",
			"tabindex":   "0",
		},
	}

	return core.Div(attrs, children...)
}

// buildFileRow creates a file row for a single file in the list
func buildFileRow(id string, index int, url string) core.Node {
	rowClass := "flex items-center gap-3 px-3 py-2 border-b border-gray-100 dark:border-gray-700"

	var preview core.Node
	if isImageURL(url) {
		// Image thumbnail
		preview = core.Img(core.Attrs{
			Src:   url,
			Class: "h-12 w-12 rounded object-cover",
			Alt:   "File thumbnail",
		})
	} else {
		// File icon (Unicode document character)
		preview = core.Div(core.Class("h-12 w-12 flex items-center justify-center text-gray-400 text-2xl"),
			core.Text("📄"),
		)
	}

	// Extract filename from URL path
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]

	return core.Div(core.Class(rowClass),
		preview,
		core.Div(core.Class("flex-1 text-sm text-gray-700 dark:text-gray-300 truncate"),
			core.Text(filename),
		),
		core.Button(core.Attrs{
			Type:  "button",
			Class: "text-gray-400 hover:text-red-500 cursor-pointer",
			Extra: map[string]string{
				"data-mfu-remove": fmt.Sprintf("%d", index),
			},
		},
			core.Text("×"),
		),
	)
}

// buildPlusIcon creates a simple plus icon
func buildPlusIcon() core.Node {
	return core.RawHTML(`<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
	</svg>`)
}
