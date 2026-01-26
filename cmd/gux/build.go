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
	Bundle   string // Bundle name (empty = default "app")
}

// BundleInfo represents a WASM bundle with its routes and imports
type BundleInfo struct {
	Name    string      // Bundle name (e.g., "admin")
	Routes  []PageRoute // Routes in this bundle
	Imports []string    // Package imports needed (e.g., "github.com/.../admin")
}

// CRUDModel represents a registered CRUD model
type CRUDModel struct {
	Name       string // e.g., "Counter"
	PluralName string // e.g., "Counters"
	Path       string // e.g., "counters"
	ListDTO    string // e.g., "UserList" - DTO type for list responses
	DetailDTO  string // e.g., "UserDetail" - DTO type for detail responses
	DTOPackage string // e.g., "dto" - package name for DTOs
	ListDTOInfo   *DTOInfo // Parsed DTO info for list responses
	DetailDTOInfo *DTOInfo // Parsed DTO info for detail responses
}

// DTOInfo holds parsed information about a DTO struct
type DTOInfo struct {
	Name      string            // e.g., "UserList"
	ModelName string            // e.g., "User" (extracted from first field's gux tag)
	Fields    []DTOFieldMapping // Field mappings
	Preloads  []string          // Preload directives from tags
}

// DTOFieldMapping represents a single field mapping from DTO to model
type DTOFieldMapping struct {
	DTOField    string // Field name in DTO, e.g., "Email"
	DTOType     string // Field type in DTO, e.g., "string"
	ModelField  string // Field name in model, e.g., "Email" (from gux:"User.Email")
	IsSlice     bool   // Whether this is a slice field (for relationships)
	SliceDTO    string // For slice fields, the element DTO type, e.g., "PostBrief"
	Preload     string // Preload directive if any
	IsNestedDTO bool   // Whether this is a nested DTO type (e.g., UserBrief)
}

// isPrimitiveType checks if a type string represents a primitive Go type
func isPrimitiveType(t string) bool {
	primitives := map[string]bool{
		"string": true, "int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true, "bool": true, "byte": true, "rune": true,
		"time.Time": true, // Common type that should be treated as primitive
	}
	return primitives[t]
}

// splitParamRoute splits a parameterized route into prefix and suffix around the parameter.
// "/admin/users/:id" -> ("/admin/users/", "")
// "/admin/users/:id/posts/new" -> ("/admin/users/", "/posts/new")
func splitParamRoute(route string) (prefix, suffix string) {
	parts := strings.Split(route, "/")
	var prefixParts, suffixParts []string
	foundParam := false

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			foundParam = true
			continue
		}
		if foundParam {
			suffixParts = append(suffixParts, part)
		} else {
			prefixParts = append(prefixParts, part)
		}
	}

	prefix = strings.Join(prefixParts, "/")
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	if len(suffixParts) > 0 {
		suffix = "/" + strings.Join(suffixParts, "/")
	}

	return prefix, suffix
}

// getParamName extracts the parameter name from a parameterized route.
// "/admin/users/:id" -> "id"
// "/admin/users/:userId/posts/:postId" -> "userId" (returns first param only)
func getParamName(route string) string {
	parts := strings.Split(route, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			return strings.TrimPrefix(part, ":")
		}
	}
	return "id" // default fallback
}

// parseDTOFile parses a DTO file and extracts field mappings from gux tags
func parseDTOFile(dtoDir string, dtoName string) (*DTOInfo, error) {
	// Find and parse all .go files in the dto directory
	entries, err := os.ReadDir(dtoDir)
	if err != nil {
		return nil, fmt.Errorf("read dto dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		fset := token.NewFileSet()
		filename := filepath.Join(dtoDir, entry.Name())
		node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		// Look for the target struct
		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != dtoName {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				info := &DTOInfo{
					Name:   dtoName,
					Fields: []DTOFieldMapping{},
				}

				// Parse each field
				for _, field := range structType.Fields.List {
					if len(field.Names) == 0 || field.Tag == nil {
						continue
					}

					fieldName := field.Names[0].Name
					tagValue := field.Tag.Value

					// Parse the gux tag: `gux:"User.Email"`
					guxTag := extractTag(tagValue, "gux")
					if guxTag == "" {
						continue
					}

					// Parse Model.Field format
					parts := strings.SplitN(guxTag, ".", 2)
					if len(parts) != 2 {
						continue
					}

					modelName := parts[0]
					modelField := parts[1]

					// Set the model name from first field
					if info.ModelName == "" {
						info.ModelName = modelName
					}

					mapping := DTOFieldMapping{
						DTOField:   fieldName,
						DTOType:    formatType(field.Type),
						ModelField: modelField,
					}

					// Check for preload tag
					preloadTag := extractTag(tagValue, "preload")
					if preloadTag != "" {
						mapping.Preload = preloadTag
						info.Preloads = append(info.Preloads, preloadTag)
					}

					// Check if it's a slice type (for relationships)
					if arrayType, ok := field.Type.(*ast.ArrayType); ok {
						mapping.IsSlice = true
						mapping.SliceDTO = formatType(arrayType.Elt)
					} else if !isPrimitiveType(mapping.DTOType) {
						// Non-primitive, non-slice type is a nested DTO
						mapping.IsNestedDTO = true
					}

					info.Fields = append(info.Fields, mapping)
				}

				return info, nil
			}
		}
	}

	return nil, fmt.Errorf("DTO %s not found in %s", dtoName, dtoDir)
}

// extractTag extracts a specific tag value from a struct tag string
func extractTag(tagStr, tagName string) string {
	// Remove backticks
	tagStr = strings.Trim(tagStr, "`")

	// Find the tag
	for _, part := range strings.Split(tagStr, " ") {
		if strings.HasPrefix(part, tagName+":") {
			value := strings.TrimPrefix(part, tagName+":")
			return strings.Trim(value, `"`)
		}
	}
	return ""
}

// formatType converts an ast type expression to a string
func formatType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return formatType(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + formatType(t.X)
	case *ast.ArrayType:
		return "[]" + formatType(t.Elt)
	default:
		return "interface{}"
	}
}

// parseCRUDModels parses the main app file to find CRUD model registrations
func parseCRUDModels(filename string) ([]CRUDModel, string, string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, "", "", fmt.Errorf("parse %s: %w", filename, err)
	}

	var models []CRUDModel
	var modelsImport string
	var dtoImport string

	// Find models and dto imports
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if strings.HasSuffix(path, "/models") || strings.Contains(path, "/models") {
			modelsImport = path
		}
		if strings.HasSuffix(path, "/dto") || strings.Contains(path, "/dto") {
			dtoImport = path
		}
	}

	// Find app.CRUD() calls
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Look for .CRUD( calls
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "CRUD" {
			return true
		}

		if len(call.Args) >= 1 {
			var model CRUDModel

			// Get model name from first argument: models.Counter{}
			switch arg := call.Args[0].(type) {
			case *ast.CompositeLit:
				// models.Counter{}
				if sel, ok := arg.Type.(*ast.SelectorExpr); ok {
					model.Name = sel.Sel.Name
				}
			case *ast.SelectorExpr:
				// models.Counter
				model.Name = arg.Sel.Name
			}

			if model.Name == "" {
				return true
			}

			model.PluralName = model.Name + "s"
			model.Path = strings.ToLower(model.Name) + "s"

			// Parse DTO options from remaining arguments
			for i := 1; i < len(call.Args); i++ {
				optCall, ok := call.Args[i].(*ast.CallExpr)
				if !ok {
					continue
				}
				// Look for core.WithListDTO or core.WithDetailDTO
				optSel, ok := optCall.Fun.(*ast.SelectorExpr)
				if !ok {
					continue
				}

				dtoName := ""
				if len(optCall.Args) >= 1 {
					// Get DTO type name: dto.UserList{}
					switch dtoArg := optCall.Args[0].(type) {
					case *ast.CompositeLit:
						if dtoSel, ok := dtoArg.Type.(*ast.SelectorExpr); ok {
							dtoName = dtoSel.Sel.Name
							if pkgIdent, ok := dtoSel.X.(*ast.Ident); ok {
								model.DTOPackage = pkgIdent.Name
							}
						}
					case *ast.SelectorExpr:
						dtoName = dtoArg.Sel.Name
						if pkgIdent, ok := dtoArg.X.(*ast.Ident); ok {
							model.DTOPackage = pkgIdent.Name
						}
					}
				}

				switch optSel.Sel.Name {
				case "WithListDTO":
					model.ListDTO = dtoName
				case "WithDetailDTO":
					model.DetailDTO = dtoName
				case "WithDTO":
					model.ListDTO = dtoName
					model.DetailDTO = dtoName
				}
			}

			models = append(models, model)
		}
		return true
	})

	return models, modelsImport, dtoImport, nil
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

