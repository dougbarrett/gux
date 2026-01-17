package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

func runSetup(tinygo bool) {
	// Get the source path for wasm_exec.js
	var srcPath string
	var err error

	if tinygo {
		// TinyGo: use tinygo env TINYGOROOT
		cmd := exec.Command("tinygo", "env", "TINYGOROOT")
		out, err := cmd.Output()
		if err != nil {
			fmt.Println("Error: TinyGo not found. Install TinyGo or use 'gux setup' without --tinygo.")
			os.Exit(1)
		}
		tinygoRoot := string(out[:len(out)-1]) // trim newline
		srcPath = filepath.Join(tinygoRoot, "targets", "wasm_exec.js")
	} else {
		// Standard Go: use go env GOROOT
		cmd := exec.Command("go", "env", "GOROOT")
		out, err := cmd.Output()
		if err != nil {
			fmt.Println("Error: Go not found")
			os.Exit(1)
		}
		goRoot := string(out[:len(out)-1]) // trim newline
		srcPath = filepath.Join(goRoot, "lib", "wasm", "wasm_exec.js")
	}

	// Check source exists
	if _, err = os.Stat(srcPath); os.IsNotExist(err) {
		fmt.Printf("Error: wasm_exec.js not found at %s\n", srcPath)
		os.Exit(1)
	}

	// Ensure public directory exists
	if err = os.MkdirAll("public", 0755); err != nil {
		fmt.Printf("Error creating public directory: %v\n", err)
		os.Exit(1)
	}

	// Copy the file
	src, err := os.Open(srcPath)
	if err != nil {
		fmt.Printf("Error opening source: %v\n", err)
		os.Exit(1)
	}
	defer src.Close()

	dst, err := os.Create("public/wasm_exec.js")
	if err != nil {
		fmt.Printf("Error creating destination: %v\n", err)
		os.Exit(1)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		fmt.Printf("Error copying file: %v\n", err)
		os.Exit(1)
	}

	compiler := "Go"
	if tinygo {
		compiler = "TinyGo"
	}
	fmt.Printf("Copied wasm_exec.js to public/ from %s installation\n", compiler)
}

// buildWasm builds the WASM module only (used by dev mode)
func buildWasm(tinygo bool) {
	// Check we're in a gux project (has cmd/app/ directory)
	if _, err := os.Stat("cmd/app"); os.IsNotExist(err) {
		fmt.Println("Error: no cmd/app/ directory found")
		fmt.Println("Run this command from your gux project root.")
		os.Exit(1)
	}

	fmt.Println("Building WASM module...")

	var cmd *exec.Cmd
	if tinygo {
		// TinyGo build (smaller output ~500KB)
		cmd = exec.Command("tinygo", "build", "-o", "public/main.wasm", "-target", "wasm", "-no-debug", "./cmd/app")
	} else {
		// Standard Go build (~5MB)
		cmd = exec.Command("go", "build", "-o", "public/main.wasm", "./cmd/app")
		cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("WASM build failed: %v\n", err)
		os.Exit(1)
	}

	// Get WASM file size for display
	wasmInfo, err := os.Stat("public/main.wasm")
	if err != nil {
		fmt.Printf("Error reading public/main.wasm: %v\n", err)
		os.Exit(1)
	}

	// Clean up any old versioned files (from previous gux versions)
	cleanOldWasmFiles("")

	wasmSize := float64(wasmInfo.Size()) / 1024 / 1024
	compiler := "Go"
	if tinygo {
		compiler = "TinyGo"
	}
	fmt.Printf("Built public/main.wasm (%.2f MB) with %s\n", wasmSize, compiler)
}

