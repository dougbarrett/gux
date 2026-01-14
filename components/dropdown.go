//go:build js && wasm

package components

import "syscall/js"

// DropdownItem represents an item in a dropdown menu
type DropdownItem struct {
	Label    string
	Icon     string
	OnClick  func()
	Disabled bool
	Divider  bool // If true, renders a divider instead of an item
}

// DropdownProps configures a Dropdown component
type DropdownProps struct {
	Trigger   js.Value // The element that triggers the dropdown
	Items     []DropdownItem
	Align     string // "left" or "right" (default "left")
	Width     string // CSS width (default "auto")
	ClassName string
}

// Dropdown creates a dropdown menu component
type Dropdown struct {
	container js.Value
	menu      js.Value
	isOpen    bool
	cleanup   js.Func
}

// NewDropdown creates a new Dropdown component
func NewDropdown(props DropdownProps) *Dropdown {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "relative inline-block")

	d := &Dropdown{container: container}

	// Wrap trigger
	triggerWrap := document.Call("createElement", "div")
	triggerWrap.Set("className", "cursor-pointer")
	triggerWrap.Call("appendChild", props.Trigger)
	container.Call("appendChild", triggerWrap)

	// Create menu
	menu := document.Call("createElement", "div")
	align := props.Align
	if align == "" {
		align = "left"
	}
	width := props.Width
	if width == "" {
		width = "auto"
	}

	alignClass := "left-0"
	if align == "right" {
		alignClass = "right-0"
	}

	className := "absolute " + alignClass + " mt-2 bg-white dark:bg-gray-800 rounded-md shadow-lg border border-gray-200 dark:border-gray-700 py-1 z-50 hidden"
	if props.ClassName != "" {
		className += " " + props.ClassName
	}
	menu.Set("className", className)
	menu.Get("style").Set("minWidth", "150px")
	if width != "auto" {
		menu.Get("style").Set("width", width)
	}

	// Add items
	for _, item := range props.Items {
		if item.Divider {
			divider := document.Call("createElement", "div")
			divider.Set("className", "border-t border-gray-200 dark:border-gray-700 my-1")
			menu.Call("appendChild", divider)
			continue
		}

		menuItem := document.Call("createElement", "button")
		itemClass := "w-full text-left px-4 py-2 text-sm flex items-center gap-2"
		if item.Disabled {
			itemClass += " text-gray-400 dark:text-gray-500 cursor-not-allowed"
		} else {
			itemClass += " text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer"
		}
		menuItem.Set("className", itemClass)
		menuItem.Set("disabled", item.Disabled)

		if item.Icon != "" {
			icon := document.Call("createElement", "span")
			icon.Set("textContent", item.Icon)
			menuItem.Call("appendChild", icon)
		}

		label := document.Call("createElement", "span")
		label.Set("textContent", item.Label)
		menuItem.Call("appendChild", label)

		if !item.Disabled && item.OnClick != nil {
			onClick := item.OnClick
			menuItem.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
				d.Close()
				onClick()
				return nil
			}))
		}

		menu.Call("appendChild", menuItem)
	}

	container.Call("appendChild", menu)
	d.menu = menu

	// Toggle on trigger click
	triggerWrap.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		args[0].Call("stopPropagation")
		d.Toggle()
		return nil
	}))

	// Close on outside click
	d.cleanup = js.FuncOf(func(this js.Value, args []js.Value) any {
		if d.isOpen {
			target := args[0].Get("target")
			if !container.Call("contains", target).Bool() {
				d.Close()
			}
		}
		return nil
	})
	js.Global().Get("document").Call("addEventListener", "click", d.cleanup)

	return d
}

// Element returns the container DOM element
func (d *Dropdown) Element() js.Value {
	return d.container
}

// Open opens the dropdown menu
func (d *Dropdown) Open() {
	d.menu.Get("classList").Call("remove", "hidden")
	d.isOpen = true
}

// Close closes the dropdown menu
func (d *Dropdown) Close() {
	d.menu.Get("classList").Call("add", "hidden")
	d.isOpen = false
}

// Toggle toggles the dropdown menu
func (d *Dropdown) Toggle() {
	if d.isOpen {
		d.Close()
	} else {
		d.Open()
	}
}

// IsOpen returns whether the dropdown is open
func (d *Dropdown) IsOpen() bool {
	return d.isOpen
}

// Destroy cleans up event listeners
func (d *Dropdown) Destroy() {
	js.Global().Get("document").Call("removeEventListener", "click", d.cleanup)
	d.cleanup.Release()
}

// ActionDropdown creates a dropdown with a button trigger
func ActionDropdown(buttonText string, items []DropdownItem) *Dropdown {
	return NewDropdown(DropdownProps{
		Trigger: Button(ButtonProps{
			Text:    buttonText + " ▼",
			Variant: ButtonSecondary,
			Size:    ButtonSM,
		}),
		Items: items,
	})
}

// IconDropdown creates a dropdown with an icon trigger (for action menus)
func IconDropdown(icon string, items []DropdownItem) *Dropdown {
	document := js.Global().Get("document")
	trigger := document.Call("createElement", "button")
	trigger.Set("className", "p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-full text-gray-600 dark:text-gray-300")
	trigger.Set("textContent", icon)

	return NewDropdown(DropdownProps{
		Trigger: trigger,
		Items:   items,
		Align:   "right",
	})
}
