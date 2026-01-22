package ui

import "github.com/dougbarrett/gux/core"

// containerMaxWidths maps size names to Tailwind max-width classes.
var containerMaxWidths = map[string]string{
	"sm":  "max-w-sm",
	"md":  "max-w-md",
	"lg":  "max-w-lg",
	"xl":  "max-w-xl",
	"2xl": "max-w-2xl",
	"7xl": "max-w-7xl",
}

// ContainerProps configures the Container component.
type ContainerProps struct {
	MaxWidth string // "sm", "md", "lg", "xl", "2xl", "7xl" (default: "7xl")
	Class    string
	Children []core.Node
}

// Container constrains content to a max-width with responsive padding.
// Commonly used as the main wrapper for page content.
//
// Example:
//
//	Container(ContainerProps{
//	    MaxWidth: "xl",
//	    Children: []core.Node{core.Text("Page content")},
//	})
func Container(props ContainerProps) core.Node {
	maxWidth := props.MaxWidth
	if maxWidth == "" {
		maxWidth = "7xl"
	}
	class := MergeClasses(
		containerMaxWidths[maxWidth],
		"mx-auto px-4 sm:px-6 lg:px-8",
		props.Class,
	)
	return core.Div(core.Class(class), props.Children...)
}

// alignClasses maps alignment values to Tailwind classes.
var alignClasses = map[string]string{
	"start":   "items-start",
	"center":  "items-center",
	"end":     "items-end",
	"stretch": "items-stretch",
}

// justifyClasses maps justify values to Tailwind classes.
var justifyClasses = map[string]string{
	"start":   "justify-start",
	"center":  "justify-center",
	"end":     "justify-end",
	"between": "justify-between",
	"around":  "justify-around",
}

// StackProps configures VStack and HStack components.
type StackProps struct {
	Gap      string // "0", "1", "2", "4", "6", "8" (default: "4")
	Align    string // "start", "center", "end", "stretch" (default: "stretch")
	Justify  string // "start", "center", "end", "between", "around"
	Class    string
	Children []core.Node
}

// VStack arranges children vertically with gap spacing.
//
// Example:
//
//	VStack(StackProps{
//	    Gap: "4",
//	    Align: "center",
//	    Children: []core.Node{
//	        core.Text("Item 1"),
//	        core.Text("Item 2"),
//	    },
//	})
func VStack(props StackProps) core.Node {
	gap := props.Gap
	if gap == "" {
		gap = "4"
	}
	class := MergeClasses(
		"flex flex-col",
		"gap-"+gap,
		alignClasses[props.Align],
		justifyClasses[props.Justify],
		props.Class,
	)
	return core.Div(core.Class(class), props.Children...)
}

// HStack arranges children horizontally with gap spacing.
//
// Example:
//
//	HStack(StackProps{
//	    Gap: "2",
//	    Justify: "between",
//	    Children: []core.Node{
//	        Button(ButtonProps{Text: "Cancel"}),
//	        Button(ButtonProps{Text: "Save", Variant: "primary"}),
//	    },
//	})
func HStack(props StackProps) core.Node {
	gap := props.Gap
	if gap == "" {
		gap = "4"
	}
	class := MergeClasses(
		"flex flex-row",
		"gap-"+gap,
		alignClasses[props.Align],
		justifyClasses[props.Justify],
		props.Class,
	)
	return core.Div(core.Class(class), props.Children...)
}

// gridColsClasses maps column counts to Tailwind grid classes.
var gridColsClasses = map[string]string{
	"1":  "grid-cols-1",
	"2":  "grid-cols-2",
	"3":  "grid-cols-3",
	"4":  "grid-cols-4",
	"6":  "grid-cols-6",
	"12": "grid-cols-12",
}

// GridProps configures the Grid component.
type GridProps struct {
	Cols     string // "1", "2", "3", "4", "6", "12" (default: "3")
	Gap      string // Same as Stack gap (default: "4")
	Class    string
	Children []core.Node
}

// Grid arranges children in a responsive grid layout.
//
// Example:
//
//	Grid(GridProps{
//	    Cols: "3",
//	    Gap: "6",
//	    Children: []core.Node{
//	        Card(CardProps{...}),
//	        Card(CardProps{...}),
//	        Card(CardProps{...}),
//	    },
//	})
func Grid(props GridProps) core.Node {
	cols := props.Cols
	if cols == "" {
		cols = "3"
	}
	gap := props.Gap
	if gap == "" {
		gap = "4"
	}
	class := MergeClasses(
		"grid",
		gridColsClasses[cols],
		"gap-"+gap,
		props.Class,
	)
	return core.Div(core.Class(class), props.Children...)
}

// DividerProps configures the Divider component.
type DividerProps struct {
	Orientation string // "horizontal" (default), "vertical"
	Class       string
}

// Divider renders a horizontal or vertical separator line.
//
// Example:
//
//	VStack(StackProps{
//	    Children: []core.Node{
//	        core.Text("Section 1"),
//	        Divider(DividerProps{}),
//	        core.Text("Section 2"),
//	    },
//	})
func Divider(props DividerProps) core.Node {
	orientation := props.Orientation
	if orientation == "" {
		orientation = "horizontal"
	}
	var baseClass string
	if orientation == "vertical" {
		baseClass = "w-px h-full bg-gray-200 dark:bg-gray-700"
	} else {
		baseClass = "h-px w-full bg-gray-200 dark:bg-gray-700"
	}
	class := MergeClasses(baseClass, props.Class)
	return core.Div(core.Class(class))
}
