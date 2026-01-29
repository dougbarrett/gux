package ui

import (
	"strings"
	"testing"
)

func TestIcon_KnownIcon(t *testing.T) {
	icon := Icon(IconProps{Name: "home", Size: IconMD})
	html := renderHTML(icon)

	if !strings.Contains(html, "<svg") {
		t.Errorf("expected SVG element, got: %s", html)
	}
	if !strings.Contains(html, "<path") {
		t.Errorf("expected path element, got: %s", html)
	}
	if !strings.Contains(html, "w-5 h-5") {
		t.Errorf("expected MD size classes, got: %s", html)
	}
}

func TestIcon_DefaultSize(t *testing.T) {
	icon := Icon(IconProps{Name: "home"})
	html := renderHTML(icon)

	if !strings.Contains(html, "w-5 h-5") {
		t.Errorf("expected default MD size classes, got: %s", html)
	}
}

func TestIcon_AllSizes(t *testing.T) {
	tests := []struct {
		size     IconSize
		expected string
	}{
		{IconXS, "w-3 h-3"},
		{IconSM, "w-4 h-4"},
		{IconMD, "w-5 h-5"},
		{IconLG, "w-6 h-6"},
		{IconXL, "w-8 h-8"},
	}

	for _, tt := range tests {
		t.Run(string(tt.size), func(t *testing.T) {
			icon := Icon(IconProps{Name: "home", Size: tt.size})
			html := renderHTML(icon)
			if !strings.Contains(html, tt.expected) {
				t.Errorf("size %s: expected %q in output, got: %s", tt.size, tt.expected, html)
			}
		})
	}
}

func TestIcon_CustomClass(t *testing.T) {
	icon := Icon(IconProps{Name: "home", Class: "text-red-500"})
	html := renderHTML(icon)

	if !strings.Contains(html, "text-red-500") {
		t.Errorf("expected custom class, got: %s", html)
	}
}

func TestIcon_UnknownRendersFallback(t *testing.T) {
	icon := Icon(IconProps{Name: "nonexistent-icon"})
	html := renderHTML(icon)

	// Should render an SVG fallback, not an empty fragment
	if !strings.Contains(html, "<svg") {
		t.Errorf("expected SVG fallback for unknown icon, got: %s", html)
	}
	// Should contain the circle placeholder
	if !strings.Contains(html, "<circle") {
		t.Errorf("expected circle element in fallback, got: %s", html)
	}
}

func TestIcon_EmptyNameRendersFallback(t *testing.T) {
	icon := Icon(IconProps{Name: ""})
	html := renderHTML(icon)

	if !strings.Contains(html, "<svg") {
		t.Errorf("expected SVG fallback for empty icon name, got: %s", html)
	}
}

func TestIcon_NewIcons(t *testing.T) {
	// Verify all icons requested in issue #32 are present
	requestedIcons := []string{
		"tag", "briefcase", "file-text", "package", "gift",
		"user-cog", "book-open", "log-out", "shopping-cart",
		"clipboard", "star", "heart", "mail", "phone",
		"globe", "calendar", "clock", "edit", "trash",
		"eye", "link", "upload", "download",
	}

	for _, name := range requestedIcons {
		t.Run(name, func(t *testing.T) {
			icon := Icon(IconProps{Name: name})
			html := renderHTML(icon)
			if !strings.Contains(html, "<path") {
				t.Errorf("icon %q should render a path element, got: %s", name, html)
			}
		})
	}
}

func TestIcon_AdditionalIcons(t *testing.T) {
	// Verify additional commonly used icons are present
	additionalIcons := []string{
		"check", "info", "warning", "error", "filter", "sort",
		"refresh", "lock", "unlock", "key", "shield", "cog",
		"database", "server", "code", "terminal", "image", "copy",
		"minus", "dots-vertical", "dots-horizontal",
		"arrow-left", "arrow-right", "arrow-up", "arrow-down",
		"external-link", "hashtag", "at-sign", "paperclip",
		"map-pin", "credit-card", "currency-dollar", "printer",
	}

	for _, name := range additionalIcons {
		t.Run(name, func(t *testing.T) {
			icon := Icon(IconProps{Name: name})
			html := renderHTML(icon)
			if !strings.Contains(html, "<path") && !strings.Contains(html, "<circle") {
				t.Errorf("icon %q should render SVG content, got: %s", name, html)
			}
		})
	}
}

func TestIconNames_ReturnsAll(t *testing.T) {
	names := IconNames()

	// Should have the original icons plus all new ones
	if len(names) < 40 {
		t.Errorf("expected at least 40 icons, got %d", len(names))
	}

	// Check that specific icons exist
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}

	required := []string{"home", "tag", "briefcase", "star", "edit", "trash", "download"}
	for _, name := range required {
		if !nameSet[name] {
			t.Errorf("IconNames() missing %q", name)
		}
	}
}

func TestIcon_SVGAttributes(t *testing.T) {
	icon := Icon(IconProps{Name: "home"})
	html := renderHTML(icon)

	if !strings.Contains(html, `fill="none"`) {
		t.Errorf("expected fill=none attribute, got: %s", html)
	}
	if !strings.Contains(html, `viewBox="0 0 24 24"`) {
		t.Errorf("expected viewBox attribute, got: %s", html)
	}
	if !strings.Contains(html, `aria-hidden="true"`) {
		t.Errorf("expected aria-hidden attribute, got: %s", html)
	}
}
