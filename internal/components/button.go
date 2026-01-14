//go:build js && wasm

package components

import "syscall/js"

type ButtonProps struct {
	Text      string
	OnClick   func()
	ClassName string
}

const defaultButtonClass = "px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 cursor-pointer mb-4 transition-colors"

func Button(props ButtonProps) js.Value {
	document := js.Global().Get("document")
	button := document.Call("createElement", "button")

	button.Set("textContent", props.Text)

	// Apply Tailwind classes
	className := defaultButtonClass
	if props.ClassName != "" {
		className = props.ClassName
	}
	button.Set("className", className)

	// Attach click handler
	if props.OnClick != nil {
		button.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			props.OnClick()
			return nil
		}))
	}

	return button
}
