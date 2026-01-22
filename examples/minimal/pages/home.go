package pages

import (
	"strconv"

	. "github.com/dougbarrett/gux/core"
)

//gux:page /
func (r *Router) Home() func() Node {
	// Loader code would go here (runs on server)
	// For this example, no server data needed

	// Component (reactive, runs on client)
	return func() Node {
		count := r.State("count", 0)

		return Div(Class("min-h-screen bg-gray-100 flex items-center justify-center"),
			Div(Class("bg-white rounded-lg shadow-lg p-8 text-center"),
				H1(Class("text-3xl font-bold mb-6"),
					Text("Hello World"),
				),
				Div(Class("flex items-center justify-center gap-4"),
					Span(Class("text-4xl font-mono"),
						Text(strconv.Itoa(count.Get())),
					),
					Button(
						Attrs{
							Class: "px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition",
							OnClick: func() {
								count.Set(count.Get() + 1)
							},
						},
						Text("Increment"),
					),
				),
			),
		)
	}
}
