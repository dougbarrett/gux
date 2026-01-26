package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Admin is the admin-only page demonstrating RBAC
func Admin(r *core.Router) func() core.Node {
	return func() core.Node {
		user := r.User()
		userName := "Admin"
		if user != nil {
			userName = user.Name
		}

		return PageLayout(
			core.Div(core.Class("max-w-4xl mx-auto"),
				// Admin header
				core.Div(core.Class("mb-8"),
					core.H1(core.Class("text-3xl font-bold text-gray-900 dark:text-white mb-2"),
						core.Text("Admin Dashboard")),
					core.P(core.Class("text-gray-600 dark:text-gray-400"),
						core.Text("Welcome, "+userName+"! This page is only visible to admins.")),
				),

				// RBAC explanation
				ui.Alert(ui.AlertProps{
					Variant: ui.AlertInfo,
					Title:   "Role-Based Access Control",
					Message: "This page is protected with .WithRoles(\"admin\"). Only users with the 'admin' role can access it. Regular users will see a 403 Forbidden error.",
					Class:   "mb-8",
				}),

				// Admin actions
				ui.Grid(ui.GridProps{
					Cols: "3",
					Gap:  "6",
					Children: []core.Node{
						adminCard("Users", "Manage user accounts", "12 users"),
						adminCard("Settings", "System configuration", "3 pending"),
						adminCard("Logs", "View system logs", "1.2k today"),
					},
				}),

				// Back to dashboard
				core.Div(core.Class("mt-8"),
					core.A(
						core.Attrs{
							Href:  "/dashboard",
							Class: "text-blue-600 dark:text-blue-400 hover:underline",
						},
						core.Text("← Back to Dashboard"),
					),
				),
			),
		)
	}
}

func adminCard(title, description, stat string) core.Node {
	return ui.Card(ui.CardProps{
		Children: []core.Node{
			ui.CardHeader(ui.CardHeaderProps{
				Children: []core.Node{
					core.H2(core.Class("text-lg font-semibold text-gray-900 dark:text-white"),
						core.Text(title)),
				},
			}),
			ui.CardContent(ui.CardContentProps{
				Children: []core.Node{
					core.P(core.Class("text-gray-600 dark:text-gray-400 mb-2"),
						core.Text(description)),
					core.P(core.Class("text-2xl font-bold text-blue-600 dark:text-blue-400"),
						core.Text(stat)),
				},
			}),
		},
	})
}
