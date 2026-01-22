package ui

import (
	"strings"
	"testing"

	"github.com/dougbarrett/gux/core"
)

// TestToastContainer_AriaLive verifies aria-live="polite" is present.
func TestToastContainer_AriaLive(t *testing.T) {
	container := ToastContainer(ToastContainerProps{})
	html := renderHTML(container)

	if !strings.Contains(html, `aria-live="polite"`) {
		t.Errorf("expected aria-live=\"polite\", got: %s", html)
	}
}

// TestToastContainer_AriaRole verifies role="status" is present.
func TestToastContainer_AriaRole(t *testing.T) {
	container := ToastContainer(ToastContainerProps{})
	html := renderHTML(container)

	if !strings.Contains(html, `role="status"`) {
		t.Errorf("expected role=\"status\", got: %s", html)
	}
}

// TestToastContainer_AriaAtomic verifies aria-atomic="true" is present.
func TestToastContainer_AriaAtomic(t *testing.T) {
	container := ToastContainer(ToastContainerProps{})
	html := renderHTML(container)

	if !strings.Contains(html, `aria-atomic="true"`) {
		t.Errorf("expected aria-atomic=\"true\", got: %s", html)
	}
}

// TestToastContainer_DefaultPosition verifies empty position defaults to ToastTopRight.
func TestToastContainer_DefaultPosition(t *testing.T) {
	container := ToastContainer(ToastContainerProps{})
	html := renderHTML(container)

	// ToastTopRight uses "top-4 right-4"
	if !strings.Contains(html, "top-4") {
		t.Errorf("expected top-4 class for default position, got: %s", html)
	}
	if !strings.Contains(html, "right-4") {
		t.Errorf("expected right-4 class for default position, got: %s", html)
	}
}

// TestToastContainer_AllPositions verifies all 6 positions produce correct classes.
func TestToastContainer_AllPositions(t *testing.T) {
	tests := []struct {
		position ToastPosition
		expected string
	}{
		{ToastTopRight, "top-4"},
		{ToastTopRight, "right-4"},
		{ToastTopLeft, "top-4"},
		{ToastTopLeft, "left-4"},
		{ToastBottomRight, "bottom-4"},
		{ToastBottomRight, "right-4"},
		{ToastBottomLeft, "bottom-4"},
		{ToastBottomLeft, "left-4"},
		{ToastTopCenter, "top-4"},
		{ToastTopCenter, "-translate-x-1/2"},
		{ToastBottomCenter, "bottom-4"},
		{ToastBottomCenter, "-translate-x-1/2"},
	}

	for _, tt := range tests {
		t.Run(string(tt.position)+"-"+tt.expected, func(t *testing.T) {
			container := ToastContainer(ToastContainerProps{
				Position: tt.position,
			})
			html := renderHTML(container)

			if !strings.Contains(html, tt.expected) {
				t.Errorf("position %s: expected %q in: %s", tt.position, tt.expected, html)
			}
		})
	}
}

// TestToastContainer_FixedPositioning verifies fixed and z-50 classes.
func TestToastContainer_FixedPositioning(t *testing.T) {
	container := ToastContainer(ToastContainerProps{})
	html := renderHTML(container)

	if !strings.Contains(html, "fixed") {
		t.Errorf("expected fixed class, got: %s", html)
	}
	if !strings.Contains(html, "z-50") {
		t.Errorf("expected z-50 class, got: %s", html)
	}
}

// TestToastContainer_Children verifies Children render inside.
func TestToastContainer_Children(t *testing.T) {
	container := ToastContainer(ToastContainerProps{
		Children: []core.Node{
			Toast(ToastProps{Message: "Test toast"}),
		},
	})
	html := renderHTML(container)

	if !strings.Contains(html, "Test toast") {
		t.Errorf("expected toast child content, got: %s", html)
	}
}

// TestToastContainer_CustomClass verifies custom class merges.
func TestToastContainer_CustomClass(t *testing.T) {
	container := ToastContainer(ToastContainerProps{
		Class: "my-custom-class",
	})
	html := renderHTML(container)

	if !strings.Contains(html, "my-custom-class") {
		t.Errorf("expected custom class, got: %s", html)
	}
	// Should still have base classes
	if !strings.Contains(html, "flex") {
		t.Errorf("expected base flex class, got: %s", html)
	}
}

// TestToastContainer_ID verifies id="toast-container" is present.
func TestToastContainer_ID(t *testing.T) {
	container := ToastContainer(ToastContainerProps{})
	html := renderHTML(container)

	if !strings.Contains(html, `id="toast-container"`) {
		t.Errorf("expected id=\"toast-container\", got: %s", html)
	}
}

