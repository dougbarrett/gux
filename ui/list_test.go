package ui

import (
	"strings"
	"testing"

	"github.com/dougbarrett/gux/core"
)

// Note: renderHTML helper is defined in button_test.go

func TestList_BasicStructure(t *testing.T) {
	html := renderHTML(List(ListProps{
		Children: []core.Node{core.Text("list content")},
	}))

	if !strings.Contains(html, "<ul") {
		t.Error("List should render ul element")
	}
	if !strings.Contains(html, "list content") {
		t.Error("List should render children")
	}
}

func TestList_Dividers(t *testing.T) {
	html := renderHTML(List(ListProps{}))

	if !strings.Contains(html, "divide-y") {
		t.Error("List should have divide-y class")
	}
	if !strings.Contains(html, "divide-gray-200") {
		t.Error("List should have divide-gray-200 class")
	}
	if !strings.Contains(html, "dark:divide-gray-700") {
		t.Error("List should have dark:divide-gray-700 class")
	}
}

func TestList_CustomClass(t *testing.T) {
	html := renderHTML(List(ListProps{Class: "my-custom-list"}))

	if !strings.Contains(html, "my-custom-list") {
		t.Error("List should include custom class")
	}
	if !strings.Contains(html, "divide-y") {
		t.Error("List should retain default divide-y class with custom class")
	}
}

func TestListItem_BasicStructure(t *testing.T) {
	html := renderHTML(ListItem(ListItemProps{
		Children: []core.Node{core.Text("item content")},
	}))

	if !strings.Contains(html, "<li") {
		t.Error("ListItem should render li element")
	}
	if !strings.Contains(html, "item content") {
		t.Error("ListItem should render children")
	}
}

func TestListItem_Padding(t *testing.T) {
	html := renderHTML(ListItem(ListItemProps{}))

	if !strings.Contains(html, "px-4") {
		t.Error("ListItem should have px-4 class")
	}
	if !strings.Contains(html, "py-4") {
		t.Error("ListItem should have py-4 class")
	}
}

func TestListItem_Hover(t *testing.T) {
	html := renderHTML(ListItem(ListItemProps{}))

	if !strings.Contains(html, "hover:bg-gray-50") {
		t.Error("ListItem should have hover:bg-gray-50 class")
	}
	if !strings.Contains(html, "dark:hover:bg-gray-800") {
		t.Error("ListItem should have dark:hover:bg-gray-800 class")
	}
}

func TestListItem_CustomClass(t *testing.T) {
	html := renderHTML(ListItem(ListItemProps{Class: "custom-item"}))

	if !strings.Contains(html, "custom-item") {
		t.Error("ListItem should include custom class")
	}
	if !strings.Contains(html, "px-4") {
		t.Error("ListItem should retain default px-4 with custom class")
	}
	if !strings.Contains(html, "hover:bg-gray-50") {
		t.Error("ListItem should retain hover styling with custom class")
	}
}

func TestList_Compound(t *testing.T) {
	html := renderHTML(List(ListProps{
		Children: []core.Node{
			ListItem(ListItemProps{
				Children: []core.Node{core.Text("First item")},
			}),
			ListItem(ListItemProps{
				Children: []core.Node{core.Text("Second item")},
			}),
			ListItem(ListItemProps{
				Children: []core.Node{core.Text("Third item")},
			}),
		},
	}))

	// Verify structure
	if !strings.Contains(html, "<ul") {
		t.Error("Compound List should include ul element")
	}
	if !strings.Contains(html, "<li") {
		t.Error("Compound List should include li elements")
	}

	// Verify content
	if !strings.Contains(html, "First item") {
		t.Error("Compound List should include 'First item'")
	}
	if !strings.Contains(html, "Second item") {
		t.Error("Compound List should include 'Second item'")
	}
	if !strings.Contains(html, "Third item") {
		t.Error("Compound List should include 'Third item'")
	}

	// Verify list dividers
	if !strings.Contains(html, "divide-y") {
		t.Error("Compound List should have divide-y for dividers")
	}

	// Verify item hover styling
	if !strings.Contains(html, "hover:bg-gray-50") {
		t.Error("Compound List items should have hover styling")
	}
}
