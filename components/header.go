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
	Title              string
	Actions            []HeaderAction
	OnMenuToggle       func() // Called when hamburger menu is clicked (mobile)
	UserMenu           *UserMenu
	NotificationCenter *NotificationCenter
	ConnectionStatus   *ConnectionStatus
}

// Header is a page header component
type Header struct {
	element            js.Value
	titleEl            js.Value
	actionsEl          js.Value
	titleText          string
	actions            []HeaderAction
	userMenu           *UserMenu
	notificationCenter *NotificationCenter
	connectionStatus   *ConnectionStatus
}

// NewHeader creates a new Header component
func NewHeader(props HeaderProps) *Header {
	document := js.Global().Get("document")

	header := document.Call("createElement", "header")
	header.Set("className", "bg-white dark:bg-gray-800 shadow dark:shadow-gray-900 px-4 md:px-6 py-4 flex justify-between items-center")

	// Left side: hamburger menu (mobile) + title
	leftDiv := document.Call("createElement", "div")
	leftDiv.Set("className", "flex items-center gap-3")

	// Hamburger menu button (mobile only)
	if props.OnMenuToggle != nil {
		menuBtn := document.Call("createElement", "button")
		menuBtn.Set("className", "md:hidden p-2 -ml-2 text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors")
		menuBtn.Set("innerHTML", `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/></svg>`)
		menuBtn.Set("ariaLabel", "Open menu")
		menuBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			props.OnMenuToggle()
			return nil
		}))
		leftDiv.Call("appendChild", menuBtn)
	}

	title := document.Call("createElement", "h1")
	title.Set("className", "text-xl md:text-2xl font-semibold text-gray-800 dark:text-gray-100 truncate")
	title.Set("textContent", props.Title)
	leftDiv.Call("appendChild", title)

	header.Call("appendChild", leftDiv)

	actionsDiv := document.Call("createElement", "div")
	actionsDiv.Set("className", "flex items-center gap-3 flex-shrink-0")
	header.Call("appendChild", actionsDiv)

	// Add NotificationCenter if provided
	if props.NotificationCenter != nil {
		actionsDiv.Call("appendChild", props.NotificationCenter.Element())
	}

	// Add ConnectionStatus if provided
	if props.ConnectionStatus != nil {
		actionsDiv.Call("appendChild", props.ConnectionStatus.Element())
	}

	// Add UserMenu if provided
	if props.UserMenu != nil {
		actionsDiv.Call("appendChild", props.UserMenu.Element())
	}

	// Add action buttons wrapper with tighter gap
	buttonsDiv := document.Call("createElement", "div")
	buttonsDiv.Set("className", "flex gap-2")
	actionsDiv.Call("appendChild", buttonsDiv)

	h := &Header{
		element:            header,
		titleEl:            title,
		actionsEl:          actionsDiv,
		titleText:          props.Title,
		actions:            props.Actions,
		userMenu:           props.UserMenu,
		notificationCenter: props.NotificationCenter,
		connectionStatus:   props.ConnectionStatus,
	}

	for _, action := range props.Actions {
		btn := Button(ButtonProps{
			Text:      action.Label,
			ClassName: "px-3 py-1 text-sm bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-200 rounded transition-colors cursor-pointer",
			OnClick:   action.OnClick,
		})
		buttonsDiv.Call("appendChild", btn)
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

// UserMenu returns the UserMenu component if set
func (h *Header) UserMenu() *UserMenu {
	return h.userMenu
}

// NotificationCenter returns the NotificationCenter component if set
func (h *Header) NotificationCenter() *NotificationCenter {
	return h.notificationCenter
}

// ConnectionStatus returns the ConnectionStatus component if set
func (h *Header) ConnectionStatus() *ConnectionStatus {
	return h.connectionStatus
}
