package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// DashboardLayout wraps content with the dashboard navigation and dark theme.
func DashboardLayout(children ...core.Node) core.Node {
	return core.Div(core.Class("min-h-screen bg-gray-900"),
		Nav(),
		core.Main(core.Class("max-w-6xl mx-auto px-4 py-8"),
			core.Frag(children...),
		),
	)
}

// Nav is the horizontal navigation bar for the dashboard.
func Nav() core.Node {
	return core.Div(core.Class("bg-gray-800 border-b border-gray-700"),
		core.Div(core.Class("max-w-6xl mx-auto px-4"),
			core.Div(core.Class("flex justify-between items-center py-4"),
				// Left side: Brand and nav links
				core.Div(core.Class("flex items-center gap-6"),
					core.Span(core.Class("text-xl font-bold text-white"),
						core.Text("SaaS App"),
					),
					core.A(
						core.Attrs{Href: "/", Class: "text-gray-300 hover:text-white transition"},
						core.Text("Dashboard"),
					),
					core.A(
						core.Attrs{Href: "/projects", Class: "text-gray-300 hover:text-white transition"},
						core.Text("Projects"),
					),
					core.A(
						core.Attrs{Href: "/settings", Class: "text-gray-300 hover:text-white transition"},
						core.Text("Settings"),
					),
				),
				// Right side: User avatar and profile link
				core.Div(core.Class("flex items-center gap-3"),
					ui.Avatar(ui.AvatarProps{
						Name: "John Doe",
						Size: ui.AvatarSM,
					}),
					core.A(
						core.Attrs{Href: "/profile", Class: "text-gray-300 hover:text-white transition"},
						core.Text("John Doe"),
					),
				),
			),
		),
	)
}
