package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/auth/.gux/api"
	"github.com/dougbarrett/gux/examples/auth/dto"
	"github.com/dougbarrett/gux/ui"
)

// Login handles user authentication
func Login(r *core.Router) func() core.Node {
	return func() core.Node {
		// Form state
		emailState := r.StateString("email", "")
		passwordState := r.StateString("password", "")
		rememberState := r.StateBool("remember", false)
		errorState := r.StateString("error", "")
		loadingState := r.StateBool("loading", false)

		// Submit handler
		submit := func() {
			// Clear previous error
			errorState.Set("")

			// Validate
			if emailState.Get() == "" {
				errorState.Set("Email is required")
				return
			}
			if passwordState.Get() == "" {
				errorState.Set("Password is required")
				return
			}

			loadingState.Set(true)

			// Call login API (simplified - in real app would be auth endpoint)
			// For demo: check if user exists and password matches
			api.Users.List(func(users []dto.UserDTO, err error) {
				loadingState.Set(false)
				if err != nil {
					errorState.Set("Login failed: " + err.Error())
					return
				}

				// Find user by email (demo only - real auth would be server-side)
				var found bool
				for _, u := range users {
					if u.Email == emailState.Get() {
						found = true
						// In demo, assume password check passed
						// Real app: server validates password, returns JWT
						r.Navigate("/dashboard")
						return
					}
				}
				if !found {
					errorState.Set("Invalid email or password")
				}
			})
		}

		// Toggle remember me
		toggleRemember := func() {
			rememberState.Set(!rememberState.Get())
		}

		// Error alert (only shown when error exists)
		var alertNode core.Node = core.Frag()
		if errorState.Get() != "" {
			alertNode = ui.Alert(ui.AlertProps{
				Variant: ui.AlertError,
				Message: errorState.Get(),
				Class:   "mb-4",
			})
		}

		return AuthLayout(
			ui.Card(ui.CardProps{
				Children: []core.Node{
					ui.CardHeader(ui.CardHeaderProps{
						Children: []core.Node{
							core.H1(core.Class("text-2xl font-bold text-center text-gray-900 dark:text-white"),
								core.Text("Sign In")),
							core.P(core.Class("text-center text-gray-600 dark:text-gray-400 mt-2"),
								core.Text("Welcome back! Please enter your credentials.")),
						},
					}),
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							alertNode,

							// Email field
							ui.FormField(ui.FormFieldProps{
								Label:    "Email",
								LabelFor: "email",
								Required: true,
								Children: []core.Node{
									ui.Input(ui.InputProps{
										ID:          "email",
										Type:        ui.InputEmail,
										Name:        "email",
										Value:       emailState.Get(),
										Placeholder: "you@example.com",
										OnChange:    func(v string) { emailState.Set(v) },
										OnEnter:     submit,
									}),
								},
							}),

							// Password field
							ui.FormField(ui.FormFieldProps{
								Label:    "Password",
								LabelFor: "password",
								Required: true,
								Children: []core.Node{
									ui.Input(ui.InputProps{
										ID:          "password",
										Type:        ui.InputPassword,
										Name:        "password",
										Value:       passwordState.Get(),
										Placeholder: "Enter your password",
										OnChange:    func(v string) { passwordState.Set(v) },
										OnEnter:     submit,
									}),
								},
							}),

							// Remember me + Forgot password row
							core.Div(core.Class("flex items-center justify-between mb-6"),
								// Remember me checkbox with click handler on wrapper
								core.Div(
									core.Attrs{
										Class:   "flex items-center gap-2 cursor-pointer",
										OnClick: toggleRemember,
									},
									ui.Checkbox(ui.CheckboxProps{
										ID:      "remember",
										Name:    "remember",
										Checked: rememberState.Get(),
									}),
									core.Span(core.Class("text-sm font-medium text-gray-700 dark:text-gray-300"),
										core.Text("Remember me")),
								),
								core.A(
									core.Attrs{
										Href:  "/forgot",
										Class: "text-sm text-blue-600 dark:text-blue-400 hover:underline",
									},
									core.Text("Forgot password?"),
								),
							),

							// Submit button
							ui.Button(ui.ButtonProps{
								Variant:  ui.ButtonPrimary,
								Class:    "w-full",
								Disabled: loadingState.Get(),
								OnClick:  submit,
								Children: []core.Node{
									core.Text(func() string {
										if loadingState.Get() {
											return "Signing in..."
										}
										return "Sign In"
									}()),
								},
							}),
						},
					}),
					ui.CardFooter(ui.CardFooterProps{
						Class: "text-center",
						Children: []core.Node{
							core.P(core.Class("text-gray-600 dark:text-gray-400"),
								core.Text("Don't have an account? "),
								core.A(
									core.Attrs{
										Href:  "/register",
										Class: "text-blue-600 dark:text-blue-400 hover:underline font-medium",
									},
									core.Text("Sign up"),
								),
							),
						},
					}),
				},
			}),
		)
	}
}
