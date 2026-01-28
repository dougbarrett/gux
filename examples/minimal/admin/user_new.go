package admin

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/guxgen/api"
	"github.com/dougbarrett/gux/examples/minimal/dto"
)

// UserNew is the form for creating a new user.
func UserNew(r *core.Router) func() core.Node {
	return func() core.Node {
		// Form state
		nameState := r.StateString("name", "")
		emailState := r.StateString("email", "")
		passwordState := r.StateString("password", "")
		roleState := r.StateString("role", "user")
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
			if nameState.Get() == "" || emailState.Get() == "" || passwordState.Get() == "" {
				errorState.Set("All fields are required")
				return
			}

			// Create user via generated API
			data := map[string]interface{}{
				"name":     nameState.Get(),
				"email":    emailState.Get(),
				"password": passwordState.Get(),
				"role":     roleState.Get(),
			}
			api.Users.Create(data, func(result *dto.UserDetail, err error) {
				if err != nil {
					errorState.Set(err.Error())
				} else {
					successState.Set("User created successfully!")
					// Clear form
					nameState.Set("")
					emailState.Set("")
					passwordState.Set("")
					roleState.Set("user")
					errorState.Set("")
					// Navigate back to users list
					r.Navigate("/admin/users")
				}
			})
		}

		return core.Div(core.Class("min-h-screen bg-gray-900"),
			Nav(),
			core.Div(core.Class("max-w-2xl mx-auto px-4 py-8"),
				// Back link
				core.A(
					core.Attrs{Href: "/admin/users", Class: "text-blue-400 hover:text-blue-300 mb-4 inline-block"},
					core.Text("← Back to Users"),
				),

				core.H1(core.Class("text-2xl font-bold text-white mb-6"),
					core.Text("Create New User"),
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

					// Password field
					formField("Password", "password", "password", passwordState.Get(), func(v string) {
						passwordState.Set(v)
					}, save),

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
							core.Option(core.Attrs{Value: "user"}, core.Text("User")),
							core.Option(core.Attrs{Value: "admin"}, core.Text("Admin")),
						),
					),

					// Submit button
					core.Button(
						core.Attrs{
							Type:    "button",
							Class:   "w-full px-4 py-2 bg-green-600 hover:bg-green-700 text-white font-medium rounded-lg transition",
							OnClick: save,
						},
						core.Text("Create User"),
					),
				),
			),
		)
	}
}

func formField(label, inputType, name, value string, onChange func(string), onEnter ...func()) core.Node {
	attrs := core.Attrs{
		Type:     inputType,
		Name:     name,
		Value:    value,
		Class:    "w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:border-blue-500",
		OnChange: onChange,
	}
	if len(onEnter) > 0 {
		attrs.OnEnter = onEnter[0]
	}
	return core.Div(core.Class("mb-4"),
		core.Label(core.Class("block text-gray-300 text-sm font-medium mb-2"),
			core.Text(label),
		),
		core.Input(attrs),
	)
}
