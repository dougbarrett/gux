//go:build js && wasm

package components

import (
	"fmt"
	"syscall/js"
)

// ComponentNode represents a node in the component tree
type ComponentNode struct {
	Name       string
	Type       string
	Props      map[string]any
	Element    js.Value
	Children   []*ComponentNode
	Expanded   bool
}

// InspectorProps configures the Inspector component
type InspectorProps struct {
	Position  string // "bottom-right", "bottom-left", "top-right", "top-left"
	Width     string
	Height    string
	Collapsed bool
}

// Inspector provides a debug tool for viewing component hierarchy
type Inspector struct {
	container    js.Value
	panel        js.Value
	treeView     js.Value
	propsView    js.Value
	isOpen       bool
	root         *ComponentNode
	selectedNode *ComponentNode
	toggle       js.Value
}

var globalInspector *Inspector

// NewInspector creates a new Inspector component
func NewInspector(props InspectorProps) *Inspector {
	document := js.Global().Get("document")

	if props.Position == "" {
		props.Position = "bottom-right"
	}
	if props.Width == "" {
		props.Width = "400px"
	}
	if props.Height == "" {
		props.Height = "300px"
	}

	i := &Inspector{
		isOpen: !props.Collapsed,
	}

	// Container
	container := document.Call("createElement", "div")
	container.Set("id", "goquery-inspector")
	container.Set("className", "fixed z-[9999]")

	// Position
	switch props.Position {
	case "bottom-right":
		container.Get("style").Set("bottom", "0")
		container.Get("style").Set("right", "0")
	case "bottom-left":
		container.Get("style").Set("bottom", "0")
		container.Get("style").Set("left", "0")
	case "top-right":
		container.Get("style").Set("top", "0")
		container.Get("style").Set("right", "0")
	case "top-left":
		container.Get("style").Set("top", "0")
		container.Get("style").Set("left", "0")
	}

	i.container = container

	// Toggle button
	toggle := document.Call("createElement", "button")
	toggle.Set("className", "absolute -top-8 right-2 bg-purple-600 text-white px-3 py-1 rounded-t text-sm font-mono")
	toggle.Set("textContent", "🔍 Inspector")
	toggle.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		i.Toggle()
		return nil
	}))
	container.Call("appendChild", toggle)
	i.toggle = toggle

	// Panel
	panel := document.Call("createElement", "div")
	panel.Set("className", "bg-gray-900 text-gray-100 font-mono text-xs shadow-lg border-t border-purple-500")
	panel.Get("style").Set("width", props.Width)
	panel.Get("style").Set("height", props.Height)
	if props.Collapsed {
		panel.Get("style").Set("display", "none")
	}

	// Header
	header := document.Call("createElement", "div")
	header.Set("className", "flex items-center justify-between bg-gray-800 px-3 py-2 border-b border-gray-700")

	title := document.Call("createElement", "span")
	title.Set("className", "text-purple-400 font-bold")
	title.Set("textContent", "GoQuery Inspector")
	header.Call("appendChild", title)

	headerButtons := document.Call("createElement", "div")
	headerButtons.Set("className", "flex gap-2")

	refreshBtn := document.Call("createElement", "button")
	refreshBtn.Set("className", "text-gray-400 hover:text-white")
	refreshBtn.Set("textContent", "↻")
	refreshBtn.Set("title", "Refresh")
	refreshBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		i.Refresh()
		return nil
	}))
	headerButtons.Call("appendChild", refreshBtn)

	closeBtn := document.Call("createElement", "button")
	closeBtn.Set("className", "text-gray-400 hover:text-white")
	closeBtn.Set("textContent", "×")
	closeBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		i.Close()
		return nil
	}))
	headerButtons.Call("appendChild", closeBtn)

	header.Call("appendChild", headerButtons)
	panel.Call("appendChild", header)

	// Content area
	content := document.Call("createElement", "div")
	content.Set("className", "flex h-full")
	content.Get("style").Set("height", "calc(100% - 36px)")

	// Tree view
	treeView := document.Call("createElement", "div")
	treeView.Set("className", "w-1/2 overflow-auto border-r border-gray-700 p-2")
	i.treeView = treeView
	content.Call("appendChild", treeView)

	// Props view
	propsView := document.Call("createElement", "div")
	propsView.Set("className", "w-1/2 overflow-auto p-2")
	i.propsView = propsView
	content.Call("appendChild", propsView)

	panel.Call("appendChild", content)
	container.Call("appendChild", panel)
	i.panel = panel

	// Append to body
	document.Get("body").Call("appendChild", container)

	// Initial scan
	i.Refresh()

	return i
}

