package pages

import (
	"github.com/dougbarrett/gux/core"
)

// About is the about page handler.
func About(r *core.Router) func() core.Node {
	return func() core.Node {
		return core.Div(core.Class("min-h-screen bg-gray-100"),
			Nav(),
			core.Div(core.Class("flex items-center justify-center pt-20"),
				core.Div(core.Class("bg-white rounded-lg shadow-lg p-8 max-w-2xl"),
					core.H1(core.Class("text-3xl font-bold mb-6 text-center"),
						core.Text("About Gux"),
					),
					core.Div(core.Class("space-y-4 text-gray-700"),
						core.P(core.Attrs{},
							core.Text("Gux is a Go framework for building hybrid web applications with SSR and WASM."),
						),
						core.P(core.Attrs{},
							core.Text("Write your UI once in Go, and it renders on the server for fast initial loads, then hydrates with WebAssembly for rich interactivity."),
						),
						core.P(core.Attrs{},
							core.Text("No JavaScript required. Just Go."),
						),
					),
					core.Div(core.Class("mt-8 p-4 bg-blue-50 rounded-lg"),
						core.H2(core.Class("font-semibold text-blue-800 mb-2"),
							core.Text("Features"),
						),
						core.Ul(core.Class("list-disc list-inside text-blue-700 space-y-1"),
							core.Li(core.Attrs{}, core.Text("Server-side rendering for SEO and fast first paint")),
							core.Li(core.Attrs{}, core.Text("WebAssembly hydration for interactivity")),
							core.Li(core.Attrs{}, core.Text("Single binary deployment")),
							core.Li(core.Attrs{}, core.Text("Type-safe components in pure Go")),
							core.Li(core.Attrs{}, core.Text("Reactive state management")),
						),
					),
				),
			),
		)
	}
}
