package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Settings is the settings page with tabbed sections.
func Settings(r *core.Router) func() core.Node {
	return func() core.Node {
		activeTab := r.StateInt("activeTab", 0)

		return DashboardLayout(r, "/settings",
			core.H1(core.Class("text-3xl font-bold text-white mb-8"),
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
												Children: []core.Node{core.Text("Notifications")},
											}),
											ui.Tab(ui.TabProps{
												Active:   activeTab.Get() == 2,
												OnSelect: func() { activeTab.Set(2) },
												Children: []core.Node{core.Text("Security")},
											}),
										},
									}),
									ui.TabPanel(ui.TabPanelProps{Active: activeTab.Get() == 0, Children: generalTab(r)}),
									ui.TabPanel(ui.TabPanelProps{Active: activeTab.Get() == 1, Children: notificationsTab(r)}),
									ui.TabPanel(ui.TabPanelProps{Active: activeTab.Get() == 2, Children: securityTab(r)}),
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
	appName := r.StateString("appName", "My SaaS App")
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

				// App name field
				core.Div(core.Attrs{},
					core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
						core.Text("Application Name"),
					),
					ui.Input(ui.InputProps{
						Type:     ui.InputText,
						Name:     "appName",
						Value:    appName.Get(),
						OnChange: func(v string) { appName.Set(v) },
						Class:    "bg-gray-700 border-gray-600 text-white",
					}),
				),

				// Timezone select
				core.Div(core.Attrs{},
					core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
						core.Text("Timezone"),
					),
					ui.Select(ui.SelectProps{
						Name:        "timezone",
						Value:       timezone.Get(),
						Placeholder: "Select timezone",
						Options: []ui.SelectOption{
							{Value: "UTC", Label: "UTC"},
							{Value: "America/New_York", Label: "America/New York"},
							{Value: "Europe/London", Label: "Europe/London"},
						},
						OnChange: func(v string) { timezone.Set(v) },
						Class:    "bg-gray-700 border-gray-600 text-white",
					}),
				),

				// Language select
				core.Div(core.Attrs{},
					core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
						core.Text("Language"),
					),
					ui.Select(ui.SelectProps{
						Name:        "language",
						Value:       language.Get(),
						Placeholder: "Select language",
						Options: []ui.SelectOption{
							{Value: "en", Label: "English"},
							{Value: "es", Label: "Spanish"},
							{Value: "fr", Label: "French"},
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
							successGeneral.Set("Settings saved successfully")
						},
						Children: []core.Node{core.Text("Save Changes")},
					}),
				),
			},
		}),
	}
}

// notificationsTab renders the Notifications settings content.
func notificationsTab(r *core.Router) []core.Node {
	emailNotifs := r.StateBool("emailNotifs", true)
	pushNotifs := r.StateBool("pushNotifs", false)
	weeklyDigest := r.StateBool("weeklyDigest", true)
	successNotifs := r.StateString("successNotifs", "")

	return []core.Node{
		ui.VStack(ui.StackProps{
			Gap: "6",
			Children: []core.Node{
				// Success message
				func() core.Node {
					if successNotifs.Get() != "" {
						return core.Div(core.Class("bg-green-900/20 border border-green-800 text-green-400 rounded-lg px-4 py-3 mb-4"),
							core.Text(successNotifs.Get()),
						)
					}
					return core.Frag()
				}(),

				// Email notifications
				settingRow(
					"Email Notifications",
					"Receive email notifications for important updates",
					ui.Switch(ui.SwitchProps{
						ID:      "emailNotifs",
						Name:    "emailNotifs",
						Checked: emailNotifs.Get(),
					}),
					func() { emailNotifs.Set(!emailNotifs.Get()) },
				),

				ui.Divider(ui.DividerProps{Class: "bg-gray-700"}),

				// Push notifications
				settingRow(
					"Push Notifications",
					"Receive push notifications in your browser",
					ui.Switch(ui.SwitchProps{
						ID:      "pushNotifs",
						Name:    "pushNotifs",
						Checked: pushNotifs.Get(),
					}),
					func() { pushNotifs.Set(!pushNotifs.Get()) },
				),

				ui.Divider(ui.DividerProps{Class: "bg-gray-700"}),

				// Weekly digest
				settingRow(
					"Weekly Digest",
					"Receive a weekly summary of your activity",
					ui.Switch(ui.SwitchProps{
						ID:      "weeklyDigest",
						Name:    "weeklyDigest",
						Checked: weeklyDigest.Get(),
					}),
					func() { weeklyDigest.Set(!weeklyDigest.Get()) },
				),

				// Save button
				core.Div(core.Class("pt-4"),
					ui.Button(ui.ButtonProps{
						Variant: ui.ButtonPrimary,
						OnClick: func() {
							successNotifs.Set("Notification preferences saved")
						},
						Children: []core.Node{core.Text("Save Preferences")},
					}),
				),
			},
		}),
	}
}