// Refresh scans the DOM and updates the component tree
func (i *Inspector) Refresh() {
	document := js.Global().Get("document")
	appEl := document.Call("getElementById", "app")

	if appEl.IsNull() || appEl.IsUndefined() {
		return
	}

	i.root = i.scanElement(appEl, 0)
	i.renderTree()
}

func (i *Inspector) scanElement(el js.Value, depth int) *ComponentNode {
	if el.IsNull() || el.IsUndefined() {
		return nil
	}

	tagName := el.Get("tagName")
	if tagName.IsUndefined() || tagName.IsNull() {
		return nil
	}

	node := &ComponentNode{
		Name:     tagName.String(),
		Type:     "element",
		Element:  el,
		Props:    make(map[string]any),
		Expanded: depth < 2,
	}

	// Get element info
	className := el.Get("className").String()
	if className != "" {
		node.Props["className"] = className
	}

	id := el.Get("id").String()
	if id != "" {
		node.Props["id"] = id
		node.Name = tagName.String() + "#" + id
	}

	role := el.Call("getAttribute", "role")
	if !role.IsNull() && role.String() != "" {
		node.Props["role"] = role.String()
		node.Type = role.String()
	}

	// Detect component type from class names
	node.Type = i.detectComponentType(className)

	// Scan children
	children := el.Get("children")
	if !children.IsUndefined() && !children.IsNull() {
		for j := 0; j < children.Length(); j++ {
			child := children.Index(j)
			childNode := i.scanElement(child, depth+1)
			if childNode != nil {
				node.Children = append(node.Children, childNode)
			}
		}
	}

	return node
}

func (i *Inspector) detectComponentType(className string) string {
	// Detect component types from Tailwind classes
	patterns := map[string]string{
		"bg-white rounded-lg shadow":         "Card",
		"flex h-screen":                      "Layout",
		"px-4 py-2":                          "Button",
		"border border-gray-300 rounded-md":  "Input",
		"fixed inset-0":                      "Modal",
		"space-y-":                           "Stack",
		"grid":                               "Grid",
		"flex":                               "Flex",
		"table":                              "Table",
		"nav":                                "Navigation",
	}

	for pattern, compType := range patterns {
		if len(className) >= len(pattern) {
			for idx := 0; idx <= len(className)-len(pattern); idx++ {
				if className[idx:idx+len(pattern)] == pattern {
					return compType
				}
			}
		}
	}

	return "Element"
}

func (i *Inspector) renderTree() {
	i.treeView.Set("innerHTML", "")

	if i.root == nil {
		return
	}

	i.renderNode(i.root, 0)
}

func (i *Inspector) renderNode(node *ComponentNode, indent int) {
	document := js.Global().Get("document")

	row := document.Call("createElement", "div")
	row.Set("className", "flex items-center hover:bg-gray-800 cursor-pointer py-0.5")
	row.Get("style").Set("paddingLeft", fmt.Sprintf("%dpx", indent*12))

	// Expand/collapse arrow
	if len(node.Children) > 0 {
		arrow := document.Call("createElement", "span")
		arrow.Set("className", "text-gray-500 mr-1 w-3")
		if node.Expanded {
			arrow.Set("textContent", "▼")
		} else {
			arrow.Set("textContent", "▶")
		}
		nodeRef := node
		arrow.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			args[0].Call("stopPropagation")
			nodeRef.Expanded = !nodeRef.Expanded
			i.renderTree()
			return nil
		}))
		row.Call("appendChild", arrow)
	} else {
		spacer := document.Call("createElement", "span")
		spacer.Set("className", "w-3 mr-1")
		row.Call("appendChild", spacer)
	}

	// Component type badge
	badge := document.Call("createElement", "span")
	badgeColor := "bg-gray-600"
	switch node.Type {
	case "Card":
		badgeColor = "bg-blue-600"
	case "Button":
		badgeColor = "bg-green-600"
	case "Input":
		badgeColor = "bg-yellow-600"
	case "Modal":
		badgeColor = "bg-purple-600"
	case "Layout":
		badgeColor = "bg-pink-600"
	case "Navigation":
		badgeColor = "bg-cyan-600"
	}
	badge.Set("className", badgeColor+" text-white px-1 rounded text-xs mr-2")
	badge.Set("textContent", node.Type)
	row.Call("appendChild", badge)

	// Element name
	name := document.Call("createElement", "span")
	name.Set("className", "text-gray-300")
	name.Set("textContent", node.Name)
	row.Call("appendChild", name)

	// Click to select
	nodeRef := node
	row.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		i.selectNode(nodeRef)
		return nil
	}))

	// Highlight on hover
	row.Call("addEventListener", "mouseenter", js.FuncOf(func(this js.Value, args []js.Value) any {
		if !nodeRef.Element.IsUndefined() && !nodeRef.Element.IsNull() {
			nodeRef.Element.Get("style").Set("outline", "2px solid #a855f7")
		}
		return nil
	}))
	row.Call("addEventListener", "mouseleave", js.FuncOf(func(this js.Value, args []js.Value) any {
		if !nodeRef.Element.IsUndefined() && !nodeRef.Element.IsNull() {
			nodeRef.Element.Get("style").Set("outline", "")
		}
		return nil
	}))

	i.treeView.Call("appendChild", row)

	// Render children if expanded
	if node.Expanded {
		for _, child := range node.Children {
			i.renderNode(child, indent+1)
		}
	}
}

