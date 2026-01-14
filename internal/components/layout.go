//go:build js && wasm

package components

import "syscall/js"

type LayoutProps struct {
	Sidebar SidebarProps
	Header  HeaderProps
}

type Layout struct {
	element js.Value
	content js.Value
	sidebar *Sidebar
	header  js.Value
}

const (
	layoutClass  = "flex min-h-screen bg-gray-50"
	mainClass    = "flex-1 flex flex-col"
	contentClass = "flex-1 p-6 overflow-auto"
)

func NewLayout(props LayoutProps) *Layout {
	document := js.Global().Get("document")

	// Main layout container
	layout := document.Call("createElement", "div")
	layout.Set("className", layoutClass)

	// Sidebar
	sidebar := NewSidebar(props.Sidebar)
	layout.Call("appendChild", sidebar.Element())

	// Main area (header + content)
	main := document.Call("createElement", "main")
	main.Set("className", mainClass)

	// Header
	header := Header(props.Header)
	main.Call("appendChild", header)

	// Content area
	content := document.Call("createElement", "div")
	content.Set("className", contentClass)
	main.Call("appendChild", content)

	layout.Call("appendChild", main)

	return &Layout{
		element: layout,
		content: content,
		sidebar: sidebar,
		header:  header,
	}
}

func (l *Layout) Element() js.Value {
	return l.element
}

func (l *Layout) Content() js.Value {
	return l.content
}

func (l *Layout) Sidebar() *Sidebar {
	return l.sidebar
}

func (l *Layout) SetContent(elements ...js.Value) {
	l.content.Set("innerHTML", "")
	for _, el := range elements {
		l.content.Call("appendChild", el)
	}
}

func (l *Layout) AppendContent(element js.Value) {
	l.content.Call("appendChild", element)
}
