package ui

import (
	"strings"
	"testing"

	"github.com/dougbarrett/gux/core"
)

// ---- Tabs tests ----

func TestTabs_Renders(t *testing.T) {
	tabs := Tabs(TabsProps{
		Children: []core.Node{core.Text("content")},
	})
	html := renderHTML(tabs)

	// Should render as div with w-full class
	if !strings.HasPrefix(html, "<div") {
		t.Errorf("expected <div> element, got: %s", html)
	}

	if !strings.Contains(html, "w-full") {
		t.Errorf("expected w-full class, got: %s", html)
	}
}

func TestTabs_CustomClass(t *testing.T) {
	tabs := Tabs(TabsProps{
		Class:    "my-custom-tabs",
		Children: []core.Node{core.Text("content")},
	})
	html := renderHTML(tabs)

	if !strings.Contains(html, "my-custom-tabs") {
		t.Errorf("expected custom class, got: %s", html)
	}

	// Should still have base class
	if !strings.Contains(html, "w-full") {
		t.Errorf("expected w-full class, got: %s", html)
	}
}

// ---- TabList tests ----

func TestTabList_AriaRole(t *testing.T) {
	tabList := TabList(TabListProps{
		Children: []core.Node{core.Text("tab")},
	})
	html := renderHTML(tabList)

	if !strings.Contains(html, `role="tablist"`) {
		t.Errorf("expected role=tablist, got: %s", html)
	}
}

func TestTabList_AriaLabel(t *testing.T) {
	tabList := TabList(TabListProps{
		Children: []core.Node{core.Text("tab")},
	})
	html := renderHTML(tabList)

	if !strings.Contains(html, `aria-label="Tabs"`) {
		t.Errorf("expected aria-label=Tabs, got: %s", html)
	}
}

func TestTabList_Classes(t *testing.T) {
	tabList := TabList(TabListProps{
		Children: []core.Node{core.Text("tab")},
	})
	html := renderHTML(tabList)

	// Should have flex and border-b classes
	if !strings.Contains(html, "flex") {
		t.Errorf("expected flex class, got: %s", html)
	}

	if !strings.Contains(html, "border-b") {
		t.Errorf("expected border-b class, got: %s", html)
	}

	if !strings.Contains(html, "overflow-x-auto") {
		t.Errorf("expected overflow-x-auto class, got: %s", html)
	}
}

func TestTabList_CustomClass(t *testing.T) {
	tabList := TabList(TabListProps{
		Class:    "my-tablist",
		Children: []core.Node{core.Text("tab")},
	})
	html := renderHTML(tabList)

	if !strings.Contains(html, "my-tablist") {
		t.Errorf("expected custom class, got: %s", html)
	}

	// Should still have base classes
	if !strings.Contains(html, "flex") {
		t.Errorf("expected flex class, got: %s", html)
	}
}

// ---- Tab tests ----

func TestTab_AriaRole(t *testing.T) {
	tab := Tab(TabProps{
		Children: []core.Node{core.Text("Tab 1")},
	})
	html := renderHTML(tab)

	// Should render as button
	if !strings.HasPrefix(html, "<button") {
		t.Errorf("expected <button> element, got: %s", html)
	}

	if !strings.Contains(html, `role="tab"`) {
		t.Errorf("expected role=tab, got: %s", html)
	}
}

func TestTab_Active(t *testing.T) {
	tab := Tab(TabProps{
		Active:   true,
		Children: []core.Node{core.Text("Active Tab")},
	})
	html := renderHTML(tab)

	// Should have aria-selected=true
	if !strings.Contains(html, `aria-selected="true"`) {
		t.Errorf("expected aria-selected=true, got: %s", html)
	}

	// Should have tabindex=0
	if !strings.Contains(html, `tabindex="0"`) {
		t.Errorf("expected tabindex=0, got: %s", html)
	}

	// Should have active styling
	if !strings.Contains(html, "border-blue-500") {
		t.Errorf("expected active border class, got: %s", html)
	}

	if !strings.Contains(html, "text-blue-600") {
		t.Errorf("expected active text class, got: %s", html)
	}
}

