//go:build js && wasm

package components

import "syscall/js"

// EmptyStateProps configures an EmptyState component
type EmptyStateProps struct {
	Icon        string // Emoji or text icon (default "📭")
	Title       string // Main heading (required)
	Description string // Explanatory text (optional)
	ActionLabel string // Button text (optional)
	OnAction    func() // Button click handler (optional)
	Compact     bool   // Smaller variant for inline use
}

// EmptyState displays a friendly empty state message
type EmptyState struct {
	container js.Value
	props     EmptyStateProps
}

// NewEmptyState creates a new EmptyState component
func NewEmptyState(props EmptyStateProps) *EmptyState {
	document := js.Global().Get("document")

	// Set default icon
	icon := props.Icon
	if icon == "" {
		icon = "📭"
	}

	// Container - centered flex column
	container := document.Call("createElement", "div")
	containerClass := "flex flex-col items-center justify-center text-center"
	if props.Compact {
		containerClass += " py-4 px-6"
	} else {
		containerClass += " py-8 px-6"
	}
	container.Set("className", containerClass)

	// Icon
	iconEl := document.Call("createElement", "div")
	iconClass := "text-gray-400 dark:text-gray-500 mb-3"
	if props.Compact {
		iconClass = "text-4xl " + iconClass
	} else {
		iconClass = "text-6xl " + iconClass
	}
	iconEl.Set("className", iconClass)
	iconEl.Set("textContent", icon)
	container.Call("appendChild", iconEl)

	// Title
	titleEl := document.Call("createElement", "h3")
	titleClass := "font-medium text-gray-900 dark:text-gray-100"
	if props.Compact {
		titleClass += " text-sm"
	} else {
		titleClass += " text-lg"
	}
	titleEl.Set("className", titleClass)
	titleEl.Set("textContent", props.Title)
	container.Call("appendChild", titleEl)

	// Description (optional)
	if props.Description != "" {
		descEl := document.Call("createElement", "p")
		descClass := "text-sm text-gray-500 dark:text-gray-400 mt-1"
		if props.Compact {
			descClass += " max-w-xs"
		} else {
			descClass += " max-w-md"
		}
		descEl.Set("className", descClass)
		descEl.Set("textContent", props.Description)
		container.Call("appendChild", descEl)
	}

	// Action button (optional)
	if props.ActionLabel != "" && props.OnAction != nil {
		btnContainer := document.Call("createElement", "div")
		btnContainer.Set("className", "mt-4")

		btn := Button(ButtonProps{
			Text:    props.ActionLabel,
			Variant: ButtonPrimary,
			Size:    ButtonSM,
			OnClick: props.OnAction,
		})

		btnContainer.Call("appendChild", btn)
		container.Call("appendChild", btnContainer)
	}

	return &EmptyState{
		container: container,
		props:     props,
	}
}

// Element returns the container DOM element
func (e *EmptyState) Element() js.Value {
	return e.container
}

// NoData creates a generic "no data" empty state
func NoData(title string) *EmptyState {
	return NewEmptyState(EmptyStateProps{
		Icon:        "📭",
		Title:       title,
		Description: "There's nothing here yet.",
	})
}

// NoResults creates a "no results found" empty state with clear filter button
func NoResults(onClear func()) *EmptyState {
	return NewEmptyState(EmptyStateProps{
		Icon:        "🔍",
		Title:       "No results found",
		Description: "Try adjusting your search or filter to find what you're looking for.",
		ActionLabel: "Clear filter",
		OnAction:    onClear,
	})
}

// NoSelection creates a "nothing selected" empty state
func NoSelection() *EmptyState {
	return NewEmptyState(EmptyStateProps{
		Icon:        "👆",
		Title:       "Nothing selected",
		Description: "Select items to see details or perform actions.",
		Compact:     true,
	})
}