// TestToast_Renders verifies Toast renders with base classes.
func TestToast_Renders(t *testing.T) {
	toast := Toast(ToastProps{Message: "Hello"})
	html := renderHTML(toast)

	// Should render as div
	if !strings.HasPrefix(html, "<div") {
		t.Errorf("expected <div> element, got: %s", html)
	}

	// Should have base classes
	if !strings.Contains(html, "rounded-lg") {
		t.Errorf("expected rounded-lg class, got: %s", html)
	}
	if !strings.Contains(html, "shadow-lg") {
		t.Errorf("expected shadow-lg class, got: %s", html)
	}
	if !strings.Contains(html, "min-w-64") {
		t.Errorf("expected min-w-64 class, got: %s", html)
	}
}

// TestToast_DefaultVariant verifies empty variant defaults to ToastInfo styling.
func TestToast_DefaultVariant(t *testing.T) {
	toast := Toast(ToastProps{Message: "Info message"})
	html := renderHTML(toast)

	// ToastInfo uses blue styling
	if !strings.Contains(html, "bg-blue-100") {
		t.Errorf("expected bg-blue-100 for default info variant, got: %s", html)
	}
	if !strings.Contains(html, "text-blue-800") {
		t.Errorf("expected text-blue-800 for default info variant, got: %s", html)
	}
}

// TestToast_AllVariants verifies all 4 variants have correct styling.
func TestToast_AllVariants(t *testing.T) {
	tests := []struct {
		variant      ToastVariant
		expectedBg   string
		expectedText string
	}{
		{ToastInfo, "bg-blue-100", "text-blue-800"},
		{ToastSuccess, "bg-green-100", "text-green-800"},
		{ToastWarning, "bg-yellow-100", "text-yellow-800"},
		{ToastError, "bg-red-100", "text-red-800"},
	}

	for _, tt := range tests {
		t.Run(string(tt.variant), func(t *testing.T) {
			toast := Toast(ToastProps{
				Message: "Test",
				Variant: tt.variant,
			})
			html := renderHTML(toast)

			if !strings.Contains(html, tt.expectedBg) {
				t.Errorf("variant %s: expected %q in: %s", tt.variant, tt.expectedBg, html)
			}
			if !strings.Contains(html, tt.expectedText) {
				t.Errorf("variant %s: expected %q in: %s", tt.variant, tt.expectedText, html)
			}
		})
	}
}

// TestToast_Message verifies message text renders.
func TestToast_Message(t *testing.T) {
	toast := Toast(ToastProps{Message: "Operation completed"})
	html := renderHTML(toast)

	if !strings.Contains(html, "Operation completed") {
		t.Errorf("expected message text, got: %s", html)
	}
}

// TestToast_WithCloseButton verifies OnClose provided renders close button with aria-label.
func TestToast_WithCloseButton(t *testing.T) {
	toast := Toast(ToastProps{
		Message: "Dismissable",
		OnClose: func() {},
	})
	html := renderHTML(toast)

	// Should have close button
	if !strings.Contains(html, "<button") {
		t.Errorf("expected close button when OnClose provided, got: %s", html)
	}

	// Should have aria-label
	if !strings.Contains(html, `aria-label="Dismiss notification"`) {
		t.Errorf("expected aria-label on close button, got: %s", html)
	}
}

// TestToast_WithoutCloseButton verifies no OnClose omits close button.
func TestToast_WithoutCloseButton(t *testing.T) {
	toast := Toast(ToastProps{Message: "No close"})
	html := renderHTML(toast)

	// Should NOT have close button
	if strings.Contains(html, "<button") {
		t.Errorf("expected no close button when OnClose not provided, got: %s", html)
	}
}

// TestToast_IconAriaHidden verifies icon has aria-hidden="true".
func TestToast_IconAriaHidden(t *testing.T) {
	toast := Toast(ToastProps{Message: "Test"})
	html := renderHTML(toast)

	if !strings.Contains(html, `aria-hidden="true"`) {
		t.Errorf("expected icon aria-hidden=\"true\", got: %s", html)
	}
}

// TestToast_CustomClass verifies custom class merges.
func TestToast_CustomClass(t *testing.T) {
	toast := Toast(ToastProps{
		Message: "Custom",
		Class:   "my-toast-class",
	})
	html := renderHTML(toast)

	if !strings.Contains(html, "my-toast-class") {
		t.Errorf("expected custom class, got: %s", html)
	}
	// Should still have base classes
	if !strings.Contains(html, "rounded-lg") {
		t.Errorf("expected base rounded-lg class, got: %s", html)
	}
}