func TestTab_Inactive(t *testing.T) {
	tab := Tab(TabProps{
		Active:   false,
		Children: []core.Node{core.Text("Inactive Tab")},
	})
	html := renderHTML(tab)

	// Should have aria-selected=false
	if !strings.Contains(html, `aria-selected="false"`) {
		t.Errorf("expected aria-selected=false, got: %s", html)
	}

	// Should have tabindex=-1
	if !strings.Contains(html, `tabindex="-1"`) {
		t.Errorf("expected tabindex=-1, got: %s", html)
	}

	// Should have inactive styling
	if !strings.Contains(html, "border-transparent") {
		t.Errorf("expected inactive border class, got: %s", html)
	}

	if !strings.Contains(html, "text-gray-500") {
		t.Errorf("expected inactive text class, got: %s", html)
	}
}

func TestTab_Disabled(t *testing.T) {
	tab := Tab(TabProps{
		Disabled: true,
		Children: []core.Node{core.Text("Disabled Tab")},
	})
	html := renderHTML(tab)

	// Should have disabled attribute
	if !strings.Contains(html, `disabled="disabled"`) {
		t.Errorf("expected disabled attribute, got: %s", html)
	}

	// Should have aria-disabled=true
	if !strings.Contains(html, `aria-disabled="true"`) {
		t.Errorf("expected aria-disabled=true, got: %s", html)
	}

	// Should have disabled styling
	if !strings.Contains(html, "text-gray-300") {
		t.Errorf("expected disabled text class, got: %s", html)
	}

	if !strings.Contains(html, "cursor-not-allowed") {
		t.Errorf("expected cursor-not-allowed class, got: %s", html)
	}
}

func TestTab_Children(t *testing.T) {
	tab := Tab(TabProps{
		Children: []core.Node{
			core.Span(core.Class("icon"), core.Text("*")),
			core.Text(" Settings"),
		},
	})
	html := renderHTML(tab)

	// Should contain children
	if !strings.Contains(html, "Settings") {
		t.Errorf("expected text content, got: %s", html)
	}

	if !strings.Contains(html, "<span") {
		t.Errorf("expected span element, got: %s", html)
	}
}

func TestTab_CustomClass(t *testing.T) {
	tab := Tab(TabProps{
		Class:    "my-tab",
		Children: []core.Node{core.Text("Custom")},
	})
	html := renderHTML(tab)

	if !strings.Contains(html, "my-tab") {
		t.Errorf("expected custom class, got: %s", html)
	}

	// Should still have base classes
	if !strings.Contains(html, "font-medium") {
		t.Errorf("expected base font class, got: %s", html)
	}
}

// ---- TabPanel tests ----

func TestTabPanel_AriaRole(t *testing.T) {
	panel := TabPanel(TabPanelProps{
		Children: []core.Node{core.Text("Content")},
	})
	html := renderHTML(panel)

	// Should render as div
	if !strings.HasPrefix(html, "<div") {
		t.Errorf("expected <div> element, got: %s", html)
	}

	if !strings.Contains(html, `role="tabpanel"`) {
		t.Errorf("expected role=tabpanel, got: %s", html)
	}
}

func TestTabPanel_Tabindex(t *testing.T) {
	panel := TabPanel(TabPanelProps{
		Children: []core.Node{core.Text("Content")},
	})
	html := renderHTML(panel)

	if !strings.Contains(html, `tabindex="0"`) {
		t.Errorf("expected tabindex=0, got: %s", html)
	}
}

