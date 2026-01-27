package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Settings renders the settings page with breadcrumbs.
func Settings(r *core.Router) func() core.Node {
	return func() core.Node {
		// Form state
		siteName := r.StateString("siteName", "Test Admin")
		siteEmail := r.StateString("siteEmail", "admin@example.com")
		enableNotifications := r.StateBool("enableNotifications", true)
		enableDarkMode := r.StateBool("enableDarkMode", true)

		return core.Div(core.Class("p-6"),
			// Page header with breadcrumbs
			ui.PageHeader(ui.PageHeaderProps{
				Title:    "Settings",
				Subtitle: "Configure application settings",
				Breadcrumbs: []ui.BreadcrumbItem{
					{Label: "Admin", Href: "/admin"},
					{Label: "Settings"},
				},
			}),

			// Settings cards
			core.Div(core.Class("grid grid-cols-1 lg:grid-cols-2 gap-6 mt-6"),
				// General settings
				ui.Card(ui.CardProps{
					Class: "bg-gray-800 border-gray-700",
					Children: []core.Node{
						ui.CardHeader(ui.CardHeaderProps{
							Children: []core.Node{
								core.H3(core.Class("text-lg font-semibold text-white"), core.Text("General Settings")),
							},
						}),
						ui.CardContent(ui.CardContentProps{
							Children: []core.Node{
								// Site name field
								ui.FormField(ui.FormFieldProps{
									Label: "Site Name",
									Class: "mb-4",
									Children: []core.Node{
										ui.Input(ui.InputProps{
											Type:     ui.InputText,
											Value:    siteName.Get(),
											Class:    "bg-gray-700 border-gray-600 text-white",
											OnChange: func(v string) { siteName.Set(v) },
										}),
									},
								}),

								// Admin email field
								ui.FormField(ui.FormFieldProps{
									Label: "Admin Email",
									Class: "mb-4",
									Children: []core.Node{
										ui.Input(ui.InputProps{
											Type:     ui.InputEmail,
											Value:    siteEmail.Get(),
											Class:    "bg-gray-700 border-gray-600 text-white",
											OnChange: func(v string) { siteEmail.Set(v) },
										}),
									},
								}),

								// Save button
								core.Div(core.Class("mt-6"),
									ui.Button(ui.ButtonProps{
										Variant: ui.ButtonPrimary,
										Children: []core.Node{
											core.Text("Save Changes"),
										},
									}),
								),
							},
						}),
					},
				}),

				// Preferences
				ui.Card(ui.CardProps{
					Class: "bg-gray-800 border-gray-700",
					Children: []core.Node{
						ui.CardHeader(ui.CardHeaderProps{
							Children: []core.Node{
								core.H3(core.Class("text-lg font-semibold text-white"), core.Text("Preferences")),
							},
						}),
						ui.CardContent(ui.CardContentProps{
							Children: []core.Node{
								// Notifications toggle
								core.Div(core.Class("flex items-center justify-between py-4 border-b border-gray-700"),
									core.Div(core.Attrs{},
										core.Div(core.Class("text-white font-medium"), core.Text("Email Notifications")),
										core.Div(core.Class("text-gray-400 text-sm"), core.Text("Receive email notifications")),
									),
									core.Div(
										core.Attrs{
											Class:   "cursor-pointer",
											OnClick: func() { enableNotifications.Set(!enableNotifications.Get()) },
										},
										ui.Switch(ui.SwitchProps{
											Checked: enableNotifications.Get(),
										}),
									),
								),

								// Dark mode toggle
								core.Div(core.Class("flex items-center justify-between py-4"),
									core.Div(core.Attrs{},
										core.Div(core.Class("text-white font-medium"), core.Text("Dark Mode")),
										core.Div(core.Class("text-gray-400 text-sm"), core.Text("Enable dark theme")),
									),
									core.Div(
										core.Attrs{
											Class:   "cursor-pointer",
											OnClick: func() { enableDarkMode.Set(!enableDarkMode.Get()) },
										},
										ui.Switch(ui.SwitchProps{
											Checked: enableDarkMode.Get(),
										}),
									),
								),
							},
						}),
					},
				}),
			),
		)
	}
}
