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
	container  js.Value
	input      js.Value
	label      js.Value
	checkboxID string
}

// NewCheckbox creates a new Checkbox component
func NewCheckbox(props CheckboxProps) *Checkbox {
	document := js.Global().Get("document")
	crypto := js.Global().Get("crypto")

	container := document.Call("createElement", "div")
	container.Set("className", "flex items-center mb-4")

	// Generate unique ID for label-input association
	checkboxID := "checkbox-" + crypto.Call("randomUUID").String()

	cb := &Checkbox{container: container, checkboxID: checkboxID}

	// Checkbox input
	input := document.Call("createElement", "input")
	input.Set("type", "checkbox")
	input.Set("id", checkboxID)
	className := "h-4 w-4 text-blue-600 border-default rounded focus:ring-blue-500 surface-base"
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
		labelClass := "ml-2 text-sm text-secondary cursor-pointer"
		if props.Disabled {
			labelClass += " text-disabled"
		}
		label.Set("className", labelClass)
		label.Set("textContent", props.Label)
		label.Set("htmlFor", checkboxID)

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
