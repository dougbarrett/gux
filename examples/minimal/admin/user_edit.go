package admin

import (
	"encoding/json"
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/guxgen/api"
	"github.com/dougbarrett/gux/examples/minimal/dto"
)

// UserEdit is the form for editing an existing user.
func UserEdit(r *core.Router) func() core.Node {
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
				core.Div(core.Class("max-w-2xl mx-auto px-4 py-8"),
					core.Div(core.Class("text-white text-center py-12"),
						core.Text("User not found"),
					),
				),
			)
		}

		// Form state - initialize with current values
		nameState := r.StateString("name", displayUser.Name)
		emailState := r.StateString("email", displayUser.Email)
		passwordState := r.StateString("password", "")
		roleState := r.StateString("role", displayUser.Role)
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
			// Validate required fields
			if nameState.Get() == "" || emailState.Get() == "" {
				errorState.Set("Name and email are required")
				return
			}

			// Build update data
			data := map[string]interface{}{
				"name":  nameState.Get(),
				"email": emailState.Get(),
				"role":  roleState.Get(),
			}
			// Only include password if provided
			if passwordState.Get() != "" {
				data["password"] = passwordState.Get()
			}

			api.Users.Update(displayUser.ID, data, func(result *dto.UserDetail, err error) {
				if err != nil {
					errorState.Set(err.Error())
				} else {
					successState.Set("User updated successfully!")
					errorState.Set("")
					// Navigate back to user detail
					r.Navigate(fmt.Sprintf("/admin/users/%d", displayUser.ID))
				}
			})
		}

		return core.Div(core.Class("min-h-screen bg-gray-900"),
			Nav(),
			core.Div(core.Class("max-w-2xl mx-auto px-4 py-8"),
				// Back link
				core.A(
					core.Attrs{Href: fmt.Sprintf("/admin/users/%d", displayUser.ID), Class: "text-blue-400 hover:text-blue-300 mb-4 inline-block"},
					core.Text("← Back to User"),
				),

				core.H1(core.Class("text-2xl font-bold text-white mb-6"),
					core.Text(fmt.Sprintf("Edit User: %s", displayUser.Name)),
				),

				messageNode,

				// Form
				core.Div(core.Class("bg-gray-800 rounded-lg p-6"),
					// Name field
					formField("Name", "text", "name", nameState.Get(), func(v string) {
						nameState.Set(v)
					}, save),

					// Email field
					formField("Email", "email", "email", emailState.Get(), func(v string) {
						emailState.Set(v)
					}, save),

					// Password field (optional for edit)
					core.Div(core.Class("mb-4"),
						core.Label(core.Class("block text-gray-300 text-sm font-medium mb-2"),
							core.Text("Password (leave blank to keep current)"),
						),
						core.Input(
							core.Attrs{
								Type:     "password",
								Name:     "password",
								Value:    passwordState.Get(),
								Class:    "w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:border-blue-500",
								OnChange: func(v string) { passwordState.Set(v) },
								OnEnter:  save,
							},
						),
					),

					// Role select
					core.Div(core.Class("mb-4"),
						core.Label(core.Class("block text-gray-300 text-sm font-medium mb-2"),
							core.Text("Role"),
						),
						core.Select(
							core.Attrs{
								Name:  "role",
								Class: "w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:border-blue-500",
								OnChange: func(v string) {
									roleState.Set(v)
								},
							},
							selectOption("user", "User", roleState.Get()),
							selectOption("admin", "Admin", roleState.Get()),
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

// selectOption creates an option element with selected state
func selectOption(value, label, selectedValue string) core.Node {
	attrs := core.Attrs{Value: value}
	if value == selectedValue {
		attrs.Extra = map[string]string{"selected": "selected"}
	}
	return core.Option(attrs, core.Text(label))
}
