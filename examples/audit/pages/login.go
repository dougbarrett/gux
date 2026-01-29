package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/audit/guxgen/api"
	"github.com/dougbarrett/gux/ui"
)

// Login is the login page.
func Login(r *core.Router) func() core.Node {
	email := r.StateString("email", "")
	password := r.StateString("password", "")
	errorMsg := r.StateString("error", "")
	loading := r.StateBool("loading", false)

	handleLogin := func() {
		if loading.Get() {
			return
		}
		if email.Get() == "" || password.Get() == "" {
			errorMsg.Set("Please enter email and password")
			return
		}
		loading.Set(true)
		errorMsg.Set("")

		data := map[string]interface{}{
			"email":    email.Get(),
			"password": password.Get(),
		}
		api.Login(data, func(resp api.LoginResponse, err error) {
			loading.Set(false)
			if err != nil {
				errorMsg.Set("Login failed. Please try again.")
				return
			}
			if !resp.Success {
				errorMsg.Set(resp.Error)
				return
			}
			r.Navigate(resp.Redirect)
		})
	}

	return func() core.Node {
		return core.Div(core.Class("min-h-screen bg-gray-900 flex items-center justify-center px-4"),
			core.Div(core.Class("max-w-md w-full"),
				core.Div(core.Class("bg-gray-800 shadow-lg rounded-lg px-8 py-10 border border-gray-700"),
					core.H2(core.Class("text-3xl font-bold text-white mb-2 text-center"),
						core.Text("Sign In"),
					),
					core.P(core.Class("text-gray-400 mb-8 text-center"),
						core.Text("Document Manager"),
					),
					core.If(errorMsg.Get() != "",
						core.Div(core.Class("bg-red-900/50 border border-red-700 text-red-200 px-4 py-3 rounded mb-6"),
							core.Text(errorMsg.Get()),
						),
					),
					core.Div(core.Class("space-y-6"),
						core.Div(core.Attrs{},
							core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
								core.Text("Email"),
							),
							ui.Input(ui.InputProps{
								ID:          "email",
								Type:        "email",
								Placeholder: "admin@example.com",
								Value:       email.Get(),
								OnChange:    func(val string) { email.Set(val) },
								Disabled:    loading.Get(),
							}),
						),
						core.Div(core.Attrs{},
							core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
								core.Text("Password"),
							),
							ui.Input(ui.InputProps{
								ID:          "password",
								Type:        "password",
								Placeholder: "Enter your password",
								Value:       password.Get(),
								OnChange:    func(val string) { password.Set(val) },
								OnEnter:     handleLogin,
								Disabled:    loading.Get(),
							}),
						),
						ui.Button(ui.ButtonProps{
							Variant:  ui.ButtonPrimary,
							Class:    "w-full",
							Disabled: loading.Get(),
							Children: []core.Node{
								func() core.Node {
									if loading.Get() {
										return core.Text("Signing in...")
									}
									return core.Text("Sign In")
								}(),
							},
							OnClick: handleLogin,
						}),
					),
					core.Div(core.Class("mt-6 pt-6 border-t border-gray-700"),
						core.P(core.Class("text-sm text-gray-400 text-center"),
							core.Text("Demo credentials: admin@example.com / admin123"),
						),
					),
				),
			),
		)
	}
}
