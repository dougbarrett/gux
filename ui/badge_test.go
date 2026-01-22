package ui

import (
	"strings"
	"testing"
)

func TestBadge_DefaultVariant(t *testing.T) {
	badge := Badge(BadgeProps{
		Text: "Status",
	})
	html := renderHTML(badge)

	// Should render as span element
	if !strings.HasPrefix(html, "<span") {
		t.Errorf("expected <span> element, got: %s", html)
	}

	// Should have default variant classes
	if !strings.Contains(html, "bg-gray-100") {
		t.Errorf("expected default variant bg-gray-100, got: %s", html)
	}
	if !strings.Contains(html, "text-gray-800") {
		t.Errorf("expected default variant text-gray-800, got: %s", html)
	}

	// Should have base classes
	if !strings.Contains(html, "rounded-full") {
		t.Errorf("expected rounded-full class, got: %s", html)
	}
	if !strings.Contains(html, "text-xs") {
		t.Errorf("expected text-xs class, got: %s", html)
	}
}

func TestBadge_AllVariants(t *testing.T) {
	tests := []struct {
		variant       BadgeVariant
		expectedClass string
	}{
		{BadgeDefault, "bg-gray-100"},
		{BadgeSuccess, "bg-green-100"},
		{BadgeWarning, "bg-yellow-100"},
		{BadgeError, "bg-red-100"},
	}

	for _, tt := range tests {
		t.Run(string(tt.variant), func(t *testing.T) {
			badge := Badge(BadgeProps{
				Text:    "Test",
				Variant: tt.variant,
			})
			html := renderHTML(badge)

			if !strings.Contains(html, tt.expectedClass) {
				t.Errorf("variant %s: expected class %q in: %s", tt.variant, tt.expectedClass, html)
			}
		})
	}
}

func TestBadge_Text(t *testing.T) {
	badge := Badge(BadgeProps{
		Text: "Active",
	})
	html := renderHTML(badge)

	// Should contain the text content
	if !strings.Contains(html, "Active") {
		t.Errorf("expected text 'Active', got: %s", html)
	}

	// Should close span tag
	if !strings.HasSuffix(html, "</span>") {
		t.Errorf("expected closing </span>, got: %s", html)
	}
}

func TestBadge_CustomClass(t *testing.T) {
	badge := Badge(BadgeProps{
		Text:  "Custom",
		Class: "my-custom-class ml-2",
	})
	html := renderHTML(badge)

	// Should include custom classes
	if !strings.Contains(html, "my-custom-class") {
		t.Errorf("expected custom class, got: %s", html)
	}
	if !strings.Contains(html, "ml-2") {
		t.Errorf("expected ml-2 class, got: %s", html)
	}

	// Should still have base classes
	if !strings.Contains(html, "inline-flex") {
		t.Errorf("expected base classes, got: %s", html)
	}
}

func TestBadge_EmptyText(t *testing.T) {
	badge := Badge(BadgeProps{
		Text:    "",
		Variant: BadgeSuccess,
	})
	html := renderHTML(badge)

	// Should render even with empty text
	if !strings.HasPrefix(html, "<span") {
		t.Errorf("expected <span> element, got: %s", html)
	}

	// Should have variant classes
	if !strings.Contains(html, "bg-green-100") {
		t.Errorf("expected success variant classes, got: %s", html)
	}
}

func TestBadge_CombinedProps(t *testing.T) {
	badge := Badge(BadgeProps{
		Text:    "Error",
		Variant: BadgeError,
		Class:   "extra-class",
	})
	html := renderHTML(badge)

	// Check variant
	if !strings.Contains(html, "bg-red-100") {
		t.Errorf("expected error variant, got: %s", html)
	}

	// Check custom class
	if !strings.Contains(html, "extra-class") {
		t.Errorf("expected extra-class, got: %s", html)
	}

	// Check text
	if !strings.Contains(html, "Error") {
		t.Errorf("expected Error content, got: %s", html)
	}

	// Check base classes
	if !strings.Contains(html, "font-medium") {
		t.Errorf("expected font-medium base class, got: %s", html)
	}
}
