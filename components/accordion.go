//go:build js && wasm

package components

import "syscall/js"

// AccordionItem represents a single accordion section
type AccordionItem struct {
	Title   string
	Content js.Value
	Open    bool // initially open
}

// AccordionProps configures an Accordion
type AccordionProps struct {
	Items       []AccordionItem
	AllowMultiple bool // allow multiple panels open at once
}

// Accordion creates a collapsible accordion component
type Accordion struct {
	element js.Value
	panels  []js.Value
}

// NewAccordion creates a new Accordion component
func NewAccordion(props AccordionProps) *Accordion {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "border border-gray-200 dark:border-gray-700 rounded-lg divide-y divide-gray-200 dark:divide-gray-700")

	acc := &Accordion{
		element: container,
		panels:  make([]js.Value, len(props.Items)),
	}

	for i, item := range props.Items {
		panel := acc.createPanel(item, i, props.AllowMultiple)
		container.Call("appendChild", panel)
		acc.panels[i] = panel
	}

	return acc
}

func (a *Accordion) createPanel(item AccordionItem, index int, allowMultiple bool) js.Value {
	document := js.Global().Get("document")

	panel := document.Call("createElement", "div")
	panel.Set("className", "")

	// Header button
	header := document.Call("createElement", "button")
	header.Set("className", "w-full px-4 py-3 flex items-center justify-between text-left hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:bg-gray-50 dark:focus:bg-gray-700 transition-colors cursor-pointer")
	header.Set("type", "button")

	title := document.Call("createElement", "span")
	title.Set("className", "font-medium text-gray-900 dark:text-gray-100")
	title.Set("textContent", item.Title)
	header.Call("appendChild", title)

	// Chevron icon
	chevron := document.Call("createElement", "span")
	chevron.Set("className", "transform transition-transform duration-200")
	chevron.Set("innerHTML", `<svg class="w-5 h-5 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>`)
	header.Call("appendChild", chevron)

	panel.Call("appendChild", header)

	// Content
	content := document.Call("createElement", "div")
	content.Set("className", "overflow-hidden transition-all duration-200")

	contentInner := document.Call("createElement", "div")
	contentInner.Set("className", "px-4 py-3 text-gray-600 dark:text-gray-300")
	contentInner.Call("appendChild", item.Content)
	content.Call("appendChild", contentInner)

	if item.Open {
		content.Get("style").Set("maxHeight", "1000px")
		chevron.Get("classList").Call("add", "rotate-180")
	} else {
		content.Get("style").Set("maxHeight", "0")
	}

	panel.Call("appendChild", content)

	// Toggle handler
	isOpen := item.Open
	header.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		isOpen = !isOpen

		if isOpen {
			// Close others if not allowing multiple
			if !allowMultiple {
				for i, p := range a.panels {
					if i != index {
						pContent := p.Get("children").Index(1)
						pChevron := p.Get("children").Index(0).Get("children").Index(1)
						pContent.Get("style").Set("maxHeight", "0")
						pChevron.Get("classList").Call("remove", "rotate-180")
					}
				}
			}
			content.Get("style").Set("maxHeight", "1000px")
			chevron.Get("classList").Call("add", "rotate-180")
		} else {
			content.Get("style").Set("maxHeight", "0")
			chevron.Get("classList").Call("remove", "rotate-180")
		}

		return nil
	}))

	return panel
}

// Element returns the DOM element
func (a *Accordion) Element() js.Value {
	return a.element
}

// OpenPanel opens a specific panel by index
func (a *Accordion) OpenPanel(index int) {
	if index < 0 || index >= len(a.panels) {
		return
	}
	panel := a.panels[index]
	content := panel.Get("children").Index(1)
	chevron := panel.Get("children").Index(0).Get("children").Index(1)
	content.Get("style").Set("maxHeight", "1000px")
	chevron.Get("classList").Call("add", "rotate-180")
}

// ClosePanel closes a specific panel by index
func (a *Accordion) ClosePanel(index int) {
	if index < 0 || index >= len(a.panels) {
		return
	}
	panel := a.panels[index]
	content := panel.Get("children").Index(1)
	chevron := panel.Get("children").Index(0).Get("children").Index(1)
	content.Get("style").Set("maxHeight", "0")
	chevron.Get("classList").Call("remove", "rotate-180")
}

// CloseAll closes all panels
func (a *Accordion) CloseAll() {
	for i := range a.panels {
		a.ClosePanel(i)
	}
}
