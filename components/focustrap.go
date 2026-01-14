//go:build js && wasm

package components

import "syscall/js"

// FocusTrap manages focus within a container element
type FocusTrap struct {
	container     js.Value
	active        bool
	keyHandler    js.Func
	previousFocus js.Value
}

// NewFocusTrap creates a focus trap for the given container
func NewFocusTrap(container js.Value) *FocusTrap {
	ft := &FocusTrap{
		container: container,
	}
	ft.setupKeyHandler()
	return ft
}

func (ft *FocusTrap) setupKeyHandler() {
	ft.keyHandler = js.FuncOf(func(this js.Value, args []js.Value) any {
		if !ft.active {
			return nil
		}

		event := args[0]
		if event.Get("key").String() != "Tab" {
			return nil
		}

		focusable := ft.getFocusableElements()
		if focusable.Length() == 0 {
			event.Call("preventDefault")
			return nil
		}

		firstEl := focusable.Index(0)
		lastEl := focusable.Index(focusable.Length() - 1)

		document := js.Global().Get("document")
		activeEl := document.Get("activeElement")

		shiftKey := event.Get("shiftKey").Bool()

		if shiftKey {
			// Shift+Tab - going backwards
			if activeEl.Call("isEqualNode", firstEl).Bool() {
				event.Call("preventDefault")
				lastEl.Call("focus")
			}
		} else {
			// Tab - going forwards
			if activeEl.Call("isEqualNode", lastEl).Bool() {
				event.Call("preventDefault")
				firstEl.Call("focus")
			}
		}

		return nil
	})
}

func (ft *FocusTrap) getFocusableElements() js.Value {
	selector := `
		a[href]:not([disabled]):not([tabindex="-1"]),
		button:not([disabled]):not([tabindex="-1"]),
		textarea:not([disabled]):not([tabindex="-1"]),
		input:not([disabled]):not([tabindex="-1"]):not([type="hidden"]),
		select:not([disabled]):not([tabindex="-1"]),
		[tabindex]:not([tabindex="-1"]):not([disabled])
	`
	return ft.container.Call("querySelectorAll", selector)
}

// Activate activates the focus trap
func (ft *FocusTrap) Activate() {
	if ft.active {
		return
	}

	document := js.Global().Get("document")

	// Store current focus
	ft.previousFocus = document.Get("activeElement")

	// Add key handler
	document.Call("addEventListener", "keydown", ft.keyHandler)

	ft.active = true

	// Focus first focusable element
	focusable := ft.getFocusableElements()
	if focusable.Length() > 0 {
		focusable.Index(0).Call("focus")
	}
}

// Deactivate deactivates the focus trap
func (ft *FocusTrap) Deactivate() {
	if !ft.active {
		return
	}

	document := js.Global().Get("document")
	document.Call("removeEventListener", "keydown", ft.keyHandler)

	ft.active = false

	// Restore previous focus
	if !ft.previousFocus.IsUndefined() && !ft.previousFocus.IsNull() {
		ft.previousFocus.Call("focus")
	}
}

// IsActive returns whether the focus trap is active
func (ft *FocusTrap) IsActive() bool {
	return ft.active
}

// Destroy cleans up the focus trap
func (ft *FocusTrap) Destroy() {
	ft.Deactivate()
	ft.keyHandler.Release()
}

// FocusFirst focuses the first focusable element in the container
func (ft *FocusTrap) FocusFirst() {
	focusable := ft.getFocusableElements()
	if focusable.Length() > 0 {
		focusable.Index(0).Call("focus")
	}
}

// FocusLast focuses the last focusable element in the container
func (ft *FocusTrap) FocusLast() {
	focusable := ft.getFocusableElements()
	if focusable.Length() > 0 {
		focusable.Index(focusable.Length() - 1).Call("focus")
	}
}

// TrapFocus is a convenience function that creates and activates a focus trap
func TrapFocus(container js.Value) *FocusTrap {
	ft := NewFocusTrap(container)
	ft.Activate()
	return ft
}
