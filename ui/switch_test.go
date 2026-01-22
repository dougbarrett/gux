package ui

import (
	"strings"
	"testing"
)

func TestSwitch_Basic(t *testing.T) {
	sw := Switch(SwitchProps{
		Name: "toggle",
	})
	html := renderHTML(sw)

	// Should be wrapped in a label
	if !strings.HasPrefix(html, "<label") {
		t.Errorf("expected <label> wrapper, got: %s", html)
	}

	// Should contain hidden checkbox input
	if !strings.Contains(html, `type="checkbox"`) {
		t.Errorf("expected hidden checkbox input, got: %s", html)
	}

	// Should contain visual switch with role="switch"
	if !strings.Contains(html, `role="switch"`) {
		t.Errorf("expected role=\"switch\" on visual element, got: %s", html)
	}
}

func TestSwitch_CheckedOn(t *testing.T) {
	sw := Switch(SwitchProps{
		Name:    "toggle",
		Checked: true,
	})
	html := renderHTML(sw)

	// Should have on styling (translate-x-6 for knob position)
	if !strings.Contains(html, "translate-x-6") {
		t.Errorf("expected translate-x-6 for on state, got: %s", html)
	}

	// Should have on color (bg-blue-600 for track)
	if !strings.Contains(html, "bg-blue-600") {
		t.Errorf("expected bg-blue-600 for on state, got: %s", html)
	}

	// Should have checked attribute on hidden input
	if !strings.Contains(html, `checked="checked"`) {
		t.Errorf("expected checked attribute on hidden input, got: %s", html)
	}
}

func TestSwitch_CheckedOff(t *testing.T) {
	sw := Switch(SwitchProps{
		Name:    "toggle",
		Checked: false,
	})
	html := renderHTML(sw)

	// Should have off styling (translate-x-1 for knob position)
	if !strings.Contains(html, "translate-x-1") {
		t.Errorf("expected translate-x-1 for off state, got: %s", html)
	}

	// Should have off color (bg-gray-300 for track)
	if !strings.Contains(html, "bg-gray-300") {
		t.Errorf("expected bg-gray-300 for off state, got: %s", html)
	}

	// Should NOT have checked attribute on hidden input
	if strings.Contains(html, `checked="checked"`) {
		t.Errorf("expected NO checked attribute when Checked=false, got: %s", html)
	}
}

func TestSwitch_HiddenInput(t *testing.T) {
	sw := Switch(SwitchProps{
		Name: "toggle",
	})
	html := renderHTML(sw)

	// Should have hidden input with sr-only class
	if !strings.Contains(html, "sr-only") {
		t.Errorf("expected sr-only class on hidden input, got: %s", html)
	}

	// Should have type="checkbox"
	if !strings.Contains(html, `type="checkbox"`) {
		t.Errorf("expected type=\"checkbox\" on hidden input, got: %s", html)
	}
}

func TestSwitch_Disabled(t *testing.T) {
	sw := Switch(SwitchProps{
		Name:     "toggle",
		Disabled: true,
	})
	html := renderHTML(sw)

	// Should have disabled attribute on hidden input
	if !strings.Contains(html, `disabled="disabled"`) {
		t.Errorf("expected disabled attribute on hidden input, got: %s", html)
	}

	// Should have disabled styling (opacity-50)
	if !strings.Contains(html, "opacity-50") {
		t.Errorf("expected opacity-50 class for disabled, got: %s", html)
	}

	// Should have cursor-not-allowed on wrapper
	if !strings.Contains(html, "cursor-not-allowed") {
		t.Errorf("expected cursor-not-allowed class for disabled, got: %s", html)
	}
}

func TestSwitch_WithLabel(t *testing.T) {
	sw := Switch(SwitchProps{
		Name:  "toggle",
		Label: "Enable feature",
	})
	html := renderHTML(sw)

	// Should contain label text
	if !strings.Contains(html, "Enable feature") {
		t.Errorf("expected label text, got: %s", html)
	}

	// Label should have appropriate styling
	if !strings.Contains(html, "text-gray-700") {
		t.Errorf("expected text-gray-700 for label styling, got: %s", html)
	}
}

