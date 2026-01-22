package ui

import (
	"strings"
	"testing"

	"github.com/dougbarrett/gux/core"
)

// Note: renderHTML helper is defined in button_test.go

func TestFormField_Basic(t *testing.T) {
	html := renderHTML(FormField(FormFieldProps{}))

	// Should render as div
	if !strings.HasPrefix(html, "<div") {
		t.Errorf("expected <div> element, got: %s", html)
	}

	// Should have mb-4 class
	if !strings.Contains(html, "mb-4") {
		t.Errorf("expected mb-4 class, got: %s", html)
	}
}

func TestFormField_Label(t *testing.T) {
	html := renderHTML(FormField(FormFieldProps{
		Label: "Email Address",
	}))

	// Should contain label element
	if !strings.Contains(html, "<label") {
		t.Errorf("expected <label> element, got: %s", html)
	}

	// Should contain label text
	if !strings.Contains(html, "Email Address") {
		t.Errorf("expected label text, got: %s", html)
	}

	// Should have correct label classes
	if !strings.Contains(html, "text-sm") {
		t.Errorf("expected text-sm class on label, got: %s", html)
	}
	if !strings.Contains(html, "font-medium") {
		t.Errorf("expected font-medium class on label, got: %s", html)
	}
	if !strings.Contains(html, "text-gray-700") {
		t.Errorf("expected text-gray-700 class on label, got: %s", html)
	}
}

func TestFormField_LabelFor(t *testing.T) {
	html := renderHTML(FormField(FormFieldProps{
		Label:    "Username",
		LabelFor: "username-input",
	}))

	// Should have for attribute
	if !strings.Contains(html, `for="username-input"`) {
		t.Errorf("expected for attribute, got: %s", html)
	}
}

func TestFormField_Required(t *testing.T) {
	html := renderHTML(FormField(FormFieldProps{
		Label:    "Password",
		Required: true,
	}))

	// Should contain asterisk
	if !strings.Contains(html, "*") {
		t.Errorf("expected asterisk for required field, got: %s", html)
	}

	// Asterisk should have red styling
	if !strings.Contains(html, "text-red-500") {
		t.Errorf("expected text-red-500 for asterisk, got: %s", html)
	}

	// Asterisk should be in a span
	if !strings.Contains(html, "<span") {
		t.Errorf("expected asterisk in span, got: %s", html)
	}
}

func TestFormField_Error(t *testing.T) {
	html := renderHTML(FormField(FormFieldProps{
		Label: "Email",
		Error: "This field is required",
	}))

	// Should contain error message
	if !strings.Contains(html, "This field is required") {
		t.Errorf("expected error message, got: %s", html)
	}

	// Error should have role="alert"
	if !strings.Contains(html, `role="alert"`) {
		t.Errorf("expected role=\"alert\" on error, got: %s", html)
	}

	// Error should have red styling
	if !strings.Contains(html, "text-red-600") {
		t.Errorf("expected text-red-600 class on error, got: %s", html)
	}
}

func TestFormField_Description(t *testing.T) {
	html := renderHTML(FormField(FormFieldProps{
		Label:       "Email",
		Description: "We will never share your email",
	}))

	// Should contain description text
	if !strings.Contains(html, "We will never share your email") {
		t.Errorf("expected description text, got: %s", html)
	}

	// Description should have gray styling
	if !strings.Contains(html, "text-gray-500") {
		t.Errorf("expected text-gray-500 class on description, got: %s", html)
	}
}

func TestFormField_ErrorOverridesDescription(t *testing.T) {
	html := renderHTML(FormField(FormFieldProps{
		Label:       "Email",
		Description: "We will never share your email",
		Error:       "Invalid email format",
	}))

	// Should contain error message
	if !strings.Contains(html, "Invalid email format") {
		t.Errorf("expected error message, got: %s", html)
	}

	// Should NOT contain description when error is present
	if strings.Contains(html, "We will never share your email") {
		t.Errorf("description should not appear when error is present, got: %s", html)
	}
}

func TestFormField_Children(t *testing.T) {
	html := renderHTML(FormField(FormFieldProps{
		Label: "Name",
		Children: []core.Node{
			Input(InputProps{
				ID:   "name-input",
				Name: "name",
			}),
		},
	}))

	// Should contain the input
	if !strings.Contains(html, "<input") {
		t.Errorf("expected input element in children, got: %s", html)
	}
	if !strings.Contains(html, `id="name-input"`) {
		t.Errorf("expected input with id, got: %s", html)
	}
}

func TestFormField_CustomClass(t *testing.T) {
	html := renderHTML(FormField(FormFieldProps{
		Class: "my-custom-class mt-8",
	}))

	// Should include custom classes
	if !strings.Contains(html, "my-custom-class") {
		t.Errorf("expected custom class, got: %s", html)
	}
	if !strings.Contains(html, "mt-8") {
		t.Errorf("expected mt-8 class, got: %s", html)
	}

	// Should still have mb-4 default class
	if !strings.Contains(html, "mb-4") {
		t.Errorf("expected mb-4 default class, got: %s", html)
	}
}

func TestFormField_CombinedProps(t *testing.T) {
	html := renderHTML(FormField(FormFieldProps{
		Label:    "Email",
		LabelFor: "email-input",
		Required: true,
		Error:    "Invalid email",
		Class:    "custom-field",
		Children: []core.Node{
			Input(InputProps{
				ID:    "email-input",
				Type:  InputEmail,
				Name:  "email",
				Error: "Invalid email",
			}),
		},
	}))

	// Check label
	if !strings.Contains(html, "Email") {
		t.Errorf("expected label text, got: %s", html)
	}

	// Check for attribute
	if !strings.Contains(html, `for="email-input"`) {
		t.Errorf("expected for attribute, got: %s", html)
	}

	// Check required asterisk
	if !strings.Contains(html, "*") {
		t.Errorf("expected asterisk, got: %s", html)
	}

	// Check error message
	if !strings.Contains(html, "Invalid email") {
		t.Errorf("expected error message, got: %s", html)
	}

	// Check role="alert"
	if !strings.Contains(html, `role="alert"`) {
		t.Errorf("expected role=alert, got: %s", html)
	}

	// Check custom class
	if !strings.Contains(html, "custom-field") {
		t.Errorf("expected custom class, got: %s", html)
	}

	// Check input is present
	if !strings.Contains(html, `type="email"`) {
		t.Errorf("expected email input, got: %s", html)
	}
}

func TestFormField_NoLabel(t *testing.T) {
	html := renderHTML(FormField(FormFieldProps{
		Children: []core.Node{
			Input(InputProps{Name: "search"}),
		},
	}))

	// Should NOT contain label element
	if strings.Contains(html, "<label") {
		t.Errorf("should not have label when Label is empty, got: %s", html)
	}

	// Should still contain the input
	if !strings.Contains(html, "<input") {
		t.Errorf("expected input element, got: %s", html)
	}
}

func TestFormField_RequiredWithoutLabel(t *testing.T) {
	html := renderHTML(FormField(FormFieldProps{
		Required: true,
		Children: []core.Node{
			Input(InputProps{Name: "field"}),
		},
	}))

	// Should NOT contain asterisk when no label
	// (asterisk is part of label rendering)
	if strings.Contains(html, "<span") && strings.Contains(html, "*") {
		// Check if the asterisk span is present
		if strings.Contains(html, "text-red-500") && strings.Contains(html, "*</span>") {
			t.Errorf("asterisk should not appear without label, got: %s", html)
		}
	}
}
