package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// valueCard renders a company value card.
func valueCard(title, description string) core.Node {
	return ui.Card(ui.CardProps{
		Children: []core.Node{
			ui.CardContent(ui.CardContentProps{
				Class: "text-center",
				Children: []core.Node{
					core.H3(core.Class("text-lg font-semibold text-gray-900 dark:text-white mb-2"),
						core.Text(title)),
					core.P(core.Class("text-gray-600 dark:text-gray-400"),
						core.Text(description)),
				},
			}),
		},
	})
}

// teamMemberCard renders a team member card with avatar.
func teamMemberCard(name, title, initials string) core.Node {
	return ui.Card(ui.CardProps{
		Children: []core.Node{
			ui.CardContent(ui.CardContentProps{
				Class: "text-center",
				Children: []core.Node{
					core.Div(core.Class("flex justify-center mb-4"),
						ui.Avatar(ui.AvatarProps{
							Name: name,
							Size: ui.AvatarLG,
						}),
					),
					core.H3(core.Class("font-semibold text-gray-900 dark:text-white"),
						core.Text(name)),
					core.P(core.Class("text-sm text-gray-600 dark:text-gray-400"),
						core.Text(title)),
				},
			}),
		},
	})
}

// About renders the about page with company story, values, and team.
func About(r *core.Router) func() core.Node {
	return func() core.Node {
		return MarketingLayout(r,
			// Story Section
			Section("Our Story", "",
				core.Div(core.Class("max-w-3xl mx-auto text-center"),
					core.P(core.Class("text-lg text-gray-600 dark:text-gray-400 mb-6"),
						core.Text("We started MarketCo in 2020 with a simple mission: make building great products easier for everyone. What began as a side project has grown into a platform trusted by thousands of teams worldwide.")),
					core.P(core.Class("text-lg text-gray-600 dark:text-gray-400"),
						core.Text("We believe in transparency, simplicity, and putting developers first. Every feature we build is designed to help you ship faster and focus on what matters most - your users.")),
				),
			),

			// Values Section
			Section("Our Values",
				"The principles that guide everything we do",
				ui.Grid(ui.GridProps{
					Cols:  "4",
					Gap:   "6",
					Class: "md:grid-cols-4 sm:grid-cols-2 grid-cols-1",
					Children: []core.Node{
						valueCard("Customer First",
							"Every decision starts with what's best for our customers"),
						valueCard("Ship Fast",
							"We believe in rapid iteration and continuous improvement"),
						valueCard("Stay Humble",
							"We're always learning and open to feedback"),
						valueCard("Think Long-term",
							"We build for sustainability, not quick wins"),
					},
				}),
			),

			// Team Section
			Section("Meet the Team",
				"The people behind the product",
				ui.Grid(ui.GridProps{
					Cols:  "4",
					Gap:   "6",
					Class: "md:grid-cols-4 grid-cols-2",
					Children: []core.Node{
						teamMemberCard("Alex Morgan", "CEO & Co-founder", "AM"),
						teamMemberCard("Jordan Lee", "CTO & Co-founder", "JL"),
						teamMemberCard("Sam Taylor", "Head of Product", "ST"),
						teamMemberCard("Chris Martinez", "Head of Engineering", "CM"),
					},
				}),
			),

			// CTA Section
			core.Section(core.Class("py-20 px-4 bg-gray-50 dark:bg-gray-800"),
				core.Div(core.Class("max-w-2xl mx-auto text-center"),
					core.H2(core.Class("text-3xl font-bold text-gray-900 dark:text-white mb-4"),
						core.Text("Want to join us?")),
					core.P(core.Class("text-lg text-gray-600 dark:text-gray-400 mb-8"),
						core.Text("We're always looking for talented people who share our values.")),
					ui.Button(ui.ButtonProps{
						Variant: ui.ButtonPrimary,
						Size:    ui.ButtonLG,
						OnClick: func() {
							// Demo: placeholder for careers link
						},
						Children: []core.Node{
							core.Text("View Open Positions"),
						},
					}),
				),
			),
		)
	}
}
