package ui

import (
	"strconv"

	"github.com/dougbarrett/gux/core"
)

// InputSize, inputSizeClasses, inputBaseClasses, inputErrorClasses, and
// inputDisabledClasses are defined in input.go and shared across form components.

// TextareaResize defines how a textarea can be resized.
type TextareaResize string

const (
	ResizeNone     TextareaResize = "none"
	ResizeVertical TextareaResize = "vertical"
	ResizeBoth     TextareaResize = "both"
)

// resizeClasses maps resize options to Tailwind classes.
var resizeClasses = map[TextareaResize]string{
	ResizeNone:     "resize-none",
	ResizeVertical: "resize-y",
	ResizeBoth:     "resize",
}

// TextareaProps configures the Textarea component.
type TextareaProps struct {
	Size        InputSize      // Size variant (default: md)
	Resize      TextareaResize // Resize behavior (default: vertical)
	Rows        int            // Number of visible rows (default: 3)
	ID          string         // Element ID for label association
	Name        string         // Textarea name attribute
	Value       string         // Current value
	Placeholder string         // Placeholder text
	Class       string         // Additional classes
	Disabled    bool           // Disabled state
	Required    bool           // Required indicator
	Error       string         // Error message (adds aria-invalid)
	OnChange    func(string)   // Value change handler (WASM only)

	// Alpine.js props
	XModel    string // x-model expression
	XOnChange string // x-on:change expression
	XOnInput  string // x-on:input expression
}

// Textarea creates a styled multi-line text input component.
//
// Example:
//
//	Textarea(TextareaProps{
//	    Name:        "description",
//	    Value:       descState.Get(),
//	    Placeholder: "Enter description...",
//	    Rows:        5,
//	    OnChange:    func(v string) { descState.Set(v) },
//	})
func Textarea(props TextareaProps) core.Node {
	// Apply defaults
	size := props.Size
	if size == "" {
		size = InputMD
	}

	resize := props.Resize
	if resize == "" {
		resize = ResizeVertical
	}

	rows := props.Rows
	if rows <= 0 {
		rows = 3
	}

	// Build class string
	class := MergeClasses(
		inputBaseClasses,
		inputSizeClasses[size],
		resizeClasses[resize],
		ConditionalClass(props.Error != "", inputErrorClasses),
		ConditionalClass(props.Disabled, inputDisabledClasses),
		props.Class,
	)

	// Build attributes
	attrs := core.Attrs{
		ID:       props.ID,
		Name:     props.Name,
		Class:    class,
		OnChange: props.OnChange,
	}

	// Alpine.js directives
	if props.XModel != "" {
		attrs.XModel = props.XModel
	}
	xOn := make(map[string]string)
	if props.XOnChange != "" {
		xOn["change"] = props.XOnChange
	}
	if props.XOnInput != "" {
		xOn["input"] = props.XOnInput
	}
	if len(xOn) > 0 {
		attrs.XOn = xOn
	}

	// Build extra attributes
	extra := make(map[string]string)
	extra["rows"] = strconv.Itoa(rows)
	if props.Placeholder != "" {
		extra["placeholder"] = props.Placeholder
	}
	if props.Disabled {
		extra["disabled"] = "disabled"
	}
	if props.Required {
		extra["required"] = "required"
		extra["aria-required"] = "true"
	}
	if props.Error != "" {
		extra["aria-invalid"] = "true"
	}
	attrs.Extra = extra

	// Pass value as child for SSR rendering (textarea needs content as child)
	return core.Textarea(attrs, core.Text(props.Value))
}
