//go:build js && wasm

package components

import (
	"fmt"
	"syscall/js"
)

// VirtualListProps configures a VirtualList component
type VirtualListProps struct {
	Items       []any           // All items (can be any type)
	ItemHeight  int             // Fixed height of each item in pixels
	RenderItem  func(item any, index int) js.Value // Function to render each item
	Height      string          // Container height (e.g., "400px", "100%")
	Width       string          // Container width (default "100%")
	Overscan    int             // Number of extra items to render above/below visible area
	OnEndReached func()         // Called when scroll reaches end (for infinite scroll)
	EndThreshold int            // Pixels from end to trigger OnEndReached
	ClassName   string
}

// VirtualList creates a virtualized list component for efficient rendering of large lists
type VirtualList struct {
	container    js.Value
	viewport     js.Value
	content      js.Value
	items        []any
	itemHeight   int
	renderItem   func(any, int) js.Value
	startIndex   int
	endIndex     int
	overscan     int
	scrollHandler js.Func
}

// NewVirtualList creates a new VirtualList component
func NewVirtualList(props VirtualListProps) *VirtualList {
	document := js.Global().Get("document")

	if props.Overscan == 0 {
		props.Overscan = 3
	}
	if props.Height == "" {
		props.Height = "400px"
	}
	if props.Width == "" {
		props.Width = "100%"
	}
	if props.EndThreshold == 0 {
		props.EndThreshold = 200
	}

	v := &VirtualList{
		items:      props.Items,
		itemHeight: props.ItemHeight,
		renderItem: props.RenderItem,
		overscan:   props.Overscan,
	}

	// Outer container
	container := document.Call("createElement", "div")
	className := "virtual-list-container"
	if props.ClassName != "" {
		className += " " + props.ClassName
	}
	container.Set("className", className)
	container.Get("style").Set("width", props.Width)
	container.Get("style").Set("height", props.Height)
	v.container = container

	// Scrollable viewport
	viewport := document.Call("createElement", "div")
	viewport.Set("className", "virtual-list-viewport")
	viewport.Get("style").Set("height", "100%")
	viewport.Get("style").Set("overflow", "auto")
	viewport.Get("style").Set("position", "relative")
	v.viewport = viewport

	// Content wrapper (creates scrollable height)
	content := document.Call("createElement", "div")
	content.Set("className", "virtual-list-content")
	content.Get("style").Set("position", "relative")
	totalHeight := len(props.Items) * props.ItemHeight
	content.Get("style").Set("height", fmt.Sprintf("%dpx", totalHeight))
	v.content = content

	viewport.Call("appendChild", content)
	container.Call("appendChild", viewport)

	// Scroll handler
	v.scrollHandler = js.FuncOf(func(this js.Value, args []js.Value) any {
		v.render()

		// Check for end reached
		if props.OnEndReached != nil {
			scrollTop := viewport.Get("scrollTop").Int()
			clientHeight := viewport.Get("clientHeight").Int()
			scrollHeight := viewport.Get("scrollHeight").Int()

			if scrollHeight-scrollTop-clientHeight < props.EndThreshold {
				props.OnEndReached()
			}
		}

		return nil
	})
	viewport.Call("addEventListener", "scroll", v.scrollHandler)

	// Initial render
	v.render()

	return v
}

