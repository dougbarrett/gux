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

// BulkAction defines an action that can be performed on selected rows
type BulkAction struct {
	Label     string                   // Button text
	Icon      string                   // Optional emoji icon
	Variant   string                   // Button style: primary, danger, secondary (default)
	OnExecute func(selectedKeys []any) // Action handler
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
	Paginated         bool                                  // Enable pagination
	PageSize          int                                   // Items per page (default 10)
	ShowPageInfo      bool                                  // Show "Showing X-Y of Z" (default true)
	OnPageChange      func(page int)                        // Callback when page changes
	Selectable        bool                                  // Enable row selection with checkboxes
	RowKey            string                                // Unique identifier field in data (default "id")
	OnSelectionChange func(selectedKeys []any)              // Callback when selection changes
	BulkActions       []BulkAction                          // Available bulk actions for selected rows
	Exportable        bool                                  // Enable export dropdown
	ExportFilename    string                                // Base filename for exports (default "export")
	ExportColumns     []string                              // Columns to export (nil = all column keys)
	EmptyState        *EmptyState                           // Custom empty state (optional)
	EmptyTitle        string                                // Title for default empty state (optional)
	EmptyDescription  string                                // Description for default empty state (optional)
}

// Table creates a data table component
type Table struct {
	container       js.Value
	tbody           js.Value
	thead           js.Value
	columns         []TableColumn
	props           TableProps
	data            []map[string]any
	allData         []map[string]any // Unfiltered data
	sortColumn      string
	sortDirection   string // "asc", "desc", or "" (none)
	filterText      string
	filterInput     js.Value
	debounceTimer   js.Value    // For debouncing filter input
	currentPage     int         // Current page (1-indexed)
	pagination      *Pagination // Pagination component instance
	paginationMount js.Value    // Container where pagination is mounted
	selectedKeys    map[any]bool // Set of selected row keys
	rowCheckboxes   []js.Value   // References to row checkboxes for updates
	selectAllCb     js.Value     // Reference to select-all checkbox
	bulkActionBar   js.Value     // Container for bulk action bar
	bulkActionCount js.Value     // Element showing selected count
	exportDropdown  *Dropdown    // Export dropdown component
	emptyStateEl    js.Value     // Container for empty state display
	tableWrapper    js.Value     // Table wrapper element (to show/hide)
}

// NewTable creates a new Table component
func NewTable(props TableProps) *Table {
	document := js.Global().Get("document")

	// Set default PageSize if not specified
	if props.PageSize == 0 {
		props.PageSize = 10
	}

	// Set default RowKey if not specified
	if props.RowKey == "" {
		props.RowKey = "id"
	}

	// Outer container - wraps everything (filter input + table + pagination)
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

	// Empty state container (hidden by default)
	emptyStateEl := document.Call("createElement", "div")
	emptyStateEl.Set("className", "hidden")

	t := &Table{
		container:    container,
		tbody:        tbody,
		thead:        thead,
		columns:      props.Columns,
		props:        props,
		currentPage:  1,
		selectedKeys: make(map[any]bool),
		tableWrapper: tableWrapper,
		emptyStateEl: emptyStateEl,
	}

	// Add toolbar if Filterable or Exportable
	if props.Filterable || props.Exportable {
		toolbar := t.createToolbar(document)
		container.Call("appendChild", toolbar)
	}

	// Add bulk action bar if selectable with bulk actions
	if props.Selectable && len(props.BulkActions) > 0 {
		bulkBar := t.createBulkActionBar(document)
		container.Call("appendChild", bulkBar)
	}

	container.Call("appendChild", tableWrapper)
	container.Call("appendChild", emptyStateEl)

	// Add pagination container if Paginated
	if props.Paginated {
		paginationMount := document.Call("createElement", "div")
		paginationMount.Set("className", "mt-4")
		container.Call("appendChild", paginationMount)
		t.paginationMount = paginationMount
	}

	// Render headers (with sort indicators)
	t.renderHeaders()

	// Render initial data
	t.SetData(props.Data)

	return t
}

