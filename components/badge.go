//go:build js && wasm

package components

import "syscall/js"

// BadgeVariant defines badge styling variants
type BadgeVariant string

const (
	BadgeDefault BadgeVariant = ""
	BadgePrimary BadgeVariant = "primary"
	BadgeSuccess BadgeVariant = "success"
	BadgeWarning BadgeVariant = "warning"
	BadgeError   BadgeVariant = "error"
	BadgeInfo    BadgeVariant = "info"
)

var badgeStyles = map[BadgeVariant]string{
	BadgeDefault: "bg-gray-100 text-gray-800",
	BadgePrimary: "bg-blue-100 text-blue-800",
	BadgeSuccess: "bg-green-100 text-green-800",
	BadgeWarning: "bg-yellow-100 text-yellow-800",
	BadgeError:   "bg-red-100 text-red-800",
	BadgeInfo:    "bg-cyan-100 text-cyan-800",
}

// BadgeProps configures a Badge component
type BadgeProps struct {
	Text      string
	Variant   BadgeVariant
	ClassName string
	Rounded   bool // Pill style
}

// Badge creates a small status label/tag
func Badge(props BadgeProps) js.Value {
	document := js.Global().Get("document")

	badge := document.Call("createElement", "span")

	variant := props.Variant
	if variant == "" {
		variant = BadgeDefault
	}

	variantClass := badgeStyles[variant]
	if variantClass == "" {
		variantClass = badgeStyles[BadgeDefault]
	}

	className := "inline-flex items-center px-2.5 py-0.5 text-xs font-medium"
	if props.Rounded {
		className += " rounded-full"
	} else {
		className += " rounded"
	}
	className += " " + variantClass

	if props.ClassName != "" {
		className = props.ClassName
	}

	badge.Set("className", className)
	badge.Set("textContent", props.Text)

	return badge
}

// Quick badge creators
func BadgeText(text string) js.Value {
	return Badge(BadgeProps{Text: text})
}

func BadgePrimaryText(text string) js.Value {
	return Badge(BadgeProps{Text: text, Variant: BadgePrimary})
}

func BadgeSuccessText(text string) js.Value {
	return Badge(BadgeProps{Text: text, Variant: BadgeSuccess})
}

func BadgeWarningText(text string) js.Value {
	return Badge(BadgeProps{Text: text, Variant: BadgeWarning})
}

func BadgeErrorText(text string) js.Value {
	return Badge(BadgeProps{Text: text, Variant: BadgeError})
}

func BadgeInfoText(text string) js.Value {
	return Badge(BadgeProps{Text: text, Variant: BadgeInfo})
}
