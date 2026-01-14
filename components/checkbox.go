//go:build js && wasm

package components

import "syscall/js"

// CheckboxProps configures a Checkbox component
type CheckboxProps struct {
	Label     string
	Checked   bool
	ClassName string
	Disabled  bool
	OnChange  func(checked bool)
}

// Checkbox creates a checkbox input with label
type Checkbox struct {
	container js.Value
	input     js.Value
	label     js.Value
}

// NewCheckbox creates a new Checkbox component
func NewCheckbox(props CheckboxProps) *Checkbox {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "flex items-center mb-4")

	cb := &Checkbox{container: container}

	// Checkbox input
	input := document.Call("createElement", "input")
	input.Set("type", "checkbox")
	className := "h-4 w-4 text-blue-600 border-gray-300 dark:border-gray-600 rounded focus:ring-blue-500 dark:bg-gray-700"
	if props.Disabled {
		className += " cursor-not-allowed"
	}
	input.Set("className", className)

	if props.Checked {
		input.Set("checked", true)
	}
	if props.Disabled {
		input.Set("disabled", true)
	}

	if props.OnChange != nil {
		input.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) any {
			checked := input.Get("checked").Bool()
			props.OnChange(checked)
			return nil
		}))
	}

	container.Call("appendChild", input)
	cb.input = input

	// Label
	if props.Label != "" {
		label := document.Call("createElement", "label")
		labelClass := "ml-2 text-sm text-gray-700 dark:text-gray-300"
		if props.Disabled {
			labelClass += " text-gray-400 dark:text-gray-500"
		}
		label.Set("className", labelClass)
		label.Set("textContent", props.Label)

		// Click label to toggle checkbox
		label.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			if !props.Disabled {
				input.Call("click")
			}
			return nil
		}))

		container.Call("appendChild", label)
		cb.label = label
	}

	return cb
}

// Element returns the container DOM element
func (c *Checkbox) Element() js.Value {
	return c.container
}

// Checked returns whether the checkbox is checked
func (c *Checkbox) Checked() bool {
	return c.input.Get("checked").Bool()
}

// SetChecked sets the checkbox state
func (c *Checkbox) SetChecked(checked bool) {
	c.input.Set("checked", checked)
}
