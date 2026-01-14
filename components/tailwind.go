//go:build js && wasm

package components

import "syscall/js"

// LoadTailwind injects Tailwind CSS into the document head.
// This blocks until Tailwind is fully loaded to ensure styles are applied.
func LoadTailwind() {
	document := js.Global().Get("document")
	head := document.Get("head")

	script := document.Call("createElement", "script")
	script.Set("src", "https://cdn.tailwindcss.com")

	// Block until loaded
	done := make(chan struct{})
	script.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) any {
		// Configure Tailwind for class-based dark mode after it loads
		js.Global().Get("tailwind").Get("config").Set("darkMode", "class")
		close(done)
		return nil
	}))

	head.Call("appendChild", script)
	<-done
}
