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
	if !strings.Contains(s, "npm install -D tailwindcss @tailwindcss/cli") {
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

func TestGetExportedTypeNames_FindsTypes(t *testing.T) {
	tmpDir, cleanup := setupGenerateTestDir(t, "myapp")
	defer cleanup()

	// Create dto/ package with several types
	os.MkdirAll(filepath.Join(tmpDir, "dto"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "dto", "user.go"), []byte(`package dto

type UserList struct {
	ID    uint   `+"`"+`json:"id"`+"`"+`
	Email string `+"`"+`json:"email"`+"`"+`
}

type UserDetail struct {
	ID    uint   `+"`"+`json:"id"`+"`"+`
	Email string `+"`"+`json:"email"`+"`"+`
}
`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "dto", "auth.go"), []byte(`package dto

type LoginRequest struct {
	Email    string
	Password string
}

type LogoutResponse struct {
	Success bool
}
`), 0644)

	types := getExportedTypeNames("dto")
	if types == nil {
		t.Fatal("getExportedTypeNames returned nil")
	}

	expected := []string{"UserList", "UserDetail", "LoginRequest", "LogoutResponse"}
	for _, name := range expected {
		if !types[name] {
			t.Errorf("expected type %q not found", name)
		}
	}

	// Should not include unexported types
	if types["unexported"] {
		t.Error("should not include unexported types")
	}
}

func TestGetExportedTypeNames_EmptyDir(t *testing.T) {
	tmpDir, cleanup := setupGenerateTestDir(t, "myapp")
	defer cleanup()

	// dto/ does not exist
	types := getExportedTypeNames(filepath.Join(tmpDir, "nonexistent"))
	if types != nil {
		t.Errorf("expected nil for nonexistent dir, got %v", types)
	}
}

func TestGetExportedTypeNames_MultipleFiles(t *testing.T) {
	_, cleanup := setupGenerateTestDir(t, "myapp")
	defer cleanup()

	os.MkdirAll("dto", 0755)
	os.WriteFile("dto/a.go", []byte("package dto\n\ntype Alpha struct{}\n"), 0644)
	os.WriteFile("dto/b.go", []byte("package dto\n\ntype Beta struct{}\n"), 0644)

	types := getExportedTypeNames("dto")
	if !types["Alpha"] || !types["Beta"] {
		t.Errorf("expected both Alpha and Beta, got %v", types)
	}
}

func TestGenerateAuthDTOWithAliases_AllAliased(t *testing.T) {
	_, cleanup := setupGenerateTestDir(t, "myapp")
	defer cleanup()

	os.MkdirAll("guxgen/dto", 0755)

	err := generateAuthDTOWithAliases("myapp", []string{"LoginRequest", "LoginResponse", "LogoutResponse", "Breadcrumb"})
	if err != nil {
		t.Fatalf("generateAuthDTOWithAliases: %v", err)
	}

	content, err := os.ReadFile("guxgen/dto/auth.go")
	if err != nil {
		t.Fatalf("read auth.go: %v", err)
	}

	s := string(content)

	// Should have type aliases for all four
	for _, typeName := range []string{"LoginRequest", "LoginResponse", "LogoutResponse", "Breadcrumb"} {
		expected := "type " + typeName + " = manualdto." + typeName
		if !strings.Contains(s, expected) {
			t.Errorf("expected alias %q, not found in:\n%s", expected, s)
		}
	}

	// Should have import
	if !strings.Contains(s, `import manualdto "myapp/dto"`) {
		t.Errorf("expected import of myapp/dto, got:\n%s", s)
	}

	// Should NOT have fresh struct definitions
	if strings.Contains(s, "struct {") {
		t.Errorf("should not have fresh struct definitions when all are aliased, got:\n%s", s)
	}
}

func TestGenerateAuthDTOWithAliases_PartialAlias(t *testing.T) {
	_, cleanup := setupGenerateTestDir(t, "myapp")
	defer cleanup()

	os.MkdirAll("guxgen/dto", 0755)

	// Only LogoutResponse exists in dto/
	err := generateAuthDTOWithAliases("myapp", []string{"LogoutResponse"})
	if err != nil {
		t.Fatalf("generateAuthDTOWithAliases: %v", err)
	}

	content, err := os.ReadFile("guxgen/dto/auth.go")
	if err != nil {
		t.Fatalf("read auth.go: %v", err)
	}

	s := string(content)

	// LogoutResponse should be aliased
	if !strings.Contains(s, "type LogoutResponse = manualdto.LogoutResponse") {
		t.Errorf("expected LogoutResponse alias, got:\n%s", s)
	}

	// LoginRequest should be a fresh struct (not aliased)
	if strings.Contains(s, "type LoginRequest = ") {
		t.Errorf("LoginRequest should NOT be aliased, got:\n%s", s)
	}
	if !strings.Contains(s, "type LoginRequest struct") {
		t.Errorf("expected fresh LoginRequest struct, got:\n%s", s)
	}

	// Breadcrumb should be a fresh struct (not aliased)
	if !strings.Contains(s, "type Breadcrumb struct") {
		t.Errorf("expected fresh Breadcrumb struct, got:\n%s", s)
	}
}

func TestTypeAlias_ManualDTOInDifferentFile(t *testing.T) {
	_, cleanup := setupModelGenTestDir(t)
	defer cleanup()

	// Create manual DTO in a differently named file (not dto/user.go)
	os.MkdirAll("dto", 0755)
	os.WriteFile("dto/accounts.go", []byte("package dto\n\ntype UserList struct{}\ntype UserDetail struct{}\n"), 0644)

	model := &ModelDefinition{
		Name:   "User",
		Preset: "auth",
		Sections: map[string][]ModelField{
			"Account": {
				{Name: "Email", Type: "string", Table: true},
			},
		},
	}

	err := GenerateModelFilesImpl(model, nil, "myapp", false, nil, nil, nil)
	if err != nil {
		t.Fatalf("GenerateModelFilesImpl: %v", err)
	}

	content, err := os.ReadFile("guxgen/dto/user_gen.go")
	if err != nil {
		t.Fatalf("read DTO type alias file: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "type UserList = manualdto.UserList") {
		t.Errorf("expected UserList type alias even though DTO is in dto/accounts.go, got:\n%s", s)
	}
	if !strings.Contains(s, "type UserDetail = manualdto.UserDetail") {
		t.Errorf("expected UserDetail type alias, got:\n%s", s)
	}
}

func TestGenerateScaffoldDTOAliases_BuiltInUser(t *testing.T) {
	_, cleanup := setupGenerateTestDir(t, "myapp")
	defer cleanup()

	os.MkdirAll("dto", 0755)
	os.MkdirAll("guxgen/dto", 0755)
	os.WriteFile("dto/user.go", []byte("package dto\n\ntype UserList struct{}\ntype UserDetail struct{}\n"), 0644)

	dtoTypes := getExportedTypeNames("dto")

	err := generateScaffoldDTOAliases("myapp", dtoTypes)
	if err != nil {
		t.Fatalf("generateScaffoldDTOAliases: %v", err)
	}

	content, err := os.ReadFile("guxgen/dto/user_gen.go")
	if err != nil {
		t.Fatalf("read user_gen.go: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "type UserList = manualdto.UserList") {
		t.Errorf("expected UserList alias, got:\n%s", s)
	}
	if !strings.Contains(s, "type UserDetail = manualdto.UserDetail") {
		t.Errorf("expected UserDetail alias, got:\n%s", s)
	}
	if !strings.Contains(s, `import manualdto "myapp/dto"`) {
		t.Errorf("expected import, got:\n%s", s)
	}
}

func TestGenerateScaffoldDTOAliases_SkipsIfAlreadyExists(t *testing.T) {
	_, cleanup := setupGenerateTestDir(t, "myapp")
	defer cleanup()

	os.MkdirAll("dto", 0755)
	os.MkdirAll("guxgen/dto", 0755)
	os.WriteFile("dto/user.go", []byte("package dto\n\ntype UserList struct{}\ntype UserDetail struct{}\n"), 0644)

	// Pre-existing file (e.g., from GenerateModelFilesImpl)
	existing := "// existing content\npackage dto\n"
	os.WriteFile("guxgen/dto/user_gen.go", []byte(existing), 0644)

	dtoTypes := getExportedTypeNames("dto")

	err := generateScaffoldDTOAliases("myapp", dtoTypes)
	if err != nil {
		t.Fatalf("generateScaffoldDTOAliases: %v", err)
	}

	// Should NOT overwrite existing file
	content, _ := os.ReadFile("guxgen/dto/user_gen.go")
	if string(content) != existing {
		t.Errorf("should not overwrite existing file, got:\n%s", string(content))
	}
}

func TestGenerateScaffoldDTOAliases_NoUserDTOs(t *testing.T) {
	_, cleanup := setupGenerateTestDir(t, "myapp")
	defer cleanup()

	os.MkdirAll("dto", 0755)
	os.MkdirAll("guxgen/dto", 0755)
	// dto/ has types but NOT UserList/UserDetail
	os.WriteFile("dto/product.go", []byte("package dto\n\ntype ProductList struct{}\n"), 0644)

	dtoTypes := getExportedTypeNames("dto")

	err := generateScaffoldDTOAliases("myapp", dtoTypes)
	if err != nil {
		t.Fatalf("generateScaffoldDTOAliases: %v", err)
	}

	// Should NOT create user_gen.go
	if _, err := os.Stat("guxgen/dto/user_gen.go"); err == nil {
		t.Error("should not create user_gen.go when UserList/UserDetail don't exist in dto/")
	}
}
