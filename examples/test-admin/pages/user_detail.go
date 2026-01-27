package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// UserDetail renders a single user detail page with breadcrumbs.
func UserDetail(r *core.Router) func() core.Node {
	// Get user ID from route params
	userID := r.Param("id")

	// Sample user data (in real app, fetch from API)
	userName := "John Doe"
	userEmail := "john@example.com"
	userRole := "user"

	return func() core.Node {
		return core.Div(core.Class("p-6"),
			// Page header with breadcrumbs
			ui.PageHeader(ui.PageHeaderProps{
				Title:    userName,
				Subtitle: "User ID: " + userID,
				Breadcrumbs: []ui.BreadcrumbItem{
					{Label: "Admin", Href: "/admin"},
					{Label: "Users", Href: "/admin/users"},
					{Label: userName},
				},
				Actions: []core.Node{
					ui.Button(ui.ButtonProps{
						Variant: ui.ButtonOutline,
						Children: []core.Node{
							ui.Icon(ui.IconProps{Name: "edit", Size: ui.IconSM, Class: "mr-2"}),
							core.Text("Edit"),
						},
					}),
					ui.Button(ui.ButtonProps{
						Variant: ui.ButtonDestructive,
						Children: []core.Node{
							ui.Icon(ui.IconProps{Name: "trash", Size: ui.IconSM, Class: "mr-2"}),
							core.Text("Delete"),
						},
					}),
				},
			}),

			// User details
			core.Div(core.Class("grid grid-cols-1 lg:grid-cols-2 gap-6 mt-6"),
				// Profile card
				ui.Card(ui.CardProps{
					Class: "bg-gray-800 border-gray-700",
					Children: []core.Node{
						ui.CardHeader(ui.CardHeaderProps{
							Children: []core.Node{
								core.H3(core.Class("text-lg font-semibold text-white"), core.Text("Profile Information")),
							},
						}),
						ui.CardContent(ui.CardContentProps{
							Children: []core.Node{
								detailRow("Name", userName),
								detailRow("Email", userEmail),
								detailRow("Role", userRole),
								detailRow("User ID", userID),
								detailRow("Created", "Jan 15, 2024"),
								detailRow("Last Login", "2 hours ago"),
							},
						}),
					},
				}),

				// Activity card
				ui.Card(ui.CardProps{
					Class: "bg-gray-800 border-gray-700",
					Children: []core.Node{
						ui.CardHeader(ui.CardHeaderProps{
							Children: []core.Node{
								core.H3(core.Class("text-lg font-semibold text-white"), core.Text("Recent Activity")),
							},
						}),
						ui.CardContent(ui.CardContentProps{
							Children: []core.Node{
								activityRow("Logged in", "2 hours ago"),
								activityRow("Updated profile", "1 day ago"),
								activityRow("Changed password", "1 week ago"),
								activityRow("Account created", "Jan 15, 2024"),
							},
						}),
					},
				}),
			),
		)
	}
}

func detailRow(label, value string) core.Node {
	return core.Div(core.Class("flex justify-between py-3 border-b border-gray-700 last:border-0"),
		core.Span(core.Class("text-gray-400"), core.Text(label)),
		core.Span(core.Class("text-white"), core.Text(value)),
	)
}

func activityRow(action, time string) core.Node {
	return core.Div(core.Class("flex justify-between py-3 border-b border-gray-700 last:border-0"),
		core.Span(core.Class("text-white"), core.Text(action)),
		core.Span(core.Class("text-gray-500 text-sm"), core.Text(time)),
	)
}
