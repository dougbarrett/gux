package ui

import (
	"strings"
	"testing"

	"github.com/dougbarrett/gux/core"
)

// Note: renderHTML helper is defined in button_test.go

func TestCard_Basic(t *testing.T) {
	html := renderHTML(Card(CardProps{}))

	if !strings.Contains(html, "<div") {
		t.Error("Card should render as div")
	}
	if !strings.Contains(html, "bg-white") {
		t.Error("Card should have bg-white class")
	}
	if !strings.Contains(html, "rounded-lg") {
		t.Error("Card should have rounded-lg class")
	}
	if !strings.Contains(html, "shadow") {
		t.Error("Card should have shadow class")
	}
}

func TestCard_CustomClass(t *testing.T) {
	html := renderHTML(Card(CardProps{Class: "my-custom-class"}))

	if !strings.Contains(html, "my-custom-class") {
		t.Error("Card should include custom class")
	}
	// Default classes should still be present
	if !strings.Contains(html, "bg-white") {
		t.Error("Card should retain default bg-white class with custom class")
	}
}

func TestCard_Children(t *testing.T) {
	html := renderHTML(Card(CardProps{
		Children: []core.Node{core.Text("Card content")},
	}))

	if !strings.Contains(html, "Card content") {
		t.Error("Card should render children")
	}
}

func TestCardHeader_Border(t *testing.T) {
	html := renderHTML(CardHeader(CardHeaderProps{}))

	if !strings.Contains(html, "border-b") {
		t.Error("CardHeader should have border-b class")
	}
	if !strings.Contains(html, "px-6") {
		t.Error("CardHeader should have px-6 padding")
	}
	if !strings.Contains(html, "py-4") {
		t.Error("CardHeader should have py-4 padding")
	}
}

func TestCardContent_Padding(t *testing.T) {
	html := renderHTML(CardContent(CardContentProps{}))

	if !strings.Contains(html, "px-6") {
		t.Error("CardContent should have px-6 padding")
	}
	if !strings.Contains(html, "py-4") {
		t.Error("CardContent should have py-4 padding")
	}
}

func TestCardFooter_Border(t *testing.T) {
	html := renderHTML(CardFooter(CardFooterProps{}))

	if !strings.Contains(html, "border-t") {
		t.Error("CardFooter should have border-t class")
	}
	if !strings.Contains(html, "px-6") {
		t.Error("CardFooter should have px-6 padding")
	}
}

func TestCard_Composition(t *testing.T) {
	html := renderHTML(Card(CardProps{
		Children: []core.Node{
			CardHeader(CardHeaderProps{
				Children: []core.Node{core.Text("Header")},
			}),
			CardContent(CardContentProps{
				Children: []core.Node{core.Text("Content")},
			}),
			CardFooter(CardFooterProps{
				Children: []core.Node{core.Text("Footer")},
			}),
		},
	}))

	if !strings.Contains(html, "Header") {
		t.Error("Composed Card should include header text")
	}
	if !strings.Contains(html, "Content") {
		t.Error("Composed Card should include content text")
	}
	if !strings.Contains(html, "Footer") {
		t.Error("Composed Card should include footer text")
	}
	// Verify structure: header, content, footer all within card
	if !strings.Contains(html, "border-b") {
		t.Error("Composed Card should include header border")
	}
	if !strings.Contains(html, "border-t") {
		t.Error("Composed Card should include footer border")
	}
}

func TestCardHeader_CustomClass(t *testing.T) {
	html := renderHTML(CardHeader(CardHeaderProps{Class: "custom-header"}))

	if !strings.Contains(html, "custom-header") {
		t.Error("CardHeader should include custom class")
	}
	if !strings.Contains(html, "border-b") {
		t.Error("CardHeader should retain default border-b with custom class")
	}
}

func TestCardContent_CustomClass(t *testing.T) {
	html := renderHTML(CardContent(CardContentProps{Class: "custom-content"}))

	if !strings.Contains(html, "custom-content") {
		t.Error("CardContent should include custom class")
	}
}

func TestCardFooter_CustomClass(t *testing.T) {
	html := renderHTML(CardFooter(CardFooterProps{Class: "custom-footer"}))

	if !strings.Contains(html, "custom-footer") {
		t.Error("CardFooter should include custom class")
	}
}
