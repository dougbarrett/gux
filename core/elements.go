package core

// Helper functions for creating nodes.
// These provide a clean DSL for building UI.

// Text creates a text node.
func Text(content string) Node {
	return TextNode{Content: content}
}

// El creates an element with the given tag, attributes, and children.
func El(tag string, attrs Attrs, children ...Node) Node {
	return Element{Tag: tag, Attrs: attrs, Children: children}
}

// Frag creates a fragment containing multiple nodes.
func Frag(children ...Node) Node {
	return Fragment{Children: children}
}

// Convenience functions for common elements

func Div(attrs Attrs, children ...Node) Node {
	return El("div", attrs, children...)
}

func Span(attrs Attrs, children ...Node) Node {
	return El("span", attrs, children...)
}

func P(attrs Attrs, children ...Node) Node {
	return El("p", attrs, children...)
}

func H1(attrs Attrs, children ...Node) Node {
	return El("h1", attrs, children...)
}

func H2(attrs Attrs, children ...Node) Node {
	return El("h2", attrs, children...)
}

func H3(attrs Attrs, children ...Node) Node {
	return El("h3", attrs, children...)
}

func A(attrs Attrs, children ...Node) Node {
	return El("a", attrs, children...)
}

func Button(attrs Attrs, children ...Node) Node {
	return El("button", attrs, children...)
}

func Input(attrs Attrs) Node {
	return El("input", attrs)
}

func Form(attrs Attrs, children ...Node) Node {
	return El("form", attrs, children...)
}

func Label(attrs Attrs, children ...Node) Node {
	return El("label", attrs, children...)
}

func Ul(attrs Attrs, children ...Node) Node {
	return El("ul", attrs, children...)
}

func Li(attrs Attrs, children ...Node) Node {
	return El("li", attrs, children...)
}

func Nav(attrs Attrs, children ...Node) Node {
	return El("nav", attrs, children...)
}

func Header(attrs Attrs, children ...Node) Node {
	return El("header", attrs, children...)
}

func Footer(attrs Attrs, children ...Node) Node {
	return El("footer", attrs, children...)
}

func Main(attrs Attrs, children ...Node) Node {
	return El("main", attrs, children...)
}

func Section(attrs Attrs, children ...Node) Node {
	return El("section", attrs, children...)
}

func Article(attrs Attrs, children ...Node) Node {
	return El("article", attrs, children...)
}

func Img(attrs Attrs) Node {
	return El("img", attrs)
}

func Table(attrs Attrs, children ...Node) Node {
	return El("table", attrs, children...)
}

func Thead(attrs Attrs, children ...Node) Node {
	return El("thead", attrs, children...)
}

func Tbody(attrs Attrs, children ...Node) Node {
	return El("tbody", attrs, children...)
}

func Tr(attrs Attrs, children ...Node) Node {
	return El("tr", attrs, children...)
}

func Th(attrs Attrs, children ...Node) Node {
	return El("th", attrs, children...)
}

func Td(attrs Attrs, children ...Node) Node {
	return El("td", attrs, children...)
}

// Class is a shorthand for creating Attrs with just a class.
func Class(class string) Attrs {
	return Attrs{Class: class}
}

func Select(attrs Attrs, children ...Node) Node {
	return El("select", attrs, children...)
}

func Option(attrs Attrs, children ...Node) Node {
	return El("option", attrs, children...)
}

func Textarea(attrs Attrs, children ...Node) Node {
	return El("textarea", attrs, children...)
}

func Video(attrs Attrs, children ...Node) Node {
	return El("video", attrs, children...)
}

func Audio(attrs Attrs, children ...Node) Node {
	return El("audio", attrs, children...)
}

func Source(attrs Attrs) Node {
	return El("source", attrs)
}

func Track(attrs Attrs) Node {
	return El("track", attrs)
}

func Iframe(attrs Attrs, children ...Node) Node {
	return El("iframe", attrs, children...)
}

func Canvas(attrs Attrs, children ...Node) Node {
	return El("canvas", attrs, children...)
}

// If conditionally renders a node based on a condition.
func If(condition bool, node Node) Node {
	if condition {
		return node
	}
	return Frag() // Empty fragment
}
