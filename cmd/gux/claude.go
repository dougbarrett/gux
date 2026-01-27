package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func runClaude() {
	// Install skill file
	installClaudeSkill(".")

	// Handle CLAUDE.md
	handleClaudeMD(".")

	fmt.Println()
	fmt.Println("Claude Code will automatically use this knowledge when")
	fmt.Println("helping you develop your Gux application.")
}

// installClaudeSkill installs the gux-framework.md skill file
func installClaudeSkill(targetDir string) {
	// Create .claude/skills directory
	skillsDir := filepath.Join(targetDir, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Read embedded skill file
	skillContent, err := templates.ReadFile("templates/claude/skills/gux-framework.md")
	if err != nil {
		fmt.Printf("Error reading skill template: %v\n", err)
		os.Exit(1)
	}

	// Write skill file
	skillPath := filepath.Join(skillsDir, "gux-framework.md")
	if err := os.WriteFile(skillPath, skillContent, 0644); err != nil {
		fmt.Printf("Error writing skill file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Claude Code integration installed:")
	fmt.Printf("  created %s\n", skillPath)
}

// handleClaudeMD creates or updates CLAUDE.md in the target directory
func handleClaudeMD(targetDir string) {
	claudeMDPath := filepath.Join(targetDir, "CLAUDE.md")

	// Check if CLAUDE.md already exists
	if _, err := os.Stat(claudeMDPath); err == nil {
		// File exists - append note about skills if not already present
		content, err := os.ReadFile(claudeMDPath)
		if err != nil {
			fmt.Printf("Warning: could not read CLAUDE.md: %v\n", err)
			return
		}

		// Check if already has our skills note
		if !strings.Contains(string(content), "## Claude Code Skills") {
			// Append skills note
			skillsNote := `

## Claude Code Skills

This project includes the Gux framework skill for Claude Code.
Run ` + "`gux claude`" + ` to update the skill file if needed.
`
			if err := os.WriteFile(claudeMDPath, append(content, []byte(skillsNote)...), 0644); err != nil {
				fmt.Printf("Warning: could not update CLAUDE.md: %v\n", err)
				return
			}
			fmt.Printf("  updated %s\n", claudeMDPath)
		}
		return
	}

	// No CLAUDE.md exists - create from template if config exists
	config, err := LoadConfig(targetDir)
	if err != nil {
		// No config file - create a basic CLAUDE.md
		createBasicClaudeMD(targetDir)
		return
	}

	// Create CLAUDE.md from template with config data
	createClaudeMDFromTemplate(targetDir, config)
}

// createBasicClaudeMD creates a minimal CLAUDE.md for projects without config
func createBasicClaudeMD(targetDir string) {
	// Get app name from directory
	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		absPath = targetDir
	}
	appName := filepath.Base(absPath)
	if appName == "." {
		cwd, err := os.Getwd()
		if err == nil {
			appName = filepath.Base(cwd)
		} else {
			appName = "myapp"
		}
	}

	content := fmt.Sprintf(`# %s

A Gux application built with Go and WebAssembly.

## Quick Start

`+"```bash"+`
gux dev          # Run development server
gux build        # Production build
gux gen          # Regenerate API client
gux clean        # Remove generated files
`+"```"+`

## Development Notes

- **Always use `+"`gux dev`"+`** (not `+"`go run`"+`) - it builds WASM and CSS
- Use `+"`r.OnLoad()`"+` for data fetching (supports SSR hydration)
- DTOs hide sensitive fields from API responses
- Generated files in `+"`guxgen/`"+` should not be edited

## Claude Code Skills

This project includes the Gux framework skill for Claude Code.
Run `+"`gux claude`"+` to update the skill file if needed.
`, appName)

	claudeMDPath := filepath.Join(targetDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMDPath, []byte(content), 0644); err != nil {
		fmt.Printf("Warning: could not create CLAUDE.md: %v\n", err)
		return
	}
	fmt.Printf("  created %s\n", claudeMDPath)
}

// createClaudeMDFromTemplate creates CLAUDE.md using the template and config
func createClaudeMDFromTemplate(targetDir string, config *GuxConfig) {
	// Read template
	tmplContent, err := templates.ReadFile("templates/CLAUDE.md.tmpl")
	if err != nil {
		// Fall back to basic if template not found
		createBasicClaudeMD(targetDir)
		return
	}

	// Get app name from directory
	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		absPath = targetDir
	}
	appName := filepath.Base(absPath)
	if appName == "." {
		cwd, err := os.Getwd()
		if err == nil {
			appName = filepath.Base(cwd)
		} else {
			appName = "myapp"
		}
	}

	// Prepare template data
	port := config.Port
	if port == "" {
		port = "8080"
	}

	data := TemplateData{
		AppName:      appName,
		ModulePath:   config.Module,
		WithAuth:     config.Auth != "" && config.Auth != "none",
		PublicSignup: config.Auth == "public",
		WithAdmin:    config.Admin,
		Port:         port,
	}

	// Parse and execute template
	tmpl, err := template.New("CLAUDE.md").Parse(string(tmplContent))
	if err != nil {
		createBasicClaudeMD(targetDir)
		return
	}

	claudeMDPath := filepath.Join(targetDir, "CLAUDE.md")
	file, err := os.Create(claudeMDPath)
	if err != nil {
		fmt.Printf("Warning: could not create CLAUDE.md: %v\n", err)
		return
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		fmt.Printf("Warning: could not write CLAUDE.md: %v\n", err)
		return
	}

	fmt.Printf("  created %s\n", claudeMDPath)
}
