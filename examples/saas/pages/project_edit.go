package pages

import (
	"encoding/json"
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/saas/.gux/api"
	"github.com/dougbarrett/gux/examples/saas/dto"
	"github.com/dougbarrett/gux/ui"
)

// ProjectEdit is the form for editing an existing project.
func ProjectEdit(r *core.Router) func() core.Node {
	var project *dto.ProjectDetail

	r.OnLoad(func() {
		// Get project ID from route params
		idStr := r.GetRouteParams()["id"]
		var id uint
		fmt.Sscanf(idStr, "%d", &id)

		if id > 0 {
			api.Projects.Get(id, func(result *dto.ProjectDetail, err error) {
				if err == nil {
					project = result
				}
			})
		}
	})

	return func() core.Node {
		// Store project in state for hydration
		projectJSON, _ := json.Marshal(project)
		projectState := r.StateString("project", string(projectJSON))

		// Parse project from state
		var displayProject *dto.ProjectDetail
		json.Unmarshal([]byte(projectState.Get()), &displayProject)

		if displayProject == nil {
			return DashboardLayout(r, "/projects",
				core.Div(core.Class("text-white text-center py-12"),
					core.Text("Project not found"),
				),
			)
		}

		// Form state - initialize with current values
		nameState := r.StateString("name", displayProject.Name)
		descriptionState := r.StateString("description", displayProject.Description)
		statusState := r.StateString("status", displayProject.Status)
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
			// Validate required fields
			if nameState.Get() == "" {
				errorState.Set("Project name is required")
				return
			}

			// Build update data
			data := map[string]interface{}{
				"name":        nameState.Get(),
				"description": descriptionState.Get(),
				"status":      statusState.Get(),
			}

			api.Projects.Update(displayProject.ID, data, func(result *dto.ProjectDetail, err error) {
				if err != nil {
					errorState.Set(err.Error())
				} else {
					successState.Set("Project updated successfully!")
					errorState.Set("")
					// Navigate back to project detail
					r.Navigate(fmt.Sprintf("/projects/%d", displayProject.ID))
				}
			})
		}

		return DashboardLayout(r, "/projects",
			// Back link
			core.A(
				core.Attrs{Href: fmt.Sprintf("/projects/%d", displayProject.ID), Class: "text-blue-400 hover:text-blue-300 mb-4 inline-block"},
				core.Text("\u2190 Back to Project"),
			),

			core.H1(core.Class("text-2xl font-bold text-white mb-6"),
				core.Text(fmt.Sprintf("Edit Project: %s", displayProject.Name)),
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
									Class:   "w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition",
									OnClick: save,
								},
								core.Text("Save Changes"),
							),
						},
					}),
				},
			}),
		)
	}
}
