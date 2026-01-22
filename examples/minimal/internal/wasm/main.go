//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/pages"
)

func main() {
	document := js.Global().Get("document")
	container := document.Call("getElementById", "app")

	count := 0

	var render func()
	render = func() {
		// Clear and re-render
		container.Set("innerHTML", "")

		node := pages.Home(count, func() {
			count++
			render()
		})

		result := node.Render(core.DOM())
		if domVal := result.DOMValue(); domVal != nil {
			container.Call("appendChild", domVal.(js.Value))
		}
	}

	render()

	// Keep alive
	select {}
}
