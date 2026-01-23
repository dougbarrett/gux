package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// MarketingLayout provides full-page layout with responsive nav and footer.
func MarketingLayout(r *core.Router, children ...core.Node) core.Node {
	return core.Div(core.Class("min-h-screen bg-white dark:bg-gray-900 flex flex-col"),
		Nav(r),
		core.Main(core.Class("flex-1"),
			children...,
		),
		Footer(),
	)
}

// Nav provides responsive navigation with mobile menu toggle.
func Nav(r *core.Router) core.Node {
	menuOpen := r.StateBool("menuOpen", false)

	toggleMenu := func() {
		menuOpen.Set(!menuOpen.Get())
	}

	return core.Nav(core.Class("bg-white dark:bg-gray-800 shadow-sm sticky top-0 z-50"),
		core.Div(core.Class("max-w-7xl mx-auto px-4 py-4 flex items-center justify-between"),
			// Logo/brand
			core.A(core.Attrs{Href: "/", Class: "text-xl font-bold text-gray-900 dark:text-white"},
				core.Text("MarketCo"),
			),
			// Desktop menu (hidden on mobile)
			core.Div(core.Class("hidden md:flex gap-6"),
				navLinks()...,
			),
			// Mobile hamburger (hidden on desktop)
			core.Button(core.Attrs{
				Class:   "md:hidden p-2 text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white",
				OnClick: toggleMenu,
			}, core.Text("\u2630")), // Unicode hamburger
		),
		// Mobile menu (conditional)
		core.If(menuOpen.Get(),
			core.Div(core.Class("md:hidden bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700"),
				core.Div(core.Class("px-4 py-2 flex flex-col gap-2"),
					navLinks()...,
				),
			),
		),
	)
}

// navLinks returns the navigation link nodes.
func navLinks() []core.Node {
	linkClass := "text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"
	return []core.Node{
		core.A(core.Attrs{Href: "/", Class: linkClass}, core.Text("Home")),
		core.A(core.Attrs{Href: "/features", Class: linkClass}, core.Text("Features")),
		core.A(core.Attrs{Href: "/pricing", Class: linkClass}, core.Text("Pricing")),
		core.A(core.Attrs{Href: "/about", Class: linkClass}, core.Text("About")),
		core.A(core.Attrs{Href: "/contact", Class: linkClass}, core.Text("Contact")),
	}
}

// footerLink represents a link in the footer.
type footerLink struct {
	Text string
	Href string
}

// footerColumn renders a footer column with title and links.
func footerColumn(title string, links []footerLink) core.Node {
	linkNodes := make([]core.Node, len(links))
	for i, link := range links {
		linkNodes[i] = core.A(core.Attrs{
			Href:  link.Href,
			Class: "block text-gray-400 hover:text-white transition-colors",
		}, core.Text(link.Text))
	}

	return core.Div(core.Attrs{},
		core.El("h4", core.Class("text-white font-semibold mb-4"),
			core.Text(title),
		),
		core.Div(core.Class("flex flex-col gap-2"),
			linkNodes...,
		),
	)
}

// Footer provides multi-column footer with links and copyright.
func Footer() core.Node {
	return core.Footer(core.Class("bg-gray-900 text-gray-400 py-12"),
		core.Div(core.Class("max-w-6xl mx-auto px-4"),
			ui.Grid(ui.GridProps{
				Cols:  "4",
				Gap:   "8",
				Class: "md:grid-cols-4 grid-cols-2",
				Children: []core.Node{
					footerColumn("Product", []footerLink{
						{Text: "Features", Href: "/features"},
						{Text: "Pricing", Href: "/pricing"},
						{Text: "Docs", Href: "#"},
					}),
					footerColumn("Company", []footerLink{
						{Text: "About", Href: "/about"},
						{Text: "Blog", Href: "#"},
						{Text: "Careers", Href: "#"},
					}),
					footerColumn("Support", []footerLink{
						{Text: "Contact", Href: "/contact"},
						{Text: "FAQ", Href: "#"},
						{Text: "Status", Href: "#"},
					}),
					footerColumn("Legal", []footerLink{
						{Text: "Privacy", Href: "#"},
						{Text: "Terms", Href: "#"},
					}),
				},
			}),
			core.Div(core.Class("border-t border-gray-800 mt-8 pt-8 text-center text-sm"),
				core.Text("\u00a9 2026 MarketCo. All rights reserved."),
			),
		),
	)
}

// Hero renders a full-width hero section with title, subtitle, and CTA buttons.
func Hero(title, subtitle string, ctaNodes ...core.Node) core.Node {
	return core.Section(core.Class("py-20 px-4 bg-gradient-to-br from-blue-600 to-purple-700 text-white"),
		core.Div(core.Class("max-w-4xl mx-auto text-center"),
			core.H1(core.Class("text-4xl md:text-5xl lg:text-6xl font-bold mb-6"),
				core.Text(title),
			),
			core.P(core.Class("text-xl md:text-2xl text-blue-100 mb-8 max-w-2xl mx-auto"),
				core.Text(subtitle),
			),
			core.Div(core.Class("flex flex-col sm:flex-row gap-4 justify-center"),
				ctaNodes...,
			),
		),
	)
}

// Section renders a generic content section with optional header.
func Section(title, description string, children ...core.Node) core.Node {
	var header core.Node
	if title != "" {
		header = core.Div(core.Class("text-center mb-12"),
			core.H2(core.Class("text-3xl font-bold text-gray-900 dark:text-white mb-4"),
				core.Text(title),
			),
			core.P(core.Class("text-lg text-gray-600 dark:text-gray-400 max-w-2xl mx-auto"),
				core.Text(description),
			),
		)
	}

	return core.Section(core.Class("py-20 px-4"),
		core.Div(core.Class("max-w-6xl mx-auto"),
			header,
			core.Frag(children...),
		),
	)
}