// createToolbar creates the toolbar containing filter and export dropdown
func (t *Table) createToolbar(document js.Value) js.Value {
	toolbar := document.Call("createElement", "div")
	toolbar.Set("className", "flex items-center gap-4 mb-4")

	// Add filter input if Filterable
	if t.props.Filterable {
		filterContainer := t.createFilterInput(document)
		toolbar.Call("appendChild", filterContainer)
	}

	// Add export dropdown if Exportable
	if t.props.Exportable {
		exportDropdown := t.createExportDropdown()
		toolbar.Call("appendChild", exportDropdown.Element())
		t.exportDropdown = exportDropdown
	}

	return toolbar
}

// createExportDropdown creates the export dropdown with CSV/JSON/PDF options
func (t *Table) createExportDropdown() *Dropdown {
	return NewDropdown(DropdownProps{
		Trigger: Button(ButtonProps{
			Text:    "📥 Export ▼",
			Variant: ButtonSecondary,
			Size:    ButtonSM,
		}),
		Items: []DropdownItem{
			{
				Label: "CSV",
				Icon:  "📄",
				OnClick: func() {
					t.exportData("csv")
				},
			},
			{
				Label: "JSON",
				Icon:  "📋",
				OnClick: func() {
					t.exportData("json")
				},
			},
			{
				Label: "PDF",
				Icon:  "📑",
				OnClick: func() {
					t.exportData("pdf")
				},
			},
		},
		Align: "right",
	})
}

// exportData exports table data in the specified format
func (t *Table) exportData(format string) {
	// Determine which data to export
	var dataToExport []map[string]any

	// If selectable and has selection, export selected rows only
	if t.props.Selectable && len(t.selectedKeys) > 0 {
		dataToExport = t.SelectedRows()
	} else {
		// Export all filtered data (respects current filter)
		dataToExport = t.filterData(t.allData)
		dataToExport = t.sortData(dataToExport)
	}

	if len(dataToExport) == 0 {
		return
	}

	// Determine filename
	filename := t.props.ExportFilename
	if filename == "" {
		filename = "export"
	}

	// Determine columns to export
	columns := t.props.ExportColumns
	if len(columns) == 0 {
		// Use all column keys from table definition
		columns = make([]string, len(t.columns))
		for i, col := range t.columns {
			columns[i] = col.Key
		}
	}

	// Export based on format
	switch format {
	case "csv":
		ExportCSV(dataToExport, columns, filename)
	case "json":
		ExportJSON(dataToExport, filename)
	case "pdf":
		// Extract headers from columns
		headers := make([]string, len(t.columns))
		for i, col := range t.columns {
			headers[i] = col.Header
		}
		ExportPDF(dataToExport, headers, columns, filename, PDFExportOptions{})
	}
}

// createFilterInput creates the filter input with search icon
func (t *Table) createFilterInput(document js.Value) js.Value {
	filterContainer := document.Call("createElement", "div")
	filterContainer.Set("className", "relative flex-1")

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
			// Reset to page 1 when filter changes
			t.currentPage = 1
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

// createBulkActionBar creates the bulk action bar that appears when rows are selected
func (t *Table) createBulkActionBar(document js.Value) js.Value {
	// Container - hidden by default
	bar := document.Call("createElement", "div")
	bar.Set("className", "mb-4 px-4 py-3 bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-800 rounded-lg flex items-center gap-4 hidden")
	t.bulkActionBar = bar

	// Selected count text
	countText := document.Call("createElement", "span")
	countText.Set("className", "text-sm font-medium text-blue-700 dark:text-blue-300")
	countText.Set("textContent", "0 items selected")
	t.bulkActionCount = countText
	bar.Call("appendChild", countText)

	// Action buttons container
	buttonsContainer := document.Call("createElement", "div")
	buttonsContainer.Set("className", "flex items-center gap-2")

	for _, action := range t.props.BulkActions {
		btn := document.Call("createElement", "button")

		// Determine button style based on variant
		btnClass := "px-3 py-1.5 text-sm font-medium rounded-md transition-colors "
		switch action.Variant {
		case "primary":
			btnClass += "bg-blue-600 hover:bg-blue-700 text-white"
		case "danger":
			btnClass += "bg-red-600 hover:bg-red-700 text-white"
		default: // secondary
			btnClass += "bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600"
		}
		btn.Set("className", btnClass)

		// Button content
		buttonText := action.Label
		if action.Icon != "" {
			buttonText = action.Icon + " " + action.Label
		}
		btn.Set("textContent", buttonText)

		// Click handler
		capturedAction := action
		btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			if capturedAction.OnExecute != nil {
				capturedAction.OnExecute(t.SelectedKeys())
			}
			return nil
		}))

		buttonsContainer.Call("appendChild", btn)
	}

	bar.Call("appendChild", buttonsContainer)

	// Clear selection link
	clearLink := document.Call("createElement", "button")
	clearLink.Set("className", "ml-auto text-sm text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-200 hover:underline")
	clearLink.Set("textContent", "Clear selection")
	clearLink.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		t.ClearSelection()
		return nil
	}))
	bar.Call("appendChild", clearLink)

	return bar
}

