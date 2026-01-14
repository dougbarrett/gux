//go:build js && wasm

package components

import "syscall/js"

// TableColumn defines a table column
type TableColumn struct {
	Header    string
	Key       string
	Width     string
	ClassName string
	Render    func(row map[string]any, value any) js.Value // Custom cell renderer
}

// TableProps configures a Table component
type TableProps struct {
	Columns   []TableColumn
	Data      []map[string]any
	Striped   bool
	Hoverable bool
	Bordered  bool
	Compact   bool
	OnRowClick func(row map[string]any, index int)
}

// Table creates a data table component
type Table struct {
	container js.Value
	tbody     js.Value
	columns   []TableColumn
	props     TableProps
}

// NewTable creates a new Table component
func NewTable(props TableProps) *Table {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "overflow-x-auto")

	table := document.Call("createElement", "table")
	tableClass := "min-w-full divide-y divide-gray-200 dark:divide-gray-700"
	if props.Bordered {
		tableClass += " border border-gray-200 dark:border-gray-700"
	}
	table.Set("className", tableClass)

	// Header
	thead := document.Call("createElement", "thead")
	thead.Set("className", "bg-gray-50 dark:bg-gray-800")

	headerRow := document.Call("createElement", "tr")
	for _, col := range props.Columns {
		th := document.Call("createElement", "th")
		thClass := "px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
		if props.Compact {
			thClass = "px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
		}
		if props.Bordered {
			thClass += " border-b border-gray-200 dark:border-gray-700"
		}
		if col.Width != "" {
			th.Get("style").Set("width", col.Width)
		}
		th.Set("className", thClass)
		th.Set("textContent", col.Header)
		headerRow.Call("appendChild", th)
	}
	thead.Call("appendChild", headerRow)
	table.Call("appendChild", thead)

	// Body
	tbody := document.Call("createElement", "tbody")
	tbodyClass := "bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700"
	tbody.Set("className", tbodyClass)
	table.Call("appendChild", tbody)

	container.Call("appendChild", table)

	t := &Table{
		container: container,
		tbody:     tbody,
		columns:   props.Columns,
		props:     props,
	}

	// Render initial data
	t.SetData(props.Data)

	return t
}

// Element returns the container DOM element
func (t *Table) Element() js.Value {
	return t.container
}

// SetData updates the table data
func (t *Table) SetData(data []map[string]any) {
	document := js.Global().Get("document")

	t.tbody.Set("innerHTML", "")

	for i, row := range data {
		tr := document.Call("createElement", "tr")

		rowClass := ""
		if t.props.Striped && i%2 == 1 {
			rowClass = "bg-gray-50 dark:bg-gray-800"
		}
		if t.props.Hoverable {
			rowClass += " hover:bg-gray-100 dark:hover:bg-gray-800"
		}
		if t.props.OnRowClick != nil {
			rowClass += " cursor-pointer"
			idx := i
			rowData := row
			tr.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
				t.props.OnRowClick(rowData, idx)
				return nil
			}))
		}
		tr.Set("className", rowClass)

		for _, col := range t.columns {
			td := document.Call("createElement", "td")
			tdClass := "px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100"
			if t.props.Compact {
				tdClass = "px-4 py-2 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100"
			}
			if t.props.Bordered {
				tdClass += " border-b border-gray-200 dark:border-gray-700"
			}
			if col.ClassName != "" {
				tdClass = col.ClassName
			}
			td.Set("className", tdClass)

			value := row[col.Key]

			if col.Render != nil {
				// Custom renderer
				rendered := col.Render(row, value)
				td.Call("appendChild", rendered)
			} else {
				// Default: show as text
				if value != nil {
					td.Set("textContent", toString(value))
				}
			}

			tr.Call("appendChild", td)
		}

		t.tbody.Call("appendChild", tr)
	}
}

// Helper to convert any to string
func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return js.Global().Get("String").Invoke(val).String()
	case float64:
		return js.Global().Get("String").Invoke(val).String()
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}
