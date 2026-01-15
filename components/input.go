//go:build js && wasm

package components

import "syscall/js"

// InputType defines the input type
type InputType string

const (
	InputText     InputType = "text"
	InputPassword InputType = "password"
	InputEmail    InputType = "email"
	InputNumber   InputType = "number"
	InputSearch   InputType = "search"
	InputTel      InputType = "tel"
	InputURL      InputType = "url"
)

// InputProps configures an Input component
type InputProps struct {
	Type        InputType
	Label       string
	Placeholder string
	Value       string
	ClassName   string
	Disabled    bool
	Required    bool
	OnChange    func(value string)
	OnEnter     func(value string)
}

// Input creates a labeled text input field
type Input struct {
	container js.Value
	input     js.Value
	label     js.Value
	errorEl   js.Value
	inputID   string
	errorID   string
}

// NewInput creates a new Input component
func NewInput(props InputProps) *Input {
	document := js.Global().Get("document")
	crypto := js.Global().Get("crypto")

	container := document.Call("createElement", "div")
	container.Set("className", "mb-4")

	inputType := props.Type
	if inputType == "" {
		inputType = InputText
	}

	// Generate unique ID for label-input association
	inputID := "input-" + crypto.Call("randomUUID").String()

	inp := &Input{container: container, inputID: inputID}

	// Label
	if props.Label != "" {
		label := document.Call("createElement", "label")
		label.Set("className", "block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1")
		label.Set("textContent", props.Label)
		label.Set("htmlFor", inputID)
		container.Call("appendChild", label)
		inp.label = label
	}

	// Input field
	input := document.Call("createElement", "input")
	className := "w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 placeholder-gray-500 dark:placeholder-gray-400"
	if props.Disabled {
		className += " bg-gray-100 dark:bg-gray-800 cursor-not-allowed"
	}
	if props.ClassName != "" {
		className = props.ClassName
	}

	input.Set("type", string(inputType))
	input.Set("id", inputID)
	input.Set("className", className)

	if props.Placeholder != "" {
		input.Set("placeholder", props.Placeholder)
	}
	if props.Value != "" {
		input.Set("value", props.Value)
	}
	if props.Disabled {
		input.Set("disabled", true)
	}
	if props.Required {
		input.Set("required", true)
	}

	// Event handlers
	if props.OnChange != nil {
		input.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
			value := input.Get("value").String()
			props.OnChange(value)
			return nil
		}))
	}

	if props.OnEnter != nil {
		input.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) any {
			if args[0].Get("key").String() == "Enter" {
				value := input.Get("value").String()
				props.OnEnter(value)
			}
			return nil
		}))
	}

	container.Call("appendChild", input)
	inp.input = input

	return inp
}

// Element returns the container DOM element
func (i *Input) Element() js.Value {
	return i.container
}

// Value returns the current input value
func (i *Input) Value() string {
	return i.input.Get("value").String()
}

// SetValue sets the input value
func (i *Input) SetValue(value string) {
	i.input.Set("value", value)
}

// Focus sets focus on the input
func (i *Input) Focus() {
	i.input.Call("focus")
}

// SetError adds error styling to the input and ARIA error attributes
func (i *Input) SetError(message string) {
	document := js.Global().Get("document")
	crypto := js.Global().Get("crypto")

	i.input.Set("className", "w-full px-3 py-2 border border-red-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-red-500")
	i.input.Call("setAttribute", "aria-invalid", "true")

	// Generate error ID if not already set
	if i.errorID == "" {
		i.errorID = "input-error-" + crypto.Call("randomUUID").String()
	}

	// Create or update error message element
	if i.errorEl.IsUndefined() || i.errorEl.IsNull() {
		i.errorEl = document.Call("createElement", "p")
		i.errorEl.Set("id", i.errorID)
		i.errorEl.Set("className", "text-red-500 text-sm mt-1")
		i.errorEl.Call("setAttribute", "role", "alert")
		i.container.Call("appendChild", i.errorEl)
	}
	i.errorEl.Set("textContent", message)
	i.errorEl.Get("classList").Call("remove", "hidden")

	// Link error message to input via aria-describedby
	i.input.Call("setAttribute", "aria-describedby", i.errorID)
}

// ClearError removes error styling and ARIA error attributes
func (i *Input) ClearError() {
	i.input.Set("className", "w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500")
	i.input.Call("removeAttribute", "aria-invalid")
	i.input.Call("removeAttribute", "aria-describedby")

	// Hide error message element
	if !i.errorEl.IsUndefined() && !i.errorEl.IsNull() {
		i.errorEl.Get("classList").Call("add", "hidden")
	}
}

// Quick input constructors

// TextInput creates a simple text input with label and placeholder
func TextInput(label, placeholder string) *Input {
	return NewInput(InputProps{Label: label, Placeholder: placeholder})
}

// EmailInput creates an email input with label and placeholder
func EmailInput(label, placeholder string) *Input {
	return NewInput(InputProps{Type: InputEmail, Label: label, Placeholder: placeholder})
}

// PasswordInput creates a password input with label and placeholder
func PasswordInput(label, placeholder string) *Input {
	return NewInput(InputProps{Type: InputPassword, Label: label, Placeholder: placeholder})
}

// NumberInput creates a number input with label and placeholder
func NumberInput(label, placeholder string) *Input {
	return NewInput(InputProps{Type: InputNumber, Label: label, Placeholder: placeholder})
}

// SearchInput creates a search input with placeholder
func SearchInput(placeholder string) *Input {
	return NewInput(InputProps{Type: InputSearch, Placeholder: placeholder})
}