// updateBulkActionBar shows/hides the bulk action bar and updates the count
func (t *Table) updateBulkActionBar() {
	if t.bulkActionBar.IsUndefined() || t.bulkActionBar.IsNull() {
		return
	}

	count := len(t.selectedKeys)
	if count > 0 {
		// Show bar and update count
		currentClass := t.bulkActionBar.Get("className").String()
		newClass := strings.Replace(currentClass, " hidden", "", 1)
		t.bulkActionBar.Set("className", newClass)

		// Update count text
		itemText := "items"
		if count == 1 {
			itemText = "item"
		}
		t.bulkActionCount.Set("textContent", toString(count)+" "+itemText+" selected")
	} else {
		// Hide bar
		currentClass := t.bulkActionBar.Get("className").String()
		if !strings.Contains(currentClass, "hidden") {
			t.bulkActionBar.Set("className", currentClass+" hidden")
		}
	}
}

// renderHeaders creates or updates the table header row with sort indicators
func (t *Table) renderHeaders() {
	document := js.Global().Get("document")
	t.thead.Set("innerHTML", "")

	headerRow := document.Call("createElement", "tr")

	// Add checkbox column header if selectable
	if t.props.Selectable {
		th := document.Call("createElement", "th")
		thClass := "px-4 py-3 w-10"
		if t.props.Compact {
			thClass = "px-2 py-2 w-10"
		}
		if t.props.Bordered {
			thClass += " border-b border-gray-200 dark:border-gray-700"
		}
		th.Set("className", thClass)

		// Create select-all checkbox
		checkbox := document.Call("createElement", "input")
		checkbox.Set("type", "checkbox")
		checkbox.Set("className", "h-4 w-4 text-blue-600 border-gray-300 dark:border-gray-600 rounded focus:ring-blue-500 dark:bg-gray-700 cursor-pointer")

		// Set initial state based on current selection
		t.updateSelectAllState(checkbox)

		// Add click handler for select-all
		checkbox.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) any {
			t.handleSelectAll(checkbox.Get("checked").Bool())
			return nil
		}))

		th.Call("appendChild", checkbox)
		t.selectAllCb = checkbox
		headerRow.Call("appendChild", th)
	}

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

	// Clear selection when data changes
	t.selectedKeys = make(map[any]bool)

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

// paginateData returns the slice of data for the current page
func (t *Table) paginateData(data []map[string]any) []map[string]any {
	if !t.props.Paginated || len(data) == 0 {
		return data
	}

	start := (t.currentPage - 1) * t.props.PageSize
	end := start + t.props.PageSize

	if start >= len(data) {
		// Current page exceeds data, return empty (shouldn't happen if page is reset properly)
		return nil
	}
	if end > len(data) {
		end = len(data)
	}

	return data[start:end]
}

