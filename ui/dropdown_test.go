package ui

import (
	"strings"
	"testing"

	"github.com/dougbarrett/gux/core"
)

// Dropdown tests

func TestDropdown_Renders(t *testing.T) {
	dropdown := Dropdown(DropdownProps{
		Children: []core.Node{core.Text("content")},
	})
	html := renderHTML(dropdown)

	// Should have wrapper with relative positioning
	if !strings.Contains(html, "relative") {
		t.Errorf("expected relative class, got: %s", html)
	}
	if !strings.Contains(html, "inline-block") {
		t.Errorf("expected inline-block class, got: %s", html)
	}
}

func TestDropdown_OpenBackdrop(t *testing.T) {
	dropdown := Dropdown(DropdownProps{
		Open:    true,
		OnClose: func() {},
		Children: []core.Node{core.Text("content")},
	})
	html := renderHTML(dropdown)

	// Should render backdrop when open
	if !strings.Contains(html, "fixed") {
		t.Errorf("expected fixed class for backdrop, got: %s", html)
	}
	if !strings.Contains(html, "inset-0") {
		t.Errorf("expected inset-0 class for backdrop, got: %s", html)
	}
	if !strings.Contains(html, "z-40") {
		t.Errorf("expected z-40 class for backdrop, got: %s", html)
	}
}

func TestDropdown_ClosedNoBackdrop(t *testing.T) {
	dropdown := Dropdown(DropdownProps{
		Open:    false,
		OnClose: func() {},
		Children: []core.Node{core.Text("content")},
	})
	html := renderHTML(dropdown)

	// Should NOT render backdrop when closed
	if strings.Contains(html, "fixed inset-0") {
		t.Errorf("expected no backdrop when closed, got: %s", html)
	}
}

func TestDropdown_CustomClass(t *testing.T) {
	dropdown := Dropdown(DropdownProps{
		Class:    "my-custom-class",
		Children: []core.Node{core.Text("content")},
	})
	html := renderHTML(dropdown)

	if !strings.Contains(html, "my-custom-class") {
		t.Errorf("expected custom class, got: %s", html)
	}
	// Should still have base classes
	if !strings.Contains(html, "relative") {
		t.Errorf("expected base relative class, got: %s", html)
	}
}

// DropdownTrigger tests

func TestDropdownTrigger_AriaHaspopup(t *testing.T) {
	trigger := DropdownTrigger(DropdownTriggerProps{
		Children: []core.Node{core.Text("Menu")},
	})
	html := renderHTML(trigger)

	if !strings.Contains(html, `aria-haspopup="menu"`) {
		t.Errorf("expected aria-haspopup=\"menu\", got: %s", html)
	}
}

func TestDropdownTrigger_AriaExpandedTrue(t *testing.T) {
	trigger := DropdownTrigger(DropdownTriggerProps{
		Open:     true,
		Children: []core.Node{core.Text("Menu")},
	})
	html := renderHTML(trigger)

	if !strings.Contains(html, `aria-expanded="true"`) {
		t.Errorf("expected aria-expanded=\"true\", got: %s", html)
	}
}

func TestDropdownTrigger_AriaExpandedFalse(t *testing.T) {
	trigger := DropdownTrigger(DropdownTriggerProps{
		Open:     false,
		Children: []core.Node{core.Text("Menu")},
	})
	html := renderHTML(trigger)

	if !strings.Contains(html, `aria-expanded="false"`) {
		t.Errorf("expected aria-expanded=\"false\", got: %s", html)
	}
}

func TestDropdownTrigger_Children(t *testing.T) {
	trigger := DropdownTrigger(DropdownTriggerProps{
		Children: []core.Node{
			core.Span(core.Class("icon"), core.Text("*")),
			core.Text("Actions"),
		},
	})
	html := renderHTML(trigger)

	if !strings.Contains(html, "<span") {
		t.Errorf("expected nested span, got: %s", html)
	}
	if !strings.Contains(html, "Actions") {
		t.Errorf("expected text content, got: %s", html)
	}
}

