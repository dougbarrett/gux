package ui

import (
	"strings"
	"testing"

	"github.com/dougbarrett/gux/core"
)

// Note: renderHTML helper is defined in button_test.go

// testUser is a simple struct for testing DataTable
type testUser struct {
	Name  string
	Email string
}

func TestDataTable_BasicRender(t *testing.T) {
	users := []testUser{
		{Name: "John", Email: "john@example.com"},
	}

	html := renderHTML(DataTable(DataTableProps[testUser]{
		Data: users,
		Columns: []ColumnDef[testUser]{
			{Header: "Name", Render: func(u testUser) core.Node { return core.Text(u.Name) }},
			{Header: "Email", Render: func(u testUser) core.Node { return core.Text(u.Email) }},
		},
	}))

	// Should contain table structure
	if !strings.Contains(html, "<table") {
		t.Error("should render table element")
	}
	if !strings.Contains(html, "<thead") {
		t.Error("should render thead element")
	}
	if !strings.Contains(html, "<tbody") {
		t.Error("should render tbody element")
	}
	if !strings.Contains(html, "<tr") {
		t.Error("should render tr elements")
	}
	if !strings.Contains(html, "<th") {
		t.Error("should render th elements")
	}
	if !strings.Contains(html, "<td") {
		t.Error("should render td elements")
	}

	// Should contain header text
	if !strings.Contains(html, "Name") {
		t.Error("should render Name header")
	}
	if !strings.Contains(html, "Email") {
		t.Error("should render Email header")
	}

	// Should contain data
	if !strings.Contains(html, "John") {
		t.Error("should render John")
	}
	if !strings.Contains(html, "john@example.com") {
		t.Error("should render john@example.com")
	}
}

func TestDataTable_Columns(t *testing.T) {
	users := []testUser{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}

	html := renderHTML(DataTable(DataTableProps[testUser]{
		Data: users,
		Columns: []ColumnDef[testUser]{
			{Header: "Name", Render: func(u testUser) core.Node { return core.Text(u.Name) }},
			{Header: "Email", Render: func(u testUser) core.Node { return core.Text(u.Email) }},
		},
	}))

	// Should contain all user data
	if !strings.Contains(html, "Alice") {
		t.Error("should render Alice")
	}
	if !strings.Contains(html, "alice@example.com") {
		t.Error("should render alice@example.com")
	}
	if !strings.Contains(html, "Bob") {
		t.Error("should render Bob")
	}
	if !strings.Contains(html, "bob@example.com") {
		t.Error("should render bob@example.com")
	}
}

func TestDataTable_Striped(t *testing.T) {
	users := []testUser{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
		{Name: "Charlie", Email: "charlie@example.com"},
	}

	html := renderHTML(DataTable(DataTableProps[testUser]{
		Data:    users,
		Striped: true,
		Columns: []ColumnDef[testUser]{
			{Header: "Name", Render: func(u testUser) core.Node { return core.Text(u.Name) }},
		},
	}))

	// Should have alternating row styling
	// Even rows (index 1, 3, etc.) should have bg-gray-50
	if !strings.Contains(html, "bg-gray-50") {
		t.Error("striped should add bg-gray-50 to odd rows")
	}
	if !strings.Contains(html, "dark:bg-gray-800") {
		t.Error("striped should add dark:bg-gray-800 to odd rows")
	}
}

func TestDataTable_Hoverable(t *testing.T) {
	users := []testUser{
		{Name: "Alice", Email: "alice@example.com"},
	}

	html := renderHTML(DataTable(DataTableProps[testUser]{
		Data:      users,
		Hoverable: true,
		Columns: []ColumnDef[testUser]{
			{Header: "Name", Render: func(u testUser) core.Node { return core.Text(u.Name) }},
		},
	}))

	// Should have hover classes
	if !strings.Contains(html, "hover:bg-gray-100") {
		t.Error("hoverable should add hover:bg-gray-100")
	}
	if !strings.Contains(html, "dark:hover:bg-gray-700") {
		t.Error("hoverable should add dark:hover:bg-gray-700")
	}
	if !strings.Contains(html, "cursor-pointer") {
		t.Error("hoverable should add cursor-pointer")
	}
}

func TestDataTable_EmptyData(t *testing.T) {
	html := renderHTML(DataTable(DataTableProps[testUser]{
		Data: []testUser{},
		Columns: []ColumnDef[testUser]{
			{Header: "Name", Render: func(u testUser) core.Node { return core.Text(u.Name) }},
		},
	}))

	// Should still render table structure with headers
	if !strings.Contains(html, "<table") {
		t.Error("empty data should still render table element")
	}
	if !strings.Contains(html, "<thead") {
		t.Error("empty data should still render thead element")
	}
	if !strings.Contains(html, "<tbody") {
		t.Error("empty data should still render tbody element")
	}
	if !strings.Contains(html, "Name") {
		t.Error("empty data should still render headers")
	}
}

