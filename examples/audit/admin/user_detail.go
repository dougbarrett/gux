package admin

import (
	"encoding/json"
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/audit/dto"
	"github.com/dougbarrett/gux/examples/audit/guxgen/api"
	"github.com/dougbarrett/gux/examples/audit/pages"
	"github.com/dougbarrett/gux/ui"
)

// UserDetail shows a single user's details.
func UserDetail(r *core.Router) func() core.Node {
	idStr := r.Param("id")
	var user *dto.UserDetail

	r.OnLoad(func() {
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
		userJSON, _ := json.Marshal(user)
		userState := r.StateString("user", string(userJSON))

		var displayUser *dto.UserDetail
		json.Unmarshal([]byte(userState.Get()), &displayUser)

		if displayUser == nil || displayUser.ID == 0 {
			return pages.AdminLayout(r, "User",
				core.Div(core.Class("text-center text-gray-400 py-12"),
					core.Text("User not found"),
				),
			)
		}

		return pages.AdminLayout(r, "User Detail",
			core.A(
				core.Attrs{Href: "/admin/users", Class: "text-blue-400 hover:text-blue-300 mb-4 inline-block", External: true},
				core.Text("\u2190 Back to Users"),
			),

			ui.Card(ui.CardProps{
				Class: "bg-gray-800 border border-gray-700",
				Children: []core.Node{
					ui.CardHeader(ui.CardHeaderProps{
						Class: "border-gray-700",
						Children: []core.Node{
							core.Div(core.Class("flex justify-between items-center"),
								core.Div(core.Class("flex items-center gap-4"),
									ui.Avatar(ui.AvatarProps{
										Name: displayUser.Name,
										Size: ui.AvatarLG,
									}),
									core.Div(core.Attrs{},
										core.H1(core.Class("text-2xl font-bold text-white"),
											core.Text(displayUser.Name),
										),
										core.Span(core.Class("text-gray-400"),
											core.Text(displayUser.Email),
										),
									),
								),
								func() core.Node {
									var variant ui.BadgeVariant
									if displayUser.Role == "admin" {
										variant = ui.BadgeError
									} else {
										variant = ui.BadgeDefault
									}
									return ui.Badge(ui.BadgeProps{
										Variant: variant,
										Text:    capitalizeFirst(displayUser.Role),
									})
								}(),
							),
						},
					}),
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							detailRow("ID", fmt.Sprintf("%d", displayUser.ID)),
							detailRow("Email", displayUser.Email),
							detailRow("Name", displayUser.Name),
							detailRow("Role", capitalizeFirst(displayUser.Role)),
							detailRow("Created", displayUser.CreatedAt.Format("January 2, 2006 at 3:04 PM")),
							detailRow("Updated", displayUser.UpdatedAt.Format("January 2, 2006 at 3:04 PM")),
						},
					}),
				},
			}),
		)
	}
}
