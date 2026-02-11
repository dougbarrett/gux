package ui

import "github.com/dougbarrett/gux/core"

// ButtonVariant defines the visual style of a button.
type ButtonVariant string

const (
	ButtonPrimary     ButtonVariant = "primary"
	ButtonSecondary   ButtonVariant = "secondary"
	ButtonOutline     ButtonVariant = "outline"
	ButtonGhost       ButtonVariant = "ghost"
	ButtonDestructive ButtonVariant = "destructive"
)

// ButtonSize defines the size of a button.
type ButtonSize string

const (
	ButtonSM ButtonSize = "sm"
	ButtonMD ButtonSize = "md"
	ButtonLG ButtonSize = "lg"
)

// ButtonProps configures a Button component.
type ButtonProps struct {
	Variant  ButtonVariant // Visual style (default: primary)
	Size     ButtonSize    // Size (default: md)
	Class    string        // Additional classes (override defaults)
	Disabled bool          // Disabled state
	Type     string        // Button type: "button", "submit", "reset" (default: "button")
	OnClick  func()        // Click handler (WASM only)
	Children []core.Node   // Button content

	// Alpine.js props (used when in Alpine mode)
	XOnClick  string // x-on:click expression (e.g., "login()")
	XDisabled string // x-bind:disabled expression (e.g., "loading")
}

// buttonVariantClasses maps variants to their Tailwind classes.
var buttonVariantClasses = map[ButtonVariant]string{
	ButtonPrimary:     "bg-blue-600 text-white hover:bg-blue-700",
	ButtonSecondary:   "bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200 hover:bg-gray-300 dark:hover:bg-gray-600",
	ButtonOutline:     "border border-gray-300 dark:border-gray-600 bg-transparent hover:bg-gray-100 dark:hover:bg-gray-800",
	ButtonGhost:       "bg-transparent hover:bg-gray-100 dark:hover:bg-gray-800",
	ButtonDestructive: "bg-red-600 text-white hover:bg-red-700",
}

// buttonSizeClasses maps sizes to their Tailwind classes.
var buttonSizeClasses = map[ButtonSize]string{
	ButtonSM: "px-3 py-1.5 text-sm",
	ButtonMD: "px-4 py-2 text-base",
	ButtonLG: "px-6 py-3 text-lg",
}

// buttonBaseClasses are the common classes applied to all buttons.
const buttonBaseClasses = "font-medium rounded-lg transition-colors cursor-pointer inline-flex items-center justify-center whitespace-nowrap"

// buttonDisabledClasses are applied when button is disabled.
const buttonDisabledClasses = "opacity-50 cursor-not-allowed"

// Button creates a button component with variants and sizes.
//
// Example:
//
//	Button(ButtonProps{
//	    Variant: ButtonPrimary,
//	    Size: ButtonMD,
//	    Children: []core.Node{core.Text("Click me")},
//	})
func Button(props ButtonProps) core.Node {
	// Apply defaults
	variant := props.Variant
	if variant == "" {
		variant = ButtonPrimary
	}

	size := props.Size
	if size == "" {
		size = ButtonMD
	}

	buttonType := props.Type
	if buttonType == "" {
		buttonType = "button"
	}

	// Build class string
	class := MergeClasses(
		buttonBaseClasses,
		buttonVariantClasses[variant],
		buttonSizeClasses[size],
		ConditionalClass(props.Disabled, buttonDisabledClasses),
		props.Class,
	)

	// Build attributes
	attrs := core.Attrs{
		Class:   class,
		Type:    buttonType,
		OnClick: props.OnClick,
	}

	// Alpine.js directives
	if props.XOnClick != "" {
		attrs.XOn = map[string]string{"click": props.XOnClick}
	}
	if props.XDisabled != "" {
		attrs.XBind = map[string]string{"disabled": props.XDisabled}
	}

	// Add disabled attribute when disabled
	if props.Disabled {
		if attrs.Extra == nil {
			attrs.Extra = make(map[string]string)
		}
		attrs.Extra["disabled"] = "disabled"
	}

	return core.El("button", attrs, props.Children...)
}
