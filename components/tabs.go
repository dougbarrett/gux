//go:build js && wasm

package components

import "syscall/js"

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
	tabButtons  []js.Value
	tabPanels   []js.Value
	activeIndex int
	props       TabsProps
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

	t := &Tabs{
		container:   container,
		tabButtons:  make([]js.Value, len(props.Tabs)),
		tabPanels:   make([]js.Value, len(props.Tabs)),
		activeIndex: props.ActiveIndex,
		props:       props,
	}

	// Tab list
	tabList := document.Call("createElement", "div")
	tabList.Set("className", "border-b border-gray-200 dark:border-gray-700")

	tabNav := document.Call("createElement", "nav")
	tabNav.Set("className", "flex space-x-8")

	for i, tab := range props.Tabs {
		btn := document.Call("createElement", "button")
		btn.Set("textContent", tab.Label)

		idx := i
		btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			t.SetActive(idx)
			return nil
		}))

		t.tabButtons[i] = btn
		tabNav.Call("appendChild", btn)
	}

	tabList.Call("appendChild", tabNav)
	container.Call("appendChild", tabList)

	// Tab panels
	panelsContainer := document.Call("createElement", "div")
	panelsContainer.Set("className", "mt-4")

	for i, tab := range props.Tabs {
		panel := document.Call("createElement", "div")
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
	activeClass := "border-b-2 border-blue-500 text-blue-600 dark:text-blue-400 py-2 px-1 font-medium text-sm cursor-pointer"
	inactiveClass := "border-b-2 border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300 dark:hover:border-gray-600 py-2 px-1 font-medium text-sm cursor-pointer"

	for i := range t.tabButtons {
		if i == t.activeIndex {
			t.tabButtons[i].Set("className", activeClass)
			t.tabPanels[i].Get("style").Set("display", "block")
		} else {
			t.tabButtons[i].Set("className", inactiveClass)
			t.tabPanels[i].Get("style").Set("display", "none")
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
