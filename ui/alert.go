package ui

import "github.com/dougbarrett/gux/core"

// AlertVariant defines the visual style of an alert.
type AlertVariant string

const (
	AlertInfo    AlertVariant = "info"
	AlertSuccess AlertVariant = "success"
	AlertWarning AlertVariant = "warning"
	AlertError   AlertVariant = "error"
)

// alertVariantStyle holds the classes for each variant.
type alertVariantStyle struct {
	bg     string // Background classes
	border string // Border classes
	text   string // Text color classes
	icon   string // Icon character
}

// alertVariantStyles maps variants to their styling.
var alertVariantStyles = map[AlertVariant]alertVariantStyle{
	AlertInfo: {
		bg:     "bg-blue-50 dark:bg-blue-950",
		border: "border-blue-200 dark:border-blue-800",
		text:   "text-blue-800 dark:text-blue-200",
		icon:   "\u24d8", // circled lowercase i
	},
	AlertSuccess: {
		bg:     "bg-green-50 dark:bg-green-950",
		border: "border-green-200 dark:border-green-800",
		text:   "text-green-800 dark:text-green-200",
		icon:   "\u2713", // checkmark
	},
	AlertWarning: {
		bg:     "bg-yellow-50 dark:bg-yellow-950",
		border: "border-yellow-200 dark:border-yellow-800",
		text:   "text-yellow-800 dark:text-yellow-200",
		icon:   "\u26a0", // warning sign
	},
	AlertError: {
		bg:     "bg-red-50 dark:bg-red-950",
		border: "border-red-200 dark:border-red-800",
		text:   "text-red-800 dark:text-red-200",
		icon:   "\u2715", // x mark
	},
}

// AlertProps configures an Alert component.
type AlertProps struct {
	Message string       // Alert text content (required)
	Variant AlertVariant // Visual style (default: info)
	Title   string       // Optional bold title before message
	OnClose func()       // Optional close handler (renders close button when provided)
	Class   string       // Additional classes
}

// Alert creates an inline alert box for feedback messages.
//
// Unlike Toast which auto-dismisses and appears in a fixed position,
// Alert is an inline element that persists until dismissed. Ideal for
// form validation errors, success messages, and important notices.
//
// Example:
//
//	Alert(AlertProps{
//	    Message: "Invalid email or password",
//	    Variant: AlertError,
//	    Title:   "Login Failed",
//	    OnClose: func() { showError.Set(false) },
//	})
func Alert(props AlertProps) core.Node {
	// Apply defaults
	variant := props.Variant
	if variant == "" {
		variant = AlertInfo
	}

	style := alertVariantStyles[variant]

	// Build class string
	class := MergeClasses(
		"px-4 py-3 rounded-lg border flex items-start gap-3",
		style.bg,
		style.border,
		style.text,
		props.Class,
	)

	// Build content children
	contentChildren := []core.Node{}

	// Optional title
	if props.Title != "" {
		contentChildren = append(contentChildren,
			core.El("strong", core.Class("block font-medium"), core.Text(props.Title)),
		)
	}

	// Message
	contentChildren = append(contentChildren,
		core.Span(core.Attrs{}, core.Text(props.Message)),
	)

	// Build main children
	children := []core.Node{
		// Icon (decorative, hidden from screen readers)
		core.Span(core.Attrs{
			Class: "text-lg flex-shrink-0",
			Extra: map[string]string{"aria-hidden": "true"},
		}, core.Text(style.icon)),
		// Content
		core.Div(core.Class("flex-1"), contentChildren...),
	}

	// Conditional close button
	if props.OnClose != nil {
		children = append(children,
			core.Button(core.Attrs{
				Class:   "ml-auto opacity-70 hover:opacity-100 cursor-pointer flex-shrink-0",
				OnClick: props.OnClose,
				Extra:   map[string]string{"aria-label": "Dismiss"},
			}, core.Text("\u00d7")), // multiplication sign (x)
		)
	}

	return core.Div(core.Attrs{
		Class: class,
		Extra: map[string]string{"role": "alert"},
	}, children...)
}
