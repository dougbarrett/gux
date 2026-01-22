package ui

import "github.com/dougbarrett/gux/core"

// ListProps configures the List component.
type ListProps struct {
	Class    string
	Children []core.Node
}

// List renders a styled unordered list with dividers between items.
// Use ListItem as children for consistent item styling.
//
// Example:
//
//	List(ListProps{
//	    Children: []core.Node{
//	        ListItem(ListItemProps{Children: []core.Node{core.Text("Item 1")}}),
//	        ListItem(ListItemProps{Children: []core.Node{core.Text("Item 2")}}),
//	        ListItem(ListItemProps{Children: []core.Node{core.Text("Item 3")}}),
//	    },
//	})
func List(props ListProps) core.Node {
	class := MergeClasses(
		"divide-y divide-gray-200 dark:divide-gray-700",
		props.Class,
	)
	return core.Ul(core.Class(class), props.Children...)
}

// ListItemProps configures the ListItem component.
type ListItemProps struct {
	Class    string
	OnClick  func() // Item click handler (WASM only)
	Children []core.Node
}

// ListItem renders a list item with padding and hover state.
//
// Example:
//
//	ListItem(ListItemProps{
//	    OnClick: func() { fmt.Println("Item clicked") },
//	    Children: []core.Node{core.Text("Clickable item")},
//	})
func ListItem(props ListItemProps) core.Node {
	class := MergeClasses(
		"px-4 py-4 hover:bg-gray-50 dark:hover:bg-gray-800",
		props.Class,
	)
	attrs := core.Attrs{
		Class:   class,
		OnClick: props.OnClick,
	}
	return core.Li(attrs, props.Children...)
}