// parseRoutesAndBundles parses the main app file to find all routes and bundles
func parseRoutesAndBundles(filename string) (map[string]*BundleInfo, map[string]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("parse %s: %w", filename, err)
	}

	// Bundle name -> BundleInfo
	bundles := make(map[string]*BundleInfo)
	bundles["app"] = &BundleInfo{Name: "app"} // Default bundle

	// Package alias -> import path (e.g., "pages" -> "github.com/.../pages")
	imports := make(map[string]string)

	// Find all imports
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		var alias string
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			parts := strings.Split(path, "/")
			alias = parts[len(parts)-1]
		}
		imports[alias] = path
	}

	// Helper to extract RouteGroup info from a call expression
	extractRouteGroupInfo := func(call *ast.CallExpr) (prefix string, bundle string, found bool) {
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "RouteGroup" {
			return "", "", false
		}

		// Get prefix (first argument)
		if len(call.Args) >= 1 {
			if lit, ok := call.Args[0].(*ast.BasicLit); ok {
				prefix = strings.Trim(lit.Value, `"`)
			}
		}

		// Check for WithBundle option in remaining arguments
		for i := 1; i < len(call.Args); i++ {
			if callArg, ok := call.Args[i].(*ast.CallExpr); ok {
				if selArg, ok := callArg.Fun.(*ast.SelectorExpr); ok {
					if selArg.Sel.Name == "WithBundle" && len(callArg.Args) >= 1 {
						if lit, ok := callArg.Args[0].(*ast.BasicLit); ok {
							bundle = strings.Trim(lit.Value, `"`)
						}
					}
				}
			}
		}

		return prefix, bundle, true
	}

	// Helper to walk up the call chain to find RouteGroup
	findRouteGroupInChain := func(call *ast.CallExpr) (prefix string, bundle string, found bool) {
		current := call
		for {
			sel, ok := current.Fun.(*ast.SelectorExpr)
			if !ok {
				return "", "", false
			}

			// Check if the receiver is a call expression
			chainedCall, ok := sel.X.(*ast.CallExpr)
			if !ok {
				return "", "", false
			}

			// Check if this call is RouteGroup
			if prefix, bundle, found := extractRouteGroupInfo(chainedCall); found {
				return prefix, bundle, true
			}

			// Check if the chained call is a method call (Hybrid, GET, etc.)
			chainedSel, ok := chainedCall.Fun.(*ast.SelectorExpr)
			if !ok {
				return "", "", false
			}

			// If it's another Hybrid/GET/etc., keep walking up
			if chainedSel.Sel.Name == "Hybrid" || chainedSel.Sel.Name == "GET" || chainedSel.Sel.Name == "POST" {
				current = chainedCall
				continue
			}

			return "", "", false
		}
	}

	// Find all Hybrid calls
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// Only process Hybrid calls
		if sel.Sel.Name != "Hybrid" || len(call.Args) < 2 {
			return true
		}

		// Get path
		path := ""
		if lit, ok := call.Args[0].(*ast.BasicLit); ok {
			path = strings.Trim(lit.Value, `"`)
		}

		// Get handler
		handler := ""
		pkgAlias := ""
		switch h := call.Args[1].(type) {
		case *ast.SelectorExpr:
			if ident, ok := h.X.(*ast.Ident); ok {
				pkgAlias = ident.Name
				handler = ident.Name + "." + h.Sel.Name
			}
		case *ast.Ident:
			handler = h.Name
		}

		if handler == "" {
			return true
		}

		// Determine bundle by walking up the call chain
		bundleName := "app"
		prefix, bundle, found := findRouteGroupInChain(call)
		if found {
			if bundle != "" {
				bundleName = bundle
				// Ensure bundle exists
				if _, exists := bundles[bundleName]; !exists {
					bundles[bundleName] = &BundleInfo{Name: bundleName}
				}
			}
			// Apply prefix to path
			if path == "/" {
				path = prefix
			} else {
				path = prefix + path
			}
		}

		route := PageRoute{
			Path:     path,
			Handler:  handler,
			IsHybrid: true,
			Bundle:   bundleName,
		}

		bundles[bundleName].Routes = append(bundles[bundleName].Routes, route)

		// Track import needed for this bundle
		if pkgAlias != "" {
			if importPath, ok := imports[pkgAlias]; ok {
				found := false
				for _, imp := range bundles[bundleName].Imports {
					if imp == importPath {
						found = true
						break
					}
				}
				if !found {
					bundles[bundleName].Imports = append(bundles[bundleName].Imports, importPath)
				}
			}
		}

		return true
	})

	return bundles, imports, nil
}

