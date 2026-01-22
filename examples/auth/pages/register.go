package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/auth/.gux/api"
	"github.com/dougbarrett/gux/examples/auth/dto"
	"github.com/dougbarrett/gux/ui"
)

// Register handles new user registration
func Register(r *core.Router) func() core.Node {
	return func() core.Node {
		// Form state
		nameState := r.StateString("name", "")
		emailState := r.StateString("email", "")
		passwordState := r.StateString("password", "")
		confirmState := r.StateString("confirm", "")
		errorState := r.StateString("error", "")
		successState := r.StateBool("success", false)
		loadingState := r.StateBool("loading", false)

		// Field-level errors
		nameError := r.StateString("nameError", "")
		emailError := r.StateString("emailError", "")
		passwordError := r.StateString("passwordError", "")
		confirmError := r.StateString("confirmError", "")

		// Clear all field errors
		clearErrors := func() {
			nameError.Set("")
			emailError.Set("")
			passwordError.Set("")
			confirmError.Set("")
			errorState.Set("")
		}

		// Submit handler
		submit := func() {
			clearErrors()
			var hasError bool

			// Validate name
			if nameState.Get() == "" {
				nameError.Set("Name is required")
				hasError = true
			}

			// Validate email
			if emailState.Get() == "" {
				emailError.Set("Email is required")
				hasError = true
			}

			// Validate password
			if passwordState.Get() == "" {
				passwordError.Set("Password is required")
				hasError = true
			} else if len(passwordState.Get()) < 8 {
				passwordError.Set("Password must be at least 8 characters")
				hasError = true
			}

			// Validate confirm
			if confirmState.Get() == "" {
				confirmError.Set("Please confirm your password")
				hasError = true
			} else if confirmState.Get() != passwordState.Get() {
				confirmError.Set("Passwords do not match")
				hasError = true
			}

			if hasError {
				return
			}

			loadingState.Set(true)

			// Create user via API
			data := map[string]interface{}{
				"name":     nameState.Get(),
				"email":    emailState.Get(),
				"password": passwordState.Get(),
			}
			api.Users.Create(data, func(result *dto.UserDTO, err error) {
				loadingState.Set(false)
				if err != nil {
					errorState.Set("Registration failed: " + err.Error())
					return
				}
				successState.Set(true)
			})
		}

		// Success state - show success message
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
									core.Text("Registration Successful!")),
								core.P(core.Class("text-gray-600 dark:text-gray-400 mb-6"),
									core.Text("Your account has been created. Please check your email to verify your account.")),
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
								core.Text("Create Account")),
							core.P(core.Class("text-center text-gray-600 dark:text-gray-400 mt-2"),
								core.Text("Join us today! Fill in your details below.")),
						},
					}),
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							alertNode,

							// Name field
							ui.FormField(ui.FormFieldProps{
								Label:    "Full Name",
								LabelFor: "name",
								Required: true,
								Error:    nameError.Get(),
								Children: []core.Node{
									ui.Input(ui.InputProps{
										ID:          "name",
										Type:        ui.InputText,
										Name:        "name",
										Value:       nameState.Get(),
										Placeholder: "John Doe",
										Error:       nameError.Get(),
										OnChange:    func(v string) { nameState.Set(v) },
										OnEnter:     submit,
									}),
								},
							}),

							// Email field
							ui.FormField(ui.FormFieldProps{
								Label:    "Email",
								LabelFor: "email",
								Required: true,
								Error:    emailError.Get(),
								Children: []core.Node{
									ui.Input(ui.InputProps{
										ID:          "email",
										Type:        ui.InputEmail,
										Name:        "email",
										Value:       emailState.Get(),
										Placeholder: "you@example.com",
										Error:       emailError.Get(),
										OnChange:    func(v string) { emailState.Set(v) },
										OnEnter:     submit,
									}),
								},
							}),

							// Password field
							ui.FormField(ui.FormFieldProps{
								Label:       "Password",
								LabelFor:    "password",
								Required:    true,
								Error:       passwordError.Get(),
								Description: "Must be at least 8 characters",
								Children: []core.Node{
									ui.Input(ui.InputProps{
										ID:          "password",
										Type:        ui.InputPassword,
										Name:        "password",
										Value:       passwordState.Get(),
										Placeholder: "Create a password",
										Error:       passwordError.Get(),
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
								Error:    confirmError.Get(),
								Children: []core.Node{
									ui.Input(ui.InputProps{
										ID:          "confirm",
										Type:        ui.InputPassword,
										Name:        "confirm",
										Value:       confirmState.Get(),
										Placeholder: "Confirm your password",
										Error:       confirmError.Get(),
										OnChange:    func(v string) { confirmState.Set(v) },
										OnEnter:     submit,
									}),
								},
							}),

							// Submit button
							ui.Button(ui.ButtonProps{
								Variant:  ui.ButtonPrimary,
								Class:    "w-full mt-2",
								Disabled: loadingState.Get(),
								OnClick:  submit,
								Children: []core.Node{
									core.Text(func() string {
										if loadingState.Get() {
											return "Creating account..."
										}
										return "Create Account"
									}()),
								},
							}),
						},
					}),
					ui.CardFooter(ui.CardFooterProps{
						Class: "text-center",
						Children: []core.Node{
							core.P(core.Class("text-gray-600 dark:text-gray-400"),
								core.Text("Already have an account? "),
								core.A(
									core.Attrs{
										Href:  "/login",
										Class: "text-blue-600 dark:text-blue-400 hover:underline font-medium",
									},
									core.Text("Sign in"),
								),
							),
						},
					}),
				},
			}),
		)
	}
}
