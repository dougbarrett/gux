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

	// Serve WASM binary
	mux.HandleFunc("/app.wasm", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/wasm")
		w.Write(wasmBinary)
	})

	// Serve wasm_exec.js
	mux.HandleFunc("/wasm_exec.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(wasmExecJS)
	})

	// Home page - SSR with WASM hydration
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// SSR the initial page (count=0, no handler for SSR)
		html := pages.Home(0, nil).Render(core.HTML()).HTML()

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Gux Counter</title>
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

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
