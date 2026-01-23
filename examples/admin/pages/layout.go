package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// AdminLayout wraps content with the admin navigation and dark theme.
func AdminLayout(children ...core.Node) core.Node {
	return core.Div(core.Class("min-h-screen bg-gray-900"),
		Nav(),
		core.Main(core.Class("max-w-6xl mx-auto px-4 py-8"),
			core.Frag(children...),
		),
	)
}

// Nav is the horizontal navigation bar for the admin panel.
func Nav() core.Node {
	return core.Div(core.Class("bg-gray-800 border-b border-gray-700"),
		core.Div(core.Class("max-w-6xl mx-auto px-4"),
			core.Div(core.Class("flex justify-between items-center py-4"),
				// Left side: Brand and nav links
				core.Div(core.Class("flex items-center gap-6"),
					core.Span(core.Class("text-xl font-bold text-white"),
						core.Text("Admin Panel"),
					),
					core.A(
						core.Attrs{Href: "/", Class: "text-gray-300 hover:text-white transition"},
						core.Text("Dashboard"),
					),
					core.A(
						core.Attrs{Href: "/users", Class: "text-gray-300 hover:text-white transition"},
						core.Text("Users"),
					),
					core.A(
						core.Attrs{Href: "/activity", Class: "text-gray-300 hover:text-white transition"},
						core.Text("Activity"),
					),
					core.A(
						core.Attrs{Href: "/settings", Class: "text-gray-300 hover:text-white transition"},
						core.Text("Settings"),
					),
				),
				// Right side: User avatar and profile link
				core.Div(core.Class("flex items-center gap-3"),
					ui.Avatar(ui.AvatarProps{
						Name: "Admin User",
						Size: ui.AvatarSM,
					}),
					core.A(
						core.Attrs{Href: "/profile", Class: "text-gray-300 hover:text-white transition"},
						core.Text("Admin User"),
					),
				),
			),
		),
	)
}

// Dashboard is the admin dashboard page (stub).
func Dashboard(r *core.Router) func() core.Node {
	return func() core.Node {
		return AdminLayout(
			core.H1(core.Class("text-2xl font-bold text-white mb-4"),
				core.Text("Dashboard"),
			),
			core.P(core.Class("text-gray-400"),
				core.Text("Dashboard - Coming Soon"),
			),
		)
	}
}

// Users is the user list page (stub).
func Users(r *core.Router) func() core.Node {
	return func() core.Node {
		return AdminLayout(
			core.H1(core.Class("text-2xl font-bold text-white mb-4"),
				core.Text("Users"),
			),
			core.P(core.Class("text-gray-400"),
				core.Text("User Management - Coming Soon"),
			),
		)
	}
}

// UserDetail is the user detail page (stub).
func UserDetail(r *core.Router) func() core.Node {
	return func() core.Node {
		return AdminLayout(
			core.H1(core.Class("text-2xl font-bold text-white mb-4"),
				core.Text("User Detail"),
			),
			core.P(core.Class("text-gray-400"),
				core.Text("User Detail - Coming Soon"),
			),
		)
	}
}

// UserEdit is the user edit page (stub).
func UserEdit(r *core.Router) func() core.Node {
	return func() core.Node {
		return AdminLayout(
			core.H1(core.Class("text-2xl font-bold text-white mb-4"),
				core.Text("Edit User"),
			),
			core.P(core.Class("text-gray-400"),
				core.Text("Edit User - Coming Soon"),
			),
		)
	}
}

// UserNew is the new user page (stub).
func UserNew(r *core.Router) func() core.Node {
	return func() core.Node {
		return AdminLayout(
			core.H1(core.Class("text-2xl font-bold text-white mb-4"),
				core.Text("New User"),
			),
			core.P(core.Class("text-gray-400"),
				core.Text("New User - Coming Soon"),
			),
		)
	}
}

// Activity is the activity log page (stub).
func Activity(r *core.Router) func() core.Node {
	return func() core.Node {
		return AdminLayout(
			core.H1(core.Class("text-2xl font-bold text-white mb-4"),
				core.Text("Activity Log"),
			),
			core.P(core.Class("text-gray-400"),
				core.Text("Activity Log - Coming Soon"),
			),
		)
	}
}

// Settings is the settings page (stub).
func Settings(r *core.Router) func() core.Node {
	return func() core.Node {
		return AdminLayout(
			core.H1(core.Class("text-2xl font-bold text-white mb-4"),
				core.Text("Settings"),
			),
			core.P(core.Class("text-gray-400"),
				core.Text("Settings - Coming Soon"),
			),
		)
	}
}
