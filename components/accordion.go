//go:build js && wasm

package components

import (
	"strconv"
	"syscall/js"
)

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
	element  js.Value
	panels   []js.Value
	headers  []js.Value // store headers for ARIA updates
	baseID   string     // base ID for generating unique IDs
}

// NewAccordion creates a new Accordion component
func NewAccordion(props AccordionProps) *Accordion {
	document := js.Global().Get("document")

	// Generate unique base ID for this accordion
	crypto := js.Global().Get("crypto")
	uuid := crypto.Call("randomUUID").String()
	baseID := "accordion-" + uuid

	container := document.Call("createElement", "div")
	container.Set("className", "border border-subtle rounded-lg divide-y divide-gray-200 dark:divide-gray-700")

	acc := &Accordion{
		element: container,
		panels:  make([]js.Value, len(props.Items)),
		headers: make([]js.Value, len(props.Items)),
		baseID:  baseID,
	}

	for i, item := range props.Items {
		panel, header := acc.createPanel(item, i, props.AllowMultiple)
		container.Call("appendChild", panel)
		acc.panels[i] = panel
		acc.headers[i] = header
	}

	return acc
}

func (a *Accordion) createPanel(item AccordionItem, index int, allowMultiple bool) (js.Value, js.Value) {
	document := js.Global().Get("document")

	// Generate unique IDs for ARIA relationships
	triggerID := a.baseID + "-trigger-" + strconv.Itoa(index)
	panelID := a.baseID + "-panel-" + strconv.Itoa(index)

	panel := document.Call("createElement", "div")
	panel.Set("className", "")

	// Header button
	header := document.Call("createElement", "button")
	header.Set("className", "w-full px-4 py-3 flex items-center justify-between text-left hover:surface-raised focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500 transition-colors cursor-pointer")
	header.Set("type", "button")

	// ARIA attributes for trigger button
	header.Set("id", triggerID)
	header.Call("setAttribute", "aria-controls", panelID)
	if item.Open {
		header.Call("setAttribute", "aria-expanded", "true")
	} else {
		header.Call("setAttribute", "aria-expanded", "false")
	}

	title := document.Call("createElement", "span")
	title.Set("className", "font-medium text-primary")
	title.Set("textContent", item.Title)
	header.Call("appendChild", title)

	// Chevron icon
	chevron := document.Call("createElement", "span")
	chevron.Set("className", "transform transition-transform duration-200")
	chevron.Set("innerHTML", `<svg class="w-5 h-5 icon-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>`)
	header.Call("appendChild", chevron)

	panel.Call("appendChild", header)

	// Content
	content := document.Call("createElement", "div")
	content.Set("className", "overflow-hidden transition-all duration-200")

	// ARIA attributes for content panel
	content.Set("id", panelID)
	content.Call("setAttribute", "role", "region")
	content.Call("setAttribute", "aria-labelledby", triggerID)

	contentInner := document.Call("createElement", "div")
	contentInner.Set("className", "px-4 py-3 text-secondary")
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
						// Update aria-expanded on other headers
						a.headers[i].Call("setAttribute", "aria-expanded", "false")
					}
				}
			}
			content.Get("style").Set("maxHeight", "1000px")
			chevron.Get("classList").Call("add", "rotate-180")
			header.Call("setAttribute", "aria-expanded", "true")
		} else {
			content.Get("style").Set("maxHeight", "0")
			chevron.Get("classList").Call("remove", "rotate-180")
			header.Call("setAttribute", "aria-expanded", "false")
		}

		return nil
	}))

	return panel, header
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
	a.headers[index].Call("setAttribute", "aria-expanded", "true")
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
	a.headers[index].Call("setAttribute", "aria-expanded", "false")
}

// CloseAll closes all panels
func (a *Accordion) CloseAll() {
	for i := range a.panels {
		a.ClosePanel(i)
	}
}
