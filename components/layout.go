//go:build js && wasm

package components

import "syscall/js"

// LayoutProps configures a Layout component
type LayoutProps struct {
	Sidebar SidebarProps
	Header  HeaderProps
}

// Layout provides a sidebar + header + content layout
type Layout struct {
	element   js.Value
	sidebar   *Sidebar
	header    *Header
	contentEl js.Value
}

// NewLayout creates a new Layout component
func NewLayout(props LayoutProps) *Layout {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "flex h-screen")

	sidebar := NewSidebar(props.Sidebar)

	// Add overlay first (so it's behind sidebar but above content)
	container.Call("appendChild", sidebar.Overlay())
	container.Call("appendChild", sidebar.Element())

	mainArea := document.Call("createElement", "div")
	mainArea.Set("className", "flex-1 flex flex-col overflow-hidden w-full")

	// Add hamburger menu toggle to header props
	headerPropsWithMenu := props.Header
	headerPropsWithMenu.OnMenuToggle = func() {
		sidebar.Toggle()
	}

	header := NewHeader(headerPropsWithMenu)
	mainArea.Call("appendChild", header.Element())

	content := document.Call("createElement", "main")
	content.Set("className", "flex-1 p-4 md:p-6 bg-gray-100 dark:bg-gray-900 overflow-auto")
	mainArea.Call("appendChild", content)

	container.Call("appendChild", mainArea)

	return &Layout{
		element:   container,
		sidebar:   sidebar,
		header:    header,
		contentEl: content,
	}
}

// Element returns the underlying DOM element
func (l *Layout) Element() js.Value {
	return l.element
}

// Sidebar returns the sidebar component
func (l *Layout) Sidebar() *Sidebar {
	return l.sidebar
}

// Header returns the header component
func (l *Layout) Header() *Header {
	return l.header
}

// SetContent replaces the main content area
func (l *Layout) SetContent(content js.Value) {
	l.contentEl.Set("innerHTML", "")
	l.contentEl.Call("appendChild", content)
}

// SetPage is a convenience method that wraps content in a TitledCard
func (l *Layout) SetPage(title, description string, content ...js.Value) {
	l.SetContent(TitledCard(title, description, content...))
}

// SetPageWithHeader sets page content and updates the header title
func (l *Layout) SetPageWithHeader(title, description string, content ...js.Value) {
	l.header.SetTitle(title)
	l.SetContent(TitledCard(title, description, content...))
}
