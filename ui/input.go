package ui

import "github.com/dougbarrett/gux/core"

// InputType defines the HTML input type.
type InputType string

const (
	InputText     InputType = "text"
	InputEmail    InputType = "email"
	InputPassword InputType = "password"
	InputNumber   InputType = "number"
	InputSearch   InputType = "search"
	InputTel      InputType = "tel"
	InputURL      InputType = "url"
)

// InputSize defines the size of an input.
type InputSize string

const (
	InputSM InputSize = "sm"
	InputMD InputSize = "md"
	InputLG InputSize = "lg"
)

// inputSizeClasses maps sizes to their Tailwind classes.
var inputSizeClasses = map[InputSize]string{
	InputSM: "px-2 py-1.5 text-sm",
	InputMD: "px-3 py-2 text-base",
	InputLG: "px-4 py-3 text-lg",
}

// inputBaseClasses are the common classes applied to all inputs.
const inputBaseClasses = "w-full border rounded-md shadow-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 border-gray-300 dark:border-gray-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 placeholder:text-gray-400 dark:placeholder:text-gray-500"

// inputErrorClasses are applied when input has an error.
const inputErrorClasses = "border-red-500 dark:border-red-400 focus:ring-red-500 focus:border-red-500"

// inputDisabledClasses are applied when input is disabled.
const inputDisabledClasses = "bg-gray-100 dark:bg-gray-700 cursor-not-allowed opacity-50"

// InputProps configures the Input component.
type InputProps struct {
	Type        InputType    // Input type (default: text)
	Size        InputSize    // Size variant (default: md)
	ID          string       // Element ID for label association
	Name        string       // Input name attribute
	Value       string       // Current value (use Bind instead for cleaner code)
	Placeholder string       // Placeholder text
	Class       string       // Additional classes
	Disabled    bool         // Disabled state
	Required    bool         // Required indicator
	Error       string       // Error message (adds aria-invalid when non-empty)
	OnChange    func(string) // Value change handler (use Bind instead for cleaner code)
	OnEnter     func()       // Enter key handler (WASM only)

	// Bind directly binds this input to a state value.
	// When set, Value and OnChange are automatically handled.
	// Usage: Bind: r.StateString("email", "")
	Bind StringState

	// Alpine.js props (used when in Alpine mode)
	XModel    string // x-model expression (e.g., "email")
	XOnChange string // x-on:change expression
	XOnInput  string // x-on:input expression
	XOnEnter  string // x-on:keydown.enter expression
}

// Input creates a styled text input component.
//
// Example with state binding (recommended):
//
//	Input(InputProps{
//	    Type:        InputEmail,
//	    Name:        "email",
//	    Bind:        r.StateString("email", ""),
//	    Placeholder: "you@example.com",
//	})
//
// Example with manual value/onChange (legacy):
//
//	Input(InputProps{
//	    Type:        InputEmail,
//	    Name:        "email",
//	    Value:       emailState.Get(),
//	    Placeholder: "you@example.com",
//	    OnChange:    func(v string) { emailState.Set(v) },
//	})
func Input(props InputProps) core.Node {
	// Handle state binding - auto-wire Value and OnChange
	value := props.Value
	onChange := props.OnChange
	if props.Bind != nil {
		value = props.Bind.Get()
		onChange = props.Bind.Set
	}

	// Apply defaults
	inputType := props.Type
	if inputType == "" {
		inputType = InputText
	}

	size := props.Size
	if size == "" {
		size = InputMD
	}

	// Build class string
	class := MergeClasses(
		inputBaseClasses,
		inputSizeClasses[size],
		ConditionalClass(props.Error != "", inputErrorClasses),
		ConditionalClass(props.Disabled, inputDisabledClasses),
		props.Class,
	)

	// Build attributes
	attrs := core.Attrs{
		ID:       props.ID,
		Type:     string(inputType),
		Name:     props.Name,
		Value:    value,
		Class:    class,
		OnChange: onChange,
		OnEnter:  props.OnEnter,
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
	if props.XOnEnter != "" {
		xOn["keydown.enter"] = props.XOnEnter
	}
	if len(xOn) > 0 {
		attrs.XOn = xOn
	}

	// Build extra attributes
	extra := make(map[string]string)
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
	if len(extra) > 0 {
		attrs.Extra = extra
	}

	return core.Input(attrs)
}
