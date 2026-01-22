//go:build js && wasm

package main

import "github.com/dougbarrett/gux/examples/minimal/pages"

func main() {
	pages.Mount("app")
	select {}
}
