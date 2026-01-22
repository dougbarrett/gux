package ui

import (
	"strings"
	"testing"

	"github.com/dougbarrett/gux/core"
)

// renderHTML renders a node to HTML string for testing.
func renderHTML(node core.Node) string {
	renderer := core.HTML()
	result := node.Render(renderer)
	return result.HTML()
}

func TestButton_DefaultVariant(t *testing.T) {
	btn := Button(ButtonProps{
		Children: []core.Node{core.Text("Click")},
	})
	html := renderHTML(btn)

	// Should render as button element
	if !strings.HasPrefix(html, "<button") {
		t.Errorf("expected <button> element, got: %s", html)
	}

	// Should have primary variant classes (default)
	if !strings.Contains(html, "bg-blue-600") {
		t.Errorf("expected primary variant classes, got: %s", html)
	}

	// Should have medium size classes (default)
	if !strings.Contains(html, "px-4 py-2") {
		t.Errorf("expected medium size classes, got: %s", html)
	}

	// Should have default type="button"
	if !strings.Contains(html, `type="button"`) {
		t.Errorf("expected type=\"button\", got: %s", html)
	}
}

func TestButton_AllVariants(t *testing.T) {
	tests := []struct {
		variant       ButtonVariant
		expectedClass string
	}{
		{ButtonPrimary, "bg-blue-600"},
		{ButtonSecondary, "bg-gray-200"},
		{ButtonOutline, "border border-gray-300"},
		{ButtonGhost, "bg-transparent"},
		{ButtonDestructive, "bg-red-600"},
	}

	for _, tt := range tests {
		t.Run(string(tt.variant), func(t *testing.T) {
			btn := Button(ButtonProps{
				Variant:  tt.variant,
				Children: []core.Node{core.Text("Button")},
			})
			html := renderHTML(btn)

			if !strings.Contains(html, tt.expectedClass) {
				t.Errorf("variant %s: expected class %q in: %s", tt.variant, tt.expectedClass, html)
			}
		})
	}
}

func TestButton_AllSizes(t *testing.T) {
	tests := []struct {
		size          ButtonSize
		expectedClass string
	}{
		{ButtonSM, "px-3 py-1.5 text-sm"},
		{ButtonMD, "px-4 py-2 text-base"},
		{ButtonLG, "px-6 py-3 text-lg"},
	}

	for _, tt := range tests {
		t.Run(string(tt.size), func(t *testing.T) {
			btn := Button(ButtonProps{
				Size:     tt.size,
				Children: []core.Node{core.Text("Button")},
			})
			html := renderHTML(btn)

			if !strings.Contains(html, tt.expectedClass) {
				t.Errorf("size %s: expected class %q in: %s", tt.size, tt.expectedClass, html)
			}
		})
	}
}

func TestButton_Disabled(t *testing.T) {
	btn := Button(ButtonProps{
		Disabled: true,
		Children: []core.Node{core.Text("Disabled")},
	})
	html := renderHTML(btn)

	// Should have disabled attribute
	if !strings.Contains(html, `disabled="disabled"`) {
		t.Errorf("expected disabled attribute, got: %s", html)
	}

	// Should have disabled styling classes
	if !strings.Contains(html, "opacity-50") {
		t.Errorf("expected opacity-50 class, got: %s", html)
	}
	if !strings.Contains(html, "cursor-not-allowed") {
		t.Errorf("expected cursor-not-allowed class, got: %s", html)
	}
}

func TestButton_CustomClass(t *testing.T) {
	btn := Button(ButtonProps{
		Class:    "my-custom-class w-full",
		Children: []core.Node{core.Text("Custom")},
	})
	html := renderHTML(btn)

	// Should include custom classes
	if !strings.Contains(html, "my-custom-class") {
		t.Errorf("expected custom class, got: %s", html)
	}
	if !strings.Contains(html, "w-full") {
		t.Errorf("expected w-full class, got: %s", html)
	}

	// Should still have base classes
	if !strings.Contains(html, "font-medium") {
		t.Errorf("expected base classes, got: %s", html)
	}
}

func TestButton_Children(t *testing.T) {
	btn := Button(ButtonProps{
		Children: []core.Node{
			core.Text("Hello "),
			core.Span(core.Class("icon"), core.Text("*")),
		},
	})
	html := renderHTML(btn)

	// Should contain text content
	if !strings.Contains(html, "Hello ") {
		t.Errorf("expected text content, got: %s", html)
	}

	// Should contain nested span
	if !strings.Contains(html, "<span") {
		t.Errorf("expected nested span, got: %s", html)
	}
	if !strings.Contains(html, "*</span>") {
		t.Errorf("expected span content, got: %s", html)
	}

	// Should close button tag
	if !strings.HasSuffix(html, "</button>") {
		t.Errorf("expected closing </button>, got: %s", html)
	}
}

func TestButton_Type(t *testing.T) {
	tests := []struct {
		name         string
		buttonType   string
		expectedType string
	}{
		{"submit", "submit", "submit"},
		{"reset", "reset", "reset"},
		{"button", "button", "button"},
		{"default (empty)", "", "button"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			btn := Button(ButtonProps{
				Type:     tt.buttonType,
				Children: []core.Node{core.Text("Btn")},
			})
			html := renderHTML(btn)

			expected := `type="` + tt.expectedType + `"`
			if !strings.Contains(html, expected) {
				t.Errorf("expected %s, got: %s", expected, html)
			}
		})
	}
}

func TestButton_CombinedProps(t *testing.T) {
	btn := Button(ButtonProps{
		Variant:  ButtonDestructive,
		Size:     ButtonLG,
		Class:    "extra-class",
		Disabled: true,
		Type:     "submit",
		Children: []core.Node{core.Text("Delete")},
	})
	html := renderHTML(btn)

	// Check variant
	if !strings.Contains(html, "bg-red-600") {
		t.Errorf("expected destructive variant, got: %s", html)
	}

	// Check size
	if !strings.Contains(html, "px-6 py-3") {
		t.Errorf("expected large size, got: %s", html)
	}

	// Check custom class
	if !strings.Contains(html, "extra-class") {
		t.Errorf("expected extra-class, got: %s", html)
	}

	// Check disabled
	if !strings.Contains(html, "opacity-50") {
		t.Errorf("expected disabled styling, got: %s", html)
	}
	if !strings.Contains(html, `disabled="disabled"`) {
		t.Errorf("expected disabled attribute, got: %s", html)
	}

	// Check type
	if !strings.Contains(html, `type="submit"`) {
		t.Errorf("expected type=submit, got: %s", html)
	}

	// Check content
	if !strings.Contains(html, "Delete") {
		t.Errorf("expected Delete content, got: %s", html)
	}
}
