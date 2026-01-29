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

// Users lists all users.
func Users(r *core.Router) func() core.Node {
	var users []dto.UserList

	r.OnLoad(func() {
		api.Users.List(func(result []dto.UserList, err error) {
			if err == nil {
				users = result
			}
		})
	})

	return func() core.Node {
		usersJSON, _ := json.Marshal(users)
		usersState := r.StateString("users", string(usersJSON))

		var displayUsers []dto.UserList
		json.Unmarshal([]byte(usersState.Get()), &displayUsers)

		return pages.AdminLayout(r, "Users",
			core.H1(core.Class("text-3xl font-bold text-white mb-6"),
				core.Text("Users"),
			),

			ui.Card(ui.CardProps{
				Class: "bg-gray-800 border border-gray-700",
				Children: []core.Node{
					ui.CardContent(ui.CardContentProps{
						Class: "p-0",
						Children: []core.Node{
							ui.DataTable(ui.DataTableProps[dto.UserList]{
								Data:      displayUsers,
								Striped:   true,
								Hoverable: true,
								Class:     "bg-gray-800",
								Columns: []ui.ColumnDef[dto.UserList]{
									{
										Header: "ID",
										Render: func(u dto.UserList) core.Node {
											return core.Span(core.Class("text-gray-400"),
												core.Text(fmt.Sprintf("%d", u.ID)),
											)
										},
									},
									{
										Header: "Name",
										Render: func(u dto.UserList) core.Node {
											return core.A(
												core.Attrs{
													Href:     fmt.Sprintf("/admin/users/%d", u.ID),
													Class:    "text-blue-400 hover:text-blue-300 font-medium",
													External: true,
												},
												core.Text(u.Name),
											)
										},
									},
									{
										Header: "Email",
										Render: func(u dto.UserList) core.Node {
											return core.Span(core.Class("text-gray-300"),
												core.Text(u.Email),
											)
										},
									},
									{
										Header: "Role",
										Render: func(u dto.UserList) core.Node {
											var variant ui.BadgeVariant
											switch u.Role {
											case "admin":
												variant = ui.BadgeError
											default:
												variant = ui.BadgeDefault
											}
											return ui.Badge(ui.BadgeProps{
												Variant: variant,
												Text:    capitalizeFirst(u.Role),
											})
										},
									},
									{
										Header: "Joined",
										Render: func(u dto.UserList) core.Node {
											return core.Span(core.Class("text-gray-400"),
												core.Text(u.CreatedAt.Format("Jan 2, 2006")),
											)
										},
									},
								},
							}),
						},
					}),
				},
			}),
		)
	}
}