// updatePagination creates or updates the pagination component based on current data
func (t *Table) updatePagination(totalItems int) {
	if !t.props.Paginated {
		return
	}

	// Calculate total pages
	totalPages := (totalItems + t.props.PageSize - 1) / t.props.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	// Reset to page 1 if current page exceeds total
	if t.currentPage > totalPages {
		t.currentPage = 1
	}

	// Show page info - default is true if TotalItems > 0
	showInfo := totalItems > 0

	// Create new pagination component
	t.pagination = NewPagination(PaginationProps{
		CurrentPage:  t.currentPage,
		TotalPages:   totalPages,
		TotalItems:   totalItems,
		ItemsPerPage: t.props.PageSize,
		ShowInfo:     showInfo,
		OnPageChange: func(page int) {
			t.currentPage = page
			if t.props.OnPageChange != nil {
				t.props.OnPageChange(page)
			}
			t.renderData()
		},
	})

	// Mount pagination
	if !t.paginationMount.IsUndefined() && !t.paginationMount.IsNull() {
		t.paginationMount.Set("innerHTML", "")
		t.paginationMount.Call("appendChild", t.pagination.Element())
	}
}

// renderData applies filter, sort, and paginate, then renders
func (t *Table) renderData() {
	document := js.Global().Get("document")

	// Apply filter first, then sort
	displayData := t.filterData(t.allData)
	displayData = t.sortData(displayData)

	// Check for empty state conditions
	filteredCount := len(displayData)
	hasData := len(t.allData) > 0
	hasFilteredData := filteredCount > 0

	// Handle empty states
	if !hasFilteredData {
		t.showEmptyState(!hasData)
		return
	}

	// Hide empty state and show table
	t.hideEmptyState()

	// Update pagination based on filtered/sorted data count
	t.updatePagination(filteredCount)

	// Apply pagination to get current page slice
	displayData = t.paginateData(displayData)

	t.tbody.Set("innerHTML", "")

	// Reset row checkboxes array
	t.rowCheckboxes = nil

	for i, row := range displayData {
		tr := document.Call("createElement", "tr")
		rowKey := t.getRowKey(row)
		isSelected := t.selectedKeys[rowKey]

		rowClass := ""
		if isSelected {
			// Selected row highlight
			rowClass = "bg-blue-50 dark:bg-blue-900/30"
		} else if t.props.Striped && i%2 == 1 {
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

		// Add checkbox cell if selectable
		if t.props.Selectable {
			td := document.Call("createElement", "td")
			tdClass := "px-4 py-4 w-10"
			if t.props.Compact {
				tdClass = "px-2 py-2 w-10"
			}
			if t.props.Bordered {
				tdClass += " border-b border-gray-200 dark:border-gray-700"
			}
			td.Set("className", tdClass)

			checkbox := document.Call("createElement", "input")
			checkbox.Set("type", "checkbox")
			checkbox.Set("className", "h-4 w-4 text-blue-600 border-gray-300 dark:border-gray-600 rounded focus:ring-blue-500 dark:bg-gray-700 cursor-pointer")
			checkbox.Set("checked", isSelected)

			// Capture key for closure
			capturedKey := rowKey
			checkbox.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) any {
				checked := checkbox.Get("checked").Bool()
				t.handleRowSelection(capturedKey, checked)
				// Re-render to update row styling
				t.renderData()
				return nil
			}))

			// Stop click propagation so row click doesn't fire
			checkbox.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
				args[0].Call("stopPropagation")
				return nil
			}))

			td.Call("appendChild", checkbox)
			tr.Call("appendChild", td)
			t.rowCheckboxes = append(t.rowCheckboxes, checkbox)
		}

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

	// Update select-all checkbox state
	t.updateSelectAllState(t.selectAllCb)
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
	if len(t.allData) > 0 {
		t.renderData()
	}
}

// Sort toggles sorting on the specified column
func (t *Table) Sort(column string) {
	t.handleHeaderClick(column)
}

// SetFilter programmatically sets the filter text
func (t *Table) SetFilter(text string) {
	t.filterText = text
	// Reset to page 1 when filter changes
	t.currentPage = 1

	// Update input field if it exists
	if !t.filterInput.IsUndefined() && !t.filterInput.IsNull() {
		t.filterInput.Set("value", text)
	}

	// Notify callback
	if t.props.OnFilter != nil {
		t.props.OnFilter(text)
	}

	// Re-render
	if len(t.allData) > 0 {
		t.renderData()
	}
}

