package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Dashboard renders the admin dashboard with breadcrumbs.
func Dashboard(r *core.Router) func() core.Node {
	return func() core.Node {
		return core.Div(core.Class("p-6"),
			// Page header with breadcrumbs
			ui.PageHeader(ui.PageHeaderProps{
				Title:    "Dashboard",
				Subtitle: "Welcome to the admin panel",
				Breadcrumbs: []ui.BreadcrumbItem{
					{Label: "Admin", Href: "/admin"},
					{Label: "Dashboard"},
				},
			}),

			// Stats cards
			core.Div(core.Class("grid grid-cols-1 md:grid-cols-3 gap-6 mt-6"),
				// Total Users card
				ui.Card(ui.CardProps{
					Class: "bg-gray-800 border-gray-700",
					Children: []core.Node{
						ui.CardHeader(ui.CardHeaderProps{
							Children: []core.Node{
								core.Div(core.Class("flex items-center gap-2"),
									ui.Icon(ui.IconProps{Name: "users", Size: ui.IconMD, Class: "text-blue-400"}),
									core.Span(core.Class("text-gray-400 text-sm"), core.Text("Total Users")),
								),
							},
						}),
						ui.CardContent(ui.CardContentProps{
							Children: []core.Node{
								core.Div(core.Class("text-3xl font-bold text-white"), core.Text("1,234")),
								core.Div(core.Class("text-sm text-green-400 mt-1"), core.Text("+12% from last month")),
							},
						}),
					},
				}),

				// Active Sessions card
				ui.Card(ui.CardProps{
					Class: "bg-gray-800 border-gray-700",
					Children: []core.Node{
						ui.CardHeader(ui.CardHeaderProps{
							Children: []core.Node{
								core.Div(core.Class("flex items-center gap-2"),
									ui.Icon(ui.IconProps{Name: "activity", Size: ui.IconMD, Class: "text-green-400"}),
									core.Span(core.Class("text-gray-400 text-sm"), core.Text("Active Sessions")),
								),
							},
						}),
						ui.CardContent(ui.CardContentProps{
							Children: []core.Node{
								core.Div(core.Class("text-3xl font-bold text-white"), core.Text("89")),
								core.Div(core.Class("text-sm text-gray-400 mt-1"), core.Text("Currently online")),
							},
						}),
					},
				}),

				// Storage Used card
				ui.Card(ui.CardProps{
					Class: "bg-gray-800 border-gray-700",
					Children: []core.Node{
						ui.CardHeader(ui.CardHeaderProps{
							Children: []core.Node{
								core.Div(core.Class("flex items-center gap-2"),
									ui.Icon(ui.IconProps{Name: "database", Size: ui.IconMD, Class: "text-purple-400"}),
									core.Span(core.Class("text-gray-400 text-sm"), core.Text("Storage Used")),
								),
							},
						}),
						ui.CardContent(ui.CardContentProps{
							Children: []core.Node{
								core.Div(core.Class("text-3xl font-bold text-white"), core.Text("2.4 GB")),
								core.Div(core.Class("text-sm text-gray-400 mt-1"), core.Text("of 10 GB")),
							},
						}),
					},
				}),
			),

			// Recent Activity section
			core.Div(core.Class("mt-8"),
				core.H2(core.Class("text-xl font-semibold text-white mb-4"), core.Text("Recent Activity")),
				ui.Card(ui.CardProps{
					Class: "bg-gray-800 border-gray-700",
					Children: []core.Node{
						ui.CardContent(ui.CardContentProps{
							Children: []core.Node{
								activityItem("User registered", "john@example.com", "2 minutes ago"),
								activityItem("Settings updated", "Admin", "15 minutes ago"),
								activityItem("User logged in", "jane@example.com", "1 hour ago"),
								activityItem("Password reset", "bob@example.com", "3 hours ago"),
							},
						}),
					},
				}),
			),
		)
	}
}

func activityItem(action, user, time string) core.Node {
	return core.Div(core.Class("flex items-center justify-between py-3 border-b border-gray-700 last:border-0"),
		core.Div(core.Attrs{},
			core.Div(core.Class("text-white"), core.Text(action)),
			core.Div(core.Class("text-sm text-gray-400"), core.Text(user)),
		),
		core.Div(core.Class("text-sm text-gray-500"), core.Text(time)),
	)
}
