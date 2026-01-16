package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

//go:embed templates/*
var templates embed.FS

// TemplateData holds the variables for template substitution
type TemplateData struct {
	AppName    string
	ModulePath string
	GuxModule  string
	GuxVersion string
}

func runInit(appName, modulePath string) {
	// Check if initializing in current directory
	initHere := appName == "."
	var targetDir string

	if initHere {
		// Initialize in current directory
		targetDir = "."

		// Get current directory name for app name
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		appName = filepath.Base(cwd)

		// Require --module when initializing in current directory
		if modulePath == "" {
			fmt.Println("Error: --module is required when initializing in current directory")
			fmt.Printf("Usage: gux init --module github.com/youruser/%s .\n", appName)
			os.Exit(1)
		}

		// Check if directory has conflicting files
		conflicts := checkForConflicts(targetDir)
		if len(conflicts) > 0 {
			fmt.Println("Error: directory contains files that would be overwritten:")
			for _, f := range conflicts {
				fmt.Printf("  - %s\n", f)
			}
			os.Exit(1)
		}
	} else {
		// Validate app name for new directory
		if !isValidAppName(appName) {
			fmt.Printf("Error: invalid app name '%s'\n", appName)
			fmt.Println("App name must contain only lowercase letters, numbers, hyphens, and underscores.")
			os.Exit(1)
		}

		targetDir = appName

		// Determine module path
		if modulePath == "" {
			modulePath = appName
			fmt.Printf("Note: No --module specified, using '%s' as module path.\n", modulePath)
			fmt.Printf("      For proper imports, consider: gux init %s --module github.com/youruser/%s\n\n", appName, appName)
		}

		// Create target directory
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
			os.Exit(1)
		}

		// Check if directory is empty
		entries, _ := os.ReadDir(targetDir)
		if len(entries) > 0 {
			fmt.Printf("Error: directory '%s' is not empty\n", targetDir)
			os.Exit(1)
		}
	}

	// Get gux version for go.mod pinning
	guxVersion := getVersion()
	if guxVersion == "dev" {
		guxVersion = "latest"
	}

	data := TemplateData{
		AppName:    appName,
		ModulePath: modulePath,
		GuxModule:  "github.com/dougbarrett/gux",
		GuxVersion: guxVersion,
	}

	// Define files to create from templates
	filesToCreate := []struct {
		tmplPath string
		destPath string
	}{
		{"templates/go.mod.tmpl", "go.mod"},
		{"templates/cmd/app/main.go.tmpl", "cmd/app/main.go"},
		{"templates/cmd/server/main.go.tmpl", "cmd/server/main.go"},
		{"templates/internal/api/types.go.tmpl", "internal/api/types.go"},
		{"templates/internal/api/example.go.tmpl", "internal/api/example.go"},
		{"templates/public/index.html.tmpl", "public/index.html"},
		{"templates/public/manifest.json.tmpl", "public/manifest.json"},
		{"templates/public/service-worker.js.tmpl", "public/service-worker.js"},
		{"templates/Dockerfile.tmpl", "Dockerfile"},
	}

	fmt.Printf("Creating Gux application '%s'...\n\n", appName)

	for _, f := range filesToCreate {
		if err := renderTemplate(targetDir, f.tmplPath, f.destPath, data); err != nil {
			fmt.Printf("Error creating %s: %v\n", f.destPath, err)
			os.Exit(1)
		}
		fmt.Printf("  created %s\n", f.destPath)
	}

	// Run go mod tidy to download dependencies
	fmt.Println("\nRunning go mod tidy...")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: go mod tidy failed: %v\n", err)
		fmt.Println("You may need to run 'go mod tidy' manually.")
	} else {
		fmt.Println("  dependencies downloaded")
	}

	printNextStepsWithDir(appName, initHere)
}

func renderTemplate(targetDir, tmplPath, destPath string, data TemplateData) error {
	// Read template
	content, err := templates.ReadFile(tmplPath)
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}

	// Parse and execute template
	tmpl, err := template.New(filepath.Base(tmplPath)).Parse(string(content))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	// Create destination directory
	fullPath := filepath.Join(targetDir, destPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}

func isValidAppName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// checkForConflicts returns a list of files that would be overwritten
func checkForConflicts(targetDir string) []string {
	filesToCheck := []string{
		"go.mod",
		"cmd/app/main.go",
		"cmd/server/main.go",
		"internal/api/types.go",
		"internal/api/example.go",
		"public/index.html",
		"public/manifest.json",
		"public/service-worker.js",
		"Dockerfile",
	}

	var conflicts []string
	for _, f := range filesToCheck {
		path := filepath.Join(targetDir, f)
		if _, err := os.Stat(path); err == nil {
			conflicts = append(conflicts, f)
		}
	}
	return conflicts
}

func printNextStepsWithDir(appName string, initHere bool) {
	if initHere {
		fmt.Printf(`
Created Gux application in current directory

Next steps:
  gux setup       # Copy wasm_exec.js to public/ (optional for local dev)
  gux dev         # Build and run dev server

Your app will be available at http://localhost:8080

Note: Docker builds don't require wasm_exec.js locally - it's copied from the TinyGo image.
`)
	} else {
		fmt.Printf(`
Created Gux application in ./%s

Next steps:
  cd %s
  gux setup       # Copy wasm_exec.js to public/ (optional for local dev)
  gux dev         # Build and run dev server

Your app will be available at http://localhost:8080

Note: Docker builds don't require wasm_exec.js locally - it's copied from the TinyGo image.
`, appName, appName)
	}
}
