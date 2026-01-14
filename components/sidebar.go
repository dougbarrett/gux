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
	Title string
	Items []NavItem
}

// Sidebar is a navigation sidebar component
type Sidebar struct {
	element  js.Value
	items    []NavItem
	navItems []js.Value
}

// NewSidebar creates a new Sidebar component
func NewSidebar(props SidebarProps) *Sidebar {
	document := js.Global().Get("document")

	sidebar := document.Call("createElement", "aside")
	sidebar.Set("className", "w-64 bg-gray-800 text-white flex flex-col h-screen")

	// Header
	header := document.Call("createElement", "div")
	header.Set("className", "p-4 border-b border-gray-700")
	title := document.Call("createElement", "h1")
	title.Set("className", "text-xl font-bold")
	title.Set("textContent", props.Title)
	header.Call("appendChild", title)
	sidebar.Call("appendChild", header)

	// Navigation
	nav := document.Call("createElement", "nav")
	nav.Set("className", "flex-1 p-4 space-y-2")

	s := &Sidebar{
		element:  sidebar,
		items:    props.Items,
		navItems: make([]js.Value, len(props.Items)),
	}

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

	return link
}
