package ui

import "github.com/dougbarrett/gux/core"

// ToastPosition defines where the toast container is positioned.
type ToastPosition string

const (
	ToastTopRight     ToastPosition = "top-right"
	ToastTopLeft      ToastPosition = "top-left"
	ToastBottomRight  ToastPosition = "bottom-right"
	ToastBottomLeft   ToastPosition = "bottom-left"
	ToastTopCenter    ToastPosition = "top-center"
	ToastBottomCenter ToastPosition = "bottom-center"
)

// toastPositionClasses maps positions to Tailwind positioning classes.
var toastPositionClasses = map[ToastPosition]string{
	ToastTopRight:     "top-4 right-4",
	ToastTopLeft:      "top-4 left-4",
	ToastBottomRight:  "bottom-4 right-4",
	ToastBottomLeft:   "bottom-4 left-4",
	ToastTopCenter:    "top-4 left-1/2 -translate-x-1/2",
	ToastBottomCenter: "bottom-4 left-1/2 -translate-x-1/2",
}

// ToastContainerProps configures a ToastContainer component.
type ToastContainerProps struct {
	Position ToastPosition // Positioning of container (default: top-right)
	Class    string        // Additional classes
	Children []core.Node   // Individual Toast components
}

// ToastContainer creates a fixed-position container for toast notifications.
//
// The container uses ARIA live region semantics to announce toast additions
// to assistive technologies. Render this once at the app root level.
//
// Example:
//
//	ToastContainer(ToastContainerProps{
//	    Position: ToastTopRight,
//	    Children: []core.Node{
//	        Toast(ToastProps{
//	            Message: "Action successful!",
//	            Variant: ToastSuccess,
//	        }),
//	    },
//	})
func ToastContainer(props ToastContainerProps) core.Node {
	// Apply defaults
	position := props.Position
	if position == "" {
		position = ToastTopRight
	}

	// Build class string
	class := MergeClasses(
		"fixed z-50 flex flex-col gap-2",
		toastPositionClasses[position],
		props.Class,
	)

	return core.Div(core.Attrs{
		ID:    "toast-container",
		Class: class,
		Extra: map[string]string{
			"role":        "status",
			"aria-live":   "polite",
			"aria-atomic": "true",
		},
	}, props.Children...)
}

// ToastVariant defines the visual style of a toast.
type ToastVariant string

const (
	ToastInfo    ToastVariant = "info"
	ToastSuccess ToastVariant = "success"
	ToastWarning ToastVariant = "warning"
	ToastError   ToastVariant = "error"
)

// toastVariantStyle holds the classes for each variant.
type toastVariantStyle struct {
	bg   string // Background classes
	text string // Text color classes
	icon string // Icon character
}

// toastVariantStyles maps variants to their styling.
var toastVariantStyles = map[ToastVariant]toastVariantStyle{
	ToastInfo: {
		bg:   "bg-blue-100 dark:bg-blue-900",
		text: "text-blue-800 dark:text-blue-200",
		icon: "i",
	},
	ToastSuccess: {
		bg:   "bg-green-100 dark:bg-green-900",
		text: "text-green-800 dark:text-green-200",
		icon: "\u2713", // checkmark
	},
	ToastWarning: {
		bg:   "bg-yellow-100 dark:bg-yellow-900",
		text: "text-yellow-800 dark:text-yellow-200",
		icon: "!",
	},
	ToastError: {
		bg:   "bg-red-100 dark:bg-red-900",
		text: "text-red-800 dark:text-red-200",
		icon: "\u2715", // x mark
	},
}

// ToastProps configures a Toast component.
type ToastProps struct {
	Message string       // Toast text content
	Variant ToastVariant // Visual style (default: info)
	OnClose func()       // Optional close handler (renders close button when provided)
	Class   string       // Additional classes
}

// Toast creates a single toast notification.
//
// The toast displays a message with variant-specific styling and an optional
// close button. Typically used within a ToastContainer with page-level state
// to manage visible toasts.
//
// Example:
//
//	Toast(ToastProps{
//	    Message: "Changes saved successfully!",
//	    Variant: ToastSuccess,
//	    OnClose: func() { showToast.Set(false) },
//	})
func Toast(props ToastProps) core.Node {
	// Apply defaults
	variant := props.Variant
	if variant == "" {
		variant = ToastInfo
	}

	style := toastVariantStyles[variant]

	// Build class string
	class := MergeClasses(
		"px-4 py-3 rounded-lg shadow-lg flex items-center gap-3 min-w-64",
		style.bg,
		style.text,
		props.Class,
	)

	// Build children
	children := []core.Node{
		// Icon (decorative, hidden from screen readers)
		core.Span(core.Attrs{
			Class: "text-lg font-bold",
			Extra: map[string]string{"aria-hidden": "true"},
		}, core.Text(style.icon)),
		// Message
		core.Span(core.Class("flex-1"), core.Text(props.Message)),
	}

	// Conditional close button
	if props.OnClose != nil {
		children = append(children,
			core.Button(core.Attrs{
				Class:   "opacity-70 hover:opacity-100 cursor-pointer",
				OnClick: props.OnClose,
				Extra:   map[string]string{"aria-label": "Dismiss notification"},
			}, core.Text("\u00d7")), // multiplication sign (x)
		)
	}

	return core.Div(core.Class(class), children...)
}
