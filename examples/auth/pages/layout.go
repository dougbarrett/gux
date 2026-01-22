package pages

import "github.com/dougbarrett/gux/core"

// AuthLayout provides centered card layout for auth pages.
func AuthLayout(children ...core.Node) core.Node {
	return core.Div(core.Class("min-h-screen bg-gray-100 dark:bg-gray-900 flex items-center justify-center py-12 px-4"),
		core.Div(core.Class("w-full max-w-md"),
			children...,
		),
	)
}

// Nav provides simple navigation for the auth example.
func Nav() core.Node {
	return core.Nav(core.Class("bg-white dark:bg-gray-800 shadow"),
		core.Div(core.Class("max-w-7xl mx-auto px-4 py-3 flex justify-between items-center"),
			core.A(
				core.Attrs{Href: "/", Class: "text-xl font-bold text-gray-900 dark:text-white"},
				core.Text("Auth Example"),
			),
			core.Div(core.Class("flex gap-4"),
				core.A(
					core.Attrs{Href: "/login", Class: "text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"},
					core.Text("Login"),
				),
				core.A(
					core.Attrs{Href: "/register", Class: "text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white"},
					core.Text("Register"),
				),
			),
		),
	)
}

// PageLayout wraps content with nav for non-auth pages.
func PageLayout(children ...core.Node) core.Node {
	return core.Div(core.Class("min-h-screen bg-gray-100 dark:bg-gray-900"),
		Nav(),
		core.Main(core.Class("max-w-7xl mx-auto px-4 py-8"),
			children...,
		),
	)
}

