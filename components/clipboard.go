//go:build js && wasm

package components

import "syscall/js"

// CopyToClipboard copies text to the clipboard
func CopyToClipboard(text string) bool {
	navigator := js.Global().Get("navigator")
	clipboard := navigator.Get("clipboard")

	if clipboard.IsUndefined() {
		// Fallback for older browsers
		return copyFallback(text)
	}

	// Use modern clipboard API
	clipboard.Call("writeText", text)
	return true
}

// copyFallback uses the old execCommand approach
func copyFallback(text string) bool {
	document := js.Global().Get("document")

	textarea := document.Call("createElement", "textarea")
	textarea.Set("value", text)
	textarea.Get("style").Set("position", "fixed")
	textarea.Get("style").Set("left", "-9999px")

	document.Get("body").Call("appendChild", textarea)
	textarea.Call("select")

	success := document.Call("execCommand", "copy").Bool()

	document.Get("body").Call("removeChild", textarea)
	return success
}

// CopyButtonProps configures a CopyButton
type CopyButtonProps struct {
	Text        string // text to copy
	Label       string // button label (default "Copy")
	CopiedLabel string // label after copy (default "Copied!")
	ShowToast   bool   // show toast notification
	OnCopy      func() // callback after copy
}

// CopyButton creates a button that copies text to clipboard
func CopyButton(props CopyButtonProps) js.Value {
	document := js.Global().Get("document")

	label := props.Label
	if label == "" {
		label = "Copy"
	}

	copiedLabel := props.CopiedLabel
	if copiedLabel == "" {
		copiedLabel = "Copied!"
	}

	btn := document.Call("createElement", "button")
	btn.Set("className", "inline-flex items-center gap-1 px-3 py-1.5 text-sm bg-gray-100 hover:bg-gray-200 text-gray-700 rounded transition-colors cursor-pointer")
	btn.Set("type", "button")

	// Icon
	icon := document.Call("createElement", "span")
	icon.Set("innerHTML", `<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"></path></svg>`)
	btn.Call("appendChild", icon)

	// Label
	labelSpan := document.Call("createElement", "span")
	labelSpan.Set("textContent", label)
	btn.Call("appendChild", labelSpan)

	btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		if CopyToClipboard(props.Text) {
			// Update label temporarily
			labelSpan.Set("textContent", copiedLabel)
			btn.Get("classList").Call("add", "bg-green-100", "text-green-700")
			btn.Get("classList").Call("remove", "bg-gray-100", "text-gray-700")

			// Show toast if enabled
			if props.ShowToast {
				Toast("Copied to clipboard!", ToastSuccess)
			}

			// Call callback
			if props.OnCopy != nil {
				props.OnCopy()
			}

			// Reset after 2 seconds
			js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
				labelSpan.Set("textContent", label)
				btn.Get("classList").Call("remove", "bg-green-100", "text-green-700")
				btn.Get("classList").Call("add", "bg-gray-100", "text-gray-700")
				return nil
			}), 2000)
		}
		return nil
	}))

	return btn
}

// CopyableText creates a text element with a copy button
func CopyableText(text string) js.Value {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "flex items-center gap-2 p-2 bg-gray-50 rounded border")

	textEl := document.Call("createElement", "code")
	textEl.Set("className", "flex-1 text-sm font-mono text-gray-800 truncate")
	textEl.Set("textContent", text)
	container.Call("appendChild", textEl)

	copyBtn := CopyButton(CopyButtonProps{
		Text:      text,
		Label:     "",
		ShowToast: true,
	})
	copyBtn.Set("className", "p-1.5 bg-gray-200 hover:bg-gray-300 rounded cursor-pointer")
	container.Call("appendChild", copyBtn)

	return container
}
