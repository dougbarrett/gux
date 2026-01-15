//go:build js && wasm

package components

import (
	"strings"
	"syscall/js"
)

// BreadcrumbItem represents a single breadcrumb
type BreadcrumbItem struct {
	Label string
	Path  string // empty for current page
	Icon  string // optional icon
}

// BreadcrumbsProps configures Breadcrumbs
type BreadcrumbsProps struct {
	Items     []BreadcrumbItem
	Separator string // default "/"
}

// Breadcrumbs creates a breadcrumb navigation component
func Breadcrumbs(props BreadcrumbsProps) js.Value {
	document := js.Global().Get("document")

	nav := document.Call("createElement", "nav")
	nav.Set("className", "text-sm")
	nav.Call("setAttribute", "aria-label", "Breadcrumb")

	// Use ordered list for semantic structure (per WAI-ARIA best practices)
	ol := document.Call("createElement", "ol")
	ol.Set("className", "flex items-center space-x-2")

	separator := props.Separator
	if separator == "" {
		separator = "/"
	}

	for i, item := range props.Items {
		// List item wrapper
		li := document.Call("createElement", "li")
		li.Set("className", "flex items-center")

		// Separator (except before first item)
		if i > 0 {
			sep := document.Call("createElement", "span")
			sep.Set("className", "mr-2 text-gray-400 dark:text-gray-500")
			sep.Set("textContent", separator)
			// Separators are decorative - hide from screen readers
			sep.Call("setAttribute", "aria-hidden", "true")
			li.Call("appendChild", sep)
		}

		// Item container
		itemEl := document.Call("createElement", "span")
		itemEl.Set("className", "flex items-center")

		// Icon if present
		if item.Icon != "" {
			icon := document.Call("createElement", "span")
			icon.Set("className", "mr-1")
			icon.Set("textContent", item.Icon)
			itemEl.Call("appendChild", icon)
		}

		// Link or text
		isLast := i == len(props.Items)-1
		if item.Path != "" && !isLast {
			link := document.Call("createElement", "a")
			link.Set("href", item.Path)
			link.Set("className", "text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 hover:underline")
			link.Set("textContent", item.Label)

			// Handle click for SPA routing
			link.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
				args[0].Call("preventDefault")
				if globalRouter != nil {
					globalRouter.Navigate(item.Path)
				}
				return nil
			}))

			itemEl.Call("appendChild", link)
		} else {
			text := document.Call("createElement", "span")
			if isLast {
				text.Set("className", "text-gray-500 dark:text-gray-400 font-medium")
				text.Set("aria-current", "page")
			} else {
				text.Set("className", "text-gray-600 dark:text-gray-400")
			}
			text.Set("textContent", item.Label)
			itemEl.Call("appendChild", text)
		}

		li.Call("appendChild", itemEl)
		ol.Call("appendChild", li)
	}

	nav.Call("appendChild", ol)
	return nav
}

// SimpleBreadcrumbs creates breadcrumbs from path strings
func SimpleBreadcrumbs(paths ...string) js.Value {
	items := make([]BreadcrumbItem, len(paths))
	currentPath := ""

	for i, label := range paths {
		if i == 0 {
			currentPath = "/"
		} else {
			currentPath += "/" + strings.ToLower(strings.ReplaceAll(label, " ", "-"))
		}

		items[i] = BreadcrumbItem{
			Label: label,
			Path:  currentPath,
		}
	}

	// Last item has no path (current page)
	if len(items) > 0 {
		items[len(items)-1].Path = ""
	}

	return Breadcrumbs(BreadcrumbsProps{Items: items})
}
