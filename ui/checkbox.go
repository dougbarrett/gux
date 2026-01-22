package ui

import "github.com/dougbarrett/gux/core"

// Checkbox class constants
const (
	// checkboxBaseClasses defines the default styling for checkbox inputs.
	checkboxBaseClasses = "h-4 w-4 rounded text-blue-600 border-gray-300 dark:border-gray-600 focus:ring-blue-500 focus:ring-offset-0 bg-white dark:bg-gray-800"

	// checkboxDisabledClasses are applied when checkbox is disabled.
	checkboxDisabledClasses = "cursor-not-allowed opacity-50"
)

// CheckboxProps configures a Checkbox component.
type CheckboxProps struct {
	ID          string // Element ID for label association
	Name        string // Input name attribute
	Checked     bool   // Checked state
	Label       string // Label text (optional - if set, wraps in label)
	Description string // Description text below label (optional)
	Class       string // Additional classes
	Disabled    bool   // Disabled state
}

// Checkbox creates a checkbox input with optional label and description.
//
// Note: core.Attrs.OnChange returns string, not bool, so checkbox change handlers
// are not directly supported. Use form submission or page-level JavaScript to
// get the checked state.
//
// Example:
//
//	Checkbox(CheckboxProps{
//	    ID:      "remember",
//	    Name:    "remember",
//	    Checked: rememberState.Get(),
//	    Label:   "Remember me",
//	})
func Checkbox(props CheckboxProps) core.Node {
	// Build checkbox input class
	checkboxClass := MergeClasses(
		checkboxBaseClasses,
		ConditionalClass(props.Disabled, checkboxDisabledClasses),
	)

	// Build attributes
	attrs := core.Attrs{
		ID:    props.ID,
		Type:  "checkbox",
		Name:  props.Name,
		Class: checkboxClass,
	}

	// Build extra attributes
	// CRITICAL: Only add "checked" when Checked=true (SSR boolean attrs are presence-based)
	extra := make(map[string]string)
	if props.Checked {
		extra["checked"] = "checked"
	}
	if props.Disabled {
		extra["disabled"] = "disabled"
	}
	if len(extra) > 0 {
		attrs.Extra = extra
	}

	input := core.Input(attrs)

	// Without label, return just the input
	if props.Label == "" {
		return input
	}

	// With label, wrap in label element
	wrapperClass := MergeClasses(
		"flex items-start gap-2",
		ConditionalClass(props.Disabled, "cursor-not-allowed"),
		props.Class,
	)

	// Build label content
	labelContent := []core.Node{
		core.Span(core.Class("text-sm font-medium text-gray-700 dark:text-gray-300"),
			core.Text(props.Label),
		),
	}

	if props.Description != "" {
		labelContent = append(labelContent,
			core.Span(core.Class("block text-sm text-gray-500 dark:text-gray-400"),
				core.Text(props.Description),
			),
		)
	}

	return core.Label(core.Class(wrapperClass),
		input,
		core.Div(core.Attrs{}, labelContent...),
	)
}
