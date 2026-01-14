//go:build js && wasm

package components

import "syscall/js"

// ButtonProps configures a Button component
type ButtonProps struct {
	Text      string
	ClassName string
	OnClick   func()
}

// Button creates a styled button element
func Button(props ButtonProps) js.Value {
	document := js.Global().Get("document")
	btn := document.Call("createElement", "button")

	className := props.ClassName
	if className == "" {
		className = "px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 cursor-pointer transition-colors"
	}

	btn.Set("className", className)
	btn.Set("textContent", props.Text)

	if props.OnClick != nil {
		btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			props.OnClick()
			return nil
		}))
	}

	return btn
}
