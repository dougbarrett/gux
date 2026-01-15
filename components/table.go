//go:build js && wasm

package components

import (
	"sort"
	"strings"
	"syscall/js"
)

// TableColumn defines a table column
type TableColumn struct {
	Header    string
	Key       string
	Width     string
	ClassName string
	Sortable  bool                                          // Whether this column is sortable
	SortKey   string                                        // Key to sort by (defaults to Key if not set)
	Render    func(row map[string]any, value any) js.Value // Custom cell renderer
}

// TableProps configures a Table component
type TableProps struct {
	Columns           []TableColumn
	Data              []map[string]any
	Striped           bool
	Hoverable         bool
	Bordered          bool
	Compact           bool
	OnRowClick        func(row map[string]any, index int)
	OnSort            func(column string, direction string) // Callback when sort changes
	Filterable        bool                                  // Enable filter input
	FilterPlaceholder string                                // Placeholder text (default "Search...")
	FilterColumns     []string                              // Columns to filter (nil = all)
	OnFilter          func(text string)                     // Callback when filter changes
}

// Table creates a data table component
type Table struct {
	container     js.Value
	tbody         js.Value
	thead         js.Value
	columns       []TableColumn
	props         TableProps
	data          []map[string]any
	allData       []map[string]any // Unfiltered data
	sortColumn    string
	sortDirection string // "asc", "desc", or "" (none)
	filterText    string
	filterInput   js.Value
	debounceTimer js.Value // For debouncing filter input
}

// NewTable creates a new Table component
func NewTable(props TableProps) *Table {
	document := js.Global().Get("document")

	// Outer container - wraps everything (filter input + table)
	container := document.Call("createElement", "div")
	container.Set("className", "w-full")

	// Table wrapper - handles overflow
	tableWrapper := document.Call("createElement", "div")
	tableWrapper.Set("className", "overflow-x-auto")

	table := document.Call("createElement", "table")
	tableClass := "min-w-full divide-y divide-gray-200 dark:divide-gray-700"
	if props.Bordered {
		tableClass += " border border-gray-200 dark:border-gray-700"
	}
	table.Set("className", tableClass)

	// Header
	thead := document.Call("createElement", "thead")
	thead.Set("className", "bg-gray-50 dark:bg-gray-800")
	table.Call("appendChild", thead)

	// Body
	tbody := document.Call("createElement", "tbody")
	tbodyClass := "bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700"
	tbody.Set("className", tbodyClass)
	table.Call("appendChild", tbody)

	tableWrapper.Call("appendChild", table)

	t := &Table{
		container: container,
		tbody:     tbody,
		thead:     thead,
		columns:   props.Columns,
		props:     props,
	}

	// Add filter input if Filterable
	if props.Filterable {
		filterContainer := t.createFilterInput(document)
		container.Call("appendChild", filterContainer)
	}

	container.Call("appendChild", tableWrapper)

	// Render headers (with sort indicators)
	t.renderHeaders()

	// Render initial data
	t.SetData(props.Data)

	return t
}

// createFilterInput creates the filter input with search icon
func (t *Table) createFilterInput(document js.Value) js.Value {
	filterContainer := document.Call("createElement", "div")
	filterContainer.Set("className", "relative mb-4")

	// Search icon
	iconSpan := document.Call("createElement", "span")
	iconSpan.Set("className", "absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none")
	iconSpan.Set("textContent", "🔍")
	filterContainer.Call("appendChild", iconSpan)

	// Input field
	input := document.Call("createElement", "input")
	input.Set("type", "text")
	placeholder := t.props.FilterPlaceholder
	if placeholder == "" {
		placeholder = "Search..."
	}
	input.Set("placeholder", placeholder)
	input.Set("className", "w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 placeholder-gray-400 dark:placeholder-gray-500")

	// Debounced input handler
	input.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		value := input.Get("value").String()

		// Clear previous timer
		if !t.debounceTimer.IsUndefined() && !t.debounceTimer.IsNull() {
			js.Global().Call("clearTimeout", t.debounceTimer)
		}

		// Set new debounced timer (150ms)
		t.debounceTimer = js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
			t.filterText = value
			if t.props.OnFilter != nil {
				t.props.OnFilter(value)
			}
			// Re-render with filter applied
			t.renderData()
			return nil
		}), 150)

		return nil
	}))

	filterContainer.Call("appendChild", input)
	t.filterInput = input

	return filterContainer
}

// renderHeaders creates or updates the table header row with sort indicators
func (t *Table) renderHeaders() {
	document := js.Global().Get("document")
	t.thead.Set("innerHTML", "")

	headerRow := document.Call("createElement", "tr")
	for _, col := range t.columns {
		th := document.Call("createElement", "th")
		thClass := "px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
		if t.props.Compact {
			thClass = "px-4 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
		}
		if t.props.Bordered {
			thClass += " border-b border-gray-200 dark:border-gray-700"
		}
		if col.Sortable {
			thClass += " cursor-pointer select-none hover:bg-gray-100 dark:hover:bg-gray-700"
		}
		if col.Width != "" {
			th.Get("style").Set("width", col.Width)
		}
		th.Set("className", thClass)

		// Header text with sort indicator
		headerText := col.Header
		if col.Sortable {
			sortKey := col.SortKey
			if sortKey == "" {
				sortKey = col.Key
			}

			// Determine sort indicator
			indicator := " ⇅" // neutral/unsorted
			if t.sortColumn == sortKey {
				if t.sortDirection == "asc" {
					indicator = " ▲"
				} else if t.sortDirection == "desc" {
					indicator = " ▼"
				}
			}
			headerText += indicator

			// Add click handler
			colSortKey := sortKey // capture for closure
			th.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
				t.handleHeaderClick(colSortKey)
				return nil
			}))
		}
		th.Set("textContent", headerText)

		headerRow.Call("appendChild", th)
	}
	t.thead.Call("appendChild", headerRow)
}