func TestDropdownTrigger_CustomClass(t *testing.T) {
	trigger := DropdownTrigger(DropdownTriggerProps{
		Class:    "trigger-custom",
		Children: []core.Node{core.Text("Menu")},
	})
	html := renderHTML(trigger)

	if !strings.Contains(html, "trigger-custom") {
		t.Errorf("expected custom class, got: %s", html)
	}
	if !strings.Contains(html, "inline-block") {
		t.Errorf("expected base inline-block class, got: %s", html)
	}
}

// DropdownMenu tests

func TestDropdownMenu_AriaRole(t *testing.T) {
	menu := DropdownMenu(DropdownMenuProps{
		Open:     true,
		Children: []core.Node{core.Text("Item")},
	})
	html := renderHTML(menu)

	if !strings.Contains(html, `role="menu"`) {
		t.Errorf("expected role=\"menu\", got: %s", html)
	}
}

func TestDropdownMenu_VisibleWhenOpen(t *testing.T) {
	menu := DropdownMenu(DropdownMenuProps{
		Open:     true,
		Children: []core.Node{core.Text("Item")},
	})
	html := renderHTML(menu)

	// Should NOT have hidden class when open
	if strings.Contains(html, "hidden") {
		t.Errorf("expected no hidden class when open, got: %s", html)
	}
}

func TestDropdownMenu_HiddenWhenClosed(t *testing.T) {
	menu := DropdownMenu(DropdownMenuProps{
		Open:     false,
		Children: []core.Node{core.Text("Item")},
	})
	html := renderHTML(menu)

	// Should have hidden class when closed
	if !strings.Contains(html, "hidden") {
		t.Errorf("expected hidden class when closed, got: %s", html)
	}
}

func TestDropdownMenu_AllPositions(t *testing.T) {
	tests := []struct {
		position      DropdownPosition
		expectedClass string
	}{
		{DropdownBottomLeft, "top-full left-0 mt-1"},
		{DropdownBottomRight, "top-full right-0 mt-1"},
		{DropdownTopLeft, "bottom-full left-0 mb-1"},
		{DropdownTopRight, "bottom-full right-0 mb-1"},
	}

	for _, tt := range tests {
		t.Run(string(tt.position), func(t *testing.T) {
			menu := DropdownMenu(DropdownMenuProps{
				Open:     true,
				Position: tt.position,
				Children: []core.Node{core.Text("Item")},
			})
			html := renderHTML(menu)

			if !strings.Contains(html, tt.expectedClass) {
				t.Errorf("position %s: expected class %q in: %s", tt.position, tt.expectedClass, html)
			}
		})
	}
}

func TestDropdownMenu_DefaultPosition(t *testing.T) {
	menu := DropdownMenu(DropdownMenuProps{
		Open:     true,
		Children: []core.Node{core.Text("Item")},
	})
	html := renderHTML(menu)

	// Default should be DropdownBottomLeft
	if !strings.Contains(html, "top-full left-0 mt-1") {
		t.Errorf("expected default DropdownBottomLeft position, got: %s", html)
	}
}

func TestDropdownMenu_Children(t *testing.T) {
	menu := DropdownMenu(DropdownMenuProps{
		Open: true,
		Children: []core.Node{
			core.Div(core.Class("item"), core.Text("First")),
			core.Div(core.Class("item"), core.Text("Second")),
		},
	})
	html := renderHTML(menu)

	if !strings.Contains(html, "First") {
		t.Errorf("expected First child, got: %s", html)
	}
	if !strings.Contains(html, "Second") {
		t.Errorf("expected Second child, got: %s", html)
	}
}

func TestDropdownMenu_CustomClass(t *testing.T) {
	menu := DropdownMenu(DropdownMenuProps{
		Open:     true,
		Class:    "menu-custom",
		Children: []core.Node{core.Text("Item")},
	})
	html := renderHTML(menu)

	if !strings.Contains(html, "menu-custom") {
		t.Errorf("expected custom class, got: %s", html)
	}
	// Should still have base classes
	if !strings.Contains(html, "absolute") {
		t.Errorf("expected base absolute class, got: %s", html)
	}
}

