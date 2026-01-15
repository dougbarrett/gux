//go:build js && wasm

package components

import (
	"strings"
	"syscall/js"
)

// Command represents a command in the palette
type Command struct {
	ID          string
	Label       string // Display text
	Description string // Optional subtitle
	Icon        string // Optional emoji/icon
	Category    string // For grouping (e.g., "Navigation", "Actions")
	OnExecute   func()
	Shortcut    string // Optional keyboard hint (e.g., "Ctrl+B")
}

// CommandPaletteProps configures a CommandPalette component
type CommandPaletteProps struct {
	Commands     []Command
	Placeholder  string // Search input placeholder
	EmptyMessage string // "No commands found"
	OnClose      func() // Called when palette closes
}

// CommandPalette creates a command palette with Cmd+K trigger
type CommandPalette struct {
	overlay          js.Value
	container        js.Value
	input            js.Value
	resultsList      js.Value
	commands         []Command
	filteredCommands []Command
	query            string
	isOpen           bool
	highlightIdx     int
	props            CommandPaletteProps
	focusTrap        *FocusTrap
	keyboardListener js.Func
	listboxID        string // ARIA: unique ID for listbox
	optionIDs        []string // ARIA: generated IDs for each option
}

// NewCommandPalette creates a new CommandPalette component
func NewCommandPalette(props CommandPaletteProps) *CommandPalette {
	document := js.Global().Get("document")

	if props.EmptyMessage == "" {
		props.EmptyMessage = "No commands found"
	}
	if props.Placeholder == "" {
		props.Placeholder = "Search commands..."
	}

	// Generate unique IDs for ARIA
	listboxID := "cmdpalette-listbox-" + js.Global().Get("crypto").Call("randomUUID").String()

	cp := &CommandPalette{
		commands:         props.Commands,
		filteredCommands: props.Commands,
		highlightIdx:     -1,
		props:            props,
		listboxID:        listboxID,
	}

	// Overlay - full screen backdrop
	overlay := document.Call("createElement", "div")
	overlay.Set("className", "fixed inset-0 bg-black/50 z-50 flex items-start justify-center pt-[20vh] hidden")
	cp.overlay = overlay

	// Palette container with dialog role
	container := document.Call("createElement", "div")
	container.Set("className", "bg-white dark:bg-gray-800 rounded-lg shadow-2xl max-w-lg w-full mx-4 overflow-hidden")
	container.Call("setAttribute", "role", "dialog")
	container.Call("setAttribute", "aria-modal", "true")
	container.Call("setAttribute", "aria-label", "Command palette")
	cp.container = container

	// Search input container
	inputContainer := document.Call("createElement", "div")
	inputContainer.Set("className", "px-4 py-3 border-b border-gray-200 dark:border-gray-700")

	// Search input with ARIA combobox attributes
	input := document.Call("createElement", "input")
	input.Set("type", "text")
	input.Set("className", "w-full bg-transparent text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none text-base")
	input.Set("placeholder", props.Placeholder)
	input.Set("autocomplete", "off")
	input.Call("setAttribute", "role", "combobox")
	input.Call("setAttribute", "aria-autocomplete", "list")
	input.Call("setAttribute", "aria-controls", listboxID)
	input.Call("setAttribute", "aria-expanded", "true")
	cp.input = input

	inputContainer.Call("appendChild", input)
	container.Call("appendChild", inputContainer)

	// Results list with listbox role
	resultsList := document.Call("createElement", "div")
	resultsList.Set("className", "max-h-[60vh] overflow-y-auto")
	resultsList.Set("id", listboxID)
	resultsList.Call("setAttribute", "role", "listbox")
	resultsList.Call("setAttribute", "aria-label", "Commands")
	cp.resultsList = resultsList
	container.Call("appendChild", resultsList)

	// Footer with keyboard hints
	footer := document.Call("createElement", "div")
	footer.Set("className", "px-4 py-2 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400")
	footer.Set("innerHTML", `
		<span class="flex items-center gap-1"><kbd class="px-1.5 py-0.5 bg-gray-200 dark:bg-gray-700 rounded text-xs">↑↓</kbd> navigate</span>
		<span class="flex items-center gap-1"><kbd class="px-1.5 py-0.5 bg-gray-200 dark:bg-gray-700 rounded text-xs">↵</kbd> select</span>
		<span class="flex items-center gap-1"><kbd class="px-1.5 py-0.5 bg-gray-200 dark:bg-gray-700 rounded text-xs">esc</kbd> close</span>
	`)
	container.Call("appendChild", footer)

	overlay.Call("appendChild", container)

	// Create focus trap
	cp.focusTrap = NewFocusTrap(container)

	// Render initial commands
	cp.renderCommands()

	// Input event handlers
	input.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		cp.query = input.Get("value").String()
		cp.filter()
		return nil
	}))

	input.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) any {
		event := args[0]
		key := event.Get("key").String()

		switch key {
		case "ArrowDown":
			event.Call("preventDefault")
			cp.highlightNext()
		case "ArrowUp":
			event.Call("preventDefault")
			cp.highlightPrev()
		case "Enter":
			event.Call("preventDefault")
			cp.executeHighlighted()
		case "Escape":
			event.Call("preventDefault")
			cp.Close()
		}
		return nil
	}))

	// Close on overlay click (not container)
	overlay.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		if args[0].Get("target").Equal(overlay) {
			cp.Close()
		}
		return nil
	}))

	return cp
}

