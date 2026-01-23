package pages

import (
	"github.com/dougbarrett/gux/core"
)

// Profile is the user profile page.
func Profile(r *core.Router) func() core.Node {
	return func() core.Node {
		return DashboardLayout(
			core.H1(core.Class("text-3xl font-bold text-white"),
				core.Text("Profile"),
			),
		)
	}
}
