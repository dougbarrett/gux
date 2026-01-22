package ui

import "github.com/dougbarrett/gux/core"

// Switch class constants
const (
	// switchTrackClasses defines the track (background) styling.
	switchTrackClasses = "relative inline-flex h-6 w-11 items-center rounded-full transition-colors duration-200 ease-in-out focus-within:outline-none focus-within:ring-2 focus-within:ring-blue-500 focus-within:ring-offset-2"

	// switchTrackOnClasses are applied when switch is on.
	switchTrackOnClasses = "bg-blue-600"

	// switchTrackOffClasses are applied when switch is off.
	switchTrackOffClasses = "bg-gray-300 dark:bg-gray-600"

	// switchKnobClasses defines the knob (moving circle) styling.
	switchKnobClasses = "inline-block h-4 w-4 rounded-full bg-white transform transition-transform duration-200 ease-in-out shadow"

	// switchKnobOnClasses position the knob when switch is on.
	switchKnobOnClasses = "translate-x-6"

	// switchKnobOffClasses position the knob when switch is off.
	switchKnobOffClasses = "translate-x-1"

	// switchDisabledClasses are applied when switch is disabled.
	switchDisabledClasses = "opacity-50 cursor-not-allowed"
)

// SwitchProps configures a Switch component.
type SwitchProps struct {
	ID       string // Element ID for the hidden input
	Name     string // Input name attribute for form submission
	Checked  bool   // Checked/on state
	Label    string // Label text (optional - displayed next to switch)
	Class    string // Additional wrapper classes
	Disabled bool   // Disabled state
}

// Switch creates a toggle switch with on/off visual states.
//
// The switch uses a hidden checkbox for form submission and a visual
// representation using spans. The aria-checked attribute reflects the
// checked state for accessibility.
//
// Note: Since this is SSR+WASM, the checked state is controlled by props.
// In WASM, clicking should toggle state via page-level handler, not internal state.
//
// Example:
//
//	Switch(SwitchProps{
//	    ID:      "notifications",
//	    Name:    "notifications",
//	    Checked: notifyState.Get(),
//	    Label:   "Enable notifications",
//	})
func Switch(props SwitchProps) core.Node {
	// Build track class based on checked state
	trackClass := MergeClasses(
		switchTrackClasses,
		ConditionalClass(props.Checked, switchTrackOnClasses),
		ConditionalClass(!props.Checked, switchTrackOffClasses),
		ConditionalClass(props.Disabled, switchDisabledClasses),
	)

	// Build knob class based on checked state
	knobClass := MergeClasses(
		switchKnobClasses,
		ConditionalClass(props.Checked, switchKnobOnClasses),
		ConditionalClass(!props.Checked, switchKnobOffClasses),
	)

	// Build hidden checkbox attributes
	hiddenInputAttrs := core.Attrs{
		ID:    props.ID,
		Type:  "checkbox",
		Name:  props.Name,
		Class: "sr-only", // Screen reader only (visually hidden)
	}

	// Build extra attributes for hidden input
	// CRITICAL: Only add "checked" when Checked=true (SSR boolean attrs are presence-based)
	extra := make(map[string]string)
	if props.Checked {
		extra["checked"] = "checked"
	}
	if props.Disabled {
		extra["disabled"] = "disabled"
	}
	if len(extra) > 0 {
		hiddenInputAttrs.Extra = extra
	}

	hiddenInput := core.Input(hiddenInputAttrs)

	// Build visual switch (track + knob)
	visualSwitch := core.Span(
		core.Attrs{
			Class: trackClass,
			Extra: map[string]string{
				"role":         "switch",
				"aria-checked": boolToAttr(props.Checked),
			},
		},
		core.Span(core.Class(knobClass)),
	)

	// Build wrapper class
	wrapperClass := MergeClasses(
		"inline-flex items-center gap-2",
		ConditionalClass(props.Disabled, "cursor-not-allowed"),
		props.Class,
	)

	// Build children
	children := []core.Node{hiddenInput, visualSwitch}

	// Add label if provided
	if props.Label != "" {
		labelSpan := core.Span(
			core.Class("text-sm font-medium text-gray-700 dark:text-gray-300"),
			core.Text(props.Label),
		)
		children = append(children, labelSpan)
	}

	// Wrap in label element
	return core.Label(core.Class(wrapperClass), children...)
}

// boolToAttr converts a bool to "true" or "false" string for aria attributes.
func boolToAttr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
