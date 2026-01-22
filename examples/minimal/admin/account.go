package admin

import (
	"github.com/dougbarrett/gux/core"
)

// Account is the admin account settings page.
func Account(r *core.Router) func() core.Node {
	// Loader - fetch account info (dummy data)
	email := ""
	role := ""
	lastLogin := ""

	r.OnLoad(func() {
		email = "admin@example.com"
		role = "Super Admin"
		lastLogin = "2024-01-22 10:30 AM"
	})

	return func() core.Node {
		userEmail := r.StateString("email", email)
		userRole := r.StateString("role", role)
		userLastLogin := r.StateString("lastLogin", lastLogin)

		return core.Div(core.Class("min-h-screen bg-gray-900"),
			Nav(),
			core.Div(core.Class("max-w-2xl mx-auto px-4 py-8"),
				core.H1(core.Class("text-3xl font-bold text-white mb-8"),
					core.Text("Account Settings"),
				),

				// Account info card
				core.Div(core.Class("bg-gray-800 rounded-lg p-6 space-y-6"),
					// Email
					core.Div(core.Class(""),
						core.Label(core.Class("block text-gray-400 text-sm mb-2"),
							core.Text("Email Address"),
						),
						core.Div(core.Class("text-white text-lg"),
							core.Text(userEmail.Get()),
						),
					),

					// Role
					core.Div(core.Class(""),
						core.Label(core.Class("block text-gray-400 text-sm mb-2"),
							core.Text("Role"),
						),
						core.Span(core.Class("inline-block bg-purple-600 text-white px-3 py-1 rounded-full text-sm"),
							core.Text(userRole.Get()),
						),
					),

					// Last login
					core.Div(core.Class(""),
						core.Label(core.Class("block text-gray-400 text-sm mb-2"),
							core.Text("Last Login"),
						),
						core.Div(core.Class("text-gray-300"),
							core.Text(userLastLogin.Get()),
						),
					),

					// Actions
					core.Div(core.Class("pt-4 border-t border-gray-700"),
						core.Button(
							core.Attrs{
								Class: "bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition mr-3",
							},
							core.Text("Change Password"),
						),
						core.Button(
							core.Attrs{
								Class: "bg-gray-700 hover:bg-gray-600 text-white px-4 py-2 rounded-lg transition",
							},
							core.Text("Edit Profile"),
						),
					),
				),

				// Danger zone
				core.Div(core.Class("mt-8 bg-red-900/20 border border-red-800 rounded-lg p-6"),
					core.H3(core.Class("text-red-400 font-semibold mb-4"),
						core.Text("Danger Zone"),
					),
					core.P(core.Class("text-gray-400 text-sm mb-4"),
						core.Text("Once you delete your account, there is no going back. Please be certain."),
					),
					core.Button(
						core.Attrs{
							Class: "bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-lg transition",
						},
						core.Text("Delete Account"),
					),
				),
			),
		)
	}
}
