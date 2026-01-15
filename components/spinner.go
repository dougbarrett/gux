//go:build js && wasm

package components

import "syscall/js"

// SpinnerSize defines spinner sizes
type SpinnerSize string

const (
	SpinnerSM SpinnerSize = "sm"
	SpinnerMD SpinnerSize = "md"
	SpinnerLG SpinnerSize = "lg"
)

var spinnerSizes = map[SpinnerSize]string{
	SpinnerSM: "h-4 w-4 border-2",
	SpinnerMD: "h-8 w-8 border-2",
	SpinnerLG: "h-12 w-12 border-4",
}

// SpinnerProps configures a Spinner component
type SpinnerProps struct {
	Size      SpinnerSize
	Color     string // Tailwind color class like "blue-500"
	Label     string // visible text label below spinner
	AriaLabel string // accessible label (defaults to "Loading" if empty)
}

// Spinner creates a loading spinner animation
func Spinner(props SpinnerProps) js.Value {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "flex flex-col items-center justify-center")
	// ARIA live region for loading status
	container.Call("setAttribute", "role", "status")
	container.Call("setAttribute", "aria-live", "polite")
	// Default to "Loading" if no custom label provided
	ariaLabel := props.AriaLabel
	if ariaLabel == "" {
		ariaLabel = "Loading"
	}
	container.Call("setAttribute", "aria-label", ariaLabel)

	size := props.Size
	if size == "" {
		size = SpinnerMD
	}

	color := props.Color
	if color == "" {
		color = "blue-500"
	}

	sizeClass := spinnerSizes[size]
	if sizeClass == "" {
		sizeClass = spinnerSizes[SpinnerMD]
	}

	// Visual spinner (decorative - aria-hidden since container has label)
	spinner := document.Call("createElement", "div")
	spinner.Set("className", sizeClass+" border-gray-200 border-t-"+color+" rounded-full animate-spin")
	spinner.Call("setAttribute", "aria-hidden", "true")
	container.Call("appendChild", spinner)

	// Add CSS animation if not exists
	addSpinnerStyles()

	if props.Label != "" {
		label := document.Call("createElement", "p")
		label.Set("className", "mt-2 text-sm text-gray-600")
		label.Set("textContent", props.Label)
		container.Call("appendChild", label)
	}

	return container
}

// SpinnerInline creates an inline spinner (no container)
func SpinnerInline(size SpinnerSize, color string) js.Value {
	document := js.Global().Get("document")

	if size == "" {
		size = SpinnerSM
	}
	if color == "" {
		color = "blue-500"
	}

	sizeClass := spinnerSizes[size]
	if sizeClass == "" {
		sizeClass = spinnerSizes[SpinnerSM]
	}

	spinner := document.Call("createElement", "div")
	spinner.Set("className", sizeClass+" border-gray-200 border-t-"+color+" rounded-full animate-spin inline-block")
	// Inline spinner has role and label directly on the element
	spinner.Call("setAttribute", "role", "status")
	spinner.Call("setAttribute", "aria-label", "Loading")

	addSpinnerStyles()

	return spinner
}

var spinnerStylesAdded = false

func addSpinnerStyles() {
	if spinnerStylesAdded {
		return
	}

	document := js.Global().Get("document")
	style := document.Call("createElement", "style")
	style.Set("textContent", `
		@keyframes spin {
			to { transform: rotate(360deg); }
		}
		.animate-spin {
			animation: spin 1s linear infinite;
		}
	`)
	document.Get("head").Call("appendChild", style)
	spinnerStylesAdded = true
}
