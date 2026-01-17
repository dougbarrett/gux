package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
