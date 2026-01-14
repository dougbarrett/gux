//go:build js && wasm

package components

import "syscall/js"

// LoadTailwind injects the Tailwind CDN script and blocks until it's loaded.
func LoadTailwind() {
	// Check if tailwindcss is already loaded
	if !js.Global().Get("tailwind").IsUndefined() {
		return
	}

	document := js.Global().Get("document")
	head := document.Get("head")

	done := make(chan struct{})

	script := document.Call("createElement", "script")
	script.Set("src", "https://cdn.tailwindcss.com")

	// Wait for script to load
	script.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) any {
		close(done)
		return nil
	}))

	head.Call("appendChild", script)

	// Block until loaded
	<-done
}
