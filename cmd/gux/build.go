package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

// getModulePath uses go list to get the module path for the current directory
func getModulePath() (string, error) {
	cmd := exec.Command("go", "list", "-m")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get module path: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// getCurrentPackagePath gets the import path for the current directory
func getCurrentPackagePath() (string, error) {
	cmd := exec.Command("go", "list", ".")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get package path: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// findMainAppFile finds the main app file (app.go or main.go with core.New())
func findMainAppFile() (string, error) {
	// Check app.go first
	if _, err := os.Stat("app.go"); err == nil {
		return "app.go", nil
	}
	// Check main.go
	if _, err := os.Stat("main.go"); err == nil {
		return "main.go", nil
	}
	return "", fmt.Errorf("no app.go or main.go found in current directory")
}

// PageRoute represents a route with its page handler
type PageRoute struct {
	Path     string
	Handler  string // e.g., "pages.Home"
	IsHybrid bool
}

// parseRoutes parses the main app file to find Hybrid routes
func parseRoutes(filename string) ([]PageRoute, string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, "", fmt.Errorf("parse %s: %w", filename, err)
	}

	var routes []PageRoute
	var pagesImport string

	// Find pages import
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if strings.HasSuffix(path, "/pages") || strings.Contains(path, "/pages") {
			pagesImport = path
			break
		}
	}

	// Find Hybrid() calls - simplified parsing
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Look for .Hybrid( calls
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Hybrid" {
			return true
		}

		if len(call.Args) >= 2 {
			// Get path (first arg)
			if lit, ok := call.Args[0].(*ast.BasicLit); ok {
				path := strings.Trim(lit.Value, `"`)

				// Get handler (second arg)
				handler := ""
				switch h := call.Args[1].(type) {
				case *ast.SelectorExpr:
					if ident, ok := h.X.(*ast.Ident); ok {
						handler = ident.Name + "." + h.Sel.Name
					}
				case *ast.Ident:
					handler = h.Name
				}

				if handler != "" {
					routes = append(routes, PageRoute{
						Path:     path,
						Handler:  handler,
						IsHybrid: true,
					})
				}
			}
		}
		return true
	})

	return routes, pagesImport, nil
}

// generateWasmEntryPoint generates .gux/wasm/main.go
func generateWasmEntryPoint(modulePath, pagesImport string, routes []PageRoute) error {
	if err := os.MkdirAll(".gux/wasm", 0755); err != nil {
		return err
	}

	// Build the router/page matching code
	var routeCode strings.Builder
	if len(routes) == 1 {
		// Single route - simple case
		routeCode.WriteString(fmt.Sprintf("\t\tcomponent := %s(router)\n", routes[0].Handler))
	} else {
		// Multiple routes - add path matching
		routeCode.WriteString("\t\tpath := js.Global().Get(\"location\").Get(\"pathname\").String()\n")
		routeCode.WriteString("\t\tvar component func() core.Node\n")
		routeCode.WriteString("\t\tswitch path {\n")
		for _, route := range routes {
			routeCode.WriteString(fmt.Sprintf("\t\tcase %q:\n", route.Path))
			routeCode.WriteString(fmt.Sprintf("\t\t\tcomponent = %s(router)\n", route.Handler))
		}
		routeCode.WriteString("\t\tdefault:\n")
		// Find the "/" route for default, or use first route
		defaultHandler := ""
		for _, r := range routes {
			if r.Path == "/" {
				defaultHandler = r.Handler
				break
			}
		}
		if defaultHandler == "" && len(routes) > 0 {
			defaultHandler = routes[0].Handler
		}
		if defaultHandler != "" {
			routeCode.WriteString(fmt.Sprintf("\t\t\tcomponent = %s(router)\n", defaultHandler))
		}
		routeCode.WriteString("\t\t}\n")
	}

	code := fmt.Sprintf(`//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/dougbarrett/gux/core"
	"%s"
)

func main() {
	document := js.Global().Get("document")
	window := js.Global()
	container := document.Call("getElementById", "app")

	var router *core.Router
	var render func()

	render = func() {
		container.Set("innerHTML", "")
%s
		node := component()
		result := node.Render(core.DOM())
		if domVal := result.DOMValue(); domVal != nil {
			container.Call("appendChild", domVal.(js.Value))
		}

		// Intercept link clicks for client-side navigation
		links := document.Call("querySelectorAll", "a[href]")
		for i := 0; i < links.Get("length").Int(); i++ {
			link := links.Call("item", i)
			href := link.Get("href").String()
			origin := window.Get("location").Get("origin").String()

			// Only intercept internal links
			if len(href) >= len(origin) && href[:len(origin)] == origin {
				link.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
					args[0].Call("preventDefault")
					path := this.Get("pathname").String()
					router.Navigate(path)
					return nil
				}))
			}
		}
	}

	// Navigate function updates URL and re-renders
	navigate := func(path string) {
		window.Get("history").Call("pushState", nil, "", path)
		render()
	}

	router = core.NewRouter(render)
	router.SetNavigate(navigate)

	// Handle browser back/forward
	window.Call("addEventListener", "popstate", js.FuncOf(func(this js.Value, args []js.Value) any {
		render()
		return nil
	}))

	render()

	select {}
}
`, pagesImport, routeCode.String())

	return os.WriteFile(".gux/wasm/main.go", []byte(code), 0644)
}

// generateAssetsFile generates assets_gen.go
func generateAssetsFile(modulePath string) error {
	code := fmt.Sprintf(`// Code generated by gux; DO NOT EDIT.
package main

import (
	_ "embed"

	"github.com/dougbarrett/gux/core"
)

//go:embed .gux/dist/app.wasm
var wasmBinary []byte

//go:embed .gux/dist/wasm_exec.js
var wasmExecJS []byte

func init() {
	core.SetDefaultAssets(wasmBinary, wasmExecJS)
}
`)
	return os.WriteFile("assets_gen.go", []byte(code), 0644)
}

