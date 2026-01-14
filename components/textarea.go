//go:build js && wasm

package components

import "syscall/js"

// TextAreaProps configures a TextArea component
type TextAreaProps struct {
	Label       string
	Placeholder string
	Value       string
	Rows        int
	ClassName   string
	Disabled    bool
	Required    bool
	OnChange    func(value string)
}

// TextArea creates a multi-line text input
type TextArea struct {
	container js.Value
	textarea  js.Value
}

// NewTextArea creates a new TextArea component
func NewTextArea(props TextAreaProps) *TextArea {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "mb-4")

	ta := &TextArea{container: container}

	// Label
	if props.Label != "" {
		label := document.Call("createElement", "label")
		label.Set("className", "block text-sm font-medium text-gray-700 mb-1")
		label.Set("textContent", props.Label)
		container.Call("appendChild", label)
	}

	// TextArea
	textarea := document.Call("createElement", "textarea")
	className := "w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-y"
	if props.Disabled {
		className += " bg-gray-100 cursor-not-allowed"
	}
	if props.ClassName != "" {
		className = props.ClassName
	}

	textarea.Set("className", className)

	rows := props.Rows
	if rows == 0 {
		rows = 4
	}
	textarea.Set("rows", rows)

	if props.Placeholder != "" {
		textarea.Set("placeholder", props.Placeholder)
	}
	if props.Value != "" {
		textarea.Set("value", props.Value)
	}
	if props.Disabled {
		textarea.Set("disabled", true)
	}
	if props.Required {
		textarea.Set("required", true)
	}

	if props.OnChange != nil {
		textarea.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
			value := textarea.Get("value").String()
			props.OnChange(value)
			return nil
		}))
	}

	container.Call("appendChild", textarea)
	ta.textarea = textarea

	return ta
}

// Element returns the container DOM element
func (t *TextArea) Element() js.Value {
	return t.container
}

// Value returns the current textarea value
func (t *TextArea) Value() string {
	return t.textarea.Get("value").String()
}

// SetValue sets the textarea value
func (t *TextArea) SetValue(value string) {
	t.textarea.Set("value", value)
}
