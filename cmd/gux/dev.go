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

	"github.com/dougbarrett/gux/cmd/gux/defaults"
)

// preparePublicDir creates cmd/server/public/ with default files, wasm_exec.js,
// and any user overrides from public/. Returns the path for cleanup.
func preparePublicDir(tinygo bool) (string, error) {
	serverPublic := filepath.Join("cmd", "server", "public")

	// Create the directory
	if err := os.MkdirAll(serverPublic, 0755); err != nil {
		return "", fmt.Errorf("create server public dir: %w", err)
	}

	// Write embedded default files
	defaultFiles := []string{"index.html", "manifest.json", "service-worker.js"}
	for _, name := range defaultFiles {
		content, err := defaults.Files.ReadFile(name)
		if err != nil {
			return "", fmt.Errorf("read embedded %s: %w", name, err)
		}
		destPath := filepath.Join(serverPublic, name)
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return "", fmt.Errorf("write %s: %w", name, err)
		}
	}

	// Copy wasm_exec.js from TinyGo or Go installation
	var wasmExecSrc string
	if tinygo {
		cmd := exec.Command("tinygo", "env", "TINYGOROOT")
		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("TinyGo not found - install TinyGo or use --go flag: %w", err)
		}
		tinygoRoot := strings.TrimSpace(string(out))
		wasmExecSrc = filepath.Join(tinygoRoot, "targets", "wasm_exec.js")
	} else {
		cmd := exec.Command("go", "env", "GOROOT")
		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("Go not found: %w", err)
		}
		goRoot := strings.TrimSpace(string(out))
		wasmExecSrc = filepath.Join(goRoot, "lib", "wasm", "wasm_exec.js")
	}

	// Check wasm_exec.js exists
	if _, err := os.Stat(wasmExecSrc); os.IsNotExist(err) {
		return "", fmt.Errorf("wasm_exec.js not found at %s", wasmExecSrc)
	}

	// Copy wasm_exec.js
	wasmExecContent, err := os.ReadFile(wasmExecSrc)
	if err != nil {
		return "", fmt.Errorf("read wasm_exec.js: %w", err)
	}
	if err := os.WriteFile(filepath.Join(serverPublic, "wasm_exec.js"), wasmExecContent, 0644); err != nil {
		return "", fmt.Errorf("write wasm_exec.js: %w", err)
	}

	// If user has a public/ directory, overlay its contents (allowing overrides)
	if _, err := os.Stat("public"); err == nil {
		if err := copyDir("public", serverPublic); err != nil {
			return "", fmt.Errorf("copy user public/: %w", err)
		}
	}

	return serverPublic, nil
}

// buildWasm builds the WASM module to the specified output directory
func buildWasm(tinygo bool, outputDir string) {
	// Check we're in a gux project (has cmd/app/ directory)
	if _, err := os.Stat("cmd/app"); os.IsNotExist(err) {
		fmt.Println("Error: no cmd/app/ directory found")
		fmt.Println("Run this command from your gux project root.")
		os.Exit(1)
	}

	fmt.Println("Building WASM module...")

	wasmOutput := filepath.Join(outputDir, "main.wasm")

	var cmd *exec.Cmd
	if tinygo {
		// TinyGo build (smaller output ~500KB)
		cmd = exec.Command("tinygo", "build", "-o", wasmOutput, "-target", "wasm", "-no-debug", "./cmd/app")
	} else {
		// Standard Go build (~5MB)
		cmd = exec.Command("go", "build", "-o", wasmOutput, "./cmd/app")
		cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("WASM build failed: %v\n", err)
		os.Exit(1)
	}

	// Get WASM file size for display
	wasmInfo, err := os.Stat(wasmOutput)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", wasmOutput, err)
		os.Exit(1)
	}

	// Clean up any old versioned files (from previous gux versions)
	cleanOldWasmFiles(outputDir)

	wasmSize := float64(wasmInfo.Size()) / 1024 / 1024
	compiler := "Go"
	if tinygo {
		compiler = "TinyGo"
	}
	fmt.Printf("Built %s (%.2f MB) with %s\n", wasmOutput, wasmSize, compiler)
}

// runBuild builds the WASM and then the server binary with all assets embedded
func runBuild(tinygo bool) {
	// Prepare public directory with defaults, wasm_exec.js, and user overrides
	serverPublic, err := preparePublicDir(tinygo)
	if err != nil {
		fmt.Printf("Error preparing public directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(serverPublic) // Clean up after build

	// Build the WASM into the prepared directory
	buildWasm(tinygo, serverPublic)

	// Build server binary with embedded assets
	fmt.Println("Building server binary with embedded assets...")

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
// These are files matching main.<hash>.wasm pattern in the specified directory.
func cleanOldWasmFiles(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		// Match main.<hash>.wasm pattern (not main.wasm)
		// Versioned files have format: main.XXXXXXXX.wasm (8 char hash)
		if name != "main.wasm" && strings.HasPrefix(name, "main.") && strings.HasSuffix(name, ".wasm") {
			os.Remove(filepath.Join(dir, name))
		}
	}
}

func runDev(port int, tinygo bool) {
	// Check if cmd/server/ exists
	serverDir := "cmd/server"
	if _, err := os.Stat(serverDir); os.IsNotExist(err) {
		fmt.Println("Error: no cmd/server/ directory found")
		os.Exit(1)
	}

	// Prepare public directory with defaults, wasm_exec.js, and user overrides
	serverPublic, err := preparePublicDir(tinygo)
	if err != nil {
		fmt.Printf("Error preparing public directory: %v\n", err)
		os.Exit(1)
	}

	// Build WASM into the prepared directory
	buildWasm(tinygo, serverPublic)

	// Cleanup function for dev artifacts
	cleanup := func() {
		os.RemoveAll(serverPublic)
	}

	// Handle Ctrl+C and other termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("\nStarting dev server on http://localhost:%d\n", port)

	// Run the server with -dir flag (serves from filesystem for hot reload)
	cmd := exec.Command("go", "run", "./cmd/server", "-port", fmt.Sprintf("%d", port), "-dir", serverPublic)
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
