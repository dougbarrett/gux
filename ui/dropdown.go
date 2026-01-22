package ui

import "github.com/dougbarrett/gux/core"

// DropdownPosition defines the positioning of the dropdown menu.
type DropdownPosition string

const (
	DropdownBottomLeft  DropdownPosition = "bottom-left"
	DropdownBottomRight DropdownPosition = "bottom-right"
	DropdownTopLeft     DropdownPosition = "top-left"
	DropdownTopRight    DropdownPosition = "top-right"
)

// dropdownPositionClasses maps positions to their Tailwind positioning classes.
var dropdownPositionClasses = map[DropdownPosition]string{
	DropdownBottomLeft:  "top-full left-0 mt-1",
	DropdownBottomRight: "top-full right-0 mt-1",
	DropdownTopLeft:     "bottom-full left-0 mb-1",
	DropdownTopRight:    "bottom-full right-0 mb-1",
}

// DropdownProps configures a Dropdown component.
type DropdownProps struct {
	Open     bool             // Controls menu visibility
	OnClose  func()           // Called on backdrop click
	Position DropdownPosition // Menu position (default: DropdownBottomLeft)
	Class    string           // Additional wrapper classes
	Children []core.Node      // DropdownTrigger and DropdownMenu
}

// Dropdown creates a dropdown container with relative positioning.
//
// When Open is true, an invisible backdrop is rendered to capture
// outside clicks and close the menu. This approach handles "click outside
// to close" without requiring global event listeners.
//
// Example:
//
//	dropdownOpen := r.StateBool("dropdown", false)
//	Dropdown(DropdownProps{
//	    Open:    dropdownOpen.Get(),
//	    OnClose: func() { dropdownOpen.Set(false) },
//	    Children: []core.Node{
//	        DropdownTrigger(DropdownTriggerProps{
//	            Open:    dropdownOpen.Get(),
//	            OnClick: func() { dropdownOpen.Set(!dropdownOpen.Get()) },
//	            Children: []core.Node{Button(ButtonProps{Children: []core.Node{core.Text("Menu")}})},
//	        }),
//	        DropdownMenu(DropdownMenuProps{
//	            Open: dropdownOpen.Get(),
//	            Children: []core.Node{
//	                DropdownItem(DropdownItemProps{Children: []core.Node{core.Text("Option 1")}}),
//	                DropdownItem(DropdownItemProps{Children: []core.Node{core.Text("Option 2")}}),
//	            },
//	        }),
//	    },
//	})
func Dropdown(props DropdownProps) core.Node {
	class := MergeClasses("relative inline-block", props.Class)

	var children []core.Node

	// Invisible backdrop for outside click detection when open
	if props.Open && props.OnClose != nil {
		children = append(children,
			core.Div(core.Attrs{
				Class:   "fixed inset-0 z-40",
				OnClick: props.OnClose,
			}),
		)
	}

	children = append(children, props.Children...)

	return core.Div(core.Class(class), children...)
}

// DropdownTriggerProps configures a DropdownTrigger component.
type DropdownTriggerProps struct {
	Open     bool        // For aria-expanded state
	OnClick  func()      // Toggle handler
	Class    string      // Additional classes
	Children []core.Node // Usually a Button
}

// DropdownTrigger wraps the element that opens the dropdown.
//
// Provides ARIA attributes for accessibility:
// - aria-haspopup="menu" indicates this controls a menu
// - aria-expanded indicates whether the menu is open
//
// Example:
//
//	DropdownTrigger(DropdownTriggerProps{
//	    Open:    isOpen,
//	    OnClick: toggleOpen,
//	    Children: []core.Node{Button(ButtonProps{Children: []core.Node{core.Text("Actions")}})},
//	})
func DropdownTrigger(props DropdownTriggerProps) core.Node {
	class := MergeClasses("inline-block", props.Class)

	return core.Div(core.Attrs{
		Class:   class,
		OnClick: props.OnClick,
		Extra: map[string]string{
			"aria-haspopup":  "menu",
			"aria-expanded": boolToAttr(props.Open),
		},
	}, props.Children...)
}

// DropdownMenuProps configures a DropdownMenu component.
type DropdownMenuProps struct {
	Open     bool             // Visibility control
	Position DropdownPosition // For positioning classes
	Class    string           // Additional classes
	Children []core.Node      // DropdownItems
}

// DropdownMenu renders the menu container with role="menu".
//
// Hidden via CSS when Open is false. Position determines which corner
// the menu anchors to relative to the trigger.
//
// Example:
//
//	DropdownMenu(DropdownMenuProps{
//	    Open:     isOpen,
//	    Position: DropdownBottomRight,
//	    Children: []core.Node{
//	        DropdownItem(DropdownItemProps{Children: []core.Node{core.Text("Edit")}}),
//	        DropdownItem(DropdownItemProps{Destructive: true, Children: []core.Node{core.Text("Delete")}}),
//	    },
//	})
func DropdownMenu(props DropdownMenuProps) core.Node {
	// Default position
	position := props.Position
	if position == "" {
		position = DropdownBottomLeft
	}

	class := MergeClasses(
		"absolute z-50 min-w-48 bg-white dark:bg-gray-800 rounded-md shadow-lg border border-gray-200 dark:border-gray-700 py-1",
		dropdownPositionClasses[position],
		ConditionalClass(!props.Open, "hidden"),
		props.Class,
	)

	return core.Div(core.Attrs{
		Class: class,
		Extra: map[string]string{
			"role": "menu",
		},
	}, props.Children...)
}

// DropdownItemProps configures a DropdownItem component.
type DropdownItemProps struct {
	OnClick     func()      // Click handler
	Disabled    bool        // Disabled state
	Destructive bool        // Red styling for delete actions
	Class       string      // Additional classes
	Children    []core.Node // Item content
}

// DropdownItem renders a menu item button with role="menuitem".
//
// Supports destructive styling for dangerous actions (delete, remove) and
// disabled state for unavailable options.
//
// Example:
//
//	DropdownItem(DropdownItemProps{
//	    OnClick:  handleEdit,
//	    Children: []core.Node{core.Text("Edit")},
//	})
//
//	DropdownItem(DropdownItemProps{
//	    OnClick:     handleDelete,
//	    Destructive: true,
//	    Children:    []core.Node{core.Text("Delete")},
//	})
func DropdownItem(props DropdownItemProps) core.Node {
	// Base classes
	baseClass := "w-full text-left px-4 py-2 text-sm transition-colors"

	// Styling based on state
	var stateClass string
	if props.Disabled {
		stateClass = "text-gray-400 dark:text-gray-500 cursor-not-allowed"
	} else if props.Destructive {
		stateClass = "text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20"
	} else {
		stateClass = "text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700"
	}

	class := MergeClasses(baseClass, stateClass, props.Class)

	extra := map[string]string{
		"role": "menuitem",
		"type": "button",
	}

	// Add disabled attribute when disabled
	if props.Disabled {
		extra["disabled"] = "disabled"
	}

	attrs := core.Attrs{
		Class: class,
		Extra: extra,
	}

	// Only attach click handler if not disabled
	if !props.Disabled {
		attrs.OnClick = props.OnClick
	}

	return core.Button(attrs, props.Children...)
}
