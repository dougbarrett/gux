//go:build js && wasm

package components

import "syscall/js"

const (
	// Expanded mode classes
	sidebarItemClass   = "flex items-center gap-3 px-4 py-3 text-gray-300 hover:bg-gray-700 hover:text-white rounded-lg transition-colors cursor-pointer"
	sidebarActiveClass = "flex items-center gap-3 px-4 py-3 bg-gray-700 text-white rounded-lg cursor-pointer"
	// Collapsed mode classes (icon centered, no text)
	sidebarItemCollapsedClass   = "flex items-center justify-center px-2 py-3 text-gray-300 hover:bg-gray-700 hover:text-white rounded-lg transition-colors cursor-pointer"
	sidebarActiveCollapsedClass = "flex items-center justify-center px-2 py-3 bg-gray-700 text-white rounded-lg cursor-pointer"
)

// NavItem represents a navigation menu item
type NavItem struct {
	Label string
	Icon  string
	Path  string
}

// SidebarProps configures a Sidebar component
type SidebarProps struct {
	Title      string
	Items      []NavItem
	OnToggle   func(isOpen bool)      // Called when sidebar is toggled on mobile
	OnCollapse func(isCollapsed bool) // Called when sidebar is collapsed/expanded on desktop
}

// Sidebar is a navigation sidebar component
type Sidebar struct {
	element          js.Value
	overlay          js.Value
	header           js.Value
	title            js.Value // Store title for show/hide on collapse
	nav              js.Value
	items            []NavItem
	navItems         []js.Value
	navLabels        []js.Value // Store label elements for show/hide on collapse
	isOpen           bool
	isCollapsed      bool
	onToggle         func(isOpen bool)
	onCollapse       func(isCollapsed bool)
	collapseBtn      js.Value
	keyboardShortcut js.Func // Stored for cleanup
}

// NewSidebar creates a new Sidebar component
func NewSidebar(props SidebarProps) *Sidebar {
	document := js.Global().Get("document")

	// Create overlay for mobile (click to close sidebar)
	overlay := document.Call("createElement", "div")
	overlay.Set("className", "fixed inset-0 bg-black bg-opacity-50 z-40 hidden md:hidden")

	sidebar := document.Call("createElement", "aside")
	// Mobile: fixed position, hidden by default with transform
	// Desktop (md+): static position, always visible, width transitions for collapse
	sidebar.Set("className", "fixed md:static inset-y-0 left-0 z-50 w-64 bg-gray-800 text-white flex flex-col h-screen transform -translate-x-full md:translate-x-0 transition-all duration-300 ease-in-out")

	// Header with collapse toggle and close button
	header := document.Call("createElement", "div")
	header.Set("className", "p-4 border-b border-gray-700 flex items-center justify-between")

	title := document.Call("createElement", "h1")
	title.Set("className", "text-xl font-bold whitespace-nowrap overflow-hidden")
	title.Set("textContent", props.Title)
	header.Call("appendChild", title)

	// Button container for collapse and close buttons
	btnContainer := document.Call("createElement", "div")
	btnContainer.Set("className", "flex items-center gap-1")

	// Collapse toggle button (desktop only) - chevron-left when expanded
	collapseBtn := document.Call("createElement", "button")
	collapseBtn.Set("className", "hidden md:block p-2 text-gray-400 hover:text-white rounded-lg hover:bg-gray-700 transition-colors")
	collapseBtn.Set("innerHTML", `<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>`)
	collapseBtn.Set("ariaLabel", "Collapse sidebar")
	btnContainer.Call("appendChild", collapseBtn)

	// Close button (mobile only)
	closeBtn := document.Call("createElement", "button")
	closeBtn.Set("className", "md:hidden p-2 text-gray-400 hover:text-white rounded-lg hover:bg-gray-700 transition-colors")
	closeBtn.Set("innerHTML", `<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>`)
	closeBtn.Set("ariaLabel", "Close menu")
	btnContainer.Call("appendChild", closeBtn)

	header.Call("appendChild", btnContainer)
	sidebar.Call("appendChild", header)

	// Navigation
	nav := document.Call("createElement", "nav")
	nav.Set("className", "flex-1 p-4 space-y-2 overflow-y-auto")

	s := &Sidebar{
		element:     sidebar,
		overlay:     overlay,
		header:      header,
		title:       title,
		nav:         nav,
		items:       props.Items,
		navItems:    make([]js.Value, len(props.Items)),
		navLabels:   make([]js.Value, len(props.Items)),
		isOpen:      false,
		isCollapsed: false,
		onToggle:    props.OnToggle,
		onCollapse:  props.OnCollapse,
		collapseBtn: collapseBtn,
	}

	// Collapse button click handler
	collapseBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		s.ToggleCollapse()
		return nil
	}))

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
		navItem, label := s.createNavItemWithLabel(document, item)
		s.navItems[i] = navItem
		s.navLabels[i] = label
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
			if s.isCollapsed {
				s.navItems[i].Set("className", sidebarActiveCollapsedClass)
			} else {
				s.navItems[i].Set("className", sidebarActiveClass)
			}
		} else {
			if s.isCollapsed {
				s.navItems[i].Set("className", sidebarItemCollapsedClass)
			} else {
				s.navItems[i].Set("className", sidebarItemClass)
			}
		}
	}
}

