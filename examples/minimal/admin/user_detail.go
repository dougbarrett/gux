package admin

import (
	"encoding/json"
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/.gux/api"
	"github.com/dougbarrett/gux/examples/minimal/dto"
)

// UserDetail shows a single user with their posts.
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

		if displayUser == nil {
			return core.Div(core.Class("min-h-screen bg-gray-900"),
				Nav(),
				core.Div(core.Class("max-w-6xl mx-auto px-4 py-8"),
					core.Div(core.Class("text-white text-center py-12"),
						core.Text("User not found"),
					),
				),
			)
		}

		return core.Div(core.Class("min-h-screen bg-gray-900"),
			Nav(),
			core.Div(core.Class("max-w-6xl mx-auto px-4 py-8"),
				// Back link
				core.A(
					core.Attrs{Href: "/admin/users", Class: "text-blue-400 hover:text-blue-300 mb-4 inline-block"},
					core.Text("← Back to Users"),
				),

				// User info card
				core.Div(core.Class("bg-gray-800 rounded-lg p-6 mb-6"),
					core.Div(core.Class("flex justify-between items-start"),
						core.Div(core.Attrs{},
							core.H1(core.Class("text-2xl font-bold text-white mb-2"),
								core.Text(displayUser.Name),
							),
							core.Div(core.Class("text-gray-400 mb-1"),
								core.Text(displayUser.Email),
							),
							core.Div(core.Class("flex items-center gap-2"),
								roleCell(displayUser.Role),
								core.Span(core.Class("text-gray-500 text-sm"),
									core.Text(fmt.Sprintf("Member since %s", formatTime(displayUser.CreatedAt))),
								),
							),
						),
						core.A(
							core.Attrs{
								Href:  fmt.Sprintf("/admin/users/%d/edit", displayUser.ID),
								Class: "px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition",
							},
							core.Text("Edit"),
						),
					),
				),

				// Posts section
				core.Div(core.Class("bg-gray-800 rounded-lg p-6"),
					core.Div(core.Class("flex justify-between items-center mb-4"),
						core.H2(core.Class("text-xl font-bold text-white"),
							core.Text(fmt.Sprintf("Posts (%d)", len(displayUser.Posts))),
						),
						core.A(
							core.Attrs{
								Href:  fmt.Sprintf("/admin/users/%d/posts/new", displayUser.ID),
								Class: "px-4 py-2 bg-green-600 hover:bg-green-700 text-white font-medium rounded-lg transition",
							},
							core.Text("+ Add Post"),
						),
					),

					// Posts list
					postsListNode(displayUser.Posts),
				),
			),
		)
	}
}

func postsListNode(posts []dto.PostBrief) core.Node {
	if len(posts) == 0 {
		return core.Div(core.Class("text-gray-400 py-8 text-center"),
			core.Text("No posts yet. Create the first one!"),
		)
	}

	var items []core.Node
	for _, post := range posts {
		items = append(items, core.Div(core.Class("border-b border-gray-700 py-4 last:border-0"),
			core.Div(core.Class("flex justify-between items-start"),
				core.Div(core.Attrs{},
					core.H3(core.Class("text-white font-medium"),
						core.Text(post.Title),
					),
					core.Div(core.Class("text-gray-500 text-sm"),
						core.Text(formatTime(post.CreatedAt)),
					),
				),
			),
		))
	}
	return core.Frag(items...)
}
