//go:build js && wasm

package components

import "syscall/js"

const (
	sidebarItemClass   = "flex items-center gap-3 px-4 py-3 text-gray-300 hover:bg-gray-700 hover:text-white rounded-lg transition-colors cursor-pointer"
	sidebarActiveClass = "flex items-center gap-3 px-4 py-3 bg-gray-700 text-white rounded-lg cursor-pointer"
)

// NavItem represents a navigation menu item
type NavItem struct {
	Label string
	Icon  string
	Path  string
}

// SidebarProps configures a Sidebar component
type SidebarProps struct {
	Title    string
	Items    []NavItem
	OnToggle func(isOpen bool) // Called when sidebar is toggled on mobile
}

// Sidebar is a navigation sidebar component
type Sidebar struct {
	element   js.Value
	overlay   js.Value
	items     []NavItem
	navItems  []js.Value
	isOpen    bool
	onToggle  func(isOpen bool)
}

// NewSidebar creates a new Sidebar component
func NewSidebar(props SidebarProps) *Sidebar {
	document := js.Global().Get("document")

	// Create overlay for mobile (click to close sidebar)
	overlay := document.Call("createElement", "div")
	overlay.Set("className", "fixed inset-0 bg-black bg-opacity-50 z-40 hidden md:hidden")

	sidebar := document.Call("createElement", "aside")
	// Mobile: fixed position, hidden by default with transform
	// Desktop (md+): static position, always visible
	sidebar.Set("className", "fixed md:static inset-y-0 left-0 z-50 w-64 bg-gray-800 text-white flex flex-col h-screen transform -translate-x-full md:translate-x-0 transition-transform duration-300 ease-in-out")

	// Header with close button for mobile
	header := document.Call("createElement", "div")
	header.Set("className", "p-4 border-b border-gray-700 flex items-center justify-between")

	title := document.Call("createElement", "h1")
	title.Set("className", "text-xl font-bold")
	title.Set("textContent", props.Title)
	header.Call("appendChild", title)

	// Close button (mobile only)
	closeBtn := document.Call("createElement", "button")
	closeBtn.Set("className", "md:hidden p-2 text-gray-400 hover:text-white rounded-lg hover:bg-gray-700 transition-colors")
	closeBtn.Set("innerHTML", `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>`)
	closeBtn.Set("ariaLabel", "Close menu")
	header.Call("appendChild", closeBtn)

	sidebar.Call("appendChild", header)

	// Navigation
	nav := document.Call("createElement", "nav")
	nav.Set("className", "flex-1 p-4 space-y-2 overflow-y-auto")

	s := &Sidebar{
		element:  sidebar,
		overlay:  overlay,
		items:    props.Items,
		navItems: make([]js.Value, len(props.Items)),
		isOpen:   false,
		onToggle: props.OnToggle,
	}

	// Close button click handler
	closeBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		s.Close()
		return nil
	}))

	// Overlay click handler (close sidebar)
	overlay.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		s.Close()
		return nil
	}))

	for i, item := range props.Items {
		navItem := s.createNavItem(document, item)
		s.navItems[i] = navItem
		nav.Call("appendChild", navItem)
	}

	sidebar.Call("appendChild", nav)

	return s
}

// Element returns the underlying DOM element
func (s *Sidebar) Element() js.Value {
	return s.element
}

// SetActive updates the active state of nav items
func (s *Sidebar) SetActive(path string) {
	for i, item := range s.items {
		if item.Path == path {
			s.navItems[i].Set("className", sidebarActiveClass)
		} else {
			s.navItems[i].Set("className", sidebarItemClass)
		}
	}
}

func (s *Sidebar) createNavItem(document js.Value, item NavItem) js.Value {
	link := Link(LinkProps{
		To:        item.Path,
		ClassName: sidebarItemClass,
	})

	if item.Icon != "" {
		icon := document.Call("createElement", "span")
		icon.Set("textContent", item.Icon)
		link.Call("appendChild", icon)
	}

	label := document.Call("createElement", "span")
	label.Set("textContent", item.Label)
	link.Call("appendChild", label)

	// Close sidebar on mobile when a nav item is clicked
	link.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		// Check if we're on mobile (sidebar is in fixed position mode)
		if s.isOpen {
			s.Close()
		}
		return nil
	}))

	return link
}

// Overlay returns the overlay element (to be added to DOM)
func (s *Sidebar) Overlay() js.Value {
	return s.overlay
}

// Open opens the sidebar on mobile
func (s *Sidebar) Open() {
	s.isOpen = true
	// Remove -translate-x-full to show sidebar
	s.element.Set("className", "fixed md:static inset-y-0 left-0 z-50 w-64 bg-gray-800 text-white flex flex-col h-screen transform translate-x-0 transition-transform duration-300 ease-in-out")
	// Show overlay
	s.overlay.Set("className", "fixed inset-0 bg-black bg-opacity-50 z-40 block md:hidden")

	if s.onToggle != nil {
		s.onToggle(true)
	}
}

// Close closes the sidebar on mobile
func (s *Sidebar) Close() {
	s.isOpen = false
	// Add -translate-x-full to hide sidebar
	s.element.Set("className", "fixed md:static inset-y-0 left-0 z-50 w-64 bg-gray-800 text-white flex flex-col h-screen transform -translate-x-full md:translate-x-0 transition-transform duration-300 ease-in-out")
	// Hide overlay
	s.overlay.Set("className", "fixed inset-0 bg-black bg-opacity-50 z-40 hidden md:hidden")

	if s.onToggle != nil {
		s.onToggle(false)
	}
}

// Toggle toggles the sidebar open/closed state
func (s *Sidebar) Toggle() {
	if s.isOpen {
		s.Close()
	} else {
		s.Open()
	}
}

// IsOpen returns whether the sidebar is currently open
func (s *Sidebar) IsOpen() bool {
	return s.isOpen
}
