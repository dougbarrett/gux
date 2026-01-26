package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Home renders the home page with hero, features, testimonials, and CTA sections.
func Home(r *core.Router) func() core.Node {
	return func() core.Node {
		return MarketingLayout(r,
			// Hero Section
			Hero(
				"Build Products People Love",
				"The modern platform for teams who ship fast. Start free, scale infinitely.",
				ui.Button(ui.ButtonProps{
					Variant: ui.ButtonGhost,
					Size:    ui.ButtonLG,
					OnClick: func() { r.Navigate("/pricing") },
					Class:   "bg-white text-blue-600 hover:bg-gray-100 font-semibold",
					Children: []core.Node{
						core.Text("Start Free Trial"),
					},
				}),
				ui.Button(ui.ButtonProps{
					Variant: ui.ButtonOutline,
					Size:    ui.ButtonLG,
					OnClick: func() { r.Navigate("/features") },
					Class:   "border-white text-white hover:bg-white/10",
					Children: []core.Node{
						core.Text("Learn More"),
					},
				}),
			),

			// Features Section
			Section("Everything You Need", "Powerful features to help you build, deploy, and scale.",
				ui.Grid(ui.GridProps{
					Cols:  "3",
					Gap:   "6",
					Class: "md:grid-cols-3 grid-cols-1",
					Children: []core.Node{
						featureCard("\u26a1", "Lightning Fast", "Blazing fast performance with edge deployment"),
						featureCard("\U0001F512", "Secure by Default", "Enterprise-grade security built in from day one"),
						featureCard("\U0001F4C8", "Scale Infinitely", "Auto-scaling infrastructure that grows with you"),
						featureCard("\U0001F4BB", "Developer First", "APIs and SDKs designed for developer happiness"),
						featureCard("\U0001F4CA", "Real-time Analytics", "Live insights into your application performance"),
						featureCard("\U0001F3A7", "24/7 Support", "Expert support team available around the clock"),
					},
				}),
			),

			// Testimonials Section
			Section("Loved by Teams Worldwide", "See what our customers have to say",
				ui.Grid(ui.GridProps{
					Cols:  "3",
					Gap:   "6",
					Class: "md:grid-cols-3 grid-cols-1",
					Children: []core.Node{
						testimonialCard(
							"This platform transformed how we ship products. Incredible experience.",
							"Sarah Chen",
							"CTO at TechStart",
						),
						testimonialCard(
							"Finally, a tool that just works. Our team productivity doubled.",
							"Marcus Johnson",
							"Lead Developer at BuildCo",
						),
						testimonialCard(
							"The best investment we made this year. Worth every penny.",
							"Emily Rodriguez",
							"Founder of GrowthLabs",
						),
					},
				}),
			),

			// Bottom CTA Section
			core.Section(core.Class("py-20 px-4 bg-gray-100 dark:bg-gray-800"),
				core.Div(core.Class("max-w-4xl mx-auto text-center"),
					core.H2(core.Class("text-3xl font-bold text-gray-900 dark:text-white mb-4"),
						core.Text("Ready to Get Started?"),
					),
					core.P(core.Class("text-lg text-gray-600 dark:text-gray-400 mb-8"),
						core.Text("Join thousands of teams building with us today."),
					),
					ui.Button(ui.ButtonProps{
						Variant: ui.ButtonPrimary,
						Size:    ui.ButtonLG,
						OnClick: func() { r.Navigate("/pricing") },
						Children: []core.Node{
							core.Text("Start Your Free Trial"),
						},
					}),
				),
			),
		)
	}
}

// featureCard renders a feature card with icon, title, and description.
func featureCard(icon, title, description string) core.Node {
	return ui.Card(ui.CardProps{
		Children: []core.Node{
			ui.CardContent(ui.CardContentProps{
				Children: []core.Node{
					core.Div(core.Class("text-3xl mb-4"),
						core.Text(icon),
					),
					core.H3(core.Class("font-semibold text-gray-900 dark:text-white mb-2"),
						core.Text(title),
					),
					core.P(core.Class("text-sm text-gray-600 dark:text-gray-400"),
						core.Text(description),
					),
				},
			}),
		},
	})
}

// testimonialCard renders a testimonial with quote, author name, and title.
func testimonialCard(quote, author, title string) core.Node {
	return ui.Card(ui.CardProps{
		Children: []core.Node{
			ui.CardContent(ui.CardContentProps{
				Children: []core.Node{
					core.P(core.Class("italic text-gray-600 dark:text-gray-400 mb-4"),
						core.Text("\""+quote+"\""),
					),
					core.Div(core.Attrs{},
						core.P(core.Class("font-semibold text-gray-900 dark:text-white"),
							core.Text(author),
						),
						core.P(core.Class("text-sm text-gray-500 dark:text-gray-500"),
							core.Text(title),
						),
					),
				},
			}),
		},
	})
}
