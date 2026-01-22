// Package pages contains shared page components.
// These compile for both server (SSR) and WASM targets.
package pages

import (
	. "github.com/dougbarrett/gux/core"
)

// Home renders the home page.
// This exact code runs on both server and client.
func Home() Node {
	return Div(Class("container mx-auto p-8"),
		H1(Class("text-3xl font-bold mb-4"),
			Text("Welcome to Gux"),
		),
		P(Class("text-gray-600 mb-8"),
			Text("A universal Go UI framework. This page can be rendered server-side or client-side."),
		),
		Div(Class("grid grid-cols-2 gap-4"),
			Card("Server Rendering", "Fast initial load, SEO friendly, works without JavaScript."),
			Card("WASM Rendering", "Full interactivity, client-side navigation, rich applications."),
		),
	)
}

// Card is a reusable component.
func Card(title, description string) Node {
	return Div(Class("bg-white rounded-lg shadow p-6"),
		H2(Class("text-xl font-semibold mb-2"),
			Text(title),
		),
		P(Class("text-gray-500"),
			Text(description),
		),
	)
}

// Counter demonstrates interactive components.
// The OnClick handler only works in WASM mode.
func Counter(count int, onIncrement func()) Node {
	return Div(Class("flex items-center gap-4"),
		Span(Class("text-2xl font-mono"),
			Text(string(rune('0'+count))), // Simple for demo
		),
		Button(
			Attrs{
				Class:   "px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600",
				OnClick: onIncrement,
			},
			Text("Increment"),
		),
	)
}
