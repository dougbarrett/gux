package ui

import (
	"strings"
	"testing"
)

func TestSelect_DefaultValues(t *testing.T) {
	sel := Select(SelectProps{
		Name: "country",
		Options: []SelectOption{
			{Value: "us", Label: "United States"},
		},
	})
	html := renderHTML(sel)

	// Should render as select element
	if !strings.HasPrefix(html, "<select") {
		t.Errorf("expected <select> element, got: %s", html)
	}

	// Should have medium size classes (default)
	if !strings.Contains(html, "px-3 py-2") {
		t.Errorf("expected medium size classes, got: %s", html)
	}

	// Should have base styling
	if !strings.Contains(html, "rounded-md") {
		t.Errorf("expected base styling classes, got: %s", html)
	}

	// Should have appearance-none for custom styling
	if !strings.Contains(html, "appearance-none") {
		t.Errorf("expected appearance-none class, got: %s", html)
	}
}

func TestSelect_AllSizes(t *testing.T) {
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
			sel := Select(SelectProps{
				Size: tt.size,
				Name: "test",
				Options: []SelectOption{
					{Value: "a", Label: "Option A"},
				},
			})
			html := renderHTML(sel)

			if !strings.Contains(html, tt.expectedClass) {
				t.Errorf("size %s: expected class %q in: %s", tt.size, tt.expectedClass, html)
			}
		})
	}
}

func TestSelect_Options(t *testing.T) {
	sel := Select(SelectProps{
		Name: "country",
		Options: []SelectOption{
			{Value: "us", Label: "United States"},
			{Value: "uk", Label: "United Kingdom"},
			{Value: "ca", Label: "Canada"},
		},
	})
	html := renderHTML(sel)

	// Should contain all options
	if !strings.Contains(html, `value="us"`) {
		t.Errorf("expected value=\"us\", got: %s", html)
	}
	if !strings.Contains(html, "United States") {
		t.Errorf("expected United States label, got: %s", html)
	}
	if !strings.Contains(html, `value="uk"`) {
		t.Errorf("expected value=\"uk\", got: %s", html)
	}
	if !strings.Contains(html, "United Kingdom") {
		t.Errorf("expected United Kingdom label, got: %s", html)
	}
	if !strings.Contains(html, `value="ca"`) {
		t.Errorf("expected value=\"ca\", got: %s", html)
	}
	if !strings.Contains(html, "Canada") {
		t.Errorf("expected Canada label, got: %s", html)
	}

	// Should close select tag
	if !strings.HasSuffix(html, "</select>") {
		t.Errorf("expected closing </select>, got: %s", html)
	}
}

func TestSelect_SelectedOption(t *testing.T) {
	sel := Select(SelectProps{
		Name:  "country",
		Value: "uk",
		Options: []SelectOption{
			{Value: "us", Label: "United States"},
			{Value: "uk", Label: "United Kingdom"},
			{Value: "ca", Label: "Canada"},
		},
	})
	html := renderHTML(sel)

	// UK option should be selected
	// The selected attribute should be near the uk value
	if !strings.Contains(html, `value="uk"`) {
		t.Errorf("expected uk option, got: %s", html)
	}

	// Count selected attributes - should be exactly one
	count := strings.Count(html, `selected="selected"`)
	if count != 1 {
		t.Errorf("expected exactly 1 selected attribute, got %d in: %s", count, html)
	}
}

func TestSelect_Placeholder(t *testing.T) {
	sel := Select(SelectProps{
		Name:        "country",
		Placeholder: "Select a country",
		Options: []SelectOption{
			{Value: "us", Label: "United States"},
		},
	})
	html := renderHTML(sel)

	// Should have placeholder text
	if !strings.Contains(html, "Select a country") {
		t.Errorf("expected placeholder text, got: %s", html)
	}

	// Placeholder should be disabled
	if !strings.Contains(html, `value=""`+` disabled="disabled"`) && !strings.Contains(html, `disabled="disabled"`+` value=""`) {
		// Check that the first option (placeholder) is disabled
		// Find the first option tag
		optStart := strings.Index(html, "<option")
		if optStart == -1 {
			t.Errorf("expected option element, got: %s", html)
			return
		}
		optEnd := strings.Index(html[optStart:], ">")
		firstOption := html[optStart : optStart+optEnd+1]
		if !strings.Contains(firstOption, `disabled="disabled"`) {
			t.Errorf("expected placeholder option to be disabled, first option: %s", firstOption)
		}
	}
}

