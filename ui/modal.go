package ui

import "github.com/dougbarrett/gux/core"

// ModalSize defines the width of a modal.
type ModalSize string

const (
	ModalSM   ModalSize = "sm"
	ModalMD   ModalSize = "md"
	ModalLG   ModalSize = "lg"
	ModalXL   ModalSize = "xl"
	ModalFull ModalSize = "full"
)

// modalSizeClasses maps sizes to their Tailwind max-width classes.
var modalSizeClasses = map[ModalSize]string{
	ModalSM:   "max-w-sm w-full mx-4",
	ModalMD:   "max-w-md w-full mx-4",
	ModalLG:   "max-w-lg w-full mx-4",
	ModalXL:   "max-w-xl w-full mx-4",
	ModalFull: "max-w-4xl w-full mx-4",
}

// ModalProps configures a Modal component.
type ModalProps struct {
	Open     bool        // Controls visibility via hidden class
	OnClose  func()      // Callback for X button and backdrop click
	Title    string      // Optional title (sets aria-labelledby when present)
	Size     ModalSize   // Width size (default: md)
	Class    string      // Additional classes for modal container
	Children []core.Node // Modal content

	// Alpine.js props
	XShow    string // x-show expression (e.g., "modalOpen")
	XOnClose string // x-on:click expression for close button (e.g., "modalOpen = false")
}

// Modal creates an accessible modal dialog.
//
// The modal is always rendered but hidden via CSS when not open.
// This ensures consistent SSR output and smooth WASM hydration.
//
// Example:
//
//	modalOpen := r.StateBool("modal", false)
//	Modal(ModalProps{
//	    Open:    modalOpen.Get(),
//	    OnClose: func() { modalOpen.Set(false) },
//	    Title:   "Confirm Action",
//	    Children: []core.Node{
//	        ModalContent(ModalContentProps{
//	            Children: []core.Node{core.Text("Are you sure?")},
//	        }),
//	        ModalFooter(ModalFooterProps{
//	            Children: []core.Node{
//	                Button(ButtonProps{
//	                    OnClick:  func() { modalOpen.Set(false) },
//	                    Children: []core.Node{core.Text("Cancel")},
//	                }),
//	            },
//	        }),
//	    },
//	})
func Modal(props ModalProps) core.Node {
	// Apply defaults
	size := props.Size
	if size == "" {
		size = ModalMD
	}

	// Overlay - covers entire screen, hidden when closed
	// Use conditional flex to avoid CSS specificity issues with hidden
	overlayClass := MergeClasses(
		"fixed inset-0 bg-black/50 items-center justify-center z-50",
		ConditionalClass(props.Open, "flex"),
		ConditionalClass(!props.Open, "hidden"),
	)

	// Modal container
	modalClass := MergeClasses(
		"bg-white dark:bg-gray-800 rounded-lg shadow-2xl border border-gray-200 dark:border-gray-600 max-h-[90vh] flex flex-col",
		modalSizeClasses[size],
		props.Class,
	)

	// Build ARIA attributes
	extra := map[string]string{
		"role":       "dialog",
		"aria-modal": "true",
	}
	if props.Title != "" {
		extra["aria-labelledby"] = "modal-title"
	}

	modalAttrs := core.Attrs{
		Class: modalClass,
		Extra: extra,
	}

	// Build content
	var content []core.Node

	// Header with title and close button (when either Title or OnClose provided)
	if props.Title != "" || props.OnClose != nil || props.XOnClose != "" {
		headerChildren := []core.Node{}

		if props.Title != "" {
			headerChildren = append(headerChildren,
				core.H3(core.Attrs{
					ID:    "modal-title",
					Class: "text-lg font-semibold text-gray-900 dark:text-gray-100",
				}, core.Text(props.Title)),
			)
		} else {
			// Spacer when no title but OnClose is provided
			headerChildren = append(headerChildren, core.Div(core.Attrs{}))
		}

		if props.OnClose != nil || props.XOnClose != "" {
			closeAttrs := core.Attrs{
				Class: "text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-2xl leading-none cursor-pointer",
				Extra: map[string]string{"aria-label": "Close"},
			}
			if props.XOnClose != "" {
				closeAttrs.XOn = map[string]string{"click": props.XOnClose}
			} else {
				closeAttrs.OnClick = props.OnClose
			}
			headerChildren = append(headerChildren,
				core.Button(closeAttrs, core.Text("×")),
			)
		}

		content = append(content,
			core.Div(core.Class("flex justify-between items-center px-6 py-4 border-b border-gray-200 dark:border-gray-700"),
				headerChildren...),
		)
	}

	// Children
	content = append(content, props.Children...)

	overlayAttrs := core.Attrs{
		Class:   overlayClass,
		OnClick: props.OnClose, // Clicking backdrop closes
	}

	// Alpine.js directives on overlay
	if props.XShow != "" {
		overlayAttrs.XShow = props.XShow
		overlayAttrs.XCloak = true
		overlayAttrs.XTransition = "true"
	}
	if props.XOnClose != "" {
		overlayAttrs.XOn = map[string]string{"click": props.XOnClose}
	}

	return core.Div(overlayAttrs,
		core.Div(modalAttrs, content...),
	)
}

// ModalContentProps configures the ModalContent component.
type ModalContentProps struct {
	Class    string      // Additional classes
	Children []core.Node // Content
}

// ModalContent wraps modal body content with appropriate padding and scroll.
//
// Example:
//
//	ModalContent(ModalContentProps{
//	    Children: []core.Node{core.Text("Modal body content here")},
//	})
func ModalContent(props ModalContentProps) core.Node {
	class := MergeClasses("px-6 py-4 overflow-y-auto flex-1", props.Class)
	return core.Div(core.Class(class), props.Children...)
}

// ModalFooterProps configures the ModalFooter component.
type ModalFooterProps struct {
	Class    string      // Additional classes
	Children []core.Node // Content (typically action buttons)
}

// ModalFooter wraps modal actions with appropriate styling.
//
// Example:
//
//	ModalFooter(ModalFooterProps{
//	    Children: []core.Node{
//	        Button(ButtonProps{Children: []core.Node{core.Text("Cancel")}}),
//	        Button(ButtonProps{Variant: ButtonPrimary, Children: []core.Node{core.Text("Save")}}),
//	    },
//	})
func ModalFooter(props ModalFooterProps) core.Node {
	class := MergeClasses("px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-2", props.Class)
	return core.Div(core.Class(class), props.Children...)
}
