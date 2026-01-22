package main

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/pages"
)

//go:embed dist/app.wasm
var wasmBinary []byte

//go:embed dist/wasm_exec.js
var wasmExecJS []byte

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/app.wasm", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/wasm")
		w.Write(wasmBinary)
	})

	mux.HandleFunc("/wasm_exec.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(wasmExecJS)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// SSR: router with no rerender (static)
		router := pages.NewRouter(nil)
		component := router.Home()
		html := component().Render(core.HTML()).HTML()

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Gux</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body>
    <div id="app">%s</div>
    <script src="/wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("/app.wasm"), go.importObject)
            .then(result => go.run(result.instance));
    </script>
</body>
</html>`, html)
	})

	fmt.Println("http://localhost:8081")
	http.ListenAndServe(":8081", mux)
}
