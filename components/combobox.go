//go:build js && wasm

package components

import (
	"strconv"
	"strings"
	"syscall/js"
)

// ComboboxOption represents an option in a combobox
type ComboboxOption struct {
	Label       string
	Value       string
	Description string
	Disabled    bool
}

// ComboboxProps configures a Combobox component
type ComboboxProps struct {
	Label        string
	Placeholder  string
	Options      []ComboboxOption
	Value        string
	Disabled     bool
	Required     bool
	AllowCustom  bool // Allow typing custom values not in options
	OnChange     func(value string)
	OnSearch     func(query string) // For async search
	EmptyMessage string             // Message when no results
}

// Combobox creates an autocomplete/combobox component
type Combobox struct {
	container     js.Value
	input         js.Value
	dropdown      js.Value
	options       []ComboboxOption
	filteredOpts  []ComboboxOption
	value         string
	isOpen        bool
	highlightIdx  int
	props         ComboboxProps
	cleanup       js.Func
	listboxID     string   // unique ID for listbox
	baseOptionID  string   // base ID for generating option IDs
}

// NewCombobox creates a new Combobox component
func NewCombobox(props ComboboxProps) *Combobox {
	document := js.Global().Get("document")

	if props.EmptyMessage == "" {
		props.EmptyMessage = "No results found"
	}

	// Generate unique IDs for ARIA relationships
	crypto := js.Global().Get("crypto")
	uuid := crypto.Call("randomUUID").String()
	listboxID := "combobox-listbox-" + uuid
	baseOptionID := "combobox-option-" + uuid

	c := &Combobox{
		options:      props.Options,
		filteredOpts: props.Options,
		value:        props.Value,
		highlightIdx: -1,
		props:        props,
		listboxID:    listboxID,
		baseOptionID: baseOptionID,
	}

	container := document.Call("createElement", "div")
	container.Set("className", "relative")

	// Label
	if props.Label != "" {
		label := document.Call("createElement", "label")
		label.Set("className", "block text-sm font-medium text-gray-700 mb-1")
		label.Set("textContent", props.Label)
		container.Call("appendChild", label)
	}

	// Input wrapper
	inputWrap := document.Call("createElement", "div")
	inputWrap.Set("className", "relative")

	// Input
	input := document.Call("createElement", "input")
	input.Set("type", "text")
	input.Set("className", "w-full px-3 py-2 pr-10 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500")
	input.Set("placeholder", props.Placeholder)
	input.Set("autocomplete", "off")

	// ARIA combobox attributes
	input.Call("setAttribute", "role", "combobox")
	input.Call("setAttribute", "aria-expanded", "false")
	input.Call("setAttribute", "aria-haspopup", "listbox")
	input.Call("setAttribute", "aria-controls", listboxID)
	input.Call("setAttribute", "aria-autocomplete", "list")

	if props.Disabled {
		input.Set("disabled", true)
		input.Set("className", "w-full px-3 py-2 pr-10 border border-gray-300 rounded-md shadow-sm bg-gray-100 cursor-not-allowed")
	}
	if props.Required {
		input.Set("required", true)
	}

	// Set initial value
	if props.Value != "" {
		for _, opt := range props.Options {
			if opt.Value == props.Value {
				input.Set("value", opt.Label)
				break
			}
		}
	}

	c.input = input
	inputWrap.Call("appendChild", input)

	// Dropdown arrow
	arrow := document.Call("createElement", "div")
	arrow.Set("className", "absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 pointer-events-none")
	arrow.Set("textContent", "▼")
	inputWrap.Call("appendChild", arrow)

	container.Call("appendChild", inputWrap)

	// Dropdown (listbox)
	dropdown := document.Call("createElement", "div")
	dropdown.Set("className", "absolute w-full mt-1 bg-white border border-gray-200 rounded-md shadow-lg max-h-60 overflow-auto z-50 hidden")

	// ARIA listbox attributes
	dropdown.Call("setAttribute", "role", "listbox")
	dropdown.Set("id", listboxID)
	dropdown.Call("setAttribute", "aria-label", "Options")

	c.dropdown = dropdown
	container.Call("appendChild", dropdown)

	c.container = container
	c.renderOptions()

	// Input events
	input.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		query := input.Get("value").String()
		c.filter(query)
		c.Open()
		if props.OnSearch != nil {
			props.OnSearch(query)
		}
		return nil
	}))

	input.Call("addEventListener", "focus", js.FuncOf(func(this js.Value, args []js.Value) any {
		c.Open()
		return nil
	}))

	input.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) any {
		key := args[0].Get("key").String()
		switch key {
		case "ArrowDown":
			args[0].Call("preventDefault")
			c.highlightNext()
		case "ArrowUp":
			args[0].Call("preventDefault")
			c.highlightPrev()
		case "Enter":
			args[0].Call("preventDefault")
			if c.highlightIdx >= 0 && c.highlightIdx < len(c.filteredOpts) {
				c.selectOption(c.filteredOpts[c.highlightIdx])
			} else if props.AllowCustom {
				c.value = input.Get("value").String()
				if props.OnChange != nil {
					props.OnChange(c.value)
				}
			}
			c.Close()
		case "Escape":
			c.Close()
		}
		return nil
	}))

	// Close on outside click
	c.cleanup = js.FuncOf(func(this js.Value, args []js.Value) any {
		if c.isOpen {
			target := args[0].Get("target")
			if !container.Call("contains", target).Bool() {
				c.Close()
			}
		}
		return nil
	})
	document.Call("addEventListener", "click", c.cleanup)

	return c
}

