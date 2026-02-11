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

	// PreserveScroll prevents automatic scroll-to-top when navigating via this link.
	// By default, client-side navigation scrolls to the top of the page.
	PreserveScroll bool

	// Media element attributes
	Poster   string // Video poster image
	Autoplay bool   // Auto-play media
	Loop     bool   // Loop playback
	Muted    bool   // Start muted
	Controls bool   // Show playback controls
	Preload  string // "none", "metadata", "auto"
	Width    string // Element width
	Height   string // Element height

	// Data attributes (rendered as data-*)
	Data map[string]string

	// Event handlers - only active in WASM, ignored in SSR
	OnClick  func()
	OnSubmit func()
	OnChange func(value string)
	OnEnter  func() // Called when Enter key is pressed

	// OnMount is called after the element's DOM node is created and children
	// are appended (WASM only, ignored in SSR). The callback receives the DOM
	// element as interface{} which can be cast to syscall/js.Value in WASM.
	// Called on every render since gux replaces DOM trees on re-render.
	// JS library initialization should be idempotent.
	OnMount func(el any)

	// OnUnmount is called before the element's DOM subtree is replaced
	// (WASM only, ignored in SSR). Use for cleanup: disposing JS library
	// instances, removing global event listeners, etc.
	OnUnmount func()

	// Alpine.js directives
	XData       string            // x-data="{open: false}"
	XShow       string            // x-show="open"
	XModel      string            // x-model="email"
	XInit       string            // x-init="..."
	XRef        string            // x-ref="inputEl"
	XText       string            // x-text="count"
	XCloak      bool              // x-cloak (boolean attr, no value)
	XTransition string            // x-transition or x-transition.duration.500ms
	XOn         map[string]string // "click" → "open = true", "submit.prevent" → "handle()"
	XBind       map[string]string // "class" → "{ active: isActive }", "disabled" → "loading"
	XIf         string            // x-if (wraps in <template>)
	XFor        string            // x-for="item in items" (wraps in <template>)
	XEffect     string            // x-effect="..."
	XHTML       string            // x-html="rawContent"

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

// RawHTMLNode renders pre-built HTML/SVG markup directly without escaping.
// Use only with trusted content (e.g., chart library SVG output).
type RawHTMLNode struct {
	Content string
}

func (n RawHTMLNode) Render(r Renderer) RenderResult {
	return r.RenderRawHTML(n.Content)
}

// RawHTML creates a node that renders raw HTML content without escaping.
// WARNING: Content is not sanitized. Only use with trusted content
// such as SVG output from charting libraries.
func RawHTML(content string) Node {
	return RawHTMLNode{Content: content}
}
