package ui

import "github.com/dougbarrett/gux/core"

// selectBaseClasses are the common classes applied to all select elements.
const selectBaseClasses = inputBaseClasses + " appearance-none"

// SelectOption represents a single option in a select dropdown.
type SelectOption struct {
	Value    string // Option value attribute
	Label    string // Display text
	Disabled bool   // Disabled state for this option
}

// StringState allows form components to directly bind to state.
// Implemented by *core.State[string] from r.StateString().
type StringState interface {
	Get() string
	Set(string)
}

// SelectProps configures the Select component.
type SelectProps struct {
	Size        InputSize      // Size variant (default: md)
	ID          string         // Element ID for label association
	Name        string         // Select name attribute
	Value       string         // Currently selected value (use Bind instead for cleaner code)
	Options     []SelectOption // Available options
	Placeholder string         // Placeholder option text (e.g., "Select...")
	Class       string         // Additional classes
	Disabled    bool           // Disabled state
	Required    bool           // Required indicator
	Error       string         // Error message (adds aria-invalid when non-empty)
	OnChange    func(string)   // Selection change handler (use Bind instead for cleaner code)

	// Bind directly binds this select to a state value.
	// When set, Value and OnChange are automatically handled.
	// Usage: Bind: r.StateString("country", "")
	Bind StringState

	// Alpine.js props (used when in Alpine mode)
	XModel    string // x-model expression (e.g., "country")
	XOnChange string // x-on:change expression
}

// Select creates a styled dropdown select component.
//
// Example with state binding (recommended):
//
//	Select(SelectProps{
//	    Name:        "country",
//	    Bind:        r.StateString("country", ""),
//	    Placeholder: "Select a country",
//	    Options: []SelectOption{
//	        {Value: "us", Label: "United States"},
//	        {Value: "uk", Label: "United Kingdom"},
//	        {Value: "ca", Label: "Canada"},
//	    },
//	})
//
// Example with manual value/onChange (legacy):
//
//	Select(SelectProps{
//	    Name:        "country",
//	    Value:       countryState.Get(),
//	    Placeholder: "Select a country",
//	    Options: []SelectOption{
//	        {Value: "us", Label: "United States"},
//	        {Value: "uk", Label: "United Kingdom"},
//	        {Value: "ca", Label: "Canada"},
//	    },
//	    OnChange: func(v string) { countryState.Set(v) },
//	})
func Select(props SelectProps) core.Node {
	// Handle state binding - auto-wire Value and OnChange
	value := props.Value
	onChange := props.OnChange
	if props.Bind != nil {
		value = props.Bind.Get()
		onChange = props.Bind.Set
	}

	// Apply defaults
	size := props.Size
	if size == "" {
		size = InputMD
	}

	// Build class string
	// Note: pr-10 must come after size classes to prevent px-* from overriding
	class := MergeClasses(
		selectBaseClasses,
		inputSizeClasses[size],
		"pr-10", // Arrow padding - must be after size classes
		ConditionalClass(props.Error != "", inputErrorClasses),
		ConditionalClass(props.Disabled, inputDisabledClasses),
		props.Class,
	)

	// Build attributes
	attrs := core.Attrs{
		ID:       props.ID,
		Name:     props.Name,
		Class:    class,
		OnChange: onChange,
	}

	// Alpine.js directives
	if props.XModel != "" {
		attrs.XModel = props.XModel
	}
	if props.XOnChange != "" {
		attrs.XOn = map[string]string{"change": props.XOnChange}
	}

	// Build extra attributes
	extra := make(map[string]string)
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
	if len(extra) > 0 {
		attrs.Extra = extra
	}

	// Build option nodes
	var options []core.Node

	// Placeholder option (if specified)
	if props.Placeholder != "" {
		placeholderAttrs := core.Attrs{
			Value: "",
		}
		placeholderExtra := map[string]string{
			"disabled": "disabled",
		}
		// Select placeholder if no value is set
		if value == "" {
			placeholderExtra["selected"] = "selected"
		}
		placeholderAttrs.Extra = placeholderExtra
		options = append(options, core.Option(placeholderAttrs, core.Text(props.Placeholder)))
	}

	// Regular options
	for _, opt := range props.Options {
		optAttrs := core.Attrs{
			Value: opt.Value,
		}

		// Build option extra attributes
		optExtra := make(map[string]string)
		if opt.Value == value {
			optExtra["selected"] = "selected"
		}
		if opt.Disabled {
			optExtra["disabled"] = "disabled"
		}
		if len(optExtra) > 0 {
			optAttrs.Extra = optExtra
		}

		options = append(options, core.Option(optAttrs, core.Text(opt.Label)))
	}

	return core.Select(attrs, options...)
}
