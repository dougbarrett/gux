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
	BadgeDefault: "bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200",
	BadgePrimary: "bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200",
	BadgeSuccess: "bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200",
	BadgeWarning: "bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200",
	BadgeError:   "bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200",
	BadgeInfo:    "bg-cyan-100 dark:bg-cyan-900 text-cyan-800 dark:text-cyan-200",
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