// createNavItemWithLabel creates a nav item and returns both the link element and the label span
func (s *Sidebar) createNavItemWithLabel(document js.Value, item NavItem) (js.Value, js.Value) {
	link := Link(LinkProps{
		To:        item.Path,
		ClassName: sidebarItemClass,
	})

	// Set relative positioning for tooltip
	link.Get("style").Set("position", "relative")

	if item.Icon != "" {
		icon := document.Call("createElement", "span")
		icon.Set("textContent", item.Icon)
		icon.Set("className", "flex-shrink-0")
		link.Call("appendChild", icon)
	}

	label := document.Call("createElement", "span")
	label.Set("textContent", item.Label)
	label.Set("className", "whitespace-nowrap overflow-hidden transition-opacity duration-200")
	link.Call("appendChild", label)

	// Create tooltip for collapsed state (hidden by default)
	tooltip := document.Call("createElement", "div")
	tooltip.Set("className", "absolute left-full ml-2 px-2 py-1 bg-gray-900 text-white text-sm rounded whitespace-nowrap opacity-0 pointer-events-none transition-opacity z-50")
	tooltip.Set("textContent", item.Label)
	link.Call("appendChild", tooltip)

	// Show tooltip on hover when collapsed
	link.Call("addEventListener", "mouseenter", js.FuncOf(func(this js.Value, args []js.Value) any {
		if s.isCollapsed {
			tooltip.Set("className", "absolute left-full ml-2 px-2 py-1 bg-gray-900 text-white text-sm rounded whitespace-nowrap opacity-100 pointer-events-none transition-opacity z-50")
		}
		return nil
	}))
	link.Call("addEventListener", "mouseleave", js.FuncOf(func(this js.Value, args []js.Value) any {
		tooltip.Set("className", "absolute left-full ml-2 px-2 py-1 bg-gray-900 text-white text-sm rounded whitespace-nowrap opacity-0 pointer-events-none transition-opacity z-50")
		return nil
	}))

	// Close sidebar on mobile when a nav item is clicked
	link.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		// Check if we're on mobile (sidebar is in fixed position mode)
		if s.isOpen {
			s.Close()
		}
		return nil
	}))

	return link, label
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

