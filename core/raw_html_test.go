package core

import (
	"strings"
	"testing"
)

func TestRawHTML_Passthrough(t *testing.T) {
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><rect width="50" height="50"/></svg>`
	node := RawHTML(svg)
	result := node.Render(HTML())

	if result.HTML() != svg {
		t.Errorf("RawHTML should pass content through unchanged.\ngot:  %s\nwant: %s", result.HTML(), svg)
	}
}

func TestRawHTML_NoEscaping(t *testing.T) {
	content := `<div class="test">&amp; <b>bold</b></div>`
	node := RawHTML(content)
	result := node.Render(HTML())

	// Should NOT double-escape the &amp;
	if strings.Contains(result.HTML(), "&amp;amp;") {
		t.Errorf("RawHTML should not escape content, got: %s", result.HTML())
	}
	if result.HTML() != content {
		t.Errorf("RawHTML content mismatch.\ngot:  %s\nwant: %s", result.HTML(), content)
	}
}

func TestRawHTML_EmptyContent(t *testing.T) {
	node := RawHTML("")
	result := node.Render(HTML())

	if result.HTML() != "" {
		t.Errorf("empty RawHTML should produce empty string, got: %s", result.HTML())
	}
}

func TestRawHTML_AsChild(t *testing.T) {
	svg := `<svg><circle cx="10" cy="10" r="5"/></svg>`
	wrapper := El("div", Attrs{Class: "chart-wrapper"}, RawHTML(svg))
	result := wrapper.Render(HTML())
	html := result.HTML()

	if !strings.Contains(html, `<div class="chart-wrapper">`) {
		t.Errorf("expected wrapper div, got: %s", html)
	}
	if !strings.Contains(html, svg) {
		t.Errorf("expected raw SVG inside wrapper, got: %s", html)
	}
}

func TestRawHTML_InFragment(t *testing.T) {
	svg := `<svg><rect/></svg>`
	frag := Frag(Text("before"), RawHTML(svg), Text("after"))
	result := frag.Render(HTML())
	html := result.HTML()

	if !strings.Contains(html, "before") || !strings.Contains(html, svg) || !strings.Contains(html, "after") {
		t.Errorf("fragment should contain all parts, got: %s", html)
	}
}

func TestRawHTML_DOMValueNil(t *testing.T) {
	node := RawHTML("<svg/>")
	result := node.Render(HTML())

	// HTML renderer should return nil for DOMValue
	if result.DOMValue() != nil {
		t.Errorf("HTML renderer DOMValue should be nil")
	}
}
