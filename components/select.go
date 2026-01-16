//go:build js && wasm

package components

import "syscall/js"

// SelectOption represents an option in a select dropdown
type SelectOption struct {
	Label    string
	Value    string
	Disabled bool
}

// SelectProps configures a Select component
type SelectProps struct {
	Label       string
	Options     []SelectOption
	Value       string
	Placeholder string
	ClassName   string
	Disabled    bool
	Required    bool
	OnChange    func(value string)
}

// Select creates a dropdown select component
type Select struct {
	container js.Value
	selectEl  js.Value
	label     js.Value
	selectID  string
}

// NewSelect creates a new Select component
func NewSelect(props SelectProps) *Select {
	document := js.Global().Get("document")
	crypto := js.Global().Get("crypto")

	container := document.Call("createElement", "div")
	container.Set("className", "mb-4")

	// Generate unique ID for label-input association
	selectID := "select-" + crypto.Call("randomUUID").String()

	s := &Select{container: container, selectID: selectID}

	// Label
	if props.Label != "" {
		label := document.Call("createElement", "label")
		label.Set("className", "block text-sm font-medium text-secondary mb-1")
		label.Set("textContent", props.Label)
		label.Set("htmlFor", selectID)
		container.Call("appendChild", label)
		s.label = label
	}

	// Select
	selectEl := document.Call("createElement", "select")
	selectEl.Set("id", selectID)
	className := "w-full px-3 py-2 border border-default rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 surface-base text-primary"
	if props.Disabled {
		className += " surface-overlay cursor-not-allowed"
	}
	if props.ClassName != "" {
		className = props.ClassName
	}

	selectEl.Set("className", className)

	if props.Disabled {
		selectEl.Set("disabled", true)
	}
	if props.Required {
		selectEl.Set("required", true)
	}

	// Placeholder option
	if props.Placeholder != "" {
		placeholder := document.Call("createElement", "option")
		placeholder.Set("value", "")
		placeholder.Set("textContent", props.Placeholder)
		placeholder.Set("disabled", true)
		if props.Value == "" {
			placeholder.Set("selected", true)
		}
		selectEl.Call("appendChild", placeholder)
	}

	// Options
	for _, opt := range props.Options {
		option := document.Call("createElement", "option")
		option.Set("value", opt.Value)
		option.Set("textContent", opt.Label)
		if opt.Disabled {
			option.Set("disabled", true)
		}
		if opt.Value == props.Value {
			option.Set("selected", true)
		}
		selectEl.Call("appendChild", option)
	}

	if props.OnChange != nil {
		selectEl.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) any {
			value := selectEl.Get("value").String()
			props.OnChange(value)
			return nil
		}))
	}

	container.Call("appendChild", selectEl)
	s.selectEl = selectEl

	return s
}

// Element returns the container DOM element
func (s *Select) Element() js.Value {
	return s.container
}

// Value returns the current selected value
func (s *Select) Value() string {
	return s.selectEl.Get("value").String()
}

// SetValue sets the selected value
func (s *Select) SetValue(value string) {
	s.selectEl.Set("value", value)
}

// Quick select constructors

// SimpleSelect creates a select with label and string options (value = label)
func SimpleSelect(label string, options ...string) *Select {
	opts := make([]SelectOption, len(options))
	for i, opt := range options {
		opts[i] = SelectOption{Label: opt, Value: opt}
	}
	return NewSelect(SelectProps{Label: label, Options: opts})
}

// SimpleSelectWithPlaceholder creates a select with placeholder and string options
func SimpleSelectWithPlaceholder(label, placeholder string, options ...string) *Select {
	opts := make([]SelectOption, len(options))
	for i, opt := range options {
		opts[i] = SelectOption{Label: opt, Value: opt}
	}
	return NewSelect(SelectProps{Label: label, Placeholder: placeholder, Options: opts})
}
