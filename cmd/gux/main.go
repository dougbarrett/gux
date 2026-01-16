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
			fmt.Println("Error: app name required")
			fmt.Println("Usage: gux init [--module <module-path>] <appname>")
			os.Exit(1)
		}

		appName := initCmd.Arg(0)
		runInit(appName, *modulePath)

	case "gen", "generate":
		genCmd := flag.NewFlagSet("gen", flag.ExitOnError)
		apiDir := genCmd.String("dir", "api", "Directory containing API interface files")
		genCmd.Parse(os.Args[2:])

		runGenerate(*apiDir)

	case "build":
		buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
		tinygo := buildCmd.Bool("tinygo", false, "Use TinyGo for smaller output (~500KB vs ~5MB)")
		buildCmd.Parse(os.Args[2:])

		runBuild(*tinygo)

	case "dev":
		devCmd := flag.NewFlagSet("dev", flag.ExitOnError)
		port := devCmd.Int("port", 8080, "Port to run dev server on")
		tinygo := devCmd.Bool("tinygo", false, "Use TinyGo for smaller output")
		devCmd.Parse(os.Args[2:])

		runDev(*port, *tinygo)

	case "setup":
		setupCmd := flag.NewFlagSet("setup", flag.ExitOnError)
		tinygo := setupCmd.Bool("tinygo", false, "Copy wasm_exec.js from TinyGo instead of Go")
		setupCmd.Parse(os.Args[2:])

		runSetup(*tinygo)

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
    gux setup [--tinygo]                          Copy wasm_exec.js from Go/TinyGo
    gux gen [--dir <api-dir>]                     Generate API client code
    gux build [--tinygo]                          Build WASM module
    gux dev [--port <port>] [--tinygo]            Build and run dev server
    gux version                                    Show version
    gux help                                       Show this help

Examples:
    gux init --module github.com/myuser/myapp myapp
    gux setup                # Copy wasm_exec.js from Go
    gux setup --tinygo       # Copy wasm_exec.js from TinyGo
    gux build --tinygo       # Build with TinyGo (~500KB)
    gux dev                  # Run dev server on :8080
    gux dev --port 3000      # Run on custom port

The init command creates a new directory with your app name and generates
a minimal Gux application scaffold including:
    - app/main.go       - WASM frontend entry point
    - server/main.go    - HTTP server
    - api/              - Shared API definitions
    - index.html        - PWA entry point
    - Dockerfile        - Multi-stage Docker build
    - manifest.json     - PWA manifest
    - offline.html      - Offline fallback
    - service-worker.js - PWA caching

After scaffolding, run:
    cd <appname>
    gux setup     # Copy wasm_exec.js from Go installation
    gux dev       # Build and run dev server`)
}
