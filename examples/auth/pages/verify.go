package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Verify handles email verification with token
func Verify(r *core.Router) func() core.Node {
	// Verification state set during load
	var token string
	var verified bool
	var verifyError string

	r.OnLoad(func() {
		params := r.GetRouteParams()
		token = params["token"]

		if token == "" {
			verifyError = "No verification token provided"
			return
		}

		// In real app: validate token via API
		// - Look up token in database
		// - Check if expired
		// - Mark user as verified
		// For demo: accept any token except "expired" or "invalid"
		if token == "expired" {
			verifyError = "This verification link has expired. Please request a new one."
		} else if token == "invalid" {
			verifyError = "This verification link is invalid."
		} else {
			verified = true
		}
	})

	return func() core.Node {
		// Error state
		if verifyError != "" {
			return AuthLayout(
				ui.Card(ui.CardProps{
					Children: []core.Node{
						ui.CardContent(ui.CardContentProps{
							Class: "text-center py-8",
							Children: []core.Node{
								core.Div(core.Class("text-red-500 text-5xl mb-4"),
									core.Text("\u2717")), // X mark
								core.H1(core.Class("text-2xl font-bold text-gray-900 dark:text-white mb-2"),
									core.Text("Verification Failed")),
								core.P(core.Class("text-gray-600 dark:text-gray-400 mb-6"),
									core.Text(verifyError)),
								ui.Button(ui.ButtonProps{
									Variant: ui.ButtonPrimary,
									OnClick: func() { r.Navigate("/login") },
									Children: []core.Node{
										core.Text("Back to Login"),
									},
								}),
							},
						}),
					},
				}),
			)
		}

		// Success state
		if verified {
			return AuthLayout(
				ui.Card(ui.CardProps{
					Children: []core.Node{
						ui.CardContent(ui.CardContentProps{
							Class: "text-center py-8",
							Children: []core.Node{
								core.Div(core.Class("text-green-500 text-5xl mb-4"),
									core.Text("\u2713")), // Checkmark
								core.H1(core.Class("text-2xl font-bold text-gray-900 dark:text-white mb-2"),
									core.Text("Email Verified!")),
								core.P(core.Class("text-gray-600 dark:text-gray-400 mb-6"),
									core.Text("Your email has been successfully verified. You can now access all features of your account.")),
								ui.Alert(ui.AlertProps{
									Variant: ui.AlertSuccess,
									Message: "Your account is now fully activated.",
									Class:   "mb-6",
								}),
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

		// Fallback (shouldn't reach here normally)
		return AuthLayout(
			ui.Card(ui.CardProps{
				Children: []core.Node{
					ui.CardContent(ui.CardContentProps{
						Class: "text-center py-8",
						Children: []core.Node{
							core.P(core.Class("text-gray-600 dark:text-gray-400"),
								core.Text("Verifying your email...")),
						},
					}),
				},
			}),
		)
	}
}
