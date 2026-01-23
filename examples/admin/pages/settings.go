package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Settings is the settings page with tabbed sections for General, Security, and Email.
func Settings(r *core.Router) func() core.Node {
	// Mock current user for role-based UI
	currentUser := struct {
		ID   uint
		Name string
		Role string
	}{ID: 1, Name: "Admin User", Role: "admin"}

	return func() core.Node {
		activeTab := r.StateInt("activeTab", 0)

		return AdminLayout(
			core.H1(core.Class("text-2xl font-bold text-white mb-6"),
				core.Text("Settings"),
			),

			// Main content card
			ui.Card(ui.CardProps{
				Class: "bg-gray-800 border border-gray-700",
				Children: []core.Node{
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							ui.Tabs(ui.TabsProps{
								Children: []core.Node{
									ui.TabList(ui.TabListProps{
										Class: "border-gray-700",
										Children: []core.Node{
											ui.Tab(ui.TabProps{
												Active:   activeTab.Get() == 0,
												OnSelect: func() { activeTab.Set(0) },
												Children: []core.Node{core.Text("General")},
											}),
											ui.Tab(ui.TabProps{
												Active:   activeTab.Get() == 1,
												OnSelect: func() { activeTab.Set(1) },
												Children: []core.Node{core.Text("Security")},
											}),
											ui.Tab(ui.TabProps{
												Active:   activeTab.Get() == 2,
												OnSelect: func() { activeTab.Set(2) },
												Children: []core.Node{core.Text("Email")},
											}),
										},
									}),
									ui.TabPanel(ui.TabPanelProps{Active: activeTab.Get() == 0, Children: generalTab(r)}),
									ui.TabPanel(ui.TabPanelProps{Active: activeTab.Get() == 1, Children: securitySettingsTab(r, currentUser.Role)}),
									ui.TabPanel(ui.TabPanelProps{Active: activeTab.Get() == 2, Children: emailTab(r)}),
								},
							}),
						},
					}),
				},
			}),
		)
	}
}

// generalTab renders the General settings content.
func generalTab(r *core.Router) []core.Node {
	siteName := r.StateString("siteName", "Admin Panel")
	siteDescription := r.StateString("siteDescription", "Administrative dashboard for managing users and settings")
	timezone := r.StateString("timezone", "UTC")
	language := r.StateString("language", "en")
	successGeneral := r.StateString("successGeneral", "")

	return []core.Node{
		ui.VStack(ui.StackProps{
			Gap: "6",
			Children: []core.Node{
				// Success message
				func() core.Node {
					if successGeneral.Get() != "" {
						return core.Div(core.Class("bg-green-900/20 border border-green-800 text-green-400 rounded-lg px-4 py-3 mb-4"),
							core.Text(successGeneral.Get()),
						)
					}
					return core.Frag()
				}(),

				// Site name field
				core.Div(core.Attrs{},
					core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
						core.Text("Site Name"),
					),
					ui.Input(ui.InputProps{
						Type:     ui.InputText,
						Name:     "siteName",
						Value:    siteName.Get(),
						OnChange: func(v string) { siteName.Set(v) },
						Class:    "bg-gray-700 border-gray-600 text-white",
					}),
				),

				// Site description field
				core.Div(core.Attrs{},
					core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
						core.Text("Site Description"),
					),
					ui.Textarea(ui.TextareaProps{
						Name:     "siteDescription",
						Value:    siteDescription.Get(),
						Rows:     3,
						OnChange: func(v string) { siteDescription.Set(v) },
						Class:    "bg-gray-700 border-gray-600 text-white",
					}),
				),

				// Timezone select
				core.Div(core.Attrs{},
					core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
						core.Text("Default Timezone"),
					),
					ui.Select(ui.SelectProps{
						Name:        "timezone",
						Value:       timezone.Get(),
						Placeholder: "Select timezone",
						Options: []ui.SelectOption{
							{Value: "UTC", Label: "UTC"},
							{Value: "America/New_York", Label: "America/New_York"},
							{Value: "Europe/London", Label: "Europe/London"},
							{Value: "Asia/Tokyo", Label: "Asia/Tokyo"},
						},
						OnChange: func(v string) { timezone.Set(v) },
						Class:    "bg-gray-700 border-gray-600 text-white",
					}),
				),

				// Language select
				core.Div(core.Attrs{},
					core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
						core.Text("Default Language"),
					),
					ui.Select(ui.SelectProps{
						Name:        "language",
						Value:       language.Get(),
						Placeholder: "Select language",
						Options: []ui.SelectOption{
							{Value: "en", Label: "English"},
							{Value: "es", Label: "Spanish"},
							{Value: "fr", Label: "French"},
							{Value: "de", Label: "German"},
						},
						OnChange: func(v string) { language.Set(v) },
						Class:    "bg-gray-700 border-gray-600 text-white",
					}),
				),

				// Save button
				core.Div(core.Class("pt-4"),
					ui.Button(ui.ButtonProps{
						Variant: ui.ButtonPrimary,
						OnClick: func() {
							successGeneral.Set("General settings saved successfully")
						},
						Children: []core.Node{core.Text("Save Changes")},
					}),
				),
			},
		}),
	}
}

