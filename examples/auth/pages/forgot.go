package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Forgot handles password reset requests
func Forgot(r *core.Router) func() core.Node {
	return func() core.Node {
		// Form state
		emailState := r.StateString("email", "")
		errorState := r.StateString("error", "")
		successState := r.StateBool("success", false)
		loadingState := r.StateBool("loading", false)

		// Submit handler
		submit := func() {
			errorState.Set("")

			if emailState.Get() == "" {
				errorState.Set("Email is required")
				return
			}

			loadingState.Set(true)

			// Simulate API call - in real app would:
			// 1. Look up user by email
			// 2. Generate reset token via PasswordReset.GenerateToken()
			// 3. Send email with reset link
			// For demo, just show success after delay
			// Using immediate success for simplicity
			loadingState.Set(false)
			successState.Set(true)
		}

		// Success state
		if successState.Get() {
			return AuthLayout(
				ui.Card(ui.CardProps{
					Children: []core.Node{
						ui.CardContent(ui.CardContentProps{
							Class: "text-center py-8",
							Children: []core.Node{
								core.Div(core.Class("text-blue-500 text-5xl mb-4"),
									core.Text("\u2709")), // Envelope
								core.H1(core.Class("text-2xl font-bold text-gray-900 dark:text-white mb-2"),
									core.Text("Check Your Email")),
								core.P(core.Class("text-gray-600 dark:text-gray-400 mb-2"),
									core.Text("If an account exists for "),
									core.Span(core.Class("font-medium"), core.Text(emailState.Get())),
									core.Text(", you will receive a password reset link.")),
								core.P(core.Class("text-sm text-gray-500 dark:text-gray-500 mb-6"),
									core.Text("The link will expire in 1 hour.")),

								// Demo link (in real app, this would be in the email)
								ui.Alert(ui.AlertProps{
									Variant: ui.AlertInfo,
									Title:   "Demo Mode",
									Message: "Click below to simulate clicking the email link.",
									Class:   "mb-4 text-left",
								}),
								ui.Button(ui.ButtonProps{
									Variant: ui.ButtonOutline,
									OnClick: func() { r.Navigate("/reset/demo-token-12345") },
									Children: []core.Node{
										core.Text("Open Reset Link"),
									},
								}),

								core.Div(core.Class("mt-6"),
									core.A(
										core.Attrs{
											Href:  "/login",
											Class: "text-blue-600 dark:text-blue-400 hover:underline",
										},
										core.Text("Back to Login"),
									),
								),
							},
						}),
					},
				}),
			)
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
								core.Text("Forgot Password")),
							core.P(core.Class("text-center text-gray-600 dark:text-gray-400 mt-2"),
								core.Text("Enter your email and we'll send you a reset link.")),
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

							// Submit button
							ui.Button(ui.ButtonProps{
								Variant:  ui.ButtonPrimary,
								Class:    "w-full",
								Disabled: loadingState.Get(),
								OnClick:  submit,
								Children: []core.Node{
									core.Text(func() string {
										if loadingState.Get() {
											return "Sending..."
										}
										return "Send Reset Link"
									}()),
								},
							}),
						},
					}),
					ui.CardFooter(ui.CardFooterProps{
						Class: "text-center",
						Children: []core.Node{
							core.A(
								core.Attrs{
									Href:  "/login",
									Class: "text-blue-600 dark:text-blue-400 hover:underline",
								},
								core.Text("Back to Login"),
							),
						},
					}),
				},
			}),
		)
	}
}
