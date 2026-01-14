//go:build js && wasm

package components

import "syscall/js"

type NavItem struct {
	Label string
	Icon  string
	Path  string // Use for routing
}

type SidebarProps struct {
	Title string
	Items []NavItem
}

type Sidebar struct {
	element  js.Value
	items    []NavItem
	navItems []js.Value
}

const (
	sidebarClass       = "w-64 bg-gray-900 text-white min-h-screen flex flex-col"
	sidebarTitleClass  = "p-4 text-xl font-bold border-b border-gray-700"
	sidebarNavClass    = "flex-1 py-4"
	sidebarItemClass   = "flex items-center px-4 py-3 text-gray-300 hover:bg-gray-800 hover:text-white cursor-pointer transition-colors no-underline"
	sidebarActiveClass = "flex items-center px-4 py-3 bg-gray-800 text-white cursor-pointer border-l-4 border-blue-500 no-underline"
)

func NewSidebar(props SidebarProps) *Sidebar {
	document := js.Global().Get("document")

	sidebar := document.Call("createElement", "aside")
	sidebar.Set("className", sidebarClass)

	// Title/Logo
	title := document.Call("createElement", "div")
	title.Set("className", sidebarTitleClass)
	title.Set("textContent", props.Title)
	sidebar.Call("appendChild", title)

	// Navigation
	nav := document.Call("createElement", "nav")
	nav.Set("className", sidebarNavClass)

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

func (s *Sidebar) Element() js.Value {
	return s.element
}

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
	// Create link element
	link := Link(LinkProps{
		To:        item.Path,
		ClassName: sidebarItemClass,
	})

	// Icon
	if item.Icon != "" {
		icon := document.Call("createElement", "span")
		icon.Set("className", "mr-3")
		icon.Set("textContent", item.Icon)
		link.Call("appendChild", icon)
	}

	// Label
	label := document.Call("createElement", "span")
	label.Set("textContent", item.Label)
	link.Call("appendChild", label)

	return link
}
