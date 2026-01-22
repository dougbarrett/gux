package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// PageInfo holds parsed information about a page
type PageInfo struct {
	Name       string   // Function name (e.g., "Home")
	Route      string   // Route path (e.g., "/")
	Package    string   // Package name
	LoaderCode string   // Code before return func()
	LoaderVars []string // Variables defined in loader
	Component  string   // The component function body
	StateVars  []string // Variables that are reassigned (reactive state)
	Imports    []string // Required imports
}

func runPages(pagesDir string) {
	info, err := os.Stat(pagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Error: directory '%s' does not exist\n", pagesDir)
			os.Exit(1)
		}
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Printf("Error: '%s' is not a directory\n", pagesDir)
		os.Exit(1)
	}

	pages, err := parsePages(pagesDir)
	if err != nil {
		fmt.Printf("Error parsing pages: %v\n", err)
		os.Exit(1)
	}

	if len(pages) == 0 {
		fmt.Printf("No pages found in '%s'\n", pagesDir)
		fmt.Println("Pages should have //gux:page annotation.")
		return
	}

	fmt.Printf("Found %d page(s):\n", len(pages))
	for _, p := range pages {
		fmt.Printf("  %s -> %s\n", p.Name, p.Route)
	}

	// Generate router.go
	if err := generateRouter(pagesDir, pages); err != nil {
		fmt.Printf("Error generating router: %v\n", err)
		os.Exit(1)
	}

	// Generate SSR handlers
	if err := generateSSR(pagesDir, pages); err != nil {
		fmt.Printf("Error generating SSR: %v\n", err)
		os.Exit(1)
	}

	// Generate WASM entry
	if err := generateWASM(pagesDir, pages); err != nil {
		fmt.Printf("Error generating WASM: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nGenerated:")
	fmt.Println("  pages/router_gen.go")
	fmt.Println("  pages/ssr_gen.go")
	fmt.Println("  pages/wasm_gen.go")
}

func parsePages(dir string) ([]PageInfo, error) {
	var pages []PageInfo

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(entry.Name(), "_gen.go") {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(fset, fullPath, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Doc == nil {
				continue
			}

			// Look for //gux:page annotation
			var route string
			for _, comment := range fn.Doc.List {
				if strings.HasPrefix(comment.Text, "//gux:page ") {
					route = strings.TrimPrefix(comment.Text, "//gux:page ")
					route = strings.TrimSpace(route)
					break
				}
			}

			if route == "" {
				continue
			}

			// Must be a method on *Router
			if fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}

			page := PageInfo{
				Name:    fn.Name.Name,
				Route:   route,
				Package: file.Name.Name,
			}

			// Parse the function body
			if fn.Body != nil {
				parseFunctionBody(fset, fn.Body, &page)
			}

			pages = append(pages, page)
		}
	}

	return pages, nil
}

func parseFunctionBody(fset *token.FileSet, body *ast.BlockStmt, page *PageInfo) {
	var loaderStmts []ast.Stmt
	var componentFunc *ast.FuncLit

	for _, stmt := range body.List {
		// Look for return statement with func() Node
		if ret, ok := stmt.(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
			if fn, ok := ret.Results[0].(*ast.FuncLit); ok {
				componentFunc = fn
				break
			}
		}
		loaderStmts = append(loaderStmts, stmt)
	}

	// Extract loader code
	if len(loaderStmts) > 0 {
		var buf bytes.Buffer
		for _, stmt := range loaderStmts {
			printer.Fprint(&buf, fset, stmt)
			buf.WriteString("\n")

			// Track variable assignments
			if assign, ok := stmt.(*ast.AssignStmt); ok {
				for _, lhs := range assign.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok {
						page.LoaderVars = append(page.LoaderVars, ident.Name)
					}
				}
			}
		}
		page.LoaderCode = buf.String()
	}

	// Extract component function
	if componentFunc != nil && componentFunc.Body != nil {
		var buf bytes.Buffer
		printer.Fprint(&buf, fset, componentFunc.Body)
		page.Component = buf.String()

		// Find state variables (variables that are reassigned)
		findStateVars(componentFunc.Body, page)
	}
}