// securitySettingsTab renders the Security settings content with role-based Danger Zone.
func securitySettingsTab(r *core.Router, userRole string) []core.Node {
	sessionTimeout := r.StateString("sessionTimeout", "30")
	passwordMinLength := r.StateString("passwordMinLength", "8")
	twoFactorRequired := r.StateBool("twoFactorRequired", false)
	successSecurity := r.StateString("successSecurity", "")

	children := []core.Node{
		// Success message
		func() core.Node {
			if successSecurity.Get() != "" {
				return core.Div(core.Class("bg-green-900/20 border border-green-800 text-green-400 rounded-lg px-4 py-3 mb-4"),
					core.Text(successSecurity.Get()),
				)
			}
			return core.Frag()
		}(),

		// Session timeout
		core.Div(core.Attrs{},
			core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
				core.Text("Session Timeout"),
			),
			ui.Select(ui.SelectProps{
				Name:  "sessionTimeout",
				Value: sessionTimeout.Get(),
				Options: []ui.SelectOption{
					{Value: "15", Label: "15 minutes"},
					{Value: "30", Label: "30 minutes"},
					{Value: "60", Label: "1 hour"},
					{Value: "240", Label: "4 hours"},
					{Value: "1440", Label: "1 day"},
				},
				OnChange: func(v string) { sessionTimeout.Set(v) },
				Class:    "bg-gray-700 border-gray-600 text-white",
			}),
		),

		// Password minimum length
		core.Div(core.Attrs{},
			core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
				core.Text("Minimum Password Length"),
			),
			ui.Select(ui.SelectProps{
				Name:  "passwordMinLength",
				Value: passwordMinLength.Get(),
				Options: []ui.SelectOption{
					{Value: "6", Label: "6 characters"},
					{Value: "8", Label: "8 characters"},
					{Value: "10", Label: "10 characters"},
					{Value: "12", Label: "12 characters"},
					{Value: "16", Label: "16 characters"},
				},
				OnChange: func(v string) { passwordMinLength.Set(v) },
				Class:    "bg-gray-700 border-gray-600 text-white",
			}),
		),

		// Two-factor authentication toggle
		settingsRow(
			"Require Two-Factor Authentication",
			"Require all users to enable 2FA for their accounts",
			ui.Switch(ui.SwitchProps{
				ID:      "twoFactorRequired",
				Name:    "twoFactorRequired",
				Checked: twoFactorRequired.Get(),
			}),
			func() { twoFactorRequired.Set(!twoFactorRequired.Get()) },
		),

		// Save button
		core.Div(core.Class("pt-4"),
			ui.Button(ui.ButtonProps{
				Variant: ui.ButtonPrimary,
				OnClick: func() {
					successSecurity.Set("Security settings saved successfully")
				},
				Children: []core.Node{core.Text("Save Changes")},
			}),
		),
	}

	// Admin-only Danger Zone
	if userRole == "admin" {
		children = append(children,
			ui.Divider(ui.DividerProps{Class: "bg-gray-700 my-6"}),
			ui.Card(ui.CardProps{
				Class: "border-red-900 bg-red-900/10",
				Children: []core.Node{
					ui.CardHeader(ui.CardHeaderProps{
						Class: "border-red-900",
						Children: []core.Node{
							core.H3(core.Class("text-lg font-medium text-red-400"),
								core.Text("Danger Zone"),
							),
						},
					}),
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							core.P(core.Class("text-gray-400 text-sm mb-4"),
								core.Text("These actions are irreversible. Please proceed with caution."),
							),
							ui.Button(ui.ButtonProps{
								Variant: ui.ButtonDestructive,
								OnClick: func() {
									// In a real app, this would call an API to reset all sessions
								},
								Children: []core.Node{core.Text("Reset All User Sessions")},
							}),
						},
					}),
				},
			}),
		)
	}

	return []core.Node{
		ui.VStack(ui.StackProps{
			Gap:      "6",
			Children: children,
		}),
	}
}

