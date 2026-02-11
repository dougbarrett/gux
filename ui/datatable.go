package ui

import (
	"strconv"

	"github.com/dougbarrett/gux/core"
)

// ColumnDef defines a column in a DataTable.
type ColumnDef[T any] struct {
	Header string           // Column header text
	Render func(T) core.Node // Cell renderer
	Width  string           // Optional width class (e.g., "w-48")
}

// DataTableProps configures the DataTable component.
type DataTableProps[T any] struct {
	Data       []T            // Data to display
	Columns    []ColumnDef[T] // Column definitions
	RowKey     func(T) string // Key generator for rows (optional)
	OnRowClick func(T)        // Row click handler (optional, WASM only)
	Striped    bool           // Alternate row background
	Hoverable  bool           // Row hover effect
	Class      string         // Additional classes

	// Alpine.js props
	XOnRowClick string // x-on:click expression for row clicks (receives row index via $el.dataset.idx)
}

// DataTable renders a typed data table using generic column definitions.
// Uses Table, Thead, Tbody, Tr, Th, Td internally.
//
// Example:
//
//	type User struct { Name string; Email string }
//	DataTable(DataTableProps[User]{
//	    Data: users,
//	    Columns: []ColumnDef[User]{
//	        {Header: "Name", Render: func(u User) core.Node { return core.Text(u.Name) }},
//	        {Header: "Email", Render: func(u User) core.Node { return core.Text(u.Email) }},
//	    },
//	})
func DataTable[T any](props DataTableProps[T]) core.Node {
	// Build headers
	var headers []core.Node
	for _, col := range props.Columns {
		headers = append(headers, Th(ThProps{
			Class:    col.Width,
			Children: []core.Node{core.Text(col.Header)},
		}))
	}

	// Build rows
	var rows []core.Node
	for i, item := range props.Data {
		var cells []core.Node
		for _, col := range props.Columns {
			cells = append(cells, Td(TdProps{
				Children: []core.Node{col.Render(item)},
			}))
		}

		// Capture for closure to avoid closure capture bug
		capturedItem := item

		// Row styling
		rowClass := ""
		if props.Striped && i%2 == 1 {
			rowClass = "bg-gray-50 dark:bg-gray-800"
		}
		if props.Hoverable {
			rowClass = MergeClasses(rowClass, "hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer")
		}

		// Row click handler
		var onClick func()
		if props.OnRowClick != nil {
			onClick = func() {
				props.OnRowClick(capturedItem)
			}
		}

		// Alpine.js row click
		if props.XOnRowClick != "" {
			rowAttrs := core.Attrs{
				Class: MergeClasses(rowClass),
				Data:  map[string]string{"idx": strconv.Itoa(i)},
				XOn:   map[string]string{"click": props.XOnRowClick},
			}
			rows = append(rows, core.Tr(rowAttrs, cells...))
		} else {
			rows = append(rows, Tr(TrProps{
				Class:    rowClass,
				OnClick:  onClick,
				Children: cells,
			}))
		}
	}

	return Table(TableProps{
		Class: props.Class,
		Children: []core.Node{
			Thead(TheadProps{
				Children: []core.Node{
					Tr(TrProps{Children: headers}),
				},
			}),
			Tbody(TbodyProps{Children: rows}),
		},
	})
}
