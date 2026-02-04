package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTestFile creates a temporary Go source file for testing parseCRUDModels.
func writeTestFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "app.go")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	return path
}

// TestParseCRUDModels_SingleModelBasic tests parsing a simple single-model app.
func TestParseCRUDModels_SingleModelBasic(t *testing.T) {
	src := `package main

import (
	"myapp/guxgen/models"
	"github.com/dougbarrett/gux/core"
)

func main() {
	app := core.New()
	app.CRUD(models.Counter{})
}
`
	path := writeTestFile(t, src)
	models, modelsImport, _, err := parseCRUDModels(path)
	if err != nil {
		t.Fatalf("parseCRUDModels: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}
	if models[0].Name != "Counter" {
		t.Errorf("model name = %q, want %q", models[0].Name, "Counter")
	}
	if models[0].PluralName != "Counters" {
		t.Errorf("plural name = %q, want %q", models[0].PluralName, "Counters")
	}
	if modelsImport != "myapp/guxgen/models" {
		t.Errorf("modelsImport = %q, want %q", modelsImport, "myapp/guxgen/models")
	}
	if models[0].ModelImportPath != "myapp/guxgen/models" {
		t.Errorf("ModelImportPath = %q, want %q", models[0].ModelImportPath, "myapp/guxgen/models")
	}
}

// TestParseCRUDModels_WithDTOs tests parsing CRUD models with WithListDTO/WithDetailDTO.
func TestParseCRUDModels_WithDTOs(t *testing.T) {
	src := `package main

import (
	"myapp/guxgen/models"
	"myapp/guxgen/dto"
	"github.com/dougbarrett/gux/core"
)

func main() {
	app := core.New()
	app.CRUD(models.User{},
		core.WithListDTO(dto.UserList{}),
		core.WithDetailDTO(dto.UserDetail{}),
	)
}
`
	path := writeTestFile(t, src)
	models, _, dtoImport, err := parseCRUDModels(path)
	if err != nil {
		t.Fatalf("parseCRUDModels: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}
	m := models[0]
	if m.Name != "User" {
		t.Errorf("model name = %q, want %q", m.Name, "User")
	}
	if m.ListDTO != "UserList" {
		t.Errorf("ListDTO = %q, want %q", m.ListDTO, "UserList")
	}
	if m.DetailDTO != "UserDetail" {
		t.Errorf("DetailDTO = %q, want %q", m.DetailDTO, "UserDetail")
	}
	if m.DTOPackage != "dto" {
		t.Errorf("DTOPackage = %q, want %q", m.DTOPackage, "dto")
	}
	if m.DTOImportPath != "myapp/guxgen/dto" {
		t.Errorf("DTOImportPath = %q, want %q", m.DTOImportPath, "myapp/guxgen/dto")
	}
	if dtoImport != "myapp/guxgen/dto" {
		t.Errorf("dtoImport = %q, want %q", dtoImport, "myapp/guxgen/dto")
	}
}

// TestParseCRUDModels_AliasedDTOImport tests that aliased DTO imports resolve
// to the canonical package name and correct import path.
func TestParseCRUDModels_AliasedDTOImport(t *testing.T) {
	src := `package main

import (
	"myapp/guxgen/models"
	guxdto "myapp/guxgen/dto"
	"github.com/dougbarrett/gux/core"
)

func main() {
	app := core.New()
	app.CRUD(models.Lead{},
		core.WithListDTO(guxdto.LeadList{}),
		core.WithDetailDTO(guxdto.LeadDetail{}),
	)
}
`
	path := writeTestFile(t, src)
	models, _, _, err := parseCRUDModels(path)
	if err != nil {
		t.Fatalf("parseCRUDModels: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}
	m := models[0]
	// DTOPackage should be the canonical name "dto", not the alias "guxdto"
	if m.DTOPackage != "dto" {
		t.Errorf("DTOPackage = %q, want %q (canonical name, not alias)", m.DTOPackage, "dto")
	}
	if m.DTOImportPath != "myapp/guxgen/dto" {
		t.Errorf("DTOImportPath = %q, want %q", m.DTOImportPath, "myapp/guxgen/dto")
	}
}

// TestParseCRUDModels_DualModelsPackages tests the key scenario from issue #47:
// both guxgen/models (bare import) and local models/ (aliased import) exist.
// Each model should track its own import path correctly.
func TestParseCRUDModels_DualModelsPackages(t *testing.T) {
	src := `package main

import (
	"myapp/guxgen/dto"
	"myapp/guxgen/models"
	usermodels "myapp/models"
	"github.com/dougbarrett/gux/core"
)

func main() {
	app := core.New()
	// Internal model from guxgen/models
	app.CRUD(models.Lead{},
		core.WithRoles("admin"),
		core.WithListDTO(dto.LeadList{}),
		core.WithDetailDTO(dto.LeadDetail{}),
	)
	// External model from local models/
	app.CRUD(usermodels.User{},
		core.WithRoles("admin"),
		core.WithListDTO(dto.UserList{}),
		core.WithDetailDTO(dto.UserDetail{}),
	)
}
`
	path := writeTestFile(t, src)
	models, _, _, err := parseCRUDModels(path)
	if err != nil {
		t.Fatalf("parseCRUDModels: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}

	// Find Lead and User models by name
	var lead, user *CRUDModel
	for i := range models {
		switch models[i].Name {
		case "Lead":
			lead = &models[i]
		case "User":
			user = &models[i]
		}
	}
	if lead == nil {
		t.Fatal("Lead model not found")
	}
	if user == nil {
		t.Fatal("User model not found")
	}

	// Lead should resolve to guxgen/models
	if lead.ModelImportPath != "myapp/guxgen/models" {
		t.Errorf("Lead.ModelImportPath = %q, want %q", lead.ModelImportPath, "myapp/guxgen/models")
	}

	// User should resolve to local models/
	if user.ModelImportPath != "myapp/models" {
		t.Errorf("User.ModelImportPath = %q, want %q", user.ModelImportPath, "myapp/models")
	}

	// Both DTOs should resolve to guxgen/dto (same package)
	if lead.DTOImportPath != "myapp/guxgen/dto" {
		t.Errorf("Lead.DTOImportPath = %q, want %q", lead.DTOImportPath, "myapp/guxgen/dto")
	}
	if user.DTOImportPath != "myapp/guxgen/dto" {
		t.Errorf("User.DTOImportPath = %q, want %q", user.DTOImportPath, "myapp/guxgen/dto")
	}

	// Both DTOs should have canonical package name "dto"
	if lead.DTOPackage != "dto" {
		t.Errorf("Lead.DTOPackage = %q, want %q", lead.DTOPackage, "dto")
	}
	if user.DTOPackage != "dto" {
		t.Errorf("User.DTOPackage = %q, want %q", user.DTOPackage, "dto")
	}
}

// TestParseCRUDModels_DualDTOPackages tests dual DTO packages:
// guxgen/dto and local dto/ with different aliases.
func TestParseCRUDModels_DualDTOPackages(t *testing.T) {
	src := `package main

import (
	"myapp/guxgen/dto"
	extdto "myapp/dto"
	"myapp/guxgen/models"
	extmodels "myapp/models"
	"github.com/dougbarrett/gux/core"
)

func main() {
	app := core.New()
	// Internal model with internal DTOs
	app.CRUD(models.Lead{},
		core.WithListDTO(dto.LeadList{}),
	)
	// External model with external DTOs
	app.CRUD(extmodels.Widget{},
		core.WithListDTO(extdto.WidgetList{}),
	)
}
`
	path := writeTestFile(t, src)
	models, _, _, err := parseCRUDModels(path)
	if err != nil {
		t.Fatalf("parseCRUDModels: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}

	var lead, widget *CRUDModel
	for i := range models {
		switch models[i].Name {
		case "Lead":
			lead = &models[i]
		case "Widget":
			widget = &models[i]
		}
	}
	if lead == nil {
		t.Fatal("Lead model not found")
	}
	if widget == nil {
		t.Fatal("Widget model not found")
	}

	// Lead: internal model, internal DTOs
	if lead.ModelImportPath != "myapp/guxgen/models" {
		t.Errorf("Lead.ModelImportPath = %q, want %q", lead.ModelImportPath, "myapp/guxgen/models")
	}
	if lead.DTOImportPath != "myapp/guxgen/dto" {
		t.Errorf("Lead.DTOImportPath = %q, want %q", lead.DTOImportPath, "myapp/guxgen/dto")
	}
	if lead.DTOPackage != "dto" {
		t.Errorf("Lead.DTOPackage = %q, want %q", lead.DTOPackage, "dto")
	}

	// Widget: external model, external DTOs
	if widget.ModelImportPath != "myapp/models" {
		t.Errorf("Widget.ModelImportPath = %q, want %q", widget.ModelImportPath, "myapp/models")
	}
	if widget.DTOImportPath != "myapp/dto" {
		t.Errorf("Widget.DTOImportPath = %q, want %q", widget.DTOImportPath, "myapp/dto")
	}
	if widget.DTOPackage != "dto" {
		t.Errorf("Widget.DTOPackage = %q, want canonical %q (not alias %q)", widget.DTOPackage, "dto", "extdto")
	}
}

// TestParseCRUDModels_ExternalModelInternalDTO tests the mixed case:
// external model (local models/) but internal DTOs (guxgen/dto).
// This is the exact scenario from issue #47.
func TestParseCRUDModels_ExternalModelInternalDTO(t *testing.T) {
	src := `package main

import (
	"myapp/guxgen/dto"
	"myapp/guxgen/models"
	usermodels "myapp/models"
	"github.com/dougbarrett/gux/core"
)

func main() {
	app := core.New()
	// Internal model with internal DTOs
	app.CRUD(models.Lead{},
		core.WithListDTO(dto.LeadList{}),
		core.WithDetailDTO(dto.LeadDetail{}),
	)
	// External model but DTOs are ALSO from guxgen/dto (not local dto/)
	app.CRUD(usermodels.User{},
		core.WithListDTO(dto.UserList{}),
		core.WithDetailDTO(dto.UserDetail{}),
	)
}
`
	path := writeTestFile(t, src)
	models, _, _, err := parseCRUDModels(path)
	if err != nil {
		t.Fatalf("parseCRUDModels: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}

	var lead, user *CRUDModel
	for i := range models {
		switch models[i].Name {
		case "Lead":
			lead = &models[i]
		case "User":
			user = &models[i]
		}
	}
	if lead == nil {
		t.Fatal("Lead model not found")
	}
	if user == nil {
		t.Fatal("User model not found")
	}

	// Lead: guxgen/models model, guxgen/dto DTOs
	if lead.ModelImportPath != "myapp/guxgen/models" {
		t.Errorf("Lead.ModelImportPath = %q, want %q", lead.ModelImportPath, "myapp/guxgen/models")
	}
	if lead.DTOImportPath != "myapp/guxgen/dto" {
		t.Errorf("Lead.DTOImportPath = %q, want %q", lead.DTOImportPath, "myapp/guxgen/dto")
	}

	// User: local models/ model, but guxgen/dto DTOs
	if user.ModelImportPath != "myapp/models" {
		t.Errorf("User.ModelImportPath = %q, want %q", user.ModelImportPath, "myapp/models")
	}
	// Critical: User's DTOs are from guxgen/dto, NOT from local dto/
	if user.DTOImportPath != "myapp/guxgen/dto" {
		t.Errorf("User.DTOImportPath = %q, want %q (DTOs are internal even though model is external)", user.DTOImportPath, "myapp/guxgen/dto")
	}
}

// TestParseCRUDModels_ExternalityFlags tests that IsExternal and DTOIsExternal
// are set correctly based on import paths (not config membership).
// This simulates the post-parse processing that happens in build/generate.
func TestParseCRUDModels_ExternalityFlags(t *testing.T) {
	src := `package main

import (
	"myapp/guxgen/dto"
	"myapp/guxgen/models"
	usermodels "myapp/models"
	"github.com/dougbarrett/gux/core"
)

func main() {
	app := core.New()
	app.CRUD(models.Lead{},
		core.WithListDTO(dto.LeadList{}),
	)
	app.CRUD(usermodels.User{},
		core.WithListDTO(dto.UserList{}),
	)
}
`
	path := writeTestFile(t, src)
	models, _, _, err := parseCRUDModels(path)
	if err != nil {
		t.Fatalf("parseCRUDModels: %v", err)
	}

	// Simulate the post-parse processing from build.go
	guxgenModelsPath := "myapp/guxgen/models"
	guxgenDTOPath := "myapp/guxgen/dto"

	for i := range models {
		if models[i].ModelImportPath != "" && models[i].ModelImportPath != guxgenModelsPath {
			models[i].IsExternal = true
		}
		if models[i].DTOImportPath != "" && models[i].DTOImportPath != guxgenDTOPath {
			models[i].DTOIsExternal = true
		}
	}

	var lead, user *CRUDModel
	for i := range models {
		switch models[i].Name {
		case "Lead":
			lead = &models[i]
		case "User":
			user = &models[i]
		}
	}

	// Lead: internal model, internal DTOs
	if lead.IsExternal {
		t.Error("Lead.IsExternal = true, want false (model is from guxgen/models)")
	}
	if lead.DTOIsExternal {
		t.Error("Lead.DTOIsExternal = true, want false (DTOs are from guxgen/dto)")
	}

	// User: external model, but internal DTOs
	if !user.IsExternal {
		t.Error("User.IsExternal = false, want true (model is from local models/)")
	}
	if user.DTOIsExternal {
		t.Error("User.DTOIsExternal = true, want false (DTOs are from guxgen/dto, not local dto/)")
	}
}

// TestParseCRUDModels_ModelWithoutDTO tests a model registered without any DTO options.
func TestParseCRUDModels_ModelWithoutDTO(t *testing.T) {
	src := `package main

import (
	"myapp/guxgen/models"
	"github.com/dougbarrett/gux/core"
)

func main() {
	app := core.New()
	app.CRUD(models.Counter{})
}
`
	path := writeTestFile(t, src)
	models, _, _, err := parseCRUDModels(path)
	if err != nil {
		t.Fatalf("parseCRUDModels: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}
	m := models[0]
	if m.ListDTO != "" {
		t.Errorf("ListDTO = %q, want empty (no WithListDTO option)", m.ListDTO)
	}
	if m.DetailDTO != "" {
		t.Errorf("DetailDTO = %q, want empty (no WithDetailDTO option)", m.DetailDTO)
	}
	if m.DTOImportPath != "" {
		t.Errorf("DTOImportPath = %q, want empty", m.DTOImportPath)
	}
}

// TestParseCRUDModels_MultipleModels tests parsing multiple CRUD registrations.
func TestParseCRUDModels_MultipleModels(t *testing.T) {
	src := `package main

import (
	"myapp/guxgen/models"
	"myapp/guxgen/dto"
	"github.com/dougbarrett/gux/core"
)

func main() {
	app := core.New()
	app.CRUD(models.User{},
		core.WithListDTO(dto.UserList{}),
	)
	app.CRUD(models.Post{},
		core.WithListDTO(dto.PostList{}),
		core.WithDetailDTO(dto.PostDetail{}),
	)
	app.CRUD(models.Tag{})
}
`
	path := writeTestFile(t, src)
	models, _, _, err := parseCRUDModels(path)
	if err != nil {
		t.Fatalf("parseCRUDModels: %v", err)
	}
	if len(models) != 3 {
		t.Fatalf("expected 3 models, got %d", len(models))
	}

	modelMap := make(map[string]*CRUDModel)
	for i := range models {
		modelMap[models[i].Name] = &models[i]
	}

	if u, ok := modelMap["User"]; !ok {
		t.Error("User model not found")
	} else {
		if u.ListDTO != "UserList" {
			t.Errorf("User.ListDTO = %q, want %q", u.ListDTO, "UserList")
		}
	}

	if p, ok := modelMap["Post"]; !ok {
		t.Error("Post model not found")
	} else {
		if p.ListDTO != "PostList" {
			t.Errorf("Post.ListDTO = %q, want %q", p.ListDTO, "PostList")
		}
		if p.DetailDTO != "PostDetail" {
			t.Errorf("Post.DetailDTO = %q, want %q", p.DetailDTO, "PostDetail")
		}
	}

	if tag, ok := modelMap["Tag"]; !ok {
		t.Error("Tag model not found")
	} else {
		if tag.ListDTO != "" || tag.DetailDTO != "" {
			t.Errorf("Tag should have no DTOs, got ListDTO=%q DetailDTO=%q", tag.ListDTO, tag.DetailDTO)
		}
	}
}

// TestParseCRUDModels_WithDTO tests the WithDTO option (sets both list and detail).
func TestParseCRUDModels_WithDTO(t *testing.T) {
	src := `package main

import (
	"myapp/guxgen/models"
	"myapp/guxgen/dto"
	"github.com/dougbarrett/gux/core"
)

func main() {
	app := core.New()
	app.CRUD(models.Product{},
		core.WithDTO(dto.ProductDTO{}),
	)
}
`
	path := writeTestFile(t, src)
	models, _, _, err := parseCRUDModels(path)
	if err != nil {
		t.Fatalf("parseCRUDModels: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}
	m := models[0]
	if m.ListDTO != "ProductDTO" {
		t.Errorf("ListDTO = %q, want %q", m.ListDTO, "ProductDTO")
	}
	if m.DetailDTO != "ProductDTO" {
		t.Errorf("DetailDTO = %q, want %q", m.DetailDTO, "ProductDTO")
	}
}

// TestParseCRUDModels_CoreAuditEntry tests parsing a core.AuditEntry{} model
// (no separate models package, uses "core" as package).
func TestParseCRUDModels_CorePackageModel(t *testing.T) {
	src := `package main

import (
	"github.com/dougbarrett/gux/core"
)

func main() {
	app := core.New()
	app.CRUD(core.AuditEntry{},
		core.WithRoles("admin"),
	)
}
`
	path := writeTestFile(t, src)
	models, _, _, err := parseCRUDModels(path)
	if err != nil {
		t.Fatalf("parseCRUDModels: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}
	if models[0].Name != "AuditEntry" {
		t.Errorf("model name = %q, want %q", models[0].Name, "AuditEntry")
	}
}

// TestIsNumericParam tests the isNumericParam helper used for endpoint path params (#51).
func TestIsNumericParam(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"id", true},
		{"user_id", true},
		{"userId", true},
		{"userID", true},
		{"post_id", true},
		{"slug", false},
		{"code", false},
		{"token", false},
		{"name", false},
		{"email", false},
		{"category", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNumericParam(tt.name)
			if got != tt.want {
				t.Errorf("isNumericParam(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// TestGenerateFieldMapping_PointerTypes tests that generateFieldMapping handles
// pointer type mismatches between model and DTO fields (#53).
func TestGenerateFieldMapping_PointerTypes(t *testing.T) {
	// Set up a temporary model file with pointer fields for getModelFieldTypes to parse
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// Create guxgen/models directory with a test model
	os.MkdirAll(filepath.Join("guxgen", "models"), 0755)
	modelContent := `package models

type Promotion struct {
	SecondaryColor *string
	AdTagURL       *string
	MinWatchTimeSec *float64
	Name           string
	Active         bool
}
`
	os.WriteFile(filepath.Join("guxgen", "models", "promotion.go"), []byte(modelContent), 0644)

	// Clear the cache since we're in a new directory
	modelFieldTypesCache = make(map[string]map[string]string)

	info := &DTOInfo{
		ModelName: "Promotion",
		Fields: []DTOFieldMapping{
			{DTOField: "Name", DTOType: "string", ModelField: "Name"},
			{DTOField: "SecondaryColor", DTOType: "string", ModelField: "SecondaryColor"},
			{DTOField: "AdTagURL", DTOType: "string", ModelField: "AdTagURL"},
			{DTOField: "MinWatchTimeSec", DTOType: "float64", ModelField: "MinWatchTimeSec"},
			{DTOField: "Active", DTOType: "bool", ModelField: "Active"},
		},
	}

	result := generateFieldMapping(info, "item")

	// String field (no pointer) should be direct assignment
	if !strings.Contains(result, "Name: item.Name,") {
		t.Errorf("expected direct assignment for Name, got:\n%s", result)
	}

	// *string model -> string DTO should use pointer dereference
	if !strings.Contains(result, "SecondaryColor: func() string { if item.SecondaryColor != nil") {
		t.Errorf("expected pointer dereference for SecondaryColor (*string -> string), got:\n%s", result)
	}

	if !strings.Contains(result, "AdTagURL: func() string { if item.AdTagURL != nil") {
		t.Errorf("expected pointer dereference for AdTagURL (*string -> string), got:\n%s", result)
	}

	// *float64 model -> float64 DTO should use pointer dereference
	if !strings.Contains(result, "MinWatchTimeSec: func() float64 { if item.MinWatchTimeSec != nil") {
		t.Errorf("expected pointer dereference for MinWatchTimeSec (*float64 -> float64), got:\n%s", result)
	}

	// bool field (no pointer mismatch) should be direct assignment
	if !strings.Contains(result, "Active: item.Active,") {
		t.Errorf("expected direct assignment for Active, got:\n%s", result)
	}
}

// TestGenerateFieldMapping_NoPointerMismatch tests that matching types use direct assignment.
func TestGenerateFieldMapping_NoPointerMismatch(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.MkdirAll(filepath.Join("guxgen", "models"), 0755)
	modelContent := `package models

type Product struct {
	Name  string
	Price float64
	Slug  *string
}
`
	os.WriteFile(filepath.Join("guxgen", "models", "product.go"), []byte(modelContent), 0644)
	modelFieldTypesCache = make(map[string]map[string]string)

	info := &DTOInfo{
		ModelName: "Product",
		Fields: []DTOFieldMapping{
			{DTOField: "Name", DTOType: "string", ModelField: "Name"},
			{DTOField: "Price", DTOType: "float64", ModelField: "Price"},
			{DTOField: "Slug", DTOType: "*string", ModelField: "Slug"},
		},
	}

	result := generateFieldMapping(info, "item")

	// Both string → should be direct
	if !strings.Contains(result, "Name: item.Name,") {
		t.Errorf("expected direct assignment for matching types, got:\n%s", result)
	}

	// Both *string → should be direct (DTO also pointer)
	if !strings.Contains(result, "Slug: item.Slug,") {
		t.Errorf("expected direct assignment for matching pointer types, got:\n%s", result)
	}
}

// TestGenerateDTOStructCode_PointerTypes tests that pointer types in inline DTOs
// are correctly mapped (#53).
func TestGenerateDTOStructCode_PointerTypes(t *testing.T) {
	fields := []ModelFieldInfo{
		{Name: "Name", Type: "string", JSONName: "name"},
		{Name: "Description", Type: "*string", JSONName: "description"},
		{Name: "Price", Type: "*float64", JSONName: "price"},
		{Name: "ParentID", Type: "*uint", JSONName: "parent_id"},
	}

	result := generateDTOStructCode("TestDTO", fields)

	if !strings.Contains(result, "Name string") {
		t.Errorf("expected string type for Name, got:\n%s", result)
	}
	// *string preserved as *string in DTOs
	if !strings.Contains(result, "Description *string") {
		t.Errorf("expected *string type for Description, got:\n%s", result)
	}
	// *float64 preserved as *float64 in DTOs
	if !strings.Contains(result, "Price *float64") {
		t.Errorf("expected *float64 type for Price, got:\n%s", result)
	}
	if !strings.Contains(result, "ParentID *uint") {
		t.Errorf("expected *uint type for ParentID, got:\n%s", result)
	}
}

// TestGenerateServerAPICode_PointerTypes tests that server API code handles
// pointer type conversions correctly (#53).
func TestGenerateServerAPICode_PointerTypes(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.MkdirAll(filepath.Join("guxgen", "models"), 0755)
	modelContent := `package models

type Widget struct {
	Description *string
	Weight      *float64
	Label       string
}
`
	os.WriteFile(filepath.Join("guxgen", "models", "widget.go"), []byte(modelContent), 0644)
	modelFieldTypesCache = make(map[string]map[string]string)

	result := generateServerAPICode("Widget", "Widgets", "models", "")

	// Since DTO preserves exact types, pointer fields should be direct assignments
	// *string model → *string DTO: direct pass-through
	if !strings.Contains(result, "Description: item.Description,") {
		t.Errorf("expected direct assignment for *string field Description, got:\n%s", result)
	}

	// *float64 model → *float64 DTO: direct pass-through
	if !strings.Contains(result, "Weight: item.Weight,") {
		t.Errorf("expected direct assignment for *float64 field Weight, got:\n%s", result)
	}

	// string field should be direct assignment
	if !strings.Contains(result, "Label: item.Label,") {
		t.Errorf("expected direct assignment for string field Label, got:\n%s", result)
	}
}

// TestResolveTypeAlias_FollowsAlias tests that type aliases (type X = pkg.Y)
// are resolved to the actual struct definition for pointer type detection (#53).
func TestResolveTypeAlias_FollowsAlias(t *testing.T) {
	// Create a temp project structure:
	// go.mod
	// dto/promotion.go          (actual struct with *string fields)
	// guxgen/dto/promotion_gen.go (type alias to dto.PromotionDetail)
	// models/promotion.go       (actual model with *float64 fields)
	// guxgen/models/promotion_gen.go (type alias to models.Promotion)
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// go.mod
	os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.23\n"), 0644)

	// dto/promotion.go - real struct with pointer fields
	os.MkdirAll("dto", 0755)
	os.WriteFile("dto/promotion.go", []byte(`package dto

type PromotionDetail struct {
	Name           string  `+"`"+`json:"name"`+"`"+`
	AdTagURL       *string `+"`"+`json:"ad_tag_url"`+"`"+`
	WebhookURL     *string `+"`"+`json:"webhook_url"`+"`"+`
	SecondaryColor *string `+"`"+`json:"secondary_color"`+"`"+`
}
`), 0644)

	// guxgen/dto/promotion_gen.go - type alias
	os.MkdirAll(filepath.Join("guxgen", "dto"), 0755)
	os.WriteFile(filepath.Join("guxgen", "dto", "promotion_gen.go"), []byte(`package dto

import manualdto "testapp/dto"

type PromotionDetail = manualdto.PromotionDetail
`), 0644)

	// models/promotion.go - real struct with pointer fields
	os.MkdirAll("models", 0755)
	os.WriteFile("models/promotion.go", []byte(`package models

import "gorm.io/gorm"

type Promotion struct {
	gorm.Model
	Name           string   `+"`"+`json:"name"`+"`"+`
	AdTagURL       *string  `+"`"+`json:"ad_tag_url"`+"`"+`
	MinWatchTimeSec *float64 `+"`"+`json:"min_watch_time_sec"`+"`"+`
}
`), 0644)

	// guxgen/models/promotion_gen.go - type alias
	os.MkdirAll(filepath.Join("guxgen", "models"), 0755)
	os.WriteFile(filepath.Join("guxgen", "models", "promotion_gen.go"), []byte(`package models

import customModels "testapp/models"

type Promotion = customModels.Promotion
`), 0644)

	// Test getDTOFieldTypes resolves the alias
	dtoFields, err := getDTOFieldTypes("PromotionDetail")
	if err != nil {
		t.Fatalf("getDTOFieldTypes: %v", err)
	}

	if dtoFields["AdTagURL"] != "*string" {
		t.Errorf("getDTOFieldTypes: AdTagURL = %q, want *string", dtoFields["AdTagURL"])
	}
	if dtoFields["WebhookURL"] != "*string" {
		t.Errorf("getDTOFieldTypes: WebhookURL = %q, want *string", dtoFields["WebhookURL"])
	}
	if dtoFields["Name"] != "string" {
		t.Errorf("getDTOFieldTypes: Name = %q, want string", dtoFields["Name"])
	}

	// Test getModelFieldTypes resolves the alias
	modelFieldTypesCache = make(map[string]map[string]string) // reset cache
	modelFields, err := getModelFieldTypes("Promotion")
	if err != nil {
		t.Fatalf("getModelFieldTypes: %v", err)
	}

	if modelFields["AdTagURL"] != "*string" {
		t.Errorf("getModelFieldTypes: AdTagURL = %q, want *string", modelFields["AdTagURL"])
	}
	if modelFields["MinWatchTimeSec"] != "*float64" {
		t.Errorf("getModelFieldTypes: MinWatchTimeSec = %q, want *float64", modelFields["MinWatchTimeSec"])
	}
}

// TestGetModelFieldTypes_UserDefinedOverridesGenerated tests that when both
// models/ and guxgen/models/ contain a struct with the same name (no alias),
// the user-defined struct in models/ takes precedence (#53).
func TestGetModelFieldTypes_UserDefinedOverridesGenerated(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.23\n"), 0644)

	// models/promotion_video.go - user-defined with pointer types
	os.MkdirAll("models", 0755)
	os.WriteFile("models/promotion_video.go", []byte(`package models

import "gorm.io/gorm"

type PromotionVideo struct {
	gorm.Model
	Title           string   `+"`"+`json:"title"`+"`"+`
	MinWatchTimeSec *float64 `+"`"+`json:"min_watch_time_sec"`+"`"+`
	AdTagURL        *string  `+"`"+`json:"ad_tag_url"`+"`"+`
	PromotionID     *uint    `+"`"+`json:"promotion_id"`+"`"+`
}
`), 0644)

	// guxgen/models/promotion_video_gen.go - generated without pointer types
	os.MkdirAll(filepath.Join("guxgen", "models"), 0755)
	os.WriteFile(filepath.Join("guxgen", "models", "promotion_video_gen.go"), []byte(`package models

import "gorm.io/gorm"

type PromotionVideo struct {
	gorm.Model
	Title           string  `+"`"+`json:"title"`+"`"+`
	MinWatchTimeSec float64 `+"`"+`json:"min_watch_time_sec"`+"`"+`
	AdTagURL        string  `+"`"+`json:"ad_tag_url"`+"`"+`
	PromotionID     uint    `+"`"+`json:"promotion_id"`+"`"+`
}
`), 0644)

	// Reset cache
	modelFieldTypesCache = make(map[string]map[string]string)

	fields, err := getModelFieldTypes("PromotionVideo")
	if err != nil {
		t.Fatalf("getModelFieldTypes: %v", err)
	}

	// User-defined pointer types should win
	tests := []struct {
		field string
		want  string
	}{
		{"Title", "string"},
		{"MinWatchTimeSec", "*float64"},
		{"AdTagURL", "*string"},
		{"PromotionID", "*uint"},
	}

	for _, tt := range tests {
		if got := fields[tt.field]; got != tt.want {
			t.Errorf("getModelFieldTypes: %s = %q, want %q", tt.field, got, tt.want)
		}
	}
}

// TestGetModelFieldTypes_OnlyGenerated tests that when only guxgen/models/
// exists (no user-defined models/), the generated struct is used correctly.
func TestGetModelFieldTypes_OnlyGenerated(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.23\n"), 0644)

	// Only guxgen/models/ exists
	os.MkdirAll(filepath.Join("guxgen", "models"), 0755)
	os.WriteFile(filepath.Join("guxgen", "models", "product_gen.go"), []byte(`package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name  string
	Price float64
}
`), 0644)

	modelFieldTypesCache = make(map[string]map[string]string)

	fields, err := getModelFieldTypes("Product")
	if err != nil {
		t.Fatalf("getModelFieldTypes: %v", err)
	}

	if fields["Name"] != "string" {
		t.Errorf("Name = %q, want string", fields["Name"])
	}
	if fields["Price"] != "float64" {
		t.Errorf("Price = %q, want float64", fields["Price"])
	}
}

// TestGetModelFieldTypes_OnlyUserDefined tests that when only models/ exists
// (no guxgen/models/), the user struct is used correctly.
func TestGetModelFieldTypes_OnlyUserDefined(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.23\n"), 0644)

	// Only models/ exists
	os.MkdirAll("models", 0755)
	os.WriteFile("models/order.go", []byte(`package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	Total    *float64
	Discount *float64
	Note     *string
}
`), 0644)

	modelFieldTypesCache = make(map[string]map[string]string)

	fields, err := getModelFieldTypes("Order")
	if err != nil {
		t.Fatalf("getModelFieldTypes: %v", err)
	}

	if fields["Total"] != "*float64" {
		t.Errorf("Total = %q, want *float64", fields["Total"])
	}
	if fields["Discount"] != "*float64" {
		t.Errorf("Discount = %q, want *float64", fields["Discount"])
	}
	if fields["Note"] != "*string" {
		t.Errorf("Note = %q, want *string", fields["Note"])
	}
}

// TestGetModelFieldTypesForImport_ResolvesCorrectPackage verifies that when a model
// import path is provided, getModelFieldTypesForImport searches the correct directory
// and does NOT fall back to core (which would pick up core.App for a user's models.App).
func TestGetModelFieldTypesForImport_ResolvesCorrectPackage(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.23\n"), 0644)

	// Create a user model named "App" (same name as core.App)
	os.MkdirAll("models", 0755)
	os.WriteFile("models/app.go", []byte(`package models

import "gorm.io/gorm"

type App struct {
	gorm.Model
	Name      string
	Subdomain string
	Status    string
	Active    bool
}
`), 0644)

	modelFieldTypesCache = make(map[string]map[string]string)

	// With import path pointing to models/, should find the user's App
	fields, err := getModelFieldTypesForImport("App", "testapp/models")
	if err != nil {
		t.Fatalf("getModelFieldTypesForImport with import path: %v", err)
	}

	// Should have user model fields
	if fields["Name"] != "string" {
		t.Errorf("expected Name=string, got %q", fields["Name"])
	}
	if fields["Subdomain"] != "string" {
		t.Errorf("expected Subdomain=string, got %q", fields["Subdomain"])
	}
	if fields["Status"] != "string" {
		t.Errorf("expected Status=string, got %q", fields["Status"])
	}
	if fields["Active"] != "bool" {
		t.Errorf("expected Active=bool, got %q", fields["Active"])
	}

	// Should NOT have core.App fields
	for _, coreField := range []string{"wasmBundles", "apiPrefix", "csrfConfig", "darkMode", "storage"} {
		if _, exists := fields[coreField]; exists {
			t.Errorf("found core.App field %q in user model — import-aware resolution failed", coreField)
		}
	}
}

// TestGetModelFieldTypesForImport_NoCoreFallback verifies that when an import path
// is provided for a non-core package, we don't fall back to core even if the model
// name matches a core type.
func TestGetModelFieldTypesForImport_NoCoreFallback(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.23\n"), 0644)

	// models/ directory exists but does NOT contain an "App" struct
	os.MkdirAll("models", 0755)
	os.WriteFile("models/agent.go", []byte(`package models

type Agent struct {
	Name string
}
`), 0644)

	modelFieldTypesCache = make(map[string]map[string]string)

	// Should error — not found in models/, and should NOT fall back to core
	_, err := getModelFieldTypesForImport("App", "testapp/models")
	if err == nil {
		t.Fatal("expected error when model not found in specified import path, got nil")
	}
}

// TestGetModelFieldTypesForImport_CoreModelStillWorks verifies that core models
// (like AuditEntry) still resolve correctly when no import path is provided.
func TestGetModelFieldTypesForImport_CoreModelStillWorks(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.23\n\nrequire github.com/dougbarrett/gux v1.28.33\n"), 0644)

	modelFieldTypesCache = make(map[string]map[string]string)

	// Empty import path should still search core as fallback
	fields, err := getModelFieldTypesForImport("AuditEntry", "")
	if err != nil {
		// Core package may not be available in test env — skip if so
		t.Skipf("core package not available in test environment: %v", err)
	}

	// Should have AuditEntry fields
	if fields["Action"] == "" {
		t.Error("expected AuditEntry to have Action field")
	}
}

// TestGetModelFieldsForImport_FiltersCorrectly verifies that getModelFieldsForImport
// properly filters out gorm.Model fields and relation fields.
func TestGetModelFieldsForImport_FiltersCorrectly(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.23\n"), 0644)

	os.MkdirAll("models", 0755)
	os.WriteFile("models/app.go", []byte(`package models

import "gorm.io/gorm"

type App struct {
	gorm.Model
	Name      string
	Subdomain string
	Status    string
	OwnerID   uint
}
`), 0644)

	modelFieldTypesCache = make(map[string]map[string]string)

	fields, err := getModelFieldsForImport("App", "testapp/models")
	if err != nil {
		t.Fatalf("getModelFieldsForImport: %v", err)
	}

	// Should have user fields
	fieldNames := make(map[string]bool)
	for _, f := range fields {
		fieldNames[f.Name] = true
	}

	for _, expected := range []string{"Name", "Subdomain", "Status", "OwnerID"} {
		if !fieldNames[expected] {
			t.Errorf("expected field %q not found in results", expected)
		}
	}

	// Should NOT have gorm.Model fields
	for _, excluded := range []string{"ID", "CreatedAt", "UpdatedAt", "DeletedAt", "Model"} {
		if fieldNames[excluded] {
			t.Errorf("gorm.Model field %q should be excluded", excluded)
		}
	}
}

// TestGetModelFieldsForImport_ExcludesRelations verifies that GORM relation fields
// (same-package struct types and slices) are excluded from generated DTOs (#77).
func TestGetModelFieldsForImport_ExcludesRelations(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.23\n"), 0644)

	os.MkdirAll("models", 0755)
	os.WriteFile("models/agent_tool.go", []byte(`package models

import "gorm.io/gorm"

type Tool struct {
	gorm.Model
	Name string
}

type AgentTool struct {
	gorm.Model
	AgentID   uint
	ToolID    uint
	SortOrder int
	Tool      Tool
	Tags      []string
}
`), 0644)

	modelFieldTypesCache = make(map[string]map[string]string)

	fields, err := getModelFieldsForImport("AgentTool", "testapp/models")
	if err != nil {
		t.Fatalf("getModelFieldsForImport: %v", err)
	}

	fieldNames := make(map[string]bool)
	for _, f := range fields {
		fieldNames[f.Name] = true
	}

	// Should include scalar fields
	for _, expected := range []string{"AgentID", "ToolID", "SortOrder"} {
		if !fieldNames[expected] {
			t.Errorf("expected field %q not found", expected)
		}
	}

	// Should exclude struct relation field
	if fieldNames["Tool"] {
		t.Error("struct relation field 'Tool' should be excluded from DTO")
	}

	// Should exclude slice fields
	if fieldNames["Tags"] {
		t.Error("slice field 'Tags' should be excluded from DTO")
	}
}

// TestToSnakeCase_Acronyms tests that toSnakeCase handles common acronyms correctly.
func TestToSnakeCase_Acronyms(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PrimaryAgentID", "primary_agent_id"},
		{"DeploymentURL", "deployment_url"},
		{"WebhookURL", "webhook_url"},
		{"AIModel", "ai_model"},
		{"AppID", "app_id"},
		{"UserID", "user_id"},
		{"HTTPSEnabled", "https_enabled"},
		{"APIKey", "api_key"},
		{"Name", "name"},
		{"CreatedAt", "created_at"},
		{"MaxTokens", "max_tokens"},
		{"SimpleField", "simple_field"},
		{"ID", "id"},
		{"URL", "url"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toSnakeCase(tt.input)
			if got != tt.expected {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestGenerateDTOStructCode_PreservesExactTypes tests that DTO struct generation
// preserves exact numeric and pointer types (#76).
func TestGenerateDTOStructCode_PreservesExactTypes(t *testing.T) {
	fields := []ModelFieldInfo{
		{Name: "MaxTokens", Type: "int64", JSONName: "max_tokens"},
		{Name: "Count", Type: "int32", JSONName: "count"},
		{Name: "BigID", Type: "uint64", JSONName: "big_id"},
		{Name: "SmallID", Type: "uint32", JSONName: "small_id"},
		{Name: "PlannerID", Type: "*uint", JSONName: "planner_id"},
		{Name: "Priority", Type: "*int64", JSONName: "priority"},
		{Name: "Score", Type: "*float64", JSONName: "score"},
		{Name: "Active", Type: "*bool", JSONName: "active"},
		{Name: "Note", Type: "*string", JSONName: "note"},
		{Name: "Name", Type: "string", JSONName: "name"},
		{Name: "Weight", Type: "float64", JSONName: "weight"},
	}

	result := generateDTOStructCode("TestModel", fields)

	expectations := []struct {
		field    string
		typeDecl string
	}{
		{"MaxTokens", "MaxTokens int64"},
		{"Count", "Count int32"},
		{"BigID", "BigID uint64"},
		{"SmallID", "SmallID uint32"},
		{"PlannerID", "PlannerID *uint"},
		{"Priority", "Priority *int64"},
		{"Score", "Score *float64"},
		{"Active", "Active *bool"},
		{"Note", "Note *string"},
		{"Name", "Name string"},
		{"Weight", "Weight float64"},
	}

	for _, e := range expectations {
		if !strings.Contains(result, e.typeDecl) {
			t.Errorf("expected %q in generated DTO, got:\n%s", e.typeDecl, result)
		}
	}
}

// TestGenerateServerAPICode_PointerPassThrough tests that server API code
// uses direct assignment for pointer types since DTO preserves them (#76).
func TestGenerateServerAPICode_PointerPassThrough(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.23\n"), 0644)
	os.MkdirAll("models", 0755)
	os.WriteFile("models/agent.go", []byte(`package models

import "gorm.io/gorm"

type Agent struct {
	gorm.Model
	Name      string
	PlannerID *uint
	MaxTokens int64
	Active    *bool
}
`), 0644)

	modelFieldTypesCache = make(map[string]map[string]string)

	result := generateServerAPICode("Agent", "Agents", "models", "testapp/models")

	// All types should be direct pass-through since DTO preserves exact types
	for _, field := range []string{"PlannerID", "MaxTokens", "Active", "Name"} {
		expected := fmt.Sprintf("%s: item.%s,", field, field)
		if !strings.Contains(result, expected) {
			t.Errorf("expected direct assignment %q in generated code, got:\n%s", expected, result)
		}
	}

	// Update assignments should also be direct
	for _, field := range []string{"PlannerID", "MaxTokens", "Active", "Name"} {
		expected := fmt.Sprintf("model.%s = item.%s", field, field)
		if !strings.Contains(result, expected) {
			t.Errorf("expected direct update assignment %q, got:\n%s", expected, result)
		}
	}
}

// TestGenerateAPIClient_NoGenerics verifies that generated client code does not
// contain generic functions like Post[T] which panic in TinyGo WASM (#78).
func TestGenerateAPIClient_NoGenerics(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.23\n"), 0644)
	os.MkdirAll("guxgen/api", 0755)
	os.MkdirAll("models", 0755)
	os.WriteFile("models/item.go", []byte(`package models
import "gorm.io/gorm"
type Item struct {
	gorm.Model
	Name string
}
`), 0644)

	modelFieldTypesCache = make(map[string]map[string]string)

	models := []CRUDModel{{
		Name:       "Item",
		PluralName: "Items",
		Path:       "items",
	}}

	err := generateAPIClient("testapp/models", "testapp/dto", models)
	if err != nil {
		t.Fatalf("generateAPIClient: %v", err)
	}

	// Check client.go (WASM)
	clientData, err := os.ReadFile("guxgen/api/client.go")
	if err != nil {
		t.Fatalf("reading client.go: %v", err)
	}
	if strings.Contains(string(clientData), "Post[T") {
		t.Error("client.go should not contain generic Post[T] — causes TinyGo WASM panic")
	}

	// Check client_server.go
	serverData, err := os.ReadFile("guxgen/api/client_server.go")
	if err != nil {
		t.Fatalf("reading client_server.go: %v", err)
	}
	if strings.Contains(string(serverData), "Post[T") {
		t.Error("client_server.go should not contain generic Post[T] — causes TinyGo WASM panic")
	}
}

// TestGenerateFormFieldCode_File tests file input generation in forms
func TestGenerateFormFieldCode_File(t *testing.T) {
	field := &ModelField{
		Name:  "Avatar",
		Type:  "string",
		Input: "file",
		Label: "Profile Picture",
	}

	tf := TemplateField{
		Name:         "Avatar",
		Type:         "string",
		DTOFieldName: "Avatar",
		Label:        "Profile Picture",
		StateVar:     "avatarState",
		IsFile:       true,
	}

	code := generateFormFieldCode(field, tf, "user")

	// Check for key components of file upload field
	tests := []struct {
		name   string
		needle string
	}{
		{"FileUpload component", "ui.FileUpload(ui.FileUploadProps{"},
		{"Upload URL", `UploadURL: "/__gux_api/upload"`},
		{"Value function", "Value: func() string {"},
		{"Key to URL conversion", `"/__gux_api/files/" + avatarState.Get()`},
		{"OnUploadComplete handler", "OnUploadComplete: func(result ui.UploadResult) {"},
		{"Set storage key", "avatarState.Set(result.Key)"},
		{"OnRemove handler", "OnRemove: func() {"},
		{"Clear state", `avatarState.Set("")`},
	}

	for _, tc := range tests {
		if !strings.Contains(code, tc.needle) {
			t.Errorf("%s: expected code to contain %q", tc.name, tc.needle)
		}
	}
}

// TestGenerateDetailFieldCode_File tests file input generation in detail views
func TestGenerateDetailFieldCode_File(t *testing.T) {
	field := &ModelField{
		Name:  "Avatar",
		Type:  "string",
		Input: "file",
		Label: "Profile Picture",
	}

	tf := TemplateField{
		Name:         "Avatar",
		Type:         "string",
		DTOFieldName: "Avatar",
		DTOType:      "*core.FileInfo",
		Label:        "Profile Picture",
		IsFile:       true,
	}

	code := generateDetailFieldCode(field, tf, "user")

	// Check for key components of file detail display
	tests := []struct {
		name   string
		needle string
	}{
		{"FileInfo access", "fi := displayItem.Avatar"},
		{"Nil check", "if fi == nil {"},
		{"isImageFile check", "if isImageFile(fi.Filename) {"},
		{"Image preview", "core.Img(core.Attrs{Src: fi.URL"},
		{"Image alt text", `Alt: "Profile Picture"`},
		{"Download link", `Extra: map[string]string{"download": "", "target": "_blank"}`},
		{"Paperclip icon", `core.Text("\U0001F4CE")`},
		{"Filename display", "core.Text(fi.Filename)"},
	}

	for _, tc := range tests {
		if !strings.Contains(code, tc.needle) {
			t.Errorf("%s: expected code to contain %q", tc.name, tc.needle)
		}
	}
}

// TestConvertToTemplateField_File tests that file fields get IsFile flag
func TestConvertToTemplateField_File(t *testing.T) {
	field := &ModelField{
		Name:  "Document",
		Type:  "string",
		Input: "file",
	}

	tf := convertToTemplateField(field, "Contract", nil)

	if !tf.IsFile {
		t.Error("expected IsFile to be true for file input field")
	}

	if tf.DTOType != "*core.FileInfo" {
		t.Errorf("DTOType = %q, want *core.FileInfo", tf.DTOType)
	}

	if tf.StateType != "String" {
		t.Errorf("StateType = %q, want String", tf.StateType)
	}

	// State should store the storage key
	if tf.StateDefault != `""` {
		t.Errorf("StateDefault = %q, want empty string", tf.StateDefault)
	}
}

// TestDTOGeneration_FileField tests that DTOs use *core.FileInfo for file fields
func TestDTOGeneration_FileField(t *testing.T) {
	modelDef := &ModelDefinition{
		Name: "Profile",
		Sections: map[string][]ModelField{
			"Main": {
				{Name: "Name", Type: "string", Table: true},
				{Name: "Avatar", Type: "string", Input: "file", Table: true},
			},
		},
	}

	data := prepareModelTemplateData(modelDef, "testapp", nil, nil, nil)

	if !data.HasFileFields {
		t.Error("expected HasFileFields to be true")
	}

	// Find the Avatar field in TableFields
	var avatarField *TemplateField
	for i := range data.TableFields {
		if data.TableFields[i].Name == "Avatar" {
			avatarField = &data.TableFields[i]
			break
		}
	}

	if avatarField == nil {
		t.Fatal("Avatar field not found in TableFields")
	}

	if avatarField.DTOType != "*core.FileInfo" {
		t.Errorf("Avatar DTOType = %q, want *core.FileInfo", avatarField.DTOType)
	}

	if !avatarField.IsFile {
		t.Error("Avatar field should have IsFile = true")
	}
}

// TestListCellGeneration_FileField tests list view file field rendering
func TestListCellGeneration_FileField(t *testing.T) {
	// The list cell code is generated via template, so we test the template field setup
	field := &ModelField{
		Name:  "Photo",
		Type:  "string",
		Input: "file",
		Table: true,
	}

	tf := convertToTemplateField(field, "Product", nil)

	if !tf.IsFile {
		t.Error("expected IsFile to be true")
	}

	// The template checks .IsFile and accesses .URL and .Filename
	// We verify the DTOType is set correctly for FileInfo access
	if tf.DTOType != "*core.FileInfo" {
		t.Errorf("DTOType = %q, want *core.FileInfo (needed for .URL and .Filename access)", tf.DTOType)
	}
}

// TestFindMainAppFile tests findMainAppFile with default and custom serverDir.
func TestFindMainAppFile(t *testing.T) {
	t.Run("finds app.go in current directory", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "app.go"), []byte("package main"), 0644)

		origDir, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(origDir)

		got, err := findMainAppFile("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "app.go" {
			t.Errorf("got %q, want %q", got, "app.go")
		}
	})

	t.Run("finds main.go in current directory", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)

		origDir, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(origDir)

		got, err := findMainAppFile("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "main.go" {
			t.Errorf("got %q, want %q", got, "main.go")
		}
	})

	t.Run("prefers app.go over main.go", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "app.go"), []byte("package main"), 0644)
		os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)

		origDir, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(origDir)

		got, err := findMainAppFile("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "app.go" {
			t.Errorf("got %q, want %q", got, "app.go")
		}
	})

	t.Run("error when no file found", func(t *testing.T) {
		dir := t.TempDir()

		origDir, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(origDir)

		_, err := findMainAppFile("")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "current directory") {
			t.Errorf("error should mention 'current directory', got: %v", err)
		}
	})

	t.Run("finds app.go in custom serverDir", func(t *testing.T) {
		dir := t.TempDir()
		subDir := filepath.Join(dir, "cmd", "platform")
		os.MkdirAll(subDir, 0755)
		os.WriteFile(filepath.Join(subDir, "app.go"), []byte("package main"), 0644)

		origDir, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(origDir)

		got, err := findMainAppFile("cmd/platform")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join("cmd", "platform", "app.go")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("finds main.go in custom serverDir", func(t *testing.T) {
		dir := t.TempDir()
		subDir := filepath.Join(dir, "cmd", "server")
		os.MkdirAll(subDir, 0755)
		os.WriteFile(filepath.Join(subDir, "main.go"), []byte("package main"), 0644)

		origDir, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(origDir)

		got, err := findMainAppFile("cmd/server")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join("cmd", "server", "main.go")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("error with custom serverDir not found", func(t *testing.T) {
		dir := t.TempDir()
		subDir := filepath.Join(dir, "cmd", "platform")
		os.MkdirAll(subDir, 0755)

		origDir, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(origDir)

		_, err := findMainAppFile("cmd/platform")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "cmd/platform") {
			t.Errorf("error should mention 'cmd/platform', got: %v", err)
		}
	})
}

// TestGenerateAssetsFile_NoWASM tests that SSR-only apps get CSS-only assets.
func TestGenerateAssetsFile_NoWASM(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// Create the styles.css file that will be embedded
	os.MkdirAll("guxgen/dist", 0755)
	os.WriteFile("guxgen/dist/styles.css", []byte("body{}"), 0644)

	err := generateAssetsFile("github.com/test/app", nil, "", false)
	if err != nil {
		t.Fatalf("generateAssetsFile: %v", err)
	}

	// Check embed_gen.go (in guxgen/dist/)
	embedContent, err := os.ReadFile("guxgen/dist/embed_gen.go")
	if err != nil {
		t.Fatalf("read embed_gen.go: %v", err)
	}
	embedCode := string(embedContent)

	if strings.Contains(embedCode, "app.wasm") {
		t.Error("SSR-only embed should not reference app.wasm")
	}
	if strings.Contains(embedCode, "wasm_exec.js") {
		t.Error("SSR-only embed should not reference wasm_exec.js")
	}
	if !strings.Contains(embedCode, "styles.css") {
		t.Error("SSR-only embed should reference styles.css")
	}

	// Check assets_gen.go (init file)
	initContent, err := os.ReadFile("assets_gen.go")
	if err != nil {
		t.Fatalf("read assets_gen.go: %v", err)
	}
	initCode := string(initContent)

	if !strings.Contains(initCode, "dist.StylesCSS") {
		t.Error("SSR-only init should reference dist.StylesCSS")
	}
	if !strings.Contains(initCode, "core.SetDefaultAssets(nil, nil, dist.StylesCSS)") {
		t.Error("SSR-only init should call SetDefaultAssets(nil, nil, dist.StylesCSS)")
	}
	if strings.Contains(initCode, "WasmBinary") {
		t.Error("SSR-only init should not reference WasmBinary")
	}
}

// TestGenerateAssetsFile_WithWASM tests that hybrid apps get full assets.
func TestGenerateAssetsFile_WithWASM(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// Create the dist files
	os.MkdirAll("guxgen/dist", 0755)
	os.WriteFile("guxgen/dist/styles.css", []byte("body{}"), 0644)
	os.WriteFile("guxgen/dist/app.wasm", []byte("wasm"), 0644)
	os.WriteFile("guxgen/dist/wasm_exec.js", []byte("js"), 0644)

	err := generateAssetsFile("github.com/test/app", []string{"app"}, "", true)
	if err != nil {
		t.Fatalf("generateAssetsFile: %v", err)
	}

	// Check embed_gen.go
	embedContent, err := os.ReadFile("guxgen/dist/embed_gen.go")
	if err != nil {
		t.Fatalf("read embed_gen.go: %v", err)
	}
	embedCode := string(embedContent)

	if !strings.Contains(embedCode, "app.wasm") {
		t.Error("hybrid embed should reference app.wasm")
	}
	if !strings.Contains(embedCode, "wasm_exec.js") {
		t.Error("hybrid embed should reference wasm_exec.js")
	}
	if !strings.Contains(embedCode, "styles.css") {
		t.Error("hybrid embed should reference styles.css")
	}

	// Check assets_gen.go
	initContent, err := os.ReadFile("assets_gen.go")
	if err != nil {
		t.Fatalf("read assets_gen.go: %v", err)
	}
	initCode := string(initContent)

	if !strings.Contains(initCode, "dist.WasmBinary") {
		t.Error("hybrid init should reference dist.WasmBinary")
	}
	if !strings.Contains(initCode, "dist.WasmExecJS") {
		t.Error("hybrid init should reference dist.WasmExecJS")
	}
	if !strings.Contains(initCode, "core.SetDefaultAssets(dist.WasmBinary, dist.WasmExecJS, dist.StylesCSS)") {
		t.Error("hybrid init should call SetDefaultAssets with all dist exports")
	}
}

// TestGenerateAssetsFile_CustomServerDir tests that custom serverDir places assets_gen.go correctly.
func TestGenerateAssetsFile_CustomServerDir(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.MkdirAll("guxgen/dist", 0755)
	os.MkdirAll("cmd/platform", 0755)
	os.WriteFile("guxgen/dist/styles.css", []byte("body{}"), 0644)
	os.WriteFile("guxgen/dist/app.wasm", []byte("wasm"), 0644)
	os.WriteFile("guxgen/dist/wasm_exec.js", []byte("js"), 0644)

	err := generateAssetsFile("github.com/test/app", []string{"app"}, "cmd/platform", true)
	if err != nil {
		t.Fatalf("generateAssetsFile: %v", err)
	}

	// embed_gen.go should be at guxgen/dist/ (no .. paths)
	embedContent, err := os.ReadFile("guxgen/dist/embed_gen.go")
	if err != nil {
		t.Fatalf("read embed_gen.go: %v", err)
	}
	embedCode := string(embedContent)

	if strings.Contains(embedCode, "..") {
		t.Error("embed_gen.go should not contain .. paths")
	}

	// assets_gen.go should be at cmd/platform/
	initContent, err := os.ReadFile("cmd/platform/assets_gen.go")
	if err != nil {
		t.Fatalf("read cmd/platform/assets_gen.go: %v", err)
	}
	initCode := string(initContent)

	// Should use import, not //go:embed
	if strings.Contains(initCode, "//go:embed") {
		t.Error("assets_gen.go should import dist package, not use //go:embed")
	}
	if !strings.Contains(initCode, "github.com/test/app/guxgen/dist") {
		t.Error("assets_gen.go should import the guxgen/dist package")
	}
}

// TestParseRoutesWithImportFollowing tests that routes in imported packages are discovered.
func TestParseRoutesWithImportFollowing(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// Create go.mod
	os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.21\n"), 0644)

	// Create entry point that imports a local package
	os.MkdirAll("cmd/server", 0755)
	os.WriteFile("cmd/server/main.go", []byte(`package main

import (
	"testapp/internal/platform"
)

func main() {
	platform.Setup()
}
`), 0644)

	// Create the imported package with Hybrid routes
	os.MkdirAll("internal/platform", 0755)
	os.WriteFile("internal/platform/app.go", []byte(`package platform

import (
	"testapp/pages"
	"github.com/dougbarrett/gux/core"
)

func Setup() {
	app := core.New()
	app.Routes().
		Hybrid("/", pages.Home).
		Hybrid("/about", pages.About)
}
`), 0644)

	// Parse with import following
	bundles, _, err := parseRoutesWithImportFollowing("cmd/server/main.go")
	if err != nil {
		t.Fatalf("parseRoutesWithImportFollowing: %v", err)
	}

	// Should find the routes from the imported package
	appBundle, ok := bundles["app"]
	if !ok {
		t.Fatal("expected 'app' bundle")
	}
	if len(appBundle.Routes) != 2 {
		t.Errorf("expected 2 routes, got %d", len(appBundle.Routes))
	}
}

// TestParseRoutesWithImportFollowing_DirectRoutes tests that direct routes skip import following.
func TestParseRoutesWithImportFollowing_DirectRoutes(t *testing.T) {
	src := `package main

import (
	"myapp/pages"
	"github.com/dougbarrett/gux/core"
)

func main() {
	app := core.New()
	app.Routes().Hybrid("/", pages.Home)
}
`
	path := writeTestFile(t, src)

	bundles, _, err := parseRoutesWithImportFollowing(path)
	if err != nil {
		t.Fatalf("parseRoutesWithImportFollowing: %v", err)
	}

	appBundle := bundles["app"]
	if len(appBundle.Routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(appBundle.Routes))
	}
	if appBundle.Routes[0].Path != "/" {
		t.Errorf("expected route path '/', got %q", appBundle.Routes[0].Path)
	}
}

// TestParseAPIEndpointsWithImportFollowing tests that API endpoints in imported packages are discovered.
func TestParseAPIEndpointsWithImportFollowing(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// Create go.mod
	os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.21\n"), 0644)

	// Create entry point that imports a local package
	os.MkdirAll("cmd/server", 0755)
	os.WriteFile("cmd/server/main.go", []byte(`package main

import (
	"testapp/internal/routes"
)

func main() {
	routes.Setup()
}
`), 0644)

	// Create the imported package with API endpoint registrations
	os.MkdirAll("internal/routes", 0755)
	os.WriteFile("internal/routes/api.go", []byte(`package routes

import (
	"testapp/dto"
	"github.com/dougbarrett/gux/core"
)

func Setup() {
	app := core.New()
	core.API(app, "POST", "/api/login", func(ctx *core.APIContext, req dto.LoginRequest) (dto.LoginResponse, error) {
		return dto.LoginResponse{}, nil
	})
	core.APIGet(app, "/api/users/:id", func(ctx *core.APIContext) (dto.UserDetail, error) {
		return dto.UserDetail{}, nil
	})
}
`), 0644)

	// Parse with import following
	endpoints, dtoImports, err := parseAPIEndpointsWithImportFollowing("cmd/server/main.go")
	if err != nil {
		t.Fatalf("parseAPIEndpointsWithImportFollowing: %v", err)
	}

	// Should find the endpoints from the imported package
	if len(endpoints) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(endpoints))
	}

	// Verify first endpoint
	if endpoints[0].Method != "POST" {
		t.Errorf("expected POST, got %s", endpoints[0].Method)
	}
	if endpoints[0].Path != "/api/login" {
		t.Errorf("expected /api/login, got %s", endpoints[0].Path)
	}
	if endpoints[0].RequestType != "LoginRequest" {
		t.Errorf("expected LoginRequest, got %s", endpoints[0].RequestType)
	}
	if endpoints[0].ResponseType != "LoginResponse" {
		t.Errorf("expected LoginResponse, got %s", endpoints[0].ResponseType)
	}

	// Verify second endpoint
	if endpoints[1].Method != "GET" {
		t.Errorf("expected GET, got %s", endpoints[1].Method)
	}
	if endpoints[1].Path != "/api/users/:id" {
		t.Errorf("expected /api/users/:id, got %s", endpoints[1].Path)
	}

	// Verify DTO imports were collected
	if _, ok := dtoImports["dto"]; !ok {
		t.Errorf("expected 'dto' in dtoImports, got: %v", dtoImports)
	}
}

// TestParseAPIEndpointsWithImportFollowing_DirectEndpoints tests that direct endpoints skip import following.
func TestParseAPIEndpointsWithImportFollowing_DirectEndpoints(t *testing.T) {
	src := `package main

import (
	"myapp/dto"
	"github.com/dougbarrett/gux/core"
)

func main() {
	app := core.New()
	core.APIGet(app, "/api/health", func(ctx *core.APIContext) (dto.HealthResponse, error) {
		return dto.HealthResponse{}, nil
	})
}
`
	path := writeTestFile(t, src)

	endpoints, _, err := parseAPIEndpointsWithImportFollowing(path)
	if err != nil {
		t.Fatalf("parseAPIEndpointsWithImportFollowing: %v", err)
	}

	if len(endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(endpoints))
	}
	if endpoints[0].Path != "/api/health" {
		t.Errorf("expected /api/health, got %s", endpoints[0].Path)
	}
}

// TestParseRoutesAndBundles_RootRouteInRouteGroup verifies that Hybrid("/", ...) inside a
// RouteGroup produces the group prefix, not an empty string (#80).
func TestParseRoutesAndBundles_RootRouteInRouteGroup(t *testing.T) {
	src := `package main

import (
	"myapp/pages"
	"myapp/admin"
	"github.com/dougbarrett/gux/core"
)

func main() {
	app := core.New()

	// Root route without RouteGroup
	app.Routes().
		Hybrid("/", pages.Home).
		Hybrid("/about", pages.About)

	// Root route inside RouteGroup
	app.RouteGroup("/admin", core.WithBundle("admin")).
		Hybrid("/", admin.Dashboard).
		Hybrid("/users", admin.Users)
}
`
	path := writeTestFile(t, src)

	bundles, _, err := parseRoutesAndBundles(path)
	if err != nil {
		t.Fatalf("parseRoutesAndBundles: %v", err)
	}

	// Check "app" bundle — root route should be "/"
	appBundle := bundles["app"]
	if appBundle == nil {
		t.Fatal("expected 'app' bundle")
	}
	for _, r := range appBundle.Routes {
		if r.Path == "" {
			t.Errorf("app bundle has empty path route; expected '/' not ''")
		}
	}
	// Verify "/" route exists
	foundRoot := false
	for _, r := range appBundle.Routes {
		if r.Path == "/" {
			foundRoot = true
		}
	}
	if !foundRoot {
		t.Errorf("app bundle missing '/' route; routes: %+v", appBundle.Routes)
	}

	// Check "admin" bundle — root route should be "/admin", not ""
	adminBundle := bundles["admin"]
	if adminBundle == nil {
		t.Fatal("expected 'admin' bundle")
	}
	for _, r := range adminBundle.Routes {
		if r.Path == "" {
			t.Errorf("admin bundle has empty path route; expected '/admin' not ''")
		}
	}
	// Verify "/admin" route exists (from Hybrid("/", ...) in RouteGroup("/admin", ...))
	foundAdminRoot := false
	for _, r := range adminBundle.Routes {
		if r.Path == "/admin" {
			foundAdminRoot = true
		}
	}
	if !foundAdminRoot {
		t.Errorf("admin bundle missing '/admin' route; routes: %+v", adminBundle.Routes)
	}
}

// TestGenerateBundleWasmEntryPoint_RootRoute verifies that the generated WASM code
// uses "/" not "" for the root route path (#80).
func TestGenerateBundleWasmEntryPoint_RootRoute(t *testing.T) {
	t.Run("normal root path", func(t *testing.T) {
		dir := t.TempDir()
		origDir, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(origDir)

		bundle := &BundleInfo{
			Name: "app",
			Routes: []PageRoute{
				{Path: "/", Handler: "pages.Home", IsHybrid: true, Bundle: "app"},
				{Path: "/about", Handler: "pages.About", IsHybrid: true, Bundle: "app"},
			},
			Imports: []BundleImport{
				{Path: "myapp/pages"},
			},
		}

		err := generateBundleWasmEntryPoint("app", bundle, "")
		if err != nil {
			t.Fatalf("generateBundleWasmEntryPoint: %v", err)
		}

		content, err := os.ReadFile("guxgen/wasm/main.go")
		if err != nil {
			t.Fatalf("read generated file: %v", err)
		}

		code := string(content)

		if strings.Contains(code, `{exact: ""}`) {
			t.Errorf("generated code contains {exact: \"\"} — should be {exact: \"/\"}")
		}
		if !strings.Contains(code, `{exact: "/"}`) {
			t.Errorf("generated code missing {exact: \"/\"}")
		}
		if strings.Contains(code, `case "":`) {
			t.Errorf("generated code contains case \"\": — should be case \"/\":")
		}
		if !strings.Contains(code, `case "/":`) {
			t.Errorf("generated code missing case \"/\":")
		}
	})

	t.Run("empty path becomes root slash", func(t *testing.T) {
		dir := t.TempDir()
		origDir, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(origDir)

		// Simulate a route with empty path (e.g., from RouteGroup with empty prefix)
		bundle := &BundleInfo{
			Name: "app",
			Routes: []PageRoute{
				{Path: "", Handler: "pages.Home", IsHybrid: true, Bundle: "app"},
				{Path: "/about", Handler: "pages.About", IsHybrid: true, Bundle: "app"},
			},
			Imports: []BundleImport{
				{Path: "myapp/pages"},
			},
		}

		err := generateBundleWasmEntryPoint("app", bundle, "")
		if err != nil {
			t.Fatalf("generateBundleWasmEntryPoint: %v", err)
		}

		content, err := os.ReadFile("guxgen/wasm/main.go")
		if err != nil {
			t.Fatalf("read generated file: %v", err)
		}

		code := string(content)

		// Empty path must be normalized to "/"
		if strings.Contains(code, `{exact: ""}`) {
			t.Errorf("generated code contains {exact: \"\"} — empty path should be normalized to \"/\"")
		}
		if !strings.Contains(code, `{exact: "/"}`) {
			t.Errorf("generated code missing {exact: \"/\"}")
		}
		if strings.Contains(code, `case "":`) {
			t.Errorf("generated code contains case \"\": — empty path should be normalized to \"/\"")
		}
		if !strings.Contains(code, `case "/":`) {
			t.Errorf("generated code missing case \"/\":")
		}
	})
}

// TestParseRoutesAndBundles_RouteGroupEmptyPrefix verifies that RouteGroup with
// empty prefix doesn't produce empty route paths (#80).
func TestParseRoutesAndBundles_RouteGroupEmptyPrefix(t *testing.T) {
	src := `package main

import (
	"myapp/pages"
	"github.com/dougbarrett/gux/core"
)

func main() {
	app := core.New()
	app.RouteGroup("", core.WithBundle("app")).
		Hybrid("/", pages.Home).
		Hybrid("/about", pages.About)
}
`
	path := writeTestFile(t, src)

	bundles, _, err := parseRoutesAndBundles(path)
	if err != nil {
		t.Fatalf("parseRoutesAndBundles: %v", err)
	}

	appBundle := bundles["app"]
	if appBundle == nil {
		t.Fatal("expected 'app' bundle")
	}

	for _, r := range appBundle.Routes {
		if r.Path == "" {
			t.Errorf("route has empty path; should be '/' not ''")
		}
	}

	// Verify root route is "/"
	foundRoot := false
	for _, r := range appBundle.Routes {
		if r.Path == "/" {
			foundRoot = true
		}
	}
	if !foundRoot {
		t.Errorf("missing '/' route; routes: %+v", appBundle.Routes)
	}
}

// --- Issue 81 tests ---

func TestGenerateFuncName_HyphenInPath(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   string
	}{
		{"GET", "/api/prebaked-tools", "GetPrebakedTools"},
		{"POST", "/api/user-settings", "UserSettings"},
		{"GET", "/api/my-long-path-name", "GetMyLongPathName"},
		{"DELETE", "/api/cache-entries/:id", "DeleteCacheEntrie"},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			got := generateFuncName(tt.method, tt.path)
			if got != tt.want {
				t.Errorf("generateFuncName(%q, %q) = %q, want %q", tt.method, tt.path, got, tt.want)
			}
		})
	}
}

func TestGenerateFuncName_NoTruncatePlural(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   string
	}{
		// "Status" ends in 's' but should NOT be truncated since :id is not the last segment
		{"GET", "/api/apps/:id/status", "GetAppsStatus"},
		// "users/:id" should singularize because :id IS the last segment
		{"GET", "/api/users/:id", "GetUser"},
		// "posts/:id" should singularize
		{"DELETE", "/api/posts/:id", "DeletePost"},
		// No param at end — keep plural
		{"GET", "/api/apps/:id/logs", "GetAppsLogs"},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			got := generateFuncName(tt.method, tt.path)
			if got != tt.want {
				t.Errorf("generateFuncName(%q, %q) = %q, want %q", tt.method, tt.path, got, tt.want)
			}
		})
	}
}

func TestGenerateEndpointFunc_NoStrayComma(t *testing.T) {
	tests := []struct {
		name string
		ep   APIEndpointInfo
	}{
		{
			name: "DELETE no params",
			ep: APIEndpointInfo{
				Method:   "DELETE",
				Path:     "/api/logout",
				FuncName: "Logout",
			},
		},
		{
			name: "POST no params no types",
			ep: APIEndpointInfo{
				Method:       "POST",
				Path:         "/api/logout",
				FuncName:     "Logout",
				ResponseType: "",
			},
		},
		{
			name: "GET no params with response",
			ep: APIEndpointInfo{
				Method:       "GET",
				Path:         "/api/tools",
				FuncName:     "GetTools",
				ResponseType: "[]string",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := generateEndpointFunc(tt.ep)
			if strings.Contains(code, "(, ") {
				t.Errorf("generated code contains stray comma: %s", code)
			}
		})
	}
}

func TestExtractMainPackageTypes_FromSourceFile(t *testing.T) {
	dir := t.TempDir()

	// Write a Go file with type definitions
	typesFile := filepath.Join(dir, "types.go")
	os.WriteFile(typesFile, []byte(`package main

type LoginRequest struct {
	Email    string `+"`"+`json:"email"`+"`"+`
	Password string `+"`"+`json:"password"`+"`"+`
}

type LoginResponse struct {
	Success bool   `+"`"+`json:"success"`+"`"+`
	Token   string `+"`"+`json:"token"`+"`"+`
}
`), 0644)

	typeNames := map[string]bool{
		"LoginRequest":  true,
		"LoginResponse": true,
	}

	result, err := extractMainPackageTypes(typesFile, typeNames)
	if err != nil {
		t.Fatalf("extractMainPackageTypes failed: %v", err)
	}
	if !strings.Contains(result, "LoginRequest") {
		t.Error("expected LoginRequest in extracted types")
	}
	if !strings.Contains(result, "LoginResponse") {
		t.Error("expected LoginResponse in extracted types")
	}
}

// TestGenerateLoadTracker verifies the load tracker files are generated correctly.
func TestGenerateLoadTracker(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.MkdirAll("guxgen/api", 0755)

	err := generateLoadTracker()
	if err != nil {
		t.Fatalf("generateLoadTracker: %v", err)
	}

	t.Run("wasm file exists and has correct content", func(t *testing.T) {
		data, err := os.ReadFile("guxgen/api/load_tracker.go")
		if err != nil {
			t.Fatalf("reading load_tracker.go: %v", err)
		}
		code := string(data)

		checks := []struct {
			name   string
			needle string
		}{
			{"build tag", "//go:build js && wasm"},
			{"package", "package api"},
			{"StartLoad func", "func StartLoad() int"},
			{"EndLoad func", "func EndLoad(epoch int)"},
			{"ResetLoads func", "func ResetLoads()"},
			{"HasPendingLoads func", "func HasPendingLoads() bool"},
			{"OnAllLoadsComplete var", "OnAllLoadsComplete func()"},
			{"epoch increment in StartLoad", "pendingLoads++"},
			{"epoch check in EndLoad", "if epoch != loadEpoch"},
			{"counter decrement", "pendingLoads--"},
			{"fires callback", "OnAllLoadsComplete()"},
		}
		for _, c := range checks {
			if !strings.Contains(code, c.needle) {
				t.Errorf("%s: expected %q in load_tracker.go", c.name, c.needle)
			}
		}
	})

	t.Run("server file exists and has no-op stubs", func(t *testing.T) {
		data, err := os.ReadFile("guxgen/api/load_tracker_server.go")
		if err != nil {
			t.Fatalf("reading load_tracker_server.go: %v", err)
		}
		code := string(data)

		checks := []struct {
			name   string
			needle string
		}{
			{"build tag", "//go:build !js || !wasm"},
			{"package", "package api"},
			{"StartLoad stub", "func StartLoad() int"},
			{"EndLoad stub", "func EndLoad(epoch int)"},
			{"ResetLoads stub", "func ResetLoads()"},
			{"HasPendingLoads stub", "func HasPendingLoads() bool"},
			{"OnAllLoadsComplete var", "var OnAllLoadsComplete func()"},
			{"returns false", "return false"},
		}
		for _, c := range checks {
			if !strings.Contains(code, c.needle) {
				t.Errorf("%s: expected %q in load_tracker_server.go", c.name, c.needle)
			}
		}
	})
}

// TestGenerateAPIClient_LoadTracking verifies StartLoad/EndLoad wraps async calls.
func TestGenerateAPIClient_LoadTracking(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.WriteFile("go.mod", []byte("module testapp\n\ngo 1.23\n"), 0644)
	os.MkdirAll("guxgen/api", 0755)
	os.MkdirAll("models", 0755)
	os.WriteFile("models/item.go", []byte(`package models
import "gorm.io/gorm"
type Item struct {
	gorm.Model
	Name string
}
`), 0644)

	modelFieldTypesCache = make(map[string]map[string]string)

	models := []CRUDModel{{
		Name:       "Item",
		PluralName: "Items",
		Path:       "items",
	}}

	err := generateAPIClient("testapp/models", "testapp/dto", models)
	if err != nil {
		t.Fatalf("generateAPIClient: %v", err)
	}

	clientData, err := os.ReadFile("guxgen/api/client.go")
	if err != nil {
		t.Fatalf("reading client.go: %v", err)
	}
	code := string(clientData)

	// Verify StartLoad/EndLoad pattern appears for each CRUD method
	startLoadCount := strings.Count(code, "epoch := StartLoad()")
	endLoadCount := strings.Count(code, "defer EndLoad(epoch)")

	if startLoadCount < 6 {
		t.Errorf("expected at least 6 StartLoad() calls (one per CRUD method), got %d", startLoadCount)
	}
	if endLoadCount < 6 {
		t.Errorf("expected at least 6 defer EndLoad(epoch) calls, got %d", endLoadCount)
	}
	if startLoadCount != endLoadCount {
		t.Errorf("StartLoad count (%d) != EndLoad count (%d)", startLoadCount, endLoadCount)
	}

	// Verify StartLoad is before go func and EndLoad is inside go func
	if !strings.Contains(code, "epoch := StartLoad()\n\tgo func() {\n\t\tdefer EndLoad(epoch)") {
		t.Error("StartLoad/EndLoad not in expected positions relative to go func()")
	}
}

// TestGenerateEndpointFunc_LoadTracking verifies StartLoad/EndLoad wraps endpoint async calls.
func TestGenerateEndpointFunc_LoadTracking(t *testing.T) {
	tests := []struct {
		name string
		ep   APIEndpointInfo
	}{
		{
			name: "GET with response",
			ep: APIEndpointInfo{
				Method:       "GET",
				Path:         "/api/items",
				FuncName:     "GetItems",
				ResponseType: "[]Item",
			},
		},
		{
			name: "POST with request and response",
			ep: APIEndpointInfo{
				Method:       "POST",
				Path:         "/api/items",
				FuncName:     "CreateItem",
				RequestType:  "CreateItemRequest",
				ResponseType: "Item",
			},
		},
		{
			name: "DELETE no response",
			ep: APIEndpointInfo{
				Method:   "DELETE",
				Path:     "/api/items/:id",
				FuncName: "DeleteItem",
				PathParams: []string{"id"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := generateEndpointFunc(tt.ep)

			if !strings.Contains(code, "epoch := StartLoad()") {
				t.Errorf("endpoint func missing StartLoad()")
			}
			if !strings.Contains(code, "defer EndLoad(epoch)") {
				t.Errorf("endpoint func missing defer EndLoad(epoch)")
			}
			// StartLoad should be before go func
			startIdx := strings.Index(code, "epoch := StartLoad()")
			goFuncIdx := strings.Index(code, "go func()")
			if startIdx > goFuncIdx {
				t.Errorf("StartLoad() should appear before go func()")
			}
		})
	}
}

// TestGenerateBundleWasmEntryPoint_WithAPIImport verifies load tracking integration in WASM main.
func TestGenerateBundleWasmEntryPoint_WithAPIImport(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	bundle := &BundleInfo{
		Name: "app",
		Routes: []PageRoute{
			{Path: "/", Handler: "pages.Home", IsHybrid: true, Bundle: "app"},
			{Path: "/about", Handler: "pages.About", IsHybrid: true, Bundle: "app"},
		},
		Imports: []BundleImport{
			{Path: "myapp/pages"},
		},
	}

	err := generateBundleWasmEntryPoint("app", bundle, "myapp/guxgen/api")
	if err != nil {
		t.Fatalf("generateBundleWasmEntryPoint: %v", err)
	}

	content, err := os.ReadFile("guxgen/wasm/main.go")
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	code := string(content)

	t.Run("imports api package", func(t *testing.T) {
		if !strings.Contains(code, `"myapp/guxgen/api"`) {
			t.Error("generated code should import api package when apiImport is provided")
		}
	})

	t.Run("has attachLinks function", func(t *testing.T) {
		if !strings.Contains(code, "attachLinks := func()") {
			t.Error("generated code should have attachLinks closure when apiImport is provided")
		}
	})

	t.Run("render calls attachLinks", func(t *testing.T) {
		if !strings.Contains(code, "attachLinks()") {
			t.Error("render should call attachLinks() when apiImport is provided")
		}
	})

	t.Run("navigate uses api.ResetLoads", func(t *testing.T) {
		if !strings.Contains(code, "api.ResetLoads()") {
			t.Error("navigate should call api.ResetLoads() when apiImport is provided")
		}
	})

	t.Run("navigate uses api.HasPendingLoads", func(t *testing.T) {
		if !strings.Contains(code, "api.HasPendingLoads()") {
			t.Error("navigate should check api.HasPendingLoads() when apiImport is provided")
		}
	})

	t.Run("initial boot wires OnAllLoadsComplete", func(t *testing.T) {
		if !strings.Contains(code, "api.OnAllLoadsComplete = func()") {
			t.Error("initial boot should wire api.OnAllLoadsComplete when apiImport is provided")
		}
	})

	t.Run("no fetchLoader function", func(t *testing.T) {
		if strings.Contains(code, "func fetchLoader(") {
			t.Error("generated code should NOT have fetchLoader when apiImport is provided")
		}
	})

	t.Run("initial boot renders conditionally on HasPendingLoads (#84)", func(t *testing.T) {
		// Check that initial boot has the HasPendingLoads conditional pattern
		if !strings.Contains(code, "api.OnAllLoadsComplete = func() { render() }\n\tloadPage()\n\tif !api.HasPendingLoads()") {
			t.Error("boot section should wire OnAllLoadsComplete, call loadPage(), then check HasPendingLoads()")
		}
		// attachLinks() should be in the else branch (async loads pending)
		if !strings.Contains(code, "attachLinks()\n\t}") {
			t.Error("boot section should call attachLinks() in else branch for async case")
		}
	})

	t.Run("attachLinks defined before render (#83)", func(t *testing.T) {
		attachLinksIdx := strings.Index(code, "attachLinks := func()")
		renderIdx := strings.Index(code, "render = func()")
		if attachLinksIdx < 0 || renderIdx < 0 {
			t.Fatal("could not find attachLinks or render definition")
		}
		if attachLinksIdx > renderIdx {
			t.Error("attachLinks must be defined before render to avoid forward reference")
		}
	})

	t.Run("popstate does not declare unused path variable (#83)", func(t *testing.T) {
		popstateIdx := strings.Index(code, "popstate")
		if popstateIdx < 0 {
			t.Fatal("could not find popstate handler")
		}
		// Find the popstate handler body (up to "return nil")
		popstateSection := code[popstateIdx:]
		returnIdx := strings.Index(popstateSection, "return nil")
		popstateBody := popstateSection[:returnIdx]
		if strings.Contains(popstateBody, "path :=") {
			t.Error("popstate handler should not declare path variable (unused)")
		}
	})
}

// TestGenerateBundleWasmEntryPoint_WithoutAPIImport verifies fetchLoader is kept when no API.
func TestGenerateBundleWasmEntryPoint_WithoutAPIImport(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	bundle := &BundleInfo{
		Name: "app",
		Routes: []PageRoute{
			{Path: "/", Handler: "pages.Home", IsHybrid: true, Bundle: "app"},
		},
		Imports: []BundleImport{
			{Path: "myapp/pages"},
		},
	}

	err := generateBundleWasmEntryPoint("app", bundle, "")
	if err != nil {
		t.Fatalf("generateBundleWasmEntryPoint: %v", err)
	}

	content, err := os.ReadFile("guxgen/wasm/main.go")
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	code := string(content)

	t.Run("has fetchLoader", func(t *testing.T) {
		if !strings.Contains(code, "func fetchLoader(") {
			t.Error("generated code should have fetchLoader when no apiImport")
		}
	})

	t.Run("no api import", func(t *testing.T) {
		if strings.Contains(code, "guxgen/api") {
			t.Error("generated code should NOT import api package when apiImport is empty")
		}
	})

	t.Run("no api calls", func(t *testing.T) {
		if strings.Contains(code, "api.ResetLoads") {
			t.Error("generated code should NOT call api.ResetLoads when apiImport is empty")
		}
		if strings.Contains(code, "api.HasPendingLoads") {
			t.Error("generated code should NOT call api.HasPendingLoads when apiImport is empty")
		}
		if strings.Contains(code, "api.OnAllLoadsComplete") {
			t.Error("generated code should NOT reference api.OnAllLoadsComplete when apiImport is empty")
		}
	})

	t.Run("no attachLinks function", func(t *testing.T) {
		if strings.Contains(code, "attachLinks := func()") {
			t.Error("generated code should NOT have attachLinks when apiImport is empty")
		}
	})

	t.Run("navigate uses fetchLoader", func(t *testing.T) {
		if !strings.Contains(code, "fetchLoader(path, func(") {
			t.Error("navigate should use fetchLoader when apiImport is empty")
		}
	})
}
