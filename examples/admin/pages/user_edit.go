package pages

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/admin/.gux/api"
	"github.com/dougbarrett/gux/examples/admin/dto"
	"github.com/dougbarrett/gux/ui"
)

// UserEdit allows editing an existing user.
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

		// Form field state
		name := r.StateString("name", func() string {
			if displayUser != nil {
				return displayUser.Name
			}
			return ""
		}())
		email := r.StateString("email", func() string {
			if displayUser != nil {
				return displayUser.Email
			}
			return ""
		}())
		role := r.StateString("role", func() string {
			if displayUser != nil {
				return displayUser.Role
			}
			return "user"
		}())
		status := r.StateString("status", func() string {
			if displayUser != nil {
				return displayUser.Status
			}
			return "active"
		}())
		password := r.StateString("password", "")

		// Message state
		errorMsg := r.StateString("error", "")
		successMsg := r.StateString("success", "")

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
				core.Attrs{Href: fmt.Sprintf("/users/%d", displayUser.ID), Class: "text-blue-400 hover:text-blue-300 mb-4 inline-block"},
				core.Text("\u2190 Back to User"),
			),

			core.H1(core.Class("text-2xl font-bold text-white mb-6"),
				core.Text("Edit User"),
			),

			// Edit form card
			ui.Card(ui.CardProps{
				Class: "bg-gray-800 border border-gray-700",
				Children: []core.Node{
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							// Error message
							func() core.Node {
								if errorMsg.Get() != "" {
									return core.Div(core.Class("bg-red-900/20 border border-red-800 text-red-400 rounded-lg px-4 py-3 mb-4"),
										core.Text(errorMsg.Get()),
									)
								}
								return core.Frag()
							}(),

							// Success message
							func() core.Node {
								if successMsg.Get() != "" {
									return core.Div(core.Class("bg-green-900/20 border border-green-800 text-green-400 rounded-lg px-4 py-3 mb-4"),
										core.Text(successMsg.Get()),
									)
								}
								return core.Frag()
							}(),

							ui.VStack(ui.StackProps{
								Gap: "4",
								Children: []core.Node{
									// Name field
									userFormField("Name", ui.Input(ui.InputProps{
										Type:        ui.InputText,
										Name:        "name",
										Value:       name.Get(),
										Placeholder: "Full name",
										OnChange:    func(v string) { name.Set(v) },
										Class:       "bg-gray-700 border-gray-600 text-white",
									})),

									// Email field
									userFormField("Email", ui.Input(ui.InputProps{
										Type:        ui.InputEmail,
										Name:        "email",
										Value:       email.Get(),
										Placeholder: "email@example.com",
										OnChange:    func(v string) { email.Set(v) },
										Class:       "bg-gray-700 border-gray-600 text-white",
									})),

									// Role select
									userFormField("Role", ui.Select(ui.SelectProps{
										Name:  "role",
										Value: role.Get(),
										Options: []ui.SelectOption{
											{Value: "admin", Label: "Admin"},
											{Value: "editor", Label: "Editor"},
											{Value: "user", Label: "User"},
										},
										OnChange: func(v string) { role.Set(v) },
										Class:    "bg-gray-700 border-gray-600 text-white",
									})),

									// Status select
									userFormField("Status", ui.Select(ui.SelectProps{
										Name:  "status",
										Value: status.Get(),
										Options: []ui.SelectOption{
											{Value: "active", Label: "Active"},
											{Value: "suspended", Label: "Suspended"},
											{Value: "pending", Label: "Pending"},
										},
										OnChange: func(v string) { status.Set(v) },
										Class:    "bg-gray-700 border-gray-600 text-white",
									})),

									// Password field (optional)
									core.Div(core.Attrs{},
										core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
											core.Text("Password (optional)"),
										),
										ui.Input(ui.InputProps{
											Type:        ui.InputPassword,
											Name:        "password",
											Value:       password.Get(),
											Placeholder: "Leave blank to keep current",
											OnChange:    func(v string) { password.Set(v) },
											Class:       "bg-gray-700 border-gray-600 text-white",
										}),
										core.Div(core.Class("text-gray-500 text-sm mt-1"),
											core.Text("Only fill if you want to change the password"),
										),
									),

									// Action buttons
									core.Div(core.Class("flex gap-2 pt-4"),
										core.Button(
											core.Attrs{
												Type:  "button",
												Class: "px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition",
												OnClick: func() {
													// Validate
													if strings.TrimSpace(name.Get()) == "" {
														errorMsg.Set("Name is required")
														successMsg.Set("")
														return
													}
													if strings.TrimSpace(email.Get()) == "" || !strings.Contains(email.Get(), "@") {
														errorMsg.Set("Valid email is required")
														successMsg.Set("")
														return
													}

													// Build update payload
													data := map[string]interface{}{
														"name":   name.Get(),
														"email":  email.Get(),
														"role":   role.Get(),
														"status": status.Get(),
													}
													if password.Get() != "" {
														if len(password.Get()) < 8 {
															errorMsg.Set("Password must be at least 8 characters")
															successMsg.Set("")
															return
														}
														data["password"] = password.Get()
													}

													api.Users.Update(displayUser.ID, data, func(result *dto.UserDetail, err error) {
														if err != nil {
															errorMsg.Set("Failed to update user")
															successMsg.Set("")
														} else {
															errorMsg.Set("")
															successMsg.Set("User updated successfully")
															password.Set("") // Clear password field
														}
													})
												},
											},
											core.Text("Save Changes"),
										),
										core.A(
											core.Attrs{
												Href:  fmt.Sprintf("/users/%d", displayUser.ID),
												Class: "px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white font-medium rounded-lg transition",
											},
											core.Text("Cancel"),
										),
									),
								},
							}),
						},
					}),
				},
			}),
		)
	}
}

// userFormField wraps a form field with a label.
func userFormField(label string, input core.Node) core.Node {
	return core.Div(core.Attrs{},
		core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
			core.Text(label),
		),
		input,
	)
}
