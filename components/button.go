//go:build js && wasm

package components

import "syscall/js"

// ButtonVariant defines button color variants
type ButtonVariant string

const (
	ButtonPrimary   ButtonVariant = "primary"
	ButtonSecondary ButtonVariant = "secondary"
	ButtonSuccess   ButtonVariant = "success"
	ButtonWarning   ButtonVariant = "warning"
	ButtonDanger    ButtonVariant = "danger"
	ButtonInfo      ButtonVariant = "info"
	ButtonGhost     ButtonVariant = "ghost"
)

var buttonVariantClasses = map[ButtonVariant]string{
	ButtonPrimary:   "bg-blue-500 text-white hover:bg-blue-600",
	ButtonSecondary: "bg-gray-200 text-gray-800 hover:bg-gray-300",
	ButtonSuccess:   "bg-green-500 text-white hover:bg-green-600",
	ButtonWarning:   "bg-yellow-500 text-white hover:bg-yellow-600",
	ButtonDanger:    "bg-red-500 text-white hover:bg-red-600",
	ButtonInfo:      "bg-cyan-500 text-white hover:bg-cyan-600",
	ButtonGhost:     "bg-transparent text-gray-600 hover:bg-gray-100",
}

// ButtonSize defines button sizes
type ButtonSize string

const (
	ButtonSM ButtonSize = "sm"
	ButtonMD ButtonSize = "md"
	ButtonLG ButtonSize = "lg"
)

var buttonSizeClasses = map[ButtonSize]string{
	ButtonSM: "px-2 py-1 text-sm",
	ButtonMD: "px-4 py-2",
	ButtonLG: "px-6 py-3 text-lg",
}

// ButtonProps configures a Button component
type ButtonProps struct {
	Text      string
	ClassName string
	Variant   ButtonVariant
	Size      ButtonSize
	OnClick   func()
}

// Button creates a styled button element
func Button(props ButtonProps) js.Value {
	document := js.Global().Get("document")
	btn := document.Call("createElement", "button")

	className := props.ClassName
	if className == "" {
		// Build class from variant and size
		variant := props.Variant
		if variant == "" {
			variant = ButtonPrimary
		}
		size := props.Size
		if size == "" {
			size = ButtonMD
		}
		className = buttonSizeClasses[size] + " " + buttonVariantClasses[variant] + " rounded cursor-pointer transition-colors"
	}

	btn.Set("className", className)
	btn.Set("textContent", props.Text)

	if props.OnClick != nil {
		btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			props.OnClick()
			return nil
		}))
	}

	return btn
}

// Convenience button constructors

// PrimaryButton creates a primary (blue) button
func PrimaryButton(text string, onClick func()) js.Value {
	return Button(ButtonProps{Text: text, Variant: ButtonPrimary, OnClick: onClick})
}

// SecondaryButton creates a secondary (gray) button
func SecondaryButton(text string, onClick func()) js.Value {
	return Button(ButtonProps{Text: text, Variant: ButtonSecondary, OnClick: onClick})
}

// SuccessButton creates a success (green) button
func SuccessButton(text string, onClick func()) js.Value {
	return Button(ButtonProps{Text: text, Variant: ButtonSuccess, OnClick: onClick})
}

// WarningButton creates a warning (yellow) button
func WarningButton(text string, onClick func()) js.Value {
	return Button(ButtonProps{Text: text, Variant: ButtonWarning, OnClick: onClick})
}

// DangerButton creates a danger (red) button
func DangerButton(text string, onClick func()) js.Value {
	return Button(ButtonProps{Text: text, Variant: ButtonDanger, OnClick: onClick})
}

// InfoButton creates an info (cyan) button
func InfoButton(text string, onClick func()) js.Value {
	return Button(ButtonProps{Text: text, Variant: ButtonInfo, OnClick: onClick})
}

// GhostButton creates a ghost (transparent) button
func GhostButton(text string, onClick func()) js.Value {
	return Button(ButtonProps{Text: text, Variant: ButtonGhost, OnClick: onClick})
}
