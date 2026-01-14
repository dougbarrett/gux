//go:build js && wasm

package components

import "syscall/js"

// AlertVariant defines alert styling variants
type AlertVariant string

const (
	AlertInfo    AlertVariant = "info"
	AlertSuccess AlertVariant = "success"
	AlertWarning AlertVariant = "warning"
	AlertError   AlertVariant = "error"
)

var alertStyles = map[AlertVariant]struct {
	bg     string
	border string
	text   string
	icon   string
}{
	AlertInfo:    {bg: "bg-blue-50", border: "border-blue-200", text: "text-blue-800", icon: "ℹ️"},
	AlertSuccess: {bg: "bg-green-50", border: "border-green-200", text: "text-green-800", icon: "✓"},
	AlertWarning: {bg: "bg-yellow-50", border: "border-yellow-200", text: "text-yellow-800", icon: "⚠️"},
	AlertError:   {bg: "bg-red-50", border: "border-red-200", text: "text-red-800", icon: "✕"},
}

// AlertProps configures an Alert component
type AlertProps struct {
	Variant     AlertVariant
	Title       string
	Message     string
	Dismissible bool
	OnDismiss   func()
}

// Alert creates an alert message component
type Alert struct {
	element js.Value
}

// NewAlert creates a new Alert component
func NewAlert(props AlertProps) *Alert {
	document := js.Global().Get("document")

	variant := props.Variant
	if variant == "" {
		variant = AlertInfo
	}

	style := alertStyles[variant]

	alert := document.Call("createElement", "div")
	alert.Set("className", style.bg+" "+style.border+" "+style.text+" border rounded-lg p-4 mb-4")

	// Content wrapper
	content := document.Call("createElement", "div")
	content.Set("className", "flex items-start")

	// Icon
	icon := document.Call("createElement", "span")
	icon.Set("className", "mr-3 text-lg")
	icon.Set("textContent", style.icon)
	content.Call("appendChild", icon)

	// Text container
	textContainer := document.Call("createElement", "div")
	textContainer.Set("className", "flex-1")

	// Title
	if props.Title != "" {
		title := document.Call("createElement", "h4")
		title.Set("className", "font-semibold mb-1")
		title.Set("textContent", props.Title)
		textContainer.Call("appendChild", title)
	}

	// Message
	if props.Message != "" {
		message := document.Call("createElement", "p")
		message.Set("className", "text-sm")
		message.Set("textContent", props.Message)
		textContainer.Call("appendChild", message)
	}

	content.Call("appendChild", textContainer)

	// Dismiss button
	if props.Dismissible {
		dismiss := document.Call("createElement", "button")
		dismiss.Set("className", "ml-4 text-lg opacity-50 hover:opacity-100 cursor-pointer")
		dismiss.Set("textContent", "×")
		dismiss.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			alert.Get("parentNode").Call("removeChild", alert)
			if props.OnDismiss != nil {
				props.OnDismiss()
			}
			return nil
		}))
		content.Call("appendChild", dismiss)
	}

	alert.Call("appendChild", content)

	return &Alert{element: alert}
}

// Element returns the DOM element
func (a *Alert) Element() js.Value {
	return a.element
}

// Quick alert creators
func AlertInfoMsg(message string) js.Value {
	return NewAlert(AlertProps{Variant: AlertInfo, Message: message}).Element()
}

func AlertSuccessMsg(message string) js.Value {
	return NewAlert(AlertProps{Variant: AlertSuccess, Message: message}).Element()
}

func AlertWarningMsg(message string) js.Value {
	return NewAlert(AlertProps{Variant: AlertWarning, Message: message}).Element()
}

func AlertErrorMsg(message string) js.Value {
	return NewAlert(AlertProps{Variant: AlertError, Message: message}).Element()
}
