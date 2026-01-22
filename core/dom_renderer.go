//go:build js && wasm

package core

import (
	"syscall/js"
)

var document = js.Global().Get("document")

// DOMRenderer renders nodes to DOM elements.
// Used for client-side WASM rendering.
type DOMRenderer struct{}

// DOM returns a new DOMRenderer.
func DOM() *DOMRenderer {
	return &DOMRenderer{}
}

func (r *DOMRenderer) RenderText(content string) RenderResult {
	textNode := document.Call("createTextNode", content)
	return &domResult{value: textNode}
}

func (r *DOMRenderer) RenderElement(tag string, attrs Attrs, children []Node) RenderResult {
	el := document.Call("createElement", tag)

	// Set attributes
	setAttr := func(name, value string) {
		if value != "" {
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
			attrs.OnChange(value)
			return nil
		})
		// Only listen to "change" event which fires on blur.
		// State updates happen when user leaves the field, not on every keystroke.
		// Combined with deferred re-renders (via ScheduleRerender), this prevents
		// focus loss and allows button clicks to work properly.
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
	for _, child := range children {
		result := child.Render(r)
		if domVal := result.DOMValue(); domVal != nil {
			el.Call("appendChild", domVal.(js.Value))
		}
	}

	return &domResult{value: el}
}

func (r *DOMRenderer) RenderFragment(children []Node) RenderResult {
	fragment := document.Call("createDocumentFragment")
	for _, child := range children {
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
