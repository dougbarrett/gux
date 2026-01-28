package pages

import (
	"strconv"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/guxgen/api"
)

// Home is a page handler.
func Home(r *core.Router) func() core.Node {
	// Loader (runs on server, also via /__gux_api/pages/index for client navigation)
	message := "Hello World"
	initialCount := 0

	// OnLoad runs only when not hydrated (server or fresh navigation)
	r.OnLoad(func() {
		api.Counters.Get(1, func(counter *api.Counter, err error) {
			if err == nil && counter != nil {
				initialCount = counter.Value
			}
		})
	})

	// Component (reactive)
	return func() core.Node {
		count := r.StateInt("count", initialCount)

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
									newValue := count.Get() + 1
									count.Set(newValue)
									// Persist to database via API
									api.Counters.Update(&api.Counter{
										ID:    1,
										Name:  "default",
										Value: newValue,
									}, func(updated *api.Counter, err error) {
										// Update handled
									})
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
