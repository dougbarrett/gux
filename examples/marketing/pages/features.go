package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Features renders the features page with detailed feature sections.
func Features(r *core.Router) func() core.Node {
	return func() core.Node {
		return MarketingLayout(r,
			// Header Section
			Section("Powerful Features for Modern Teams", "Everything you need to build, ship, and scale your products."),

			// Feature Detail Sections
			featureDetailSection(
				"Lightning Fast Performance",
				"Deploy to the edge in seconds. Our global CDN ensures your users get blazing fast load times no matter where they are. With automatic optimization and smart caching, your app stays responsive under any load.",
				[]string{
					"Global edge deployment",
					"Automatic image optimization",
					"Smart caching strategies",
					"Sub-100ms response times",
				},
				true, // visual on right
				"bg-blue-100 dark:bg-blue-900/30",
				"Performance Metrics",
			),
			featureDetailSection(
				"Enterprise-Grade Security",
				"Security isn't an afterthought - it's built into every layer. From automatic SSL to advanced threat protection, your data stays safe without any extra configuration.",
				[]string{
					"Automatic SSL certificates",
					"DDoS protection included",
					"SOC 2 Type II compliant",
					"Role-based access control",
				},
				false, // visual on left
				"bg-green-100 dark:bg-green-900/30",
				"Security Dashboard",
			),
			featureDetailSection(
				"Built for Developers",
				"We obsess over developer experience. Clean APIs, comprehensive docs, and tools that get out of your way so you can focus on building great products.",
				[]string{
					"RESTful and GraphQL APIs",
					"SDKs for all major languages",
					"Comprehensive documentation",
					"Active community and support",
				},
				true, // visual on right
				"bg-purple-100 dark:bg-purple-900/30",
				"Developer Tools",
			),

			// Bottom CTA Section
			core.Section(core.Class("py-20 px-4 bg-gray-100 dark:bg-gray-800"),
				core.Div(core.Class("max-w-4xl mx-auto text-center"),
					core.H2(core.Class("text-3xl font-bold text-gray-900 dark:text-white mb-4"),
						core.Text("Ready to see it in action?"),
					),
					ui.Button(ui.ButtonProps{
						Variant: ui.ButtonPrimary,
						Size:    ui.ButtonLG,
						OnClick: func() { r.Navigate("/pricing") },
						Children: []core.Node{
							core.Text("Start Free Trial"),
						},
					}),
				),
			),
		)
	}
}

// featureDetailSection renders a detailed feature section with text and visual placeholder.
func featureDetailSection(title, description string, bullets []string, visualRight bool, visualBg, visualText string) core.Node {
	// Build bullet list
	bulletNodes := make([]core.Node, len(bullets))
	for i, bullet := range bullets {
		bulletNodes[i] = core.Li(core.Class("flex items-center gap-2"),
			core.Span(core.Class("text-green-500"), core.Text("\u2713")),
			core.Text(bullet),
		)
	}

	// Text content
	textContent := core.Div(core.Class("flex flex-col justify-center"),
		core.H3(core.Class("text-2xl font-bold text-gray-900 dark:text-white mb-4"),
			core.Text(title),
		),
		core.P(core.Class("text-gray-600 dark:text-gray-400 mb-6"),
			core.Text(description),
		),
		core.El("ul", core.Class("space-y-3 text-gray-600 dark:text-gray-400"),
			bulletNodes...,
		),
	)

	// Visual placeholder
	visualContent := core.Div(core.Class("flex items-center justify-center rounded-lg "+visualBg+" h-64 md:h-auto"),
		core.Span(core.Class("text-gray-500 dark:text-gray-400 font-medium"),
			core.Text(visualText),
		),
	)

	// Arrange based on visual position
	var children []core.Node
	if visualRight {
		children = []core.Node{textContent, visualContent}
	} else {
		children = []core.Node{visualContent, textContent}
	}

	return core.Section(core.Class("py-16 px-4"),
		core.Div(core.Class("max-w-6xl mx-auto"),
			ui.Grid(ui.GridProps{
				Cols:  "2",
				Gap:   "12",
				Class: "md:grid-cols-2 grid-cols-1 items-center",
				Children: children,
			}),
		),
	)
}
