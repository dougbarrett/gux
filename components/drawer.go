//go:build js && wasm

package components

import "syscall/js"

// DrawerPosition defines which side the drawer opens from
type DrawerPosition string

const (
	DrawerLeft   DrawerPosition = "left"
	DrawerRight  DrawerPosition = "right"
	DrawerTop    DrawerPosition = "top"
	DrawerBottom DrawerPosition = "bottom"
)

// DrawerProps configures a Drawer component
type DrawerProps struct {
	Title      string
	Content    js.Value
	Position   DrawerPosition // Default: right
	Width      string         // For left/right drawers (default "320px")
	Height     string         // For top/bottom drawers (default "auto")
	ShowClose  bool           // Show close button (default true)
	Overlay    bool           // Show overlay behind drawer (default true)
	CloseOnEsc bool           // Close on Escape key (default true)
	OnClose    func()
}

// Drawer creates a slide-out panel component
type Drawer struct {
	overlay   js.Value
	drawer    js.Value
	isOpen    bool
	props     DrawerProps
	escHandler js.Func
}

// NewDrawer creates a new Drawer component
func NewDrawer(props DrawerProps) *Drawer {
	document := js.Global().Get("document")

	// Defaults
	if props.Position == "" {
		props.Position = DrawerRight
	}
	if props.Width == "" {
		props.Width = "320px"
	}
	if props.Height == "" {
		props.Height = "auto"
	}

	d := &Drawer{props: props}

	// Overlay
	overlay := document.Call("createElement", "div")
	overlay.Set("className", "fixed inset-0 bg-black bg-opacity-50 z-40 hidden transition-opacity duration-300 opacity-0")
	d.overlay = overlay

	// Drawer panel
	drawer := document.Call("createElement", "div")
	baseClass := "fixed bg-white dark:bg-gray-800 shadow-xl z-50 transition-transform duration-300 ease-in-out flex flex-col"

	var transformHidden, transformVisible string
	var positionClass string

	switch props.Position {
	case DrawerLeft:
		positionClass = "top-0 left-0 h-full"
		transformHidden = "translateX(-100%)"
		transformVisible = "translateX(0)"
		drawer.Get("style").Set("width", props.Width)
	case DrawerRight:
		positionClass = "top-0 right-0 h-full"
		transformHidden = "translateX(100%)"
		transformVisible = "translateX(0)"
		drawer.Get("style").Set("width", props.Width)
	case DrawerTop:
		positionClass = "top-0 left-0 w-full"
		transformHidden = "translateY(-100%)"
		transformVisible = "translateY(0)"
		if props.Height != "auto" {
			drawer.Get("style").Set("height", props.Height)
		}
	case DrawerBottom:
		positionClass = "bottom-0 left-0 w-full"
		transformHidden = "translateY(100%)"
		transformVisible = "translateY(0)"
		if props.Height != "auto" {
			drawer.Get("style").Set("height", props.Height)
		}
	}

	drawer.Set("className", baseClass+" "+positionClass)
	drawer.Get("style").Set("transform", transformHidden)
	d.drawer = drawer

	// Store transforms for animation
	drawer.Set("data-transform-hidden", transformHidden)
	drawer.Set("data-transform-visible", transformVisible)

	// Header
	if props.Title != "" || props.ShowClose {
		header := document.Call("createElement", "div")
		header.Set("className", "flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700")

		if props.Title != "" {
			title := document.Call("createElement", "h2")
			title.Set("className", "text-lg font-semibold text-gray-900 dark:text-gray-100")
			title.Set("textContent", props.Title)
			header.Call("appendChild", title)
		}

		if props.ShowClose {
			closeBtn := document.Call("createElement", "button")
			closeBtn.Set("className", "p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded text-gray-500 dark:text-gray-400 text-xl")
			closeBtn.Set("textContent", "×")
			closeBtn.Call("setAttribute", "aria-label", "Close drawer")
			closeBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
				d.Close()
				return nil
			}))
			header.Call("appendChild", closeBtn)
		}

		drawer.Call("appendChild", header)
	}

	// Content
	content := document.Call("createElement", "div")
	content.Set("className", "flex-1 overflow-auto p-4")
	if !props.Content.IsUndefined() && !props.Content.IsNull() {
		content.Call("appendChild", props.Content)
	}
	drawer.Call("appendChild", content)

	// Overlay click to close
	if props.Overlay {
		overlay.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			d.Close()
			return nil
		}))
	}

	// Escape key handler
	if props.CloseOnEsc {
		d.escHandler = js.FuncOf(func(this js.Value, args []js.Value) any {
			if d.isOpen && args[0].Get("key").String() == "Escape" {
				d.Close()
			}
			return nil
		})
		document.Call("addEventListener", "keydown", d.escHandler)
	}

	// Append to body
	body := document.Get("body")
	body.Call("appendChild", overlay)
	body.Call("appendChild", drawer)

	return d
}

