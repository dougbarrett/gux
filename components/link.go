//go:build js && wasm

package components

import "syscall/js"

// LinkProps configures a Link component
type LinkProps struct {
	To        string
	ClassName string
	Children  func(parent js.Value)
}

// Link creates a client-side navigation link
func Link(props LinkProps) js.Value {
	document := js.Global().Get("document")
	a := document.Call("createElement", "a")

	a.Set("href", props.To)
	if props.ClassName != "" {
		a.Set("className", props.ClassName)
	}

	// Prevent default navigation, use router instead
	a.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		args[0].Call("preventDefault")
		if globalRouter != nil {
			globalRouter.Navigate(props.To)
		}
		return nil
	}))

	if props.Children != nil {
		props.Children(a)
	}

	return a
}
