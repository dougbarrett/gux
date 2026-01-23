package pages

import (
	"strings"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Contact renders the contact page with form and company info.
func Contact(r *core.Router) func() core.Node {
	return func() core.Node {
		// Form state
		nameState := r.StateString("name", "")
		emailState := r.StateString("email", "")
		subjectState := r.StateString("subject", "")
		messageState := r.StateString("message", "")
		errorState := r.StateString("error", "")
		successState := r.StateBool("success", false)

		// Submit handler
		submit := func() {
			errorState.Set("")

			// Validate required fields
			if strings.TrimSpace(nameState.Get()) == "" {
				errorState.Set("Name is required")
				return
			}
			email := strings.TrimSpace(emailState.Get())
			if email == "" || !strings.Contains(email, "@") {
				errorState.Set("Valid email is required")
				return
			}
			if strings.TrimSpace(messageState.Get()) == "" {
				errorState.Set("Message is required")
				return
			}

			// Demo mode - just show success
			successState.Set(true)
		}

		// Reset form handler
		resetForm := func() {
			nameState.Set("")
			emailState.Set("")
			subjectState.Set("")
			messageState.Set("")
			errorState.Set("")
			successState.Set(false)
		}

		// Success state - show success message
		if successState.Get() {
			return MarketingLayout(r,
				Section("", "",
					core.Div(core.Class("max-w-lg mx-auto"),
						ui.Card(ui.CardProps{
							Children: []core.Node{
								ui.CardContent(ui.CardContentProps{
									Class: "text-center py-8",
									Children: []core.Node{
										core.Div(core.Class("text-green-500 text-5xl mb-4"),
											core.Text("\u2713")), // Checkmark
										core.H2(core.Class("text-2xl font-bold text-gray-900 dark:text-white mb-2"),
											core.Text("Message Sent!")),
										core.P(core.Class("text-gray-600 dark:text-gray-400 mb-6"),
											core.Text("Thanks for your message! We'll get back to you within 24 hours.")),
										ui.Button(ui.ButtonProps{
											Variant: ui.ButtonPrimary,
											OnClick: resetForm,
											Children: []core.Node{
												core.Text("Send Another Message"),
											},
										}),
									},
								}),
							},
						}),
					),
				),
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

		return MarketingLayout(r,
			Section("", "",
				ui.Grid(ui.GridProps{
					Cols:  "2",
					Gap:   "12",
					Class: "md:grid-cols-2 grid-cols-1",
					Children: []core.Node{
						// Contact Info Column
						core.Div(core.Attrs{},
							core.H2(core.Class("text-3xl font-bold text-gray-900 dark:text-white mb-4"),
								core.Text("Get in Touch")),
							core.P(core.Class("text-lg text-gray-600 dark:text-gray-400 mb-8"),
								core.Text("Have a question or want to learn more? We'd love to hear from you.")),

							// Contact details
							ui.VStack(ui.StackProps{
								Gap: "4",
								Children: []core.Node{
									core.Div(core.Class("flex items-center gap-3"),
										core.Span(core.Class("text-blue-600 text-xl"), core.Text("\u2709")), // Envelope
										core.Span(core.Class("text-gray-700 dark:text-gray-300"),
											core.Text("hello@marketco.com")),
									),
									core.Div(core.Class("flex items-center gap-3"),
										core.Span(core.Class("text-blue-600 text-xl"), core.Text("\u260e")), // Telephone
										core.Span(core.Class("text-gray-700 dark:text-gray-300"),
											core.Text("+1 (555) 123-4567")),
									),
									core.Div(core.Class("flex items-center gap-3"),
										core.Span(core.Class("text-blue-600 text-xl"), core.Text("\u2302")), // House
										core.Span(core.Class("text-gray-700 dark:text-gray-300"),
											core.Text("123 Main St, San Francisco, CA 94102")),
									),
								},
							}),

							// Social links
							core.Div(core.Class("mt-8"),
								core.H3(core.Class("text-sm font-semibold text-gray-900 dark:text-white uppercase tracking-wide mb-4"),
									core.Text("Follow Us")),
								ui.HStack(ui.StackProps{
									Gap: "4",
									Children: []core.Node{
										core.A(core.Attrs{
											Href:  "#",
											Class: "text-gray-400 hover:text-blue-500 transition-colors",
										}, core.Text("Twitter")),
										core.A(core.Attrs{
											Href:  "#",
											Class: "text-gray-400 hover:text-blue-700 transition-colors",
										}, core.Text("LinkedIn")),
										core.A(core.Attrs{
											Href:  "#",
											Class: "text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors",
										}, core.Text("GitHub")),
									},
								}),
							),
						),

						// Contact Form Column
						ui.Card(ui.CardProps{
							Children: []core.Node{
								ui.CardHeader(ui.CardHeaderProps{
									Children: []core.Node{
										core.H2(core.Class("text-xl font-bold text-gray-900 dark:text-white"),
											core.Text("Send us a Message")),
									},
								}),
								ui.CardContent(ui.CardContentProps{
									Children: []core.Node{
										alertNode,

										// Name field
										ui.FormField(ui.FormFieldProps{
											Label:    "Name",
											LabelFor: "name",
											Required: true,
											Children: []core.Node{
												ui.Input(ui.InputProps{
													ID:          "name",
													Type:        ui.InputText,
													Name:        "name",
													Value:       nameState.Get(),
													Placeholder: "Your name",
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

										// Subject field (optional)
										ui.FormField(ui.FormFieldProps{
											Label:    "Subject",
											LabelFor: "subject",
											Children: []core.Node{
												ui.Input(ui.InputProps{
													ID:          "subject",
													Type:        ui.InputText,
													Name:        "subject",
													Value:       subjectState.Get(),
													Placeholder: "What's this about?",
													OnChange:    func(v string) { subjectState.Set(v) },
													OnEnter:     submit,
												}),
											},
										}),

										// Message field
										ui.FormField(ui.FormFieldProps{
											Label:    "Message",
											LabelFor: "message",
											Required: true,
											Children: []core.Node{
												ui.Textarea(ui.TextareaProps{
													ID:          "message",
													Name:        "message",
													Value:       messageState.Get(),
													Placeholder: "How can we help you?",
													Rows:        5,
													OnChange:    func(v string) { messageState.Set(v) },
												}),
											},
										}),

										// Submit button
										ui.Button(ui.ButtonProps{
											Variant: ui.ButtonPrimary,
											Class:   "w-full",
											OnClick: submit,
											Children: []core.Node{
												core.Text("Send Message"),
											},
										}),
									},
								}),
							},
						}),
					},
				}),
			),
		)
	}
}
