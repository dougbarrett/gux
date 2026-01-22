package ui

import "github.com/dougbarrett/gux/core"

// TooltipPosition defines the position of a tooltip relative to its trigger.
type TooltipPosition string

const (
	TooltipTop    TooltipPosition = "top"
	TooltipBottom TooltipPosition = "bottom"
	TooltipLeft   TooltipPosition = "left"
	TooltipRight  TooltipPosition = "right"
)

// tooltipPositionClasses maps positions to their Tailwind positioning classes.
var tooltipPositionClasses = map[TooltipPosition]string{
	TooltipTop:    "bottom-full left-1/2 -translate-x-1/2 mb-2",
	TooltipBottom: "top-full left-1/2 -translate-x-1/2 mt-2",
	TooltipLeft:   "right-full top-1/2 -translate-y-1/2 mr-2",
	TooltipRight:  "left-full top-1/2 -translate-y-1/2 ml-2",
}

// tooltipBaseClasses are the common classes applied to all tooltips.
const tooltipBaseClasses = "absolute z-50 px-2 py-1 text-sm text-white bg-gray-900 rounded shadow-lg whitespace-nowrap"

// tooltipVisibilityClasses control showing/hiding via CSS group-hover and group-focus-within.
const tooltipVisibilityClasses = "opacity-0 invisible group-hover:opacity-100 group-hover:visible group-focus-within:opacity-100 group-focus-within:visible"

// tooltipTransitionClasses provide smooth appearance animation.
const tooltipTransitionClasses = "transition-opacity duration-200 pointer-events-none"

// TooltipProps configures a Tooltip component.
type TooltipProps struct {
	Content  string          // Tooltip text content
	Position TooltipPosition // Position relative to trigger (default: TooltipTop)
	Class    string          // Additional classes for the wrapper
	Children []core.Node     // The trigger element(s)
}

// Tooltip wraps content with a tooltip that appears on hover/focus.
//
// The tooltip visibility is CSS-driven using Tailwind's group-hover and
// group-focus-within utilities. This works in SSR because the tooltip is
// always in the DOM but hidden via CSS opacity/visibility classes.
//
// Example:
//
//	Tooltip(TooltipProps{
//	    Content: "Click to submit the form",
//	    Children: []core.Node{
//	        Button(ButtonProps{
//	            Children: []core.Node{core.Text("Submit")},
//	        }),
//	    },
//	})
//
//	// With custom position
//	Tooltip(TooltipProps{
//	    Content:  "View details",
//	    Position: TooltipRight,
//	    Children: []core.Node{core.Text("Info")},
//	})
func Tooltip(props TooltipProps) core.Node {
	// Apply defaults
	position := props.Position
	if position == "" {
		position = TooltipTop
	}

	// Wrapper with group for hover/focus state
	wrapperClass := MergeClasses("relative inline-block group", props.Class)

	// Tooltip element - hidden until hover/focus
	tooltipClass := MergeClasses(
		tooltipBaseClasses,
		tooltipVisibilityClasses,
		tooltipTransitionClasses,
		tooltipPositionClasses[position],
	)

	return core.Div(core.Class(wrapperClass),
		// Trigger content (children first)
		core.Frag(props.Children...),
		// Tooltip element
		core.Div(core.Attrs{
			Class: tooltipClass,
			Extra: map[string]string{"role": "tooltip"},
		}, core.Text(props.Content)),
	)
}
