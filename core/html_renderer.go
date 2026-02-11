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
	writeAttr("poster", attrs.Poster)
	writeAttr("preload", attrs.Preload)
	writeAttr("width", attrs.Width)
	writeAttr("height", attrs.Height)

	// Boolean attributes
	if attrs.Autoplay {
		b.WriteString(" autoplay")
	}
	if attrs.Loop {
		b.WriteString(" loop")
	}
	if attrs.Muted {
		b.WriteString(" muted")
	}
	if attrs.Controls {
		b.WriteString(" controls")
	}

	// External link marker (bypass client-side navigation)
	if attrs.External {
		writeAttr("data-gux-external", "true")
	}

	// Preserve scroll position on navigation
	if attrs.PreserveScroll {
		writeAttr("data-gux-preserve-scroll", "true")
	}

	// Data attributes
	for k, v := range attrs.Data {
		writeAttr("data-"+k, v)
	}

	// Extra attributes
	for k, v := range attrs.Extra {
		writeAttr(k, v)
	}

	// Alpine.js directives
	writeAttr("x-data", attrs.XData)
	writeAttr("x-show", attrs.XShow)
	writeAttr("x-model", attrs.XModel)
	writeAttr("x-init", attrs.XInit)
	writeAttr("x-ref", attrs.XRef)
	writeAttr("x-text", attrs.XText)
	writeAttr("x-effect", attrs.XEffect)
	writeAttr("x-html", attrs.XHTML)
	if attrs.XCloak {
		b.WriteString(" x-cloak")
	}
	if attrs.XTransition != "" {
		if attrs.XTransition == "true" {
			b.WriteString(" x-transition")
		} else {
			writeAttr("x-transition", attrs.XTransition)
		}
	}
	for event, expr := range attrs.XOn {
		writeAttr("x-on:"+event, expr)
	}
	for prop, expr := range attrs.XBind {
		writeAttr("x-bind:"+prop, expr)
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
	} else {
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
	}

	// Wrap in <template> for x-if or x-for
	result := b.String()
	if attrs.XIf != "" {
		result = fmt.Sprintf(`<template x-if="%s">%s</template>`, html.EscapeString(attrs.XIf), result)
	}
	if attrs.XFor != "" {
		result = fmt.Sprintf(`<template x-for="%s">%s</template>`, html.EscapeString(attrs.XFor), result)
	}

	return &htmlResult{html: result}
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
