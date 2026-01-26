package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/saas/.gux/api"
	"github.com/dougbarrett/gux/examples/saas/dto"
	"github.com/dougbarrett/gux/ui"
)

// ProjectNew is the form for creating a new project.
func ProjectNew(r *core.Router) func() core.Node {
	return func() core.Node {
		// Form state
		nameState := r.StateString("name", "")
		descriptionState := r.StateString("description", "")
		statusState := r.StateString("status", "active")
		errorState := r.StateString("error", "")
		successState := r.StateString("success", "")

		var messageNode core.Node = core.Frag()
		if successState.Get() != "" {
			messageNode = core.Div(core.Class("mb-4 p-4 bg-green-900 border border-green-700 rounded-lg text-green-300"),
				core.Text(successState.Get()),
			)
		} else if errorState.Get() != "" {
			messageNode = core.Div(core.Class("mb-4 p-4 bg-red-900 border border-red-700 rounded-lg text-red-300"),
				core.Text(errorState.Get()),
			)
		}

		// Save function - shared by button click and Enter key
		save := func() {
			// Validate
			if nameState.Get() == "" {
				errorState.Set("Project name is required")
				return
			}

			// Create project via generated API
			data := map[string]interface{}{
				"name":        nameState.Get(),
				"description": descriptionState.Get(),
				"status":      statusState.Get(),
			}
			api.Projects.Create(data, func(result *dto.ProjectDetail, err error) {
				if err != nil {
					errorState.Set(err.Error())
				} else {
					successState.Set("Project created successfully!")
					// Clear form
					nameState.Set("")
					descriptionState.Set("")
					statusState.Set("active")
					errorState.Set("")
					// Navigate to projects list
					r.Navigate("/projects")
				}
			})
		}

		return DashboardLayout(r, "/projects",
			// Back link
			core.A(
				core.Attrs{Href: "/projects", Class: "text-blue-400 hover:text-blue-300 mb-4 inline-block"},
				core.Text("\u2190 Back to Projects"),
			),

			core.H1(core.Class("text-2xl font-bold text-white mb-6"),
				core.Text("Create New Project"),
			),

			messageNode,

			// Form
			ui.Card(ui.CardProps{
				Class: "bg-gray-800 border border-gray-700",
				Children: []core.Node{
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							// Name field
							formField("Name", "text", "name", nameState.Get(), func(v string) {
								nameState.Set(v)
							}, save),

							// Description field
							core.Div(core.Class("mb-4"),
								core.Label(core.Class("block text-gray-300 text-sm font-medium mb-2"),
									core.Text("Description"),
								),
								core.Textarea(
									core.Attrs{
										Name:  "description",
										Class: "w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:border-blue-500",
										OnChange: func(v string) {
											descriptionState.Set(v)
										},
										Extra: map[string]string{"rows": "4"},
									},
									core.Text(descriptionState.Get()),
								),
							),

							// Status select
							core.Div(core.Class("mb-6"),
								core.Label(core.Class("block text-gray-300 text-sm font-medium mb-2"),
									core.Text("Status"),
								),
								core.Select(
									core.Attrs{
										Name:  "status",
										Class: "w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:border-blue-500",
										OnChange: func(v string) {
											statusState.Set(v)
										},
									},
									selectOption("active", "Active", statusState.Get()),
									selectOption("completed", "Completed", statusState.Get()),
									selectOption("archived", "Archived", statusState.Get()),
								),
							),

							// Submit button
							core.Button(
								core.Attrs{
									Type:    "button",
									Class:   "w-full px-4 py-2 bg-green-600 hover:bg-green-700 text-white font-medium rounded-lg transition",
									OnClick: save,
								},
								core.Text("Create Project"),
							),
						},
					}),
				},
			}),
		)
	}
}

// formField renders a form input field with label.
func formField(label, inputType, name, value string, onChange func(string), onEnter ...func()) core.Node {
	attrs := core.Attrs{
		Type:     inputType,
		Name:     name,
		Value:    value,
		Class:    "w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:border-blue-500",
		OnChange: onChange,
	}
	if len(onEnter) > 0 {
		attrs.OnEnter = onEnter[0]
	}
	return core.Div(core.Class("mb-4"),
		core.Label(core.Class("block text-gray-300 text-sm font-medium mb-2"),
			core.Text(label),
		),
		core.Input(attrs),
	)
}

// selectOption creates an option element with selected state.
func selectOption(value, label, selectedValue string) core.Node {
	attrs := core.Attrs{Value: value}
	if value == selectedValue {
		attrs.Extra = map[string]string{"selected": "selected"}
	}
	return core.Option(attrs, core.Text(label))
}
