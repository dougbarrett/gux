// Server entry point - renders pages as HTML.
package main

import (
	"fmt"
	"net/http"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/pages"
)

func main() {
	http.HandleFunc("/", handleHome)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	// Render the page to HTML
	renderer := core.HTML()
	html := pages.Home().Render(renderer).HTML()

	// Wrap in full document
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Gux Minimal Example</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 min-h-screen">
    %s
</body>
</html>`, html)
}
