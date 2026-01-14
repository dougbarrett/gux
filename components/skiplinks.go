//go:build js && wasm

package components

import "syscall/js"

// SkipLink represents a skip navigation link
type SkipLink struct {
	Label  string
	Target string // CSS selector or element ID (without #)
}

// SkipLinksProps configures SkipLinks component
type SkipLinksProps struct {
	Links     []SkipLink
	ClassName string
}

// SkipLinks creates accessible skip navigation links
// These links are hidden until focused (via Tab key)
func SkipLinks(props SkipLinksProps) js.Value {
	document := js.Global().Get("document")

	// Default skip link if none provided
	if len(props.Links) == 0 {
		props.Links = []SkipLink{
			{Label: "Skip to main content", Target: "main"},
		}
	}

	container := document.Call("createElement", "div")
	container.Set("className", "skip-links")

	for _, link := range props.Links {
		a := document.Call("createElement", "a")
		a.Set("href", "#"+link.Target)
		a.Set("className", "sr-only focus:not-sr-only focus:absolute focus:top-0 focus:left-0 focus:z-50 focus:bg-blue-600 focus:text-white focus:px-4 focus:py-2 focus:rounded focus:m-2 focus:outline-none focus:ring-2 focus:ring-white")
		a.Set("textContent", link.Label)

		// Handle click to focus target
		target := link.Target
		a.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			args[0].Call("preventDefault")

			// Try to find target element
			targetEl := document.Call("getElementById", target)
			if targetEl.IsNull() {
				// Try as CSS selector
				targetEl = document.Call("querySelector", target)
			}

			if !targetEl.IsNull() && !targetEl.IsUndefined() {
				// Make target focusable if not already
				if targetEl.Get("tabIndex").Int() < 0 {
					targetEl.Set("tabIndex", -1)
				}
				targetEl.Call("focus")
				targetEl.Call("scrollIntoView", map[string]any{"behavior": "smooth"})
			}

			return nil
		}))

		container.Call("appendChild", a)
	}

	return container
}

// DefaultSkipLinks creates standard skip links for main, nav, and footer
func DefaultSkipLinks() js.Value {
	return SkipLinks(SkipLinksProps{
		Links: []SkipLink{
			{Label: "Skip to main content", Target: "main"},
			{Label: "Skip to navigation", Target: "nav"},
			{Label: "Skip to footer", Target: "footer"},
		},
	})
}

// MainSkipLink creates a single skip link to main content
func MainSkipLink() js.Value {
	return SkipLinks(SkipLinksProps{
		Links: []SkipLink{
			{Label: "Skip to main content", Target: "main"},
		},
	})
}

// AddSkipLinksCSS adds necessary CSS for skip links to the document
// Call this once during app initialization
func AddSkipLinksCSS() {
	document := js.Global().Get("document")

	// Check if already added
	existing := document.Call("getElementById", "skip-links-css")
	if !existing.IsNull() {
		return
	}

	style := document.Call("createElement", "style")
	style.Set("id", "skip-links-css")
	style.Set("textContent", `
		.sr-only {
			position: absolute;
			width: 1px;
			height: 1px;
			padding: 0;
			margin: -1px;
			overflow: hidden;
			clip: rect(0, 0, 0, 0);
			white-space: nowrap;
			border-width: 0;
		}
		.focus\:not-sr-only:focus {
			position: absolute;
			width: auto;
			height: auto;
			padding: 0;
			margin: 0;
			overflow: visible;
			clip: auto;
			white-space: normal;
		}
	`)

	head := document.Get("head")
	head.Call("appendChild", style)
}

// VisuallyHidden creates a visually hidden element (for screen readers)
func VisuallyHidden(content string) js.Value {
	document := js.Global().Get("document")
	span := document.Call("createElement", "span")
	span.Set("className", "sr-only")
	span.Set("textContent", content)
	return span
}

// AriaLive creates a live region for screen reader announcements
type AriaLive struct {
	element js.Value
}

// NewAriaLive creates a new ARIA live region
func NewAriaLive(politeness string) *AriaLive {
	if politeness == "" {
		politeness = "polite" // polite or assertive
	}

	document := js.Global().Get("document")
	el := document.Call("createElement", "div")
	el.Set("className", "sr-only")
	el.Set("aria-live", politeness)
	el.Set("aria-atomic", "true")

	// Append to body
	document.Get("body").Call("appendChild", el)

	return &AriaLive{element: el}
}

// Announce makes an announcement to screen readers
func (a *AriaLive) Announce(message string) {
	// Clear then set to trigger announcement
	a.element.Set("textContent", "")

	// Use setTimeout to ensure the change is detected
	js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
		a.element.Set("textContent", message)
		return nil
	}), 100)
}

// Clear clears the live region
func (a *AriaLive) Clear() {
	a.element.Set("textContent", "")
}

// Destroy removes the live region from the DOM
func (a *AriaLive) Destroy() {
	a.element.Call("remove")
}

// Global announcer instance
var globalAnnouncer *AriaLive

// Announce makes a screen reader announcement using the global announcer
func Announce(message string) {
	if globalAnnouncer == nil {
		globalAnnouncer = NewAriaLive("polite")
	}
	globalAnnouncer.Announce(message)
}

// AnnounceAssertive makes an urgent screen reader announcement
func AnnounceAssertive(message string) {
	announcer := NewAriaLive("assertive")
	announcer.Announce(message)

	// Clean up after announcement
	js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
		announcer.Destroy()
		return nil
	}), 5000)
}