// generateWasmEntryPoint generates .gux/wasm/main.go
func generateWasmEntryPoint(modulePath, pagesImport string, routes []PageRoute) error {
	if err := os.MkdirAll(".gux/wasm", 0755); err != nil {
		return err
	}

	// Build the router/page matching code (same logic as bundle version)
	var routeCode strings.Builder
	if len(routes) == 1 && !strings.Contains(routes[0].Path, ":") {
		// Single non-parameterized route - simple case
		routeCode.WriteString(fmt.Sprintf("\t\tcomponent := %s(router)\n", routes[0].Handler))
	} else {
		// Multiple routes or parameterized routes - need pattern matching
		routeCode.WriteString("\t\tpath := js.Global().Get(\"location\").Get(\"pathname\").String()\n")
		routeCode.WriteString("\t\tvar component func() core.Node\n")

		// Separate exact routes from parameterized routes
		var exactRoutes, paramRoutes []PageRoute
		for _, route := range routes {
			if strings.Contains(route.Path, ":") {
				paramRoutes = append(paramRoutes, route)
			} else {
				exactRoutes = append(exactRoutes, route)
			}
		}

		// Generate switch for exact matches first
		if len(exactRoutes) > 0 {
			routeCode.WriteString("\t\tswitch path {\n")
			for _, route := range exactRoutes {
				routeCode.WriteString(fmt.Sprintf("\t\tcase %q:\n", route.Path))
				routeCode.WriteString(fmt.Sprintf("\t\t\tcomponent = %s(router)\n", route.Handler))
			}
			routeCode.WriteString("\t\t}\n")
		}

		// Generate if-else for parameterized routes
		if len(paramRoutes) > 0 {
			routeCode.WriteString("\t\tif component == nil {\n")
			for i, route := range paramRoutes {
				prefix, suffix := splitParamRoute(route.Path)
				paramName := getParamName(route.Path)
				if i == 0 {
					routeCode.WriteString(fmt.Sprintf("\t\t\tif matchRoute(path, %q, %q) {\n", prefix, suffix))
				} else {
					routeCode.WriteString(fmt.Sprintf("\t\t\t} else if matchRoute(path, %q, %q) {\n", prefix, suffix))
				}
				// Extract and set route params before rendering
				routeCode.WriteString(fmt.Sprintf("\t\t\t\trouter.SetRouteParams(map[string]string{%q: extractRouteParam(path, %q, %q)})\n", paramName, prefix, suffix))
				routeCode.WriteString(fmt.Sprintf("\t\t\t\tcomponent = %s(router)\n", route.Handler))
			}
			routeCode.WriteString("\t\t\t}\n")
			routeCode.WriteString("\t\t}\n")
		}

		// Default fallback
		routeCode.WriteString("\t\tif component == nil {\n")
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
	"encoding/json"
	"strings"
	"syscall/js"

	"github.com/dougbarrett/gux/core"
	"%s"
)

// matchRoute checks if a path matches a parameterized route pattern.
func matchRoute(path, prefix, suffix string) bool {
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	rest := path[len(prefix):]
	if suffix == "" {
		return len(rest) > 0 && !strings.Contains(rest, "/")
	}
	if !strings.HasSuffix(rest, suffix) {
		return false
	}
	paramValue := rest[:len(rest)-len(suffix)]
	return len(paramValue) > 0 && !strings.Contains(paramValue, "/")
}

// extractRouteParam extracts the parameter value from a path given prefix and suffix.
func extractRouteParam(path, prefix, suffix string) string {
	rest := path[len(prefix):]
	if suffix == "" {
		return rest
	}
	return rest[:len(rest)-len(suffix)]
}

// fetchLoader fetches page state from loader endpoint
func fetchLoader(path string, callback func(map[string]any)) {
	loaderPath := "/__gux_api/pages" + path
	if path == "/" {
		loaderPath = "/__gux_api/pages/index"
	}

	promise := js.Global().Call("fetch", loaderPath)
	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
		resp := args[0]
		if resp.Get("ok").Bool() {
			resp.Call("json").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
				// Convert JS object to Go map
				state := make(map[string]any)
				jsObj := args[0]
				keys := js.Global().Get("Object").Call("keys", jsObj)
				for i := 0; i < keys.Get("length").Int(); i++ {
					key := keys.Index(i).String()
					val := jsObj.Get(key)
					switch val.Type() {
					case js.TypeNumber:
						state[key] = val.Float()
					case js.TypeBoolean:
						state[key] = val.Bool()
					case js.TypeString:
						state[key] = val.String()
					}
				}
				callback(state)
				return nil
			}))
		} else {
			callback(nil)
		}
		return nil
	}))
}

