package ui

import "github.com/dougbarrett/gux/core"

// Radio class constants
const (
	// radioBaseClasses defines the default styling for radio inputs.
	radioBaseClasses = "h-4 w-4 text-blue-600 border-gray-300 dark:border-gray-600 focus:ring-blue-500 focus:ring-offset-0 bg-white dark:bg-gray-800"

	// radioDisabledClasses are applied when radio is disabled.
	radioDisabledClasses = "cursor-not-allowed opacity-50"
)

// RadioOptionProps configures a single radio option.
type RadioOptionProps struct {
	Value       string // Option value
	Label       string // Display label
	Description string // Optional description text below label
	Disabled    bool   // Disabled state for this option
}

// RadioGroupProps configures the RadioGroup component.
type RadioGroupProps struct {
	Name     string             // Group name (required - groups radios together)
	Value    string             // Currently selected value
	Options  []RadioOptionProps // Available options
	Inline   bool               // Display horizontally (default: false = vertical)
	Class    string             // Additional wrapper classes
	Disabled bool               // Disable all options
	Required bool               // Required indicator
	Error    string             // Error message (for aria-invalid)
}

// RadioGroup creates a group of mutually exclusive radio options.
//
// Example:
//
//	RadioGroup(RadioGroupProps{
//	    Name:  "status",
//	    Value: statusState.Get(),
//	    Options: []RadioOptionProps{
//	        {Value: "active", Label: "Active"},
//	        {Value: "inactive", Label: "Inactive"},
//	        {Value: "pending", Label: "Pending", Description: "Awaiting approval"},
//	    },
//	})
func RadioGroup(props RadioGroupProps) core.Node {
	// Build wrapper class
	wrapperClass := MergeClasses(
		ConditionalClass(props.Inline, "flex flex-row gap-4"),
		ConditionalClass(!props.Inline, "flex flex-col gap-2"),
		props.Class,
	)

	// Build option nodes
	var children []core.Node
	for _, opt := range props.Options {
		isChecked := opt.Value == props.Value
		isDisabled := props.Disabled || opt.Disabled

		// Build radio input class
		radioClass := MergeClasses(
			radioBaseClasses,
			ConditionalClass(isDisabled, radioDisabledClasses),
		)

		// Build input attributes
		inputAttrs := core.Attrs{
			Type:  "radio",
			Name:  props.Name,
			Value: opt.Value,
			Class: radioClass,
		}

		// Build extra attributes
		// CRITICAL: Only add "checked" when this option is selected (SSR boolean attrs are presence-based)
		extra := make(map[string]string)
		if isChecked {
			extra["checked"] = "checked"
		}
		if isDisabled {
			extra["disabled"] = "disabled"
		}
		if props.Required {
			extra["required"] = "required"
		}
		if len(extra) > 0 {
			inputAttrs.Extra = extra
		}

		input := core.Input(inputAttrs)

		// Build label content
		labelContent := []core.Node{
			core.Span(core.Class("text-sm font-medium text-gray-700 dark:text-gray-300"),
				core.Text(opt.Label),
			),
		}

		if opt.Description != "" {
			labelContent = append(labelContent,
				core.Span(core.Class("block text-sm text-gray-500 dark:text-gray-400"),
					core.Text(opt.Description),
				),
			)
		}

		// Build option wrapper class
		optionClass := MergeClasses(
			"flex items-start gap-2",
			ConditionalClass(isDisabled, "opacity-50 cursor-not-allowed"),
		)

		// Create label wrapping input + content
		children = append(children,
			core.Label(core.Class(optionClass),
				input,
				core.Div(core.Attrs{}, labelContent...),
			),
		)
	}

	// Build wrapper attributes
	wrapperAttrs := core.Attrs{
		Class: wrapperClass,
		Extra: map[string]string{"role": "radiogroup"},
	}

	// Add aria-invalid if there's an error
	if props.Error != "" {
		wrapperAttrs.Extra["aria-invalid"] = "true"
	}

	return core.Div(wrapperAttrs, children...)
}
