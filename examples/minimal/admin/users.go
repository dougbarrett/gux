package admin

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/.gux/api"
	"github.com/dougbarrett/gux/examples/minimal/dto"
)

// Users is the admin users management page.
func Users(r *core.Router) func() core.Node {
	var users []dto.UserList

	r.OnLoad(func() {
		// Fetch users via generated API client
		// Server-side: queries DB directly with field mappings from gux tags
		// WASM: calls CRUD endpoint which handles DTO conversion
		api.Users.List(func(result []dto.UserList, err error) {
			if err == nil {
				users = result
			}
		})
	})

	return func() core.Node {
		// Store users in state for hydration
		usersJSON, _ := json.Marshal(users)
		usersState := r.StateString("users", string(usersJSON))

		// Parse users from state (for client-side)
		var displayUsers []dto.UserList
		json.Unmarshal([]byte(usersState.Get()), &displayUsers)

		return core.Div(core.Class("min-h-screen bg-gray-900"),
			Nav(),
			core.Div(core.Class("max-w-6xl mx-auto px-4 py-8"),
				core.H1(core.Class("text-3xl font-bold text-white mb-8"),
					core.Text("User Management"),
				),

				// Add User Button
				core.Div(core.Class("mb-4"),
					core.A(
						core.Attrs{Href: "/admin/users/new", Class: "inline-flex items-center px-4 py-2 bg-green-600 hover:bg-green-700 text-white font-medium rounded-lg transition"},
						core.Text("+ Add User"),
					),
				),

				// Users table
				core.Div(core.Class("bg-gray-800 rounded-lg overflow-hidden"),
					core.Table(core.Class("w-full"),
						core.Thead(core.Class("bg-gray-700"),
							core.Tr(core.Attrs{},
								tableHeader("ID"),
								tableHeader("Name"),
								tableHeader("Email"),
								tableHeader("Role"),
								tableHeader("Created"),
								tableHeader("Actions"),
							),
						),
						core.Tbody(core.Attrs{},
							userRowsNode(displayUsers),
						),
					),
				),

				// Note about DTOs
				core.Div(core.Class("mt-6 p-4 bg-gray-800 rounded-lg border border-gray-700"),
					core.Div(core.Class("text-green-400 font-semibold mb-2"),
						core.Text("DTO Security in Action"),
					),
					core.Div(core.Class("text-gray-400 text-sm"),
						core.Text("Notice: The PasswordHash field is NOT included in this response. "),
						core.Text("The API uses UserList DTO which only exposes safe fields (mapped via gux tags)."),
					),
				),
			),
		)
	}
}

func tableHeader(text string) core.Node {
	return core.Th(core.Class("px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider"),
		core.Text(text),
	)
}

func userRowsNode(users []dto.UserList) core.Node {
	if len(users) == 0 {
		return core.Tr(core.Attrs{},
			core.Td(core.Class("px-6 py-4 text-gray-400 text-center"),
				core.Text("No users found"),
			),
		)
	}

	var rows []core.Node
	for _, user := range users {
		rows = append(rows, core.Tr(core.Class("border-t border-gray-700 hover:bg-gray-750"),
			tableCell(fmt.Sprintf("%d", user.ID)),
			tableCellLink(user.Name, fmt.Sprintf("/admin/users/%d", user.ID)),
			tableCell(user.Email),
			roleCell(user.Role),
			tableCell(formatTime(user.CreatedAt)),
			actionsCell(user.ID),
		))
	}
	return core.Frag(rows...)
}

func tableCellLink(text, href string) core.Node {
	return core.Td(core.Class("px-6 py-4 whitespace-nowrap text-sm"),
		core.A(
			core.Attrs{Href: href, Class: "text-blue-400 hover:text-blue-300 font-medium"},
			core.Text(text),
		),
	)
}

func actionsCell(userID uint) core.Node {
	return core.Td(core.Class("px-6 py-4 whitespace-nowrap text-sm"),
		core.A(
			core.Attrs{Href: fmt.Sprintf("/admin/users/%d", userID), Class: "text-blue-400 hover:text-blue-300 mr-3"},
			core.Text("View"),
		),
	)
}

func tableCell(text string) core.Node {
	return core.Td(core.Class("px-6 py-4 whitespace-nowrap text-sm text-gray-300"),
		core.Text(text),
	)
}

func roleCell(role string) core.Node {
	colorClass := "bg-gray-600"
	if role == "admin" {
		colorClass = "bg-purple-600"
	} else if role == "user" {
		colorClass = "bg-blue-600"
	}

	return core.Td(core.Class("px-6 py-4 whitespace-nowrap"),
		core.Span(core.Class("px-2 py-1 text-xs font-medium rounded "+colorClass+" text-white"),
			core.Text(role),
		),
	)
}

func formatTime(t time.Time) string {
	return t.Format("Jan 2, 2006")
}
