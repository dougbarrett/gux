package main

import (
	_ "embed"
	"fmt"
	"net/http"

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

	// Register page routes using generated code
	pages.RegisterRoutes(mux, func(content string) string {
		return fmt.Sprintf(`<!DOCTYPE html>
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
</html>`, content)
	})

	fmt.Println("http://localhost:8081")
	http.ListenAndServe(":8081", mux)
}
