package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// runGenNew generates all guxgen files without building WASM or binary
func runGenNew(watch bool, updateDockerfile bool, serverDir string) {
	// Initial generation
	if err := generateGuxFiles(serverDir); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Regenerate Dockerfile from template if requested
	if updateDockerfile {
		if err := regenerateDockerfile(); err != nil {
			fmt.Printf("Error updating Dockerfile: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Dockerfile updated from template.")
	}

	if !watch {
		fmt.Println("\nGeneration complete! Run 'gux build' or 'gux dev' to compile.")
		return
	}

	// Watch mode
	fmt.Println("\nWatching for file changes... (Ctrl+C to stop)")
	watchAndRegenerate(false, nil, serverDir)
}

// generateGuxFiles generates all files in the guxgen directory.
// serverDir optionally specifies a subdirectory containing the app entry point.
func generateGuxFiles(serverDir string) error {
	// Get module path
	modulePath, err := getModulePath()
	if err != nil {
		return fmt.Errorf("getting module path: %w", err)
	}

	// Regenerate models from gux.config.json if it exists
	if config, err := LoadModelsConfig("."); err == nil && len(config.Models) > 0 {
		fmt.Printf("Regenerating %d model(s) from config...\n", len(config.Models))

		// Build display fields map for all models
		displayFields := BuildDisplayFieldsMap(".")

		// Collect Brief DTOs needed across all models
		briefDTOs := make(map[string]BriefDTOInfo)

		// Collect model names for hooks bridge generation
		var modelNames []string

		for name, model := range config.Models {
			// Apply roles to auth preset models
			if model.Preset == "auth" && len(config.Roles) > 0 {
				applyRolesToAuthModel(&model, config.Roles)
			}
			model.Name = name
			modelNames = append(modelNames, name)

			// Collect relations for Brief DTO generation
			// Generate briefs for ALL relations (including external models not in config)
			for _, fields := range model.Sections {
				for _, field := range fields {
					if field.Relation != "" {
						if _, exists := briefDTOs[field.Relation]; !exists {
							display := displayFields[field.Relation]
							info := BriefDTOInfo{
								Model:        field.Relation,
								DisplayField: display,
							}
							if isDisplayTemplate(display) {
								info.DisplayFields = extractDisplayFields(display)
								info.DisplayFormat = display
							} else {
								info.DisplayFields = extractDisplayFields(display)
							}
							briefDTOs[field.Relation] = info
						}
					}
				}
			}

			if err := GenerateModelFiles(&model, config.OptionSets); err != nil {
				fmt.Printf("  %s: error - %v\n", name, err)
			} else {
				fmt.Printf("  %s: done\n", name)
			}
		}

		// Generate shared Brief DTOs file
		if len(briefDTOs) > 0 {
			if err := GenerateSharedBriefsFile(briefDTOs); err != nil {
				fmt.Printf("  briefs_gen.go: error - %v\n", err)
			} else {
				fmt.Printf("  guxgen/dto/briefs_gen.go: done\n")
			}
		}

		// Generate consolidated hooks bridge file
		if err := GenerateHooksBridgeFile(modelNames, modulePath); err != nil {
			fmt.Printf("  hooks_gen.go: error - %v\n", err)
		}

		fmt.Println()
	}

	// Regenerate scaffold files in guxgen/ from templates
	if err := regenerateScaffoldFiles(modulePath); err != nil {
		fmt.Printf("Warning: scaffold regeneration failed: %v\n", err)
	}

	// Find main app file
	appFile, err := findMainAppFile(serverDir)
	if err != nil {
		return fmt.Errorf("finding app file: %w", err)
	}

	// Parse routes and bundles from app file (follows imports if needed)
	bundles, _, err := parseRoutesWithImportFollowing(appFile)
	if err != nil {
		return fmt.Errorf("parsing routes: %w", err)
	}

	// Count total routes
	totalRoutes := 0
	for _, bundle := range bundles {
		totalRoutes += len(bundle.Routes)
	}
	if totalRoutes == 0 {
		fmt.Println("Warning: no Hybrid routes found")
	}

	// Collect bundle names (only bundles that have routes)
	bundleNames := make([]string, 0, len(bundles))
	for name, bundle := range bundles {
		if len(bundle.Routes) > 0 {
			bundleNames = append(bundleNames, name)
		}
	}
	hasWASM := len(bundleNames) > 0

	// Parse CRUD models from app file (follows imports if needed)
	crudModels, modelsImport, dtoImport, err := parseCRUDModelsWithImportFollowing(appFile)
	if err != nil {
		return fmt.Errorf("parsing CRUD models: %w", err)
	}

	// Check if we have a gux.config.json - if so, use guxgen/ paths
	hasGuxConfig := false
	if config, err := LoadModelsConfig("."); err == nil && len(config.Models) > 0 {
		hasGuxConfig = true

		guxgenModelsPath := modulePath + "/guxgen/models"
		guxgenDTOPath := modulePath + "/guxgen/dto"

		// Determine model and DTO externality from actual import paths.
		for i := range crudModels {
			if crudModels[i].ModelImportPath != "" && crudModels[i].ModelImportPath != guxgenModelsPath {
				crudModels[i].IsExternal = true
			} else if crudModels[i].ModelImportPath == "" {
				if _, inConfig := config.Models[crudModels[i].Name]; !inConfig {
					crudModels[i].IsExternal = true
				}
			}

			if crudModels[i].DTOImportPath != "" && crudModels[i].DTOImportPath != guxgenDTOPath {
				crudModels[i].DTOIsExternal = true
			}
		}

		// Auto-populate ListDTO/DetailDTO for config models that don't have
		// explicit WithListDTO/WithDetailDTO in app.go. Scaffolded models always
		// generate DTOs in guxgen/dto/ ({Model}List, {Model}Detail), so the
		// CRUD client should use them even if app.go omits the DTO options.
		for i := range crudModels {
			if crudModels[i].IsExternal {
				continue
			}
			if crudModels[i].ListDTO == "" {
				crudModels[i].ListDTO = crudModels[i].Name + "List"
				if crudModels[i].DTOPackage == "" {
					crudModels[i].DTOPackage = "dto"
				}
			}
			if crudModels[i].DetailDTO == "" {
				crudModels[i].DetailDTO = crudModels[i].Name + "Detail"
				if crudModels[i].DTOPackage == "" {
					crudModels[i].DTOPackage = "dto"
				}
			}
		}
	}

	// For projects with gux.config.json, always use guxgen/ paths
	// This ensures generated DTOs and models are correctly imported
	// Note: use modulePath (from go.mod) instead of getCurrentPackagePath() (go list .)
	// because go list fails when generated packages don't exist yet (e.g., after gux clean)
	if hasGuxConfig {
		modelsImport = modulePath + "/guxgen/models"
		dtoImport = modulePath + "/guxgen/dto"
	} else {
		// Legacy: If no models import found, construct it from module path
		if modelsImport == "" && len(crudModels) > 0 {
			modelsImport = modulePath + "/guxgen/models"
		}

		// Legacy: If no dto import found but we have DTOs, construct it
		if dtoImport == "" {
			for _, m := range crudModels {
				if m.ListDTO != "" || m.DetailDTO != "" {
					dtoImport = modulePath + "/guxgen/dto"
					break
				}
			}
		}
	}

	// Generate API client if there are CRUD models
	if len(crudModels) > 0 {
		fmt.Println("Generating API client...")

		// Parse DTO files to get field mappings (check guxgen/dto first, then dto)
		for i := range crudModels {
			m := &crudModels[i]
			if m.ListDTO != "" {
				info, err := parseDTOFile("guxgen/dto", m.ListDTO)
				if err != nil {
					// Fallback to dto/ for backwards compatibility
					info, err = parseDTOFile("dto", m.ListDTO)
				}
				if err != nil {
					fmt.Printf("Warning: could not parse DTO %s: %v\n", m.ListDTO, err)
				} else {
					m.ListDTOInfo = info
				}
			}
			if m.DetailDTO != "" {
				info, err := parseDTOFile("guxgen/dto", m.DetailDTO)
				if err != nil {
					// Fallback to dto/ for backwards compatibility
					info, err = parseDTOFile("dto", m.DetailDTO)
				}
				if err != nil {
					fmt.Printf("Warning: could not parse DTO %s: %v\n", m.DetailDTO, err)
				} else {
					m.DetailDTOInfo = info
				}
			}
		}

		if err := generateAPIClient(modelsImport, dtoImport, crudModels); err != nil {
			return fmt.Errorf("generating API client: %w", err)
		}
	}

	// Parse typed API endpoints from app file (follows imports if needed)
	apiEndpoints, endpointDTOImports, err := parseAPIEndpointsWithImportFollowing(appFile)
	if err != nil {
		return fmt.Errorf("parsing API endpoints: %w", err)
	}

	// Generate API endpoint client if there are endpoints
	if len(apiEndpoints) > 0 {
		fmt.Printf("Generating API endpoint clients (%d endpoints)...\n", len(apiEndpoints))

		// Merge CRUD dto import into endpoint imports if not already present
		if dtoImport != "" && len(endpointDTOImports) == 0 {
			// Derive alias from import path
			parts := strings.Split(dtoImport, "/")
			endpointDTOImports = map[string]string{parts[len(parts)-1]: dtoImport}
		}

		// Collect CRUD plural names to detect name collisions with endpoint functions
		crudNames := make(map[string]bool)
		for _, m := range crudModels {
			crudNames[m.PluralName] = true
		}

		if err := generateEndpointClient(apiEndpoints, endpointDTOImports, crudNames, appFile); err != nil {
			return fmt.Errorf("generating API endpoint client: %w", err)
		}
	}

	// Generate JavaScript API client for Alpine.js mode
	fmt.Println("Generating JavaScript API client...")
	if err := generateAPIJS(apiEndpoints, crudModels); err != nil {
		return fmt.Errorf("generating API JS: %w", err)
	}

	// Determine if we have any async API calls (CRUD or endpoints)
	hasAsyncAPI := len(crudModels) > 0 || len(apiEndpoints) > 0

	// Generate load tracker for async CRUD/endpoint operations
	if hasAsyncAPI {
		if err := generateLoadTracker(); err != nil {
			return fmt.Errorf("generating load tracker: %w", err)
		}
	}

	// Generate WASM entry points for each bundle
	if hasWASM {
		fmt.Println("Generating WASM entry points...")
		for name, bundle := range bundles {
			if len(bundle.Routes) == 0 {
				continue // Skip bundles with no routes
			}
			apiImport := ""
			if hasAsyncAPI {
				apiImport = modulePath + "/guxgen/api"
			}
			if err := generateBundleWasmEntryPoint(name, bundle, apiImport); err != nil {
				return fmt.Errorf("generating WASM entry for bundle %s: %w", name, err)
			}
			fmt.Printf("  Generated entry point for bundle: %s (%d routes)\n", name, len(bundle.Routes))
		}

		// Copy wasm_exec.js (use TinyGo by default)
		if err := copyWasmExec(true); err != nil {
			return fmt.Errorf("copying wasm_exec.js: %w", err)
		}
	} else {
		fmt.Println("No Hybrid routes — skipping WASM generation")
	}

	// Generate assets_gen.go with all bundles
	fmt.Println("Generating assets...")
	if err := generateAssetsFile(modulePath, bundleNames, serverDir, hasWASM); err != nil {
		return fmt.Errorf("generating assets: %w", err)
	}

	return nil
}

// watchAndRegenerate watches for file changes and regenerates/rebuilds
// If fullBuild is true, it rebuilds WASM and binary (for dev mode)
// If notifyReload is not nil, it's called after successful rebuild
func watchAndRegenerate(fullBuild bool, notifyReload func(), serverDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Error creating watcher: %v\n", err)
		os.Exit(1)
	}
	defer watcher.Close()

	// Add directories to watch (excluding guxgen and bin)
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			// Skip guxgen, bin, .git, and hidden directories
			name := info.Name()
			if name == "guxgen" || name == "bin" || name == ".git" || (len(name) > 0 && name[0] == '.') {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error setting up watcher: %v\n", err)
		os.Exit(1)
	}

	// Debounce timer
	var debounceTimer *time.Timer
	debounceDuration := 500 * time.Millisecond

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Only watch .go, .templ, .css, .html files
			ext := filepath.Ext(event.Name)
			if ext != ".go" && ext != ".templ" && ext != ".css" && ext != ".html" {
				continue
			}
			// Skip generated files
			base := filepath.Base(event.Name)
			if base == "assets_gen.go" || strings.HasSuffix(base, "_gen.go") {
				continue
			}

			// Debounce: reset timer on each event
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(debounceDuration, func() {
				fmt.Printf("\n[%s] File changed: %s\n", time.Now().Format("15:04:05"), event.Name)
				if fullBuild {
					// Full rebuild for dev mode
					fmt.Println("Rebuilding...")
					// Note: caller handles actual rebuild
					if notifyReload != nil {
						notifyReload()
					}
				} else {
					// Just regenerate guxgen files
					if err := generateGuxFiles(serverDir); err != nil {
						fmt.Printf("Error regenerating: %v\n", err)
					} else {
						fmt.Println("Regeneration complete!")
					}
				}
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}

// regenerateDockerfile regenerates the Dockerfile from the embedded template
func regenerateDockerfile() error {
	// Get module path from go.mod
	modulePath, err := getModulePath()
	if err != nil {
		return fmt.Errorf("getting module path: %w", err)
	}

	// Get gux version
	guxVersion := getVersion()
	if guxVersion == "dev" {
		guxVersion = "latest"
	}

	// Get app name from current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}
	appName := filepath.Base(cwd)

	data := TemplateData{
		AppName:    appName,
		ModulePath: modulePath,
		GuxModule:  "github.com/dougbarrett/gux",
		GuxVersion: guxVersion,
	}

	return renderTemplate(".", "templates/Dockerfile.tmpl", "Dockerfile", data)
}

// regenerateScaffoldFiles regenerates scaffold template files in guxgen/
// that were originally created by gux init (layout, dashboard, settings, auth DTOs).
// This ensures gux clean && gux gen produces a complete guxgen/ directory.
func regenerateScaffoldFiles(modulePath string) error {
	guxConfig, err := LoadConfig(".")
	if err != nil {
		// No config file — nothing to regenerate
		return nil
	}

	// Derive app name from directory (same as gux init)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}
	appName := filepath.Base(cwd)

	guxVersion := getVersion()
	if guxVersion == "dev" {
		guxVersion = "latest"
	}

	withAuth := guxConfig.Auth != "" && guxConfig.Auth != "none"
	publicSignup := guxConfig.Auth == "public"

	data := TemplateData{
		AppName:      appName,
		ModulePath:   modulePath,
		GuxModule:    "github.com/dougbarrett/gux",
		GuxVersion:   guxVersion,
		WithAuth:     withAuth,
		PublicSignup: publicSignup,
		WithAdmin:    guxConfig.Admin,
		Port:         guxConfig.Port,
	}

	// Regenerate admin scaffold files
	if guxConfig.Admin {
		scaffoldFiles := []struct {
			tmplPath string
			destPath string
		}{
			{"templates/admin/admin/layout.go.tmpl", "guxgen/admin/layout.go"},
			{"templates/admin/admin/dashboard.go.tmpl", "guxgen/admin/dashboard.go"},
			{"templates/admin/admin/settings.go.tmpl", "guxgen/admin/settings.go"},
		}

		// Admin without auth also gets static user pages and breadcrumb DTO
		if !withAuth {
			scaffoldFiles = append(scaffoldFiles,
				struct {
					tmplPath string
					destPath string
				}{"templates/admin/admin/users.go.tmpl", "guxgen/admin/users.go"},
				struct {
					tmplPath string
					destPath string
				}{"templates/admin/admin/user_new.go.tmpl", "guxgen/admin/user_new.go"},
				struct {
					tmplPath string
					destPath string
				}{"templates/admin/admin/user_detail.go.tmpl", "guxgen/admin/user_detail.go"},
				struct {
					tmplPath string
					destPath string
				}{"templates/admin/dto/breadcrumb.go.tmpl", "guxgen/dto/breadcrumb.go"},
			)
		}

		for _, f := range scaffoldFiles {
			if err := renderTemplate(".", f.tmplPath, f.destPath, data); err != nil {
				return fmt.Errorf("regenerate %s: %w", f.destPath, err)
			}
		}
	}

	// Regenerate auth DTO files (LoginRequest, LoginResponse, LogoutResponse, Breadcrumb)
	if withAuth {
		if err := renderTemplate(".", "templates/admin/dto/auth.go.tmpl", "guxgen/dto/auth.go", data); err != nil {
			return fmt.Errorf("regenerate guxgen/dto/auth.go: %w", err)
		}

		// Check if project has dto/ with conflicting auth types.
		// If so, regenerate guxgen/dto/auth.go with type aliases so that
		// guxgen/dto.X and dto.X are the same Go type.
		dtoTypes := getExportedTypeNames("dto")
		if len(dtoTypes) > 0 {
			authTypes := []string{"LoginRequest", "LoginResponse", "LogoutResponse", "Breadcrumb"}
			var aliasedTypes []string
			for _, t := range authTypes {
				if dtoTypes[t] {
					aliasedTypes = append(aliasedTypes, t)
				}
			}
			if len(aliasedTypes) > 0 {
				if err := generateAuthDTOWithAliases(modulePath, aliasedTypes); err != nil {
					return fmt.Errorf("regenerate guxgen/dto/auth.go with aliases: %w", err)
				}
			}

			// Scaffold admin templates reference UserList and UserDetail
			// (dashboard, users list, user detail). When the User model is
			// the built-in auth model (not in gux.config.json),
			// GenerateModelFilesImpl won't run for it, so we generate
			// type aliases here for any User DTOs found in dto/.
			if err := generateScaffoldDTOAliases(modulePath, dtoTypes); err != nil {
				return fmt.Errorf("generate scaffold DTO aliases: %w", err)
			}
		}
	}

	return nil
}

// getExportedTypeNames parses a Go package directory and returns all exported type names.
func getExportedTypeNames(dir string) map[string]bool {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		return nil
	}
	types := make(map[string]bool)
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					continue
				}
				for _, spec := range genDecl.Specs {
					typeSpec := spec.(*ast.TypeSpec)
					if typeSpec.Name.IsExported() {
						types[typeSpec.Name.Name] = true
					}
				}
			}
		}
	}
	return types
}

