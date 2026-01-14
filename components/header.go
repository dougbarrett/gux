//go:build js && wasm

package components

import "syscall/js"

// HeaderAction represents a header action button
type HeaderAction struct {
	Label   string
	OnClick func()
}

// HeaderProps configures a Header component
type HeaderProps struct {
	Title   string
	Actions []HeaderAction
}

// Header is a page header component
type Header struct {
	element    js.Value
	titleEl    js.Value
	actionsEl  js.Value
	titleText  string
	actions    []HeaderAction
}

// NewHeader creates a new Header component
func NewHeader(props HeaderProps) *Header {
	document := js.Global().Get("document")

	header := document.Call("createElement", "header")
	header.Set("className", "bg-white dark:bg-gray-800 shadow dark:shadow-gray-900 px-6 py-4 flex justify-between items-center")

	title := document.Call("createElement", "h1")
	title.Set("className", "text-2xl font-semibold text-gray-800 dark:text-gray-100")
	title.Set("textContent", props.Title)
	header.Call("appendChild", title)

	actionsDiv := document.Call("createElement", "div")
	actionsDiv.Set("className", "flex gap-2")
	header.Call("appendChild", actionsDiv)

	h := &Header{
		element:   header,
		titleEl:   title,
		actionsEl: actionsDiv,
		titleText: props.Title,
		actions:   props.Actions,
	}

	for _, action := range props.Actions {
		btn := Button(ButtonProps{
			Text:      action.Label,
			ClassName: "px-3 py-1 text-sm bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-200 rounded transition-colors cursor-pointer",
			OnClick:   action.OnClick,
		})
		actionsDiv.Call("appendChild", btn)
	}

	return h
}

// Element returns the underlying DOM element
func (h *Header) Element() js.Value {
	return h.element
}

// SetTitle updates the header title
func (h *Header) SetTitle(title string) {
	h.titleText = title
	h.titleEl.Set("textContent", title)
}
