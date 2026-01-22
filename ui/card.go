package ui

import "github.com/dougbarrett/gux/core"

// CardProps configures the Card component.
type CardProps struct {
	Class    string
	Children []core.Node
}

// Card renders a styled container div with shadow and rounded corners.
// Use CardHeader, CardContent, and CardFooter as children for structured layout.
//
// Example:
//
//	Card(CardProps{
//	    Children: []core.Node{
//	        CardHeader(CardHeaderProps{Children: []core.Node{core.Text("Title")}}),
//	        CardContent(CardContentProps{Children: []core.Node{core.Text("Content")}}),
//	        CardFooter(CardFooterProps{Children: []core.Node{core.Text("Footer")}}),
//	    },
//	})
func Card(props CardProps) core.Node {
	class := MergeClasses(
		"bg-white dark:bg-gray-800 rounded-lg shadow dark:shadow-gray-900",
		props.Class,
	)
	return core.Div(core.Class(class), props.Children...)
}

// CardHeaderProps configures the CardHeader component.
type CardHeaderProps struct {
	Class    string
	Children []core.Node
}

// CardHeader renders the header section of a Card with bottom border.
//
// Example:
//
//	CardHeader(CardHeaderProps{
//	    Children: []core.Node{core.H2(core.Class("text-lg font-semibold"), core.Text("Title"))},
//	})
func CardHeader(props CardHeaderProps) core.Node {
	class := MergeClasses(
		"px-6 py-4 border-b border-gray-200 dark:border-gray-700",
		props.Class,
	)
	return core.Div(core.Class(class), props.Children...)
}

// CardContentProps configures the CardContent component.
type CardContentProps struct {
	Class    string
	Children []core.Node
}

// CardContent renders the main content section of a Card with padding.
//
// Example:
//
//	CardContent(CardContentProps{
//	    Children: []core.Node{core.P(core.Class("text-gray-600"), core.Text("Description"))},
//	})
func CardContent(props CardContentProps) core.Node {
	class := MergeClasses("px-6 py-4", props.Class)
	return core.Div(core.Class(class), props.Children...)
}

// CardFooterProps configures the CardFooter component.
type CardFooterProps struct {
	Class    string
	Children []core.Node
}

// CardFooter renders the footer section of a Card with top border.
//
// Example:
//
//	CardFooter(CardFooterProps{
//	    Children: []core.Node{Button(ButtonProps{Text: "Save"})},
//	})
func CardFooter(props CardFooterProps) core.Node {
	class := MergeClasses(
		"px-6 py-4 border-t border-gray-200 dark:border-gray-700",
		props.Class,
	)
	return core.Div(core.Class(class), props.Children...)
}