func TestSelect_PlaceholderSelected(t *testing.T) {
	sel := Select(SelectProps{
		Name:        "country",
		Value:       "", // No value selected
		Placeholder: "Select a country",
		Options: []SelectOption{
			{Value: "us", Label: "United States"},
		},
	})
	html := renderHTML(sel)

	// Placeholder should be selected when value is empty
	// Find the placeholder option (value="")
	optStart := strings.Index(html, "<option")
	if optStart == -1 {
		t.Errorf("expected option element, got: %s", html)
		return
	}
	optEnd := strings.Index(html[optStart:], ">")
	firstOption := html[optStart : optStart+optEnd+1]

	if !strings.Contains(firstOption, `selected="selected"`) {
		t.Errorf("expected placeholder to be selected when value is empty, first option: %s", firstOption)
	}
}

func TestSelect_PlaceholderNotSelectedWhenValueSet(t *testing.T) {
	sel := Select(SelectProps{
		Name:        "country",
		Value:       "us",
		Placeholder: "Select a country",
		Options: []SelectOption{
			{Value: "us", Label: "United States"},
		},
	})
	html := renderHTML(sel)

	// Placeholder should NOT be selected when a value is set
	// Find the placeholder option (value="")
	optStart := strings.Index(html, "<option")
	if optStart == -1 {
		t.Errorf("expected option element, got: %s", html)
		return
	}
	optEnd := strings.Index(html[optStart:], ">")
	firstOption := html[optStart : optStart+optEnd+1]

	// First option (placeholder) should NOT have selected
	if strings.Contains(firstOption, `selected="selected"`) && strings.Contains(firstOption, `value=""`) {
		t.Errorf("placeholder should NOT be selected when value is set, first option: %s", firstOption)
	}

	// Only one selected should exist in the whole HTML
	count := strings.Count(html, `selected="selected"`)
	if count != 1 {
		t.Errorf("expected exactly 1 selected attribute (on us option), got %d in: %s", count, html)
	}
}

func TestSelect_DisabledOption(t *testing.T) {
	sel := Select(SelectProps{
		Name: "country",
		Options: []SelectOption{
			{Value: "us", Label: "United States"},
			{Value: "uk", Label: "United Kingdom", Disabled: true},
			{Value: "ca", Label: "Canada"},
		},
	})
	html := renderHTML(sel)

	// UK option should be disabled
	// We need to check that the uk option has disabled attribute
	// Find UK option
	ukStart := strings.Index(html, `value="uk"`)
	if ukStart == -1 {
		t.Errorf("expected uk option, got: %s", html)
		return
	}

	// Find the <option that contains this value
	optStart := strings.LastIndex(html[:ukStart], "<option")
	nextOption := strings.Index(html[ukStart:], "</option>")
	ukOption := html[optStart : ukStart+nextOption]

	if !strings.Contains(ukOption, `disabled="disabled"`) {
		t.Errorf("expected UK option to be disabled, got: %s", ukOption)
	}
}

func TestSelect_Disabled(t *testing.T) {
	sel := Select(SelectProps{
		Name:     "country",
		Disabled: true,
		Options: []SelectOption{
			{Value: "us", Label: "United States"},
		},
	})
	html := renderHTML(sel)

	// Should have disabled attribute on select element
	// Check the select tag itself has disabled
	selectEnd := strings.Index(html, ">")
	selectTag := html[:selectEnd+1]

	if !strings.Contains(selectTag, `disabled="disabled"`) {
		t.Errorf("expected disabled attribute on select, got: %s", selectTag)
	}

	// Should have disabled styling classes
	if !strings.Contains(html, "opacity-50") {
		t.Errorf("expected opacity-50 class, got: %s", html)
	}
	if !strings.Contains(html, "cursor-not-allowed") {
		t.Errorf("expected cursor-not-allowed class, got: %s", html)
	}
}

