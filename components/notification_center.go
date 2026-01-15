//go:build js && wasm

package components

import "syscall/js"

// Notification represents a single notification item
type Notification struct {
	ID      string
	Title   string
	Message string
	Time    string
	Read    bool
	Type    string // "info", "success", "warning", "error"
}

// NotificationCenterProps configures a NotificationCenter component
type NotificationCenterProps struct {
	Notifications       []Notification
	OnMarkRead          func(id string)
	OnMarkAllRead       func()
	OnClear             func()
	OnNotificationClick func(id string)
}

// NotificationCenter creates a notification bell with dropdown
type NotificationCenter struct {
	element       js.Value
	dropdown      *Dropdown
	badgeEl       js.Value
	listContainer js.Value
	emptyState    js.Value
	notifications []Notification
	props         NotificationCenterProps
}

// NewNotificationCenter creates a new NotificationCenter component
func NewNotificationCenter(props NotificationCenterProps) *NotificationCenter {
	document := js.Global().Get("document")

	// Create bell button trigger with badge container
	// Use button as outer container so ARIA attributes are valid
	triggerContainer := document.Call("createElement", "button")
	triggerContainer.Set("className", "relative p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-full text-gray-600 dark:text-gray-300")
	triggerContainer.Call("setAttribute", "aria-label", "Notifications")

	// Bell SVG icon
	bellIcon := document.Call("createElement", "span")
	bellIcon.Set("innerHTML", `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"/></svg>`)
	bellIcon.Call("setAttribute", "aria-hidden", "true")

	triggerContainer.Call("appendChild", bellIcon)

	// Unread badge (positioned absolute top-right)
	// aria-hidden: badge is supplementary visual info, not part of accessible name
	badgeEl := Badge(BadgeProps{
		Text:    "0",
		Variant: BadgeError,
		Rounded: true,
	})
	badgeEl.Set("className", "absolute -top-1 -right-1 min-w-[18px] h-[18px] flex items-center justify-center text-xs font-bold bg-red-500 text-white rounded-full hidden")
	badgeEl.Call("setAttribute", "aria-hidden", "true")
	triggerContainer.Call("appendChild", badgeEl)

	// Create dropdown content
	content := document.Call("createElement", "div")
	content.Set("className", "w-80")

	// Header with title and mark all read
	header := document.Call("createElement", "div")
	header.Set("className", "px-4 py-3 flex items-center justify-between border-b border-gray-200 dark:border-gray-700")

	title := document.Call("createElement", "h3")
	title.Set("className", "text-sm font-semibold text-gray-900 dark:text-white")
	title.Set("textContent", "Notifications")
	header.Call("appendChild", title)

	markAllBtn := document.Call("createElement", "button")
	markAllBtn.Set("className", "text-xs text-blue-600 dark:text-blue-400 hover:underline")
	markAllBtn.Set("textContent", "Mark all read")
	if props.OnMarkAllRead != nil {
		markAllBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			args[0].Call("stopPropagation")
			props.OnMarkAllRead()
			return nil
		}))
	}
	header.Call("appendChild", markAllBtn)
	content.Call("appendChild", header)

	// Scrollable notification list
	listContainer := document.Call("createElement", "div")
	listContainer.Set("className", "max-h-80 overflow-y-auto")
	content.Call("appendChild", listContainer)

	// Empty state (hidden by default)
	emptyState := document.Call("createElement", "div")
	emptyState.Set("className", "py-8 text-center text-gray-500 dark:text-gray-400 text-sm hidden")
	emptyState.Set("textContent", "No notifications")
	listContainer.Call("appendChild", emptyState)

	// Footer with clear all
	footer := document.Call("createElement", "div")
	footer.Set("className", "px-4 py-2 border-t border-gray-200 dark:border-gray-700")

	clearBtn := document.Call("createElement", "button")
	clearBtn.Set("className", "w-full text-center text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200")
	clearBtn.Set("textContent", "Clear all")
	if props.OnClear != nil {
		clearBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			args[0].Call("stopPropagation")
			props.OnClear()
			return nil
		}))
	}
	footer.Call("appendChild", clearBtn)
	content.Call("appendChild", footer)

	// Create dropdown
	dropdown := NewDropdown(DropdownProps{
		Trigger: triggerContainer,
		Items:   []DropdownItem{},
		Align:   "right",
	})

	// Replace menu content with custom content
	dropdown.menu.Set("innerHTML", "")
	dropdown.menu.Call("appendChild", content)
	dropdown.menu.Get("style").Set("minWidth", "")
	dropdown.menu.Get("style").Set("width", "20rem")

	nc := &NotificationCenter{
		element:       dropdown.Element(),
		dropdown:      dropdown,
		badgeEl:       badgeEl,
		listContainer: listContainer,
		emptyState:    emptyState,
		notifications: props.Notifications,
		props:         props,
	}

	// Render initial notifications
	nc.renderNotifications()

	return nc
}