func (cp *CommandPalette) renderCommands() {
	document := js.Global().Get("document")
	cp.resultsList.Set("innerHTML", "")

	// Reset option IDs
	cp.optionIDs = nil

	if len(cp.filteredCommands) == 0 {
		empty := document.Call("createElement", "div")
		empty.Set("className", "px-4 py-8 text-center text-gray-500 dark:text-gray-400")
		empty.Set("textContent", cp.props.EmptyMessage)
		cp.resultsList.Call("appendChild", empty)
		cp.input.Call("removeAttribute", "aria-activedescendant")
		return
	}

	// Generate option IDs for each filtered command
	crypto := js.Global().Get("crypto")
	for i := 0; i < len(cp.filteredCommands); i++ {
		cp.optionIDs = append(cp.optionIDs, "cmdpalette-option-"+crypto.Call("randomUUID").String())
	}

	// Group commands by category
	categories := make(map[string][]Command)
	categoryOrder := []string{}

	for _, cmd := range cp.filteredCommands {
		cat := cmd.Category
		if cat == "" {
			cat = "Commands"
		}
		if _, exists := categories[cat]; !exists {
			categoryOrder = append(categoryOrder, cat)
		}
		categories[cat] = append(categories[cat], cmd)
	}

	// Track overall index for highlighting
	overallIdx := 0

	for _, category := range categoryOrder {
		cmds := categories[category]

		// Category header (not a listbox option)
		header := document.Call("createElement", "div")
		header.Set("className", "px-4 py-2 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider bg-gray-50 dark:bg-gray-900/50")
		header.Set("textContent", category)
		header.Call("setAttribute", "role", "presentation") // Not an option
		cp.resultsList.Call("appendChild", header)

		// Command items
		for _, cmd := range cmds {
			item := cp.renderCommandItem(cmd, overallIdx)
			cp.resultsList.Call("appendChild", item)
			overallIdx++
		}
	}

	// Update aria-activedescendant on input
	cp.updateARIAActiveDescendant()
}

