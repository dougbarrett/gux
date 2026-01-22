package ui

import (
	"strings"
	"testing"

	"github.com/dougbarrett/gux/core"
)

// Note: renderHTML helper is defined in button_test.go

func TestTable_BasicStructure(t *testing.T) {
	html := renderHTML(Table(TableProps{
		Children: []core.Node{core.Text("table content")},
	}))

	if !strings.Contains(html, "<table") {
		t.Error("Table should render table element")
	}
	if !strings.Contains(html, "table content") {
		t.Error("Table should render children")
	}
}

func TestTable_OverflowWrapper(t *testing.T) {
	html := renderHTML(Table(TableProps{}))

	if !strings.Contains(html, "<div") {
		t.Error("Table should be wrapped in div")
	}
	if !strings.Contains(html, "overflow-x-auto") {
		t.Error("Table wrapper should have overflow-x-auto class")
	}
}

func TestTable_DefaultStyling(t *testing.T) {
	html := renderHTML(Table(TableProps{}))

	if !strings.Contains(html, "min-w-full") {
		t.Error("Table should have min-w-full class")
	}
	if !strings.Contains(html, "divide-y") {
		t.Error("Table should have divide-y class")
	}
	if !strings.Contains(html, "divide-gray-200") {
		t.Error("Table should have divide-gray-200 class")
	}
}

func TestTable_CustomClass(t *testing.T) {
	html := renderHTML(Table(TableProps{Class: "my-custom-table"}))

	if !strings.Contains(html, "my-custom-table") {
		t.Error("Table should include custom class")
	}
	if !strings.Contains(html, "min-w-full") {
		t.Error("Table should retain default min-w-full class with custom class")
	}
}

func TestThead_Styling(t *testing.T) {
	html := renderHTML(Thead(TheadProps{}))

	if !strings.Contains(html, "<thead") {
		t.Error("Thead should render thead element")
	}
	if !strings.Contains(html, "bg-gray-50") {
		t.Error("Thead should have bg-gray-50 class")
	}
	if !strings.Contains(html, "dark:bg-gray-800") {
		t.Error("Thead should have dark:bg-gray-800 class")
	}
}

func TestThead_CustomClass(t *testing.T) {
	html := renderHTML(Thead(TheadProps{Class: "custom-head"}))

	if !strings.Contains(html, "custom-head") {
		t.Error("Thead should include custom class")
	}
	if !strings.Contains(html, "bg-gray-50") {
		t.Error("Thead should retain default bg-gray-50 with custom class")
	}
}

func TestThead_Children(t *testing.T) {
	html := renderHTML(Thead(TheadProps{
		Children: []core.Node{core.Text("header content")},
	}))

	if !strings.Contains(html, "header content") {
		t.Error("Thead should render children")
	}
}

func TestTbody_Styling(t *testing.T) {
	html := renderHTML(Tbody(TbodyProps{}))

	if !strings.Contains(html, "<tbody") {
		t.Error("Tbody should render tbody element")
	}
	if !strings.Contains(html, "bg-white") {
		t.Error("Tbody should have bg-white class")
	}
	if !strings.Contains(html, "dark:bg-gray-900") {
		t.Error("Tbody should have dark:bg-gray-900 class")
	}
	if !strings.Contains(html, "divide-y") {
		t.Error("Tbody should have divide-y class")
	}
}

func TestTbody_CustomClass(t *testing.T) {
	html := renderHTML(Tbody(TbodyProps{Class: "custom-body"}))

	if !strings.Contains(html, "custom-body") {
		t.Error("Tbody should include custom class")
	}
	if !strings.Contains(html, "bg-white") {
		t.Error("Tbody should retain default bg-white with custom class")
	}
}

func TestTbody_Children(t *testing.T) {
	html := renderHTML(Tbody(TbodyProps{
		Children: []core.Node{core.Text("body content")},
	}))

	if !strings.Contains(html, "body content") {
		t.Error("Tbody should render children")
	}
}

func TestTr_BasicStructure(t *testing.T) {
	html := renderHTML(Tr(TrProps{
		Children: []core.Node{core.Text("row content")},
	}))

	if !strings.Contains(html, "<tr") {
		t.Error("Tr should render tr element")
	}
	if !strings.Contains(html, "row content") {
		t.Error("Tr should render children")
	}
}

func TestTr_CustomClass(t *testing.T) {
	html := renderHTML(Tr(TrProps{Class: "custom-row"}))

	if !strings.Contains(html, "custom-row") {
		t.Error("Tr should include custom class")
	}
}

func TestTh_Styling(t *testing.T) {
	html := renderHTML(Th(ThProps{}))

	if !strings.Contains(html, "<th") {
		t.Error("Th should render th element")
	}
	if !strings.Contains(html, "px-6") {
		t.Error("Th should have px-6 class")
	}
	if !strings.Contains(html, "py-3") {
		t.Error("Th should have py-3 class")
	}
	if !strings.Contains(html, "text-left") {
		t.Error("Th should have text-left class")
	}
	if !strings.Contains(html, "text-xs") {
		t.Error("Th should have text-xs class")
	}
	if !strings.Contains(html, "font-medium") {
		t.Error("Th should have font-medium class")
	}
	if !strings.Contains(html, "uppercase") {
		t.Error("Th should have uppercase class")
	}
	if !strings.Contains(html, "tracking-wider") {
		t.Error("Th should have tracking-wider class")
	}
}