// buildWasmNew builds the WASM module using TinyGo
func buildWasmNew(tinygo bool) error {
	if err := os.MkdirAll(".gux/dist", 0755); err != nil {
		return err
	}

	fmt.Println("Building WASM...")

	var cmd *exec.Cmd
	if tinygo {
		cmd = exec.Command("tinygo", "build", "-o", ".gux/dist/app.wasm", "-target", "wasm", "./.gux/wasm")
	} else {
		cmd = exec.Command("go", "build", "-o", ".gux/dist/app.wasm", "./.gux/wasm")
		cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("WASM build failed: %w", err)
	}

	// Get size for display
	info, err := os.Stat(".gux/dist/app.wasm")
	if err != nil {
		return err
	}
	sizeMB := float64(info.Size()) / 1024 / 1024
	compiler := "TinyGo"
	if !tinygo {
		compiler = "Go"
	}
	fmt.Printf("Built .gux/dist/app.wasm (%.2f MB) with %s\n", sizeMB, compiler)

	return nil
}

// copyWasmExec copies wasm_exec.js from TinyGo/Go installation
func copyWasmExec(tinygo bool) error {
	// Ensure dist directory exists
	if err := os.MkdirAll(".gux/dist", 0755); err != nil {
		return err
	}

	var src string
	if tinygo {
		cmd := exec.Command("tinygo", "env", "TINYGOROOT")
		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("TinyGo not found: %w", err)
		}
		src = filepath.Join(strings.TrimSpace(string(out)), "targets", "wasm_exec.js")
	} else {
		cmd := exec.Command("go", "env", "GOROOT")
		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("Go not found: %w", err)
		}
		src = filepath.Join(strings.TrimSpace(string(out)), "lib", "wasm", "wasm_exec.js")
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read wasm_exec.js: %w", err)
	}

	return os.WriteFile(".gux/dist/wasm_exec.js", data, 0644)
}

// buildBinary builds the final server binary
func buildBinary() error {
	fmt.Println("Building binary...")

	cmd := exec.Command("go", "build", "-o", "bin/app", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("binary build failed: %w", err)
	}

	info, err := os.Stat("bin/app")
	if err != nil {
		return err
	}
	sizeMB := float64(info.Size()) / 1024 / 1024
	fmt.Printf("Built bin/app (%.2f MB)\n", sizeMB)

	return nil
}

// runBuildNew is the new build command for the simplified architecture
func runBuildNew(tinygo bool) {
	// Get module path
	modulePath, err := getModulePath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Find main app file
	appFile, err := findMainAppFile()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Parse routes from app file
	routes, pagesImport, err := parseRoutes(appFile)
	if err != nil {
		fmt.Printf("Error parsing routes: %v\n", err)
		os.Exit(1)
	}

	if len(routes) == 0 {
		fmt.Println("Warning: no Hybrid routes found")
	}

	// If no pages import found, construct it from current package path
	if pagesImport == "" {
		pkgPath, err := getCurrentPackagePath()
		if err != nil {
			fmt.Printf("Error getting package path: %v\n", err)
			os.Exit(1)
		}
		pagesImport = pkgPath + "/pages"
	}

	// Generate WASM entry point
	fmt.Println("Generating WASM entry point...")
	if err := generateWasmEntryPoint(modulePath, pagesImport, routes); err != nil {
		fmt.Printf("Error generating WASM entry: %v\n", err)
		os.Exit(1)
	}

	// Copy wasm_exec.js
	if err := copyWasmExec(tinygo); err != nil {
		fmt.Printf("Error copying wasm_exec.js: %v\n", err)
		os.Exit(1)
	}

	// Build WASM
	if err := buildWasmNew(tinygo); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Generate assets_gen.go
	fmt.Println("Generating assets...")
	if err := generateAssetsFile(modulePath); err != nil {
		fmt.Printf("Error generating assets: %v\n", err)
		os.Exit(1)
	}

	// Create bin directory
	if err := os.MkdirAll("bin", 0755); err != nil {
		fmt.Printf("Error creating bin/: %v\n", err)
		os.Exit(1)
	}

	// Build final binary
	if err := buildBinary(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nBuild complete! Run with: ./bin/app")
}

// runClean removes all generated files
func runClean() {
	removed := []string{}

	if err := os.RemoveAll("bin"); err == nil {
		if _, err := os.Stat("bin"); os.IsNotExist(err) {
			removed = append(removed, "bin/")
		}
	} else {
		removed = append(removed, "bin/")
	}

	if err := os.RemoveAll(".gux"); err == nil {
		removed = append(removed, ".gux/")
	}

	if err := os.Remove("assets_gen.go"); err == nil {
		removed = append(removed, "assets_gen.go")
	}

	if len(removed) > 0 {
		fmt.Printf("Cleaned: %s\n", strings.Join(removed, ", "))
	} else {
		fmt.Println("Nothing to clean")
	}
}

// runDevNew builds, runs, and cleans up on exit
func runDevNew(tinygo bool) {
	// Build first
	runBuildNew(tinygo)

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("\nStarting dev server...\n")

	// Run the binary
	cmd := exec.Command("./bin/app")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start: %v\n", err)
		runClean()
		os.Exit(1)
	}

	// Wait for signal or process exit
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-sigChan:
		fmt.Println("\nShutting down...")
		cmd.Process.Signal(os.Interrupt)
		<-done
	case err := <-done:
		if err != nil {
			fmt.Printf("Server exited with error: %v\n", err)
		}
	}

	// Clean up
	fmt.Println("Cleaning up...")
	runClean()
}