func (cp *CommandPalette) renderCommandItem(cmd Command, index int) js.Value {
	document := js.Global().Get("document")

	item := document.Call("createElement", "div")
	baseClass := "px-4 py-2 cursor-pointer flex items-center gap-3"
	if index == cp.highlightIdx {
		item.Set("className", baseClass+" bg-blue-50 dark:bg-blue-900/30")
	} else {
		item.Set("className", baseClass+" hover:bg-gray-100 dark:hover:bg-gray-700/50")
	}
	item.Set("data-index", index)

	// ARIA option attributes
	item.Set("id", cp.optionIDs[index])
	item.Call("setAttribute", "role", "option")
	if index == cp.highlightIdx {
		item.Call("setAttribute", "aria-selected", "true")
	} else {
		item.Call("setAttribute", "aria-selected", "false")
	}

	// Icon
	if cmd.Icon != "" {
		icon := document.Call("createElement", "span")
		icon.Set("className", "text-lg w-6 text-center flex-shrink-0")
		icon.Set("textContent", cmd.Icon)
		item.Call("appendChild", icon)
	}

	// Label and description container
	labelContainer := document.Call("createElement", "div")
	labelContainer.Set("className", "flex-1 min-w-0")

	label := document.Call("createElement", "div")
	label.Set("className", "text-sm font-medium text-gray-900 dark:text-gray-100 truncate")
	label.Set("textContent", cmd.Label)
	labelContainer.Call("appendChild", label)

	if cmd.Description != "" {
		desc := document.Call("createElement", "div")
		desc.Set("className", "text-xs text-gray-500 dark:text-gray-400 truncate")
		desc.Set("textContent", cmd.Description)
		labelContainer.Call("appendChild", desc)
	}

	item.Call("appendChild", labelContainer)

	// Shortcut hint
	if cmd.Shortcut != "" {
		shortcut := document.Call("createElement", "span")
		shortcut.Set("className", "text-xs text-gray-400 dark:text-gray-500 flex-shrink-0")
		shortcut.Set("textContent", cmd.Shortcut)
		item.Call("appendChild", shortcut)
	}

	// Click handler
	command := cmd
	item.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		cp.executeCommand(command)
		return nil
	}))

	// Hover handler to update highlight visually without re-rendering
	idx := index
	item.Call("addEventListener", "mouseenter", js.FuncOf(func(this js.Value, args []js.Value) any {
		cp.highlightIdx = idx
		cp.updateHighlightStyles()
		return nil
	}))

	return item
}

// updateHighlightStyles updates highlight visually without re-rendering DOM
func (cp *CommandPalette) updateHighlightStyles() {
	items := cp.resultsList.Call("querySelectorAll", "[data-index]")
	baseClass := "px-4 py-2 cursor-pointer flex items-center gap-3"

	for i := 0; i < items.Length(); i++ {
		item := items.Index(i)
		idx := item.Get("dataset").Get("index").String()
		highlightStr := js.ValueOf(cp.highlightIdx).String()

		if idx == highlightStr {
			item.Set("className", baseClass+" bg-blue-50 dark:bg-blue-900/30")
			item.Call("setAttribute", "aria-selected", "true")
		} else {
			item.Set("className", baseClass+" hover:bg-gray-100 dark:hover:bg-gray-700/50")
			item.Call("setAttribute", "aria-selected", "false")
		}
	}

	// Update aria-activedescendant on input
	cp.updateARIAActiveDescendant()
}

// updateARIAActiveDescendant updates the input's aria-activedescendant to point to highlighted option
func (cp *CommandPalette) updateARIAActiveDescendant() {
	if cp.highlightIdx >= 0 && cp.highlightIdx < len(cp.optionIDs) {
		cp.input.Call("setAttribute", "aria-activedescendant", cp.optionIDs[cp.highlightIdx])
	} else {
		cp.input.Call("removeAttribute", "aria-activedescendant")
	}
}

func (cp *CommandPalette) filter() {
	query := strings.ToLower(cp.query)
	cp.filteredCommands = nil
	cp.highlightIdx = -1

	for _, cmd := range cp.commands {
		if strings.Contains(strings.ToLower(cmd.Label), query) ||
			strings.Contains(strings.ToLower(cmd.Description), query) {
			cp.filteredCommands = append(cp.filteredCommands, cmd)
		}
	}

	// Auto-highlight first result if there are results
	if len(cp.filteredCommands) > 0 {
		cp.highlightIdx = 0
	}

	cp.renderCommands()
}

func (cp *CommandPalette) highlightNext() {
	if len(cp.filteredCommands) == 0 {
		return
	}
	cp.highlightIdx++
	if cp.highlightIdx >= len(cp.filteredCommands) {
		cp.highlightIdx = 0
	}
	cp.updateHighlightStyles()
	cp.scrollToHighlighted()
}