// Collapse collapses the sidebar to icons-only mode (desktop)
func (s *Sidebar) Collapse() {
	s.isCollapsed = true
	// Update sidebar width: w-16 for collapsed (icons-only)
	s.element.Set("className", "fixed md:static inset-y-0 left-0 z-50 w-16 bg-gray-800 text-white flex flex-col h-screen transform -translate-x-full md:translate-x-0 transition-all duration-300 ease-in-out")

	// Update header padding for collapsed state
	s.header.Set("className", "p-2 border-b border-gray-700 flex flex-col items-center gap-2")

	// Hide title when collapsed
	s.title.Set("className", "hidden")

	// Update nav padding
	s.nav.Set("className", "flex-1 p-2 space-y-2 overflow-y-auto")

	// Update collapse button icon to chevron-right (expand)
	s.collapseBtn.Set("innerHTML", `<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>`)
	s.collapseBtn.Set("ariaLabel", "Expand sidebar")

	// Hide labels, update nav item classes
	for i, label := range s.navLabels {
		label.Set("className", "hidden")
		// Update nav item class for collapsed mode
		if s.navItems[i].Get("className").String() == sidebarActiveClass {
			s.navItems[i].Set("className", sidebarActiveCollapsedClass)
		} else {
			s.navItems[i].Set("className", sidebarItemCollapsedClass)
		}
	}

	if s.onCollapse != nil {
		s.onCollapse(true)
	}
}

// Expand expands the sidebar to full width (desktop)
func (s *Sidebar) Expand() {
	s.isCollapsed = false
	// Update sidebar width: w-64 for expanded
	s.element.Set("className", "fixed md:static inset-y-0 left-0 z-50 w-64 bg-gray-800 text-white flex flex-col h-screen transform -translate-x-full md:translate-x-0 transition-all duration-300 ease-in-out")

	// Update header padding for expanded state
	s.header.Set("className", "p-4 border-b border-gray-700 flex items-center justify-between")

	// Show title when expanded
	s.title.Set("className", "text-xl font-bold whitespace-nowrap overflow-hidden")

	// Update nav padding
	s.nav.Set("className", "flex-1 p-4 space-y-2 overflow-y-auto")

	// Update collapse button icon to chevron-left (collapse)
	s.collapseBtn.Set("innerHTML", `<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>`)
	s.collapseBtn.Set("ariaLabel", "Collapse sidebar")

	// Show labels, update nav item classes
	for i, label := range s.navLabels {
		label.Set("className", "whitespace-nowrap overflow-hidden transition-opacity duration-200")
		// Update nav item class for expanded mode
		if s.navItems[i].Get("className").String() == sidebarActiveCollapsedClass {
			s.navItems[i].Set("className", sidebarActiveClass)
		} else {
			s.navItems[i].Set("className", sidebarItemClass)
		}
	}

	if s.onCollapse != nil {
		s.onCollapse(false)
	}
}

// ToggleCollapse toggles the sidebar between collapsed and expanded states
func (s *Sidebar) ToggleCollapse() {
	if s.isCollapsed {
		s.Expand()
	} else {
		s.Collapse()
	}
}

// IsCollapsed returns whether the sidebar is currently collapsed
func (s *Sidebar) IsCollapsed() bool {
	return s.isCollapsed
}

// RegisterKeyboardShortcut registers Cmd/Ctrl+B to toggle sidebar collapse
func (s *Sidebar) RegisterKeyboardShortcut() {
	document := js.Global().Get("document")

	s.keyboardShortcut = js.FuncOf(func(this js.Value, args []js.Value) any {
		event := args[0]
		key := event.Get("key").String()

		// Check for Cmd+B (Mac) or Ctrl+B (Windows/Linux)
		// event.key returns lowercase "b" for the B key
		if (key == "b" || key == "B") && (event.Get("metaKey").Bool() || event.Get("ctrlKey").Bool()) {
			event.Call("preventDefault") // Prevent browser default (e.g., bold in rich text)
			s.ToggleCollapse()
		}
		return nil
	})

	document.Call("addEventListener", "keydown", s.keyboardShortcut)
}

// UnregisterKeyboardShortcut removes the Cmd/Ctrl+B keyboard shortcut
func (s *Sidebar) UnregisterKeyboardShortcut() {
	if s.keyboardShortcut.Truthy() {
		document := js.Global().Get("document")
		document.Call("removeEventListener", "keydown", s.keyboardShortcut)
		s.keyboardShortcut.Release()
	}
}
