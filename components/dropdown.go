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
	container    js.Value
	menu         js.Value
	isOpen       bool
	cleanup      js.Func
	highlightIdx int
	menuItems    []js.Value
	keyHandler   js.Func
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

	// Accessibility attributes
	menu.Call("setAttribute", "tabindex", "0")
	menu.Call("setAttribute", "role", "menu")
	menu.Call("setAttribute", "aria-orientation", "vertical")

	// Add items
	itemIdx := 0
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
		menuItem.Set("data-index", itemIdx)
		menuItem.Call("setAttribute", "role", "menuitem")
		menuItem.Set("id", js.Global().Get("crypto").Call("randomUUID").String())

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

		// Add mouseenter handler to sync highlight on hover
		if !item.Disabled {
			idx := itemIdx
			menuItem.Call("addEventListener", "mouseenter", js.FuncOf(func(this js.Value, args []js.Value) any {
				d.highlightIdx = idx
				d.updateHighlightStyles()
				return nil
			}))
		}

		menu.Call("appendChild", menuItem)
		d.menuItems = append(d.menuItems, menuItem)
		itemIdx++
	}

	container.Call("appendChild", menu)
	d.menu = menu

	// Toggle on trigger click
	triggerWrap.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		args[0].Call("stopPropagation")
		d.Toggle()
		return nil
	}))

	// Close on blur (when focus leaves dropdown)
	menu.Call("addEventListener", "focusout", js.FuncOf(func(this js.Value, args []js.Value) any {
		if !d.isOpen {
			return nil
		}
		event := args[0]
		relatedTarget := event.Get("relatedTarget")
		// If focus is moving outside the dropdown container, close it
		if relatedTarget.IsNull() || !container.Call("contains", relatedTarget).Bool() {
			d.Close()
		}
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
	if d.isOpen {
		return
	}
	d.menu.Get("classList").Call("remove", "hidden")
	d.isOpen = true
	d.highlightIdx = 0
	d.updateHighlightStyles()

	// Focus menu for keyboard navigation
	d.menu.Call("focus")

	// Register keydown handler
	d.keyHandler = js.FuncOf(func(this js.Value, args []js.Value) any {
		event := args[0]
		key := event.Get("key").String()

		switch key {
		case "ArrowDown":
			event.Call("preventDefault")
			d.highlightNext()
		case "ArrowUp":
			event.Call("preventDefault")
			d.highlightPrev()
		case "Enter":
			event.Call("preventDefault")
			d.executeHighlighted()
		case "Escape":
			event.Call("preventDefault")
			d.Close()
		}
		return nil
	})
	js.Global().Get("document").Call("addEventListener", "keydown", d.keyHandler)
}

// Close closes the dropdown menu
func (d *Dropdown) Close() {
	if !d.isOpen {
		return
	}
	d.menu.Get("classList").Call("add", "hidden")
	d.isOpen = false

	// Remove keydown handler
	js.Global().Get("document").Call("removeEventListener", "keydown", d.keyHandler)
	d.keyHandler.Release()
}

// updateHighlightStyles updates highlight visually without re-rendering DOM
func (d *Dropdown) updateHighlightStyles() {
	baseClass := "w-full text-left px-4 py-2 text-sm flex items-center gap-2 text-gray-700 dark:text-gray-200 cursor-pointer"
	highlightClass := baseClass + " bg-gray-100 dark:bg-gray-700"
	normalClass := baseClass + " hover:bg-gray-100 dark:hover:bg-gray-700"

	for i, item := range d.menuItems {
		if item.Get("disabled").Bool() {
			continue
		}
		if i == d.highlightIdx {
			item.Set("className", highlightClass)
			// Update aria-activedescendant for screen readers
			itemId := item.Get("id").String()
			d.menu.Call("setAttribute", "aria-activedescendant", itemId)
		} else {
			item.Set("className", normalClass)
		}
	}
}

// highlightNext moves highlight to next item
func (d *Dropdown) highlightNext() {
	if len(d.menuItems) == 0 {
		return
	}
	d.highlightIdx++
	if d.highlightIdx >= len(d.menuItems) {
		d.highlightIdx = 0
	}
	// Skip disabled items
	for d.menuItems[d.highlightIdx].Get("disabled").Bool() {
		d.highlightIdx++
		if d.highlightIdx >= len(d.menuItems) {
			d.highlightIdx = 0
		}
	}
	d.updateHighlightStyles()
}

// highlightPrev moves highlight to previous item
func (d *Dropdown) highlightPrev() {
	if len(d.menuItems) == 0 {
		return
	}
	d.highlightIdx--
	if d.highlightIdx < 0 {
		d.highlightIdx = len(d.menuItems) - 1
	}
	// Skip disabled items
	for d.menuItems[d.highlightIdx].Get("disabled").Bool() {
		d.highlightIdx--
		if d.highlightIdx < 0 {
			d.highlightIdx = len(d.menuItems) - 1
		}
	}
	d.updateHighlightStyles()
}

// executeHighlighted executes the currently highlighted item
func (d *Dropdown) executeHighlighted() {
	if d.highlightIdx >= 0 && d.highlightIdx < len(d.menuItems) {
		item := d.menuItems[d.highlightIdx]
		if !item.Get("disabled").Bool() {
			item.Call("click")
		}
	}
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
	// Close first to clean up keyHandler
	d.Close()
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
