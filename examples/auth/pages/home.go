package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Home is the landing page for the auth example
func Home(r *core.Router) func() core.Node {
	return func() core.Node {
		return PageLayout(
			// Hero section
			core.Div(core.Class("text-center py-16"),
				core.H1(core.Class("text-4xl font-bold text-gray-900 dark:text-white mb-4"),
					core.Text("Auth Example")),
				core.P(core.Class("text-xl text-gray-600 dark:text-gray-400 mb-8 max-w-2xl mx-auto"),
					core.Text("A complete authentication example built with Gux framework. Demonstrates login, registration, password reset, and email verification patterns.")),
				core.Div(core.Class("flex gap-4 justify-center"),
					ui.Button(ui.ButtonProps{
						Variant: ui.ButtonPrimary,
						Size:    ui.ButtonLG,
						OnClick: func() { r.Navigate("/register") },
						Children: []core.Node{
							core.Text("Get Started"),
						},
					}),
					ui.Button(ui.ButtonProps{
						Variant: ui.ButtonOutline,
						Size:    ui.ButtonLG,
						OnClick: func() { r.Navigate("/login") },
						Children: []core.Node{
							core.Text("Sign In"),
						},
					}),
				),
			),

			// Features section
			core.Div(core.Class("py-16"),
				core.H2(core.Class("text-2xl font-bold text-gray-900 dark:text-white text-center mb-8"),
					core.Text("Authentication Features")),
				ui.Grid(ui.GridProps{
					Cols:  "3",
					Gap:   "6",
					Class: "max-w-4xl mx-auto",
					Children: []core.Node{
						featureCard("Login", "Secure sign-in with email and password. Includes remember me and loading states."),
						featureCard("Register", "User registration with validation. Password confirmation and field-level errors."),
						featureCard("Password Reset", "Forgot password flow with email request and token-based reset."),
						featureCard("Email Verification", "Verify user email addresses with secure tokens."),
						featureCard("Form Validation", "Client-side validation with inline error messages."),
						featureCard("Session Management", "Demonstrates auth state and protected routes."),
					},
				}),
			),

			// Demo credentials
			ui.Alert(ui.AlertProps{
				Variant: ui.AlertInfo,
				Title:   "Demo Credentials",
				Message: "Email: demo@example.com / Password: demo123",
				Class:   "max-w-md mx-auto",
			}),
		)
	}
}

func featureCard(title, description string) core.Node {
	return ui.Card(ui.CardProps{
		Children: []core.Node{
			ui.CardContent(ui.CardContentProps{
				Children: []core.Node{
					core.H3(core.Class("font-semibold text-gray-900 dark:text-white mb-2"),
						core.Text(title)),
					core.P(core.Class("text-sm text-gray-600 dark:text-gray-400"),
						core.Text(description)),
				},
			}),
		},
	})
}
