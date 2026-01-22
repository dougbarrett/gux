package ui

import (
	"strings"
	"testing"
)

func TestCheckbox_Basic(t *testing.T) {
	checkbox := Checkbox(CheckboxProps{
		Name: "agree",
	})
	html := renderHTML(checkbox)

	// Should render as input type=checkbox
	if !strings.Contains(html, "<input") {
		t.Errorf("expected <input> element, got: %s", html)
	}
	if !strings.Contains(html, `type="checkbox"`) {
		t.Errorf("expected type=\"checkbox\", got: %s", html)
	}
}

func TestCheckbox_Checked(t *testing.T) {
	checkbox := Checkbox(CheckboxProps{
		Name:    "agree",
		Checked: true,
	})
	html := renderHTML(checkbox)

	// CRITICAL: checked attribute MUST be present when Checked=true
	if !strings.Contains(html, `checked="checked"`) {
		t.Errorf("expected checked=\"checked\" when Checked=true, got: %s", html)
	}
}

func TestCheckbox_Unchecked(t *testing.T) {
	checkbox := Checkbox(CheckboxProps{
		Name:    "agree",
		Checked: false,
	})
	html := renderHTML(checkbox)

	// CRITICAL: checked attribute MUST NOT be present when Checked=false
	if strings.Contains(html, "checked") {
		t.Errorf("expected NO checked attribute when Checked=false, got: %s", html)
	}
}

func TestCheckbox_Disabled(t *testing.T) {
	checkbox := Checkbox(CheckboxProps{
		Name:     "agree",
		Disabled: true,
	})
	html := renderHTML(checkbox)

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

func TestCheckbox_WithLabel(t *testing.T) {
	checkbox := Checkbox(CheckboxProps{
		Name:  "agree",
		Label: "I agree to the terms",
	})
	html := renderHTML(checkbox)

	// Should be wrapped in a label
	if !strings.HasPrefix(html, "<label") {
		t.Errorf("expected <label> wrapper, got: %s", html)
	}
	if !strings.HasSuffix(html, "</label>") {
		t.Errorf("expected closing </label>, got: %s", html)
	}

	// Should contain the label text
	if !strings.Contains(html, "I agree to the terms") {
		t.Errorf("expected label text, got: %s", html)
	}

	// Should still contain the input
	if !strings.Contains(html, `type="checkbox"`) {
		t.Errorf("expected checkbox input inside label, got: %s", html)
	}
}

func TestCheckbox_WithDescription(t *testing.T) {
	checkbox := Checkbox(CheckboxProps{
		Name:        "newsletter",
		Label:       "Subscribe to newsletter",
		Description: "Get weekly updates about new features",
	})
	html := renderHTML(checkbox)

	// Should contain description text
	if !strings.Contains(html, "Get weekly updates about new features") {
		t.Errorf("expected description text, got: %s", html)
	}

	// Description should have appropriate styling
	if !strings.Contains(html, "text-gray-500") {
		t.Errorf("expected description styling (text-gray-500), got: %s", html)
	}
}

func TestCheckbox_ID(t *testing.T) {
	checkbox := Checkbox(CheckboxProps{
		ID:   "my-checkbox",
		Name: "agree",
	})
	html := renderHTML(checkbox)

	if !strings.Contains(html, `id="my-checkbox"`) {
		t.Errorf("expected id attribute, got: %s", html)
	}
}

func TestCheckbox_Name(t *testing.T) {
	checkbox := Checkbox(CheckboxProps{
		Name: "agree_terms",
	})
	html := renderHTML(checkbox)

	if !strings.Contains(html, `name="agree_terms"`) {
		t.Errorf("expected name attribute, got: %s", html)
	}
}

func TestCheckbox_CustomClass(t *testing.T) {
	checkbox := Checkbox(CheckboxProps{
		Name:  "agree",
		Label: "Agree",
		Class: "my-custom-class extra-spacing",
	})
	html := renderHTML(checkbox)

	// Custom class should be on the wrapper
	if !strings.Contains(html, "my-custom-class") {
		t.Errorf("expected custom class on wrapper, got: %s", html)
	}
	if !strings.Contains(html, "extra-spacing") {
		t.Errorf("expected extra-spacing class on wrapper, got: %s", html)
	}
}

func TestCheckbox_DisabledWithLabel(t *testing.T) {
	checkbox := Checkbox(CheckboxProps{
		Name:     "agree",
		Label:    "Agree",
		Disabled: true,
	})
	html := renderHTML(checkbox)

	// Wrapper should have cursor-not-allowed
	// Check that the label wrapper has the disabled cursor
	if !strings.Contains(html, "cursor-not-allowed") {
		t.Errorf("expected cursor-not-allowed on wrapper, got: %s", html)
	}
}

func TestCheckbox_CheckedWithLabel(t *testing.T) {
	checkbox := Checkbox(CheckboxProps{
		Name:    "agree",
		Label:   "Agree",
		Checked: true,
	})
	html := renderHTML(checkbox)

	// Should be wrapped in label AND have checked attribute
	if !strings.HasPrefix(html, "<label") {
		t.Errorf("expected <label> wrapper, got: %s", html)
	}
	if !strings.Contains(html, `checked="checked"`) {
		t.Errorf("expected checked attribute, got: %s", html)
	}
}

func TestCheckbox_CombinedProps(t *testing.T) {
	checkbox := Checkbox(CheckboxProps{
		ID:          "terms-checkbox",
		Name:        "terms",
		Checked:     true,
		Label:       "Accept Terms",
		Description: "Read our terms of service",
		Class:       "mt-4",
		Disabled:    true,
	})
	html := renderHTML(checkbox)

	// Check ID
	if !strings.Contains(html, `id="terms-checkbox"`) {
		t.Errorf("expected id attribute, got: %s", html)
	}

	// Check name
	if !strings.Contains(html, `name="terms"`) {
		t.Errorf("expected name attribute, got: %s", html)
	}

	// Check checked
	if !strings.Contains(html, `checked="checked"`) {
		t.Errorf("expected checked attribute, got: %s", html)
	}

	// Check label text
	if !strings.Contains(html, "Accept Terms") {
		t.Errorf("expected label text, got: %s", html)
	}

	// Check description
	if !strings.Contains(html, "Read our terms of service") {
		t.Errorf("expected description text, got: %s", html)
	}

	// Check custom class
	if !strings.Contains(html, "mt-4") {
		t.Errorf("expected custom class, got: %s", html)
	}

	// Check disabled
	if !strings.Contains(html, `disabled="disabled"`) {
		t.Errorf("expected disabled attribute, got: %s", html)
	}
}
