package ui

import (
	"strings"
	"testing"
)

func TestSplitButton_DefaultRendering(t *testing.T) {
	sb := SplitButton(SplitButtonProps{
		Label: "Save",
		Items: []SplitButtonItem{
			{Label: "Save as Draft"},
		},
	})
	html := renderHTML(sb)

	// Wrapper should have relative inline-flex
	if !strings.Contains(html, "relative") {
		t.Errorf("expected relative class, got: %s", html)
	}
	if !strings.Contains(html, "inline-flex") {
		t.Errorf("expected inline-flex class, got: %s", html)
	}

	// Primary button with default primary variant
	if !strings.Contains(html, "bg-blue-600") {
		t.Errorf("expected primary variant classes, got: %s", html)
	}

	// Primary button label
	if !strings.Contains(html, "Save") {
		t.Errorf("expected label text, got: %s", html)
	}

	// Should have two button elements (primary + chevron)
	if count := strings.Count(html, "<button"); count != 3 {
		// 2 for split button + 1 for the DropdownItem
		t.Errorf("expected 3 button elements (primary + chevron + 1 menu item), got %d in: %s", count, html)
	}

	// Primary button should have rounded-r-none
	if !strings.Contains(html, "rounded-r-none") {
		t.Errorf("expected rounded-r-none on primary button, got: %s", html)
	}

	// Chevron button should have rounded-l-none
	if !strings.Contains(html, "rounded-l-none") {
		t.Errorf("expected rounded-l-none on chevron button, got: %s", html)
	}

	// Chevron button should have border-l
	if !strings.Contains(html, "border-l") {
		t.Errorf("expected border-l on chevron button, got: %s", html)
	}

	// Should contain an SVG icon (chevron-down)
	if !strings.Contains(html, "<svg") {
		t.Errorf("expected SVG chevron icon, got: %s", html)
	}

	// Default position should be DropdownBottomRight
	if !strings.Contains(html, "right-0") {
		t.Errorf("expected default DropdownBottomRight position, got: %s", html)
	}
}

func TestSplitButton_AllVariants(t *testing.T) {
	tests := []struct {
		variant       ButtonVariant
		expectedClass string
		dividerClass  string
	}{
		{ButtonPrimary, "bg-blue-600", "border-blue-500"},
		{ButtonSecondary, "bg-gray-200", "border-gray-300"},
		{ButtonOutline, "border border-gray-300", "border-gray-300"},
		{ButtonGhost, "bg-transparent", "border-gray-300"},
		{ButtonDestructive, "bg-red-600", "border-red-500"},
	}

	for _, tt := range tests {
		t.Run(string(tt.variant), func(t *testing.T) {
			sb := SplitButton(SplitButtonProps{
				Label:   "Action",
				Variant: tt.variant,
				Items:   []SplitButtonItem{{Label: "Alt"}},
			})
			html := renderHTML(sb)

			if !strings.Contains(html, tt.expectedClass) {
				t.Errorf("variant %s: expected class %q in: %s", tt.variant, tt.expectedClass, html)
			}
			if !strings.Contains(html, tt.dividerClass) {
				t.Errorf("variant %s: expected divider class %q in: %s", tt.variant, tt.dividerClass, html)
			}
		})
	}
}

func TestSplitButton_AllSizes(t *testing.T) {
	tests := []struct {
		size               ButtonSize
		primarySizeClass   string
		chevronSizeClass   string
	}{
		{ButtonSM, "px-3 py-1.5 text-sm", "px-2 py-1.5 text-sm"},
		{ButtonMD, "px-4 py-2 text-base", "px-2.5 py-2 text-base"},
		{ButtonLG, "px-6 py-3 text-lg", "px-3 py-3 text-lg"},
	}

	for _, tt := range tests {
		t.Run(string(tt.size), func(t *testing.T) {
			sb := SplitButton(SplitButtonProps{
				Label: "Action",
				Size:  tt.size,
				Items: []SplitButtonItem{{Label: "Alt"}},
			})
			html := renderHTML(sb)

			// Primary should have standard size classes
			if !strings.Contains(html, tt.primarySizeClass) {
				t.Errorf("size %s: expected primary size class %q in: %s", tt.size, tt.primarySizeClass, html)
			}
			// Chevron should have compact size classes
			if !strings.Contains(html, tt.chevronSizeClass) {
				t.Errorf("size %s: expected chevron size class %q in: %s", tt.size, tt.chevronSizeClass, html)
			}
		})
	}
}

