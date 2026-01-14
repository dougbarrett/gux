//go:build js && wasm

package components

import "syscall/js"

// LoadTailwind injects Tailwind CSS into the document head.
// This blocks until Tailwind is fully loaded to ensure styles are applied.
func LoadTailwind() {
	document := js.Global().Get("document")
	head := document.Get("head")

	script := document.Call("createElement", "script")
	script.Set("src", "https://cdn.tailwindcss.com")

	// Block until loaded
	done := make(chan struct{})
	script.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) any {
		// Configure Tailwind for class-based dark mode after it loads
		js.Global().Get("tailwind").Get("config").Set("darkMode", "class")

		// Inject custom CSS for mobile utilities
		injectMobileStyles()

		close(done)
		return nil
	}))

	head.Call("appendChild", script)
	<-done
}

// injectMobileStyles adds custom CSS utilities for mobile responsiveness
func injectMobileStyles() {
	document := js.Global().Get("document")
	head := document.Get("head")

	style := document.Call("createElement", "style")
	style.Set("textContent", `
		/* Hide scrollbar but keep scroll functionality */
		.scrollbar-hide {
			-ms-overflow-style: none;
			scrollbar-width: none;
		}
		.scrollbar-hide::-webkit-scrollbar {
			display: none;
		}

		/* Smooth touch scrolling on mobile */
		.overflow-x-auto {
			-webkit-overflow-scrolling: touch;
		}

		/* Prevent text selection on navigation items */
		nav a, nav button {
			-webkit-user-select: none;
			user-select: none;
		}

		/* Better touch targets on mobile */
		@media (max-width: 768px) {
			button, a, input, select, textarea {
				min-height: 44px;
			}
		}
	`)
	head.Call("appendChild", style)
}
