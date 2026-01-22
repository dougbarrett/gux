package pages

import (
	"github.com/dougbarrett/gux/core"
)

// Nav creates the navigation bar component.
func Nav() core.Node {
	return core.Nav(core.Class("bg-white shadow-sm"),
		core.Div(core.Class("max-w-4xl mx-auto px-4 py-3 flex items-center justify-between"),
			core.A(core.Attrs{Class: "text-xl font-bold text-blue-600", Href: "/"},
				core.Text("Gux"),
			),
			core.Div(core.Class("flex gap-6"),
				core.A(core.Attrs{Class: "text-gray-600 hover:text-blue-600 transition", Href: "/"},
					core.Text("Home"),
				),
				core.A(core.Attrs{Class: "text-gray-600 hover:text-blue-600 transition", Href: "/about"},
					core.Text("About"),
				),
			),
		),
	)
}