// generateAuthDTOWithAliases generates guxgen/dto/auth.go with type aliases for
// auth types that exist in the project's dto/ package, and fresh definitions for
// types that don't. This ensures guxgen/dto.X and dto.X are the same Go type.
func generateAuthDTOWithAliases(modulePath string, aliasedTypes []string) error {
	aliasSet := make(map[string]bool)
	for _, t := range aliasedTypes {
		aliasSet[t] = true
	}

	var buf strings.Builder
	buf.WriteString("// Code generated by gux; DO NOT EDIT.\npackage dto\n\n")
	buf.WriteString(fmt.Sprintf("import manualdto \"%s/dto\"\n\n", modulePath))

	// Type aliases for types found in the project's dto/ package
	for _, t := range aliasedTypes {
		buf.WriteString(fmt.Sprintf("type %s = manualdto.%s\n\n", t, t))
	}

	// Fresh definitions for types NOT found in the project's dto/ package
	if !aliasSet["LoginRequest"] {
		buf.WriteString("// LoginRequest is the request body for login.\ntype LoginRequest struct {\n\tEmail    string `json:\"email\"`\n\tPassword string `json:\"password\"`\n}\n\n")
	}
	if !aliasSet["LoginResponse"] {
		buf.WriteString("// LoginResponse is the response from login.\ntype LoginResponse struct {\n\tSuccess  bool   `json:\"success\"`\n\tRedirect string `json:\"redirect,omitempty\"`\n\tError    string `json:\"error,omitempty\"`\n}\n\n")
	}
	if !aliasSet["LogoutResponse"] {
		buf.WriteString("// LogoutResponse is the response from logout.\ntype LogoutResponse struct {\n\tSuccess  bool   `json:\"success\"`\n\tRedirect string `json:\"redirect,omitempty\"`\n}\n\n")
	}
	if !aliasSet["Breadcrumb"] {
		buf.WriteString("// Breadcrumb represents a breadcrumb navigation item.\ntype Breadcrumb struct {\n\tLabel string\n\tHref  string\n}\n")
	}

	return os.WriteFile(filepath.Join("guxgen", "dto", "auth.go"), []byte(buf.String()), 0644)
}

