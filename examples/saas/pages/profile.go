package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Profile is the user profile page with avatar and edit form.
func Profile(r *core.Router) func() core.Node {
	// Mock user data (in real app, would fetch from API)
	name := "John Doe"
	email := "john@example.com"
	role := "Admin"
	joinDate := "January 15, 2024"

	r.OnLoad(func() {
		// In a real app, this would fetch user data from API
		// The mock values above simulate loaded data
	})

	return func() core.Node {
		// State for profile data
		nameState := r.StateString("name", name)
		emailState := r.StateString("email", email)
		bioState := r.StateString("bio", "")
		editMode := r.StateBool("editMode", false)
		successState := r.StateString("success", "")
		errorState := r.StateString("error", "")

		return DashboardLayout(
			core.H1(core.Class("text-3xl font-bold text-white mb-8"),
				core.Text("Profile"),
			),

			// Success/error messages
			func() core.Node {
				if successState.Get() != "" {
					return core.Div(core.Class("bg-green-900/20 border border-green-800 text-green-400 rounded-lg px-4 py-3 mb-6"),
						core.Text(successState.Get()),
					)
				}
				return core.Frag()
			}(),
			func() core.Node {
				if errorState.Get() != "" {
					return core.Div(core.Class("bg-red-900/20 border border-red-800 text-red-400 rounded-lg px-4 py-3 mb-6"),
						core.Text(errorState.Get()),
					)
				}
				return core.Frag()
			}(),

			// Profile header card
			ui.Card(ui.CardProps{
				Class: "bg-gray-800 border border-gray-700 mb-8",
				Children: []core.Node{
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							core.Div(core.Class("flex items-start gap-6"),
								// Avatar
								ui.Avatar(ui.AvatarProps{
									Name: nameState.Get(),
									Size: ui.AvatarLG,
								}),

								// User info
								core.Div(core.Class("flex-1"),
									core.H2(core.Class("text-xl font-bold text-white"),
										core.Text(nameState.Get()),
									),
									core.Div(core.Class("text-gray-400"),
										core.Text(emailState.Get()),
									),
									core.Div(core.Class("flex items-center gap-4 mt-2"),
										ui.Badge(ui.BadgeProps{
											Text:  role,
											Class: "bg-purple-600 text-white",
										}),
										core.Span(core.Class("text-gray-500 text-sm"),
											core.Text("Member since "+joinDate),
										),
									),
								),

								// Edit button
								func() core.Node {
									if !editMode.Get() {
										return ui.Button(ui.ButtonProps{
											Variant: ui.ButtonPrimary,
											OnClick: func() {
												editMode.Set(true)
												successState.Set("")
												errorState.Set("")
											},
											Children: []core.Node{core.Text("Edit Profile")},
										})
									}
									return core.Frag()
								}(),
							),
						},
					}),
				},
			}),

			// Profile details / Edit form
			func() core.Node {
				if editMode.Get() {
					return profileEditForm(r, nameState, bioState, editMode, successState, errorState)
				}
				return profileViewMode(nameState, emailState, bioState)
			}(),

			// Danger zone
			dangerZone(r),
		)
	}
}

// profileViewMode renders the read-only profile view.
func profileViewMode(nameState, emailState, bioState *core.State[string]) core.Node {
	return ui.Card(ui.CardProps{
		Class: "bg-gray-800 border border-gray-700",
		Children: []core.Node{
			ui.CardHeader(ui.CardHeaderProps{
				Class: "border-gray-700",
				Children: []core.Node{
					core.H3(core.Class("text-lg font-semibold text-white"),
						core.Text("Profile Information"),
					),
				},
			}),
			ui.CardContent(ui.CardContentProps{
				Children: []core.Node{
					ui.VStack(ui.StackProps{
						Gap: "4",
						Children: []core.Node{
							// Name
							profileField("Full Name", nameState.Get()),
							// Email
							profileField("Email Address", emailState.Get()),
							// Bio
							func() core.Node {
								bio := bioState.Get()
								if bio == "" {
									bio = "No bio provided"
								}
								return profileField("Bio", bio)
							}(),
						},
					}),
				},
			}),
		},
	})
}

