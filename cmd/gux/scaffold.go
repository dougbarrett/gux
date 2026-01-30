package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*
var templates embed.FS

// AuthMode represents the authentication scaffolding mode
type AuthMode int

const (
	AuthModeNone    AuthMode = iota // No authentication
	AuthModePrivate                 // Auth with admin-only (no public signup)
	AuthModePublic                  // Auth with public signup enabled
)

// TemplateData holds the variables for template substitution
type TemplateData struct {
	AppName      string
	ModulePath   string
	GuxModule    string
	GuxVersion   string
	WithAuth     bool   // true if --auth or --auth-public
	PublicSignup bool   // true if --auth-public (allows public registration)
	WithAdmin    bool   // true if --admin (admin panel with sidebar layout)
	Port         string // Server port (default: 8080)
}

// GuxConfig is the configuration file format for gux.config.json
type GuxConfig struct {
	Module string                     `json:"module"`           // Go module path
	Auth   string                     `json:"auth,omitempty"`   // "none", "private", "public"
	Admin  bool                       `json:"admin,omitempty"`  // Include admin panel
	Claude bool                       `json:"claude,omitempty"` // Include Claude Code integration
	Port   string                     `json:"port,omitempty"`   // Server port
	Roles  []SelectOption             `json:"roles,omitempty"`  // Available user roles (used by auth preset)
	Models map[string]ModelDefinition `json:"models,omitempty"` // Model definitions for scaffolding
}

const configFileName = "gux.config.json"

