package admin

import (
	"github.com/dougbarrett/gux/core"
)

// Nav is the admin navigation bar.
func Nav() core.Node {
	return core.Div(core.Class("bg-gray-800 border-b border-gray-700"),
		core.Div(core.Class("max-w-6xl mx-auto px-4"),
			core.Div(core.Class("flex justify-between items-center py-4"),
				core.Div(core.Class("flex items-center gap-6"),
					core.Span(core.Class("text-xl font-bold text-white"),
						core.Text("Admin Panel"),
					),
					core.A(
						core.Attrs{Href: "/admin", Class: "text-gray-300 hover:text-white transition"},
						core.Text("Dashboard"),
					),
					core.A(
						core.Attrs{Href: "/admin/users", Class: "text-gray-300 hover:text-white transition"},
						core.Text("Users"),
					),
					core.A(
						core.Attrs{Href: "/admin/posts", Class: "text-gray-300 hover:text-white transition"},
						core.Text("Posts"),
					),
					core.A(
						core.Attrs{Href: "/admin/account", Class: "text-gray-300 hover:text-white transition"},
						core.Text("Account"),
					),
				),
				core.A(
					core.Attrs{Href: "/", Class: "text-gray-400 hover:text-white transition text-sm", External: true},
					core.Text("← Back to Site"),
				),
			),
		),
	)
}
