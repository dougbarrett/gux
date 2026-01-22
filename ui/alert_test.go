package ui

import (
	"strings"
	"testing"
)

// TestAlert_Renders verifies Alert renders with base classes.
func TestAlert_Renders(t *testing.T) {
	alert := Alert(AlertProps{Message: "Test alert"})
	html := renderHTML(alert)

	// Should render as div
	if !strings.HasPrefix(html, "<div") {
		t.Errorf("expected <div> element, got: %s", html)
	}

	// Should have base classes
	if !strings.Contains(html, "rounded-lg") {
		t.Errorf("expected rounded-lg class, got: %s", html)
	}
	if !strings.Contains(html, "border") {
		t.Errorf("expected border class, got: %s", html)
	}
	if !strings.Contains(html, "flex") {
		t.Errorf("expected flex class, got: %s", html)
	}
}

// TestAlert_AriaRole verifies role="alert" is present.
func TestAlert_AriaRole(t *testing.T) {
	alert := Alert(AlertProps{Message: "Important"})
	html := renderHTML(alert)

	if !strings.Contains(html, `role="alert"`) {
		t.Errorf("expected role=\"alert\", got: %s", html)
	}
}

// TestAlert_DefaultVariant verifies empty variant defaults to AlertInfo styling.
func TestAlert_DefaultVariant(t *testing.T) {
	alert := Alert(AlertProps{Message: "Info message"})
	html := renderHTML(alert)

	// AlertInfo uses blue styling
	if !strings.Contains(html, "bg-blue-50") {
		t.Errorf("expected bg-blue-50 for default info variant, got: %s", html)
	}
	if !strings.Contains(html, "border-blue-200") {
		t.Errorf("expected border-blue-200 for default info variant, got: %s", html)
	}
	if !strings.Contains(html, "text-blue-800") {
		t.Errorf("expected text-blue-800 for default info variant, got: %s", html)
	}
}

// TestAlert_AllVariants verifies all 4 variants have correct styling.
func TestAlert_AllVariants(t *testing.T) {
	tests := []struct {
		variant        AlertVariant
		expectedBg     string
		expectedBorder string
		expectedText   string
	}{
		{AlertInfo, "bg-blue-50", "border-blue-200", "text-blue-800"},
		{AlertSuccess, "bg-green-50", "border-green-200", "text-green-800"},
		{AlertWarning, "bg-yellow-50", "border-yellow-200", "text-yellow-800"},
		{AlertError, "bg-red-50", "border-red-200", "text-red-800"},
	}

	for _, tt := range tests {
		t.Run(string(tt.variant), func(t *testing.T) {
			alert := Alert(AlertProps{
				Message: "Test",
				Variant: tt.variant,
			})
			html := renderHTML(alert)

			if !strings.Contains(html, tt.expectedBg) {
				t.Errorf("variant %s: expected %q in: %s", tt.variant, tt.expectedBg, html)
			}
			if !strings.Contains(html, tt.expectedBorder) {
				t.Errorf("variant %s: expected %q in: %s", tt.variant, tt.expectedBorder, html)
			}
			if !strings.Contains(html, tt.expectedText) {
				t.Errorf("variant %s: expected %q in: %s", tt.variant, tt.expectedText, html)
			}
		})
	}
}

// TestAlert_Message verifies message text renders.
func TestAlert_Message(t *testing.T) {
	alert := Alert(AlertProps{Message: "Operation failed"})
	html := renderHTML(alert)

	if !strings.Contains(html, "Operation failed") {
		t.Errorf("expected message text, got: %s", html)
	}
}

// TestAlert_Title verifies Title renders in strong tag when provided.
func TestAlert_Title(t *testing.T) {
	alert := Alert(AlertProps{
		Message: "Please try again",
		Title:   "Error",
	})
	html := renderHTML(alert)

	// Should have strong tag
	if !strings.Contains(html, "<strong") {
		t.Errorf("expected <strong> tag for title, got: %s", html)
	}

	// Should contain title text
	if !strings.Contains(html, "Error") {
		t.Errorf("expected title text, got: %s", html)
	}
}

// TestAlert_NoTitle verifies no strong tag when Title is empty.
func TestAlert_NoTitle(t *testing.T) {
	alert := Alert(AlertProps{Message: "Simple message"})
	html := renderHTML(alert)

	// Should NOT have strong tag
	if strings.Contains(html, "<strong") {
		t.Errorf("expected no <strong> tag when title empty, got: %s", html)
	}
}

// TestAlert_WithCloseButton verifies OnClose provided renders close button with aria-label.
func TestAlert_WithCloseButton(t *testing.T) {
	alert := Alert(AlertProps{
		Message: "Dismissable",
		OnClose: func() {},
	})
	html := renderHTML(alert)

	// Should have close button
	if !strings.Contains(html, "<button") {
		t.Errorf("expected close button when OnClose provided, got: %s", html)
	}

	// Should have aria-label
	if !strings.Contains(html, `aria-label="Dismiss"`) {
		t.Errorf("expected aria-label on close button, got: %s", html)
	}
}

// TestAlert_WithoutCloseButton verifies no OnClose omits close button.
func TestAlert_WithoutCloseButton(t *testing.T) {
	alert := Alert(AlertProps{Message: "No close"})
	html := renderHTML(alert)

	// Should NOT have close button
	if strings.Contains(html, "<button") {
		t.Errorf("expected no close button when OnClose not provided, got: %s", html)
	}
}

// TestAlert_IconAriaHidden verifies icon has aria-hidden="true".
func TestAlert_IconAriaHidden(t *testing.T) {
	alert := Alert(AlertProps{Message: "Test"})
	html := renderHTML(alert)

	if !strings.Contains(html, `aria-hidden="true"`) {
		t.Errorf("expected icon aria-hidden=\"true\", got: %s", html)
	}
}

// TestAlert_CustomClass verifies custom class merges with base classes.
func TestAlert_CustomClass(t *testing.T) {
	alert := Alert(AlertProps{
		Message: "Custom",
		Class:   "my-alert-class",
	})
	html := renderHTML(alert)

	if !strings.Contains(html, "my-alert-class") {
		t.Errorf("expected custom class, got: %s", html)
	}
	// Should still have base classes
	if !strings.Contains(html, "rounded-lg") {
		t.Errorf("expected base rounded-lg class, got: %s", html)
	}
}
