package admin

import (
	"encoding/json"
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/.gux/api"
	"github.com/dougbarrett/gux/examples/minimal/dto"
)

// PostEdit is the form for editing an existing post.
func PostEdit(r *core.Router) func() core.Node {
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
				core.Div(core.Class("max-w-2xl mx-auto px-4 py-8"),
					core.Div(core.Class("text-white text-center py-12"),
						core.Text("Post not found"),
					),
				),
			)
		}

		// Form state - initialize with current values
		titleState := r.StateString("title", displayPost.Title)
		contentState := r.StateString("content", displayPost.Content)
		errorState := r.StateString("error", "")
		successState := r.StateString("success", "")

		var messageNode core.Node = core.Frag()
		if successState.Get() != "" {
			messageNode = core.Div(core.Class("mb-4 p-4 bg-green-900 border border-green-700 rounded-lg text-green-300"),
				core.Text(successState.Get()),
			)
		} else if errorState.Get() != "" {
			messageNode = core.Div(core.Class("mb-4 p-4 bg-red-900 border border-red-700 rounded-lg text-red-300"),
				core.Text(errorState.Get()),
			)
		}

		// Save function - shared by button click and Enter key
		save := func() {
			// Validate
			if titleState.Get() == "" || contentState.Get() == "" {
				errorState.Set("Title and content are required")
				return
			}

			// Update post via generated API
			data := map[string]interface{}{
				"title":   titleState.Get(),
				"content": contentState.Get(),
			}
			api.Posts.Update(displayPost.ID, data, func(result *dto.PostDetail, err error) {
				if err != nil {
					errorState.Set(err.Error())
				} else {
					successState.Set("Post updated successfully!")
					errorState.Set("")
					// Navigate back to post detail
					r.Navigate(fmt.Sprintf("/admin/posts/%d", displayPost.ID))
				}
			})
		}

		return core.Div(core.Class("min-h-screen bg-gray-900"),
			Nav(),
			core.Div(core.Class("max-w-2xl mx-auto px-4 py-8"),
				// Back link
				core.A(
					core.Attrs{Href: fmt.Sprintf("/admin/posts/%d", displayPost.ID), Class: "text-blue-400 hover:text-blue-300 mb-4 inline-block"},
					core.Text("← Back to Post"),
				),

				core.H1(core.Class("text-2xl font-bold text-white mb-6"),
					core.Text(fmt.Sprintf("Edit Post: %s", displayPost.Title)),
				),

				messageNode,

				// Form
				core.Div(core.Class("bg-gray-800 rounded-lg p-6"),
					// Title field
					formField("Title", "text", "title", titleState.Get(), func(v string) {
						titleState.Set(v)
					}, save),

					// Content field (textarea - no OnEnter since Enter creates newlines)
					core.Div(core.Class("mb-4"),
						core.Label(core.Class("block text-gray-300 text-sm font-medium mb-2"),
							core.Text("Content"),
						),
						core.Textarea(
							core.Attrs{
								Name:  "content",
								Class: "w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:border-blue-500 h-32",
								OnChange: func(v string) {
									contentState.Set(v)
								},
							},
							core.Text(contentState.Get()),
						),
					),

					// Submit button
					core.Button(
						core.Attrs{
							Type:    "button",
							Class:   "w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition",
							OnClick: save,
						},
						core.Text("Save Changes"),
					),
				),
			),
		)
	}
}