// ClearFilter resets the filter to show all rows
func (t *Table) ClearFilter() {
	t.SetFilter("")
}

// SetPage navigates to a specific page
func (t *Table) SetPage(page int) {
	if page < 1 {
		page = 1
	}
	totalPages := t.TotalPages()
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}
	t.currentPage = page
	if t.props.OnPageChange != nil {
		t.props.OnPageChange(page)
	}
	t.renderData()
}

// NextPage navigates to the next page
func (t *Table) NextPage() {
	if t.currentPage < t.TotalPages() {
		t.SetPage(t.currentPage + 1)
	}
}

// PrevPage navigates to the previous page
func (t *Table) PrevPage() {
	if t.currentPage > 1 {
		t.SetPage(t.currentPage - 1)
	}
}

// TotalPages returns the total number of pages
func (t *Table) TotalPages() int {
	if !t.props.Paginated || len(t.allData) == 0 {
		return 1
	}
	// Get filtered count
	filteredData := t.filterData(t.allData)
	totalItems := len(filteredData)
	return (totalItems + t.props.PageSize - 1) / t.props.PageSize
}

// CurrentPage returns the current page number
func (t *Table) CurrentPage() int {
	return t.currentPage
}

// getRowKey extracts the unique key from a row based on RowKey prop
func (t *Table) getRowKey(row map[string]any) any {
	return row[t.props.RowKey]
}

// getVisibleRowKeys returns keys for currently visible rows (after filter/sort/paginate)
func (t *Table) getVisibleRowKeys() []any {
	displayData := t.filterData(t.allData)
	displayData = t.sortData(displayData)
	displayData = t.paginateData(displayData)

	keys := make([]any, 0, len(displayData))
	for _, row := range displayData {
		key := t.getRowKey(row)
		if key != nil {
			keys = append(keys, key)
		}
	}
	return keys
}

// updateSelectAllState updates the select-all checkbox state (checked, unchecked, or indeterminate)
func (t *Table) updateSelectAllState(checkbox js.Value) {
	if checkbox.IsUndefined() || checkbox.IsNull() {
		return
	}

	visibleKeys := t.getVisibleRowKeys()
	if len(visibleKeys) == 0 {
		checkbox.Set("checked", false)
		checkbox.Set("indeterminate", false)
		return
	}

	selectedCount := 0
	for _, key := range visibleKeys {
		if t.selectedKeys[key] {
			selectedCount++
		}
	}

	if selectedCount == 0 {
		checkbox.Set("checked", false)
		checkbox.Set("indeterminate", false)
	} else if selectedCount == len(visibleKeys) {
		checkbox.Set("checked", true)
		checkbox.Set("indeterminate", false)
	} else {
		checkbox.Set("checked", false)
		checkbox.Set("indeterminate", true)
	}
}

// handleSelectAll handles click on select-all checkbox
func (t *Table) handleSelectAll(checked bool) {
	visibleKeys := t.getVisibleRowKeys()

	if checked {
		// Select all visible rows
		for _, key := range visibleKeys {
			t.selectedKeys[key] = true
		}
	} else {
		// Deselect all visible rows
		for _, key := range visibleKeys {
			delete(t.selectedKeys, key)
		}
	}

	// Update row checkboxes and notify
	t.updateRowCheckboxes()
	t.notifySelectionChange()
}

// handleRowSelection handles click on individual row checkbox
func (t *Table) handleRowSelection(key any, checked bool) {
	if checked {
		t.selectedKeys[key] = true
	} else {
		delete(t.selectedKeys, key)
	}

	// Update select-all checkbox state
	t.updateSelectAllState(t.selectAllCb)
	t.notifySelectionChange()
}

