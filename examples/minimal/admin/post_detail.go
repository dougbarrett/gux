package admin

import (
	"encoding/json"
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/.gux/api"
	"github.com/dougbarrett/gux/examples/minimal/dto"
)

// PostDetail shows a single post with its content.
func PostDetail(r *core.Router) func() core.Node {
	var post *dto.PostDetail

	r.OnLoad(func() {
		// Get post ID from route params
		idStr := r.GetRouteParams()["id"]
		var id uint
		fmt.Sscanf(idStr, "%d", &id)

		if id > 0 {
			api.Posts.Get(id, func(result *dto.PostDetail, err error) {
				if err == nil {
					post = result
				}
			})
		}
	})

	return func() core.Node {
		// Store post in state for hydration
		postJSON, _ := json.Marshal(post)
		postState := r.StateString("post", string(postJSON))

		// Parse post from state
		var displayPost *dto.PostDetail
		json.Unmarshal([]byte(postState.Get()), &displayPost)

		if displayPost == nil {
			return core.Div(core.Class("min-h-screen bg-gray-900"),
				Nav(),
				core.Div(core.Class("max-w-6xl mx-auto px-4 py-8"),
					core.Div(core.Class("text-white text-center py-12"),
						core.Text("Post not found"),
					),
				),
			)
		}

		return core.Div(core.Class("min-h-screen bg-gray-900"),
			Nav(),
			core.Div(core.Class("max-w-4xl mx-auto px-4 py-8"),
				// Back link
				core.A(
					core.Attrs{Href: "/admin/posts", Class: "text-blue-400 hover:text-blue-300 mb-4 inline-block"},
					core.Text("← Back to Posts"),
				),

				// Post info card
				core.Div(core.Class("bg-gray-800 rounded-lg p-6"),
					core.Div(core.Class("flex justify-between items-start mb-6"),
						core.Div(core.Attrs{},
							core.H1(core.Class("text-2xl font-bold text-white mb-2"),
								core.Text(displayPost.Title),
							),
							core.Div(core.Class("flex items-center gap-4 text-gray-400 text-sm"),
								core.Span(core.Attrs{},
									core.Text(fmt.Sprintf("By %s", displayPost.Author.Name)),
								),
								core.Span(core.Attrs{},
									core.Text(formatTime(displayPost.CreatedAt)),
								),
							),
						),
						core.A(
							core.Attrs{
								Href:  fmt.Sprintf("/admin/posts/%d/edit", displayPost.ID),
								Class: "px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition",
							},
							core.Text("Edit"),
						),
					),

					// Content
					core.Div(core.Class("prose prose-invert max-w-none"),
						core.Div(core.Class("text-gray-300 whitespace-pre-wrap"),
							core.Text(displayPost.Content),
						),
					),

					// Metadata
					core.Div(core.Class("mt-6 pt-6 border-t border-gray-700"),
						core.Div(core.Class("flex gap-6 text-sm text-gray-500"),
							core.Span(core.Attrs{},
								core.Text(fmt.Sprintf("Post ID: %d", displayPost.ID)),
							),
							core.A(
								core.Attrs{
									Href:  fmt.Sprintf("/admin/users/%d", displayPost.UserID),
									Class: "text-blue-400 hover:text-blue-300",
								},
								core.Text(fmt.Sprintf("Author: %s", displayPost.Author.Name)),
							),
							core.Span(core.Attrs{},
								core.Text(fmt.Sprintf("Updated: %s", formatTime(displayPost.UpdatedAt))),
							),
						),
					),
				),
			),
		)
	}
}
