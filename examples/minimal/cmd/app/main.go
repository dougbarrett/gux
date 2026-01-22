package main

import (
	_ "embed"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/pages"
)

//go:embed dist/app.wasm
var wasmBinary []byte

//go:embed dist/wasm_exec.js
var wasmExecJS []byte

func main() {
	core.New().
		SetAssets(wasmBinary, wasmExecJS).
		SetTitle("Gux Counter").
		Hybrid("/", pages.Home).
		Run(":8081")
}