func findStateVars(body *ast.BlockStmt, page *PageInfo) {
	// Track declared variables and those that get reassigned
	declared := make(map[string]bool)
	reassigned := make(map[string]bool)

	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			if node.Tok == token.DEFINE { // :=
				for _, lhs := range node.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok {
						declared[ident.Name] = true
					}
				}
			} else if node.Tok == token.ASSIGN { // =
				for _, lhs := range node.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok {
						reassigned[ident.Name] = true
					}
				}
			}
		case *ast.IncDecStmt: // count++ or count--
			if ident, ok := node.X.(*ast.Ident); ok {
				reassigned[ident.Name] = true
			}
		}
		return true
	})

	// State vars are those declared AND reassigned
	for name := range declared {
		if reassigned[name] {
			page.StateVars = append(page.StateVars, name)
		}
	}
}

var routerTemplate = `// Code generated by gux pages; DO NOT EDIT.
package {{.Package}}

// Router provides page context and state management.
type Router struct {
	state    map[string]any
	rerender func()
}

// NewRouter creates a new router instance.
func NewRouter(rerender func()) *Router {
	return &Router{
		state:    make(map[string]any),
		rerender: rerender,
	}
}

// State represents a reactive state value.
type State[T any] struct {
	key    string
	router *Router
}

// Get returns the current state value.
func (s *State[T]) Get() T {
	if val, ok := s.router.state[s.key]; ok {
		return val.(T)
	}
	var zero T
	return zero
}

// Set updates the state and triggers re-render.
func (s *State[T]) Set(val T) {
	s.router.state[s.key] = val
	if s.router.rerender != nil {
		s.router.rerender()
	}
}

// StateInt creates an integer state.
func (r *Router) StateInt(key string, initial int) *State[int] {
	if _, ok := r.state[key]; !ok {
		r.state[key] = initial
	}
	return &State[int]{key: key, router: r}
}

// StateBool creates a boolean state.
func (r *Router) StateBool(key string, initial bool) *State[bool] {
	if _, ok := r.state[key]; !ok {
		r.state[key] = initial
	}
	return &State[bool]{key: key, router: r}
}

// StateString creates a string state.
func (r *Router) StateString(key string, initial string) *State[string] {
	if _, ok := r.state[key]; !ok {
		r.state[key] = initial
	}
	return &State[string]{key: key, router: r}
}
`

var ssrTemplate = `// Code generated by gux pages; DO NOT EDIT.
//go:build !js

package {{.Package}}

import (
	"fmt"
	"net/http"

	"github.com/dougbarrett/gux/core"
)

// RegisterRoutes registers all page routes with the given mux.
func RegisterRoutes(mux *http.ServeMux, shell func(content string) string) {
	router := NewRouter(nil) // No rerender for SSR
{{range .Pages}}
	mux.HandleFunc("{{.Route}}", func(w http.ResponseWriter, r *http.Request) {
		component := router.{{.Name}}()
		html := component().Render(core.HTML()).HTML()
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, shell(html))
	})
{{end}}
}
`

var wasmTemplate = `// Code generated by gux pages; DO NOT EDIT.
//go:build js && wasm

package {{.Package}}

import (
	"syscall/js"

	"github.com/dougbarrett/gux/core"
)

// Mount initializes the WASM application.
func Mount(containerID string) {
	document := js.Global().Get("document")
	container := document.Call("getElementById", containerID)

	var router *Router
	var render func()

	render = func() {
		container.Set("innerHTML", "")

		// Route based on current path
		path := js.Global().Get("location").Get("pathname").String()
		var component func() core.Node

		switch path {
{{range .Pages}}		case "{{.Route}}":
			component = router.{{.Name}}()
{{end}}		default:
			component = router.{{(index .Pages 0).Name}}()
		}

		node := component()
		result := node.Render(core.DOM())
		if domVal := result.DOMValue(); domVal != nil {
			container.Call("appendChild", domVal.(js.Value))
		}
	}

	router = NewRouter(render)
	render()
}
`

func generateRouter(dir string, pages []PageInfo) error {
	if len(pages) == 0 {
		return nil
	}

	tmpl, err := template.New("router").Parse(routerTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]any{
		"Package": pages[0].Package,
	})
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "router_gen.go"), buf.Bytes(), 0644)
}

func generateSSR(dir string, pages []PageInfo) error {
	if len(pages) == 0 {
		return nil
	}

	tmpl, err := template.New("ssr").Parse(ssrTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]any{
		"Package": pages[0].Package,
		"Pages":   pages,
	})
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "ssr_gen.go"), buf.Bytes(), 0644)
}

func generateWASM(dir string, pages []PageInfo) error {
	if len(pages) == 0 {
		return nil
	}

	tmpl, err := template.New("wasm").Parse(wasmTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]any{
		"Package": pages[0].Package,
		"Pages":   pages,
	})
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "wasm_gen.go"), buf.Bytes(), 0644)
}
