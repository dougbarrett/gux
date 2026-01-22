package ui

import (
	"strings"
	"testing"
)

func TestRadioGroup_Basic(t *testing.T) {
	rg := RadioGroup(RadioGroupProps{
		Name: "status",
		Options: []RadioOptionProps{
			{Value: "a", Label: "Option A"},
			{Value: "b", Label: "Option B"},
		},
	})
	html := renderHTML(rg)

	// Should render as div with radiogroup role
	if !strings.HasPrefix(html, "<div") {
		t.Errorf("expected <div> wrapper, got: %s", html)
	}
	if !strings.Contains(html, `role="radiogroup"`) {
		t.Errorf("expected role=\"radiogroup\", got: %s", html)
	}

	// Should contain radio inputs
	if !strings.Contains(html, `type="radio"`) {
		t.Errorf("expected type=\"radio\" inputs, got: %s", html)
	}

	// Should contain option labels
	if !strings.Contains(html, "Option A") {
		t.Errorf("expected Option A label, got: %s", html)
	}
	if !strings.Contains(html, "Option B") {
		t.Errorf("expected Option B label, got: %s", html)
	}
}

func TestRadioGroup_SelectedValue(t *testing.T) {
	rg := RadioGroup(RadioGroupProps{
		Name:  "status",
		Value: "b",
		Options: []RadioOptionProps{
			{Value: "a", Label: "Option A"},
			{Value: "b", Label: "Option B"},
			{Value: "c", Label: "Option C"},
		},
	})
	html := renderHTML(rg)

	// CRITICAL: Only one option should have checked attribute
	checkedCount := strings.Count(html, `checked="checked"`)
	if checkedCount != 1 {
		t.Errorf("expected exactly 1 checked attribute, got %d in: %s", checkedCount, html)
	}

	// The checked input should have value="b"
	// Check that value="b" appears near checked
	if !strings.Contains(html, `value="b"`) {
		t.Errorf("expected value=\"b\" in: %s", html)
	}
}

func TestRadioGroup_NoSelection(t *testing.T) {
	rg := RadioGroup(RadioGroupProps{
		Name:  "status",
		Value: "", // No selection
		Options: []RadioOptionProps{
			{Value: "a", Label: "Option A"},
			{Value: "b", Label: "Option B"},
		},
	})
	html := renderHTML(rg)

	// CRITICAL: No options should have checked attribute
	if strings.Contains(html, "checked") {
		t.Errorf("expected NO checked attributes when Value=\"\", got: %s", html)
	}
}

func TestRadioGroup_Inline(t *testing.T) {
	rg := RadioGroup(RadioGroupProps{
		Name:   "status",
		Inline: true,
		Options: []RadioOptionProps{
			{Value: "a", Label: "Option A"},
			{Value: "b", Label: "Option B"},
		},
	})
	html := renderHTML(rg)

	// Should have flex-row for inline layout
	if !strings.Contains(html, "flex-row") {
		t.Errorf("expected flex-row class for inline layout, got: %s", html)
	}
}

func TestRadioGroup_Stacked(t *testing.T) {
	rg := RadioGroup(RadioGroupProps{
		Name:   "status",
		Inline: false, // Default stacked
		Options: []RadioOptionProps{
			{Value: "a", Label: "Option A"},
			{Value: "b", Label: "Option B"},
		},
	})
	html := renderHTML(rg)

	// Should have flex-col for stacked layout (default)
	if !strings.Contains(html, "flex-col") {
		t.Errorf("expected flex-col class for stacked layout, got: %s", html)
	}
}

func TestRadioGroup_DisabledGroup(t *testing.T) {
	rg := RadioGroup(RadioGroupProps{
		Name:     "status",
		Disabled: true,
		Options: []RadioOptionProps{
			{Value: "a", Label: "Option A"},
			{Value: "b", Label: "Option B"},
		},
	})
	html := renderHTML(rg)

	// All options should have disabled attribute
	disabledCount := strings.Count(html, `disabled="disabled"`)
	if disabledCount != 2 {
		t.Errorf("expected 2 disabled attributes (all options), got %d in: %s", disabledCount, html)
	}

	// Should have disabled styling
	if !strings.Contains(html, "opacity-50") {
		t.Errorf("expected opacity-50 class for disabled, got: %s", html)
	}
}

func TestRadioGroup_DisabledOption(t *testing.T) {
	rg := RadioGroup(RadioGroupProps{
		Name: "status",
		Options: []RadioOptionProps{
			{Value: "a", Label: "Option A"},
			{Value: "b", Label: "Option B", Disabled: true},
			{Value: "c", Label: "Option C"},
		},
	})
	html := renderHTML(rg)

	// Only one option should have disabled attribute
	disabledCount := strings.Count(html, `disabled="disabled"`)
	if disabledCount != 1 {
		t.Errorf("expected 1 disabled attribute (single option), got %d in: %s", disabledCount, html)
	}
}

