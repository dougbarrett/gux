package ui

import "github.com/dougbarrett/gux/core"

// TableProps configures the Table component.
type TableProps struct {
	Class    string
	Children []core.Node
}

// Table renders a styled HTML table wrapped in an overflow container.
// Use Thead, Tbody, Tr, Th, and Td as children for structured layout.
//
// Example:
//
//	Table(TableProps{
//	    Children: []core.Node{
//	        Thead(TheadProps{Children: []core.Node{
//	            Tr(TrProps{Children: []core.Node{
//	                Th(ThProps{Children: []core.Node{core.Text("Name")}}),
//	                Th(ThProps{Children: []core.Node{core.Text("Email")}}),
//	            }}),
//	        }}),
//	        Tbody(TbodyProps{Children: []core.Node{
//	            Tr(TrProps{Children: []core.Node{
//	                Td(TdProps{Children: []core.Node{core.Text("John")}}),
//	                Td(TdProps{Children: []core.Node{core.Text("john@example.com")}}),
//	            }}),
//	        }}),
//	    },
//	})
func Table(props TableProps) core.Node {
	tableClass := MergeClasses(
		"min-w-full divide-y divide-gray-200 dark:divide-gray-700",
		props.Class,
	)
	// Wrap in overflow-x-auto div to prevent wide tables breaking layout
	return core.Div(core.Class("overflow-x-auto"),
		core.Table(core.Class(tableClass), props.Children...),
	)
}

// TheadProps configures the Thead component.
type TheadProps struct {
	Class    string
	Children []core.Node
}

// Thead renders the table header section with background styling.
//
// Example:
//
//	Thead(TheadProps{
//	    Children: []core.Node{
//	        Tr(TrProps{Children: []core.Node{
//	            Th(ThProps{Children: []core.Node{core.Text("Column")}}),
//	        }}),
//	    },
//	})
func Thead(props TheadProps) core.Node {
	class := MergeClasses(
		"bg-gray-50 dark:bg-gray-800",
		props.Class,
	)
	return core.Thead(core.Class(class), props.Children...)
}

// TbodyProps configures the Tbody component.
type TbodyProps struct {
	Class    string
	Children []core.Node
}

// Tbody renders the table body section with row dividers.
//
// Example:
//
//	Tbody(TbodyProps{
//	    Children: []core.Node{
//	        Tr(TrProps{Children: []core.Node{
//	            Td(TdProps{Children: []core.Node{core.Text("Cell")}}),
//	        }}),
//	    },
//	})
func Tbody(props TbodyProps) core.Node {
	class := MergeClasses(
		"bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700",
		props.Class,
	)
	return core.Tbody(core.Class(class), props.Children...)
}

// TrProps configures the Tr (table row) component.
type TrProps struct {
	Class    string
	OnClick  func() // Row click handler (WASM only)
	Children []core.Node
}

// Tr renders a table row with optional click handler.
//
// Example:
//
//	Tr(TrProps{
//	    OnClick: func() { fmt.Println("Row clicked") },
//	    Children: []core.Node{
//	        Td(TdProps{Children: []core.Node{core.Text("Cell")}}),
//	    },
//	})
func Tr(props TrProps) core.Node {
	attrs := core.Attrs{
		Class:   MergeClasses(props.Class),
		OnClick: props.OnClick,
	}
	return core.Tr(attrs, props.Children...)
}

// ThProps configures the Th (table header cell) component.
type ThProps struct {
	Class    string
	Children []core.Node
}

// Th renders a table header cell with uppercase styling.
//
// Example:
//
//	Th(ThProps{Children: []core.Node{core.Text("Name")}})
func Th(props ThProps) core.Node {
	class := MergeClasses(
		"px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider",
		props.Class,
	)
	return core.Th(core.Class(class), props.Children...)
}

// TdProps configures the Td (table data cell) component.
type TdProps struct {
	Class    string
	Children []core.Node
}

// Td renders a table data cell with standard styling.
//
// Example:
//
//	Td(TdProps{Children: []core.Node{core.Text("Cell content")}})
func Td(props TdProps) core.Node {
	class := MergeClasses(
		"px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100",
		props.Class,
	)
	return core.Td(core.Class(class), props.Children...)
}
