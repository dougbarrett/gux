package pages

import "github.com/dougbarrett/gux/core"

// Home renders the home page.
func Home(r *core.Router) func() core.Node {
	return func() core.Node {
		return MarketingLayout(r,
			core.Div(core.Class("py-20 px-4 text-center"),
				core.H1(core.Class("text-4xl font-bold text-gray-900 dark:text-white"),
					core.Text("Home"),
				),
				core.P(core.Class("text-gray-600 dark:text-gray-400 mt-4"),
					core.Text("Coming soon..."),
				),
			),
		)
	}
}
