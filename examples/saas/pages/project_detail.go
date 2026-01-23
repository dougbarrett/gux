package pages

import (
	"encoding/json"
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/saas/.gux/api"
	"github.com/dougbarrett/gux/examples/saas/dto"
	"github.com/dougbarrett/gux/ui"
)

// ProjectDetail shows a single project with edit/delete options.
func ProjectDetail(r *core.Router) func() core.Node {
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

		// Delete modal state
		deleteModalOpen := r.StateBool("deleteModal", false)

		if displayProject == nil {
			return DashboardLayout(
				core.Div(core.Class("text-white text-center py-12"),
					core.Text("Project not found"),
				),
			)
		}

		return DashboardLayout(
			// Back link
			core.A(
				core.Attrs{Href: "/projects", Class: "text-blue-400 hover:text-blue-300 mb-4 inline-block"},
				core.Text("\u2190 Back to Projects"),
			),

			// Project info card
			ui.Card(ui.CardProps{
				Class: "bg-gray-800 border border-gray-700 mb-6",
				Children: []core.Node{
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							core.Div(core.Class("flex justify-between items-start"),
								core.Div(core.Attrs{},
									core.H1(core.Class("text-2xl font-bold text-white mb-2"),
										core.Text(displayProject.Name),
									),
									core.Div(core.Class("text-gray-400 mb-4"),
										core.Text(displayProject.Description),
									),
									core.Div(core.Class("flex items-center gap-4"),
										statusBadge(displayProject.Status),
										core.Span(core.Class("text-gray-500 text-sm"),
											core.Text(fmt.Sprintf("Created %s", displayProject.CreatedAt.Format("Jan 2, 2006"))),
										),
									),
								),
								core.Div(core.Class("flex gap-2"),
									core.A(
										core.Attrs{
											Href:  fmt.Sprintf("/projects/%d/edit", displayProject.ID),
											Class: "px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition",
										},
										core.Text("Edit"),
									),
									core.Button(
										core.Attrs{
											Type:  "button",
											Class: "px-4 py-2 bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg transition",
											OnClick: func() {
												deleteModalOpen.Set(true)
											},
										},
										core.Text("Delete"),
									),
								),
							),
						},
					}),
				},
			}),

			// Delete confirmation modal
			deleteConfirmationModal(r, displayProject, deleteModalOpen),
		)
	}
}

// deleteConfirmationModal renders the delete confirmation modal.
func deleteConfirmationModal(r *core.Router, project *dto.ProjectDetail, openState *core.State[bool]) core.Node {
	return ui.Modal(ui.ModalProps{
		Open:  openState.Get(),
		Title: "Delete Project",
		Size:  ui.ModalSM,
		OnClose: func() {
			openState.Set(false)
		},
		Children: []core.Node{
			ui.ModalContent(ui.ModalContentProps{
				Children: []core.Node{
					core.P(core.Class("text-gray-600 dark:text-gray-300"),
						core.Text(fmt.Sprintf("Are you sure you want to delete \"%s\"? This action cannot be undone.", project.Name)),
					),
				},
			}),
			ui.ModalFooter(ui.ModalFooterProps{
				Children: []core.Node{
					core.Button(
						core.Attrs{
							Type:  "button",
							Class: "px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white font-medium rounded-lg transition",
							OnClick: func() {
								openState.Set(false)
							},
						},
						core.Text("Cancel"),
					),
					core.Button(
						core.Attrs{
							Type:  "button",
							Class: "px-4 py-2 bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg transition",
							OnClick: func() {
								api.Projects.Delete(project.ID, func(err error) {
									if err == nil {
										r.Navigate("/projects")
									}
								})
							},
						},
						core.Text("Delete"),
					),
				},
			}),
		},
	})
}
