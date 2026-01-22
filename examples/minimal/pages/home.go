package pages

import (
	"strconv"

	"github.com/dougbarrett/gux/core"
)

// Home is a page handler.
func Home(r *core.Router) func() core.Node {
	// Loader (runs on server)
	message := "Hello World"

	// Component (reactive, runs on client)
	return func() core.Node {
		count := r.StateInt("count", 0)

		return core.Div(core.Class("min-h-screen bg-gray-100"),
			Nav(),
			core.Div(core.Class("flex items-center justify-center pt-20"),
				core.Div(core.Class("bg-white rounded-lg shadow-lg p-8 text-center"),
					core.H1(core.Class("text-3xl font-bold mb-6"),
						core.Text(message),
					),
					core.Div(core.Class("flex items-center justify-center gap-4"),
						core.Span(core.Class("text-4xl font-mono"),
							core.Text(strconv.Itoa(count.Get())),
						),
						core.Button(
							core.Attrs{
								Class: "px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition",
								OnClick: func() {
									count.Set(count.Get() + 1)
								},
							},
							core.Text("Increment"),
						),
					),
				),
			),
		)
	}
}
