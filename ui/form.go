package ui

import "github.com/dougbarrett/gux/core"

// FormFieldProps configures the FormField wrapper component.
type FormFieldProps struct {
	Label       string      // Label text
	LabelFor    string      // ID to associate label with input
	Required    bool        // Show required indicator (asterisk)
	Error       string      // Error message to display
	Description string      // Help text below input (shown only when no error)
	Class       string      // Additional wrapper classes
	Children    []core.Node // Input component(s)
}

// FormField creates a form field wrapper with label, input, and error display.
//
// Example:
//
//	FormField(FormFieldProps{
//	    Label:    "Email",
//	    LabelFor: "email-input",
//	    Required: true,
//	    Error:    emailError,
//	    Children: []core.Node{
//	        Input(InputProps{
//	            ID:   "email-input",
//	            Type: InputEmail,
//	            Name: "email",
//	        }),
//	    },
//	})
func FormField(props FormFieldProps) core.Node {
	var children []core.Node

	// Label
	if props.Label != "" {
		labelClass := "block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"

		labelChildren := []core.Node{core.Text(props.Label)}
		if props.Required {
			labelChildren = append(labelChildren,
				core.Span(core.Class("text-red-500 ml-1"), core.Text("*")),
			)
		}

		labelAttrs := core.Attrs{Class: labelClass}
		if props.LabelFor != "" {
			labelAttrs.Extra = map[string]string{"for": props.LabelFor}
		}

		children = append(children, core.Label(labelAttrs, labelChildren...))
	}

	// Input(s)
	children = append(children, props.Children...)

	// Error message
	if props.Error != "" {
		children = append(children,
			core.P(
				core.Attrs{
					Class: "text-sm text-red-600 dark:text-red-400 mt-1",
					Extra: map[string]string{"role": "alert"},
				},
				core.Text(props.Error),
			),
		)
	} else if props.Description != "" {
		// Description (only shown when no error)
		children = append(children,
			core.P(
				core.Class("text-sm text-gray-500 dark:text-gray-400 mt-1"),
				core.Text(props.Description),
			),
		)
	}

	class := MergeClasses("mb-4", props.Class)
	return core.Div(core.Class(class), children...)
}
