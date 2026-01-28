//go:build js && wasm

package core

import (
	"syscall/js"
)

var document = js.Global().Get("document")

// SVG namespace URI
const svgNamespace = "http://www.w3.org/2000/svg"

// svgElements contains tags that should be created in the SVG namespace
var svgElements = map[string]bool{
	"svg":            true,
	"path":           true,
	"circle":         true,
	"rect":           true,
	"line":           true,
	"polygon":        true,
	"polyline":       true,
	"ellipse":        true,
	"g":              true,
	"defs":           true,
	"use":            true,
	"text":           true,
	"tspan":          true,
	"clipPath":       true,
	"mask":           true,
	"linearGradient": true,
	"radialGradient": true,
	"stop":           true,
	"symbol":         true,
	"marker":         true,
	"pattern":        true,
	"filter":         true,
	"feBlend":        true,
	"feColorMatrix":  true,
	"feGaussianBlur": true,
	"feOffset":       true,
	"feMerge":        true,
	"feMergeNode":    true,
	"animate":        true,
	"animateMotion":  true,
	"animateTransform": true,
	"foreignObject":  true,
}

// DOMRenderer renders nodes to DOM elements.
// Used for client-side WASM rendering.
type DOMRenderer struct {
	inSVG bool // Track if we're inside an SVG element
}

// DOM returns a new DOMRenderer.
func DOM() *DOMRenderer {
	return &DOMRenderer{inSVG: false}
}

// withSVGContext returns a new renderer that's in SVG context
func (r *DOMRenderer) withSVGContext() *DOMRenderer {
	return &DOMRenderer{inSVG: true}
}

func (r *DOMRenderer) RenderText(content string) RenderResult {
	textNode := document.Call("createTextNode", content)
	return &domResult{value: textNode}
}

func (r *DOMRenderer) RenderElement(tag string, attrs Attrs, children []Node) RenderResult {
	// Debug: track all tags
	js.Global().Set("__lastTag", tag)

	// Determine if this element should be in SVG namespace
	isSVGElement := svgElements[tag] || r.inSVG

	var el js.Value
	if isSVGElement {
		js.Global().Set("__svgCreated", tag)
		el = document.Call("createElementNS", svgNamespace, tag)
	} else {
		el = document.Call("createElement", tag)
	}

	// Set attributes
	setAttr := func(name, value string) {
		// For value attribute, always set it (including empty string for placeholder options)
		// For other attributes, skip empty strings
		if value != "" || name == "value" {
			el.Call("setAttribute", name, value)
		}
	}

	setAttr("id", attrs.ID)
	setAttr("class", attrs.Class)
	setAttr("href", attrs.Href)
	setAttr("src", attrs.Src)
	setAttr("alt", attrs.Alt)
	setAttr("type", attrs.Type)
	setAttr("name", attrs.Name)
	setAttr("value", attrs.Value)

	// External link marker (bypass client-side navigation)
	if attrs.External {
		setAttr("data-gux-external", "true")
	}

	// Data attributes
	for k, v := range attrs.Data {
		setAttr("data-"+k, v)
	}

	// Extra attributes
	for k, v := range attrs.Extra {
		setAttr(k, v)
	}

	// Event handlers
	if attrs.OnClick != nil {
		cb := js.FuncOf(func(this js.Value, args []js.Value) any {
			attrs.OnClick()
			return nil
		})
		el.Call("addEventListener", "click", cb)
	}

	if attrs.OnSubmit != nil {
		cb := js.FuncOf(func(this js.Value, args []js.Value) any {
			args[0].Call("preventDefault")
			attrs.OnSubmit()
			return nil
		})
		el.Call("addEventListener", "submit", cb)
	}

	if attrs.OnChange != nil {
		cb := js.FuncOf(func(this js.Value, args []js.Value) any {
			value := args[0].Get("target").Get("value").String()
			// Suppress re-renders during change events.
			// This allows button clicks to work - the state is updated but
			// the DOM isn't replaced, so the click event completes normally.
			// The re-render will happen after the click handler's side effects
			// (like navigation or showing success messages).
			SetInChangeEvent(true)
			attrs.OnChange(value)
			SetInChangeEvent(false)
			return nil
		})
		// Only listen to "change" event which fires on blur.
		el.Call("addEventListener", "change", cb)
	}

	if attrs.OnEnter != nil {
		cb := js.FuncOf(func(this js.Value, args []js.Value) any {
			event := args[0]
			if event.Get("key").String() == "Enter" {
				event.Call("preventDefault")
				// Get the current value and trigger change before calling OnEnter
				value := event.Get("target").Get("value").String()
				if attrs.OnChange != nil {
					attrs.OnChange(value)
				}
				attrs.OnEnter()
			}
			return nil
		})
		el.Call("addEventListener", "keydown", cb)
	}

	// Render and append children
	// Use SVG context for children if we're inside an SVG element
	childRenderer := r
	if isSVGElement && !r.inSVG {
		childRenderer = r.withSVGContext()
	}

	for _, child := range children {
		if child == nil {
			continue
		}
		result := child.Render(childRenderer)
		if domVal := result.DOMValue(); domVal != nil {
			el.Call("appendChild", domVal.(js.Value))
		}
	}

	return &domResult{value: el}
}

func (r *DOMRenderer) RenderFragment(children []Node) RenderResult {
	fragment := document.Call("createDocumentFragment")
	for _, child := range children {
		if child == nil {
			continue
		}
		result := child.Render(r)
		if domVal := result.DOMValue(); domVal != nil {
			fragment.Call("appendChild", domVal.(js.Value))
		}
	}
	return &domResult{value: fragment}
}

// domResult holds the result of DOM rendering.
type domResult struct {
	value js.Value
}

func (r *domResult) HTML() string {
	return ""
}

func (r *domResult) DOMValue() any {
	return r.value
}
