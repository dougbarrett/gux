package ui

import (
	"strings"
	"testing"
)

func TestAvatar_WithImage(t *testing.T) {
	avatar := Avatar(AvatarProps{
		Src: "/images/user.jpg",
		Alt: "John Doe",
	})
	html := renderHTML(avatar)

	// Should render as div containing img
	if !strings.HasPrefix(html, "<div") {
		t.Errorf("expected <div> element, got: %s", html)
	}

	// Should contain img element
	if !strings.Contains(html, "<img") {
		t.Errorf("expected <img> element, got: %s", html)
	}

	// Should have src attribute
	if !strings.Contains(html, `src="/images/user.jpg"`) {
		t.Errorf("expected src attribute, got: %s", html)
	}

	// Should have alt attribute
	if !strings.Contains(html, `alt="John Doe"`) {
		t.Errorf("expected alt attribute, got: %s", html)
	}

	// Should have image styling
	if !strings.Contains(html, "object-cover") {
		t.Errorf("expected object-cover class on img, got: %s", html)
	}
}

func TestAvatar_WithInitials(t *testing.T) {
	avatar := Avatar(AvatarProps{
		Name: "John Doe",
	})
	html := renderHTML(avatar)

	// Should render as div (no img)
	if !strings.HasPrefix(html, "<div") {
		t.Errorf("expected <div> element, got: %s", html)
	}

	// Should NOT contain img element
	if strings.Contains(html, "<img") {
		t.Errorf("expected no <img> element for initials fallback, got: %s", html)
	}

	// Should contain initials
	if !strings.Contains(html, "JD") {
		t.Errorf("expected initials JD, got: %s", html)
	}

	// Should have fallback styling
	if !strings.Contains(html, "bg-gray-200") {
		t.Errorf("expected fallback bg-gray-200, got: %s", html)
	}
	if !strings.Contains(html, "font-medium") {
		t.Errorf("expected font-medium class, got: %s", html)
	}
}

func TestAvatar_AllSizes(t *testing.T) {
	tests := []struct {
		size          AvatarSize
		expectedClass string
	}{
		{AvatarSM, "w-8 h-8 text-xs"},
		{AvatarMD, "w-10 h-10 text-sm"},
		{AvatarLG, "w-12 h-12 text-base"},
	}

	for _, tt := range tests {
		t.Run(string(tt.size), func(t *testing.T) {
			avatar := Avatar(AvatarProps{
				Name: "Test",
				Size: tt.size,
			})
			html := renderHTML(avatar)

			// Check for size-specific classes
			parts := strings.Fields(tt.expectedClass)
			for _, part := range parts {
				if !strings.Contains(html, part) {
					t.Errorf("size %s: expected class %q in: %s", tt.size, part, html)
				}
			}
		})
	}
}

func TestAvatar_DefaultSize(t *testing.T) {
	avatar := Avatar(AvatarProps{
		Name: "Test",
	})
	html := renderHTML(avatar)

	// Default should be MD (w-10 h-10)
	if !strings.Contains(html, "w-10") {
		t.Errorf("expected default w-10, got: %s", html)
	}
	if !strings.Contains(html, "h-10") {
		t.Errorf("expected default h-10, got: %s", html)
	}
}

func TestAvatar_InitialsEdgeCases(t *testing.T) {
	tests := []struct {
		name             string
		inputName        string
		expectedInitials string
	}{
		{"single name", "John", "J"},
		{"two names", "John Doe", "JD"},
		{"three names", "John William Doe", "JD"},
		{"empty name", "", "?"},
		{"whitespace only", "   ", "?"},
		{"lowercase", "john doe", "JD"},
		{"mixed case", "jOHN dOE", "JD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avatar := Avatar(AvatarProps{
				Name: tt.inputName,
			})
			html := renderHTML(avatar)

			if !strings.Contains(html, tt.expectedInitials) {
				t.Errorf("expected initials %q for name %q, got: %s", tt.expectedInitials, tt.inputName, html)
			}
		})
	}
}

func TestAvatar_CustomClass(t *testing.T) {
	avatar := Avatar(AvatarProps{
		Name:  "Test",
		Class: "my-custom-class border-2",
	})
	html := renderHTML(avatar)

	// Should include custom classes
	if !strings.Contains(html, "my-custom-class") {
		t.Errorf("expected custom class, got: %s", html)
	}
	if !strings.Contains(html, "border-2") {
		t.Errorf("expected border-2 class, got: %s", html)
	}

	// Should still have base classes
	if !strings.Contains(html, "rounded-full") {
		t.Errorf("expected base classes, got: %s", html)
	}
}

func TestAvatar_ImageWithSize(t *testing.T) {
	avatar := Avatar(AvatarProps{
		Src:  "/images/user.jpg",
		Alt:  "Test User",
		Size: AvatarLG,
	})
	html := renderHTML(avatar)

	// Should have large size classes
	if !strings.Contains(html, "w-12") {
		t.Errorf("expected w-12 for large size, got: %s", html)
	}

	// Should contain img
	if !strings.Contains(html, "<img") {
		t.Errorf("expected img element, got: %s", html)
	}
}

func TestAvatar_ImageTakesPrecedence(t *testing.T) {
	// When both Src and Name are provided, Src should be used
	avatar := Avatar(AvatarProps{
		Src:  "/images/user.jpg",
		Alt:  "Test",
		Name: "John Doe",
	})
	html := renderHTML(avatar)

	// Should render image, not initials
	if !strings.Contains(html, "<img") {
		t.Errorf("expected img element when Src is provided, got: %s", html)
	}

	// Should NOT contain initials
	if strings.Contains(html, "JD") {
		t.Errorf("expected no initials when Src is provided, got: %s", html)
	}
}

func TestAvatar_BaseClasses(t *testing.T) {
	avatar := Avatar(AvatarProps{
		Name: "Test",
	})
	html := renderHTML(avatar)

	// Should have all base classes
	baseClasses := []string{"rounded-full", "overflow-hidden", "flex", "items-center", "justify-center"}
	for _, class := range baseClasses {
		if !strings.Contains(html, class) {
			t.Errorf("expected base class %q, got: %s", class, html)
		}
	}
}