func (i *Inspector) selectNode(node *ComponentNode) {
	i.selectedNode = node
	i.renderProps()
}

func (i *Inspector) renderProps() {
	document := js.Global().Get("document")
	i.propsView.Set("innerHTML", "")

	if i.selectedNode == nil {
		placeholder := document.Call("createElement", "div")
		placeholder.Set("className", "text-gray-500 text-center mt-4")
		placeholder.Set("textContent", "Select an element to view props")
		i.propsView.Call("appendChild", placeholder)
		return
	}

	// Header
	header := document.Call("createElement", "div")
	header.Set("className", "text-purple-400 font-bold mb-2")
	header.Set("textContent", i.selectedNode.Name)
	i.propsView.Call("appendChild", header)

	// Type
	typeRow := document.Call("createElement", "div")
	typeRow.Set("className", "mb-2")
	typeRow.Set("innerHTML", fmt.Sprintf("<span class='text-gray-500'>type:</span> <span class='text-green-400'>%s</span>", i.selectedNode.Type))
	i.propsView.Call("appendChild", typeRow)

	// Props
	if len(i.selectedNode.Props) > 0 {
		propsHeader := document.Call("createElement", "div")
		propsHeader.Set("className", "text-gray-500 mt-2 mb-1")
		propsHeader.Set("textContent", "Props:")
		i.propsView.Call("appendChild", propsHeader)

		for key, value := range i.selectedNode.Props {
			propRow := document.Call("createElement", "div")
			propRow.Set("className", "ml-2 mb-1")

			valueStr := fmt.Sprintf("%v", value)
			if len(valueStr) > 50 {
				valueStr = valueStr[:50] + "..."
			}

			propRow.Set("innerHTML", fmt.Sprintf("<span class='text-cyan-400'>%s</span>: <span class='text-orange-300'>%q</span>", key, valueStr))
			i.propsView.Call("appendChild", propRow)
		}
	}

	// Element info
	if !i.selectedNode.Element.IsUndefined() && !i.selectedNode.Element.IsNull() {
		rect := i.selectedNode.Element.Call("getBoundingClientRect")

		dimensionsHeader := document.Call("createElement", "div")
		dimensionsHeader.Set("className", "text-gray-500 mt-3 mb-1")
		dimensionsHeader.Set("textContent", "Dimensions:")
		i.propsView.Call("appendChild", dimensionsHeader)

		dimensions := document.Call("createElement", "div")
		dimensions.Set("className", "ml-2 text-gray-300")
		dimensions.Set("innerHTML", fmt.Sprintf(
			"width: %.0fpx<br>height: %.0fpx<br>x: %.0fpx<br>y: %.0fpx",
			rect.Get("width").Float(),
			rect.Get("height").Float(),
			rect.Get("x").Float(),
			rect.Get("y").Float(),
		))
		i.propsView.Call("appendChild", dimensions)
	}
}

// Open opens the inspector panel
func (i *Inspector) Open() {
	i.panel.Get("style").Set("display", "")
	i.isOpen = true
}

// Close closes the inspector panel
func (i *Inspector) Close() {
	i.panel.Get("style").Set("display", "none")
	i.isOpen = false
}

// Toggle toggles the inspector panel
func (i *Inspector) Toggle() {
	if i.isOpen {
		i.Close()
	} else {
		i.Open()
	}
}

// IsOpen returns whether the inspector is open
func (i *Inspector) IsOpen() bool {
	return i.isOpen
}

// Element returns the container element
func (i *Inspector) Element() js.Value {
	return i.container
}

// Destroy removes the inspector from the DOM
func (i *Inspector) Destroy() {
	i.container.Call("remove")
}

// InitInspector initializes the global inspector (call once in development)
func InitInspector() *Inspector {
	if globalInspector != nil {
		return globalInspector
	}
	globalInspector = NewInspector(InspectorProps{
		Position:  "bottom-right",
		Collapsed: true,
	})
	return globalInspector
}

// GetInspector returns the global inspector
func GetInspector() *Inspector {
	return globalInspector
}
