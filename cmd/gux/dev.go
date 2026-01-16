package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func runBuild(tinygo bool) {
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
		fmt.Printf("Build failed: %v\n", err)
		os.Exit(1)
	}

	// Get file size and hash for versioning
	info, err := os.Stat("public/main.wasm")
	if err != nil {
		fmt.Printf("Error reading public/main.wasm: %v\n", err)
		os.Exit(1)
	}

	// Compute content hash for cache busting
	hash, err := hashFile("public/main.wasm")
	if err != nil {
		fmt.Printf("Error hashing public/main.wasm: %v\n", err)
		os.Exit(1)
	}

	// Create versioned filename
	versionedName := fmt.Sprintf("main.%s.wasm", hash)

	// Remove old versioned files
	cleanOldWasmFiles(versionedName)

	// Rename to versioned filename
	if err := os.Rename("public/main.wasm", filepath.Join("public", versionedName)); err != nil {
		fmt.Printf("Error renaming to versioned file: %v\n", err)
		os.Exit(1)
	}

	// Update index.html with new filename
	if err := updateIndexHTML(versionedName); err != nil {
		fmt.Printf("Error updating public/index.html: %v\n", err)
		os.Exit(1)
	}

	size := float64(info.Size()) / 1024 / 1024
	compiler := "Go"
	if tinygo {
		compiler = "TinyGo"
	}
	fmt.Printf("Built public/%s (%.2f MB) with %s\n", versionedName, size, compiler)
}

// hashFile computes SHA256 hash of file content, returns first 8 chars
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil))[:8], nil
}

// cleanOldWasmFiles removes old versioned wasm files, keeping the current one
func cleanOldWasmFiles(keepFile string) {
	entries, err := os.ReadDir("public")
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		// Match main.<hash>.wasm pattern (not main.wasm) and not the one we're keeping
		// Versioned files have format: main.XXXXXXXX.wasm (8 char hash)
		if name != "main.wasm" && strings.HasPrefix(name, "main.") && strings.HasSuffix(name, ".wasm") && name != keepFile {
			os.Remove(filepath.Join("public", name))
		}
	}
}

// updateIndexHTML replaces the WASM filename reference in public/index.html
func updateIndexHTML(wasmFile string) error {
	content, err := os.ReadFile("public/index.html")
	if err != nil {
		return err
	}

	// Replace any main.*.wasm or main.wasm reference
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		// Look for fetch("/main or fetch("main with .wasm
		if (strings.Contains(line, `fetch("/main`) || strings.Contains(line, `fetch("main`)) && strings.Contains(line, `.wasm"`) {
			// Replace the wasm filename in the fetch call, preserving the leading slash
			start := strings.Index(line, `fetch("`)
			if start != -1 {
				end := strings.Index(line[start:], `.wasm"`)
				if end != -1 {
					lines[i] = line[:start] + `fetch("/` + wasmFile + `"` + line[start+end+6:]
				}
			}
		}
	}

	return os.WriteFile("public/index.html", []byte(strings.Join(lines, "\n")), 0644)
}

func runDev(port int, tinygo bool) {
	// Check we're in a gux project
	if _, err := os.Stat("cmd/app"); os.IsNotExist(err) {
		fmt.Println("Error: no cmd/app/ directory found")
		fmt.Println("Run this command from your gux project root.")
		os.Exit(1)
	}

	// Check for wasm_exec.js
	if _, err := os.Stat("public/wasm_exec.js"); os.IsNotExist(err) {
		fmt.Println("Error: public/wasm_exec.js not found")
		fmt.Println("Run 'gux setup' first to copy wasm_exec.js from your Go installation.")
		os.Exit(1)
	}

	// Build first
	runBuild(tinygo)

	fmt.Printf("\nStarting dev server on http://localhost:%d\n", port)

	// Check if cmd/server/ exists
	serverDir := "cmd/server"
	if _, err := os.Stat(serverDir); os.IsNotExist(err) {
		fmt.Println("Error: no cmd/server/ directory found")
		os.Exit(1)
	}

	// Run the server
	cmd := exec.Command("go", "run", "./cmd/server", "-port", fmt.Sprintf("%d", port), "-dir", "public")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(".")

	if err := cmd.Run(); err != nil {
		fmt.Printf("Server failed: %v\n", err)
		os.Exit(1)
	}
}