func main() {
	document := js.Global().Get("document")
	window := js.Global()
	container := document.Call("getElementById", "app")

	var router *core.Router
	var render func()

	render = func() {
		// Save focus state before re-render
		activeElement := document.Get("activeElement")
		var focusName string
		if !activeElement.IsNull() && !activeElement.IsUndefined() {
			focusName = activeElement.Call("getAttribute", "name").String()
		}

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

			// Skip external links (marked with data-gux-external)
			if link.Call("getAttribute", "data-gux-external").String() == "true" {
				continue
			}

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

		// Restore focus after re-render
		if focusName != "" {
			newElement := document.Call("querySelector", "[name=\""+focusName+"\"]")
			if !newElement.IsNull() && !newElement.IsUndefined() {
				newElement.Call("focus")
			}
		}
	}

	// Navigate fetches page data then renders
	navigate := func(path string) {
		window.Get("history").Call("pushState", nil, "", path)
		fetchLoader(path, func(state map[string]any) {
			if state != nil {
				router.Hydrate(state)
			}
			render()
		})
	}

	router = core.NewRouter(render)
	router.SetNavigate(navigate)

	// Hydrate state from SSR
	stateEl := document.Call("getElementById", "__gux_state")
	if !stateEl.IsNull() && !stateEl.IsUndefined() {
		stateJSON := stateEl.Get("textContent").String()
		var state map[string]any
		if err := json.Unmarshal([]byte(stateJSON), &state); err == nil {
			router.Hydrate(state)
		}
	}

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

// generateBundleWasmEntryPoint generates a WASM entry point for a specific bundle
func generateBundleWasmEntryPoint(bundleName string, bundle *BundleInfo) error {
	wasmDir := ".gux/wasm_" + bundleName
	if bundleName == "app" {
		wasmDir = ".gux/wasm"
	}

	if err := os.MkdirAll(wasmDir, 0755); err != nil {
		return err
	}

	routes := bundle.Routes
	imports := bundle.Imports

	// Build the router/page matching code
	var routeCode strings.Builder
	if len(routes) == 1 && !strings.Contains(routes[0].Path, ":") {
		// Single non-parameterized route - simple case
		routeCode.WriteString(fmt.Sprintf("\t\tcomponent := %s(router)\n", routes[0].Handler))
	} else {
		// Multiple routes or parameterized routes - need pattern matching
		routeCode.WriteString("\t\tpath := js.Global().Get(\"location\").Get(\"pathname\").String()\n")
		routeCode.WriteString("\t\tvar component func() core.Node\n")

		// Separate exact routes from parameterized routes
		var exactRoutes, paramRoutes []PageRoute
		for _, route := range routes {
			if strings.Contains(route.Path, ":") {
				paramRoutes = append(paramRoutes, route)
			} else {
				exactRoutes = append(exactRoutes, route)
			}
		}

		// Generate switch for exact matches first
		if len(exactRoutes) > 0 {
			routeCode.WriteString("\t\tswitch path {\n")
			for _, route := range exactRoutes {
				routeCode.WriteString(fmt.Sprintf("\t\tcase %q:\n", route.Path))
				routeCode.WriteString(fmt.Sprintf("\t\t\tcomponent = %s(router)\n", route.Handler))
			}
			routeCode.WriteString("\t\t}\n")
		}

		// Generate if-else for parameterized routes
		if len(paramRoutes) > 0 {
			routeCode.WriteString("\t\tif component == nil {\n")
			for i, route := range paramRoutes {
				// Generate matching code for parameterized route
				prefix, suffix := splitParamRoute(route.Path)
				paramName := getParamName(route.Path)
				if i == 0 {
					routeCode.WriteString(fmt.Sprintf("\t\t\tif matchRoute(path, %q, %q) {\n", prefix, suffix))
				} else {
					routeCode.WriteString(fmt.Sprintf("\t\t\t} else if matchRoute(path, %q, %q) {\n", prefix, suffix))
				}
				// Extract and set route params before rendering
				routeCode.WriteString(fmt.Sprintf("\t\t\t\trouter.SetRouteParams(map[string]string{%q: extractRouteParam(path, %q, %q)})\n", paramName, prefix, suffix))
				routeCode.WriteString(fmt.Sprintf("\t\t\t\tcomponent = %s(router)\n", route.Handler))
			}
			routeCode.WriteString("\t\t\t}\n")
			routeCode.WriteString("\t\t}\n")
		}

		// Default fallback
		routeCode.WriteString("\t\tif component == nil {\n")
		defaultHandler := ""
		for _, r := range routes {
			if r.Path == "/" || strings.HasSuffix(r.Path, "/") {
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

	// Build imports section
	var importSection strings.Builder
	for _, imp := range imports {
		parts := strings.Split(imp, "/")
		alias := parts[len(parts)-1]
		importSection.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
		_ = alias // Alias inferred from path
	}

	code := fmt.Sprintf(`//go:build js && wasm

package main

import (
	"encoding/json"
	"strings"
	"syscall/js"

	"github.com/dougbarrett/gux/core"
%s)

// matchRoute checks if a path matches a parameterized route pattern.
// prefix is the part before the param, suffix is the part after.
// e.g., matchRoute("/admin/users/123", "/admin/users/", "") returns true
// e.g., matchRoute("/admin/users/123/posts/new", "/admin/users/", "/posts/new") returns true
func matchRoute(path, prefix, suffix string) bool {
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	rest := path[len(prefix):]
	if suffix == "" {
		// No suffix - just need something after the prefix (the param value)
		return len(rest) > 0 && !strings.Contains(rest, "/")
	}
	// Has suffix - need to match suffix at the end
	if !strings.HasSuffix(rest, suffix) {
		return false
	}
	// Extract the param value (between prefix and suffix)
	paramValue := rest[:len(rest)-len(suffix)]
	return len(paramValue) > 0 && !strings.Contains(paramValue, "/")
}

// extractRouteParam extracts the parameter value from a path given prefix and suffix.
// e.g., extractRouteParam("/admin/users/123", "/admin/users/", "") returns "123"
// e.g., extractRouteParam("/admin/users/123/posts/new", "/admin/users/", "/posts/new") returns "123"
func extractRouteParam(path, prefix, suffix string) string {
	rest := path[len(prefix):]
	if suffix == "" {
		return rest
	}
	return rest[:len(rest)-len(suffix)]
}

// fetchLoader fetches page state from loader endpoint
func fetchLoader(path string, callback func(map[string]any)) {
	loaderPath := "/__gux_api/pages" + path
	if path == "/" {
		loaderPath = "/__gux_api/pages/index"
	}

	promise := js.Global().Call("fetch", loaderPath)
	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
		resp := args[0]
		if resp.Get("ok").Bool() {
			resp.Call("json").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
				// Convert JS object to Go map
				state := make(map[string]any)
				jsObj := args[0]
				keys := js.Global().Get("Object").Call("keys", jsObj)
				for i := 0; i < keys.Get("length").Int(); i++ {
					key := keys.Index(i).String()
					val := jsObj.Get(key)
					switch val.Type() {
					case js.TypeNumber:
						state[key] = val.Float()
					case js.TypeBoolean:
						state[key] = val.Bool()
					case js.TypeString:
						state[key] = val.String()
					}
				}
				callback(state)
				return nil
			}))
		} else {
			callback(nil)
		}
		return nil
	}))
}

func main() {
	document := js.Global().Get("document")
	window := js.Global()
	container := document.Call("getElementById", "app")

	var router *core.Router
	var render func()

	render = func() {
		// Save focus state before re-render
		activeElement := document.Get("activeElement")
		var focusName string
		if !activeElement.IsNull() && !activeElement.IsUndefined() {
			focusName = activeElement.Call("getAttribute", "name").String()
		}

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

			// Skip external links (marked with data-gux-external)
			if link.Call("getAttribute", "data-gux-external").String() == "true" {
				continue
			}

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

		// Restore focus after re-render
		if focusName != "" {
			newElement := document.Call("querySelector", "[name=\""+focusName+"\"]")
			if !newElement.IsNull() && !newElement.IsUndefined() {
				newElement.Call("focus")
			}
		}
	}

	// Navigate fetches page data then renders
	navigate := func(path string) {
		window.Get("history").Call("pushState", nil, "", path)
		fetchLoader(path, func(state map[string]any) {
			if state != nil {
				router.Hydrate(state)
			}
			render()
		})
	}

	router = core.NewRouter(render)
	router.SetNavigate(navigate)

	// Hydrate state from SSR
	stateEl := document.Call("getElementById", "__gux_state")
	if !stateEl.IsNull() && !stateEl.IsUndefined() {
		stateJSON := stateEl.Get("textContent").String()
		var state map[string]any
		if err := json.Unmarshal([]byte(stateJSON), &state); err == nil {
			router.Hydrate(state)
		}
	}

	// Handle browser back/forward
	window.Call("addEventListener", "popstate", js.FuncOf(func(this js.Value, args []js.Value) any {
		render()
		return nil
	}))

	render()

	select {}
}
`, importSection.String(), routeCode.String())

	return os.WriteFile(wasmDir+"/main.go", []byte(code), 0644)
}

// buildWasmBundle builds a specific WASM bundle
func buildWasmBundle(bundleName string, tinygo bool) error {
	wasmDir := ".gux/wasm_" + bundleName
	outputFile := ".gux/dist/" + bundleName + ".wasm"

	if bundleName == "app" {
		wasmDir = ".gux/wasm"
		outputFile = ".gux/dist/app.wasm"
	}

	if err := os.MkdirAll(".gux/dist", 0755); err != nil {
		return err
	}

	fmt.Printf("Building WASM bundle: %s...\n", bundleName)

	var cmd *exec.Cmd
	if tinygo {
		cmd = exec.Command("tinygo", "build", "-o", outputFile, "-target", "wasm", "./"+wasmDir)
	} else {
		cmd = exec.Command("go", "build", "-o", outputFile, "./"+wasmDir)
		cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("WASM build failed for bundle %s: %w", bundleName, err)
	}

	// Get size for display
	info, err := os.Stat(outputFile)
	if err != nil {
		return err
	}
	sizeMB := float64(info.Size()) / 1024 / 1024
	compiler := "TinyGo"
	if !tinygo {
		compiler = "Go"
	}
	fmt.Printf("Built %s (%.2f MB) with %s\n", outputFile, sizeMB, compiler)

	return nil
}

// generateServerDTOCode generates server-side API code for models with parsed DTO info
func generateServerDTOCode(m CRUDModel) string {
	var sb strings.Builder

	// Determine which DTO to use for list and detail
	listDTO := m.ListDTO
	detailDTO := m.DetailDTO
	if detailDTO == "" {
		detailDTO = listDTO
	}

	// Get the model name from the DTO info
	modelName := m.Name
	if m.ListDTOInfo != nil && m.ListDTOInfo.ModelName != "" {
		modelName = m.ListDTOInfo.ModelName
	}

	// Build preload chain for list
	listPreloads := ""
	if m.ListDTOInfo != nil {
		for _, p := range m.ListDTOInfo.Preloads {
			listPreloads += fmt.Sprintf(".Preload(\"%s\")", p)
		}
	}

	// Build preload chain for detail
	detailPreloads := ""
	if m.DetailDTOInfo != nil {
		for _, p := range m.DetailDTOInfo.Preloads {
			detailPreloads += fmt.Sprintf(".Preload(\"%s\")", p)
		}
	}

	// Generate field mapping for list DTO
	listFieldMapping := generateFieldMapping(m.ListDTOInfo, "item")
	listNestedMapping := generateNestedDTOMapping(m.ListDTOInfo, "result[i]", "item", true)

	// Generate field mapping for detail DTO
	detailFieldMapping := generateFieldMapping(m.DetailDTOInfo, "item")
	detailNestedMapping := generateNestedDTOMapping(m.DetailDTOInfo, "result", "item", false)

	// Check if detail DTO has preloads (relationships) - if so, try FromModel pattern
	hasRelationships := m.DetailDTOInfo != nil && len(m.DetailDTOInfo.Preloads) > 0

	// Generate the Get method body based on whether we have relationships
	var getMethodBody string
	if hasRelationships {
		// Use FromModel pattern for DTOs with relationships (if implemented)
		// Uses type assertion to check if FromModel exists at runtime
		getMethodBody = fmt.Sprintf(`// Get returns a single %s by ID.
func (a *%sAPI) Get(id uint, callback func(*dto.%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	var item models.%s
	if err := db%s.First(&item, id).Error; err != nil {
		callback(nil, err)
		return
	}
	// Try DTOMapper.FromModel if implemented (for complex relationships)
	empty := &dto.%s{}
	if mapper, ok := interface{}(empty).(interface{ FromModel(interface{}) interface{} }); ok {
		mapped := mapper.FromModel(item)
		if result, ok := mapped.(dto.%s); ok {
			callback(&result, nil)
			return
		}
	}
	// Fallback to basic mapping
	result := &dto.%s{
%s	}
%s	callback(result, nil)
}`,
			m.Name, m.PluralName, detailDTO, modelName, detailPreloads,
			detailDTO, detailDTO, detailDTO, detailFieldMapping, detailNestedMapping)
	} else {
		// Use simple field mapping for DTOs without relationships
		getMethodBody = fmt.Sprintf(`// Get returns a single %s by ID.
func (a *%sAPI) Get(id uint, callback func(*dto.%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	var item models.%s
	if err := db%s.First(&item, id).Error; err != nil {
		callback(nil, err)
		return
	}
	result := &dto.%s{
%s	}
%s	callback(result, nil)
}`,
			m.Name, m.PluralName, detailDTO, modelName, detailPreloads, detailDTO, detailFieldMapping, detailNestedMapping)
	}

	sb.WriteString(fmt.Sprintf(`
// %sAPI provides CRUD operations for %s.
type %sAPI struct{}

// %s is the API client for %s operations.
var %s = &%sAPI{}

// List returns all %s records.
func (a *%sAPI) List(callback func([]dto.%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	var items []models.%s
	if err := db%s.Find(&items).Error; err != nil {
		callback(nil, err)
		return
	}
	// Convert to DTOs using field mappings from gux tags
	result := make([]dto.%s, len(items))
	for i, item := range items {
		result[i] = dto.%s{
%s		}
%s	}
	callback(result, nil)
}

%s

// Create creates a new %s (server-side stub - actual creation handled by CRUD endpoint).
func (a *%sAPI) Create(data map[string]interface{}, callback func(*dto.%s, error)) {
	// Server-side: this is typically not called directly
	// The CRUD endpoint handles creation with hooks
	callback(nil, nil)
}

// Update updates an existing %s (server-side stub).
func (a *%sAPI) Update(id uint, data map[string]interface{}, callback func(*dto.%s, error)) {
	// Server-side: this is typically not called directly
	callback(nil, nil)
}

// Delete deletes a %s by ID (server-side stub).
func (a *%sAPI) Delete(id uint, callback func(error)) {
	// Server-side: this is typically not called directly
	callback(nil)
}
`,
		m.PluralName, m.Name,                    // type comment
		m.PluralName,                            // type name
		m.PluralName, m.Name,                    // var comment
		m.PluralName, m.PluralName,              // var declaration
		m.Name,                                  // List comment
		m.PluralName, listDTO,                   // List signature
		modelName,                               // List query model
		listPreloads,                            // List preloads
		listDTO,                                 // List result make
		listDTO,                                 // List DTO type
		listFieldMapping,                        // List field mapping
		listNestedMapping,                       // List nested DTO mapping
		getMethodBody,                           // Get method (generated above)
		m.Name,                                  // Create comment
		m.PluralName, detailDTO,                 // Create signature
		m.Name,                                  // Update comment
		m.PluralName, detailDTO,                 // Update signature
		m.Name,                                  // Delete comment
		m.PluralName,                            // Delete signature
	))

	return sb.String()
}

// generateFieldMapping generates the field assignment code from parsed DTO info
func generateFieldMapping(info *DTOInfo, varName string) string {
	if info == nil {
		return ""
	}

	var sb strings.Builder
	for _, f := range info.Fields {
		if f.IsSlice {
			// For slice fields (relationships), we need to map each element
			sb.WriteString(fmt.Sprintf("\t\t\t// %s mapped from %s.%s (slice)\n",
				f.DTOField, info.ModelName, f.ModelField))
			// Skip slice mapping in simple generation - requires nested loop
			continue
		}
		if f.IsNestedDTO {
			// Skip nested DTOs here - they are handled by generateNestedDTOMapping
			continue
		}
		sb.WriteString(fmt.Sprintf("\t\t\t%s: %s.%s,\n", f.DTOField, varName, f.ModelField))
	}
	return sb.String()
}

// generateNestedDTOMapping generates code to map nested DTO fields after the main struct literal
// Returns empty string if no nested DTOs, otherwise returns the mapping code
func generateNestedDTOMapping(info *DTOInfo, resultVar, itemVar string, forList bool) string {
	if info == nil {
		return ""
	}

	var sb strings.Builder
	for _, f := range info.Fields {
		if !f.IsNestedDTO {
			continue
		}

		// Parse nested DTO type to get field mappings (e.g., UserBrief -> ID, Name)
		nestedInfo, err := parseDTOFile("dto", f.DTOType)
		if err != nil {
			// Fall back to simple assignment if we can't parse nested DTO
			continue
		}

		// Generate null check and nested mapping
		if forList {
			// For list: result[i].Author = dto.UserBrief{...}
			sb.WriteString(fmt.Sprintf("\t\tif %s.%s != nil {\n", itemVar, f.ModelField))
			sb.WriteString(fmt.Sprintf("\t\t\t%s.%s = dto.%s{\n", resultVar, f.DTOField, f.DTOType))
		} else {
			// For single: result.Author = dto.UserBrief{...}
			sb.WriteString(fmt.Sprintf("\tif %s.%s != nil {\n", itemVar, f.ModelField))
			sb.WriteString(fmt.Sprintf("\t\t%s.%s = dto.%s{\n", resultVar, f.DTOField, f.DTOType))
		}

		// Add nested field mappings
		for _, nf := range nestedInfo.Fields {
			if nf.IsSlice || nf.IsNestedDTO {
				continue // Skip complex nested fields for now
			}
			if forList {
				sb.WriteString(fmt.Sprintf("\t\t\t\t%s: %s.%s.%s,\n", nf.DTOField, itemVar, f.ModelField, nf.ModelField))
			} else {
				sb.WriteString(fmt.Sprintf("\t\t\t%s: %s.%s.%s,\n", nf.DTOField, itemVar, f.ModelField, nf.ModelField))
			}
		}

		if forList {
			sb.WriteString("\t\t\t}\n")
			sb.WriteString("\t\t}\n")
		} else {
			sb.WriteString("\t\t}\n")
			sb.WriteString("\t}\n")
		}
	}
	return sb.String()
}

// generateAPIClient generates .gux/api/client.go with WASM-compatible DTOs
func generateAPIClient(modelsImport string, dtoImport string, models []CRUDModel) error {
	if len(models) == 0 {
		return nil // No CRUD models, skip API client generation
	}

	if err := os.MkdirAll(".gux/api", 0755); err != nil {
		return err
	}

	// Generate DTO structs for models WITHOUT custom DTOs
	var dtoCode strings.Builder
	for _, m := range models {
		// Skip models with custom DTOs
		if m.ListDTO != "" || m.DetailDTO != "" {
			continue
		}
		dtoCode.WriteString(fmt.Sprintf(`
// %s is a WASM-compatible DTO for %s.
type %s struct {
	ID          uint   `+"`json:\"ID\"`"+`
	CreatedAt   string `+"`json:\"CreatedAt,omitempty\"`"+`
	UpdatedAt   string `+"`json:\"UpdatedAt,omitempty\"`"+`
	Name        string `+"`json:\"name,omitempty\"`"+`
	Description string `+"`json:\"description,omitempty\"`"+`
}
`, m.Name, m.Name, m.Name))
	}

	// Generate model-specific API code
	var modelCode strings.Builder
	for _, m := range models {
		// Models with DTOs get specialized read-only API
		if m.ListDTO != "" || m.DetailDTO != "" {
			listType := m.ListDTO
			if listType == "" {
				listType = m.Name
			}
			detailType := m.DetailDTO
			if detailType == "" {
				detailType = m.Name
			}
			dtoPkg := m.DTOPackage
			if dtoPkg == "" {
				dtoPkg = "dto"
			}

			modelCode.WriteString(fmt.Sprintf(`
// %sAPI provides read operations for %s with DTO responses.
type %sAPI struct {
	baseURL string
}

// %s is the API client for %s operations.
var %s = &%sAPI{baseURL: "/__gux_api/crud/%s"}

// List returns all %s records using %s DTO.
func (a *%sAPI) List(callback func([]%s.%s, error)) {
	go func() {
		resp, err := fetch("GET", a.baseURL, nil)
		if err != nil {
			callback(nil, err)
			return
		}
		var items []%s.%s
		if err := json.Unmarshal(resp, &items); err != nil {
			callback(nil, err)
			return
		}
		callback(items, nil)
	}()
}

// Get returns a single %s by ID using %s DTO.
func (a *%sAPI) Get(id uint, callback func(*%s.%s, error)) {
	go func() {
		resp, err := fetch("GET", fmt.Sprintf("%%s/%%d", a.baseURL, id), nil)
		if err != nil {
			callback(nil, err)
			return
		}
		var item %s.%s
		if err := json.Unmarshal(resp, &item); err != nil {
			callback(nil, err)
			return
		}
		callback(&item, nil)
	}()
}

// Create creates a new %s with the given data.
func (a *%sAPI) Create(data map[string]interface{}, callback func(*%s.%s, error)) {
	go func() {
		body, _ := json.Marshal(data)
		resp, err := fetch("POST", a.baseURL, body)
		if err != nil {
			callback(nil, err)
			return
		}
		var item %s.%s
		if err := json.Unmarshal(resp, &item); err != nil {
			callback(nil, err)
			return
		}
		callback(&item, nil)
	}()
}

// Update updates an existing %s.
func (a *%sAPI) Update(id uint, data map[string]interface{}, callback func(*%s.%s, error)) {
	go func() {
		body, _ := json.Marshal(data)
		resp, err := fetch("PUT", fmt.Sprintf("%%s/%%d", a.baseURL, id), body)
		if err != nil {
			callback(nil, err)
			return
		}
		var item %s.%s
		if err := json.Unmarshal(resp, &item); err != nil {
			callback(nil, err)
			return
		}
		callback(&item, nil)
	}()
}

// Delete deletes a %s by ID.
func (a *%sAPI) Delete(id uint, callback func(error)) {
	go func() {
		_, err := fetch("DELETE", fmt.Sprintf("%%s/%%d", a.baseURL, id), nil)
		callback(err)
	}()
}
`,
				m.PluralName, m.Name,               // type comment
				m.PluralName,                       // type name
				m.PluralName, m.Name,               // var comment
				m.PluralName, m.PluralName, m.Path, // var declaration
				m.Name, listType,                   // List comment
				m.PluralName, dtoPkg, listType,     // List signature
				dtoPkg, listType,                   // List body
				m.Name, detailType,                 // Get comment
				m.PluralName, dtoPkg, detailType,   // Get signature
				dtoPkg, detailType,                 // Get body
				m.Name,                             // Create comment
				m.PluralName, dtoPkg, detailType,   // Create signature
				dtoPkg, detailType,                 // Create body
				m.Name,                             // Update comment
				m.PluralName, dtoPkg, detailType,   // Update signature
				dtoPkg, detailType,                 // Update body
				m.Name,                             // Delete comment
				m.PluralName,                       // Delete signature
			))
			continue
		}

		// Models without DTOs get full CRUD with inline DTO
		modelCode.WriteString(fmt.Sprintf(`
// %sAPI provides CRUD operations for %s.
type %sAPI struct {
	baseURL string
}

// %s is the API client for %s operations.
var %s = &%sAPI{baseURL: "/__gux_api/crud/%s"}

// List returns all %s records.
func (a *%sAPI) List(callback func([]%s, error)) {
	go func() {
		resp, err := fetch("GET", a.baseURL, nil)
		if err != nil {
			callback(nil, err)
			return
		}
		var items []%s
		if err := json.Unmarshal(resp, &items); err != nil {
			callback(nil, err)
			return
		}
		callback(items, nil)
	}()
}

// Get returns a single %s by ID.
func (a *%sAPI) Get(id uint, callback func(*%s, error)) {
	go func() {
		resp, err := fetch("GET", fmt.Sprintf("%%s/%%d", a.baseURL, id), nil)
		if err != nil {
			callback(nil, err)
			return
		}
		var item %s
		if err := json.Unmarshal(resp, &item); err != nil {
			callback(nil, err)
			return
		}
		callback(&item, nil)
	}()
}

// Create creates a new %s.
func (a *%sAPI) Create(item *%s, callback func(*%s, error)) {
	go func() {
		body, _ := json.Marshal(item)
		resp, err := fetch("POST", a.baseURL, body)
		if err != nil {
			callback(nil, err)
			return
		}
		var result %s
		if err := json.Unmarshal(resp, &result); err != nil {
			callback(nil, err)
			return
		}
		callback(&result, nil)
	}()
}

// Update updates an existing %s.
func (a *%sAPI) Update(item *%s, callback func(*%s, error)) {
	go func() {
		body, _ := json.Marshal(item)
		resp, err := fetch("PUT", fmt.Sprintf("%%s/%%d", a.baseURL, item.ID), body)
		if err != nil {
			callback(nil, err)
			return
		}
		var result %s
		if err := json.Unmarshal(resp, &result); err != nil {
			callback(nil, err)
			return
		}
		callback(&result, nil)
	}()
}

// Delete deletes a %s by ID.
func (a *%sAPI) Delete(id uint, callback func(error)) {
	go func() {
		_, err := fetch("DELETE", fmt.Sprintf("%%s/%%d", a.baseURL, id), nil)
		callback(err)
	}()
}
`,
			m.PluralName, m.Name,               // type comment
			m.PluralName,                       // type name
			m.PluralName, m.Name,               // var comment
			m.PluralName, m.PluralName, m.Path, // var declaration
			m.Name,                             // List comment
			m.PluralName, m.Name,               // List signature
			m.Name,                             // List body
			m.Name,                             // Get comment
			m.PluralName, m.Name,               // Get signature
			m.Name,                             // Get body
			m.Name,                             // Create comment
			m.PluralName, m.Name, m.Name,       // Create signature
			m.Name,                             // Create body
			m.Name,                             // Update comment
			m.PluralName, m.Name, m.Name,       // Update signature
			m.Name,                             // Update body
			m.Name,                             // Delete comment
			m.PluralName,                       // Delete signature
		))
	}

	// Note: modelsImport is unused since we generate our own DTOs
	_ = modelsImport

	// Build imports list
	imports := `"encoding/json"
	"errors"
	"fmt"
	"syscall/js"`

	// Add dto import if any model uses DTOs
	if dtoImport != "" {
		for _, m := range models {
			if m.ListDTO != "" || m.DetailDTO != "" {
				imports += fmt.Sprintf("\n\n\t\"%s\"", dtoImport)
				break
			}
		}
	}

	code := fmt.Sprintf(`//go:build js && wasm

// Code generated by gux; DO NOT EDIT.
package api

import (
	%s
)

// Init is a no-op in WASM (database not available).
func Init(db interface{}) {}

// fetch makes an HTTP request and returns the response body.
func fetch(method, url string, body []byte) ([]byte, error) {
	done := make(chan struct{})
	var result []byte
	var fetchErr error

	// Create JS object for headers
	jsOpts := js.Global().Get("Object").New()
	jsOpts.Set("method", method)
	headers := js.Global().Get("Object").New()
	headers.Set("Content-Type", "application/json")
	jsOpts.Set("headers", headers)
	if body != nil {
		jsOpts.Set("body", string(body))
	}

	// Call fetch
	promise := js.Global().Call("fetch", url, jsOpts)

	// Handle response
	var thenFunc, catchFunc js.Func
	thenFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resp := args[0]
		if !resp.Get("ok").Bool() {
			fetchErr = errors.New("request failed: " + resp.Get("statusText").String())
			close(done)
			return nil
		}
		// Get response text
		resp.Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			result = []byte(args[0].String())
			close(done)
			return nil
		}))
		return nil
	})
	catchFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fetchErr = errors.New("fetch error: " + args[0].Get("message").String())
		close(done)
		return nil
	})

	promise.Call("then", thenFunc).Call("catch", catchFunc)

	<-done
	return result, fetchErr
}
%s%s
`, imports, dtoCode.String(), modelCode.String())

	if err := os.WriteFile(".gux/api/client.go", []byte(code), 0644); err != nil {
		return err
	}

	// Generate server-side implementation (real DB queries)
	var stubDtoCode strings.Builder
	for _, m := range models {
		// Skip models with DTOs - they import from the dto package
		if m.ListDTO != "" || m.DetailDTO != "" {
			continue
		}
		stubDtoCode.WriteString(fmt.Sprintf(`
// %s is the API DTO for %s.
type %s struct {
	ID          uint   `+"`json:\"ID\"`"+`
	CreatedAt   string `+"`json:\"CreatedAt,omitempty\"`"+`
	UpdatedAt   string `+"`json:\"UpdatedAt,omitempty\"`"+`
	Name        string `+"`json:\"name,omitempty\"`"+`
	Description string `+"`json:\"description,omitempty\"`"+`
}
`, m.Name, m.Name, m.Name))
	}

	var stubCode strings.Builder
	for _, m := range models {
		// Models WITH parsed DTO info get field-mapped code
		if m.ListDTOInfo != nil || m.DetailDTOInfo != nil {
			stubCode.WriteString(generateServerDTOCode(m))
			continue
		}

		// Models WITHOUT DTOs get hardcoded Counter-like code
		if m.ListDTO != "" || m.DetailDTO != "" {
			continue // Skip - no parsed info available
		}
		stubCode.WriteString(fmt.Sprintf(`
// %sAPI provides CRUD operations for %s.
type %sAPI struct{}

// %s is the API client for %s operations.
var %s = &%sAPI{}

// List returns all %s records.
func (a *%sAPI) List(callback func([]%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	var items []models.%s
	if err := db.Find(&items).Error; err != nil {
		callback(nil, err)
		return
	}
	// Convert to DTOs
	result := make([]%s, len(items))
	for i, item := range items {
		result[i] = %s{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
		}
	}
	callback(result, nil)
}

// Get returns a single %s by ID.
func (a *%sAPI) Get(id uint, callback func(*%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	var item models.%s
	if err := db.First(&item, id).Error; err != nil {
		callback(nil, err)
		return
	}
	result := &%s{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
	}
	callback(result, nil)
}

// Create creates a new %s.
func (a *%sAPI) Create(item *%s, callback func(*%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	model := models.%s{
		Name:        item.Name,
		Description: item.Description,
	}
	if err := db.Create(&model).Error; err != nil {
		callback(nil, err)
		return
	}
	result := &%s{
		ID:          model.ID,
		Name:        model.Name,
		Description: model.Description,
	}
	callback(result, nil)
}

// Update updates an existing %s.
func (a *%sAPI) Update(item *%s, callback func(*%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	var model models.%s
	if err := db.First(&model, item.ID).Error; err != nil {
		callback(nil, err)
		return
	}
	model.Name = item.Name
	model.Description = item.Description
	if err := db.Save(&model).Error; err != nil {
		callback(nil, err)
		return
	}
	result := &%s{
		ID:          model.ID,
		Name:        model.Name,
		Description: model.Description,
	}
	callback(result, nil)
}

// Delete deletes a %s by ID.
func (a *%sAPI) Delete(id uint, callback func(error)) {
	if db == nil {
		callback(nil)
		return
	}
	callback(db.Delete(&models.%s{}, id).Error)
}
`,
			m.PluralName, m.Name,       // type comment
			m.PluralName,               // type name
			m.PluralName, m.Name,       // var comment
			m.PluralName, m.PluralName, // var declaration
			m.Name,                     // List comment
			m.PluralName, m.Name,       // List signature
			m.Name,                     // List query
			m.Name, m.Name,             // List convert
			m.Name,                     // Get comment
			m.PluralName, m.Name,       // Get signature
			m.Name,                     // Get query
			m.Name,                     // Get result
			m.Name,                     // Create comment
			m.PluralName, m.Name, m.Name, // Create signature
			m.Name,                     // Create model
			m.Name,                     // Create result
			m.Name,                     // Update comment
			m.PluralName, m.Name, m.Name, // Update signature
			m.Name,                     // Update query
			m.Name,                     // Update result
			m.Name,                     // Delete comment
			m.PluralName,               // Delete signature
			m.Name,                     // Delete model
		))
	}

	// Build server-side imports
	serverImports := fmt.Sprintf(`"gorm.io/gorm"

	"%s"`, modelsImport)

	// Add dto import if any models use DTOs with parsed info
	for _, m := range models {
		if m.ListDTOInfo != nil || m.DetailDTOInfo != nil {
			serverImports += fmt.Sprintf(`

	"%s"`, dtoImport)
			break
		}
	}

	stubFile := fmt.Sprintf(`//go:build !js || !wasm

// Code generated by gux; DO NOT EDIT.
// Server-side implementation - queries database directly.
package api

import (
	%s
)

var db *gorm.DB

// Init initializes the API with a database connection.
// Called automatically by the gux framework.
func Init(database *gorm.DB) {
	db = database
}
%s%s
`, serverImports, stubDtoCode.String(), stubCode.String())

	return os.WriteFile(".gux/api/client_server.go", []byte(stubFile), 0644)
}

// generateAssetsFile generates assets_gen.go with support for multiple bundles
func generateAssetsFile(modulePath string, bundles []string) error {
	var embedCode strings.Builder
	var initCode strings.Builder

	// Default bundle (app.wasm)
	embedCode.WriteString(`//go:embed .gux/dist/app.wasm
var wasmBinary []byte

//go:embed .gux/dist/wasm_exec.js
var wasmExecJS []byte

//go:embed .gux/dist/styles.css
var stylesCSS []byte
`)

	// Additional bundles
	for _, bundle := range bundles {
		if bundle != "app" {
			embedCode.WriteString(fmt.Sprintf(`
//go:embed .gux/dist/%s.wasm
var wasm%s []byte
`, bundle, strings.Title(bundle)))
		}
	}

	// Init function
	initCode.WriteString("\tfunc init() {\n")
	initCode.WriteString("\t\tcore.SetDefaultAssets(wasmBinary, wasmExecJS, stylesCSS)\n")
	for _, bundle := range bundles {
		if bundle != "app" {
			initCode.WriteString(fmt.Sprintf("\t\tcore.SetDefaultBundle(%q, wasm%s)\n", bundle, strings.Title(bundle)))
		}
	}
	initCode.WriteString("\t}\n")

	code := fmt.Sprintf(`// Code generated by gux; DO NOT EDIT.
package main

import (
	_ "embed"

	"github.com/dougbarrett/gux/core"
)

%s
%s`, embedCode.String(), initCode.String())

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

// getGuxModulePath finds the gux module in the Go module cache
func getGuxModulePath() (string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/dougbarrett/gux")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to find gux module: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// readSafelist reads the Tailwind safelist file from the gux module
func readSafelist(guxPath string) ([]string, error) {
	safelistPath := filepath.Join(guxPath, "ui", "tailwind-safelist.txt")
	data, err := os.ReadFile(safelistPath)
	if err != nil {
		return nil, err
	}

	var classes []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		classes = append(classes, line)
	}
	return classes, nil
}

// generateSafelistFile generates a file containing class names for Tailwind to scan
func generateSafelistFile(classes []string) error {
	// Create a file that Tailwind's @source can scan for class names
	// Format: class="classname" so Tailwind recognizes them
	var sb strings.Builder
	sb.WriteString("<!-- Auto-generated safelist for gux/ui components -->\n")
	sb.WriteString("<!-- Tailwind scans this file for class names -->\n")

	for _, class := range classes {
		sb.WriteString(fmt.Sprintf("<div class=\"%s\"></div>\n", class))
	}

	return os.WriteFile(".gux/styles/safelist.html", []byte(sb.String()), 0644)
}

// generateTailwindConfig generates Tailwind config files in .gux/
func generateTailwindConfig() error {
	if err := os.MkdirAll(".gux/styles", 0755); err != nil {
		return err
	}

	// Try to find gux module and read safelist
	var safelistImport string
	guxPath, err := getGuxModulePath()
	if err == nil {
		classes, err := readSafelist(guxPath)
		if err == nil && len(classes) > 0 {
			if err := generateSafelistFile(classes); err == nil {
				safelistImport = "\n/* Include gux/ui component classes */\n@source \"./safelist.html\";\n"
				fmt.Printf("  Including %d classes from gux/ui safelist\n", len(classes))
			}
		}
	}

	// Generate input.css for Tailwind v4+
	// Uses @source directive to scan Go files for class names
	inputCSS := fmt.Sprintf(`@import "tailwindcss";

/* Enable class-based dark mode (requires class="dark" on html element) */
@custom-variant dark (&:where(.dark, .dark *));
%s
/* Scan Go files for Tailwind classes */
@source "./pages/**/*.go";
@source "./components/**/*.go";
@source "./*.go";
`, safelistImport)

	return os.WriteFile(".gux/styles/input.css", []byte(inputCSS), 0644)
}

// buildTailwind builds Tailwind CSS using the CLI
func buildTailwind() error {
	fmt.Println("Building Tailwind CSS...")

	// Generate config files
	if err := generateTailwindConfig(); err != nil {
		return fmt.Errorf("failed to generate Tailwind config: %w", err)
	}

	if err := os.MkdirAll(".gux/dist", 0755); err != nil {
		return err
	}

	// Try global tailwindcss first (works with npm global, ruby gem, standalone binary)
	// Then fall back to npx if global not found
	// Tailwind v4 doesn't need -c config flag, uses CSS directives
	args := []string{
		"-i", ".gux/styles/input.css",
		"-o", ".gux/dist/styles.css",
		"--minify",
	}

	cmd := exec.Command("tailwindcss", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Try npx as fallback (for local npm installs)
		cmd = exec.Command("npx", append([]string{"tailwindcss"}, args...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Tailwind CSS build failed (install with: npm install -D tailwindcss or gem install tailwindcss-ruby): %w", err)
		}
	}

	// Get size for display
	info, err := os.Stat(".gux/dist/styles.css")
	if err != nil {
		return err
	}
	sizeKB := float64(info.Size()) / 1024
	fmt.Printf("Built .gux/dist/styles.css (%.2f KB)\n", sizeKB)

	return nil
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

	// Parse routes and bundles from app file
	bundles, imports, err := parseRoutesAndBundles(appFile)
	if err != nil {
		fmt.Printf("Error parsing routes: %v\n", err)
		os.Exit(1)
	}

	// Count total routes
	totalRoutes := 0
	for _, bundle := range bundles {
		totalRoutes += len(bundle.Routes)
	}
	if totalRoutes == 0 {
		fmt.Println("Warning: no Hybrid routes found")
	}

	// Collect bundle names
	bundleNames := make([]string, 0, len(bundles))
	for name := range bundles {
		bundleNames = append(bundleNames, name)
	}

	// Parse CRUD models from app file
	crudModels, modelsImport, dtoImport, err := parseCRUDModels(appFile)
	if err != nil {
		fmt.Printf("Error parsing CRUD models: %v\n", err)
		os.Exit(1)
	}

	// If no models import found, construct it from current package path
	if modelsImport == "" && len(crudModels) > 0 {
		pkgPath, err := getCurrentPackagePath()
		if err != nil {
			fmt.Printf("Error getting package path: %v\n", err)
			os.Exit(1)
		}
		modelsImport = pkgPath + "/models"
	}

	// If no dto import found but we have DTOs, construct it
	if dtoImport == "" {
		for _, m := range crudModels {
			if m.ListDTO != "" || m.DetailDTO != "" {
				pkgPath, err := getCurrentPackagePath()
				if err == nil {
					dtoImport = pkgPath + "/dto"
				}
				break
			}
		}
	}

	// Generate API client if there are CRUD models
	if len(crudModels) > 0 {
		fmt.Println("Generating API client...")

		// Parse DTO files to get field mappings
		for i := range crudModels {
			m := &crudModels[i]
			if m.ListDTO != "" {
				info, err := parseDTOFile("dto", m.ListDTO)
				if err != nil {
					fmt.Printf("Warning: could not parse DTO %s: %v\n", m.ListDTO, err)
				} else {
					m.ListDTOInfo = info
				}
			}
			if m.DetailDTO != "" {
				info, err := parseDTOFile("dto", m.DetailDTO)
				if err != nil {
					fmt.Printf("Warning: could not parse DTO %s: %v\n", m.DetailDTO, err)
				} else {
					m.DetailDTOInfo = info
				}
			}
		}

		if err := generateAPIClient(modelsImport, dtoImport, crudModels); err != nil {
			fmt.Printf("Error generating API client: %v\n", err)
			os.Exit(1)
		}
	}

	// Generate WASM entry points for each bundle
	fmt.Println("Generating WASM entry points...")
	for name, bundle := range bundles {
		if len(bundle.Routes) == 0 {
			continue // Skip bundles with no routes
		}
		if err := generateBundleWasmEntryPoint(name, bundle); err != nil {
			fmt.Printf("Error generating WASM entry for bundle %s: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("  Generated entry point for bundle: %s (%d routes)\n", name, len(bundle.Routes))
	}

	// Copy wasm_exec.js
	if err := copyWasmExec(tinygo); err != nil {
		fmt.Printf("Error copying wasm_exec.js: %v\n", err)
		os.Exit(1)
	}

	// Build WASM for each bundle
	for name, bundle := range bundles {
		if len(bundle.Routes) == 0 {
			continue // Skip bundles with no routes
		}
		if err := buildWasmBundle(name, tinygo); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Build Tailwind CSS
	if err := buildTailwind(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Generate assets_gen.go with all bundles
	fmt.Println("Generating assets...")
	if err := generateAssetsFile(modulePath, bundleNames); err != nil {
		fmt.Printf("Error generating assets: %v\n", err)
		os.Exit(1)
	}

	// Silence unused variable warning for imports (used for debugging)
	_ = imports

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

// runDevNew builds and runs the server (does NOT clean up on exit to preserve .gux files)
func runDevNew(tinygo bool, watch bool) {
	// Build first
	runBuildNew(tinygo)

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("\nStarting dev server...\n")
	if watch {
		fmt.Println("Hot reload enabled - watching for file changes...")
	}

	// Channel to signal rebuild needed
	rebuildChan := make(chan struct{}, 1)

	// Start file watcher if watch mode enabled
	if watch {
		go watchAndRegenerate(true, func() {
			select {
			case rebuildChan <- struct{}{}:
			default:
			}
		})
	}

	for {
		// Run the binary
		cmd := exec.Command("./bin/app")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		// Set hot reload env var if watch mode enabled
		if watch {
			cmd.Env = append(os.Environ(), "GUX_HOT_RELOAD=1")
		}

		if err := cmd.Start(); err != nil {
			fmt.Printf("Failed to start: %v\n", err)
			os.Exit(1)
		}

		// Wait for signal, rebuild request, or process exit
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-sigChan:
			fmt.Println("\nShutting down...")
			cmd.Process.Signal(os.Interrupt)
			<-done
			return // Exit without cleaning
		case <-rebuildChan:
			// Kill current server and rebuild
			fmt.Println("\nRebuilding...")
			cmd.Process.Signal(os.Interrupt)
			<-done
			runBuildNew(tinygo)
			fmt.Println("\nRestarting server...")
			continue // Restart loop
		case err := <-done:
			if err != nil {
				fmt.Printf("Server exited with error: %v\n", err)
			}
			return // Exit without cleaning
		}
	}
}