// DropdownItem tests

func TestDropdownItem_AriaRole(t *testing.T) {
	item := DropdownItem(DropdownItemProps{
		Children: []core.Node{core.Text("Edit")},
	})
	html := renderHTML(item)

	if !strings.Contains(html, `role="menuitem"`) {
		t.Errorf("expected role=\"menuitem\", got: %s", html)
	}
}

func TestDropdownItem_Renders(t *testing.T) {
	item := DropdownItem(DropdownItemProps{
		Children: []core.Node{core.Text("Edit")},
	})
	html := renderHTML(item)

	// Should be a button element
	if !strings.HasPrefix(html, "<button") {
		t.Errorf("expected <button> element, got: %s", html)
	}

	// Should have base classes
	if !strings.Contains(html, "w-full") {
		t.Errorf("expected w-full class, got: %s", html)
	}
	if !strings.Contains(html, "text-left") {
		t.Errorf("expected text-left class, got: %s", html)
	}
	if !strings.Contains(html, "px-4") {
		t.Errorf("expected px-4 class, got: %s", html)
	}
	if !strings.Contains(html, "py-2") {
		t.Errorf("expected py-2 class, got: %s", html)
	}

	// Should have type=button
	if !strings.Contains(html, `type="button"`) {
		t.Errorf("expected type=\"button\", got: %s", html)
	}

	// Should have default text styling
	if !strings.Contains(html, "text-gray-700") {
		t.Errorf("expected default text-gray-700 class, got: %s", html)
	}
}

func TestDropdownItem_Destructive(t *testing.T) {
	item := DropdownItem(DropdownItemProps{
		Destructive: true,
		Children:    []core.Node{core.Text("Delete")},
	})
	html := renderHTML(item)

	// Should have red styling
	if !strings.Contains(html, "text-red-600") {
		t.Errorf("expected text-red-600 class, got: %s", html)
	}
	if !strings.Contains(html, "hover:bg-red-50") {
		t.Errorf("expected hover:bg-red-50 class, got: %s", html)
	}
}

func TestDropdownItem_Disabled(t *testing.T) {
	item := DropdownItem(DropdownItemProps{
		Disabled: true,
		Children: []core.Node{core.Text("Unavailable")},
	})
	html := renderHTML(item)

	// Should have disabled attribute
	if !strings.Contains(html, `disabled="disabled"`) {
		t.Errorf("expected disabled attribute, got: %s", html)
	}

	// Should have disabled styling
	if !strings.Contains(html, "text-gray-400") {
		t.Errorf("expected text-gray-400 class, got: %s", html)
	}
	if !strings.Contains(html, "cursor-not-allowed") {
		t.Errorf("expected cursor-not-allowed class, got: %s", html)
	}
}

func TestDropdownItem_Children(t *testing.T) {
	item := DropdownItem(DropdownItemProps{
		Children: []core.Node{
			core.Span(core.Class("icon"), core.Text("*")),
			core.Text(" Edit"),
		},
	})
	html := renderHTML(item)

	if !strings.Contains(html, "<span") {
		t.Errorf("expected nested span, got: %s", html)
	}
	if !strings.Contains(html, "Edit") {
		t.Errorf("expected Edit text, got: %s", html)
	}
}

func TestDropdownItem_CustomClass(t *testing.T) {
	item := DropdownItem(DropdownItemProps{
		Class:    "item-custom",
		Children: []core.Node{core.Text("Edit")},
	})
	html := renderHTML(item)

	if !strings.Contains(html, "item-custom") {
		t.Errorf("expected custom class, got: %s", html)
	}
	// Should still have base classes
	if !strings.Contains(html, "w-full") {
		t.Errorf("expected base w-full class, got: %s", html)
	}
}
