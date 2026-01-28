package admin

import (
	"encoding/json"
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/guxgen/api"
	"github.com/dougbarrett/gux/examples/minimal/dto"
)

// Posts is the admin posts management page.
func Posts(r *core.Router) func() core.Node {
	var posts []dto.PostList

	r.OnLoad(func() {
		api.Posts.List(func(result []dto.PostList, err error) {
			if err == nil {
				posts = result
			}
		})
	})

	return func() core.Node {
		// Store posts in state for hydration
		postsJSON, _ := json.Marshal(posts)
		postsState := r.StateString("posts", string(postsJSON))

		// Parse posts from state (for client-side)
		var displayPosts []dto.PostList
		json.Unmarshal([]byte(postsState.Get()), &displayPosts)

		return core.Div(core.Class("min-h-screen bg-gray-900"),
			Nav(),
			core.Div(core.Class("max-w-6xl mx-auto px-4 py-8"),
				core.H1(core.Class("text-3xl font-bold text-white mb-8"),
					core.Text("Posts Management"),
				),

				// Add Post Button
				core.Div(core.Class("mb-4"),
					core.A(
						core.Attrs{Href: "/admin/posts/new", Class: "inline-flex items-center px-4 py-2 bg-green-600 hover:bg-green-700 text-white font-medium rounded-lg transition"},
						core.Text("+ Add Post"),
					),
				),

				// Posts table
				core.Div(core.Class("bg-gray-800 rounded-lg overflow-hidden"),
					core.Table(core.Class("w-full"),
						core.Thead(core.Class("bg-gray-700"),
							core.Tr(core.Attrs{},
								tableHeader("ID"),
								tableHeader("Title"),
								tableHeader("Author"),
								tableHeader("Created"),
								tableHeader("Actions"),
							),
						),
						core.Tbody(core.Attrs{},
							postRowsNode(displayPosts),
						),
					),
				),
			),
		)
	}
}

func postRowsNode(posts []dto.PostList) core.Node {
	if len(posts) == 0 {
		return core.Tr(core.Attrs{},
			core.Td(core.Class("px-6 py-4 text-gray-400 text-center"),
				core.Text("No posts found"),
			),
		)
	}

	var rows []core.Node
	for _, post := range posts {
		authorName := post.Author.Name
		if authorName == "" {
			authorName = fmt.Sprintf("User #%d", post.UserID)
		}
		rows = append(rows, core.Tr(core.Class("border-t border-gray-700 hover:bg-gray-750"),
			tableCell(fmt.Sprintf("%d", post.ID)),
			tableCellLink(post.Title, fmt.Sprintf("/admin/posts/%d", post.ID)),
			tableCellLink(authorName, fmt.Sprintf("/admin/users/%d", post.UserID)),
			tableCell(formatTime(post.CreatedAt)),
			postActionsCell(post.ID),
		))
	}
	return core.Frag(rows...)
}

func postActionsCell(postID uint) core.Node {
	return core.Td(core.Class("px-6 py-4 whitespace-nowrap text-sm"),
		core.A(
			core.Attrs{Href: fmt.Sprintf("/admin/posts/%d", postID), Class: "text-blue-400 hover:text-blue-300 mr-3"},
			core.Text("View"),
		),
		core.A(
			core.Attrs{Href: fmt.Sprintf("/admin/posts/%d/edit", postID), Class: "text-yellow-400 hover:text-yellow-300"},
			core.Text("Edit"),
		),
	)
}
