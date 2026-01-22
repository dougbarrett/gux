package ui

import (
	"strings"
	"testing"

	"github.com/dougbarrett/gux/core"
)

func TestTooltip_DefaultPosition(t *testing.T) {
	tip := Tooltip(TooltipProps{
		Content:  "Help text",
		Children: []core.Node{core.Text("Hover me")},
	})
	html := renderHTML(tip)

	// Should have TooltipTop position classes (default)
	if !strings.Contains(html, "bottom-full") {
		t.Errorf("expected bottom-full (top position), got: %s", html)
	}
	if !strings.Contains(html, "-translate-x-1/2") {
		t.Errorf("expected -translate-x-1/2 (top position), got: %s", html)
	}
	if !strings.Contains(html, "mb-2") {
		t.Errorf("expected mb-2 (top position margin), got: %s", html)
	}
}

func TestTooltip_AllPositions(t *testing.T) {
	tests := []struct {
		position      TooltipPosition
		expectedClass string
	}{
		{TooltipTop, "bottom-full"},
		{TooltipBottom, "top-full"},
		{TooltipLeft, "right-full"},
		{TooltipRight, "left-full"},
	}

	for _, tt := range tests {
		t.Run(string(tt.position), func(t *testing.T) {
			tip := Tooltip(TooltipProps{
				Content:  "Tip",
				Position: tt.position,
				Children: []core.Node{core.Text("Trigger")},
			})
			html := renderHTML(tip)

			if !strings.Contains(html, tt.expectedClass) {
				t.Errorf("position %s: expected class %q in: %s", tt.position, tt.expectedClass, html)
			}
		})
	}
}

func TestTooltip_AriaRole(t *testing.T) {
	tip := Tooltip(TooltipProps{
		Content:  "Accessible tooltip",
		Children: []core.Node{core.Text("Trigger")},
	})
	html := renderHTML(tip)

	if !strings.Contains(html, `role="tooltip"`) {
		t.Errorf("expected role=\"tooltip\", got: %s", html)
	}
}

func TestTooltip_Content(t *testing.T) {
	tip := Tooltip(TooltipProps{
		Content:  "This is the tooltip content",
		Children: []core.Node{core.Text("Trigger")},
	})
	html := renderHTML(tip)

	if !strings.Contains(html, "This is the tooltip content") {
		t.Errorf("expected tooltip content in output, got: %s", html)
	}
}

func TestTooltip_VisibilityClasses(t *testing.T) {
	tip := Tooltip(TooltipProps{
		Content:  "Tip",
		Children: []core.Node{core.Text("Trigger")},
	})
	html := renderHTML(tip)

	// Should have initial hidden state
	if !strings.Contains(html, "opacity-0") {
		t.Errorf("expected opacity-0 class, got: %s", html)
	}
	if !strings.Contains(html, "invisible") {
		t.Errorf("expected invisible class, got: %s", html)
	}

	// Should have hover visibility classes
	if !strings.Contains(html, "group-hover:opacity-100") {
		t.Errorf("expected group-hover:opacity-100 class, got: %s", html)
	}
	if !strings.Contains(html, "group-hover:visible") {
		t.Errorf("expected group-hover:visible class, got: %s", html)
	}

	// Should have focus visibility classes
	if !strings.Contains(html, "group-focus-within:opacity-100") {
		t.Errorf("expected group-focus-within:opacity-100 class, got: %s", html)
	}
	if !strings.Contains(html, "group-focus-within:visible") {
		t.Errorf("expected group-focus-within:visible class, got: %s", html)
	}
}

func TestTooltip_Children(t *testing.T) {
	tip := Tooltip(TooltipProps{
		Content: "Tip",
		Children: []core.Node{
			core.Button(core.Attrs{Class: "trigger-button"}, core.Text("Click me")),
		},
	})
	html := renderHTML(tip)

	// Children should render
	if !strings.Contains(html, "trigger-button") {
		t.Errorf("expected trigger-button class in children, got: %s", html)
	}
	if !strings.Contains(html, "Click me") {
		t.Errorf("expected child text content, got: %s", html)
	}

	// Children should appear before tooltip (check order: button comes before role="tooltip")
	buttonIdx := strings.Index(html, "trigger-button")
	tooltipIdx := strings.Index(html, `role="tooltip"`)
	if buttonIdx > tooltipIdx {
		t.Errorf("expected children before tooltip element, got: %s", html)
	}
}

func TestTooltip_CustomClass(t *testing.T) {
	tip := Tooltip(TooltipProps{
		Content:  "Tip",
		Class:    "my-custom-wrapper extra-class",
		Children: []core.Node{core.Text("Trigger")},
	})
	html := renderHTML(tip)

	if !strings.Contains(html, "my-custom-wrapper") {
		t.Errorf("expected my-custom-wrapper class, got: %s", html)
	}
	if !strings.Contains(html, "extra-class") {
		t.Errorf("expected extra-class class, got: %s", html)
	}
}

func TestTooltip_GroupClass(t *testing.T) {
	tip := Tooltip(TooltipProps{
		Content:  "Tip",
		Children: []core.Node{core.Text("Trigger")},
	})
	html := renderHTML(tip)

	// Wrapper must have "group" class for Tailwind group-hover to work
	if !strings.Contains(html, "group") {
		t.Errorf("expected group class on wrapper, got: %s", html)
	}
}

func TestTooltip_TransitionClasses(t *testing.T) {
	tip := Tooltip(TooltipProps{
		Content:  "Tip",
		Children: []core.Node{core.Text("Trigger")},
	})
	html := renderHTML(tip)

	// Should have transition classes for smooth appearance
	if !strings.Contains(html, "transition-opacity") {
		t.Errorf("expected transition-opacity class, got: %s", html)
	}
	if !strings.Contains(html, "duration-200") {
		t.Errorf("expected duration-200 class, got: %s", html)
	}
	if !strings.Contains(html, "pointer-events-none") {
		t.Errorf("expected pointer-events-none class, got: %s", html)
	}
}

func TestTooltip_BaseClasses(t *testing.T) {
	tip := Tooltip(TooltipProps{
		Content:  "Tip",
		Children: []core.Node{core.Text("Trigger")},
	})
	html := renderHTML(tip)

	// Should have tooltip base styling classes
	if !strings.Contains(html, "z-50") {
		t.Errorf("expected z-50 class, got: %s", html)
	}
	if !strings.Contains(html, "bg-gray-900") {
		t.Errorf("expected bg-gray-900 class, got: %s", html)
	}
	if !strings.Contains(html, "text-white") {
		t.Errorf("expected text-white class, got: %s", html)
	}
	if !strings.Contains(html, "rounded") {
		t.Errorf("expected rounded class, got: %s", html)
	}
	if !strings.Contains(html, "shadow-lg") {
		t.Errorf("expected shadow-lg class, got: %s", html)
	}
}
