package pages

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/admin/.gux/api"
	"github.com/dougbarrett/gux/examples/admin/dto"
	"github.com/dougbarrett/gux/ui"
)

// Dashboard is the admin dashboard page with stats and recent activity.
func Dashboard(r *core.Router) func() core.Node {
	var users []dto.UserList
	var activities []dto.ActivityLogList

	r.OnLoad(func() {
		// Fetch all users to calculate stats
		api.Users.List(func(result []dto.UserList, err error) {
			if err == nil {
				users = result
			}
		})
		// Fetch activity logs for recent activity section
		api.ActivityLogs.List(func(result []dto.ActivityLogList, err error) {
			if err == nil {
				activities = result
			}
		})
	})

	return func() core.Node {
		// Calculate stats from loaded users
		totalUsers := len(users)
		activeUsers := 0
		suspendedUsers := 0
		pendingUsers := 0
		for _, u := range users {
			switch u.Status {
			case "active":
				activeUsers++
			case "suspended":
				suspendedUsers++
			case "pending":
				pendingUsers++
			}
		}

		// Store stats in state for hydration
		total := r.StateInt("totalUsers", totalUsers)
		active := r.StateInt("activeUsers", activeUsers)
		suspended := r.StateInt("suspendedUsers", suspendedUsers)
		pending := r.StateInt("pendingUsers", pendingUsers)

		// Store activities in state for hydration
		activitiesJSON, _ := json.Marshal(activities)
		activitiesState := r.StateString("activities", string(activitiesJSON))

		// Parse activities from state (for client-side)
		var displayActivities []dto.ActivityLogList
		json.Unmarshal([]byte(activitiesState.Get()), &displayActivities)

		return AdminLayout(
			core.H1(core.Class("text-3xl font-bold text-white mb-8"),
				core.Text("Dashboard"),
			),

			// Stats grid
			ui.Grid(ui.GridProps{
				Cols: "4",
				Gap:  "6",
				Children: []core.Node{
					statCard("Total Users", formatInt(total.Get()), "text-white"),
					statCard("Active Users", formatInt(active.Get()), "text-green-400"),
					statCard("Suspended", formatInt(suspended.Get()), "text-red-400"),
					statCard("Pending", formatInt(pending.Get()), "text-yellow-400"),
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
							recentActivityList(displayActivities),
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

// recentActivityList renders the list of recent activities.
func recentActivityList(activities []dto.ActivityLogList) core.Node {
	if len(activities) == 0 {
		return core.Div(core.Class("text-gray-400 py-4 text-center"),
			core.Text("No recent activity"),
		)
	}

	// Show up to 5 most recent activities
	limit := 5
	if len(activities) < limit {
		limit = len(activities)
	}

	var items []core.Node
	for i := 0; i < limit; i++ {
		items = append(items, dashboardActivityItem(activities[i].Description, formatActivityTime(activities[i].CreatedAt)))
	}
	return core.Div(core.Class("space-y-3"), items...)
}

// dashboardActivityItem renders a single activity item for the dashboard.
func dashboardActivityItem(description, time string) core.Node {
	return core.Div(core.Class("flex justify-between items-center py-2 border-b border-gray-700 last:border-0"),
		core.Span(core.Class("text-gray-300"), core.Text(description)),
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

// formatActivityTime formats a time.Time as "Jan 2, 3:04 PM" for dashboard display.
func formatActivityTime(t time.Time) string {
	return t.Format("Jan 2, 3:04 PM")
}
