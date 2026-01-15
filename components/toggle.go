//go:build js && wasm

package components

import "syscall/js"

// ToggleSize defines toggle switch sizes
type ToggleSize string

const (
	ToggleSM ToggleSize = "sm"
	ToggleMD ToggleSize = "md"
	ToggleLG ToggleSize = "lg"
)

// ToggleProps configures a Toggle component
type ToggleProps struct {
	Label       string
	Description string
	Checked     bool
	Disabled    bool
	Size        ToggleSize
	OnChange    func(checked bool)
}

// Toggle creates a toggle switch component
type Toggle struct {
	container js.Value
	toggle    js.Value
	knob      js.Value
	checked   bool
	disabled  bool
	onChange  func(bool)
}

// NewToggle creates a new Toggle component
func NewToggle(props ToggleProps) *Toggle {
	document := js.Global().Get("document")

	size := props.Size
	if size == "" {
		size = ToggleMD
	}

	t := &Toggle{
		checked:  props.Checked,
		disabled: props.Disabled,
		onChange: props.OnChange,
	}

	container := document.Call("createElement", "label")
	container.Set("className", "flex items-center gap-3 cursor-pointer")
	if props.Disabled {
		container.Set("className", "flex items-center gap-3 cursor-not-allowed opacity-50")
	}

	// Toggle track
	var trackWidth, trackHeight, knobSize, knobTranslate string
	switch size {
	case ToggleSM:
		trackWidth = "36px"
		trackHeight = "20px"
		knobSize = "16px"
		knobTranslate = "16px"
	case ToggleMD:
		trackWidth = "44px"
		trackHeight = "24px"
		knobSize = "20px"
		knobTranslate = "20px"
	case ToggleLG:
		trackWidth = "52px"
		trackHeight = "28px"
		knobSize = "24px"
		knobTranslate = "24px"
	}

	toggle := document.Call("createElement", "div")
	toggle.Set("className", "relative inline-block rounded-full transition-colors duration-200 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2")
	toggle.Get("style").Set("width", trackWidth)
	toggle.Get("style").Set("height", trackHeight)
	toggle.Set("role", "switch")
	toggle.Set("aria-checked", props.Checked)
	toggle.Set("tabindex", "0")

	t.updateTrackColor(toggle)

	// Knob
	knob := document.Call("createElement", "div")
	knob.Set("className", "absolute bg-white rounded-full shadow transition-transform duration-200")
	knob.Get("style").Set("width", knobSize)
	knob.Get("style").Set("height", knobSize)
	knob.Get("style").Set("top", "2px")
	knob.Get("style").Set("left", "2px")
	knob.Set("data-translate", knobTranslate)

	if props.Checked {
		knob.Get("style").Set("transform", "translateX("+knobTranslate+")")
	}

	toggle.Call("appendChild", knob)
	t.toggle = toggle
	t.knob = knob

	container.Call("appendChild", toggle)

	// Label text
	if props.Label != "" || props.Description != "" {
		textContainer := document.Call("createElement", "div")

		if props.Label != "" {
			label := document.Call("createElement", "div")
			label.Set("className", "text-sm font-medium text-gray-900 dark:text-gray-100")
			label.Set("textContent", props.Label)
			textContainer.Call("appendChild", label)
		}

		if props.Description != "" {
			desc := document.Call("createElement", "div")
			desc.Set("className", "text-xs text-gray-500 dark:text-gray-400")
			desc.Set("textContent", props.Description)
			textContainer.Call("appendChild", desc)
		}

		container.Call("appendChild", textContainer)
	}

	t.container = container

	// Click handler
	if !props.Disabled {
		container.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			t.Toggle()
			return nil
		}))

		// Keyboard handler
		toggle.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) any {
			key := args[0].Get("key").String()
			if key == " " || key == "Enter" {
				args[0].Call("preventDefault")
				t.Toggle()
			}
			return nil
		}))
	}

	return t
}

func (t *Toggle) updateTrackColor(toggle js.Value) {
	if t.checked {
		toggle.Get("style").Set("backgroundColor", "#3b82f6") // blue-500
	} else {
		// Check if dark mode is active
		isDark := js.Global().Get("document").Get("documentElement").Get("classList").Call("contains", "dark").Bool()
		if isDark {
			toggle.Get("style").Set("backgroundColor", "#4b5563") // gray-600
		} else {
			toggle.Get("style").Set("backgroundColor", "#d1d5db") // gray-300
		}
	}
}

// Element returns the container DOM element
func (t *Toggle) Element() js.Value {
	return t.container
}

// Checked returns whether the toggle is checked
func (t *Toggle) Checked() bool {
	return t.checked
}

// SetChecked sets the toggle state
func (t *Toggle) SetChecked(checked bool) {
	if t.checked == checked {
		return
	}
	t.checked = checked
	t.toggle.Set("aria-checked", checked)
	t.updateTrackColor(t.toggle)

	translate := t.knob.Get("data-translate").String()
	if checked {
		t.knob.Get("style").Set("transform", "translateX("+translate+")")
	} else {
		t.knob.Get("style").Set("transform", "translateX(0)")
	}
}

// Toggle toggles the switch state
func (t *Toggle) Toggle() {
	if t.disabled {
		return
	}
	t.SetChecked(!t.checked)
	if t.onChange != nil {
		t.onChange(t.checked)
	}
}

// SimpleToggle creates a basic toggle with label
func SimpleToggle(label string, checked bool, onChange func(bool)) *Toggle {
	return NewToggle(ToggleProps{
		Label:    label,
		Checked:  checked,
		OnChange: onChange,
	})
}

// ToggleWithDescription creates a toggle with label and description
func ToggleWithDescription(label, description string, checked bool, onChange func(bool)) *Toggle {
	return NewToggle(ToggleProps{
		Label:       label,
		Description: description,
		Checked:     checked,
		OnChange:    onChange,
	})
}