// renderNotifications renders the notification list
func (nc *NotificationCenter) renderNotifications() {
	document := js.Global().Get("document")

	// Clear existing items (but keep empty state)
	children := nc.listContainer.Get("children")
	length := children.Get("length").Int()
	for i := length - 1; i >= 0; i-- {
		child := children.Index(i)
		if child.Equal(nc.emptyState) {
			continue
		}
		nc.listContainer.Call("removeChild", child)
	}

	// Show/hide empty state
	if len(nc.notifications) == 0 {
		nc.emptyState.Get("classList").Call("remove", "hidden")
		nc.badgeEl.Get("classList").Call("add", "hidden")
		return
	}

	nc.emptyState.Get("classList").Call("add", "hidden")

	// Calculate unread count
	unreadCount := 0
	for _, n := range nc.notifications {
		if !n.Read {
			unreadCount++
		}
	}

	// Update badge
	if unreadCount > 0 {
		nc.badgeEl.Set("textContent", itoa(unreadCount))
		nc.badgeEl.Get("classList").Call("remove", "hidden")
	} else {
		nc.badgeEl.Get("classList").Call("add", "hidden")
	}

	// Render notification items (insert before empty state)
	for _, notification := range nc.notifications {
		item := nc.createNotificationItem(document, notification)
		nc.listContainer.Call("insertBefore", item, nc.emptyState)
	}
}

// createNotificationItem creates a single notification item element
func (nc *NotificationCenter) createNotificationItem(document js.Value, notification Notification) js.Value {
	item := document.Call("createElement", "div")

	bgClass := "bg-white dark:bg-gray-800"
	if !notification.Read {
		bgClass = "bg-blue-50 dark:bg-gray-750"
	}
	item.Set("className", "px-4 py-3 flex gap-3 hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer border-b border-gray-100 dark:border-gray-700 last:border-b-0 "+bgClass)

	// Type indicator dot
	dot := document.Call("createElement", "div")
	dotColor := "bg-blue-600"
	switch notification.Type {
	case "success":
		dotColor = "bg-green-500"
	case "warning":
		dotColor = "bg-yellow-500"
	case "error":
		dotColor = "bg-red-500"
	}
	dot.Set("className", "w-2 h-2 rounded-full mt-2 flex-shrink-0 "+dotColor)
	item.Call("appendChild", dot)

	// Content
	content := document.Call("createElement", "div")
	content.Set("className", "flex-1 min-w-0")

	titleEl := document.Call("createElement", "div")
	titleClass := "text-sm text-gray-900 dark:text-white truncate"
	if !notification.Read {
		titleClass += " font-semibold"
	}
	titleEl.Set("className", titleClass)
	titleEl.Set("textContent", notification.Title)
	content.Call("appendChild", titleEl)

	messageEl := document.Call("createElement", "div")
	messageEl.Set("className", "text-xs text-gray-500 dark:text-gray-400 truncate mt-0.5")
	messageEl.Set("textContent", notification.Message)
	content.Call("appendChild", messageEl)

	timeEl := document.Call("createElement", "div")
	timeEl.Set("className", "text-xs text-gray-400 dark:text-gray-500 mt-1")
	timeEl.Set("textContent", notification.Time)
	content.Call("appendChild", timeEl)

	item.Call("appendChild", content)

	// Click handlers
	id := notification.ID
	if nc.props.OnNotificationClick != nil {
		item.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			nc.props.OnNotificationClick(id)
			return nil
		}))
	}

	return item
}

// Element returns the DOM element
func (nc *NotificationCenter) Element() js.Value {
	return nc.element
}

// SetNotifications updates the notification list
func (nc *NotificationCenter) SetNotifications(notifications []Notification) {
	nc.notifications = notifications
	nc.renderNotifications()
}

// UnreadCount returns the number of unread notifications
func (nc *NotificationCenter) UnreadCount() int {
	count := 0
	for _, n := range nc.notifications {
		if !n.Read {
			count++
		}
	}
	return count
}

// Open opens the dropdown
func (nc *NotificationCenter) Open() {
	nc.dropdown.Open()
}

// Close closes the dropdown
func (nc *NotificationCenter) Close() {
	nc.dropdown.Close()
}

// Destroy cleans up event listeners
func (nc *NotificationCenter) Destroy() {
	nc.dropdown.Destroy()
}
