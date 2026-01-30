package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupGenerateTestDir creates a temp directory with go.mod and changes into it.
// Returns the temp dir path and a cleanup function.
func setupGenerateTestDir(t *testing.T, modulePath string) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "gux-gen-test-*")
	if err != nil {
		t.Fatal(err)
	}
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)

	// Write a minimal go.mod
	goMod := "module " + modulePath + "\n\ngo 1.21\n"
	os.WriteFile("go.mod", []byte(goMod), 0644)

	return tmpDir, func() {
		os.Chdir(origDir)
		os.RemoveAll(tmpDir)
	}
}

func TestRegenerateDockerfile_CreatesFile(t *testing.T) {
	_, cleanup := setupGenerateTestDir(t, "github.com/testuser/myapp")
	defer cleanup()

	err := regenerateDockerfile()
	if err != nil {
		t.Fatalf("regenerateDockerfile() error: %v", err)
	}

	content, err := os.ReadFile("Dockerfile")
	if err != nil {
		t.Fatalf("could not read Dockerfile: %v", err)
	}

	s := string(content)

	// Should contain the gux module reference
	if !strings.Contains(s, "github.com/dougbarrett/gux") {
		t.Error("Dockerfile missing gux module reference")
	}

	// Should contain key Dockerfile instructions
	if !strings.Contains(s, "FROM tinygo/tinygo:latest AS builder") {
		t.Error("Dockerfile missing builder stage")
	}
	if !strings.Contains(s, "FROM alpine:3.21") {
		t.Error("Dockerfile missing alpine production stage")
	}
	if !strings.Contains(s, "gux gen") {
		t.Error("Dockerfile missing gux gen command")
	}
	if !strings.Contains(s, "gux build") {
		t.Error("Dockerfile missing gux build command")
	}
	if !strings.Contains(s, "HEALTHCHECK") {
		t.Error("Dockerfile missing HEALTHCHECK")
	}
}

func TestRegenerateDockerfile_OverwritesExisting(t *testing.T) {
	_, cleanup := setupGenerateTestDir(t, "github.com/testuser/myapp")
	defer cleanup()

	// Write an old Dockerfile
	os.WriteFile("Dockerfile", []byte("# old dockerfile\nFROM scratch\n"), 0644)

	err := regenerateDockerfile()
	if err != nil {
		t.Fatalf("regenerateDockerfile() error: %v", err)
	}

	content, err := os.ReadFile("Dockerfile")
	if err != nil {
		t.Fatalf("could not read Dockerfile: %v", err)
	}

	s := string(content)
	if strings.Contains(s, "# old dockerfile") {
		t.Error("Dockerfile was not overwritten")
	}
	if !strings.Contains(s, "FROM tinygo/tinygo:latest AS builder") {
		t.Error("Dockerfile missing new content after overwrite")
	}
}

func TestRegenerateDockerfile_UsesModulePath(t *testing.T) {
	_, cleanup := setupGenerateTestDir(t, "github.com/acme/coolapp")
	defer cleanup()

	err := regenerateDockerfile()
	if err != nil {
		t.Fatalf("regenerateDockerfile() error: %v", err)
	}

	content, err := os.ReadFile("Dockerfile")
	if err != nil {
		t.Fatalf("could not read Dockerfile: %v", err)
	}

	s := string(content)
	// The template header should contain the app name (directory basename)
	// The gux install line should reference the gux module, not the app module
	if !strings.Contains(s, "github.com/dougbarrett/gux/cmd/gux@") {
		t.Error("Dockerfile missing gux install command")
	}
}

func TestRegenerateDockerfile_NodejsInstall(t *testing.T) {
	_, cleanup := setupGenerateTestDir(t, "github.com/testuser/myapp")
	defer cleanup()

	err := regenerateDockerfile()
	if err != nil {
		t.Fatalf("regenerateDockerfile() error: %v", err)
	}

	content, err := os.ReadFile("Dockerfile")
	if err != nil {
		t.Fatalf("could not read Dockerfile: %v", err)
	}

	s := string(content)
	// Should install nodejs and npm for Tailwind CSS
	if !strings.Contains(s, "nodejs npm") {
		t.Error("Dockerfile missing Node.js/npm installation")
	}
	if !strings.Contains(s, "npm install -D tailwindcss") {
		t.Error("Dockerfile missing tailwindcss npm install")
	}
}

func TestRegenerateDockerfile_AppNameFromDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gux-gen-test-appname-*")
	if err != nil {
		t.Fatal(err)
	}
	// Create a subdirectory with a known name
	appDir := filepath.Join(tmpDir, "my-cool-app")
	os.MkdirAll(appDir, 0755)

	origDir, _ := os.Getwd()
	os.Chdir(appDir)
	defer func() {
		os.Chdir(origDir)
		os.RemoveAll(tmpDir)
	}()

	os.WriteFile("go.mod", []byte("module github.com/test/my-cool-app\n\ngo 1.21\n"), 0644)

	err = regenerateDockerfile()
	if err != nil {
		t.Fatalf("regenerateDockerfile() error: %v", err)
	}

	content, err := os.ReadFile("Dockerfile")
	if err != nil {
		t.Fatalf("could not read Dockerfile: %v", err)
	}

	s := string(content)
	// The template comment should reference the app name
	if !strings.Contains(s, "my-cool-app") {
		t.Error("Dockerfile does not contain app name from directory")
	}
}