func TestTh_CustomClass(t *testing.T) {
	html := renderHTML(Th(ThProps{Class: "custom-header-cell"}))

	if !strings.Contains(html, "custom-header-cell") {
		t.Error("Th should include custom class")
	}
	if !strings.Contains(html, "px-6") {
		t.Error("Th should retain default px-6 with custom class")
	}
}

func TestTh_Children(t *testing.T) {
	html := renderHTML(Th(ThProps{
		Children: []core.Node{core.Text("Column Header")},
	}))

	if !strings.Contains(html, "Column Header") {
		t.Error("Th should render children")
	}
}

func TestTd_Styling(t *testing.T) {
	html := renderHTML(Td(TdProps{}))

	if !strings.Contains(html, "<td") {
		t.Error("Td should render td element")
	}
	if !strings.Contains(html, "px-6") {
		t.Error("Td should have px-6 class")
	}
	if !strings.Contains(html, "py-4") {
		t.Error("Td should have py-4 class")
	}
	if !strings.Contains(html, "whitespace-nowrap") {
		t.Error("Td should have whitespace-nowrap class")
	}
	if !strings.Contains(html, "text-sm") {
		t.Error("Td should have text-sm class")
	}
	if !strings.Contains(html, "text-gray-900") {
		t.Error("Td should have text-gray-900 class")
	}
}

func TestTd_CustomClass(t *testing.T) {
	html := renderHTML(Td(TdProps{Class: "custom-cell"}))

	if !strings.Contains(html, "custom-cell") {
		t.Error("Td should include custom class")
	}
	if !strings.Contains(html, "px-6") {
		t.Error("Td should retain default px-6 with custom class")
	}
}

func TestTd_Children(t *testing.T) {
	html := renderHTML(Td(TdProps{
		Children: []core.Node{core.Text("Cell Content")},
	}))

	if !strings.Contains(html, "Cell Content") {
		t.Error("Td should render children")
	}
}

func TestTable_Compound(t *testing.T) {
	html := renderHTML(Table(TableProps{
		Children: []core.Node{
			Thead(TheadProps{
				Children: []core.Node{
					Tr(TrProps{
						Children: []core.Node{
							Th(ThProps{Children: []core.Node{core.Text("Name")}}),
							Th(ThProps{Children: []core.Node{core.Text("Email")}}),
						},
					}),
				},
			}),
			Tbody(TbodyProps{
				Children: []core.Node{
					Tr(TrProps{
						Children: []core.Node{
							Td(TdProps{Children: []core.Node{core.Text("John Doe")}}),
							Td(TdProps{Children: []core.Node{core.Text("john@example.com")}}),
						},
					}),
					Tr(TrProps{
						Children: []core.Node{
							Td(TdProps{Children: []core.Node{core.Text("Jane Smith")}}),
							Td(TdProps{Children: []core.Node{core.Text("jane@example.com")}}),
						},
					}),
				},
			}),
		},
	}))

	// Verify all parts are present
	if !strings.Contains(html, "<table") {
		t.Error("Compound Table should include table element")
	}
	if !strings.Contains(html, "<thead") {
		t.Error("Compound Table should include thead element")
	}
	if !strings.Contains(html, "<tbody") {
		t.Error("Compound Table should include tbody element")
	}
	if !strings.Contains(html, "<tr") {
		t.Error("Compound Table should include tr elements")
	}
	if !strings.Contains(html, "<th") {
		t.Error("Compound Table should include th elements")
	}
	if !strings.Contains(html, "<td") {
		t.Error("Compound Table should include td elements")
	}

	// Verify content
	if !strings.Contains(html, "Name") {
		t.Error("Compound Table should include header text 'Name'")
	}
	if !strings.Contains(html, "Email") {
		t.Error("Compound Table should include header text 'Email'")
	}
	if !strings.Contains(html, "John Doe") {
		t.Error("Compound Table should include cell text 'John Doe'")
	}
	if !strings.Contains(html, "john@example.com") {
		t.Error("Compound Table should include cell text 'john@example.com'")
	}
	if !strings.Contains(html, "Jane Smith") {
		t.Error("Compound Table should include cell text 'Jane Smith'")
	}

	// Verify overflow wrapper
	if !strings.Contains(html, "overflow-x-auto") {
		t.Error("Compound Table should have overflow wrapper")
	}

	// Verify thead styling
	if !strings.Contains(html, "bg-gray-50") {
		t.Error("Compound Table thead should have bg-gray-50")
	}

	// Verify tbody styling
	if !strings.Contains(html, "bg-white") {
		t.Error("Compound Table tbody should have bg-white")
	}
}