func TestSplitButton_Disabled(t *testing.T) {
	sb := SplitButton(SplitButtonProps{
		Label:    "Save",
		Disabled: true,
		Items:    []SplitButtonItem{{Label: "Alt"}},
	})
	html := renderHTML(sb)

	// Both buttons should have disabled attribute
	count := strings.Count(html, `disabled="disabled"`)
	if count < 2 {
		t.Errorf("expected at least 2 disabled attributes (primary + chevron), got %d in: %s", count, html)
	}

	// Should have disabled styling
	if !strings.Contains(html, "opacity-50") {
		t.Errorf("expected opacity-50 class, got: %s", html)
	}
	if !strings.Contains(html, "cursor-not-allowed") {
		t.Errorf("expected cursor-not-allowed class, got: %s", html)
	}
}

func TestSplitButton_CustomClass(t *testing.T) {
	sb := SplitButton(SplitButtonProps{
		Label: "Save",
		Class: "my-custom-wrapper",
		Items: []SplitButtonItem{{Label: "Alt"}},
	})
	html := renderHTML(sb)

	if !strings.Contains(html, "my-custom-wrapper") {
		t.Errorf("expected custom class, got: %s", html)
	}
	// Should still have base classes
	if !strings.Contains(html, "relative") {
		t.Errorf("expected base relative class, got: %s", html)
	}
	if !strings.Contains(html, "inline-flex") {
		t.Errorf("expected base inline-flex class, got: %s", html)
	}
}

func TestSplitButton_MenuItems(t *testing.T) {
	sb := SplitButton(SplitButtonProps{
		Label: "Save",
		Items: []SplitButtonItem{
			{Label: "Save as Draft"},
			{Label: "Save & Close"},
			{Label: "Discard"},
		},
	})
	html := renderHTML(sb)

	// All items should render with role="menuitem"
	menuItemCount := strings.Count(html, `role="menuitem"`)
	if menuItemCount != 3 {
		t.Errorf("expected 3 menuitems, got %d in: %s", menuItemCount, html)
	}

	// Each label should appear
	if !strings.Contains(html, "Save as Draft") {
		t.Errorf("expected 'Save as Draft' label, got: %s", html)
	}
	if !strings.Contains(html, "Save &amp; Close") || !strings.Contains(html, "Save &") {
		// HTML encoding may vary
		if !strings.Contains(html, "Save &") {
			t.Errorf("expected 'Save & Close' label, got: %s", html)
		}
	}
	if !strings.Contains(html, "Discard") {
		t.Errorf("expected 'Discard' label, got: %s", html)
	}
}

func TestSplitButton_OpenShowsMenu(t *testing.T) {
	sb := SplitButton(SplitButtonProps{
		Label: "Save",
		Open:  true,
		Items: []SplitButtonItem{{Label: "Alt"}},
	})
	html := renderHTML(sb)

	// Menu should be visible (no "hidden" class on the menu div)
	// The menu div has role="menu"
	menuIdx := strings.Index(html, `role="menu"`)
	if menuIdx == -1 {
		t.Fatalf("expected role=\"menu\" element, got: %s", html)
	}

	// Extract the menu div's class - it should NOT contain "hidden"
	// We check the class attribute before role="menu"
	menuSection := html[:menuIdx]
	lastDivIdx := strings.LastIndex(menuSection, "<div")
	if lastDivIdx == -1 {
		t.Fatalf("expected menu div, got: %s", html)
	}
	menuDivHTML := html[lastDivIdx : menuIdx+20]
	if strings.Contains(menuDivHTML, "hidden") {
		t.Errorf("expected menu to be visible when open, got: %s", menuDivHTML)
	}
}

func TestSplitButton_ClosedHidesMenu(t *testing.T) {
	sb := SplitButton(SplitButtonProps{
		Label: "Save",
		Open:  false,
		Items: []SplitButtonItem{{Label: "Alt"}},
	})
	html := renderHTML(sb)

	// Menu should have "hidden" class
	menuIdx := strings.Index(html, `role="menu"`)
	if menuIdx == -1 {
		t.Fatalf("expected role=\"menu\" element, got: %s", html)
	}

	menuSection := html[:menuIdx]
	lastDivIdx := strings.LastIndex(menuSection, "<div")
	if lastDivIdx == -1 {
		t.Fatalf("expected menu div, got: %s", html)
	}
	menuDivHTML := html[lastDivIdx : menuIdx+20]
	if !strings.Contains(menuDivHTML, "hidden") {
		t.Errorf("expected menu to be hidden when closed, got: %s", menuDivHTML)
	}
}

