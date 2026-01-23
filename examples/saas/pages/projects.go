package pages

import (
	"encoding/json"
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/saas/.gux/api"
	"github.com/dougbarrett/gux/examples/saas/dto"
	"github.com/dougbarrett/gux/ui"
)

// Projects is the project list page with DataTable.
func Projects(r *core.Router) func() core.Node {
	var projects []dto.ProjectList

	r.OnLoad(func() {
		// Fetch projects via generated API client
		api.Projects.List(func(result []dto.ProjectList, err error) {
			if err == nil {
				projects = result
			}
		})
	})

	return func() core.Node {
		// Store projects in state for hydration
		projectsJSON, _ := json.Marshal(projects)
		projectsState := r.StateString("projects", string(projectsJSON))

		// Parse projects from state (for client-side)
		var displayProjects []dto.ProjectList
		json.Unmarshal([]byte(projectsState.Get()), &displayProjects)

		return DashboardLayout(
			core.H1(core.Class("text-3xl font-bold text-white mb-8"),
				core.Text("Projects"),
			),

			// Add Project button
			core.Div(core.Class("mb-6"),
				core.A(
					core.Attrs{Href: "/projects/new", Class: "inline-flex items-center px-4 py-2 bg-green-600 hover:bg-green-700 text-white font-medium rounded-lg transition"},
					core.Text("+ Add Project"),
				),
			),

			// Projects DataTable
			ui.Card(ui.CardProps{
				Class: "bg-gray-800 border border-gray-700",
				Children: []core.Node{
					projectsTable(r, displayProjects),
				},
			}),
		)
	}
}

// projectsTable renders the projects DataTable.
func projectsTable(r *core.Router, projects []dto.ProjectList) core.Node {
	if len(projects) == 0 {
		return ui.CardContent(ui.CardContentProps{
			Children: []core.Node{
				core.Div(core.Class("text-gray-400 py-8 text-center"),
					core.Text("No projects found. Create your first project!"),
				),
			},
		})
	}

	return ui.DataTable(ui.DataTableProps[dto.ProjectList]{
		Data: projects,
		Columns: []ui.ColumnDef[dto.ProjectList]{
			{
				Header: "Name",
				Render: func(p dto.ProjectList) core.Node {
					return core.A(
						core.Attrs{
							Href:  fmt.Sprintf("/projects/%d", p.ID),
							Class: "text-blue-400 hover:text-blue-300 font-medium",
						},
						core.Text(p.Name),
					)
				},
			},
			{
				Header: "Status",
				Render: func(p dto.ProjectList) core.Node {
					return statusBadge(p.Status)
				},
			},
			{
				Header: "Created",
				Render: func(p dto.ProjectList) core.Node {
					return core.Text(p.CreatedAt.Format("Jan 2, 2006"))
				},
			},
		},
		Striped:   true,
		Hoverable: true,
		OnRowClick: func(p dto.ProjectList) {
			r.Navigate(fmt.Sprintf("/projects/%d", p.ID))
		},
		Class: "dark:bg-gray-800",
	})
}

// statusBadge renders a badge for project status.
func statusBadge(status string) core.Node {
	var variant ui.BadgeVariant
	switch status {
	case "active":
		variant = ui.BadgeSuccess
	case "completed":
		variant = ui.BadgeDefault
	case "archived":
		variant = ui.BadgeWarning
	default:
		variant = ui.BadgeDefault
	}
	return ui.Badge(ui.BadgeProps{
		Text:    status,
		Variant: variant,
	})
}
