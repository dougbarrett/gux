package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
)

// InitOptions holds the options gathered from interactive prompts
type InitOptions struct {
	AppName      string
	ModulePath   string
	AuthMode     AuthMode
	WithAdmin    bool
	AdminEmail   string
	AdminPassword string
}

// RunInteractiveInit runs the interactive prompts for gux init
func RunInteractiveInit(appName string, providedModule string) {
	opts := InitOptions{
		AppName: appName,
	}

	// Determine if we're initializing in current directory
	initHere := appName == "."
	if initHere {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		opts.AppName = filepath.Base(cwd)
	}

	// Build form groups
	var groups []*huh.Group

	// Module path input (only if not provided via flag)
	if providedModule == "" {
		defaultModule := fmt.Sprintf("github.com/youruser/%s", opts.AppName)
		if initHere {
			defaultModule = fmt.Sprintf("github.com/youruser/%s", opts.AppName)
		}

		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title("Go module path").
				Description("e.g., github.com/youruser/"+opts.AppName).
				Placeholder(defaultModule).
				Value(&opts.ModulePath).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("module path is required")
					}
					return nil
				}),
		))
	} else {
		opts.ModulePath = providedModule
	}

	// Auth mode selection
	var authChoice string
	groups = append(groups, huh.NewGroup(
		huh.NewSelect[string]().
			Title("Authentication").
			Options(
				huh.NewOption("None - No authentication", "none"),
				huh.NewOption("Admin-only - Login required, no public signup", "admin"),
				huh.NewOption("Public signup - Users can register", "public"),
			).
			Value(&authChoice),
	))

	// Admin panel selection
	groups = append(groups, huh.NewGroup(
		huh.NewConfirm().
			Title("Include admin panel with sidebar layout?").
			Affirmative("Yes").
			Negative("No").
			Value(&opts.WithAdmin),
	))

	// Run the form
	form := huh.NewForm(groups...)
	err := form.Run()
	if err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println("\nAborted.")
			os.Exit(0)
		}
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Convert auth choice to AuthMode
	switch authChoice {
	case "admin":
		opts.AuthMode = AuthModePrivate
	case "public":
		opts.AuthMode = AuthModePublic
	default:
		opts.AuthMode = AuthModeNone
	}

	// If auth is enabled, prompt for credentials
	if opts.AuthMode != AuthModeNone {
		if err := promptForCredentials(&opts); err != nil {
			if err == huh.ErrUserAborted {
				fmt.Println("\nAborted.")
				os.Exit(0)
			}
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Run the init with collected options
	runInitWithEnv(appName, opts)
}

// promptForCredentials prompts for admin credentials
func promptForCredentials(opts *InitOptions) error {
	var passwordChoice string
	var customPassword string

	// Default email
	opts.AdminEmail = "admin@example.com"

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Admin email").
				Value(&opts.AdminEmail).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("email is required")
					}
					if !strings.Contains(s, "@") {
						return fmt.Errorf("invalid email format")
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Admin password").
				Options(
					huh.NewOption("Generate secure password", "generate"),
					huh.NewOption("Enter manually", "manual"),
				).
				Value(&passwordChoice),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if passwordChoice == "generate" {
		opts.AdminPassword = generateSecurePassword(16)
	} else {
		// Manual password entry
		passwordForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter password").
					EchoMode(huh.EchoModePassword).
					Value(&customPassword).
					Validate(func(s string) error {
						if len(s) < 8 {
							return fmt.Errorf("password must be at least 8 characters")
						}
						return nil
					}),
			),
		)
		if err := passwordForm.Run(); err != nil {
			return err
		}
		opts.AdminPassword = customPassword
	}

	return nil
}

// generateSecurePassword generates a secure random password
func generateSecurePassword(length int) string {
	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a reasonable default if crypto/rand fails
		return "ChangeMe123!"
	}

	// Encode to base64 and trim to desired length
	password := base64.URLEncoding.EncodeToString(bytes)
	if len(password) > length {
		password = password[:length]
	}

	return password
}

// runInitWithEnv runs init and creates .env file with credentials
func runInitWithEnv(originalAppName string, opts InitOptions) {
	// Run the standard init
	runInit(originalAppName, opts.ModulePath, opts.AuthMode, opts.WithAdmin)

	// If auth is enabled, create .env file with credentials
	if opts.AuthMode != AuthModeNone {
		targetDir := originalAppName
		if originalAppName == "." {
			targetDir = "."
		}

		envContent := fmt.Sprintf(`# Admin credentials (created by gux init)
ADMIN_EMAIL=%s
ADMIN_PASSWORD=%s

# Server port (optional)
# PORT=8080
`, opts.AdminEmail, opts.AdminPassword)

		envPath := filepath.Join(targetDir, ".env")
		if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
			fmt.Printf("Warning: could not create .env file: %v\n", err)
		} else {
			fmt.Println("  created .env")
			fmt.Printf("\n  Admin credentials:\n")
			fmt.Printf("    Email:    %s\n", opts.AdminEmail)
			fmt.Printf("    Password: %s\n", opts.AdminPassword)
			fmt.Println("\n  Save these credentials - the password won't be shown again!")
		}
	}
}

// ShouldRunInteractive determines if we should run interactive mode
// Returns true if no feature flags were provided
func ShouldRunInteractive(withAuth, withAuthPublic, withAdmin bool) bool {
	return !withAuth && !withAuthPublic && !withAdmin
}
