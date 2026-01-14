//go:build js && wasm

package components

import "syscall/js"

// TooltipPosition defines where the tooltip appears
type TooltipPosition string

const (
	TooltipTop    TooltipPosition = "top"
	TooltipBottom TooltipPosition = "bottom"
	TooltipLeft   TooltipPosition = "left"
	TooltipRight  TooltipPosition = "right"
)

// TooltipProps configures a Tooltip
type TooltipProps struct {
	Text     string
	Position TooltipPosition
	Delay    int // milliseconds, default 200
}

// WithTooltip wraps an element with a tooltip
func WithTooltip(element js.Value, props TooltipProps) js.Value {
	document := js.Global().Get("document")

	// Wrapper with relative positioning
	wrapper := document.Call("createElement", "div")
	wrapper.Set("className", "relative inline-block")

	// Tooltip element (hidden by default)
	tooltip := document.Call("createElement", "div")

	position := props.Position
	if position == "" {
		position = TooltipTop
	}

	baseClass := "absolute z-50 px-2 py-1 text-sm text-white bg-gray-900 rounded shadow-lg whitespace-nowrap opacity-0 invisible transition-all duration-200 pointer-events-none"

	var positionClass string
	switch position {
	case TooltipTop:
		positionClass = "bottom-full left-1/2 -translate-x-1/2 mb-2"
	case TooltipBottom:
		positionClass = "top-full left-1/2 -translate-x-1/2 mt-2"
	case TooltipLeft:
		positionClass = "right-full top-1/2 -translate-y-1/2 mr-2"
	case TooltipRight:
		positionClass = "left-full top-1/2 -translate-y-1/2 ml-2"
	}

	tooltip.Set("className", baseClass+" "+positionClass)
	tooltip.Set("textContent", props.Text)

	delay := props.Delay
	if delay == 0 {
		delay = 200
	}

	var timeoutID js.Value

	// Show on hover
	element.Call("addEventListener", "mouseenter", js.FuncOf(func(this js.Value, args []js.Value) any {
		timeoutID = js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
			tooltip.Get("classList").Call("remove", "opacity-0", "invisible")
			tooltip.Get("classList").Call("add", "opacity-100", "visible")
			return nil
		}), delay)
		return nil
	}))

	// Hide on leave
	element.Call("addEventListener", "mouseleave", js.FuncOf(func(this js.Value, args []js.Value) any {
		js.Global().Call("clearTimeout", timeoutID)
		tooltip.Get("classList").Call("remove", "opacity-100", "visible")
		tooltip.Get("classList").Call("add", "opacity-0", "invisible")
		return nil
	}))

	wrapper.Call("appendChild", element)
	wrapper.Call("appendChild", tooltip)

	return wrapper
}

// Tooltip creates a simple tooltip-wrapped text span
func Tooltip(text, tooltip string) js.Value {
	document := js.Global().Get("document")
	span := document.Call("createElement", "span")
	span.Set("textContent", text)
	span.Set("className", "border-b border-dotted border-gray-400 cursor-help")
	return WithTooltip(span, TooltipProps{Text: tooltip})
}