func TestSelect_Required(t *testing.T) {
	sel := Select(SelectProps{
		Name:     "country",
		Required: true,
		Options: []SelectOption{
			{Value: "us", Label: "United States"},
		},
	})
	html := renderHTML(sel)

	// Should have required attribute
	if !strings.Contains(html, `required="required"`) {
		t.Errorf("expected required attribute, got: %s", html)
	}

	// Should have aria-required for accessibility
	if !strings.Contains(html, `aria-required="true"`) {
		t.Errorf("expected aria-required attribute, got: %s", html)
	}
}

func TestSelect_Error(t *testing.T) {
	sel := Select(SelectProps{
		Name:  "country",
		Error: "Please select a country",
		Options: []SelectOption{
			{Value: "us", Label: "United States"},
		},
	})
	html := renderHTML(sel)

	// Should have aria-invalid for accessibility
	if !strings.Contains(html, `aria-invalid="true"`) {
		t.Errorf("expected aria-invalid attribute, got: %s", html)
	}

	// Should have error styling classes
	if !strings.Contains(html, "border-red-500") {
		t.Errorf("expected border-red-500 class, got: %s", html)
	}
}

func TestSelect_CustomClass(t *testing.T) {
	sel := Select(SelectProps{
		Name:  "country",
		Class: "my-custom-class w-64",
		Options: []SelectOption{
			{Value: "us", Label: "United States"},
		},
	})
	html := renderHTML(sel)

	// Should include custom classes
	if !strings.Contains(html, "my-custom-class") {
		t.Errorf("expected custom class, got: %s", html)
	}
	if !strings.Contains(html, "w-64") {
		t.Errorf("expected w-64 class, got: %s", html)
	}

	// Should still have base classes
	if !strings.Contains(html, "rounded-md") {
		t.Errorf("expected base classes, got: %s", html)
	}
}

func TestSelect_ID(t *testing.T) {
	sel := Select(SelectProps{
		ID:   "country-select",
		Name: "country",
		Options: []SelectOption{
			{Value: "us", Label: "United States"},
		},
	})
	html := renderHTML(sel)

	if !strings.Contains(html, `id="country-select"`) {
		t.Errorf("expected id attribute, got: %s", html)
	}
}

func TestSelect_CombinedProps(t *testing.T) {
	sel := Select(SelectProps{
		ID:          "country-field",
		Name:        "country",
		Size:        InputLG,
		Value:       "uk",
		Placeholder: "Choose country",
		Class:       "extra-class",
		Required:    true,
		Error:       "Required field",
		Options: []SelectOption{
			{Value: "us", Label: "United States"},
			{Value: "uk", Label: "United Kingdom"},
			{Value: "ca", Label: "Canada", Disabled: true},
		},
	})
	html := renderHTML(sel)

	// Check ID
	if !strings.Contains(html, `id="country-field"`) {
		t.Errorf("expected id, got: %s", html)
	}

	// Check name
	if !strings.Contains(html, `name="country"`) {
		t.Errorf("expected name, got: %s", html)
	}

	// Check size
	if !strings.Contains(html, "px-4 py-3") {
		t.Errorf("expected large size, got: %s", html)
	}

	// Check placeholder exists
	if !strings.Contains(html, "Choose country") {
		t.Errorf("expected placeholder text, got: %s", html)
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

	// Check all options present
	if !strings.Contains(html, "United States") {
		t.Errorf("expected United States option, got: %s", html)
	}
	if !strings.Contains(html, "United Kingdom") {
		t.Errorf("expected United Kingdom option, got: %s", html)
	}
	if !strings.Contains(html, "Canada") {
		t.Errorf("expected Canada option, got: %s", html)
	}

	// Check UK is selected (only one selected)
	count := strings.Count(html, `selected="selected"`)
	if count != 1 {
		t.Errorf("expected exactly 1 selected (UK), got %d in: %s", count, html)
	}
}
