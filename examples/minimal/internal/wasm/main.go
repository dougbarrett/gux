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

	var router *core.Router
	var render func()

	render = func() {
		container.Set("innerHTML", "")

		// Get current path and route to appropriate page
		// For now, just use Home
		component := pages.Home(router)

		node := component()
		result := node.Render(core.DOM())
		if domVal := result.DOMValue(); domVal != nil {
			container.Call("appendChild", domVal.(js.Value))
		}
	}

	router = core.NewRouter(render)
	render()

	select {}
}
