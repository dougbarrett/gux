package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Reset handles password reset with token
func Reset(r *core.Router) func() core.Node {
	// Get token from route params during load
	var token string
	var tokenValid bool

	r.OnLoad(func() {
		params := r.GetRouteParams()
		token = params["token"]
		// In real app: validate token via API
		// For demo: accept any non-empty token
		tokenValid = token != ""
	})

	return func() core.Node {
		// State
		passwordState := r.StateString("password", "")
		confirmState := r.StateString("confirm", "")
		errorState := r.StateString("error", "")
		successState := r.StateBool("success", false)
		loadingState := r.StateBool("loading", false)

		// Invalid/expired token
		if !tokenValid {
			return AuthLayout(
				ui.Card(ui.CardProps{
					Children: []core.Node{
						ui.CardContent(ui.CardContentProps{
							Class: "text-center py-8",
							Children: []core.Node{
								core.Div(core.Class("text-red-500 text-5xl mb-4"),
									core.Text("\u2717")), // X mark
								core.H1(core.Class("text-2xl font-bold text-gray-900 dark:text-white mb-2"),
									core.Text("Invalid Reset Link")),
								core.P(core.Class("text-gray-600 dark:text-gray-400 mb-6"),
									core.Text("This password reset link is invalid or has expired. Please request a new one.")),
								ui.Button(ui.ButtonProps{
									Variant: ui.ButtonPrimary,
									OnClick: func() { r.Navigate("/forgot") },
									Children: []core.Node{
										core.Text("Request New Link"),
									},
								}),
							},
						}),
					},
				}),
			)
		}

		// Success state
		if successState.Get() {
			return AuthLayout(
				ui.Card(ui.CardProps{
					Children: []core.Node{
						ui.CardContent(ui.CardContentProps{
							Class: "text-center py-8",
							Children: []core.Node{
								core.Div(core.Class("text-green-500 text-5xl mb-4"),
									core.Text("\u2713")), // Checkmark
								core.H1(core.Class("text-2xl font-bold text-gray-900 dark:text-white mb-2"),
									core.Text("Password Reset Complete")),
								core.P(core.Class("text-gray-600 dark:text-gray-400 mb-6"),
									core.Text("Your password has been successfully changed. You can now log in with your new password.")),
								ui.Button(ui.ButtonProps{
									Variant: ui.ButtonPrimary,
									OnClick: func() { r.Navigate("/login") },
									Children: []core.Node{
										core.Text("Continue to Login"),
									},
								}),
							},
						}),
					},
				}),
			)
		}

		// Submit handler
		submit := func() {
			errorState.Set("")

			// Validate password
			if passwordState.Get() == "" {
				errorState.Set("Password is required")
				return
			}
			if len(passwordState.Get()) < 8 {
				errorState.Set("Password must be at least 8 characters")
				return
			}

			// Validate confirm
			if confirmState.Get() != passwordState.Get() {
				errorState.Set("Passwords do not match")
				return
			}

			loadingState.Set(true)

			// In real app: call API to reset password with token
			// For demo: just show success
			loadingState.Set(false)
			successState.Set(true)
		}

		// Error alert
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
								core.Text("Reset Password")),
							core.P(core.Class("text-center text-gray-600 dark:text-gray-400 mt-2"),
								core.Text("Enter your new password below.")),
						},
					}),
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							alertNode,

							// Password field
							ui.FormField(ui.FormFieldProps{
								Label:       "New Password",
								LabelFor:    "password",
								Required:    true,
								Description: "Must be at least 8 characters",
								Children: []core.Node{
									ui.Input(ui.InputProps{
										ID:          "password",
										Type:        ui.InputPassword,
										Name:        "password",
										Value:       passwordState.Get(),
										Placeholder: "Enter new password",
										OnChange:    func(v string) { passwordState.Set(v) },
										OnEnter:     submit,
									}),
								},
							}),

							// Confirm password field
							ui.FormField(ui.FormFieldProps{
								Label:    "Confirm Password",
								LabelFor: "confirm",
								Required: true,
								Children: []core.Node{
									ui.Input(ui.InputProps{
										ID:          "confirm",
										Type:        ui.InputPassword,
										Name:        "confirm",
										Value:       confirmState.Get(),
										Placeholder: "Confirm new password",
										OnChange:    func(v string) { confirmState.Set(v) },
										OnEnter:     submit,
									}),
								},
							}),

							// Submit button
							ui.Button(ui.ButtonProps{
								Variant:  ui.ButtonPrimary,
								Class:    "w-full",
								Disabled: loadingState.Get(),
								OnClick:  submit,
								Children: []core.Node{
									core.Text(func() string {
										if loadingState.Get() {
											return "Resetting..."
										}
										return "Reset Password"
									}()),
								},
							}),
						},
					}),
				},
			}),
		)
	}
}
