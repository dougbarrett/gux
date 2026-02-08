package ui

import "github.com/dougbarrett/gux/core"

// SplitButtonItem represents one option in the split button dropdown.
type SplitButtonItem struct {
	Label   string
	OnClick func()
}

// SplitButtonProps configures a SplitButton component.
type SplitButtonProps struct {
	Label    string            // Primary button text
	OnClick  func()            // Primary button click handler
	Variant  ButtonVariant     // Visual style (default: primary)
	Size     ButtonSize        // Size (default: md)
	Class    string            // Additional classes on outer wrapper
	Disabled bool              // Disables both primary and chevron
	Items    []SplitButtonItem // Dropdown menu items
	Open     bool              // Controls dropdown visibility (caller manages state)
	OnToggle func()            // Called when chevron is clicked (caller toggles Open)
	OnClose  func()            // Called on outside click (caller sets Open=false)
	Position DropdownPosition  // Menu position (default: DropdownBottomRight)
}

// splitButtonDividerClasses maps variants to their border-left color.
var splitButtonDividerClasses = map[ButtonVariant]string{
	ButtonPrimary:     "border-blue-500",
	ButtonSecondary:   "border-gray-300 dark:border-gray-600",
	ButtonOutline:     "border-gray-300 dark:border-gray-600",
	ButtonGhost:       "border-gray-300 dark:border-gray-600",
	ButtonDestructive: "border-red-500",
}

// splitButtonChevronSizeClasses maps sizes to compact padding for the chevron button.
var splitButtonChevronSizeClasses = map[ButtonSize]string{
	ButtonSM: "px-2 py-1.5 text-sm",
	ButtonMD: "px-2.5 py-2 text-base",
	ButtonLG: "px-3 py-3 text-lg",
}

// SplitButton creates a split button with a primary action and a dropdown of
// additional actions. The caller manages Open state via r.StateBool.
//
// Example:
//
//	open := r.StateBool("split", false)
//	SplitButton(SplitButtonProps{
//	    Label:   "Save",
//	    OnClick: func() { /* primary action */ },
//	    Items: []SplitButtonItem{
//	        {Label: "Save as Draft", OnClick: func() { /* ... */ }},
//	        {Label: "Save & Close", OnClick: func() { /* ... */ }},
//	    },
//	    Open:     open.Get(),
//	    OnToggle: func() { open.Set(!open.Get()) },
//	    OnClose:  func() { open.Set(false) },
//	})
func SplitButton(props SplitButtonProps) core.Node {
	variant := props.Variant
	if variant == "" {
		variant = ButtonPrimary
	}

	size := props.Size
	if size == "" {
		size = ButtonMD
	}

	position := props.Position
	if position == "" {
		position = DropdownBottomRight
	}

	// Primary button classes
	primaryClass := MergeClasses(
		buttonBaseClasses,
		buttonVariantClasses[variant],
		buttonSizeClasses[size],
		ConditionalClass(props.Disabled, buttonDisabledClasses),
		"rounded-r-none",
	)

	// Chevron button classes
	chevronClass := MergeClasses(
		buttonBaseClasses,
		buttonVariantClasses[variant],
		splitButtonChevronSizeClasses[size],
		ConditionalClass(props.Disabled, buttonDisabledClasses),
		"rounded-l-none border-l",
		splitButtonDividerClasses[variant],
	)

	// Build children
	var children []core.Node

	// Backdrop for outside click
	if props.Open && props.OnClose != nil {
		children = append(children,
			core.Div(core.Attrs{
				Class:   "fixed inset-0 z-40",
				OnClick: props.OnClose,
			}),
		)
	}

	// Primary button
	primaryAttrs := core.Attrs{
		Class:   primaryClass,
		Type:    "button",
		OnClick: props.OnClick,
	}
	if props.Disabled {
		primaryAttrs.Extra = map[string]string{
			"disabled": "disabled",
		}
	}
	children = append(children,
		core.El("button", primaryAttrs, core.Text(props.Label)),
	)

	// Chevron button
	chevronExtra := map[string]string{
		"aria-haspopup":  "menu",
		"aria-expanded": boolToAttr(props.Open),
	}
	if props.Disabled {
		chevronExtra["disabled"] = "disabled"
	}
	chevronAttrs := core.Attrs{
		Class:   chevronClass,
		Type:    "button",
		OnClick: props.OnToggle,
		Extra:   chevronExtra,
	}
	children = append(children,
		core.El("button", chevronAttrs, Icon(IconProps{Name: "chevron-down", Size: IconSM})),
	)

	// Dropdown menu with items
	var menuItems []core.Node
	for _, item := range props.Items {
		menuItems = append(menuItems,
			DropdownItem(DropdownItemProps{
				OnClick:  item.OnClick,
				Children: []core.Node{core.Text(item.Label)},
			}),
		)
	}
	children = append(children,
		DropdownMenu(DropdownMenuProps{
			Open:     props.Open,
			Position: position,
			Children: menuItems,
		}),
	)

	// Wrapper
	wrapperClass := MergeClasses("relative inline-flex", props.Class)
	return core.Div(core.Class(wrapperClass), children...)
}
