package admin

import (
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/.gux/api"
	"github.com/dougbarrett/gux/examples/minimal/dto"
)

// PostNew is the form for creating a new post for a user.
func PostNew(r *core.Router) func() core.Node {
	// Get user ID from route params
	var userID uint
	idStr := r.GetRouteParams()["id"]
	fmt.Sscanf(idStr, "%d", &userID)

	return func() core.Node {
		// Form state
		titleState := r.StateString("title", "")
		contentState := r.StateString("content", "")
		errorState := r.StateString("error", "")
		successState := r.StateString("success", "")
		userIDState := r.StateInt("userId", int(userID))

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

		backURL := fmt.Sprintf("/admin/users/%d", userIDState.Get())

		// Save function - shared by button click and Enter key
		save := func() {
			// Validate
			if titleState.Get() == "" || contentState.Get() == "" {
				errorState.Set("Title and content are required")
				return
			}

			// Create post via generated API
			data := map[string]interface{}{
				"title":   titleState.Get(),
				"content": contentState.Get(),
				"user_id": userIDState.Get(),
			}
			api.Posts.Create(data, func(_ *dto.PostDetail, err error) {
				if err != nil {
					errorState.Set(err.Error())
				} else {
					successState.Set("Post created successfully!")
					// Clear form
					titleState.Set("")
					contentState.Set("")
					errorState.Set("")
					// Navigate back to user detail
					r.Navigate(backURL)
				}
			})
		}

		return core.Div(core.Class("min-h-screen bg-gray-900"),
			Nav(),
			core.Div(core.Class("max-w-2xl mx-auto px-4 py-8"),
				// Back link
				core.A(
					core.Attrs{Href: backURL, Class: "text-blue-400 hover:text-blue-300 mb-4 inline-block"},
					core.Text("← Back to User"),
				),

				core.H1(core.Class("text-2xl font-bold text-white mb-6"),
					core.Text("Create New Post"),
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
							Class:   "w-full px-4 py-2 bg-green-600 hover:bg-green-700 text-white font-medium rounded-lg transition",
							OnClick: save,
						},
						core.Text("Create Post"),
					),
				),
			),
		)
	}
}