// updateRowCheckboxes updates the checked state of all row checkboxes
func (t *Table) updateRowCheckboxes() {
	visibleKeys := t.getVisibleRowKeys()
	for i, checkbox := range t.rowCheckboxes {
		if i < len(visibleKeys) {
			key := visibleKeys[i]
			checkbox.Set("checked", t.selectedKeys[key])
		}
	}
	// Update select-all too
	t.updateSelectAllState(t.selectAllCb)
}

// notifySelectionChange calls the OnSelectionChange callback with current selection
func (t *Table) notifySelectionChange() {
	// Update bulk action bar visibility
	t.updateBulkActionBar()

	if t.props.OnSelectionChange != nil {
		keys := make([]any, 0, len(t.selectedKeys))
		for key := range t.selectedKeys {
			keys = append(keys, key)
		}
		t.props.OnSelectionChange(keys)
	}
}

// SelectedKeys returns the keys of currently selected rows
func (t *Table) SelectedKeys() []any {
	keys := make([]any, 0, len(t.selectedKeys))
	for key := range t.selectedKeys {
		keys = append(keys, key)
	}
	return keys
}

// SelectedRows returns the full row data for selected rows
func (t *Table) SelectedRows() []map[string]any {
	rows := make([]map[string]any, 0, len(t.selectedKeys))
	for _, row := range t.allData {
		key := t.getRowKey(row)
		if t.selectedKeys[key] {
			rows = append(rows, row)
		}
	}
	return rows
}

// SelectAll selects all visible rows (respects filter/pagination)
func (t *Table) SelectAll() {
	visibleKeys := t.getVisibleRowKeys()
	for _, key := range visibleKeys {
		t.selectedKeys[key] = true
	}
	t.renderData()
	t.notifySelectionChange()
}

// ClearSelection deselects all rows
func (t *Table) ClearSelection() {
	t.selectedKeys = make(map[any]bool)
	t.renderData()
	t.notifySelectionChange()
}

// SetSelection programmatically sets the selection to specific keys
func (t *Table) SetSelection(keys []any) {
	t.selectedKeys = make(map[any]bool)
	for _, key := range keys {
		t.selectedKeys[key] = true
	}
	t.renderData()
	t.notifySelectionChange()
}

// IsSelected checks if a row with the given key is selected
func (t *Table) IsSelected(key any) bool {
	return t.selectedKeys[key]
}

// SelectionCount returns the number of selected rows
func (t *Table) SelectionCount() int {
	return len(t.selectedKeys)
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

// showEmptyState displays the appropriate empty state
// noData=true means the table has no data at all
// noData=false means the filter returned no results
func (t *Table) showEmptyState(noData bool) {
	// Hide table and pagination
	t.tableWrapper.Set("className", "hidden")
	if !t.paginationMount.IsUndefined() && !t.paginationMount.IsNull() {
		t.paginationMount.Set("className", "hidden")
	}

	// Clear and populate empty state container
	t.emptyStateEl.Set("innerHTML", "")
	t.emptyStateEl.Set("className", "")

	var emptyState *EmptyState

	// Use custom empty state if provided
	if t.props.EmptyState != nil {
		emptyState = t.props.EmptyState
	} else if noData {
		// No data at all - show "no data" state
		title := t.props.EmptyTitle
		if title == "" {
			title = "No data"
		}
		desc := t.props.EmptyDescription
		if desc == "" {
			desc = "There's nothing here yet."
		}
		emptyState = NewEmptyState(EmptyStateProps{
			Icon:        "📭",
			Title:       title,
			Description: desc,
		})
	} else {
		// Has data but filter returned nothing - show "no results" state
		emptyState = NoResults(func() {
			t.ClearFilter()
		})
	}

	t.emptyStateEl.Call("appendChild", emptyState.Element())
}

// hideEmptyState hides the empty state and shows the table
func (t *Table) hideEmptyState() {
	// Show table
	t.tableWrapper.Set("className", "overflow-x-auto")

	// Show pagination if enabled
	if !t.paginationMount.IsUndefined() && !t.paginationMount.IsNull() {
		t.paginationMount.Set("className", "mt-4")
	}

	// Hide empty state
	t.emptyStateEl.Set("className", "hidden")
}
