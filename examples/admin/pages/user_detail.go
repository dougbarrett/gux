package pages

import (
	"encoding/json"
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/admin/.gux/api"
	"github.com/dougbarrett/gux/examples/admin/dto"
	"github.com/dougbarrett/gux/ui"
)

// UserDetail shows a single user with edit/delete options.
func UserDetail(r *core.Router) func() core.Node {
	var user *dto.UserDetail

	r.OnLoad(func() {
		// Get user ID from route params
		idStr := r.GetRouteParams()["id"]
		var id uint
		fmt.Sscanf(idStr, "%d", &id)

		if id > 0 {
			api.Users.Get(id, func(result *dto.UserDetail, err error) {
				if err == nil {
					user = result
				}
			})
		}
	})

	return func() core.Node {
		// Store user in state for hydration
		userJSON, _ := json.Marshal(user)
		userState := r.StateString("user", string(userJSON))

		// Parse user from state
		var displayUser *dto.UserDetail
		json.Unmarshal([]byte(userState.Get()), &displayUser)

		// Delete modal state
		showDeleteModal := r.StateBool("showDeleteModal", false)

		if displayUser == nil || displayUser.ID == 0 {
			return AdminLayout(
				core.Div(core.Class("text-white text-center py-12"),
					core.Text("User not found"),
				),
			)
		}

		return AdminLayout(
			// Back link
			core.A(
				core.Attrs{Href: "/users", Class: "text-blue-400 hover:text-blue-300 mb-4 inline-block"},
				core.Text("\u2190 Back to Users"),
			),

			// User info card
			ui.Card(ui.CardProps{
				Class: "bg-gray-800 border border-gray-700 mb-6",
				Children: []core.Node{
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							core.Div(core.Class("flex justify-between items-start"),
								core.Div(core.Attrs{},
									core.H1(core.Class("text-2xl font-bold text-white mb-2"),
										core.Text(displayUser.Name),
									),
									core.Div(core.Class("text-gray-400 mb-4"),
										core.Text(displayUser.Email),
									),
									core.Div(core.Class("flex items-center gap-4"),
										roleBadge(displayUser.Role),
										userStatusBadge(displayUser.Status),
									),
								),
								core.Div(core.Class("flex gap-2"),
									core.A(
										core.Attrs{
											Href:  fmt.Sprintf("/users/%d/edit", displayUser.ID),
											Class: "px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition",
										},
										core.Text("Edit"),
									),
									core.Button(
										core.Attrs{
											Type:  "button",
											Class: "px-4 py-2 bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg transition",
											OnClick: func() {
												showDeleteModal.Set(true)
											},
										},
										core.Text("Delete"),
									),
								),
							),
						},
					}),
				},
			}),

			// User details card
			ui.Card(ui.CardProps{
				Class: "bg-gray-800 border border-gray-700",
				Children: []core.Node{
					ui.CardHeader(ui.CardHeaderProps{
						Class: "border-gray-700",
						Children: []core.Node{
							core.H2(core.Class("text-lg font-semibold text-white"),
								core.Text("User Details"),
							),
						},
					}),
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							userDetailRow("Email", displayUser.Email),
							userDetailRow("Role", capitalizeFirst(displayUser.Role)),
							userDetailRow("Status", capitalizeFirst(displayUser.Status)),
							userDetailRow("Last Login", func() string {
								if displayUser.LastLoginAt == nil {
									return "Never"
								}
								return displayUser.LastLoginAt.Format("Jan 2, 2006 at 3:04 PM")
							}()),
							userDetailRow("Created", displayUser.CreatedAt.Format("Jan 2, 2006 at 3:04 PM")),
							userDetailRow("Updated", displayUser.UpdatedAt.Format("Jan 2, 2006 at 3:04 PM")),
						},
					}),
				},
			}),

			// Delete confirmation modal
			userDeleteModal(r, displayUser, showDeleteModal),
		)
	}
}

// userDetailRow renders a label-value pair for user details.
func userDetailRow(label, value string) core.Node {
	return core.Div(core.Class("flex py-3 border-b border-gray-700 last:border-0"),
		core.Span(core.Class("w-32 text-gray-400 font-medium"),
			core.Text(label),
		),
		core.Span(core.Class("text-white"),
			core.Text(value),
		),
	)
}

// userDeleteModal renders the delete confirmation modal.
func userDeleteModal(r *core.Router, user *dto.UserDetail, openState *core.State[bool]) core.Node {
	return ui.Modal(ui.ModalProps{
		Open:  openState.Get(),
		Title: "Delete User",
		Size:  ui.ModalSM,
		OnClose: func() {
			openState.Set(false)
		},
		Children: []core.Node{
			ui.ModalContent(ui.ModalContentProps{
				Children: []core.Node{
					core.P(core.Class("text-gray-600 dark:text-gray-300"),
						core.Text(fmt.Sprintf("Are you sure you want to delete \"%s\"? This action cannot be undone.", user.Name)),
					),
				},
			}),
			ui.ModalFooter(ui.ModalFooterProps{
				Children: []core.Node{
					core.Button(
						core.Attrs{
							Type:  "button",
							Class: "px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white font-medium rounded-lg transition",
							OnClick: func() {
								openState.Set(false)
							},
						},
						core.Text("Cancel"),
					),
					core.Button(
						core.Attrs{
							Type:  "button",
							Class: "px-4 py-2 bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg transition",
							OnClick: func() {
								api.Users.Delete(user.ID, func(err error) {
									if err == nil {
										r.Navigate("/users")
									}
								})
							},
						},
						core.Text("Delete"),
					),
				},
			}),
		},
	})
}