func TestDataTable_CustomClass(t *testing.T) {
	html := renderHTML(DataTable(DataTableProps[testUser]{
		Data:  []testUser{{Name: "Alice", Email: "alice@example.com"}},
		Class: "my-custom-table",
		Columns: []ColumnDef[testUser]{
			{Header: "Name", Render: func(u testUser) core.Node { return core.Text(u.Name) }},
		},
	}))

	if !strings.Contains(html, "my-custom-table") {
		t.Error("should apply custom class")
	}
	// Should retain default table styling
	if !strings.Contains(html, "min-w-full") {
		t.Error("should retain default min-w-full class from Table")
	}
}

func TestDataTable_ColumnWidth(t *testing.T) {
	html := renderHTML(DataTable(DataTableProps[testUser]{
		Data: []testUser{{Name: "Alice", Email: "alice@example.com"}},
		Columns: []ColumnDef[testUser]{
			{Header: "Name", Width: "w-48", Render: func(u testUser) core.Node { return core.Text(u.Name) }},
			{Header: "Email", Render: func(u testUser) core.Node { return core.Text(u.Email) }},
		},
	}))

	if !strings.Contains(html, "w-48") {
		t.Error("should apply column width class")
	}
}

func TestDataTable_StripedAndHoverable(t *testing.T) {
	users := []testUser{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}

	html := renderHTML(DataTable(DataTableProps[testUser]{
		Data:      users,
		Striped:   true,
		Hoverable: true,
		Columns: []ColumnDef[testUser]{
			{Header: "Name", Render: func(u testUser) core.Node { return core.Text(u.Name) }},
		},
	}))

	// Should have both striped and hoverable classes
	if !strings.Contains(html, "bg-gray-50") {
		t.Error("should have striped bg-gray-50")
	}
	if !strings.Contains(html, "hover:bg-gray-100") {
		t.Error("should have hoverable hover:bg-gray-100")
	}
	if !strings.Contains(html, "cursor-pointer") {
		t.Error("should have cursor-pointer")
	}
}

func TestDataTable_UsesTableCompound(t *testing.T) {
	html := renderHTML(DataTable(DataTableProps[testUser]{
		Data: []testUser{{Name: "Alice", Email: "alice@example.com"}},
		Columns: []ColumnDef[testUser]{
			{Header: "Name", Render: func(u testUser) core.Node { return core.Text(u.Name) }},
		},
	}))

	// Verify Table compound styling is applied
	if !strings.Contains(html, "overflow-x-auto") {
		t.Error("should use Table which adds overflow-x-auto wrapper")
	}
	if !strings.Contains(html, "divide-y") {
		t.Error("should have divide-y from Table")
	}
	if !strings.Contains(html, "bg-gray-50") {
		t.Error("should have bg-gray-50 from Thead")
	}
	if !strings.Contains(html, "bg-white") {
		t.Error("should have bg-white from Tbody")
	}
	if !strings.Contains(html, "uppercase") {
		t.Error("should have uppercase from Th")
	}
	if !strings.Contains(html, "whitespace-nowrap") {
		t.Error("should have whitespace-nowrap from Td")
	}
}

func TestDataTable_MultipleColumns(t *testing.T) {
	type product struct {
		Name  string
		Price string
		Stock int
	}

	products := []product{
		{Name: "Widget", Price: "$9.99", Stock: 10},
		{Name: "Gadget", Price: "$19.99", Stock: 5},
	}

	html := renderHTML(DataTable(DataTableProps[product]{
		Data: products,
		Columns: []ColumnDef[product]{
			{Header: "Product", Render: func(p product) core.Node { return core.Text(p.Name) }},
			{Header: "Price", Render: func(p product) core.Node { return core.Text(p.Price) }},
			{Header: "Stock", Render: func(p product) core.Node {
				return core.Span(core.Class("font-bold"), core.Text(string(rune('0'+p.Stock))))
			}},
		},
	}))

	// Verify headers
	if !strings.Contains(html, "Product") {
		t.Error("should render Product header")
	}
	if !strings.Contains(html, "Price") {
		t.Error("should render Price header")
	}
	if !strings.Contains(html, "Stock") {
		t.Error("should render Stock header")
	}

	// Verify data
	if !strings.Contains(html, "Widget") {
		t.Error("should render Widget")
	}
	if !strings.Contains(html, "$9.99") {
		t.Error("should render $9.99")
	}
	if !strings.Contains(html, "Gadget") {
		t.Error("should render Gadget")
	}
	if !strings.Contains(html, "$19.99") {
		t.Error("should render $19.99")
	}

	// Verify custom rendering (font-bold span for stock)
	if !strings.Contains(html, "font-bold") {
		t.Error("should apply custom rendering for stock column")
	}
}