func (c *Combobox) renderOptions() {
	document := js.Global().Get("document")
	c.dropdown.Set("innerHTML", "")

	if len(c.filteredOpts) == 0 {
		empty := document.Call("createElement", "div")
		empty.Set("className", "px-3 py-2 text-sm text-gray-500")
		empty.Set("textContent", c.props.EmptyMessage)
		c.dropdown.Call("appendChild", empty)
		// Clear aria-activedescendant when no options
		c.input.Call("removeAttribute", "aria-activedescendant")
		return
	}

	for i, opt := range c.filteredOpts {
		item := document.Call("createElement", "div")
		itemClass := "px-3 py-2 cursor-pointer"
		if opt.Disabled {
			itemClass = "px-3 py-2 text-gray-400 cursor-not-allowed"
		} else if i == c.highlightIdx {
			itemClass = "px-3 py-2 cursor-pointer bg-blue-50"
		} else {
			itemClass = "px-3 py-2 cursor-pointer hover:bg-gray-100"
		}
		item.Set("className", itemClass)

		// ARIA option attributes
		optionID := c.baseOptionID + "-" + strconv.Itoa(i)
		item.Call("setAttribute", "role", "option")
		item.Set("id", optionID)
		if i == c.highlightIdx {
			item.Call("setAttribute", "aria-selected", "true")
			// Update aria-activedescendant on input
			c.input.Call("setAttribute", "aria-activedescendant", optionID)
		} else {
			item.Call("setAttribute", "aria-selected", "false")
		}

		label := document.Call("createElement", "div")
		label.Set("className", "text-sm font-medium")
		label.Set("textContent", opt.Label)
		item.Call("appendChild", label)

		if opt.Description != "" {
			desc := document.Call("createElement", "div")
			desc.Set("className", "text-xs text-gray-500")
			desc.Set("textContent", opt.Description)
			item.Call("appendChild", desc)
		}

		if !opt.Disabled {
			option := opt
			item.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
				c.selectOption(option)
				c.Close()
				return nil
			}))
		}

		c.dropdown.Call("appendChild", item)
	}

	// Clear aria-activedescendant if nothing is highlighted
	if c.highlightIdx < 0 {
		c.input.Call("removeAttribute", "aria-activedescendant")
	}
}

