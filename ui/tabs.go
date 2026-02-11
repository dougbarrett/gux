package ui

import "github.com/dougbarrett/gux/core"

// TabsProps configures the Tabs container.
type TabsProps struct {
	Class    string      // Additional wrapper classes
	Children []core.Node // TabList and TabPanels
}

// Tabs creates a tabs container.
//
// Tabs is a compound component that wraps TabList and TabPanel children.
// State management (which tab is active) lives at the page level.
//
// Example:
//
//	activeTab := r.StateInt("activeTab", 0)
//	Tabs(TabsProps{
//	    Children: []core.Node{
//	        TabList(TabListProps{
//	            Children: []core.Node{
//	                Tab(TabProps{Active: activeTab.Get() == 0, OnSelect: func() { activeTab.Set(0) }, Children: []core.Node{core.Text("Tab 1")}}),
//	                Tab(TabProps{Active: activeTab.Get() == 1, OnSelect: func() { activeTab.Set(1) }, Children: []core.Node{core.Text("Tab 2")}}),
//	            },
//	        }),
//	        TabPanel(TabPanelProps{Active: activeTab.Get() == 0, Children: []core.Node{core.Text("Content 1")}}),
//	        TabPanel(TabPanelProps{Active: activeTab.Get() == 1, Children: []core.Node{core.Text("Content 2")}}),
//	    },
//	})
func Tabs(props TabsProps) core.Node {
	class := MergeClasses("w-full", props.Class)
	return core.Div(core.Class(class), props.Children...)
}

// TabListProps configures the TabList component.
type TabListProps struct {
	Class    string      // Additional wrapper classes
	Children []core.Node // Tab elements
}

// TabList creates a container for tab buttons.
//
// TabList provides the tablist role and contains Tab elements.
// It includes horizontal scrolling support for many tabs.
//
// Example:
//
//	TabList(TabListProps{
//	    Children: []core.Node{
//	        Tab(TabProps{Active: true, Children: []core.Node{core.Text("First")}}),
//	        Tab(TabProps{Active: false, Children: []core.Node{core.Text("Second")}}),
//	    },
//	})
func TabList(props TabListProps) core.Node {
	class := MergeClasses(
		"flex border-b border-gray-200 dark:border-gray-700 overflow-x-auto",
		props.Class,
	)
	return core.Div(core.Attrs{
		Class: class,
		Extra: map[string]string{
			"role":       "tablist",
			"aria-label": "Tabs",
		},
	}, props.Children...)
}

// TabProps configures a Tab component.
type TabProps struct {
	Active   bool        // Whether this tab is selected
	OnSelect func()      // Called when tab is clicked
	Disabled bool        // Disabled state
	Class    string      // Additional classes
	Children []core.Node // Tab label content

	// Alpine.js props
	XActive   string // x-bind:class expression for active state
	XOnSelect string // x-on:click expression for tab selection
}

// Tab class constants
const (
	// tabBaseClasses are applied to all tabs.
	tabBaseClasses = "px-4 py-2 font-medium text-sm cursor-pointer border-b-2 transition-colors whitespace-nowrap"

	// tabFocusClasses handle focus styling.
	tabFocusClasses = "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset"

	// tabActiveClasses are applied when Active=true.
	tabActiveClasses = "border-blue-500 text-blue-600 dark:text-blue-400"

	// tabInactiveClasses are applied when Active=false and not Disabled.
	tabInactiveClasses = "border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:border-gray-300"

	// tabDisabledClasses are applied when Disabled=true.
	tabDisabledClasses = "border-transparent text-gray-300 dark:text-gray-600 cursor-not-allowed"
)

// Tab creates a tab button.
//
// Tab renders as a button with ARIA tab semantics. It uses roving tabindex
// for keyboard navigation: active tab has tabindex="0", others have "-1".
//
// Example:
//
//	Tab(TabProps{
//	    Active:   isActive,
//	    OnSelect: func() { setActive(index) },
//	    Children: []core.Node{core.Text("Settings")},
//	})
func Tab(props TabProps) core.Node {
	// Build class based on state
	class := MergeClasses(
		tabBaseClasses,
		tabFocusClasses,
		ConditionalClass(props.Active, tabActiveClasses),
		ConditionalClass(!props.Active && !props.Disabled, tabInactiveClasses),
		ConditionalClass(props.Disabled, tabDisabledClasses),
		props.Class,
	)

	// Build extra attributes
	extra := map[string]string{
		"role":          "tab",
		"aria-selected": boolToAttr(props.Active),
	}

	// Roving tabindex: active tab gets 0, others get -1
	if props.Active {
		extra["tabindex"] = "0"
	} else {
		extra["tabindex"] = "-1"
	}

	// Disabled state
	if props.Disabled {
		extra["disabled"] = "disabled"
		extra["aria-disabled"] = "true"
	}

	attrs := core.Attrs{
		Class:   class,
		OnClick: props.OnSelect,
		Extra:   extra,
	}

	// Alpine.js directives
	if props.XOnSelect != "" {
		attrs.XOn = map[string]string{"click": props.XOnSelect}
	}
	if props.XActive != "" {
		attrs.XBind = map[string]string{"class": props.XActive}
	}

	return core.Button(attrs, props.Children...)
}

// TabPanelProps configures a TabPanel component.
type TabPanelProps struct {
	Active   bool        // Whether this panel is visible
	Class    string      // Additional classes
	Children []core.Node // Panel content

	// Alpine.js props
	XShow string // x-show expression for panel visibility
}

// TabPanel creates a tab content panel.
//
// TabPanel renders with tabpanel role. Inactive panels are hidden via CSS
// class but remain in the DOM for consistent SSR output.
//
// Example:
//
//	TabPanel(TabPanelProps{
//	    Active:   activeTab == 0,
//	    Children: []core.Node{core.Text("Panel content here")},
//	})
func TabPanel(props TabPanelProps) core.Node {
	class := MergeClasses(
		"mt-4",
		ConditionalClass(!props.Active, "hidden"),
		props.Class,
	)

	attrs := core.Attrs{
		Class: class,
		Extra: map[string]string{
			"role":     "tabpanel",
			"tabindex": "0",
		},
	}

	// Alpine.js directives
	if props.XShow != "" {
		attrs.XShow = props.XShow
		attrs.XCloak = true
	}

	return core.Div(attrs, props.Children...)
}
