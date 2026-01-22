package core

// Renderer defines how nodes are converted to their final form.
// Two implementations exist:
//   - HTMLRenderer: produces HTML strings (server-side)
//   - DOMRenderer: produces js.Value DOM elements (WASM)
type Renderer interface {
	RenderText(content string) RenderResult
	RenderElement(tag string, attrs Attrs, children []Node) RenderResult
	RenderFragment(children []Node) RenderResult
}