// handleHeaderClick cycles sort direction: none -> asc -> desc -> none
func (t *Table) handleHeaderClick(sortKey string) {
	if t.sortColumn == sortKey {
		// Cycle: asc -> desc -> none
		switch t.sortDirection {
		case "asc":
			t.sortDirection = "desc"
		case "desc":
			t.sortColumn = ""
			t.sortDirection = ""
		default:
			t.sortDirection = "asc"
		}
	} else {
		// New column, start with asc
		t.sortColumn = sortKey
		t.sortDirection = "asc"
	}

	// Re-render headers to update indicators
	t.renderHeaders()

	// Call OnSort callback if provided
	if t.props.OnSort != nil {
		t.props.OnSort(t.sortColumn, t.sortDirection)
	}

	// Re-render data with new sort
	if len(t.allData) > 0 {
		t.renderData()
	}
}

// Element returns the container DOM element
func (t *Table) Element() js.Value {
	return t.container
}

// SetData updates the table data
func (t *Table) SetData(data []map[string]any) {
	// Store unfiltered data
	t.allData = data
	t.data = data

	// Render with current filter/sort state
	t.renderData()
}

// filterData returns rows that match the current filter text
func (t *Table) filterData(data []map[string]any) []map[string]any {
	if t.filterText == "" || len(data) == 0 {
		return data
	}

	filterLower := strings.ToLower(t.filterText)
	var filtered []map[string]any

	for _, row := range data {
		// Check specified columns or all columns
		columnsToCheck := t.props.FilterColumns
		if len(columnsToCheck) == 0 {
			// Check all columns
			for _, col := range t.columns {
				columnsToCheck = append(columnsToCheck, col.Key)
			}
		}

		// Include row if ANY column contains the filter text
		for _, colKey := range columnsToCheck {
			value := row[colKey]
			if value != nil {
				strValue := strings.ToLower(toString(value))
				if strings.Contains(strValue, filterLower) {
					filtered = append(filtered, row)
					break
				}
			}
		}
	}

	return filtered
}

// renderData applies filter and sort, then renders
func (t *Table) renderData() {
	document := js.Global().Get("document")

	// Apply filter first, then sort
	displayData := t.filterData(t.allData)
	displayData = t.sortData(displayData)

	t.tbody.Set("innerHTML", "")

	for i, row := range displayData {
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

// sortData returns a sorted copy of the data based on current sort state
func (t *Table) sortData(data []map[string]any) []map[string]any {
	if t.sortColumn == "" || t.sortDirection == "" || len(data) == 0 {
		return data
	}

	// Create a copy to avoid mutating original
	sorted := make([]map[string]any, len(data))
	copy(sorted, data)

	sort.SliceStable(sorted, func(i, j int) bool {
		valI := sorted[i][t.sortColumn]
		valJ := sorted[j][t.sortColumn]

		// Handle nil values - sort to end
		if valI == nil && valJ == nil {
			return false
		}
		if valI == nil {
			return false // nil goes to end
		}
		if valJ == nil {
			return true // non-nil before nil
		}

		result := compareValues(valI, valJ)

		// Reverse for descending
		if t.sortDirection == "desc" {
			return result > 0
		}
		return result < 0
	})

	return sorted
}

// compareValues compares two values for sorting
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func compareValues(a, b any) int {
	// Try string comparison (case-insensitive)
	if strA, okA := a.(string); okA {
		if strB, okB := b.(string); okB {
			return strings.Compare(strings.ToLower(strA), strings.ToLower(strB))
		}
	}

	// Try numeric comparison
	numA := toFloat64(a)
	numB := toFloat64(b)
	if numA != nil && numB != nil {
		if *numA < *numB {
			return -1
		}
		if *numA > *numB {
			return 1
		}
		return 0
	}

	// Try bool comparison (false < true)
	if boolA, okA := a.(bool); okA {
		if boolB, okB := b.(bool); okB {
			if boolA == boolB {
				return 0
			}
			if !boolA && boolB {
				return -1
			}
			return 1
		}
	}

	// Fallback: convert to string
	return strings.Compare(toString(a), toString(b))
}

// toFloat64 attempts to convert a value to float64
func toFloat64(v any) *float64 {
	switch val := v.(type) {
	case int:
		f := float64(val)
		return &f
	case int64:
		f := float64(val)
		return &f
	case float64:
		return &val
	case float32:
		f := float64(val)
		return &f
	}
	return nil
}

// SetSort programmatically sets the sort column and direction
func (t *Table) SetSort(column, direction string) {
	t.sortColumn = column
	t.sortDirection = direction
	t.renderHeaders()
	if len(t.data) > 0 {
		t.SetData(t.data)
	}
}

// Sort toggles sorting on the specified column
func (t *Table) Sort(column string) {
	t.handleHeaderClick(column)
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
