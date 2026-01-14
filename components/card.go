//go:build js && wasm

package components

import "syscall/js"

// CardProps configures a Card component
type CardProps struct {
	ClassName string
	Children  []js.Value
}

// Card creates a styled card container.
// Default styling: white background, rounded corners, shadow, padding.
func Card(children ...js.Value) js.Value {
	return CardWithClass("", children...)
}

// CardWithClass creates a card with additional custom classes.
func CardWithClass(extraClass string, children ...js.Value) js.Value {
	className := "bg-white rounded-lg shadow p-6"
	if extraClass != "" {
		className += " " + extraClass
	}
	return Div(className, children...)
}

// TitledCard creates a card with a title and optional description
func TitledCard(title, description string, children ...js.Value) js.Value {
	content := []js.Value{H2(title)}
	if description != "" {
		content = append(content, Text(description))
	}
	content = append(content, children...)
	return Card(content...)
}

// Section creates a titled section with content
func Section(title string, children ...js.Value) js.Value {
	content := []js.Value{H3(title)}
	content = append(content, children...)
	return Div("space-y-3", content...)
}