func (cp *CommandPalette) highlightPrev() {
	if len(cp.filteredCommands) == 0 {
		return
	}
	cp.highlightIdx--
	if cp.highlightIdx < 0 {
		cp.highlightIdx = len(cp.filteredCommands) - 1
	}
	cp.updateHighlightStyles()
	cp.scrollToHighlighted()
}

func (cp *CommandPalette) scrollToHighlighted() {
	if cp.highlightIdx >= 0 {
		items := cp.resultsList.Call("querySelectorAll", "[data-index]")
		for i := 0; i < items.Length(); i++ {
			item := items.Index(i)
			if item.Get("dataset").Get("index").String() == js.ValueOf(cp.highlightIdx).String() {
				item.Call("scrollIntoView", map[string]any{"block": "nearest"})
				break
			}
		}
	}
}

func (cp *CommandPalette) executeHighlighted() {
	if cp.highlightIdx >= 0 && cp.highlightIdx < len(cp.filteredCommands) {
		cp.executeCommand(cp.filteredCommands[cp.highlightIdx])
	}
}

func (cp *CommandPalette) executeCommand(cmd Command) {
	cp.Close()
	if cmd.OnExecute != nil {
		cmd.OnExecute()
	}
}

// Element returns the overlay DOM element
func (cp *CommandPalette) Element() js.Value {
	return cp.overlay
}

// Open shows the command palette
func (cp *CommandPalette) Open() {
	if cp.isOpen {
		return
	}

	cp.isOpen = true
	cp.query = ""
	cp.input.Set("value", "")
	cp.filteredCommands = cp.commands
	cp.highlightIdx = 0 // Pre-highlight first item
	cp.renderCommands()

	cp.overlay.Get("classList").Call("remove", "hidden")

	// Prevent body scroll
	js.Global().Get("document").Get("body").Get("style").Set("overflow", "hidden")

	// Activate focus trap and focus input
	cp.focusTrap.Activate()
	cp.input.Call("focus")
}

// Close hides the command palette
func (cp *CommandPalette) Close() {
	if !cp.isOpen {
		return
	}

	cp.isOpen = false
	cp.overlay.Get("classList").Call("add", "hidden")

	// Restore body scroll
	js.Global().Get("document").Get("body").Get("style").Set("overflow", "")

	// Deactivate focus trap
	cp.focusTrap.Deactivate()

	if cp.props.OnClose != nil {
		cp.props.OnClose()
	}
}

// IsOpen returns whether the palette is currently open
func (cp *CommandPalette) IsOpen() bool {
	return cp.isOpen
}

// SetCommands updates the available commands
func (cp *CommandPalette) SetCommands(commands []Command) {
	cp.commands = commands
	cp.filteredCommands = commands
	cp.renderCommands()
}

// RegisterKeyboardShortcut registers global Cmd+K / Ctrl+K listener
func (cp *CommandPalette) RegisterKeyboardShortcut() {
	cp.keyboardListener = js.FuncOf(func(this js.Value, args []js.Value) any {
		event := args[0]
		key := strings.ToLower(event.Get("key").String())

		// Check for Cmd+K (Mac) or Ctrl+K (Windows/Linux)
		metaKey := event.Get("metaKey").Bool()
		ctrlKey := event.Get("ctrlKey").Bool()

		if key == "k" && (metaKey || ctrlKey) {
			event.Call("preventDefault")
			if cp.isOpen {
				cp.Close()
			} else {
				cp.Open()
			}
		}
		return nil
	})

	js.Global().Get("document").Call("addEventListener", "keydown", cp.keyboardListener)
}

// UnregisterKeyboardShortcut removes the global keyboard listener
func (cp *CommandPalette) UnregisterKeyboardShortcut() {
	if !cp.keyboardListener.IsUndefined() {
		js.Global().Get("document").Call("removeEventListener", "keydown", cp.keyboardListener)
		cp.keyboardListener.Release()
	}
}

// Destroy cleans up the command palette
func (cp *CommandPalette) Destroy() {
	cp.UnregisterKeyboardShortcut()
	cp.focusTrap.Destroy()
}