// Open opens the drawer
func (d *Drawer) Open() {
	if d.isOpen {
		return
	}
	d.isOpen = true

	// Show overlay
	d.overlay.Get("classList").Call("remove", "hidden")
	// Trigger reflow for animation
	_ = d.overlay.Get("offsetHeight")
	d.overlay.Get("classList").Call("remove", "opacity-0")
	d.overlay.Get("classList").Call("add", "opacity-100")

	// Slide in drawer
	transformVisible := d.drawer.Get("data-transform-visible").String()
	d.drawer.Get("style").Set("transform", transformVisible)

	// Prevent body scroll
	js.Global().Get("document").Get("body").Get("style").Set("overflow", "hidden")
}

// Close closes the drawer
func (d *Drawer) Close() {
	if !d.isOpen {
		return
	}
	d.isOpen = false

	// Hide overlay
	d.overlay.Get("classList").Call("remove", "opacity-100")
	d.overlay.Get("classList").Call("add", "opacity-0")

	// Slide out drawer
	transformHidden := d.drawer.Get("data-transform-hidden").String()
	d.drawer.Get("style").Set("transform", transformHidden)

	// Re-enable body scroll
	js.Global().Get("document").Get("body").Get("style").Set("overflow", "")

	// Hide overlay after animation
	js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
		if !d.isOpen {
			d.overlay.Get("classList").Call("add", "hidden")
		}
		return nil
	}), 300)

	if d.props.OnClose != nil {
		d.props.OnClose()
	}
}

// Toggle toggles the drawer
func (d *Drawer) Toggle() {
	if d.isOpen {
		d.Close()
	} else {
		d.Open()
	}
}

// IsOpen returns whether the drawer is open
func (d *Drawer) IsOpen() bool {
	return d.isOpen
}

// SetContent updates the drawer content
func (d *Drawer) SetContent(content js.Value) {
	contentArea := d.drawer.Call("querySelector", ".overflow-auto")
	contentArea.Set("innerHTML", "")
	contentArea.Call("appendChild", content)
}

// Element returns the drawer element (for reference, not mounting)
func (d *Drawer) Element() js.Value {
	return d.drawer
}

// Destroy removes the drawer from DOM and cleans up
func (d *Drawer) Destroy() {
	if d.props.CloseOnEsc {
		js.Global().Get("document").Call("removeEventListener", "keydown", d.escHandler)
		d.escHandler.Release()
	}
	d.overlay.Call("remove")
	d.drawer.Call("remove")
}

// RightDrawer creates a drawer that slides from the right
func RightDrawer(title string, content js.Value) *Drawer {
	return NewDrawer(DrawerProps{
		Title:      title,
		Content:    content,
		Position:   DrawerRight,
		ShowClose:  true,
		Overlay:    true,
		CloseOnEsc: true,
	})
}

// LeftDrawer creates a drawer that slides from the left (navigation style)
func LeftDrawer(title string, content js.Value) *Drawer {
	return NewDrawer(DrawerProps{
		Title:      title,
		Content:    content,
		Position:   DrawerLeft,
		ShowClose:  true,
		Overlay:    true,
		CloseOnEsc: true,
	})
}

// BottomSheet creates a drawer that slides from the bottom (mobile style)
func BottomSheet(content js.Value) *Drawer {
	return NewDrawer(DrawerProps{
		Content:    content,
		Position:   DrawerBottom,
		Height:     "50vh",
		ShowClose:  false,
		Overlay:    true,
		CloseOnEsc: true,
	})
}
