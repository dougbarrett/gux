package pages

import (
	"github.com/dougbarrett/gux/core"
)

// Projects is the project list page.
func Projects(r *core.Router) func() core.Node {
	return func() core.Node {
		return DashboardLayout(
			core.H1(core.Class("text-3xl font-bold text-white"),
				core.Text("Projects"),
			),
		)
	}
}

// ProjectNew is the create project page.
func ProjectNew(r *core.Router) func() core.Node {
	return func() core.Node {
		return DashboardLayout(
			core.H1(core.Class("text-3xl font-bold text-white"),
				core.Text("New Project"),
			),
		)
	}
}

// ProjectDetail is the project detail page.
func ProjectDetail(r *core.Router) func() core.Node {
	return func() core.Node {
		return DashboardLayout(
			core.H1(core.Class("text-3xl font-bold text-white"),
				core.Text("Project Detail"),
			),
		)
	}
}

// ProjectEdit is the edit project page.
func ProjectEdit(r *core.Router) func() core.Node {
	return func() core.Node {
		return DashboardLayout(
			core.H1(core.Class("text-3xl font-bold text-white"),
				core.Text("Edit Project"),
			),
		)
	}
}

// Settings is the settings page.
func Settings(r *core.Router) func() core.Node {
	return func() core.Node {
		return DashboardLayout(
			core.H1(core.Class("text-3xl font-bold text-white"),
				core.Text("Settings"),
			),
		)
	}
}

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