func (c *Combobox) filter(query string) {
	query = strings.ToLower(query)
	c.filteredOpts = nil
	c.highlightIdx = -1

	for _, opt := range c.options {
		if strings.Contains(strings.ToLower(opt.Label), query) ||
			strings.Contains(strings.ToLower(opt.Value), query) ||
			strings.Contains(strings.ToLower(opt.Description), query) {
			c.filteredOpts = append(c.filteredOpts, opt)
		}
	}

	c.renderOptions()
}

func (c *Combobox) selectOption(opt ComboboxOption) {
	c.value = opt.Value
	c.input.Set("value", opt.Label)
	if c.props.OnChange != nil {
		c.props.OnChange(opt.Value)
	}
}

func (c *Combobox) highlightNext() {
	if len(c.filteredOpts) == 0 {
		return
	}
	c.highlightIdx++
	if c.highlightIdx >= len(c.filteredOpts) {
		c.highlightIdx = 0
	}
	c.renderOptions()
	c.scrollToHighlighted()
}

func (c *Combobox) highlightPrev() {
	if len(c.filteredOpts) == 0 {
		return
	}
	c.highlightIdx--
	if c.highlightIdx < 0 {
		c.highlightIdx = len(c.filteredOpts) - 1
	}
	c.renderOptions()
	c.scrollToHighlighted()
}

func (c *Combobox) scrollToHighlighted() {
	if c.highlightIdx >= 0 {
		items := c.dropdown.Get("children")
		if c.highlightIdx < items.Length() {
			item := items.Index(c.highlightIdx)
			item.Call("scrollIntoView", map[string]any{"block": "nearest"})
		}
	}
}

// Element returns the container DOM element
func (c *Combobox) Element() js.Value {
	return c.container
}

// Open opens the dropdown
func (c *Combobox) Open() {
	c.dropdown.Get("classList").Call("remove", "hidden")
	c.isOpen = true
	c.input.Call("setAttribute", "aria-expanded", "true")
}

// Close closes the dropdown
func (c *Combobox) Close() {
	c.dropdown.Get("classList").Call("add", "hidden")
	c.isOpen = false
	c.highlightIdx = -1
	c.input.Call("setAttribute", "aria-expanded", "false")
	c.input.Call("removeAttribute", "aria-activedescendant")
}

// Value returns the current value
func (c *Combobox) Value() string {
	return c.value
}

// SetValue sets the current value
func (c *Combobox) SetValue(value string) {
	c.value = value
	for _, opt := range c.options {
		if opt.Value == value {
			c.input.Set("value", opt.Label)
			return
		}
	}
	if c.props.AllowCustom {
		c.input.Set("value", value)
	}
}

// SetOptions updates the available options
func (c *Combobox) SetOptions(options []ComboboxOption) {
	c.options = options
	c.filteredOpts = options
	c.renderOptions()
}

// Destroy cleans up event listeners
func (c *Combobox) Destroy() {
	js.Global().Get("document").Call("removeEventListener", "click", c.cleanup)
	c.cleanup.Release()
}

// SimpleCombobox creates a combobox with string options
func SimpleCombobox(label, placeholder string, options ...string) *Combobox {
	opts := make([]ComboboxOption, len(options))
	for i, opt := range options {
		opts[i] = ComboboxOption{Label: opt, Value: opt}
	}
	return NewCombobox(ComboboxProps{
		Label:       label,
		Placeholder: placeholder,
		Options:     opts,
	})
}

// SearchableSelect creates a searchable select dropdown
func SearchableSelect(label, placeholder string, options []ComboboxOption, onChange func(string)) *Combobox {
	return NewCombobox(ComboboxProps{
		Label:       label,
		Placeholder: placeholder,
		Options:     options,
		OnChange:    onChange,
	})
}
