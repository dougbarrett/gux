//go:build js && wasm

package components

import "syscall/js"

type LinkProps struct {
	To        string
	Text      string
	ClassName string
	Children  js.Value // Optional: use custom element instead of text
}

func Link(props LinkProps) js.Value {
	document := js.Global().Get("document")

	a := document.Call("createElement", "a")
	a.Set("href", props.To)

	if props.ClassName != "" {
		a.Set("className", props.ClassName)
	}

	// Use children if provided, otherwise use text
	if !props.Children.IsUndefined() && !props.Children.IsNull() {
		a.Call("appendChild", props.Children)
	} else {
		a.Set("textContent", props.Text)
	}

	// Handle click - prevent default and use router
	a.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		args[0].Call("preventDefault")
		if router := GetRouter(); router != nil {
			router.Navigate(props.To)
		}
		return nil
	}))

	return a
}
