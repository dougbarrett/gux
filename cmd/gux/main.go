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
		initCmd.Parse(os.Args[2:])

		if initCmd.NArg() < 1 {
			fmt.Println("Error: app name required (use '.' for current directory)")
			fmt.Println("Usage: gux init [--module <module-path>] <appname>")
			fmt.Println("       gux init --module <module-path> .")
			os.Exit(1)
		}

		appName := initCmd.Arg(0)
		runInit(appName, *modulePath)

	case "gen", "generate":
		genCmd := flag.NewFlagSet("gen", flag.ExitOnError)
		apiDir := genCmd.String("dir", "internal/api", "Directory containing API interface files")
		genCmd.Parse(os.Args[2:])

		runGenerate(*apiDir)

	case "build":
		buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
		useGo := buildCmd.Bool("go", false, "Use standard Go instead of TinyGo (~5MB vs ~500KB)")
		buildCmd.Parse(os.Args[2:])

		runBuild(!*useGo) // TinyGo is default

	case "dev":
		devCmd := flag.NewFlagSet("dev", flag.ExitOnError)
		port := devCmd.Int("port", 8080, "Port to run dev server on")
		useGo := devCmd.Bool("go", false, "Use standard Go instead of TinyGo")
		devCmd.Parse(os.Args[2:])

		runDev(*port, !*useGo) // TinyGo is default

	case "setup":
		setupCmd := flag.NewFlagSet("setup", flag.ExitOnError)
		useGo := setupCmd.Bool("go", false, "Copy wasm_exec.js from standard Go instead of TinyGo")
		setupCmd.Parse(os.Args[2:])

		runSetup(!*useGo) // TinyGo is default

	case "claude":
		runClaude()

	case "version", "-v", "--version":
		fmt.Printf("gux version %s\n", getVersion())

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`gux - Gux application scaffolding tool

Usage:
    gux init [--module <module-path>] <appname>   Create a new Gux application
    gux init --module <module-path> .             Initialize in current directory
    gux setup [--go]                              Copy wasm_exec.js to public/
    gux gen [--dir <api-dir>]                     Generate API client code
    gux build [--go]                              Build WASM and server binary
    gux dev [--port <port>] [--go]                Build and run dev server
    gux claude                                    Install Claude Code skill
    gux version                                   Show version
    gux help                                      Show this help

TinyGo is the default compiler (~500KB WASM). Use --go for standard Go (~5MB).

Examples:
    gux init --module github.com/myuser/myapp myapp   # Create new directory
    gux init --module github.com/myuser/myapp .       # Use current directory
    gux setup                # Copy wasm_exec.js from TinyGo to public/
    gux setup --go           # Copy wasm_exec.js from standard Go to public/
    gux build                # Build with TinyGo (~500KB WASM)
    gux build --go           # Build with standard Go (~5MB WASM)
    gux dev                  # Run dev server on :8080 (TinyGo)
    gux dev --port 3000      # Run on custom port
    gux claude               # Install Claude Code skill for AI assistance

The init command creates a Gux application scaffold including:
    - cmd/app/main.go       - WASM frontend entry point
    - cmd/server/main.go    - HTTP server
    - internal/api/         - Shared API definitions
    - public/               - Static files (index.html, manifest.json, etc.)
    - Dockerfile            - Multi-stage Docker build

After scaffolding, run:
    gux setup     # Copy wasm_exec.js to public/
    gux claude    # Install Claude Code skill (optional)
    gux dev       # Build and run dev server`)
}
