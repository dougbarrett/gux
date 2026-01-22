//go:build js && wasm

// WASM entry point - renders pages to DOM.
package main

import (
	"syscall/js"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/pages"
)

func main() {
	// Get the app container
	document := js.Global().Get("document")
	container := document.Call("getElementById", "app")

	// Clear existing content
	container.Set("innerHTML", "")

	// Render the page to DOM
	renderer := core.DOM()
	result := pages.Home().Render(renderer)

	// Mount to container
	if domVal := result.DOMValue(); domVal != nil {
		container.Call("appendChild", domVal.(js.Value))
	}

	// Keep WASM alive
	select {}
}
