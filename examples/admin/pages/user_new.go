package pages

import (
	"strings"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/admin/.gux/api"
	"github.com/dougbarrett/gux/examples/admin/dto"
	"github.com/dougbarrett/gux/ui"
)

// UserNew creates a new user.
func UserNew(r *core.Router) func() core.Node {
	return func() core.Node {
		// Form field state
		name := r.StateString("name", "")
		email := r.StateString("email", "")
		role := r.StateString("role", "user")
		status := r.StateString("status", "active")
		password := r.StateString("password", "")

		// Message state
		errorMsg := r.StateString("error", "")

		return AdminLayout(
			// Back link
			core.A(
				core.Attrs{Href: "/users", Class: "text-blue-400 hover:text-blue-300 mb-4 inline-block"},
				core.Text("\u2190 Back to Users"),
			),

			core.H1(core.Class("text-2xl font-bold text-white mb-6"),
				core.Text("Create New User"),
			),

			// Create form card
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

									// Password field (required for new user)
									userFormField("Password", ui.Input(ui.InputProps{
										Type:        ui.InputPassword,
										Name:        "password",
										Value:       password.Get(),
										Placeholder: "Minimum 8 characters",
										OnChange:    func(v string) { password.Set(v) },
										Class:       "bg-gray-700 border-gray-600 text-white",
									})),

									// Action buttons
									core.Div(core.Class("flex gap-2 pt-4"),
										core.Button(
											core.Attrs{
												Type:  "button",
												Class: "px-4 py-2 bg-green-600 hover:bg-green-700 text-white font-medium rounded-lg transition",
												OnClick: func() {
													// Validate
													if strings.TrimSpace(name.Get()) == "" {
														errorMsg.Set("Name is required")
														return
													}
													if strings.TrimSpace(email.Get()) == "" || !strings.Contains(email.Get(), "@") {
														errorMsg.Set("Valid email is required")
														return
													}
													if password.Get() == "" {
														errorMsg.Set("Password is required")
														return
													}
													if len(password.Get()) < 8 {
														errorMsg.Set("Password must be at least 8 characters")
														return
													}

													// Build create payload
													data := map[string]interface{}{
														"name":     name.Get(),
														"email":    email.Get(),
														"role":     role.Get(),
														"status":   status.Get(),
														"password": password.Get(),
													}

													api.Users.Create(data, func(result *dto.UserDetail, err error) {
														if err != nil {
															errorMsg.Set("Failed to create user")
														} else {
															// Navigate to users list on success
															r.Navigate("/users")
														}
													})
												},
											},
											core.Text("Create User"),
										),
										core.A(
											core.Attrs{
												Href:  "/users",
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
