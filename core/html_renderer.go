package core

import (
	"fmt"
	"html"
	"strings"
)

// HTMLRenderer renders nodes to HTML strings.
// Used for server-side rendering.
type HTMLRenderer struct{}

// HTML returns a new HTMLRenderer.
func HTML() *HTMLRenderer {
	return &HTMLRenderer{}
}

func (r *HTMLRenderer) RenderText(content string) RenderResult {
	return &htmlResult{html: html.EscapeString(content)}
}

func (r *HTMLRenderer) RenderElement(tag string, attrs Attrs, children []Node) RenderResult {
	var b strings.Builder

	b.WriteString("<")
	b.WriteString(tag)

	// Write attributes
	writeAttr := func(name, value string) {
		if value != "" {
			b.WriteString(fmt.Sprintf(` %s="%s"`, name, html.EscapeString(value)))
		}
	}

	writeAttr("id", attrs.ID)
	writeAttr("class", attrs.Class)
	writeAttr("href", attrs.Href)
	writeAttr("src", attrs.Src)
	writeAttr("alt", attrs.Alt)
	writeAttr("type", attrs.Type)
	writeAttr("name", attrs.Name)
	writeAttr("value", attrs.Value)

	// External link marker (bypass client-side navigation)
	if attrs.External {
		writeAttr("data-gux-external", "true")
	}

	// Data attributes
	for k, v := range attrs.Data {
		writeAttr("data-"+k, v)
	}

	// Extra attributes
	for k, v := range attrs.Extra {
		writeAttr(k, v)
	}

	// Self-closing tags
	selfClosing := map[string]bool{
		"area": true, "base": true, "br": true, "col": true,
		"embed": true, "hr": true, "img": true, "input": true,
		"link": true, "meta": true, "source": true, "track": true,
		"wbr": true,
	}

	if selfClosing[tag] {
		b.WriteString(" />")
		return &htmlResult{html: b.String()}
	}

	b.WriteString(">")

	// Render children
	for _, child := range children {
		if child == nil {
			continue
		}
		result := child.Render(r)
		b.WriteString(result.HTML())
	}

	b.WriteString("</")
	b.WriteString(tag)
	b.WriteString(">")

	return &htmlResult{html: b.String()}
}

func (r *HTMLRenderer) RenderFragment(children []Node) RenderResult {
	var b strings.Builder
	for _, child := range children {
		if child == nil {
			continue
		}
		result := child.Render(r)
		b.WriteString(result.HTML())
	}
	return &htmlResult{html: b.String()}
}

func (r *HTMLRenderer) RenderRawHTML(content string) RenderResult {
	return &htmlResult{html: content}
}

// htmlResult holds the result of HTML rendering.
type htmlResult struct {
	html string
}

func (r *htmlResult) HTML() string {
	return r.html
}

func (r *htmlResult) DOMValue() any {
	return nil
}