// SaveConfig writes the configuration to gux.config.json
func SaveConfig(targetDir string, modulePath string, authMode AuthMode, withAdmin bool, withClaude bool, port string, roles []SelectOption, models map[string]ModelDefinition) error {
	config := GuxConfig{
		Module: modulePath,
		Admin:  withAdmin,
		Claude: withClaude,
		Port:   port,
		Roles:  roles,
		Models: models,
	}

	switch authMode {
	case AuthModeNone:
		config.Auth = "none"
	case AuthModePrivate:
		config.Auth = "private"
	case AuthModePublic:
		config.Auth = "public"
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	configPath := filepath.Join(targetDir, configFileName)
	return os.WriteFile(configPath, data, 0644)
}

// LoadConfig reads the configuration from gux.config.json
func LoadConfig(dir string) (*GuxConfig, error) {
	configPath := filepath.Join(dir, configFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config GuxConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ConfigExists checks if gux.config.json exists in the given directory
func ConfigExists(dir string) bool {
	configPath := filepath.Join(dir, configFileName)
	_, err := os.Stat(configPath)
	return err == nil
}

// ParseAuthMode converts a string auth mode to AuthMode
func ParseAuthMode(auth string) AuthMode {
	switch auth {
	case "private":
		return AuthModePrivate
	case "public":
		return AuthModePublic
	default:
		return AuthModeNone
	}
}

// authModeToString converts AuthMode to string for config storage
func authModeToString(mode AuthMode) string {
	switch mode {
	case AuthModePrivate:
		return "private"
	case AuthModePublic:
		return "public"
	default:
		return "none"
	}
}

func runInit(appName, modulePath string, authMode AuthMode, withAdmin bool, withClaude bool, port string, roles []SelectOption, models map[string]ModelDefinition) {
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

		// Check if directory has conflicting files (skip when re-init from config)
		if !ConfigExists(targetDir) {
			conflicts := checkForConflicts(targetDir)
			if len(conflicts) > 0 {
				fmt.Println("Error: directory contains files that would be overwritten:")
				for _, f := range conflicts {
					fmt.Printf("  - %s\n", f)
				}
				os.Exit(1)
			}
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

	// Default port if not specified
	if port == "" {
		port = "8080"
	}

	data := TemplateData{
		AppName:      appName,
		ModulePath:   modulePath,
		GuxModule:    "github.com/dougbarrett/gux",
		GuxVersion:   guxVersion,
		WithAuth:     authMode != AuthModeNone,
		PublicSignup: authMode == AuthModePublic,
		WithAdmin:    withAdmin,
		Port:         port,
	}

	// Define files to create from templates
	// Note: public/ files (index.html, manifest.json, service-worker.js) are NOT created here.
	// They are embedded in gux and injected at build time. Users can override by creating
	// their own public/ directory with custom files.
	filesToCreate := []struct {
		tmplPath string
		destPath string
	}{
		{"templates/go.mod.tmpl", "go.mod"},
		{"templates/Dockerfile.tmpl", "Dockerfile"},
	}

	// Add app.go based on auth and admin mode
	if withAdmin && authMode != AuthModeNone {
		// Auth + Admin combined
		filesToCreate = append(filesToCreate, struct {
			tmplPath string
			destPath string
		}{"templates/admin/app_auth.go.tmpl", "app.go"})
	} else if withAdmin {
		// Admin only (no auth)
		filesToCreate = append(filesToCreate, struct {
			tmplPath string
			destPath string
		}{"templates/admin/app.go.tmpl", "app.go"})
	} else if authMode != AuthModeNone {
		// Auth only (no admin)
		filesToCreate = append(filesToCreate, struct {
			tmplPath string
			destPath string
		}{"templates/auth/app.go.tmpl", "app.go"})
	} else {
		// Basic app (no auth, no admin)
		filesToCreate = append(filesToCreate, struct {
			tmplPath string
			destPath string
		}{"templates/app.go.tmpl", "app.go"})
	}

	// Add standard files (non-auth, non-admin)
	if authMode == AuthModeNone && !withAdmin {
		filesToCreate = append(filesToCreate,
			struct {
				tmplPath string
				destPath string
			}{"templates/models/item.go.tmpl", "models/item.go"},
			struct {
				tmplPath string
				destPath string
			}{"templates/pages/home.go.tmpl", "pages/home.go"},
			struct {
				tmplPath string
				destPath string
			}{"templates/pages/items.go.tmpl", "pages/items.go"},
			struct {
				tmplPath string
				destPath string
			}{"templates/pages/item_new.go.tmpl", "pages/item_new.go"},
		)
	}

	// Add admin-specific files (with or without auth)
	if withAdmin {
		filesToCreate = append(filesToCreate,
			struct {
				tmplPath string
				destPath string
			}{"templates/admin/admin/layout.go.tmpl", "guxgen/admin/layout.go"},
			struct {
				tmplPath string
				destPath string
			}{"templates/admin/admin/dashboard.go.tmpl", "guxgen/admin/dashboard.go"},
			struct {
				tmplPath string
				destPath string
			}{"templates/admin/admin/settings.go.tmpl", "guxgen/admin/settings.go"},
		)

		// Add user admin pages only if auth is NOT enabled (no User model to scaffold)
		if authMode == AuthModeNone {
			filesToCreate = append(filesToCreate,
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
				}{"templates/admin/dto/breadcrumb.go.tmpl", "dto/breadcrumb.go"},
			)
		}
	}

	// Add auth-specific files
	// Note: models/user.go and dto/user.go are now generated via scaffolding with auth preset
	if authMode != AuthModeNone {
		filesToCreate = append(filesToCreate,
			struct {
				tmplPath string
				destPath string
			}{"templates/auth/env.example.tmpl", ".env.example"},
		)

		// Add pages based on whether admin is also enabled
		if withAdmin {
			// Auth + Admin: login page standalone (no sidebar), home redirect
			filesToCreate = append(filesToCreate,
				struct {
					tmplPath string
					destPath string
				}{"templates/admin/pages/home_redirect.go.tmpl", "pages/home_redirect.go"},
				struct {
					tmplPath string
					destPath string
				}{"templates/admin/pages/login.go.tmpl", "pages/login.go"},
				struct {
					tmplPath string
					destPath string
				}{"templates/admin/dto/auth.go.tmpl", "dto/auth.go"},
			)
		} else {
			// Auth only: use standard auth pages
			filesToCreate = append(filesToCreate,
				struct {
					tmplPath string
					destPath string
				}{"templates/auth/pages/layout.go.tmpl", "pages/layout.go"},
				struct {
					tmplPath string
					destPath string
				}{"templates/auth/pages/home.go.tmpl", "pages/home.go"},
				struct {
					tmplPath string
					destPath string
				}{"templates/auth/pages/login.go.tmpl", "pages/login.go"},
				struct {
					tmplPath string
					destPath string
				}{"templates/auth/pages/dashboard.go.tmpl", "pages/dashboard.go"},
			)
		}

		// Add register page only for public signup
		if authMode == AuthModePublic {
			if withAdmin {
				filesToCreate = append(filesToCreate, struct {
					tmplPath string
					destPath string
				}{"templates/admin/pages/register.go.tmpl", "pages/register.go"})
			} else {
				filesToCreate = append(filesToCreate, struct {
					tmplPath string
					destPath string
				}{"templates/auth/pages/register.go.tmpl", "pages/register.go"})
			}
		}
	}

	fmt.Printf("Creating Gux application '%s'...\n\n", appName)

	// When re-initializing from config, skip files that already exist
	reinit := initHere && ConfigExists(targetDir)
	for _, f := range filesToCreate {
		destFullPath := filepath.Join(targetDir, f.destPath)
		if reinit {
			if _, err := os.Stat(destFullPath); err == nil {
				fmt.Printf("  skipped %s (already exists)\n", f.destPath)
				continue
			}
		}
		if err := renderTemplate(targetDir, f.tmplPath, f.destPath, data); err != nil {
			fmt.Printf("Error creating %s: %v\n", f.destPath, err)
			os.Exit(1)
		}
		fmt.Printf("  created %s\n", f.destPath)
	}

	// Create or update .gitignore
	if err := updateGitignore(targetDir, authMode != AuthModeNone); err != nil {
		fmt.Printf("Warning: could not update .gitignore: %v\n", err)
	} else {
		fmt.Println("  updated .gitignore")
	}

	// Generate User model via scaffolding if auth is enabled
	if authMode != AuthModeNone && models != nil {
		if userModel, ok := models["User"]; ok {
			fmt.Println("\nGenerating User model via scaffolding...")
			// Change to target directory for scaffolding
			originalDir, _ := os.Getwd()
			if targetDir != "." {
				os.Chdir(targetDir)
			}
			userModel.Name = "User"
			if err := GenerateModelFilesImpl(&userModel, nil, modulePath, withAdmin, nil, nil, nil); err != nil {
				fmt.Printf("Warning: could not scaffold User model: %v\n", err)
			}
			if targetDir != "." {
				os.Chdir(originalDir)
			}
		}
	}

	// Run gux gen to generate API client (must run before go mod tidy)
	fmt.Println("\nRunning gux gen...")
	genCmd := exec.Command("gux", "gen")
	genCmd.Dir = targetDir
	genCmd.Stdout = os.Stdout
	genCmd.Stderr = os.Stderr
	if err := genCmd.Run(); err != nil {
		fmt.Printf("Warning: gux gen failed: %v\n", err)
		fmt.Println("You may need to run 'gux gen' manually.")
	} else {
		fmt.Println("  API client generated")
	}

	// Run go mod tidy to download dependencies (after gux gen creates guxgen/)
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

	// Save configuration for future use
	if err := SaveConfig(targetDir, modulePath, authMode, withAdmin, withClaude, port, roles, models); err != nil {
		fmt.Printf("Warning: could not save config: %v\n", err)
	} else {
		fmt.Println("  saved gux.config.json")
	}

	// Install Claude Code integration if requested
	if withClaude {
		fmt.Println("\nInstalling Claude Code integration...")
		installClaudeSkill(targetDir)
		createClaudeMDFromTemplate(targetDir, &GuxConfig{
			Module: modulePath,
			Auth:   authModeToString(authMode),
			Admin:  withAdmin,
			Claude: withClaude,
			Port:   port,
		})
	}

	printNextStepsWithDir(appName, initHere, authMode)

	// Check for updates
	checkForUpdates()
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
		"app.go",
		"models/item.go",
		"pages/home.go",
		"pages/items.go",
		"pages/item_new.go",
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

// updateGitignore creates or updates .gitignore with gux-generated file entries
func updateGitignore(targetDir string, withAuth bool) error {
	gitignorePath := filepath.Join(targetDir, ".gitignore")
	guxEntries := []string{
		"# Gux generated files",
		"guxgen/",
		"assets_gen.go",
	}

	// Add .env to gitignore if auth is enabled
	if withAuth {
		guxEntries = append(guxEntries, "", "# Environment (contains secrets)", ".env")
	}

	// Check if .gitignore exists
	existingContent := ""
	if data, err := os.ReadFile(gitignorePath); err == nil {
		existingContent = string(data)
	}

	// Check which entries are already present
	var entriesToAdd []string
	for _, entry := range guxEntries {
		// Skip comment lines when checking for duplicates
		if strings.HasPrefix(entry, "#") {
			continue
		}
		if !strings.Contains(existingContent, entry) {
			entriesToAdd = append(entriesToAdd, entry)
		}
	}

	// If all entries already exist, nothing to do
	if len(entriesToAdd) == 0 {
		return nil
	}

	// Open file for appending (or create if doesn't exist)
	file, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open .gitignore: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Add newline if file has content and doesn't end with newline
	if existingContent != "" && !strings.HasSuffix(existingContent, "\n") {
		writer.WriteString("\n")
	}

	// Add blank line before gux section if file has content
	if existingContent != "" {
		writer.WriteString("\n")
	}

	// Write gux entries
	for _, entry := range guxEntries {
		writer.WriteString(entry + "\n")
	}

	return writer.Flush()
}

func printNextStepsWithDir(appName string, initHere bool, authMode AuthMode) {
	var cdCmd string
	if !initHere {
		cdCmd = fmt.Sprintf("  cd %s\n", appName)
	}

	var authNote string
	if authMode != AuthModeNone {
		authNote = `
Authentication is enabled. Before running:
  1. Copy .env.example to .env
  2. Set ADMIN_EMAIL and ADMIN_PASSWORD in .env
  3. The admin user will be created on first run
`
		if authMode == AuthModePublic {
			authNote += "\nPublic signup is enabled at /register\n"
		} else {
			authNote += "\nPublic signup is disabled. Only the seeded admin can log in.\n"
		}
	}

	location := "current directory"
	if !initHere {
		location = fmt.Sprintf("./%s", appName)
	}

	fmt.Printf(`
Created Gux application in %s
%s
Next steps:
%s  gux dev         # Build and run dev server

Your app will be available at http://localhost:8080

To customize the HTML shell, create public/index.html (it will override the default).
`, location, authNote, cdCmd)
}
