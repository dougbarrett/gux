package ui

import (
	"strings"
	"testing"
)

func TestTextarea_DefaultValues(t *testing.T) {
	ta := Textarea(TextareaProps{
		Name: "description",
	})
	html := renderHTML(ta)

	// Should render as textarea element
	if !strings.HasPrefix(html, "<textarea") {
		t.Errorf("expected <textarea> element, got: %s", html)
	}

	// Should have medium size classes (default)
	if !strings.Contains(html, "px-3 py-2") {
		t.Errorf("expected medium size classes, got: %s", html)
	}

	// Should have vertical resize (default)
	if !strings.Contains(html, "resize-y") {
		t.Errorf("expected resize-y class, got: %s", html)
	}

	// Should have 3 rows (default)
	if !strings.Contains(html, `rows="3"`) {
		t.Errorf("expected rows=\"3\", got: %s", html)
	}
}

func TestTextarea_AllSizes(t *testing.T) {
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
			ta := Textarea(TextareaProps{
				Size: tt.size,
				Name: "test",
			})
			html := renderHTML(ta)

			if !strings.Contains(html, tt.expectedClass) {
				t.Errorf("size %s: expected class %q in: %s", tt.size, tt.expectedClass, html)
			}
		})
	}
}

func TestTextarea_ResizeOptions(t *testing.T) {
	tests := []struct {
		resize        TextareaResize
		expectedClass string
	}{
		{ResizeNone, "resize-none"},
		{ResizeVertical, "resize-y"},
		{ResizeBoth, "resize"},
	}

	for _, tt := range tests {
		t.Run(string(tt.resize), func(t *testing.T) {
			ta := Textarea(TextareaProps{
				Resize: tt.resize,
				Name:   "test",
			})
			html := renderHTML(ta)

			if !strings.Contains(html, tt.expectedClass) {
				t.Errorf("resize %s: expected class %q in: %s", tt.resize, tt.expectedClass, html)
			}
		})
	}
}

func TestTextarea_Rows(t *testing.T) {
	tests := []struct {
		name         string
		rows         int
		expectedRows string
	}{
		{"explicit 5", 5, `rows="5"`},
		{"explicit 10", 10, `rows="10"`},
		{"zero defaults to 3", 0, `rows="3"`},
		{"negative defaults to 3", -1, `rows="3"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := Textarea(TextareaProps{
				Rows: tt.rows,
				Name: "test",
			})
			html := renderHTML(ta)

			if !strings.Contains(html, tt.expectedRows) {
				t.Errorf("rows %d: expected %q in: %s", tt.rows, tt.expectedRows, html)
			}
		})
	}
}

func TestTextarea_Placeholder(t *testing.T) {
	ta := Textarea(TextareaProps{
		Name:        "description",
		Placeholder: "Enter your message...",
	})
	html := renderHTML(ta)

	if !strings.Contains(html, `placeholder="Enter your message..."`) {
		t.Errorf("expected placeholder attribute, got: %s", html)
	}
}

func TestTextarea_Disabled(t *testing.T) {
	ta := Textarea(TextareaProps{
		Name:     "description",
		Disabled: true,
	})
	html := renderHTML(ta)

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

func TestTextarea_Required(t *testing.T) {
	ta := Textarea(TextareaProps{
		Name:     "description",
		Required: true,
	})
	html := renderHTML(ta)

	// Should have required attribute
	if !strings.Contains(html, `required="required"`) {
		t.Errorf("expected required attribute, got: %s", html)
	}

	// Should have aria-required for accessibility
	if !strings.Contains(html, `aria-required="true"`) {
		t.Errorf("expected aria-required attribute, got: %s", html)
	}
}

func TestTextarea_Error(t *testing.T) {
	ta := Textarea(TextareaProps{
		Name:  "description",
		Error: "This field is required",
	})
	html := renderHTML(ta)

	// Should have aria-invalid for accessibility
	if !strings.Contains(html, `aria-invalid="true"`) {
		t.Errorf("expected aria-invalid attribute, got: %s", html)
	}

	// Should have error styling classes
	if !strings.Contains(html, "border-red-500") {
		t.Errorf("expected border-red-500 class, got: %s", html)
	}
}

func TestTextarea_Value(t *testing.T) {
	ta := Textarea(TextareaProps{
		Name:  "description",
		Value: "Hello, World!",
	})
	html := renderHTML(ta)

	// Value should render as textarea content for SSR
	if !strings.Contains(html, "Hello, World!") {
		t.Errorf("expected value as content, got: %s", html)
	}

	// Should close textarea tag properly
	if !strings.HasSuffix(html, "</textarea>") {
		t.Errorf("expected closing </textarea>, got: %s", html)
	}
}

func TestTextarea_CustomClass(t *testing.T) {
	ta := Textarea(TextareaProps{
		Name:  "description",
		Class: "my-custom-class min-h-32",
	})
	html := renderHTML(ta)

	// Should include custom classes
	if !strings.Contains(html, "my-custom-class") {
		t.Errorf("expected custom class, got: %s", html)
	}
	if !strings.Contains(html, "min-h-32") {
		t.Errorf("expected min-h-32 class, got: %s", html)
	}

	// Should still have base classes
	if !strings.Contains(html, "rounded-md") {
		t.Errorf("expected base classes, got: %s", html)
	}
}

func TestTextarea_ID(t *testing.T) {
	ta := Textarea(TextareaProps{
		ID:   "description-field",
		Name: "description",
	})
	html := renderHTML(ta)

	if !strings.Contains(html, `id="description-field"`) {
		t.Errorf("expected id attribute, got: %s", html)
	}
}

func TestTextarea_CombinedProps(t *testing.T) {
	ta := Textarea(TextareaProps{
		ID:          "desc",
		Name:        "description",
		Size:        InputLG,
		Resize:      ResizeNone,
		Rows:        6,
		Value:       "Test content",
		Placeholder: "Enter here",
		Class:       "extra-class",
		Required:    true,
		Error:       "Some error",
	})
	html := renderHTML(ta)

	// Check ID
	if !strings.Contains(html, `id="desc"`) {
		t.Errorf("expected id, got: %s", html)
	}

	// Check name
	if !strings.Contains(html, `name="description"`) {
		t.Errorf("expected name, got: %s", html)
	}

	// Check size
	if !strings.Contains(html, "px-4 py-3") {
		t.Errorf("expected large size, got: %s", html)
	}

	// Check resize
	if !strings.Contains(html, "resize-none") {
		t.Errorf("expected resize-none, got: %s", html)
	}

	// Check rows
	if !strings.Contains(html, `rows="6"`) {
		t.Errorf("expected rows=6, got: %s", html)
	}

	// Check value as content
	if !strings.Contains(html, "Test content") {
		t.Errorf("expected value content, got: %s", html)
	}

	// Check placeholder
	if !strings.Contains(html, `placeholder="Enter here"`) {
		t.Errorf("expected placeholder, got: %s", html)
	}

	// Check custom class
	if !strings.Contains(html, "extra-class") {
		t.Errorf("expected extra-class, got: %s", html)
	}

	// Check required
	if !strings.Contains(html, `required="required"`) {
		t.Errorf("expected required, got: %s", html)
	}

	// Check error styling
	if !strings.Contains(html, "border-red-500") {
		t.Errorf("expected error styling, got: %s", html)
	}
	if !strings.Contains(html, `aria-invalid="true"`) {
		t.Errorf("expected aria-invalid, got: %s", html)
	}
}
