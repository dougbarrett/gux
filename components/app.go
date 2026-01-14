//go:build js && wasm

package components

import "syscall/js"

// App provides helpers for initializing a WASM application.
type App struct {
	root js.Value
}

// NewApp initializes the application, loads Tailwind, and returns an App instance.
// It clears the target element and prepares it for rendering.
func NewApp(elementID string) *App {
	// Load Tailwind CSS (blocks until loaded)
	LoadTailwind()

	document := js.Global().Get("document")

	// Reset body styles for full-page layout
	body := document.Get("body")
	body.Set("className", "m-0 p-0")

	// Get and clear the app container
	root := document.Call("getElementById", elementID)
	root.Set("innerHTML", "")

	return &App{root: root}
}

// Root returns the root DOM element.
func (a *App) Root() js.Value {
	return a.root
}

// Mount appends an element to the root.
func (a *App) Mount(el js.Value) {
	a.root.Call("appendChild", el)
}

// Run blocks forever, keeping the WASM instance alive.
// Call this at the end of your main function.
func (a *App) Run() {
	select {}
}

// Document returns the global document object.
func Document() js.Value {
	return js.Global().Get("document")
}
