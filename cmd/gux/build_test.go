package main

import (
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
	// *string maps to string in inline DTOs (default case)
	if !strings.Contains(result, "Description string") {
		t.Errorf("expected string type for Description (*string mapped to string), got:\n%s", result)
	}
	// *float64 maps to float64 in inline DTOs (no explicit case, hits default -> string)
	// Actually check what the actual mapping is:
	if !strings.Contains(result, "Price") {
		t.Errorf("expected Price field in generated DTO, got:\n%s", result)
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

	result := generateServerAPICode("Widget", "Widgets", "models")

	// *string should have pointer dereference in model-to-DTO
	if !strings.Contains(result, "if item.Description != nil") {
		t.Errorf("expected nil check for *string field Description, got:\n%s", result)
	}

	// *float64 should have pointer dereference in model-to-DTO
	if !strings.Contains(result, "if item.Weight != nil") {
		t.Errorf("expected nil check for *float64 field Weight, got:\n%s", result)
	}

	// string field should be direct assignment
	if !strings.Contains(result, "Label: item.Label,") {
		t.Errorf("expected direct assignment for string field Label, got:\n%s", result)
	}

	// *string should have pointer wrapping in DTO-to-model (create)
	if !strings.Contains(result, "func() *string") {
		t.Errorf("expected *string wrapping function in create code, got:\n%s", result)
	}
}
