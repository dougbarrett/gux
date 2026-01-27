package pages

import (
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Users renders the users list page with breadcrumbs.
func Users(r *core.Router) func() core.Node {
	return func() core.Node {
		// Sample user data (in real app, fetch from API)
		users := []struct {
			ID    uint
			Name  string
			Email string
			Role  string
		}{
			{1, "Admin User", "admin@example.com", "admin"},
			{2, "John Doe", "john@example.com", "user"},
			{3, "Jane Smith", "jane@example.com", "user"},
			{4, "Bob Wilson", "bob@example.com", "user"},
		}

		// Build user rows
		userRows := make([]core.Node, len(users))
		for i, user := range users {
			userRows[i] = ui.Tr(ui.TrProps{
				Class: "hover:bg-gray-700",
				Children: []core.Node{
					ui.Td(ui.TdProps{
						Class:    "text-white",
						Children: []core.Node{core.Text(fmt.Sprintf("%d", user.ID))},
					}),
					ui.Td(ui.TdProps{
						Children: []core.Node{
							core.A(core.Attrs{
								Href:  fmt.Sprintf("/admin/users/%d", user.ID),
								Class: "text-blue-400 hover:text-blue-300",
							}, core.Text(user.Name)),
						},
					}),
					ui.Td(ui.TdProps{
						Class:    "text-gray-300",
						Children: []core.Node{core.Text(user.Email)},
					}),
					ui.Td(ui.TdProps{
						Children: []core.Node{
							ui.Badge(ui.BadgeProps{
								Text: user.Role,
								Variant: func() ui.BadgeVariant {
									if user.Role == "admin" {
										return ui.BadgeDefault
									}
									return ui.BadgeDefault
								}(),
							}),
						},
					}),
					ui.Td(ui.TdProps{
						Children: []core.Node{
							core.A(core.Attrs{
								Href:  fmt.Sprintf("/admin/users/%d", user.ID),
								Class: "text-gray-400 hover:text-white",
							},
								ui.Icon(ui.IconProps{Name: "eye", Size: ui.IconSM}),
							),
						},
					}),
				},
			})
		}

		return core.Div(core.Class("p-6"),
			// Page header with breadcrumbs and action button
			ui.PageHeader(ui.PageHeaderProps{
				Title:    "Users",
				Subtitle: "Manage system users",
				Breadcrumbs: []ui.BreadcrumbItem{
					{Label: "Admin", Href: "/admin"},
					{Label: "Users"},
				},
				Actions: []core.Node{
					ui.Button(ui.ButtonProps{
						Variant: ui.ButtonPrimary,
						Children: []core.Node{
							ui.Icon(ui.IconProps{Name: "plus", Size: ui.IconSM, Class: "mr-2"}),
							core.Text("Add User"),
						},
					}),
				},
			}),

			// Users table
			core.Div(core.Class("mt-6"),
				ui.Card(ui.CardProps{
					Class: "bg-gray-800 border-gray-700",
					Children: []core.Node{
						ui.CardContent(ui.CardContentProps{
							Class: "p-0",
							Children: []core.Node{
								ui.Table(ui.TableProps{
									Class: "text-gray-300",
									Children: []core.Node{
										ui.Thead(ui.TheadProps{
											Class: "bg-gray-900",
											Children: []core.Node{
												ui.Tr(ui.TrProps{
													Children: []core.Node{
														ui.Th(ui.ThProps{Class: "text-gray-400", Children: []core.Node{core.Text("ID")}}),
														ui.Th(ui.ThProps{Class: "text-gray-400", Children: []core.Node{core.Text("Name")}}),
														ui.Th(ui.ThProps{Class: "text-gray-400", Children: []core.Node{core.Text("Email")}}),
														ui.Th(ui.ThProps{Class: "text-gray-400", Children: []core.Node{core.Text("Role")}}),
														ui.Th(ui.ThProps{Class: "text-gray-400", Children: []core.Node{core.Text("Actions")}}),
													},
												}),
											},
										}),
										ui.Tbody(ui.TbodyProps{
											Children: userRows,
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
