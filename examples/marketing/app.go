package main

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/marketing/pages"
)

func main() {
	app := core.New()
	app.SetTitle("Marketing Example")

	// Marketing routes (single bundle)
	app.Routes().
		Hybrid("/", pages.Home).
		Hybrid("/features", pages.Features).
		Hybrid("/pricing", pages.Pricing).
		Hybrid("/about", pages.About).
		Hybrid("/contact", pages.Contact).
		Hybrid("/video", pages.VideoDemo)

	app.Run(":8083")
}
