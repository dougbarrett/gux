package pages

import (
	. "github.com/dougbarrett/gux/core"
	"strconv"
)

// Home renders a simple page with an interactive counter.
func Home(count int, onIncrement func()) Node {
	return Div(Class("min-h-screen bg-gray-100 flex items-center justify-center"),
		Div(Class("bg-white rounded-lg shadow-lg p-8 text-center"),
			H1(Class("text-3xl font-bold mb-6"),
				Text("Hello World"),
			),
			Div(Class("flex items-center justify-center gap-4"),
				Span(Class("text-4xl font-mono"),
					Text(strconv.Itoa(count)),
				),
				Button(
					Attrs{
						Class:   "px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition",
						OnClick: onIncrement,
					},
					Text("Increment"),
				),
			),
		),
	)
}