// runBuild builds the WASM and then the server binary with all assets embedded
func runBuild(tinygo bool) {
	// Check for wasm_exec.js
	if _, err := os.Stat("public/wasm_exec.js"); os.IsNotExist(err) {
		fmt.Println("Error: public/wasm_exec.js not found")
		fmt.Println("Run 'gux setup' first to copy wasm_exec.js from your Go/TinyGo installation.")
		os.Exit(1)
	}

	// Build the WASM first
	buildWasm(tinygo)

	// Copy public/ to cmd/server/public/ for embedding
	// (go:embed paths are relative to the source file)
	fmt.Println("Building server binary with embedded assets...")

	serverPublic := filepath.Join("cmd", "server", "public")
	if err := copyDir("public", serverPublic); err != nil {
		fmt.Printf("Error copying public directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(serverPublic) // Clean up after build

	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", "server", "./cmd/server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Set CGO_ENABLED=0 for static linking (works on Alpine/musl-based images)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	if err := cmd.Run(); err != nil {
		fmt.Printf("Server build failed: %v\n", err)
		os.Exit(1)
	}

	// Get server binary size for display
	serverInfo, err := os.Stat("server")
	if err != nil {
		fmt.Printf("Error reading server binary: %v\n", err)
		os.Exit(1)
	}

	serverSize := float64(serverInfo.Size()) / 1024 / 1024
	fmt.Printf("Built ./server (%.2f MB) with all assets embedded\n", serverSize)
	fmt.Println("\nRun with: ./server")
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

// cleanOldWasmFiles removes old versioned wasm files from previous gux versions.
// These are files matching main.<hash>.wasm pattern.
func cleanOldWasmFiles(keepFile string) {
	entries, err := os.ReadDir("public")
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		// Match main.<hash>.wasm pattern (not main.wasm)
		// Versioned files have format: main.XXXXXXXX.wasm (8 char hash)
		if name != "main.wasm" && strings.HasPrefix(name, "main.") && strings.HasSuffix(name, ".wasm") && name != keepFile {
			os.Remove(filepath.Join("public", name))
		}
	}
}

func runDev(port int, tinygo bool) {
	// Check for wasm_exec.js
	if _, err := os.Stat("public/wasm_exec.js"); os.IsNotExist(err) {
		fmt.Println("Error: public/wasm_exec.js not found")
		fmt.Println("Run 'gux setup' first to copy wasm_exec.js from your Go installation.")
		os.Exit(1)
	}

	// Build WASM only (not the full binary - we'll use go run for dev)
	buildWasm(tinygo)

	// Check if cmd/server/ exists
	serverDir := "cmd/server"
	if _, err := os.Stat(serverDir); os.IsNotExist(err) {
		fmt.Println("Error: no cmd/server/ directory found")
		os.Exit(1)
	}

	// Copy public/ to cmd/server/public/ for go:embed to compile
	// (even though dev mode uses -dir flag, the embed directive must resolve)
	serverPublic := filepath.Join("cmd", "server", "public")
	if err := copyDir("public", serverPublic); err != nil {
		fmt.Printf("Error copying public directory: %v\n", err)
		os.Exit(1)
	}

	// Cleanup function for dev artifacts
	cleanup := func() {
		os.RemoveAll(serverPublic)
		os.Remove("public/main.wasm")
	}

	// Handle Ctrl+C and other termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("\nStarting dev server on http://localhost:%d\n", port)

	// Run the server with -dir flag (serves from filesystem for hot reload)
	cmd := exec.Command("go", "run", "./cmd/server", "-port", fmt.Sprintf("%d", port), "-dir", "public")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(".")

	// Start server in background so we can handle signals
	if err := cmd.Start(); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
		cleanup()
		os.Exit(1)
	}

	// Wait for either server to exit or signal
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-sigChan:
		fmt.Println("\nShutting down...")
		cmd.Process.Signal(os.Interrupt)
		<-done // Wait for process to exit
		cleanup()
	case err := <-done:
		cleanup()
		if err != nil {
			fmt.Printf("Server exited with error: %v\n", err)
			os.Exit(1)
		}
	}
}
