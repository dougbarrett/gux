//go:build js && wasm

package components

import (
	"syscall/js"
	"time"
)

// ToastVariant defines toast styling variants
type ToastVariant string

const (
	ToastInfo    ToastVariant = "info"
	ToastSuccess ToastVariant = "success"
	ToastWarning ToastVariant = "warning"
	ToastError   ToastVariant = "error"
)

var toastStyles = map[ToastVariant]struct {
	bg   string
	text string
	icon string
}{
	ToastInfo:    {bg: "bg-blue-600", text: "text-white", icon: "ℹ"},
	ToastSuccess: {bg: "bg-green-600", text: "text-white", icon: "✓"},
	ToastWarning: {bg: "bg-yellow-500", text: "text-white", icon: "⚠"},
	ToastError:   {bg: "bg-red-600", text: "text-white", icon: "✕"},
}

// ToastManager manages toast notifications
type ToastManager struct {
	container js.Value
}

var globalToastManager *ToastManager

// InitToasts initializes the global toast manager (call once on app startup)
func InitToasts() *ToastManager {
	if globalToastManager != nil {
		return globalToastManager
	}

	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("id", "toast-container")
	container.Set("className", "fixed top-4 right-4 z-50 flex flex-col gap-2")
	// ARIA live region for toast notifications
	container.Call("setAttribute", "role", "status")
	container.Call("setAttribute", "aria-live", "polite")
	container.Call("setAttribute", "aria-atomic", "true")

	document.Get("body").Call("appendChild", container)

	globalToastManager = &ToastManager{container: container}
	return globalToastManager
}

// ToastProps configures a toast notification
type ToastProps struct {
	Variant  ToastVariant
	Message  string
	Duration time.Duration // 0 = no auto-dismiss
}

// Show displays a toast notification
func (tm *ToastManager) Show(props ToastProps) {
	document := js.Global().Get("document")

	variant := props.Variant
	if variant == "" {
		variant = ToastInfo
	}

	style := toastStyles[variant]
	duration := props.Duration
	if duration == 0 {
		duration = 3 * time.Second
	}

	toast := document.Call("createElement", "div")
	toast.Set("className", style.bg+" "+style.text+" px-4 py-3 rounded-lg shadow-lg flex items-center gap-3 min-w-64 transform transition-all duration-300 translate-x-full opacity-0")

	// Icon (decorative)
	icon := document.Call("createElement", "span")
	icon.Set("className", "text-lg")
	icon.Set("textContent", style.icon)
	icon.Call("setAttribute", "aria-hidden", "true")
	toast.Call("appendChild", icon)

	// Message
	message := document.Call("createElement", "span")
	message.Set("className", "flex-1")
	message.Set("textContent", props.Message)
	toast.Call("appendChild", message)

	// Close button
	closeBtn := document.Call("createElement", "button")
	closeBtn.Set("className", "opacity-70 hover:opacity-100 cursor-pointer text-lg")
	closeBtn.Set("textContent", "×")
	closeBtn.Call("setAttribute", "aria-label", "Dismiss notification")

	var removeToast js.Func
	removeToast = js.FuncOf(func(this js.Value, args []js.Value) any {
		toast.Get("classList").Call("add", "translate-x-full", "opacity-0")
		go func() {
			time.Sleep(300 * time.Millisecond)
			if toast.Get("parentNode").Truthy() {
				tm.container.Call("removeChild", toast)
			}
		}()
		return nil
	})

	closeBtn.Call("addEventListener", "click", removeToast)
	toast.Call("appendChild", closeBtn)

	tm.container.Call("appendChild", toast)

	// Animate in
	go func() {
		time.Sleep(10 * time.Millisecond)
		toast.Get("classList").Call("remove", "translate-x-full", "opacity-0")

		// Auto-dismiss
		if duration > 0 {
			time.Sleep(duration)
			removeToast.Invoke()
		}
	}()
}

// Global toast functions for convenience

// Toast shows a toast with the global manager
func Toast(message string, variant ToastVariant) {
	if globalToastManager == nil {
		InitToasts()
	}
	globalToastManager.Show(ToastProps{Message: message, Variant: variant})
}

// ShowInfo shows an info toast
func ShowInfo(message string) {
	Toast(message, ToastInfo)
}

// ShowSuccess shows a success toast
func ShowSuccess(message string) {
	Toast(message, ToastSuccess)
}

// ShowWarning shows a warning toast
func ShowWarning(message string) {
	Toast(message, ToastWarning)
}

// ShowError shows an error toast
func ShowError(message string) {
	Toast(message, ToastError)
}

// ToastWithDuration shows a toast with custom duration
func ToastWithDuration(message string, variant ToastVariant, duration time.Duration) {
	if globalToastManager == nil {
		InitToasts()
	}
	globalToastManager.Show(ToastProps{Message: message, Variant: variant, Duration: duration})
}
