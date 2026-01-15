//go:build js && wasm

package components

import "syscall/js"

// ModalProps configures a Modal component
type ModalProps struct {
	Title      string
	Content    js.Value
	Footer     js.Value
	Width      string // sm, md, lg, xl, full
	OnClose    func()
	CloseOnEsc bool
}

// Modal creates a modal dialog overlay
type Modal struct {
	overlay js.Value
	modal   js.Value
	content js.Value
	isOpen  bool
	titleID string // ARIA: unique ID for aria-labelledby
}

var modalWidths = map[string]string{
	"sm":   "max-w-sm",
	"md":   "max-w-md",
	"lg":   "max-w-lg",
	"xl":   "max-w-xl",
	"full": "max-w-4xl",
}

// NewModal creates a new Modal component
func NewModal(props ModalProps) *Modal {
	document := js.Global().Get("document")

	// Overlay
	overlay := document.Call("createElement", "div")
	overlay.Set("className", "fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 hidden")

	// Modal container
	width := props.Width
	if width == "" {
		width = "md"
	}
	widthClass := modalWidths[width]
	if widthClass == "" {
		widthClass = modalWidths["md"]
	}

	modal := document.Call("createElement", "div")
	modal.Set("className", "bg-white dark:bg-gray-800 rounded-lg shadow-xl "+widthClass+" w-full mx-4 max-h-[90vh] flex flex-col")

	// Generate unique ID for ARIA labelledby
	titleID := ""
	if props.Title != "" {
		titleID = "modal-title-" + js.Global().Get("crypto").Call("randomUUID").String()
	}

	// Add ARIA dialog attributes
	modal.Call("setAttribute", "role", "dialog")
	modal.Call("setAttribute", "aria-modal", "true")
	if titleID != "" {
		modal.Call("setAttribute", "aria-labelledby", titleID)
	}

	m := &Modal{
		overlay: overlay,
		modal:   modal,
		titleID: titleID,
	}

	// Header
	if props.Title != "" {
		header := document.Call("createElement", "div")
		header.Set("className", "flex justify-between items-center px-6 py-4 border-b border-gray-200 dark:border-gray-700")

		title := document.Call("createElement", "h3")
		title.Set("className", "text-lg font-semibold text-gray-900 dark:text-gray-100")
		title.Set("textContent", props.Title)
		title.Set("id", m.titleID) // ARIA: referenced by aria-labelledby
		header.Call("appendChild", title)

		closeBtn := document.Call("createElement", "button")
		closeBtn.Set("className", "text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-2xl leading-none cursor-pointer")
		closeBtn.Set("innerHTML", "&times;")
		closeBtn.Call("setAttribute", "aria-label", "Close") // ARIA: accessible name for close button
		closeBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			m.Close()
			return nil
		}))
		header.Call("appendChild", closeBtn)

		modal.Call("appendChild", header)
	}

	// Content
	content := document.Call("createElement", "div")
	content.Set("className", "px-6 py-4 overflow-y-auto flex-1")
	if !props.Content.IsUndefined() && !props.Content.IsNull() {
		content.Call("appendChild", props.Content)
	}
	modal.Call("appendChild", content)
	m.content = content

	// Footer
	if !props.Footer.IsUndefined() && !props.Footer.IsNull() {
		footer := document.Call("createElement", "div")
		footer.Set("className", "px-6 py-4 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 rounded-b-lg")
		footer.Call("appendChild", props.Footer)
		modal.Call("appendChild", footer)
	}

	overlay.Call("appendChild", modal)

	// Close on overlay click
	overlay.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		if args[0].Get("target").Equal(overlay) {
			m.Close()
		}
		return nil
	}))

	// Close on Escape key
	if props.CloseOnEsc {
		js.Global().Get("document").Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) any {
			if m.isOpen && args[0].Get("key").String() == "Escape" {
				m.Close()
			}
			return nil
		}))
	}

	// Store onClose callback
	if props.OnClose != nil {
		m.overlay.Set("_onClose", js.FuncOf(func(this js.Value, args []js.Value) any {
			props.OnClose()
			return nil
		}))
	}

	return m
}

// Element returns the overlay DOM element
func (m *Modal) Element() js.Value {
	return m.overlay
}

// Open shows the modal
func (m *Modal) Open() {
	m.overlay.Get("classList").Call("remove", "hidden")
	m.isOpen = true
	// Prevent body scroll
	js.Global().Get("document").Get("body").Get("style").Set("overflow", "hidden")
}

// Close hides the modal
func (m *Modal) Close() {
	m.overlay.Get("classList").Call("add", "hidden")
	m.isOpen = false
	// Restore body scroll
	js.Global().Get("document").Get("body").Get("style").Set("overflow", "")

	// Call onClose callback
	onClose := m.overlay.Get("_onClose")
	if !onClose.IsUndefined() {
		onClose.Invoke()
	}
}

// IsOpen returns whether the modal is currently open
func (m *Modal) IsOpen() bool {
	return m.isOpen
}

// SetContent replaces the modal content
func (m *Modal) SetContent(content js.Value) {
	m.content.Set("innerHTML", "")
	m.content.Call("appendChild", content)
}

// TitleID returns the unique ID used for aria-labelledby
func (m *Modal) TitleID() string {
	return m.titleID
}

// ModalElement returns the inner modal container (for ARIA attribute access)
func (m *Modal) ModalElement() js.Value {
	return m.modal
}