func TestTabPanel_Active(t *testing.T) {
	panel := TabPanel(TabPanelProps{
		Active:   true,
		Children: []core.Node{core.Text("Visible Content")},
	})
	html := renderHTML(panel)

	// Should NOT have hidden class
	if strings.Contains(html, "hidden") {
		t.Errorf("expected no hidden class for active panel, got: %s", html)
	}

	// Should have mt-4 base class
	if !strings.Contains(html, "mt-4") {
		t.Errorf("expected mt-4 class, got: %s", html)
	}
}

func TestTabPanel_Inactive(t *testing.T) {
	panel := TabPanel(TabPanelProps{
		Active:   false,
		Children: []core.Node{core.Text("Hidden Content")},
	})
	html := renderHTML(panel)

	// Should have hidden class
	if !strings.Contains(html, "hidden") {
		t.Errorf("expected hidden class for inactive panel, got: %s", html)
	}
}

func TestTabPanel_Children(t *testing.T) {
	panel := TabPanel(TabPanelProps{
		Active: true,
		Children: []core.Node{
			core.P(core.Class("text"), core.Text("Paragraph")),
			core.Div(core.Class("box"), core.Text("Box")),
		},
	})
	html := renderHTML(panel)

	if !strings.Contains(html, "<p") {
		t.Errorf("expected p element, got: %s", html)
	}

	if !strings.Contains(html, "Paragraph") {
		t.Errorf("expected paragraph text, got: %s", html)
	}

	if !strings.Contains(html, "Box") {
		t.Errorf("expected box text, got: %s", html)
	}
}

func TestTabPanel_CustomClass(t *testing.T) {
	panel := TabPanel(TabPanelProps{
		Active:   true,
		Class:    "my-panel",
		Children: []core.Node{core.Text("Content")},
	})
	html := renderHTML(panel)

	if !strings.Contains(html, "my-panel") {
		t.Errorf("expected custom class, got: %s", html)
	}

	// Should still have base class
	if !strings.Contains(html, "mt-4") {
		t.Errorf("expected mt-4 class, got: %s", html)
	}
}

// ---- Integration test ----

func TestTabs_FullComposition(t *testing.T) {
	// Test that all components work together
	tabs := Tabs(TabsProps{
		Children: []core.Node{
			TabList(TabListProps{
				Children: []core.Node{
					Tab(TabProps{Active: true, Children: []core.Node{core.Text("Tab 1")}}),
					Tab(TabProps{Active: false, Children: []core.Node{core.Text("Tab 2")}}),
					Tab(TabProps{Disabled: true, Children: []core.Node{core.Text("Tab 3")}}),
				},
			}),
			TabPanel(TabPanelProps{Active: true, Children: []core.Node{core.Text("Content 1")}}),
			TabPanel(TabPanelProps{Active: false, Children: []core.Node{core.Text("Content 2")}}),
			TabPanel(TabPanelProps{Active: false, Children: []core.Node{core.Text("Content 3")}}),
		},
	})
	html := renderHTML(tabs)

	// Check structure
	if !strings.Contains(html, `role="tablist"`) {
		t.Errorf("expected tablist role, got: %s", html)
	}

	// Check tabs
	if strings.Count(html, `role="tab"`) != 3 {
		t.Errorf("expected 3 tabs, got: %s", html)
	}

	// Check panels
	if strings.Count(html, `role="tabpanel"`) != 3 {
		t.Errorf("expected 3 tabpanels, got: %s", html)
	}

	// Check active tab has correct attributes
	if !strings.Contains(html, `aria-selected="true"`) {
		t.Errorf("expected one tab with aria-selected=true, got: %s", html)
	}

	// Check disabled tab
	if !strings.Contains(html, `aria-disabled="true"`) {
		t.Errorf("expected disabled tab, got: %s", html)
	}

	// Check content rendering
	if !strings.Contains(html, "Content 1") {
		t.Errorf("expected Content 1, got: %s", html)
	}
	if !strings.Contains(html, "Content 2") {
		t.Errorf("expected Content 2, got: %s", html)
	}
	if !strings.Contains(html, "Content 3") {
		t.Errorf("expected Content 3, got: %s", html)
	}
}
