package ui

import (
	"strings"
	"testing"
)

// Note: renderHTML helper is defined in button_test.go

func TestInput_DefaultValues(t *testing.T) {
	input := Input(InputProps{})
	html := renderHTML(input)

	// Should render as input element
	if !strings.HasPrefix(html, "<input") {
		t.Errorf("expected <input> element, got: %s", html)
	}

	// Should have type=text (default)
	if !strings.Contains(html, `type="text"`) {
		t.Errorf("expected type=\"text\", got: %s", html)
	}

	// Should have medium size classes (default)
	if !strings.Contains(html, "px-3 py-2") {
		t.Errorf("expected medium size classes (px-3 py-2), got: %s", html)
	}
	if !strings.Contains(html, "text-base") {
		t.Errorf("expected text-base class, got: %s", html)
	}

	// Should have base classes
	if !strings.Contains(html, "w-full") {
		t.Errorf("expected w-full class, got: %s", html)
	}
	if !strings.Contains(html, "rounded-md") {
		t.Errorf("expected rounded-md class, got: %s", html)
	}
}

func TestInput_AllTypes(t *testing.T) {
	tests := []struct {
		inputType    InputType
		expectedType string
	}{
		{InputText, "text"},
		{InputEmail, "email"},
		{InputPassword, "password"},
		{InputNumber, "number"},
		{InputSearch, "search"},
		{InputTel, "tel"},
		{InputURL, "url"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedType, func(t *testing.T) {
			input := Input(InputProps{Type: tt.inputType})
			html := renderHTML(input)

			expected := `type="` + tt.expectedType + `"`
			if !strings.Contains(html, expected) {
				t.Errorf("expected %s, got: %s", expected, html)
			}
		})
	}
}

func TestInput_AllSizes(t *testing.T) {
	tests := []struct {
		size          InputSize
		expectedClass string
	}{
		{InputSM, "px-2 py-1.5 text-sm"},
		{InputMD, "px-3 py-2 text-base"},
		{InputLG, "px-4 py-3 text-lg"},
	}

	for _, tt := range tests {
		t.Run(string(tt.size), func(t *testing.T) {
			input := Input(InputProps{Size: tt.size})
			html := renderHTML(input)

			if !strings.Contains(html, tt.expectedClass) {
				t.Errorf("size %s: expected class %q in: %s", tt.size, tt.expectedClass, html)
			}
		})
	}
}

func TestInput_Placeholder(t *testing.T) {
	input := Input(InputProps{
		Placeholder: "Enter your name",
	})
	html := renderHTML(input)

	if !strings.Contains(html, `placeholder="Enter your name"`) {
		t.Errorf("expected placeholder attribute, got: %s", html)
	}
}

func TestInput_Disabled(t *testing.T) {
	input := Input(InputProps{
		Disabled: true,
	})
	html := renderHTML(input)

	// Should have disabled attribute
	if !strings.Contains(html, `disabled="disabled"`) {
		t.Errorf("expected disabled attribute, got: %s", html)
	}

	// Should have disabled styling classes
	if !strings.Contains(html, "bg-gray-100") {
		t.Errorf("expected bg-gray-100 class, got: %s", html)
	}
	if !strings.Contains(html, "cursor-not-allowed") {
		t.Errorf("expected cursor-not-allowed class, got: %s", html)
	}
	if !strings.Contains(html, "opacity-50") {
		t.Errorf("expected opacity-50 class, got: %s", html)
	}
}

func TestInput_Required(t *testing.T) {
	input := Input(InputProps{
		Required: true,
	})
	html := renderHTML(input)

	// Should have required attribute
	if !strings.Contains(html, `required="required"`) {
		t.Errorf("expected required attribute, got: %s", html)
	}

	// Should have aria-required attribute
	if !strings.Contains(html, `aria-required="true"`) {
		t.Errorf("expected aria-required attribute, got: %s", html)
	}
}

func TestInput_Error(t *testing.T) {
	input := Input(InputProps{
		Error: "This field is required",
	})
	html := renderHTML(input)

	// Should have aria-invalid attribute
	if !strings.Contains(html, `aria-invalid="true"`) {
		t.Errorf("expected aria-invalid attribute, got: %s", html)
	}

	// Should have error styling classes
	if !strings.Contains(html, "border-red-500") {
		t.Errorf("expected border-red-500 class, got: %s", html)
	}
	if !strings.Contains(html, "focus:ring-red-500") {
		t.Errorf("expected focus:ring-red-500 class, got: %s", html)
	}
}

func TestInput_CustomClass(t *testing.T) {
	input := Input(InputProps{
		Class: "my-custom-class w-1/2",
	})
	html := renderHTML(input)

	// Should include custom classes
	if !strings.Contains(html, "my-custom-class") {
		t.Errorf("expected custom class, got: %s", html)
	}
	if !strings.Contains(html, "w-1/2") {
		t.Errorf("expected w-1/2 class, got: %s", html)
	}

	// Should still have base classes
	if !strings.Contains(html, "rounded-md") {
		t.Errorf("expected base classes, got: %s", html)
	}
}

func TestInput_ID(t *testing.T) {
	input := Input(InputProps{
		ID: "email-input",
	})
	html := renderHTML(input)

	if !strings.Contains(html, `id="email-input"`) {
		t.Errorf("expected id attribute, got: %s", html)
	}
}

func TestInput_NameAndValue(t *testing.T) {
	input := Input(InputProps{
		Name:  "email",
		Value: "test@example.com",
	})
	html := renderHTML(input)

	if !strings.Contains(html, `name="email"`) {
		t.Errorf("expected name attribute, got: %s", html)
	}
	if !strings.Contains(html, `value="test@example.com"`) {
		t.Errorf("expected value attribute, got: %s", html)
	}
}

func TestInput_CombinedProps(t *testing.T) {
	input := Input(InputProps{
		Type:        InputEmail,
		Size:        InputLG,
		ID:          "contact-email",
		Name:        "email",
		Value:       "user@test.com",
		Placeholder: "Enter email",
		Class:       "extra-class",
		Required:    true,
		Error:       "Invalid email",
	})
	html := renderHTML(input)

	// Check type
	if !strings.Contains(html, `type="email"`) {
		t.Errorf("expected type=email, got: %s", html)
	}

	// Check size
	if !strings.Contains(html, "px-4 py-3") {
		t.Errorf("expected large size classes, got: %s", html)
	}

	// Check ID
	if !strings.Contains(html, `id="contact-email"`) {
		t.Errorf("expected id, got: %s", html)
	}

	// Check name and value
	if !strings.Contains(html, `name="email"`) {
		t.Errorf("expected name, got: %s", html)
	}
	if !strings.Contains(html, `value="user@test.com"`) {
		t.Errorf("expected value, got: %s", html)
	}

	// Check placeholder
	if !strings.Contains(html, `placeholder="Enter email"`) {
		t.Errorf("expected placeholder, got: %s", html)
	}

	// Check custom class
	if !strings.Contains(html, "extra-class") {
		t.Errorf("expected extra-class, got: %s", html)
	}

	// Check required attributes
	if !strings.Contains(html, `required="required"`) {
		t.Errorf("expected required, got: %s", html)
	}
	if !strings.Contains(html, `aria-required="true"`) {
		t.Errorf("expected aria-required, got: %s", html)
	}

	// Check error attributes
	if !strings.Contains(html, `aria-invalid="true"`) {
		t.Errorf("expected aria-invalid, got: %s", html)
	}
	if !strings.Contains(html, "border-red-500") {
		t.Errorf("expected error styling, got: %s", html)
	}
}
