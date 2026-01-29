package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// AdminLayout wraps content with the admin navigation and dark theme.
func AdminLayout(r *core.Router, title string, children ...core.Node) core.Node {
	return core.Div(core.Class("min-h-screen bg-gray-900"),
		Nav(r),
		core.Main(core.Class("max-w-6xl mx-auto px-4 py-8"),
			core.Frag(children...),
		),
	)
}

// Nav renders the horizontal navigation bar.
func Nav(r *core.Router) core.Node {
	return core.Div(core.Class("bg-gray-800 border-b border-gray-700"),
		core.Div(core.Class("max-w-6xl mx-auto px-4"),
			core.Div(core.Class("flex justify-between items-center py-4"),
				core.Div(core.Class("flex items-center gap-6"),
					core.Span(core.Class("text-xl font-bold text-white"),
						core.Text("Document Manager"),
					),
					core.A(
						core.Attrs{Href: "/admin/dashboard", Class: "text-gray-300 hover:text-white transition", External: true},
						core.Text("Dashboard"),
					),
					core.A(
						core.Attrs{Href: "/admin/documents", Class: "text-gray-300 hover:text-white transition", External: true},
						core.Text("Documents"),
					),
					core.A(
						core.Attrs{Href: "/admin/users", Class: "text-gray-300 hover:text-white transition", External: true},
						core.Text("Users"),
					),
					core.A(
						core.Attrs{Href: "/admin/audit", Class: "text-gray-300 hover:text-white transition", External: true},
						core.Text("Audit Log"),
					),
				),
				core.Div(core.Class("flex items-center gap-3"),
					func() core.Node {
						user := r.User()
						if user != nil {
							return core.Div(core.Class("flex items-center gap-3"),
								ui.Avatar(ui.AvatarProps{
									Name: user.Name,
									Size: ui.AvatarSM,
								}),
								core.Span(core.Class("text-gray-300"),
									core.Text(user.Name),
								),
							)
						}
						return core.Frag()
					}(),
				),
			),
		),
	)
}