// emailTab renders the Email settings content.
func emailTab(r *core.Router) []core.Node {
	smtpHost := r.StateString("smtpHost", "smtp.example.com")
	smtpPort := r.StateString("smtpPort", "587")
	smtpUser := r.StateString("smtpUser", "admin@example.com")
	emailFromName := r.StateString("emailFromName", "Admin Panel")
	emailFromAddress := r.StateString("emailFromAddress", "noreply@example.com")
	successEmail := r.StateString("successEmail", "")

	return []core.Node{
		ui.VStack(ui.StackProps{
			Gap: "6",
			Children: []core.Node{
				// Success message
				func() core.Node {
					if successEmail.Get() != "" {
						return core.Div(core.Class("bg-green-900/20 border border-green-800 text-green-400 rounded-lg px-4 py-3 mb-4"),
							core.Text(successEmail.Get()),
						)
					}
					return core.Frag()
				}(),

				// SMTP Configuration section header
				core.H3(core.Class("text-lg font-medium text-white"),
					core.Text("SMTP Configuration"),
				),
				core.P(core.Class("text-gray-400 text-sm -mt-4"),
					core.Text("SMTP credentials are managed via environment variables"),
				),

				// SMTP Host (readonly display)
				core.Div(core.Attrs{},
					core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
						core.Text("SMTP Host"),
					),
					ui.Input(ui.InputProps{
						Type:     ui.InputText,
						Name:     "smtpHost",
						Value:    smtpHost.Get(),
						Disabled: true,
						Class:    "bg-gray-700 border-gray-600 text-gray-400",
					}),
				),

				// SMTP Port (readonly display)
				core.Div(core.Attrs{},
					core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
						core.Text("SMTP Port"),
					),
					ui.Input(ui.InputProps{
						Type:     ui.InputText,
						Name:     "smtpPort",
						Value:    smtpPort.Get(),
						Disabled: true,
						Class:    "bg-gray-700 border-gray-600 text-gray-400",
					}),
				),

				// SMTP User (readonly display)
				core.Div(core.Attrs{},
					core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
						core.Text("SMTP User"),
					),
					ui.Input(ui.InputProps{
						Type:     ui.InputText,
						Name:     "smtpUser",
						Value:    smtpUser.Get(),
						Disabled: true,
						Class:    "bg-gray-700 border-gray-600 text-gray-400",
					}),
				),

				ui.Divider(ui.DividerProps{Class: "bg-gray-700"}),

				// Email Sender section header
				core.H3(core.Class("text-lg font-medium text-white"),
					core.Text("Email Sender"),
				),

				// From Name
				core.Div(core.Attrs{},
					core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
						core.Text("From Name"),
					),
					ui.Input(ui.InputProps{
						Type:     ui.InputText,
						Name:     "emailFromName",
						Value:    emailFromName.Get(),
						OnChange: func(v string) { emailFromName.Set(v) },
						Class:    "bg-gray-700 border-gray-600 text-white",
					}),
				),

				// From Address
				core.Div(core.Attrs{},
					core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
						core.Text("From Address"),
					),
					ui.Input(ui.InputProps{
						Type:     ui.InputEmail,
						Name:     "emailFromAddress",
						Value:    emailFromAddress.Get(),
						OnChange: func(v string) { emailFromAddress.Set(v) },
						Class:    "bg-gray-700 border-gray-600 text-white",
					}),
				),

				// Save button
				core.Div(core.Class("pt-4"),
					ui.Button(ui.ButtonProps{
						Variant: ui.ButtonPrimary,
						OnClick: func() {
							successEmail.Set("Email settings saved successfully")
						},
						Children: []core.Node{core.Text("Save Changes")},
					}),
				),
			},
		}),
	}
}

// settingsRow creates a row with label, description, and toggle control.
func settingsRow(title, description string, control core.Node, onClick func()) core.Node {
	return core.Div(
		core.Attrs{
			Class:   "flex items-center justify-between py-2 cursor-pointer",
			OnClick: onClick,
		},
		core.Div(core.Attrs{},
			core.Div(core.Class("text-white font-medium"),
				core.Text(title),
			),
			core.Div(core.Class("text-gray-400 text-sm"),
				core.Text(description),
			),
		),
		control,
	)
}
