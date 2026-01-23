package pages

import (
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Dashboard is the main dashboard page with stats and recent activity.
func Dashboard(r *core.Router) func() core.Node {
	// Loader - fetch dashboard stats (dummy data for now)
	totalProjects := 0
	activeProjects := 0
	completedProjects := 0

	r.OnLoad(func() {
		// In a real app, this would fetch from database
		totalProjects = 42
		activeProjects = 28
		completedProjects = 14
	})

	return func() core.Node {
		// Store stats in state for hydration
		total := r.StateInt("totalProjects", totalProjects)
		active := r.StateInt("activeProjects", activeProjects)
		completed := r.StateInt("completedProjects", completedProjects)

		return DashboardLayout(
			core.H1(core.Class("text-3xl font-bold text-white mb-8"),
				core.Text("Dashboard"),
			),

			// Stats grid
			ui.Grid(ui.GridProps{
				Cols: "3",
				Gap:  "6",
				Children: []core.Node{
					statCard("Total Projects", formatInt(total.Get()), "text-white"),
					statCard("Active Projects", formatInt(active.Get()), "text-green-400"),
					statCard("Completed", formatInt(completed.Get()), "text-blue-400"),
				},
			}),

			// Recent activity section
			ui.Card(ui.CardProps{
				Class: "mt-8 bg-gray-800 border border-gray-700",
				Children: []core.Node{
					ui.CardHeader(ui.CardHeaderProps{
						Class: "border-gray-700",
						Children: []core.Node{
							core.H2(core.Class("text-xl font-semibold text-white"),
								core.Text("Recent Activity"),
							),
						},
					}),
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							core.Div(core.Class("space-y-3"),
								activityItem("Project 'Dashboard' created", "2 minutes ago"),
								activityItem("Project 'API Gateway' completed", "15 minutes ago"),
								activityItem("Project 'Auth Service' updated", "1 hour ago"),
								activityItem("Project 'Frontend' archived", "3 hours ago"),
							),
						},
					}),
				},
			}),
		)
	}
}

// statCard renders a stat card with title, value, and color.
func statCard(title, value, colorClass string) core.Node {
	return ui.Card(ui.CardProps{
		Class: "bg-gray-800 border border-gray-700",
		Children: []core.Node{
			ui.CardContent(ui.CardContentProps{
				Children: []core.Node{
					core.Div(core.Class("text-gray-400 text-sm"),
						core.Text(title),
					),
					core.Div(core.Class("text-3xl font-bold mt-2 "+colorClass),
						core.Text(value),
					),
				},
			}),
		},
	})
}

// activityItem renders a single activity item.
func activityItem(title, time string) core.Node {
	return core.Div(core.Class("flex justify-between items-center py-2 border-b border-gray-700 last:border-0"),
		core.Span(core.Class("text-gray-300"), core.Text(title)),
		core.Span(core.Class("text-gray-500 text-sm"), core.Text(time)),
	)
}

// formatInt formats an integer with commas for thousands.
func formatInt(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d", n)
}