// profileField renders a single profile field row.
func profileField(label, value string) core.Node {
	return core.Div(core.Attrs{},
		core.Div(core.Class("text-sm font-medium text-gray-400 mb-1"),
			core.Text(label),
		),
		core.Div(core.Class("text-white"),
			core.Text(value),
		),
	)
}

// profileEditForm renders the profile edit form.
func profileEditForm(r *core.Router, nameState, bioState *core.State[string], editMode *core.State[bool], successState, errorState *core.State[string]) core.Node {
	return ui.Card(ui.CardProps{
		Class: "bg-gray-800 border border-gray-700",
		Children: []core.Node{
			ui.CardHeader(ui.CardHeaderProps{
				Class: "border-gray-700",
				Children: []core.Node{
					core.H3(core.Class("text-lg font-semibold text-white"),
						core.Text("Edit Profile"),
					),
				},
			}),
			ui.CardContent(ui.CardContentProps{
				Children: []core.Node{
					ui.VStack(ui.StackProps{
						Gap: "4",
						Children: []core.Node{
							// Name input
							core.Div(core.Attrs{},
								core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
									core.Text("Full Name"),
								),
								ui.Input(ui.InputProps{
									Type:     ui.InputText,
									Name:     "name",
									Value:    nameState.Get(),
									OnChange: func(v string) { nameState.Set(v) },
									Class:    "bg-gray-700 border-gray-600 text-white",
								}),
							),

							// Email (read-only)
							core.Div(core.Attrs{},
								core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
									core.Text("Email Address"),
								),
								ui.Input(ui.InputProps{
									Type:     ui.InputEmail,
									Name:     "email",
									Value:    "john@example.com",
									Disabled: true,
									Class:    "bg-gray-700 border-gray-600 text-gray-400",
								}),
								core.P(core.Class("text-sm text-gray-500 mt-1"),
									core.Text("Contact admin to change email"),
								),
							),

							// Bio textarea
							core.Div(core.Attrs{},
								core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
									core.Text("Bio"),
								),
								ui.Textarea(ui.TextareaProps{
									Name:        "bio",
									Value:       bioState.Get(),
									Placeholder: "Tell us about yourself...",
									Rows:        4,
									OnChange:    func(v string) { bioState.Set(v) },
									Class:       "bg-gray-700 border-gray-600 text-white",
								}),
							),

							// Action buttons
							core.Div(core.Class("flex gap-3 pt-4"),
								ui.Button(ui.ButtonProps{
									Variant: ui.ButtonPrimary,
									OnClick: func() {
										// Mock save
										successState.Set("Profile updated successfully")
										errorState.Set("")
										editMode.Set(false)
									},
									Children: []core.Node{core.Text("Save Changes")},
								}),
								ui.Button(ui.ButtonProps{
									Variant: ui.ButtonSecondary,
									OnClick: func() {
										editMode.Set(false)
										successState.Set("")
										errorState.Set("")
									},
									Children: []core.Node{core.Text("Cancel")},
								}),
							),
						},
					}),
				},
			}),
		},
	})
}

// dangerZone renders the account deletion section.
func dangerZone(r *core.Router) core.Node {
	return core.Div(core.Class("mt-8 bg-red-900/20 border border-red-800 rounded-lg p-6"),
		core.H3(core.Class("text-red-400 font-semibold mb-4"),
			core.Text("Danger Zone"),
		),
		core.P(core.Class("text-gray-400 text-sm mb-4"),
			core.Text("Once you delete your account, there is no going back. Please be certain."),
		),
		ui.Button(ui.ButtonProps{
			Variant: ui.ButtonDestructive,
			OnClick: func() {
				// In real app, would show confirmation modal
			},
			Children: []core.Node{core.Text("Delete Account")},
		}),
	)
}
