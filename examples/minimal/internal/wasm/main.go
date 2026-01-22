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

	var router *pages.Router
	var render func()

	render = func() {
		container.Set("innerHTML", "")

		// Get the component function from the page
		component := router.Home()

		// Render the component
		node := component()
		result := node.Render(core.DOM())
		if domVal := result.DOMValue(); domVal != nil {
			container.Call("appendChild", domVal.(js.Value))
		}
	}

	router = pages.NewRouter(render)
	render()

	select {}
}
