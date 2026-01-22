package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Dashboard is the authenticated user's home page
func Dashboard(r *core.Router) func() core.Node {
	return func() core.Node {
		return PageLayout(
			core.Div(core.Class("max-w-4xl mx-auto"),
				// Welcome section
				core.Div(core.Class("mb-8"),
					core.H1(core.Class("text-3xl font-bold text-gray-900 dark:text-white mb-2"),
						core.Text("Welcome back!")),
					core.P(core.Class("text-gray-600 dark:text-gray-400"),
						core.Text("You are successfully logged in.")),
				),

				// Success alert
				ui.Alert(ui.AlertProps{
					Variant: ui.AlertSuccess,
					Title:   "Authentication Complete",
					Message: "Your session is active. In a real application, this page would show your dashboard content.",
					Class:   "mb-8",
				}),

				// Action cards
				ui.Grid(ui.GridProps{
					Cols: "2",
					Gap:  "6",
					Children: []core.Node{
						ui.Card(ui.CardProps{
							Children: []core.Node{
								ui.CardHeader(ui.CardHeaderProps{
									Children: []core.Node{
										core.H2(core.Class("text-lg font-semibold text-gray-900 dark:text-white"),
											core.Text("Profile Settings")),
									},
								}),
								ui.CardContent(ui.CardContentProps{
									Children: []core.Node{
										core.P(core.Class("text-gray-600 dark:text-gray-400 mb-4"),
											core.Text("Update your name, email, or password.")),
										ui.Button(ui.ButtonProps{
											Variant: ui.ButtonOutline,
											Children: []core.Node{
												core.Text("Edit Profile"),
											},
										}),
									},
								}),
							},
						}),
						ui.Card(ui.CardProps{
							Children: []core.Node{
								ui.CardHeader(ui.CardHeaderProps{
									Children: []core.Node{
										core.H2(core.Class("text-lg font-semibold text-gray-900 dark:text-white"),
											core.Text("Sign Out")),
									},
								}),
								ui.CardContent(ui.CardContentProps{
									Children: []core.Node{
										core.P(core.Class("text-gray-600 dark:text-gray-400 mb-4"),
											core.Text("End your session and return to login.")),
										ui.Button(ui.ButtonProps{
											Variant: ui.ButtonDestructive,
											OnClick: func() { r.Navigate("/login") },
											Children: []core.Node{
												core.Text("Sign Out"),
											},
										}),
									},
								}),
							},
						}),
					},
				}),
			),
		)
	}
}
