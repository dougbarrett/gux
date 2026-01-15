//go:build js && wasm

package components

import (
	"strconv"
	"syscall/js"
)

// Tab represents a single tab
type Tab struct {
	Label   string
	Content js.Value
	OnSelect func()
}

// TabsProps configures a Tabs component
type TabsProps struct {
	Tabs        []Tab
	ActiveIndex int
	ClassName   string
	OnChange    func(index int)
}

// Tabs creates a tabbed content component
type Tabs struct {
	container   js.Value
	tabNav      js.Value   // tablist element for keyboard handler
	tabButtons  []js.Value
	tabPanels   []js.Value
	tabIDs      []string // unique IDs for tabs
	panelIDs    []string // unique IDs for panels
	activeIndex int
	props       TabsProps
	keyHandler  js.Func // keyboard navigation handler
}

// NewTabs creates a new Tabs component
func NewTabs(props TabsProps) *Tabs {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	className := "w-full"
	if props.ClassName != "" {
		className = props.ClassName
	}
	container.Set("className", className)

	// Generate unique IDs for tabs and panels
	crypto := js.Global().Get("crypto")
	tabIDs := make([]string, len(props.Tabs))
	panelIDs := make([]string, len(props.Tabs))
	for i := range props.Tabs {
		uuid := crypto.Call("randomUUID").String()
		tabIDs[i] = "tabs-tab-" + strconv.Itoa(i) + "-" + uuid
		panelIDs[i] = "tabs-panel-" + strconv.Itoa(i) + "-" + uuid
	}

	t := &Tabs{
		container:   container,
		tabButtons:  make([]js.Value, len(props.Tabs)),
		tabPanels:   make([]js.Value, len(props.Tabs)),
		tabIDs:      tabIDs,
		panelIDs:    panelIDs,
		activeIndex: props.ActiveIndex,
		props:       props,
	}

	// Tab list - scrollable on mobile
	tabList := document.Call("createElement", "div")
	tabList.Set("className", "border-b border-gray-200 dark:border-gray-700 overflow-x-auto scrollbar-hide")

	tabNav := document.Call("createElement", "nav")
	tabNav.Set("className", "flex space-x-4 md:space-x-8 min-w-max px-1")
	tabNav.Call("setAttribute", "role", "tablist")
	tabNav.Call("setAttribute", "aria-label", "Tabs")

	for i, tab := range props.Tabs {
		btn := document.Call("createElement", "button")
		btn.Set("textContent", tab.Label)
		btn.Set("type", "button")

		// ARIA tab attributes
		btn.Call("setAttribute", "role", "tab")
		btn.Set("id", tabIDs[i])
		btn.Call("setAttribute", "aria-controls", panelIDs[i])
		if i == props.ActiveIndex {
			btn.Call("setAttribute", "aria-selected", "true")
			btn.Call("setAttribute", "tabindex", "0")
		} else {
			btn.Call("setAttribute", "aria-selected", "false")
			btn.Call("setAttribute", "tabindex", "-1")
		}

		idx := i
		btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			t.SetActive(idx)
			return nil
		}))

		t.tabButtons[i] = btn
		tabNav.Call("appendChild", btn)
	}

	t.tabNav = tabNav

	// Keyboard navigation handler - WAI-ARIA Tabs pattern
	t.keyHandler = js.FuncOf(func(this js.Value, args []js.Value) any {
		event := args[0]
		key := event.Get("key").String()

		switch key {
		case "ArrowRight":
			event.Call("preventDefault")
			// Move to next tab, wrap to first if at end
			nextIdx := (t.activeIndex + 1) % len(t.tabButtons)
			t.SetActive(nextIdx)
			t.tabButtons[nextIdx].Call("focus")
		case "ArrowLeft":
			event.Call("preventDefault")
			// Move to previous tab, wrap to last if at start
			prevIdx := t.activeIndex - 1
			if prevIdx < 0 {
				prevIdx = len(t.tabButtons) - 1
			}
			t.SetActive(prevIdx)
			t.tabButtons[prevIdx].Call("focus")
		case "Home":
			event.Call("preventDefault")
			t.SetActive(0)
			t.tabButtons[0].Call("focus")
		case "End":
			event.Call("preventDefault")
			lastIdx := len(t.tabButtons) - 1
			t.SetActive(lastIdx)
			t.tabButtons[lastIdx].Call("focus")
		}
		return nil
	})
	tabNav.Call("addEventListener", "keydown", t.keyHandler)

	tabList.Call("appendChild", tabNav)
	container.Call("appendChild", tabList)

	// Tab panels
	panelsContainer := document.Call("createElement", "div")
	panelsContainer.Set("className", "mt-4")

	for i, tab := range props.Tabs {
		panel := document.Call("createElement", "div")

		// ARIA tabpanel attributes
		panel.Call("setAttribute", "role", "tabpanel")
		panel.Set("id", panelIDs[i])
		panel.Call("setAttribute", "aria-labelledby", tabIDs[i])
		panel.Call("setAttribute", "tabindex", "0")

		if !tab.Content.IsUndefined() && !tab.Content.IsNull() {
			panel.Call("appendChild", tab.Content)
		}
		t.tabPanels[i] = panel
		panelsContainer.Call("appendChild", panel)
	}

	container.Call("appendChild", panelsContainer)

	// Set initial active state
	t.updateStyles()

	return t
}

// Element returns the container DOM element
func (t *Tabs) Element() js.Value {
	return t.container
}

// SetActive sets the active tab by index
func (t *Tabs) SetActive(index int) {
	if index < 0 || index >= len(t.tabButtons) {
		return
	}

	t.activeIndex = index
	t.updateStyles()

	// Call tab's onSelect callback
	if t.props.Tabs[index].OnSelect != nil {
		t.props.Tabs[index].OnSelect()
	}

	// Call onChange callback
	if t.props.OnChange != nil {
		t.props.OnChange(index)
	}
}

// ActiveIndex returns the currently active tab index
func (t *Tabs) ActiveIndex() int {
	return t.activeIndex
}

func (t *Tabs) updateStyles() {
	activeClass := "border-b-2 border-blue-500 text-blue-600 dark:text-blue-400 py-2 px-1 font-medium text-sm cursor-pointer focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset"
	inactiveClass := "border-b-2 border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600 py-2 px-1 font-medium text-sm cursor-pointer focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset"

	for i := range t.tabButtons {
		if i == t.activeIndex {
			t.tabButtons[i].Set("className", activeClass)
			t.tabPanels[i].Get("style").Set("display", "block")
			// Update ARIA states for active tab
			t.tabButtons[i].Call("setAttribute", "aria-selected", "true")
			t.tabButtons[i].Call("setAttribute", "tabindex", "0")
		} else {
			t.tabButtons[i].Set("className", inactiveClass)
			t.tabPanels[i].Get("style").Set("display", "none")
			// Update ARIA states for inactive tabs
			t.tabButtons[i].Call("setAttribute", "aria-selected", "false")
			t.tabButtons[i].Call("setAttribute", "tabindex", "-1")
		}
	}
}

// SetTabContent updates the content of a specific tab
func (t *Tabs) SetTabContent(index int, content js.Value) {
	if index < 0 || index >= len(t.tabPanels) {
		return
	}

	t.tabPanels[index].Set("innerHTML", "")
	t.tabPanels[index].Call("appendChild", content)
}
