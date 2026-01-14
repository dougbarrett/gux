//go:build js && wasm

package components

import "syscall/js"

// El creates a generic DOM element with optional class and children.
// This is the foundation for building other components.
func El(tag string, className string, children ...js.Value) js.Value {
	document := js.Global().Get("document")
	el := document.Call("createElement", tag)

	if className != "" {
		el.Set("className", className)
	}

	for _, child := range children {
		el.Call("appendChild", child)
	}

	return el
}

// Div creates a div element with optional class and children.
func Div(className string, children ...js.Value) js.Value {
	return El("div", className, children...)
}

// Span creates a span element with optional class and text content.
func Span(className string, text string) js.Value {
	el := El("span", className)
	if text != "" {
		el.Set("textContent", text)
	}
	return el
}