// settingRow creates a row with label, description, and toggle control.
func settingRow(title, description string, control core.Node, onClick func()) core.Node {
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

// securityTab renders the Security settings content.
func securityTab(r *core.Router) []core.Node {
	currentPassword := r.StateString("currentPassword", "")
	newPassword := r.StateString("newPassword", "")
	confirmPassword := r.StateString("confirmPassword", "")
	twoFactor := r.StateBool("twoFactor", false)
	successSecurity := r.StateString("successSecurity", "")
	errorSecurity := r.StateString("errorSecurity", "")

	return []core.Node{
		ui.VStack(ui.StackProps{
			Gap: "6",
			Children: []core.Node{
				// Success/error messages
				func() core.Node {
					if successSecurity.Get() != "" {
						return core.Div(core.Class("bg-green-900/20 border border-green-800 text-green-400 rounded-lg px-4 py-3 mb-4"),
							core.Text(successSecurity.Get()),
						)
					}
					return core.Frag()
				}(),
				func() core.Node {
					if errorSecurity.Get() != "" {
						return core.Div(core.Class("bg-red-900/20 border border-red-800 text-red-400 rounded-lg px-4 py-3 mb-4"),
							core.Text(errorSecurity.Get()),
						)
					}
					return core.Frag()
				}(),

				// Change password section
				core.Div(core.Attrs{},
					core.H3(core.Class("text-lg font-medium text-white mb-4"),
						core.Text("Change Password"),
					),

					ui.VStack(ui.StackProps{
						Gap: "4",
						Children: []core.Node{
							// Current password
							core.Div(core.Attrs{},
								core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
									core.Text("Current Password"),
								),
								ui.Input(ui.InputProps{
									Type:     ui.InputPassword,
									Name:     "currentPassword",
									Value:    currentPassword.Get(),
									OnChange: func(v string) { currentPassword.Set(v) },
									Class:    "bg-gray-700 border-gray-600 text-white",
								}),
							),

							// New password
							core.Div(core.Attrs{},
								core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
									core.Text("New Password"),
								),
								ui.Input(ui.InputProps{
									Type:     ui.InputPassword,
									Name:     "newPassword",
									Value:    newPassword.Get(),
									OnChange: func(v string) { newPassword.Set(v) },
									Class:    "bg-gray-700 border-gray-600 text-white",
								}),
							),

							// Confirm password
							core.Div(core.Attrs{},
								core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
									core.Text("Confirm New Password"),
								),
								ui.Input(ui.InputProps{
									Type:     ui.InputPassword,
									Name:     "confirmPassword",
									Value:    confirmPassword.Get(),
									OnChange: func(v string) { confirmPassword.Set(v) },
									Class:    "bg-gray-700 border-gray-600 text-white",
								}),
							),

							// Change password button
							ui.Button(ui.ButtonProps{
								Variant: ui.ButtonPrimary,
								OnClick: func() {
									errorSecurity.Set("")
									successSecurity.Set("")
									if currentPassword.Get() == "" {
										errorSecurity.Set("Please enter your current password")
										return
									}
									if newPassword.Get() == "" {
										errorSecurity.Set("Please enter a new password")
										return
									}
									if len(newPassword.Get()) < 8 {
										errorSecurity.Set("Password must be at least 8 characters")
										return
									}
									if newPassword.Get() != confirmPassword.Get() {
										errorSecurity.Set("Passwords do not match")
										return
									}
									successSecurity.Set("Password changed successfully")
									currentPassword.Set("")
									newPassword.Set("")
									confirmPassword.Set("")
								},
								Children: []core.Node{core.Text("Change Password")},
							}),
						},
					}),
				),

				ui.Divider(ui.DividerProps{Class: "bg-gray-700"}),

				// Two-factor authentication section
				core.Div(core.Attrs{},
					core.H3(core.Class("text-lg font-medium text-white mb-4"),
						core.Text("Two-Factor Authentication"),
					),
					settingRow(
						"Enable 2FA",
						"Add an extra layer of security to your account",
						ui.Switch(ui.SwitchProps{
							ID:      "twoFactor",
							Name:    "twoFactor",
							Checked: twoFactor.Get(),
						}),
						func() { twoFactor.Set(!twoFactor.Get()) },
					),
				),
			},
		}),
	}
}
