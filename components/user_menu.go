//go:build js && wasm

package components

import "syscall/js"

// UserMenuProps configures a UserMenu component
type UserMenuProps struct {
	Name       string
	Email      string
	AvatarSrc  string // Optional - falls back to initials
	OnProfile  func()
	OnSettings func()
	OnLogout   func()
}

// UserMenu creates a user profile dropdown with avatar trigger
type UserMenu struct {
	element  js.Value
	dropdown *Dropdown
}

// NewUserMenu creates a new UserMenu component
func NewUserMenu(props UserMenuProps) *UserMenu {
	document := js.Global().Get("document")

	// Create button trigger wrapping avatar (button needed for valid ARIA)
	trigger := document.Call("createElement", "button")
	trigger.Set("className", "rounded-full focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2")
	trigger.Call("setAttribute", "aria-label", "User menu for "+props.Name)

	// Create avatar inside the button
	avatar := Avatar(AvatarProps{
		Src:  props.AvatarSrc,
		Name: props.Name,
		Size: AvatarSM,
	})
	trigger.Call("appendChild", avatar)

	// Create dropdown content container
	content := document.Call("createElement", "div")
	content.Set("className", "w-64")

	// Header section with larger avatar + name + email
	header := document.Call("createElement", "div")
	header.Set("className", "px-4 py-3 flex items-center gap-3")

	headerAvatar := Avatar(AvatarProps{
		Src:  props.AvatarSrc,
		Name: props.Name,
		Size: AvatarMD,
	})
	header.Call("appendChild", headerAvatar)

	headerInfo := document.Call("createElement", "div")
	headerInfo.Set("className", "flex-1 min-w-0")

	nameEl := document.Call("createElement", "div")
	nameEl.Set("className", "text-sm font-medium text-gray-900 dark:text-white truncate")
	nameEl.Set("textContent", props.Name)
	headerInfo.Call("appendChild", nameEl)

	emailEl := document.Call("createElement", "div")
	emailEl.Set("className", "text-xs text-gray-500 dark:text-gray-400 truncate")
	emailEl.Set("textContent", props.Email)
	headerInfo.Call("appendChild", emailEl)

	header.Call("appendChild", headerInfo)
	content.Call("appendChild", header)

	// Divider
	divider1 := document.Call("createElement", "div")
	divider1.Set("className", "border-t border-gray-200 dark:border-gray-700")
	content.Call("appendChild", divider1)

	// Menu items
	menuItems := document.Call("createElement", "div")
	menuItems.Set("className", "py-1")

	// Profile item
	profileItem := createMenuItem(document, "👤", "Profile", false, props.OnProfile)
	menuItems.Call("appendChild", profileItem)

	// Settings item
	settingsItem := createMenuItem(document, "⚙️", "Settings", false, props.OnSettings)
	menuItems.Call("appendChild", settingsItem)

	content.Call("appendChild", menuItems)

	// Divider before logout
	divider2 := document.Call("createElement", "div")
	divider2.Set("className", "border-t border-gray-200 dark:border-gray-700")
	content.Call("appendChild", divider2)

	// Logout section
	logoutSection := document.Call("createElement", "div")
	logoutSection.Set("className", "py-1")

	logoutItem := createMenuItem(document, "🚪", "Logout", true, props.OnLogout)
	logoutSection.Call("appendChild", logoutItem)

	content.Call("appendChild", logoutSection)

	// Create dropdown with custom content
	dropdown := NewDropdown(DropdownProps{
		Trigger: trigger,
		Items:   []DropdownItem{}, // Empty - we use custom content
		Align:   "right",
	})

	// Replace the empty menu content with our custom content
	dropdown.menu.Set("innerHTML", "")
	dropdown.menu.Call("appendChild", content)
	dropdown.menu.Get("style").Set("minWidth", "")
	dropdown.menu.Get("style").Set("width", "16rem")

	return &UserMenu{
		element:  dropdown.Element(),
		dropdown: dropdown,
	}
}

// createMenuItem creates a menu item button
func createMenuItem(document js.Value, icon, label string, danger bool, onClick func()) js.Value {
	item := document.Call("createElement", "button")

	baseClass := "w-full text-left px-4 py-2 text-sm flex items-center gap-2 hover:bg-gray-100 dark:hover:bg-gray-700"
	if danger {
		baseClass += " text-red-600 dark:text-red-400"
	} else {
		baseClass += " text-gray-700 dark:text-gray-200"
	}
	item.Set("className", baseClass)

	iconSpan := document.Call("createElement", "span")
	iconSpan.Set("textContent", icon)
	item.Call("appendChild", iconSpan)

	labelSpan := document.Call("createElement", "span")
	labelSpan.Set("textContent", label)
	item.Call("appendChild", labelSpan)

	if onClick != nil {
		item.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			onClick()
			return nil
		}))
	}

	return item
}

// Element returns the DOM element
func (u *UserMenu) Element() js.Value {
	return u.element
}

// Open opens the dropdown menu
func (u *UserMenu) Open() {
	u.dropdown.Open()
}

// Close closes the dropdown menu
func (u *UserMenu) Close() {
	u.dropdown.Close()
}

// Destroy cleans up event listeners
func (u *UserMenu) Destroy() {
	u.dropdown.Destroy()
}
