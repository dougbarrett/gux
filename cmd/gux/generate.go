package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// runGenNew generates all guxgen files without building WASM or binary
func runGenNew(watch bool) {
	// Initial generation
	if err := generateGuxFiles(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !watch {
		fmt.Println("\nGeneration complete! Run 'gux build' or 'gux dev' to compile.")
		return
	}

	// Watch mode
	fmt.Println("\nWatching for file changes... (Ctrl+C to stop)")
	watchAndRegenerate(false, nil)
}

// generateGuxFiles generates all files in the guxgen directory
func generateGuxFiles() error {
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
			for _, fields := range model.Sections {
				for _, field := range fields {
					if field.Relation != "" {
						if _, exists := briefDTOs[field.Relation]; !exists {
							briefDTOs[field.Relation] = BriefDTOInfo{
								Model:        field.Relation,
								DisplayField: displayFields[field.Relation],
							}
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

	// Find main app file
	appFile, err := findMainAppFile()
	if err != nil {
		return fmt.Errorf("finding app file: %w", err)
	}

	// Parse routes and bundles from app file
	bundles, _, err := parseRoutesAndBundles(appFile)
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

	// Collect bundle names
	bundleNames := make([]string, 0, len(bundles))
	for name := range bundles {
		bundleNames = append(bundleNames, name)
	}

	// Parse CRUD models from app file
	crudModels, modelsImport, dtoImport, err := parseCRUDModels(appFile)
	if err != nil {
		return fmt.Errorf("parsing CRUD models: %w", err)
	}

	// Check if we have a gux.config.json - if so, use guxgen/ paths
	hasGuxConfig := false
	if config, err := LoadModelsConfig("."); err == nil && len(config.Models) > 0 {
		hasGuxConfig = true
	}

	// For projects with gux.config.json, always use guxgen/ paths
	// This ensures generated DTOs and models are correctly imported
	if hasGuxConfig {
		pkgPath, err := getCurrentPackagePath()
		if err != nil {
			return fmt.Errorf("getting package path: %w", err)
		}
		modelsImport = pkgPath + "/guxgen/models"
		dtoImport = pkgPath + "/guxgen/dto"
	} else {
		// Legacy: If no models import found, construct it from current package path
		if modelsImport == "" && len(crudModels) > 0 {
			pkgPath, err := getCurrentPackagePath()
			if err != nil {
				return fmt.Errorf("getting package path: %w", err)
			}
			modelsImport = pkgPath + "/guxgen/models"
		}

		// Legacy: If no dto import found but we have DTOs, construct it
		if dtoImport == "" {
			for _, m := range crudModels {
				if m.ListDTO != "" || m.DetailDTO != "" {
					pkgPath, err := getCurrentPackagePath()
					if err == nil {
						dtoImport = pkgPath + "/guxgen/dto"
					}
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

	// Parse typed API endpoints from app file
	apiEndpoints, endpointDTOImport, err := parseAPIEndpoints(appFile)
	if err != nil {
		return fmt.Errorf("parsing API endpoints: %w", err)
	}

	// Generate API endpoint client if there are endpoints
	if len(apiEndpoints) > 0 {
		fmt.Printf("Generating API endpoint clients (%d endpoints)...\n", len(apiEndpoints))

		// Use dto import from endpoints or from CRUD
		if endpointDTOImport == "" {
			endpointDTOImport = dtoImport
		}

		if err := generateEndpointClient(apiEndpoints, endpointDTOImport); err != nil {
			return fmt.Errorf("generating API endpoint client: %w", err)
		}
	}

	// Generate WASM entry points for each bundle
	fmt.Println("Generating WASM entry points...")
	for name, bundle := range bundles {
		if len(bundle.Routes) == 0 {
			continue // Skip bundles with no routes
		}
		if err := generateBundleWasmEntryPoint(name, bundle); err != nil {
			return fmt.Errorf("generating WASM entry for bundle %s: %w", name, err)
		}
		fmt.Printf("  Generated entry point for bundle: %s (%d routes)\n", name, len(bundle.Routes))
	}

	// Copy wasm_exec.js (use TinyGo by default)
	if err := copyWasmExec(true); err != nil {
		return fmt.Errorf("copying wasm_exec.js: %w", err)
	}

	// Generate assets_gen.go with all bundles
	fmt.Println("Generating assets...")
	if err := generateAssetsFile(modulePath, bundleNames); err != nil {
		return fmt.Errorf("generating assets: %w", err)
	}

	return nil
}

// watchAndRegenerate watches for file changes and regenerates/rebuilds
// If fullBuild is true, it rebuilds WASM and binary (for dev mode)
// If notifyReload is not nil, it's called after successful rebuild
func watchAndRegenerate(fullBuild bool, notifyReload func()) {
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
					if err := generateGuxFiles(); err != nil {
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

// Legacy generate functions below - kept for backward compatibility

func runGenerate(apiDir string) {
	// Check if directory exists
	info, err := os.Stat(apiDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Error: directory '%s' does not exist\n", apiDir)
			os.Exit(1)
		}
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Printf("Error: '%s' is not a directory\n", apiDir)
		os.Exit(1)
	}

	// Find all .go files with @client annotation
	files, err := findAPIFiles(apiDir)
	if err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Printf("No API interface files found in '%s'\n", apiDir)
		fmt.Println("API files should contain a '@client' annotation in interface comments.")
		return
	}

	fmt.Printf("Generating API clients from %d file(s)...\n\n", len(files))

	// Generate shared client code once
	sharedCode, err := GenerateClientSharedCode()
	if err != nil {
		fmt.Printf("Error generating shared client code: %v\n", err)
		os.Exit(1)
	}
	sharedPath := filepath.Join(apiDir, "client_shared_gen.go")
	if err := os.WriteFile(sharedPath, []byte(sharedCode), 0644); err != nil {
		fmt.Printf("Error writing shared client code: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  generated: %s\n\n", sharedPath)

	for _, file := range files {
		// Generate output filename: foo.go -> foo_client_gen.go
		base := strings.TrimSuffix(filepath.Base(file), ".go")
		outputFile := base + "_client_gen.go"

		fmt.Printf("  %s:\n", filepath.Base(file))

		if err := GenerateAPI(file, outputFile); err != nil {
			fmt.Printf("Error generating %s: %v\n", file, err)
			os.Exit(1)
		}
	}

	fmt.Printf("\nGenerated %d API file(s) + shared client code\n", len(files))

	// Check for updates
	checkForUpdates()
}

// findAPIFiles finds all .go files in the directory that contain @client annotation
func findAPIFiles(dir string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}

		// Skip generated files
		if strings.HasSuffix(name, "_gen.go") {
			continue
		}

		// Check if file contains @client annotation
		fullPath := filepath.Join(dir, name)
		if hasClientAnnotation(fullPath) {
			files = append(files, fullPath)
		}
	}

	return files, nil
}

// hasClientAnnotation checks if a file contains the @client annotation
func hasClientAnnotation(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "@client") {
			return true
		}
	}

	return false
}
