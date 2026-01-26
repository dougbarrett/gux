package main

import (
	"fmt"
	"os"
	"sort"
	"text/template"

	"golang.org/x/mod/modfile"
)

// Pattern holds metadata and template for a help pattern.
type Pattern struct {
	Name        string // e.g., "page", "page:list"
	Description string // e.g., "Basic page template with OnLoad pattern"
	FilePath    string // e.g., "pages/example.go"
	Template    string // The actual template code
}

// patterns is the registry of all available help patterns.
// Will be populated in Plan 02 with actual patterns.
var patterns = map[string]Pattern{
	// Example placeholder - real patterns added in next plan
}

// getModulePathForHelp reads go.mod and extracts the module path.
// Uses modfile.ModulePath for tolerant parsing (doesn't require valid go.mod).
// Returns a placeholder and prints warning to stderr if go.mod is not found.
func getModulePathForHelp() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: Could not read go.mod. Using placeholder import path.")
		fmt.Fprintln(os.Stderr, "Run from a Go module directory or replace 'your-module-path' manually.")
		fmt.Fprintln(os.Stderr, "")
		return "your-module-path"
	}

	path := modfile.ModulePath(data)
	if path == "" {
		fmt.Fprintln(os.Stderr, "Warning: Could not parse module path from go.mod. Using placeholder.")
		fmt.Fprintln(os.Stderr, "Replace 'your-module-path' with your actual module path.")
		fmt.Fprintln(os.Stderr, "")
		return "your-module-path"
	}

	return path
}

// printPatternList prints all available patterns sorted alphabetically.
func printPatternList() {
	fmt.Println("Available patterns:")

	if len(patterns) == 0 {
		fmt.Println("  (no patterns registered yet)")
		fmt.Println("")
		fmt.Println("Patterns will be added in a future update.")
		return
	}

	// Sort pattern names alphabetically
	names := make([]string, 0, len(patterns))
	for name := range patterns {
		names = append(names, name)
	}
	sort.Strings(names)

	// Print each pattern
	for _, name := range names {
		p := patterns[name]
		fmt.Printf("  %-20s %s\n", name, p.Description)
	}

	fmt.Println("")
	fmt.Println("Usage: gux help <pattern>")
	fmt.Println("Example: gux help page > pages/mypage.go")
}

// runHelpPattern handles the `gux help <pattern>` command.
// If name is empty, lists all patterns.
// If pattern exists, renders it with module path substitution.
// If pattern not found, prints error and lists available patterns.
func runHelpPattern(name string) {
	if name == "" {
		printPatternList()
		return
	}

	pattern, exists := patterns[name]
	if !exists {
		fmt.Fprintf(os.Stderr, "Unknown pattern: %s\n\n", name)
		printPatternList()
		os.Exit(1)
	}

	// Get module path for template substitution
	modulePath := getModulePathForHelp()

	// Template data
	data := struct {
		ModulePath string
	}{
		ModulePath: modulePath,
	}

	// Parse and execute template
	tmpl, err := template.New(name).Parse(pattern.Template)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing pattern template: %v\n", err)
		os.Exit(1)
	}

	// Output to stdout for piping
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing pattern template: %v\n", err)
		os.Exit(1)
	}
}
