package core

import (
	"context"
	"fmt"
	"html"
	"io"
	"strings"
)

// NodeComponent wraps a Node as a templ.Component for use in templ templates.
// Use AsComponent() to convert any Node tree into a streamable component.
type NodeComponent struct {
	Node Node
}

// Render implements the templ.Component interface, streaming the Node tree
// directly to an io.Writer without building an intermediate string.
func (nc NodeComponent) Render(ctx context.Context, w io.Writer) error {
	return renderNodeToWriter(ctx, w, nc.Node)
}

// AsComponent converts any Node to a templ.Component.
func AsComponent(node Node) NodeComponent {
	return NodeComponent{Node: node}
}

// renderNodeToWriter dispatches to type-specific streaming renderers.
func renderNodeToWriter(ctx context.Context, w io.Writer, node Node) error {
	switch n := node.(type) {
	case Element:
		return renderElementToWriter(ctx, w, n.Tag, n.Attrs, n.Children)
	case TextNode:
		_, err := io.WriteString(w, html.EscapeString(n.Content))
		return err
	case Fragment:
		for _, child := range n.Children {
			if child == nil {
				continue
			}
			if err := renderNodeToWriter(ctx, w, child); err != nil {
				return err
			}
		}
		return nil
	case RawHTMLNode:
		_, err := io.WriteString(w, n.Content)
		return err
	default:
		// Fall back to HTMLRenderer for unknown node types
		result := node.Render(HTML())
		_, err := io.WriteString(w, result.HTML())
		return err
	}
}

// selfClosingTags for streaming renderer
var selfClosingTags = map[string]bool{
	"area": true, "base": true, "br": true, "col": true,
	"embed": true, "hr": true, "img": true, "input": true,
	"link": true, "meta": true, "source": true, "track": true,
	"wbr": true,
}

func renderElementToWriter(ctx context.Context, w io.Writer, tag string, attrs Attrs, children []Node) error {
	var b strings.Builder

	b.WriteString("<")
	b.WriteString(tag)

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
	if attrs.External {
		writeAttr("data-gux-external", "true")
	}
	if attrs.PreserveScroll {
		writeAttr("data-gux-preserve-scroll", "true")
	}

	for k, v := range attrs.Data {
		writeAttr("data-"+k, v)
	}
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

	if selfClosingTags[tag] {
		b.WriteString(" />")
	} else {
		b.WriteString(">")
	}

	// Wrap in <template> for x-if/x-for
	needsTemplateWrap := attrs.XIf != "" || attrs.XFor != ""
	if needsTemplateWrap {
		if attrs.XFor != "" {
			fmt.Fprintf(w, `<template x-for="%s">`, html.EscapeString(attrs.XFor))
		}
		if attrs.XIf != "" {
			fmt.Fprintf(w, `<template x-if="%s">`, html.EscapeString(attrs.XIf))
		}
	}

	if _, err := io.WriteString(w, b.String()); err != nil {
		return err
	}

	if !selfClosingTags[tag] {
		for _, child := range children {
			if child == nil {
				continue
			}
			if err := renderNodeToWriter(ctx, w, child); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "</%s>", tag); err != nil {
			return err
		}
	}

	if needsTemplateWrap {
		if attrs.XIf != "" {
			io.WriteString(w, "</template>")
		}
		if attrs.XFor != "" {
			io.WriteString(w, "</template>")
		}
	}

	return nil
}