func TestSplitButton_BackdropWhenOpen(t *testing.T) {
	sb := SplitButton(SplitButtonProps{
		Label:   "Save",
		Open:    true,
		OnClose: func() {},
		Items:   []SplitButtonItem{{Label: "Alt"}},
	})
	html := renderHTML(sb)

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

func TestSplitButton_NoBackdropWhenClosed(t *testing.T) {
	sb := SplitButton(SplitButtonProps{
		Label:   "Save",
		Open:    false,
		OnClose: func() {},
		Items:   []SplitButtonItem{{Label: "Alt"}},
	})
	html := renderHTML(sb)

	if strings.Contains(html, "fixed inset-0") {
		t.Errorf("expected no backdrop when closed, got: %s", html)
	}
}

func TestSplitButton_AriaAttributes(t *testing.T) {
	tests := []struct {
		name     string
		open     bool
		expected string
	}{
		{"open", true, `aria-expanded="true"`},
		{"closed", false, `aria-expanded="false"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := SplitButton(SplitButtonProps{
				Label: "Save",
				Open:  tt.open,
				Items: []SplitButtonItem{{Label: "Alt"}},
			})
			html := renderHTML(sb)

			if !strings.Contains(html, `aria-haspopup="menu"`) {
				t.Errorf("expected aria-haspopup=\"menu\", got: %s", html)
			}
			if !strings.Contains(html, tt.expected) {
				t.Errorf("expected %s, got: %s", tt.expected, html)
			}
		})
	}
}

func TestSplitButton_Positions(t *testing.T) {
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
			sb := SplitButton(SplitButtonProps{
				Label:    "Save",
				Position: tt.position,
				Items:    []SplitButtonItem{{Label: "Alt"}},
			})
			html := renderHTML(sb)

			if !strings.Contains(html, tt.expectedClass) {
				t.Errorf("position %s: expected class %q in: %s", tt.position, tt.expectedClass, html)
			}
		})
	}
}

func TestSplitButton_EmptyItems(t *testing.T) {
	sb := SplitButton(SplitButtonProps{
		Label: "Save",
	})
	html := renderHTML(sb)

	// Should still render without items
	if !strings.Contains(html, "Save") {
		t.Errorf("expected label text, got: %s", html)
	}

	// Menu should exist but be empty (no menuitems)
	if !strings.Contains(html, `role="menu"`) {
		t.Errorf("expected menu element even with no items, got: %s", html)
	}
	if strings.Contains(html, `role="menuitem"`) {
		t.Errorf("expected no menuitems with empty items, got: %s", html)
	}
}

func TestSplitButton_MenuHasRoleMenu(t *testing.T) {
	sb := SplitButton(SplitButtonProps{
		Label: "Save",
		Open:  true,
		Items: []SplitButtonItem{{Label: "Alt"}},
	})
	html := renderHTML(sb)

	if !strings.Contains(html, `role="menu"`) {
		t.Errorf("expected role=\"menu\" on dropdown, got: %s", html)
	}
}

func TestSplitButton_MenuItemsAreButtons(t *testing.T) {
	sb := SplitButton(SplitButtonProps{
		Label: "Save",
		Open:  true,
		Items: []SplitButtonItem{
			{Label: "Option A"},
		},
	})
	html := renderHTML(sb)

	// DropdownItem renders as <button type="button" role="menuitem">
	if !strings.Contains(html, `type="button"`) {
		t.Errorf("expected menu items to have type=\"button\", got: %s", html)
	}
}

func TestSplitButton_PrimaryButtonType(t *testing.T) {
	sb := SplitButton(SplitButtonProps{
		Label: "Save",
		Items: []SplitButtonItem{{Label: "Alt"}},
	})
	html := renderHTML(sb)

	// Both primary and chevron should be type="button"
	typeCount := strings.Count(html, `type="button"`)
	if typeCount < 2 {
		t.Errorf("expected at least 2 type=\"button\" attributes (primary + chevron), got %d in: %s", typeCount, html)
	}
}

func TestSplitButton_ChevronIconSize(t *testing.T) {
	sb := SplitButton(SplitButtonProps{
		Label: "Save",
		Items: []SplitButtonItem{{Label: "Alt"}},
	})
	html := renderHTML(sb)

	// Chevron icon should use IconSM (w-4 h-4)
	if !strings.Contains(html, "w-4") {
		t.Errorf("expected w-4 (IconSM) on chevron icon, got: %s", html)
	}
	if !strings.Contains(html, "h-4") {
		t.Errorf("expected h-4 (IconSM) on chevron icon, got: %s", html)
	}
}

func TestSplitButton_StructuralOrder(t *testing.T) {
	sb := SplitButton(SplitButtonProps{
		Label: "Primary",
		Open:  true,
		Items: []SplitButtonItem{{Label: "Secondary"}},
	})
	html := renderHTML(sb)

	// Primary button should appear before chevron (SVG) and menu
	primaryIdx := strings.Index(html, "Primary")
	chevronIdx := strings.Index(html, "<svg")
	menuIdx := strings.Index(html, `role="menu"`)

	if primaryIdx == -1 || chevronIdx == -1 || menuIdx == -1 {
		t.Fatalf("missing expected content in: %s", html)
	}

	if primaryIdx > chevronIdx {
		t.Errorf("primary button should appear before chevron")
	}
	if chevronIdx > menuIdx {
		t.Errorf("chevron should appear before menu")
	}
}
