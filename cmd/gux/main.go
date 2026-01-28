package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
)

// getVersion returns the version from module info (set by go install @vX.Y.Z)
func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return "dev"
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		initCmd := flag.NewFlagSet("init", flag.ExitOnError)
		modulePath := initCmd.String("module", "", "Go module path (e.g., github.com/user/myapp)")
		withAuth := initCmd.Bool("auth", false, "Include authentication (admin-only, no public signup)")
		withAuthPublic := initCmd.Bool("auth-public", false, "Include authentication with public signup")
		withAdmin := initCmd.Bool("admin", false, "Include admin panel with sidebar layout")
		withClaude := initCmd.Bool("claude", false, "Include Claude Code skills and CLAUDE.md")
		port := initCmd.String("port", "8080", "Server port (default: 8080)")
		nonInteractive := initCmd.Bool("yes", false, "Non-interactive mode with defaults (no auth, no admin)")
		fromConfig := initCmd.Bool("config", false, "Use settings from gux.config.json")
		initCmd.Parse(os.Args[2:])

		if initCmd.NArg() < 1 {
			fmt.Println("Error: app name required (use '.' for current directory)")
			fmt.Println("Usage: gux init [--module <module-path>] [--auth|--auth-public] [--admin] <appname>")
			fmt.Println("       gux init --module <module-path> .")
			fmt.Println("       gux init --config .")
			os.Exit(1)
		}

		appName := initCmd.Arg(0)

		// Check for config file if --config flag or initializing in current directory
		if *fromConfig || (appName == "." && ConfigExists(".")) {
			config, err := LoadConfig(".")
			if err != nil {
				if *fromConfig {
					fmt.Printf("Error: could not read gux.config.json: %v\n", err)
					os.Exit(1)
				}
				// Config doesn't exist, continue with normal flow
			} else {
				// Use config values, but allow flags to override
				if *modulePath == "" {
					*modulePath = config.Module
				}
				if *port == "8080" && config.Port != "" {
					*port = config.Port
				}
				if !*withAdmin && config.Admin {
					*withAdmin = true
				}
				if !*withClaude && config.Claude {
					*withClaude = true
				}
				if !*withAuth && !*withAuthPublic {
					switch config.Auth {
					case "private":
						*withAuth = true
					case "public":
						*withAuthPublic = true
					}
				}
				fmt.Println("Using settings from gux.config.json")
				*nonInteractive = true // Skip interactive mode when using config
			}
		}

		// Determine if we should run interactive mode
		// Interactive if: no auth flags, no admin flag, no claude flag, and not --yes
		if ShouldRunInteractive(*withAuth, *withAuthPublic, *withAdmin, *withClaude) && !*nonInteractive {
			RunInteractiveInit(appName, *modulePath, *port)
		} else {
			// Non-interactive mode with explicit flags
			authMode := AuthModeNone
			if *withAuthPublic {
				authMode = AuthModePublic
			} else if *withAuth {
				authMode = AuthModePrivate
			}
			runInit(appName, *modulePath, authMode, *withAdmin, *withClaude, *port)
		}

	case "gen", "generate":
		genCmd := flag.NewFlagSet("gen", flag.ExitOnError)
		watch := genCmd.Bool("watch", false, "Watch for file changes and regenerate")
		genCmd.Parse(os.Args[2:])

		runGenNew(*watch)

	case "pages":
		pagesCmd := flag.NewFlagSet("pages", flag.ExitOnError)
		pagesDir := pagesCmd.String("dir", "pages", "Directory containing page files")
		pagesCmd.Parse(os.Args[2:])

		runPages(*pagesDir)

	case "build":
		buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
		useGo := buildCmd.Bool("go", false, "Use standard Go instead of TinyGo (~5MB vs ~500KB)")
		buildCmd.Parse(os.Args[2:])

		runBuildNew(!*useGo) // TinyGo is default

	case "dev":
		devCmd := flag.NewFlagSet("dev", flag.ExitOnError)
		useGo := devCmd.Bool("go", false, "Use standard Go instead of TinyGo")
		watch := devCmd.Bool("watch", false, "Watch for file changes and hot reload")
		devCmd.Parse(os.Args[2:])

		runDevNew(!*useGo, *watch) // TinyGo is default

	case "clean":
		runClean()

	case "model":
		runModelCommand(os.Args[2:])

	case "claude":
		runClaude()

	case "update":
		updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
		checkOnly := updateCmd.Bool("check", false, "Only check for updates, don't install")
		updateCmd.Parse(os.Args[2:])

		runUpdate(*checkOnly)

	case "version", "-v", "--version":
		fmt.Printf("gux version %s\n", getVersion())

	case "help":
		if len(os.Args) > 2 {
			// gux help <pattern> - show pattern template
			runHelpPattern(os.Args[2])
		} else {
			// gux help - list available patterns
			runHelpPattern("")
		}

	case "-h", "--help":
		// Show general usage
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`gux - Gux application tool

Usage:
    gux init <appname>                            Create a new Gux application (interactive)
    gux init [options] <appname>                  Create with explicit options (non-interactive)
    gux init --module <module-path> .             Initialize in current directory
    gux gen [--watch]                             Generate guxgen files (API, WASM entry points)
    gux model add                                 Interactive model builder
    gux model add --from-config                   Generate models from gux.config.json
    gux model list                                List models in config
    gux build [--go]                              Build WASM and server binary
    gux dev [--go] [--watch]                      Build and run server
    gux clean                                     Remove generated files
    gux claude                                    Install Claude Code skill and CLAUDE.md
    gux update [--check]                          Update gux to latest version
    gux version                                   Show version
    gux help                                      Show this help
    gux help <pattern>                            Show boilerplate for a pattern

Init options:
    --module <path>     Go module path (e.g., github.com/user/myapp)
    --auth              Include authentication (admin-only, no public signup)
    --auth-public       Include authentication with public signup
    --admin             Include admin panel with sidebar layout
    --claude            Include Claude Code skills and CLAUDE.md
    --port <port>       Server port (default: 8080)
    --yes               Non-interactive mode with defaults (no auth, no admin)
    --config            Use settings from gux.config.json (auto-detected with '.')

Interactive mode:
    Running 'gux init myapp' without flags launches an interactive setup wizard
    that guides you through selecting authentication and admin panel options.
    Use --yes or provide explicit flags to skip interactive mode.

TinyGo is the default compiler (~1MB WASM). Use --go for standard Go (~5MB).

Examples:
    gux init myapp                                # Interactive setup wizard
    gux init --yes myapp                          # Quick start with defaults
    gux init --auth --admin --module github.com/x/y app   # Admin with auth
    gux init --claude --module github.com/x/y app         # With Claude integration
    gux init --port 3000 myapp                    # Custom port
    gux init --module github.com/myuser/myapp .           # Current directory
    gux init --config .                           # Re-init using gux.config.json
    gux init .                                    # Auto-detects gux.config.json
    gux gen                  # Generate guxgen files only
    gux gen --watch          # Generate and watch for changes
    gux build                # Build with TinyGo
    gux build --go           # Build with standard Go
    gux dev                  # Build and run server
    gux dev --watch          # Build, run, and hot reload on changes
    gux clean                # Remove bin/, guxgen/, assets_gen.go
    gux claude               # Install Claude Code skill and CLAUDE.md
    gux update               # Update gux to latest release
    gux help page            # Show basic page template
    gux help                 # List all available patterns

Project structure:
    app.go                   - Your application entry point
    pages/                   - Your page components
    guxgen/                    - Generated files (API client, WASM entry points)
    assets_gen.go            - Asset embedding (gitignored)
    bin/                     - Built binary (gitignored)

After scaffolding, run:
    gux dev       # Build and run server
    gux dev --watch  # Build, run, and hot reload on file changes`)
}
