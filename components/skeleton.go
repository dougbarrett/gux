//go:build js && wasm

package components

import "syscall/js"

// SkeletonProps configures a Skeleton loader
type SkeletonProps struct {
	Width    string // e.g., "w-full", "w-32", "100px"
	Height   string // e.g., "h-4", "h-8", "20px"
	Rounded  bool   // rounded corners
	Circle   bool   // circular shape
	Animate  bool   // pulse animation (default true)
	Class    string // additional classes
}

// Skeleton creates a placeholder loading element
func Skeleton(props SkeletonProps) js.Value {
	document := js.Global().Get("document")

	el := document.Call("createElement", "div")

	width := props.Width
	if width == "" {
		width = "w-full"
	}

	height := props.Height
	if height == "" {
		height = "h-4"
	}

	className := "bg-gray-200 dark:bg-gray-700"

	// Handle explicit pixel values vs Tailwind classes
	if width[0] >= '0' && width[0] <= '9' {
		el.Get("style").Set("width", width)
	} else {
		className += " " + width
	}

	if height[0] >= '0' && height[0] <= '9' {
		el.Get("style").Set("height", height)
	} else {
		className += " " + height
	}

	if props.Circle {
		className += " rounded-full"
	} else if props.Rounded {
		className += " rounded"
	}

	// Animation is on by default
	if props.Animate || (!props.Animate && props.Width == "" && props.Height == "") {
		className += " animate-pulse"
	}

	if props.Class != "" {
		className += " " + props.Class
	}

	el.Set("className", className)

	return el
}

// SkeletonText creates a text placeholder
func SkeletonText(lines int) js.Value {
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")
	container.Set("className", "space-y-2")

	for i := 0; i < lines; i++ {
		width := "w-full"
		if i == lines-1 && lines > 1 {
			width = "w-3/4" // Last line is shorter
		}
		container.Call("appendChild", Skeleton(SkeletonProps{
			Width:   width,
			Height:  "h-4",
			Rounded: true,
			Animate: true,
		}))
	}

	return container
}

// SkeletonAvatar creates a circular avatar placeholder
func SkeletonAvatar(size string) js.Value {
	if size == "" {
		size = "w-10 h-10"
	}
	document := js.Global().Get("document")
	el := document.Call("createElement", "div")
	el.Set("className", "bg-gray-200 dark:bg-gray-700 rounded-full animate-pulse "+size)
	return el
}

// SkeletonCard creates a card placeholder with avatar, title, and text
func SkeletonCard() js.Value {
	document := js.Global().Get("document")

	card := document.Call("createElement", "div")
	card.Set("className", "bg-white dark:bg-gray-800 rounded-lg shadow p-4 space-y-4")

	// Header with avatar and title
	header := document.Call("createElement", "div")
	header.Set("className", "flex items-center space-x-3")
	header.Call("appendChild", SkeletonAvatar("w-10 h-10"))

	titleGroup := document.Call("createElement", "div")
	titleGroup.Set("className", "space-y-2 flex-1")
	titleGroup.Call("appendChild", Skeleton(SkeletonProps{Width: "w-1/3", Height: "h-4", Rounded: true, Animate: true}))
	titleGroup.Call("appendChild", Skeleton(SkeletonProps{Width: "w-1/4", Height: "h-3", Rounded: true, Animate: true}))
	header.Call("appendChild", titleGroup)
	card.Call("appendChild", header)

	// Image placeholder
	card.Call("appendChild", Skeleton(SkeletonProps{Width: "w-full", Height: "h-48", Rounded: true, Animate: true}))

	// Text lines
	card.Call("appendChild", SkeletonText(3))

	return card
}

// SkeletonTable creates a table placeholder
func SkeletonTable(rows, cols int) js.Value {
	document := js.Global().Get("document")

	table := document.Call("createElement", "div")
	table.Set("className", "space-y-2")

	// Header row
	headerRow := document.Call("createElement", "div")
	headerRow.Set("className", "flex space-x-4 pb-2 border-b border-gray-200 dark:border-gray-700")
	for j := 0; j < cols; j++ {
		headerRow.Call("appendChild", Skeleton(SkeletonProps{Width: "flex-1", Height: "h-4", Rounded: true, Animate: true}))
	}
	table.Call("appendChild", headerRow)

	// Data rows
	for i := 0; i < rows; i++ {
		row := document.Call("createElement", "div")
		row.Set("className", "flex space-x-4 py-2")
		for j := 0; j < cols; j++ {
			row.Call("appendChild", Skeleton(SkeletonProps{Width: "flex-1", Height: "h-4", Rounded: true, Animate: true}))
		}
		table.Call("appendChild", row)
	}

	return table
}