func TestRadioGroup_Required(t *testing.T) {
	rg := RadioGroup(RadioGroupProps{
		Name:     "status",
		Required: true,
		Options: []RadioOptionProps{
			{Value: "a", Label: "Option A"},
			{Value: "b", Label: "Option B"},
		},
	})
	html := renderHTML(rg)

	// Required attribute should be on radio inputs
	if !strings.Contains(html, `required="required"`) {
		t.Errorf("expected required attribute, got: %s", html)
	}
}

func TestRadioGroup_WithDescriptions(t *testing.T) {
	rg := RadioGroup(RadioGroupProps{
		Name: "status",
		Options: []RadioOptionProps{
			{Value: "a", Label: "Option A", Description: "First choice"},
			{Value: "b", Label: "Option B", Description: "Second choice"},
		},
	})
	html := renderHTML(rg)

	// Should contain description text
	if !strings.Contains(html, "First choice") {
		t.Errorf("expected description \"First choice\", got: %s", html)
	}
	if !strings.Contains(html, "Second choice") {
		t.Errorf("expected description \"Second choice\", got: %s", html)
	}

	// Description should have appropriate styling
	if !strings.Contains(html, "text-gray-500") {
		t.Errorf("expected description styling (text-gray-500), got: %s", html)
	}
}

func TestRadioGroup_SameName(t *testing.T) {
	rg := RadioGroup(RadioGroupProps{
		Name: "my-radio-group",
		Options: []RadioOptionProps{
			{Value: "a", Label: "Option A"},
			{Value: "b", Label: "Option B"},
			{Value: "c", Label: "Option C"},
		},
	})
	html := renderHTML(rg)

	// All radios should have the same name attribute
	nameCount := strings.Count(html, `name="my-radio-group"`)
	if nameCount != 3 {
		t.Errorf("expected all 3 radios to have name=\"my-radio-group\", got %d in: %s", nameCount, html)
	}
}

func TestRadioGroup_CustomClass(t *testing.T) {
	rg := RadioGroup(RadioGroupProps{
		Name:  "status",
		Class: "my-custom-class mt-4",
		Options: []RadioOptionProps{
			{Value: "a", Label: "Option A"},
		},
	})
	html := renderHTML(rg)

	// Custom class should be on the wrapper
	if !strings.Contains(html, "my-custom-class") {
		t.Errorf("expected custom class on wrapper, got: %s", html)
	}
	if !strings.Contains(html, "mt-4") {
		t.Errorf("expected mt-4 class on wrapper, got: %s", html)
	}
}

func TestRadioGroup_Error(t *testing.T) {
	rg := RadioGroup(RadioGroupProps{
		Name:  "status",
		Error: "Please select an option",
		Options: []RadioOptionProps{
			{Value: "a", Label: "Option A"},
		},
	})
	html := renderHTML(rg)

	// Should have aria-invalid on wrapper
	if !strings.Contains(html, `aria-invalid="true"`) {
		t.Errorf("expected aria-invalid on wrapper, got: %s", html)
	}
}

func TestRadioGroup_CombinedProps(t *testing.T) {
	rg := RadioGroup(RadioGroupProps{
		Name:     "priority",
		Value:    "high",
		Inline:   true,
		Class:    "mt-2",
		Required: true,
		Options: []RadioOptionProps{
			{Value: "low", Label: "Low"},
			{Value: "medium", Label: "Medium", Description: "Recommended"},
			{Value: "high", Label: "High"},
		},
	})
	html := renderHTML(rg)

	// Check name
	if !strings.Contains(html, `name="priority"`) {
		t.Errorf("expected name=\"priority\", got: %s", html)
	}

	// Check selected value (high)
	checkedCount := strings.Count(html, `checked="checked"`)
	if checkedCount != 1 {
		t.Errorf("expected 1 checked attribute, got %d", checkedCount)
	}

	// Check inline layout
	if !strings.Contains(html, "flex-row") {
		t.Errorf("expected flex-row for inline, got: %s", html)
	}

	// Check custom class
	if !strings.Contains(html, "mt-2") {
		t.Errorf("expected mt-2 class, got: %s", html)
	}

	// Check required
	if !strings.Contains(html, `required="required"`) {
		t.Errorf("expected required attribute, got: %s", html)
	}

	// Check description
	if !strings.Contains(html, "Recommended") {
		t.Errorf("expected description \"Recommended\", got: %s", html)
	}
}
