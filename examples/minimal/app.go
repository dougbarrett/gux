package main

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/pages"
)

func main() {
	app := core.New()

	app.SetTitle("Gux Counter")

	app.Routes().
		Hybrid("/", pages.Home).
		Hybrid("/about", pages.About)

	app.Run(":8081")
}
