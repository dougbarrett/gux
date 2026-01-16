package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func runClaude() {
	// Create .claude/skills directory
	skillsDir := filepath.Join(".claude", "skills")
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

	fmt.Println("Claude Code skill installed successfully!")
	fmt.Println()
	fmt.Printf("  created %s\n", skillPath)
	fmt.Println()
	fmt.Println("The gux-framework skill provides Claude with comprehensive")
	fmt.Println("knowledge of the Gux framework including:")
	fmt.Println("  - CLI commands and project scaffolding")
	fmt.Println("  - API code generation")
	fmt.Println("  - Component library (45+ components)")
	fmt.Println("  - State management patterns")
	fmt.Println("  - Server utilities and deployment")
	fmt.Println()
	fmt.Println("Claude Code will automatically use this knowledge when")
	fmt.Println("helping you develop your Gux application.")
}
