package admin

import (
	"fmt"

	"github.com/dougbarrett/gux/core"
)

// Dashboard is the admin dashboard page.
func Dashboard(r *core.Router) func() core.Node {
	// Loader - fetch admin stats (dummy data for now)
	totalUsers := 0
	totalOrders := 0
	revenue := "$0"

	r.OnLoad(func() {
		// In a real app, this would fetch from database
		totalUsers = 1234
		totalOrders = 567
		revenue = "$45,678"
	})

	return func() core.Node {
		users := r.StateInt("totalUsers", totalUsers)
		orders := r.StateInt("totalOrders", totalOrders)
		rev := r.StateString("revenue", revenue)

		return core.Div(core.Class("min-h-screen bg-gray-900"),
			Nav(),
			core.Div(core.Class("max-w-6xl mx-auto px-4 py-8"),
				core.H1(core.Class("text-3xl font-bold text-white mb-8"),
					core.Text("Admin Dashboard"),
				),

				// Stats cards
				core.Div(core.Class("grid grid-cols-1 md:grid-cols-3 gap-6"),
					// Users card
					core.Div(core.Class("bg-gray-800 rounded-lg p-6"),
						core.Div(core.Class("text-gray-400 text-sm"),
							core.Text("Total Users"),
						),
						core.Div(core.Class("text-3xl font-bold text-white mt-2"),
							core.Text(formatInt(users.Get())),
						),
					),

					// Orders card
					core.Div(core.Class("bg-gray-800 rounded-lg p-6"),
						core.Div(core.Class("text-gray-400 text-sm"),
							core.Text("Total Orders"),
						),
						core.Div(core.Class("text-3xl font-bold text-white mt-2"),
							core.Text(formatInt(orders.Get())),
						),
					),

					// Revenue card
					core.Div(core.Class("bg-gray-800 rounded-lg p-6"),
						core.Div(core.Class("text-gray-400 text-sm"),
							core.Text("Revenue"),
						),
						core.Div(core.Class("text-3xl font-bold text-green-400 mt-2"),
							core.Text(rev.Get()),
						),
					),
				),

				// Recent activity section
				core.Div(core.Class("mt-8 bg-gray-800 rounded-lg p-6"),
					core.H2(core.Class("text-xl font-semibold text-white mb-4"),
						core.Text("Recent Activity"),
					),
					core.Div(core.Class("space-y-3"),
						activityItem("New user registered", "2 minutes ago"),
						activityItem("Order #1234 completed", "15 minutes ago"),
						activityItem("Payment received", "1 hour ago"),
						activityItem("New support ticket", "3 hours ago"),
					),
				),
			),
		)
	}
}

func activityItem(title, time string) core.Node {
	return core.Div(core.Class("flex justify-between items-center py-2 border-b border-gray-700"),
		core.Span(core.Class("text-gray-300"), core.Text(title)),
		core.Span(core.Class("text-gray-500 text-sm"), core.Text(time)),
	)
}

func formatInt(n int) string {
	// Simple number formatting
	if n >= 1000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d", n)
}