// generateScaffoldDTOAliases generates type aliases in guxgen/dto/ for model
// DTOs referenced by scaffold admin templates that exist in the project's dto/
// but might not be generated by GenerateModelFilesImpl (e.g., the built-in User
// model when it's not in gux.config.json).
func generateScaffoldDTOAliases(modulePath string, dtoTypes map[string]bool) error {
	// Model DTOs referenced by scaffold admin templates
	scaffoldModels := []string{"User"}

	for _, model := range scaffoldModels {
		listType := model + "List"
		detailType := model + "Detail"

		if !dtoTypes[listType] && !dtoTypes[detailType] {
			continue // Neither type exists in dto/
		}

		aliasPath := filepath.Join("guxgen", "dto", ToSnakeCase(model)+"_gen.go")

		// Skip if already generated (e.g., by GenerateModelFilesImpl)
		if _, err := os.Stat(aliasPath); err == nil {
			continue
		}

		var buf strings.Builder
		buf.WriteString(fmt.Sprintf("// Code generated by gux; DO NOT EDIT.\npackage dto\n\nimport manualdto \"%s/dto\"\n\n", modulePath))
		if dtoTypes[listType] {
			buf.WriteString(fmt.Sprintf("type %s = manualdto.%s\n\n", listType, listType))
		}
		if dtoTypes[detailType] {
			buf.WriteString(fmt.Sprintf("type %s = manualdto.%s\n\n", detailType, detailType))
		}

		if err := os.MkdirAll(filepath.Dir(aliasPath), 0755); err != nil {
			return fmt.Errorf("create guxgen/dto dir: %w", err)
		}
		if err := os.WriteFile(aliasPath, []byte(buf.String()), 0644); err != nil {
			return fmt.Errorf("generate scaffold DTO alias %s: %w", aliasPath, err)
		}
	}
	return nil
}

