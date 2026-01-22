// Package core provides the fundamental abstractions for gux's universal rendering.
//
// The key insight: components return Nodes, not DOM elements or HTML strings.
// Nodes are an intermediate representation that can be rendered to either target.
package core

// Node represents a renderable element in the UI tree.
// This is the universal currency - components produce Nodes,
// renderers consume them.
type Node interface {
	// Render converts this node to its final form using the given renderer.
	// For HTML: returns a string node containing the markup.
	// For DOM: returns a node wrapping the js.Value element.
	Render(r Renderer) RenderResult
}

// RenderResult is the output of rendering a node.
type RenderResult interface {
	// HTML returns the HTML string representation.
	// Only valid when rendered with an HTML renderer.
	HTML() string

	// DOMValue returns the underlying js.Value.
	// Only valid when rendered with a DOM renderer.
	// Returns nil when rendered server-side.
	DOMValue() any
}

// Attrs holds element attributes.
type Attrs struct {
	ID    string
	Class string
	Href  string
	Src   string
	Alt   string
	Type  string
	Name  string
	Value string

	// External marks a link to bypass client-side navigation.
	// Use this for links that cross bundle boundaries (e.g., admin -> public).
	External bool

	// Data attributes (rendered as data-*)
	Data map[string]string

	// Event handlers - only active in WASM, ignored in SSR
	OnClick  func()
	OnSubmit func()
	OnChange func(value string)
	OnEnter  func() // Called when Enter key is pressed

	// Additional attributes not covered above
	Extra map[string]string
}

// TextNode represents plain text content.
type TextNode struct {
	Content string
}

func (t TextNode) Render(r Renderer) RenderResult {
	return r.RenderText(t.Content)
}

// Element represents an HTML element with tag, attributes, and children.
type Element struct {
	Tag      string
	Attrs    Attrs
	Children []Node
}

func (e Element) Render(r Renderer) RenderResult {
	return r.RenderElement(e.Tag, e.Attrs, e.Children)
}

// Fragment represents multiple nodes without a wrapper element.
type Fragment struct {
	Children []Node
}

func (f Fragment) Render(r Renderer) RenderResult {
	return r.RenderFragment(f.Children)
}