func (v *VirtualList) render() {
	if v.itemHeight == 0 || len(v.items) == 0 {
		return
	}

	scrollTop := v.viewport.Get("scrollTop").Int()
	clientHeight := v.viewport.Get("clientHeight").Int()

	// Calculate visible range
	startIndex := scrollTop / v.itemHeight
	visibleCount := (clientHeight / v.itemHeight) + 1

	// Apply overscan
	startIndex = startIndex - v.overscan
	if startIndex < 0 {
		startIndex = 0
	}

	endIndex := startIndex + visibleCount + (v.overscan * 2)
	if endIndex > len(v.items) {
		endIndex = len(v.items)
	}

	// Only re-render if range changed
	if startIndex == v.startIndex && endIndex == v.endIndex {
		return
	}

	v.startIndex = startIndex
	v.endIndex = endIndex

	// Clear existing items
	v.content.Set("innerHTML", "")

	document := js.Global().Get("document")

	// Render visible items
	for i := startIndex; i < endIndex; i++ {
		item := v.renderItem(v.items[i], i)

		// Wrap in positioned container
		wrapper := document.Call("createElement", "div")
		wrapper.Get("style").Set("position", "absolute")
		wrapper.Get("style").Set("top", fmt.Sprintf("%dpx", i*v.itemHeight))
		wrapper.Get("style").Set("left", "0")
		wrapper.Get("style").Set("right", "0")
		wrapper.Get("style").Set("height", fmt.Sprintf("%dpx", v.itemHeight))
		wrapper.Call("appendChild", item)

		v.content.Call("appendChild", wrapper)
	}
}

// Element returns the container DOM element
func (v *VirtualList) Element() js.Value {
	return v.container
}

// SetItems updates the items and re-renders
func (v *VirtualList) SetItems(items []any) {
	v.items = items

	// Update content height
	totalHeight := len(items) * v.itemHeight
	v.content.Get("style").Set("height", fmt.Sprintf("%dpx", totalHeight))

	// Force re-render
	v.startIndex = -1
	v.endIndex = -1
	v.render()
}

// AppendItems adds items to the end (for infinite scroll)
func (v *VirtualList) AppendItems(items []any) {
	v.items = append(v.items, items...)

	// Update content height
	totalHeight := len(v.items) * v.itemHeight
	v.content.Get("style").Set("height", fmt.Sprintf("%dpx", totalHeight))

	v.render()
}

// ScrollTo scrolls to a specific index
func (v *VirtualList) ScrollTo(index int) {
	scrollTop := index * v.itemHeight
	v.viewport.Set("scrollTop", scrollTop)
}

// ScrollToTop scrolls to the top
func (v *VirtualList) ScrollToTop() {
	v.viewport.Set("scrollTop", 0)
}

// ScrollToBottom scrolls to the bottom
func (v *VirtualList) ScrollToBottom() {
	v.viewport.Set("scrollTop", v.viewport.Get("scrollHeight"))
}

// GetVisibleRange returns the currently visible index range
func (v *VirtualList) GetVisibleRange() (start, end int) {
	return v.startIndex, v.endIndex
}

// ItemCount returns the total number of items
func (v *VirtualList) ItemCount() int {
	return len(v.items)
}

// Destroy cleans up event listeners
func (v *VirtualList) Destroy() {
	v.viewport.Call("removeEventListener", "scroll", v.scrollHandler)
	v.scrollHandler.Release()
}

// StringList creates a virtual list of strings
func StringList(items []string, height string, itemHeight int) *VirtualList {
	anyItems := make([]any, len(items))
	for i, item := range items {
		anyItems[i] = item
	}

	return NewVirtualList(VirtualListProps{
		Items:      anyItems,
		ItemHeight: itemHeight,
		Height:     height,
		RenderItem: func(item any, index int) js.Value {
			document := js.Global().Get("document")
			div := document.Call("createElement", "div")
			div.Set("className", "px-4 py-2 border-b border-gray-100 hover:bg-gray-50")
			div.Set("textContent", item.(string))
			return div
		},
	})
}

// CardList creates a virtual list with card-style items
func CardList(items []any, height string, itemHeight int, renderContent func(any, int) js.Value) *VirtualList {
	return NewVirtualList(VirtualListProps{
		Items:      items,
		ItemHeight: itemHeight,
		Height:     height,
		RenderItem: func(item any, index int) js.Value {
			document := js.Global().Get("document")
			card := document.Call("createElement", "div")
			card.Set("className", "mx-2 my-1 p-3 bg-white rounded shadow-sm border")
			content := renderContent(item, index)
			card.Call("appendChild", content)
			return card
		},
	})
}