func TestSwitch_AriaChecked(t *testing.T) {
	tests := []struct {
		name        string
		checked     bool
		expectedVal string
	}{
		{"checked true", true, `aria-checked="true"`},
		{"checked false", false, `aria-checked="false"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sw := Switch(SwitchProps{
				Name:    "toggle",
				Checked: tt.checked,
			})
			html := renderHTML(sw)

			if !strings.Contains(html, tt.expectedVal) {
				t.Errorf("expected %s, got: %s", tt.expectedVal, html)
			}
		})
	}
}

func TestSwitch_Name(t *testing.T) {
	sw := Switch(SwitchProps{
		Name: "my_toggle_setting",
	})
	html := renderHTML(sw)

	if !strings.Contains(html, `name="my_toggle_setting"`) {
		t.Errorf("expected name attribute on hidden input, got: %s", html)
	}
}

func TestSwitch_ID(t *testing.T) {
	sw := Switch(SwitchProps{
		ID:   "toggle-id",
		Name: "toggle",
	})
	html := renderHTML(sw)

	if !strings.Contains(html, `id="toggle-id"`) {
		t.Errorf("expected id attribute on hidden input, got: %s", html)
	}
}

func TestSwitch_CustomClass(t *testing.T) {
	sw := Switch(SwitchProps{
		Name:  "toggle",
		Class: "my-custom-class mt-4",
	})
	html := renderHTML(sw)

	// Custom class should be on the wrapper
	if !strings.Contains(html, "my-custom-class") {
		t.Errorf("expected custom class on wrapper, got: %s", html)
	}
	if !strings.Contains(html, "mt-4") {
		t.Errorf("expected mt-4 class on wrapper, got: %s", html)
	}
}

func TestSwitch_CombinedProps(t *testing.T) {
	sw := Switch(SwitchProps{
		ID:       "notifications-switch",
		Name:     "notifications",
		Checked:  true,
		Label:    "Enable notifications",
		Class:    "mb-4",
		Disabled: true,
	})
	html := renderHTML(sw)

	// Check ID
	if !strings.Contains(html, `id="notifications-switch"`) {
		t.Errorf("expected id attribute, got: %s", html)
	}

	// Check name
	if !strings.Contains(html, `name="notifications"`) {
		t.Errorf("expected name attribute, got: %s", html)
	}

	// Check checked state (on styling)
	if !strings.Contains(html, `checked="checked"`) {
		t.Errorf("expected checked attribute, got: %s", html)
	}
	if !strings.Contains(html, "translate-x-6") {
		t.Errorf("expected translate-x-6 for on state, got: %s", html)
	}
	if !strings.Contains(html, `aria-checked="true"`) {
		t.Errorf("expected aria-checked=\"true\", got: %s", html)
	}

	// Check label
	if !strings.Contains(html, "Enable notifications") {
		t.Errorf("expected label text, got: %s", html)
	}

	// Check custom class
	if !strings.Contains(html, "mb-4") {
		t.Errorf("expected mb-4 class, got: %s", html)
	}

	// Check disabled
	if !strings.Contains(html, `disabled="disabled"`) {
		t.Errorf("expected disabled attribute, got: %s", html)
	}
	if !strings.Contains(html, "opacity-50") {
		t.Errorf("expected opacity-50 for disabled, got: %s", html)
	}
}

func TestSwitch_TrackStyling(t *testing.T) {
	sw := Switch(SwitchProps{
		Name: "toggle",
	})
	html := renderHTML(sw)

	// Should have track styling
	if !strings.Contains(html, "h-6") {
		t.Errorf("expected track height h-6, got: %s", html)
	}
	if !strings.Contains(html, "w-11") {
		t.Errorf("expected track width w-11, got: %s", html)
	}
	if !strings.Contains(html, "rounded-full") {
		t.Errorf("expected rounded-full for track, got: %s", html)
	}
}

func TestSwitch_KnobStyling(t *testing.T) {
	sw := Switch(SwitchProps{
		Name: "toggle",
	})
	html := renderHTML(sw)

	// Should have knob styling
	if !strings.Contains(html, "h-4") {
		t.Errorf("expected knob height h-4, got: %s", html)
	}
	if !strings.Contains(html, "w-4") {
		t.Errorf("expected knob width w-4, got: %s", html)
	}
	if !strings.Contains(html, "bg-white") {
		t.Errorf("expected bg-white for knob, got: %s", html)
	}
}
