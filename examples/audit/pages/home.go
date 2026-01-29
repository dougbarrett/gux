package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Home is the public landing page.
func Home(r *core.Router) func() core.Node {
	return func() core.Node {
		return core.Div(core.Class("min-h-screen bg-gray-900 flex items-center justify-center"),
			core.Div(core.Class("text-center max-w-lg"),
				core.H1(core.Class("text-4xl font-bold text-white mb-4"),
					core.Text("Document Manager"),
				),
				core.P(core.Class("text-gray-400 mb-8"),
					core.Text("A secure document management system with full audit logging for compliance tracking."),
				),
				ui.Button(ui.ButtonProps{
					Variant: ui.ButtonPrimary,
					Size:    ui.ButtonLG,
					Children: []core.Node{core.Text("Sign In")},
					OnClick: func() {
						r.Navigate("/login")
					},
				}),
			),
		)
	}
}
