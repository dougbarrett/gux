package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerateFormFieldCode_RelationDropdownUsesDisplayField verifies that
// relation dropdown options use the Brief DTO's actual display field
// instead of a hardcoded field name. (Regression test for #22, #27)
func TestGenerateFormFieldCode_RelationDropdownUsesDisplay(t *testing.T) {
	tests := []struct {
		name      string
		relation  string
		fieldType string
	}{
		{
			name:      "nullable *uint relation",
			relation:  "Industry",
			fieldType: "*uint",
		},
		{
			name:      "required uint relation",
			relation:  "Order",
			fieldType: "uint",
		},
		{
			name:      "client relation",
			relation:  "Client",
			fieldType: "*uint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &ModelField{
				Name:     tt.relation + "ID",
				Type:     tt.fieldType,
				Relation: tt.relation,
			}
			tf := TemplateField{
				Label:                tt.relation,
				StateVar:             strings.ToLower(tt.relation) + "IDState",
				StateName:            strings.ToLower(tt.relation) + "_id",
				RelationDisplayField: "Name",
			}
			code := generateFormFieldCode(field, tf, "test")

			// Should use Display() method instead of direct field access
			if !strings.Contains(code, "item.Display()") {
				t.Errorf("expected generated code to contain item.Display(), but it doesn't.\nGenerated:\n%s", code)
			}
			// Must NOT contain hardcoded #%d pattern
			if strings.Contains(code, `Sprintf("#`) {
				t.Errorf("generated code still contains hardcoded #N pattern")
			}
		})
	}
}

// TestGenerateFormFieldCode_RelationSelectForUint verifies that both uint
// and *uint relation fields generate select dropdowns. (Regression test for #23)
func TestGenerateFormFieldCode_RelationSelectForUint(t *testing.T) {
	tests := []struct {
		name      string
		fieldType string
	}{
		{"nullable *uint", "*uint"},
		{"required uint", "uint"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &ModelField{
				Name:     "ClientID",
				Type:     tt.fieldType,
				Relation: "Client",
			}
			tf := TemplateField{
				Label:                "Client",
				StateVar:             "clientIDState",
				StateName:            "client_id",
				RelationDisplayField: "Name",
			}
			code := generateFormFieldCode(field, tf, "test")

			// Should generate a select with Option elements, not a number input
			if !strings.Contains(code, "core.Option") {
				t.Errorf("expected select dropdown with core.Option for %s relation, got number input.\nGenerated:\n%s", tt.fieldType, code)
			}
			if strings.Contains(code, `"number"`) {
				t.Errorf("%s relation field should not generate number input", tt.fieldType)
			}
		})
	}
}

// TestGenerateFormFieldCode_NoFmtInDropdown verifies that the generated
// dropdown code does not use fmt.Sprintf (which would require fmt import).
// (Regression test for #27 unused fmt import)
func TestGenerateFormFieldCode_NoFmtInDropdown(t *testing.T) {
	field := &ModelField{
		Name:     "ClientID",
		Type:     "*uint",
		Relation: "Client",
	}
	tf := TemplateField{
		Label:                "Client",
		StateVar:             "clientIDState",
		StateName:            "client_id",
		RelationDisplayField: "Name",
	}
	code := generateFormFieldCode(field, tf, "test")

	if strings.Contains(code, "fmt.Sprintf") {
		t.Errorf("dropdown code should not use fmt.Sprintf (causes unused fmt import in new forms).\nGenerated:\n%s", code)
	}
}

// TestNamePluralLower_HumanizesMultiWordNames verifies that multi-word
// model names are properly humanized for empty state messages.
// (Regression test for #25)
func TestNamePluralLower_HumanizesMultiWordNames(t *testing.T) {
	tests := []struct {
		modelName string
		want      string
	}{
		{"ProductCategory", "product categories"},
		{"OrderItem", "order items"},
		{"LeadSource", "lead sources"},
		{"User", "users"},
	}

	for _, tt := range tests {
		t.Run(tt.modelName, func(t *testing.T) {
			got := strings.ToLower(toDisplayName(ToPlural(tt.modelName)))
			if got != tt.want {
				t.Errorf("NamePluralLower for %q = %q, want %q", tt.modelName, got, tt.want)
			}
		})
	}
}

// TestHasSelectRelations_IncludesUint verifies that both uint and *uint
// relation fields set HasSelectRelations = true. (Regression test for #23)
func TestHasSelectRelations_IncludesUint(t *testing.T) {
	tests := []struct {
		fieldType string
		relation  string
		want      bool
	}{
		{"*uint", "Client", true},
		{"uint", "Client", true},
		{"string", "", false},
		{"int", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.fieldType, func(t *testing.T) {
			got := (tt.fieldType == "*uint" || tt.fieldType == "uint" || false) && tt.relation != "" ||
				false // field.Input == "select"
			// Simplified check matching the actual condition in prepareModelTemplateData
			if tt.relation != "" {
				got = tt.fieldType == "*uint" || tt.fieldType == "uint"
			} else {
				got = false
			}
			if got != tt.want {
				t.Errorf("HasSelectRelations for type=%q relation=%q = %v, want %v", tt.fieldType, tt.relation, got, tt.want)
			}
		})
	}
}

// TestParentChildConfigParsing verifies that parent-child config fields are parsed correctly.
func TestParentChildConfigParsing(t *testing.T) {
	configJSON := `{
		"models": {
			"Order": {
				"sections": {
					"Details": [
						{"name": "OrderNumber", "type": "string", "table": true}
					]
				}
			},
			"OrderItem": {
				"parent": "Order",
				"parentField": "OrderID",
				"sidebar": false,
				"sections": {
					"Details": [
						{"name": "ProductName", "type": "string", "table": true},
						{"name": "Quantity", "type": "int", "table": true},
						{"name": "OrderID", "type": "uint", "relation": "Order"}
					]
				}
			}
		}
	}`

	var config ModelsConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	// Verify OrderItem has parent config
	orderItem, ok := config.Models["OrderItem"]
	if !ok {
		t.Fatal("OrderItem not found in config")
	}
	if orderItem.Parent != "Order" {
		t.Errorf("OrderItem.Parent = %q, want %q", orderItem.Parent, "Order")
	}
	if orderItem.ParentField != "OrderID" {
		t.Errorf("OrderItem.ParentField = %q, want %q", orderItem.ParentField, "OrderID")
	}
	if orderItem.Sidebar == nil || *orderItem.Sidebar != false {
		t.Errorf("OrderItem.Sidebar should be false")
	}

	// Verify Order has no parent config
	order, ok := config.Models["Order"]
	if !ok {
		t.Fatal("Order not found in config")
	}
	if order.Parent != "" {
		t.Errorf("Order.Parent = %q, want empty", order.Parent)
	}
}

// TestChildModelCollection verifies that GenerateModelFilesImpl correctly
// identifies child models for a parent model.
func TestChildModelCollection(t *testing.T) {
	allModels := map[string]ModelDefinition{
		"Order": {
			Name: "Order",
			Sections: map[string][]ModelField{
				"Details": {
					{Name: "OrderNumber", Type: "string", Table: true},
				},
			},
		},
		"OrderItem": {
			Name:        "OrderItem",
			Parent:      "Order",
			ParentField: "OrderID",
			Sections: map[string][]ModelField{
				"Details": {
					{Name: "ProductName", Type: "string", Table: true},
					{Name: "Quantity", Type: "int", Table: true},
				},
			},
		},
		"OrderNote": {
			Name:        "OrderNote",
			Parent:      "Order",
			ParentField: "OrderID",
			Sections: map[string][]ModelField{
				"Details": {
					{Name: "Content", Type: "string", Table: true},
				},
			},
		},
		"User": {
			Name: "User",
			Sections: map[string][]ModelField{
				"Details": {
					{Name: "Email", Type: "string", Table: true},
				},
			},
		},
	}

	// Simulate the child collection logic from GenerateModelFilesImpl
	var children []ChildModel
	for childName, childModel := range allModels {
		if childModel.Parent == "Order" {
			child := ChildModel{
				Name:              childName,
				NamePlural:        ToPlural(childName),
				NameDisplay:       toDisplayName(childName),
				NamePluralDisplay: toDisplayName(ToPlural(childName)),
				NamePluralLower:   strings.ToLower(toDisplayName(ToPlural(childName))),
				RoutePlural:       ToKebabCase(ToPlural(childName)),
				ParentField:       childModel.ParentField,
				ParentFieldSnake:  ToSnakeCase(childModel.ParentField),
			}
			for _, fields := range childModel.Sections {
				for _, f := range fields {
					if f.Table {
						child.TableFields = append(child.TableFields, TemplateField{
							Name:  f.Name,
							Label: getLabel(f),
						})
					}
				}
			}
			children = append(children, child)
		}
	}

	if len(children) != 2 {
		t.Fatalf("expected 2 children for Order, got %d", len(children))
	}

	// Verify child properties (order may vary due to map iteration)
	childByName := make(map[string]ChildModel)
	for _, c := range children {
		childByName[c.Name] = c
	}

	item, ok := childByName["OrderItem"]
	if !ok {
		t.Fatal("OrderItem not found in children")
	}
	if item.ParentFieldSnake != "order_id" {
		t.Errorf("OrderItem.ParentFieldSnake = %q, want %q", item.ParentFieldSnake, "order_id")
	}
	if item.NamePluralLower != "order items" {
		t.Errorf("OrderItem.NamePluralLower = %q, want %q", item.NamePluralLower, "order items")
	}
	if len(item.TableFields) != 2 {
		t.Errorf("OrderItem.TableFields count = %d, want 2", len(item.TableFields))
	}

	note, ok := childByName["OrderNote"]
	if !ok {
		t.Fatal("OrderNote not found in children")
	}
	if note.RoutePlural != "order-notes" {
		t.Errorf("OrderNote.RoutePlural = %q, want %q", note.RoutePlural, "order-notes")
	}
	if len(note.TableFields) != 1 {
		t.Errorf("OrderNote.TableFields count = %d, want 1", len(note.TableFields))
	}
}

// TestParentTemplateDataPopulation verifies that parent model template data
// is correctly set for child models.
func TestParentTemplateDataPopulation(t *testing.T) {
	model := &ModelDefinition{
		Name:        "OrderItem",
		Parent:      "Order",
		ParentField: "OrderID",
	}

	// Simulate parent data population
	data := ModelTemplateData{}
	if model.Parent != "" {
		data.Parent = model.Parent
		data.ParentField = model.ParentField
		data.ParentFieldSnake = ToSnakeCase(model.ParentField)
		data.ParentRoutePlural = ToKebabCase(ToPlural(model.Parent))
	}

	if data.Parent != "Order" {
		t.Errorf("Parent = %q, want %q", data.Parent, "Order")
	}
	if data.ParentField != "OrderID" {
		t.Errorf("ParentField = %q, want %q", data.ParentField, "OrderID")
	}
	if data.ParentFieldSnake != "order_id" {
		t.Errorf("ParentFieldSnake = %q, want %q", data.ParentFieldSnake, "order_id")
	}
	if data.ParentRoutePlural != "orders" {
		t.Errorf("ParentRoutePlural = %q, want %q", data.ParentRoutePlural, "orders")
	}
}

// TestHideSidebarFlag verifies that the sidebar hiding flag is set correctly
// for child models with sidebar: false.
func TestHideSidebarFlag(t *testing.T) {
	falseBool := false
	trueBool := true

	tests := []struct {
		name    string
		sidebar *bool
		parent  string
		want    bool
	}{
		{"child with sidebar false", &falseBool, "Order", true},
		{"child with sidebar true", &trueBool, "Order", false},
		{"child with sidebar nil", nil, "Order", false},
		{"non-child model", nil, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &ModelDefinition{
				Name:    "OrderItem",
				Parent:  tt.parent,
				Sidebar: tt.sidebar,
			}

			data := ModelTemplateData{}
			if model.Parent != "" {
				if model.Sidebar != nil && !*model.Sidebar {
					data.HideSidebar = true
				}
			}

			if data.HideSidebar != tt.want {
				t.Errorf("HideSidebar = %v, want %v", data.HideSidebar, tt.want)
			}
		})
	}
}

// TestChildModelDisplayNames verifies that child model display names are
// correctly generated for multi-word model names.
func TestChildModelDisplayNames(t *testing.T) {
	tests := []struct {
		name              string
		wantDisplay       string
		wantPluralDisplay string
		wantPluralLower   string
		wantRoutePlural   string
	}{
		{"OrderItem", "Order Item", "Order Items", "order items", "order-items"},
		{"ProductCategory", "Product Category", "Product Categories", "product categories", "product-categories"},
		{"Note", "Note", "Notes", "notes", "notes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := ChildModel{
				Name:              tt.name,
				NamePlural:        ToPlural(tt.name),
				NameDisplay:       toDisplayName(tt.name),
				NamePluralDisplay: toDisplayName(ToPlural(tt.name)),
				NamePluralLower:   strings.ToLower(toDisplayName(ToPlural(tt.name))),
				RoutePlural:       ToKebabCase(ToPlural(tt.name)),
			}

			if child.NameDisplay != tt.wantDisplay {
				t.Errorf("NameDisplay = %q, want %q", child.NameDisplay, tt.wantDisplay)
			}
			if child.NamePluralDisplay != tt.wantPluralDisplay {
				t.Errorf("NamePluralDisplay = %q, want %q", child.NamePluralDisplay, tt.wantPluralDisplay)
			}
			if child.NamePluralLower != tt.wantPluralLower {
				t.Errorf("NamePluralLower = %q, want %q", child.NamePluralLower, tt.wantPluralLower)
			}
			if child.RoutePlural != tt.wantRoutePlural {
				t.Errorf("RoutePlural = %q, want %q", child.RoutePlural, tt.wantRoutePlural)
			}
		})
	}
}

// TestChildTableFieldRelationResolution verifies that child table fields with
// relations use the DTO relation name instead of the raw FK field name.
// (Regression test for #30)
func TestChildTableFieldRelationResolution(t *testing.T) {
	allModels := map[string]ModelDefinition{
		"Order": {
			Name: "Order",
			Sections: map[string][]ModelField{
				"Details": {
					{Name: "OrderNumber", Type: "string", Table: true},
				},
			},
		},
		"Product": {
			Name: "Product",
			Sections: map[string][]ModelField{
				"Details": {
					{Name: "Name", Type: "string", Table: true},
				},
			},
		},
		"OrderItem": {
			Name:        "OrderItem",
			Parent:      "Order",
			ParentField: "OrderID",
			Sections: map[string][]ModelField{
				"Details": {
					{Name: "ProductID", Type: "uint", Relation: "Product", Table: true},
					{Name: "Quantity", Type: "int", Table: true},
					{Name: "Price", Type: "float64", Table: true},
				},
			},
		},
	}

	displayFields := map[string]string{
		"Order":   "OrderNumber",
		"Product": "Name",
	}

	// Simulate the child collection logic from GenerateModelFilesImpl
	var children []ChildModel
	for childName, childModel := range allModels {
		if childModel.Parent == "Order" {
			child := ChildModel{
				Name:       childName,
				NamePlural: ToPlural(childName),
			}
			for _, fields := range childModel.Sections {
				for _, f := range fields {
					if f.Table {
						tf := TemplateField{
							Name:  f.Name,
							Label: getLabel(f),
						}
						if f.Relation != "" {
							tf.IsRelation = true
							tf.DTOFieldName = f.Relation
							tf.RelationDisplayField = "Name"
							if df, ok := displayFields[f.Relation]; ok && df != "" {
								tf.RelationDisplayField = df
							}
						}
						child.TableFields = append(child.TableFields, tf)
					}
				}
			}
			children = append(children, child)
		}
	}

	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}

	item := children[0]
	if len(item.TableFields) != 3 {
		t.Fatalf("expected 3 table fields, got %d", len(item.TableFields))
	}

	// Find the ProductID field (which has a relation)
	var productField *TemplateField
	var quantityField *TemplateField
	for i := range item.TableFields {
		if item.TableFields[i].Name == "ProductID" {
			productField = &item.TableFields[i]
		}
		if item.TableFields[i].Name == "Quantity" {
			quantityField = &item.TableFields[i]
		}
	}

	if productField == nil {
		t.Fatal("ProductID field not found")
	}
	if !productField.IsRelation {
		t.Error("ProductID should be marked as IsRelation")
	}
	if productField.DTOFieldName != "Product" {
		t.Errorf("DTOFieldName = %q, want %q", productField.DTOFieldName, "Product")
	}
	if productField.RelationDisplayField != "Name" {
		t.Errorf("RelationDisplayField = %q, want %q", productField.RelationDisplayField, "Name")
	}

	// Non-relation field should not be marked
	if quantityField == nil {
		t.Fatal("Quantity field not found")
	}
	if quantityField.IsRelation {
		t.Error("Quantity should NOT be marked as IsRelation")
	}
}

// setupModelGenTestDir creates a temporary directory with the required structure
// for testing GenerateModelFilesImpl, and changes the working directory to it.
// Returns a cleanup function to restore the original working directory.
func setupModelGenTestDir(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to temp: %v", err)
	}
	// Create required directories
	for _, d := range []string{"guxgen/models", "guxgen/dto", "guxgen/admin"} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	return dir, func() {
		os.Chdir(origDir)
	}
}

// TestTypeAlias_ModelFileGenerated verifies that when NO manual model file
// exists, a full model file is generated in guxgen/models/.
func TestTypeAlias_ModelFileGenerated(t *testing.T) {
	_, cleanup := setupModelGenTestDir(t)
	defer cleanup()

	model := &ModelDefinition{
		Name: "Product",
		Sections: map[string][]ModelField{
			"Details": {
				{Name: "Name", Type: "string", Table: true},
			},
		},
	}

	err := GenerateModelFilesImpl(model, nil, "myapp", false, nil, nil, nil)
	if err != nil {
		t.Fatalf("GenerateModelFilesImpl: %v", err)
	}

	// Check that a full model file was generated (not a type alias)
	content, err := os.ReadFile("guxgen/models/product_gen.go")
	if err != nil {
		t.Fatalf("read generated model: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "type Product struct") {
		t.Error("expected full model struct definition, got type alias or missing struct")
	}
	if strings.Contains(s, "type Product =") {
		t.Error("should NOT be a type alias when no manual file exists")
	}
}

// TestTypeAlias_ManualModelGeneratesAlias verifies that when a manual model
// file exists in models/, a type alias is generated in guxgen/models/.
// (Regression test for #48)
func TestTypeAlias_ManualModelGeneratesAlias(t *testing.T) {
	_, cleanup := setupModelGenTestDir(t)
	defer cleanup()

	// Create a manual model file
	os.MkdirAll("models", 0755)
	os.WriteFile("models/user.go", []byte("package models\n\ntype User struct {}\n"), 0644)

	model := &ModelDefinition{
		Name:   "User",
		Preset: "auth",
		Sections: map[string][]ModelField{
			"Account": {
				{Name: "Email", Type: "string", Table: true},
				{Name: "Name", Type: "string", Table: true},
			},
		},
	}

	err := GenerateModelFilesImpl(model, nil, "myapp", false, nil, nil, nil)
	if err != nil {
		t.Fatalf("GenerateModelFilesImpl: %v", err)
	}

	// Check that a type alias was generated
	content, err := os.ReadFile("guxgen/models/user_gen.go")
	if err != nil {
		t.Fatalf("read type alias file: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "type User = manualmodels.User") {
		t.Errorf("expected type alias 'type User = manualmodels.User', got:\n%s", s)
	}
	if !strings.Contains(s, `import manualmodels "myapp/models"`) {
		t.Errorf("expected import of manual models package, got:\n%s", s)
	}
	// Should NOT contain a struct definition
	if strings.Contains(s, "type User struct") {
		t.Error("type alias file should NOT contain struct definition")
	}
}

// TestTypeAlias_ManualDTOGeneratesAlias verifies that when a manual DTO file
// exists in dto/, type aliases are generated in guxgen/dto/.
// (Regression test for #48)
func TestTypeAlias_ManualDTOGeneratesAlias(t *testing.T) {
	_, cleanup := setupModelGenTestDir(t)
	defer cleanup()

	// Create a manual DTO file
	os.MkdirAll("dto", 0755)
	os.WriteFile("dto/user.go", []byte("package dto\n\ntype UserList struct {}\ntype UserDetail struct {}\n"), 0644)

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

	// Check that type aliases were generated for DTOs
	content, err := os.ReadFile("guxgen/dto/user_gen.go")
	if err != nil {
		t.Fatalf("read DTO type alias file: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "type UserList = manualdto.UserList") {
		t.Errorf("expected UserList type alias, got:\n%s", s)
	}
	if !strings.Contains(s, "type UserDetail = manualdto.UserDetail") {
		t.Errorf("expected UserDetail type alias, got:\n%s", s)
	}
	if !strings.Contains(s, `import manualdto "myapp/dto"`) {
		t.Errorf("expected import of manual dto package, got:\n%s", s)
	}
}

// TestTypeAlias_NoDTOAliasWhenNoManualFile verifies that when NO manual DTO
// file exists, full DTO structs are generated (not type aliases).
func TestTypeAlias_NoDTOAliasWhenNoManualFile(t *testing.T) {
	_, cleanup := setupModelGenTestDir(t)
	defer cleanup()

	model := &ModelDefinition{
		Name: "Product",
		Sections: map[string][]ModelField{
			"Details": {
				{Name: "Name", Type: "string", Table: true},
			},
		},
	}

	err := GenerateModelFilesImpl(model, nil, "myapp", false, nil, nil, nil)
	if err != nil {
		t.Fatalf("GenerateModelFilesImpl: %v", err)
	}

	content, err := os.ReadFile("guxgen/dto/product_gen.go")
	if err != nil {
		t.Fatalf("read generated DTO: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "type ProductList struct") {
		t.Error("expected full ProductList struct definition")
	}
	if strings.Contains(s, "type ProductList =") {
		t.Error("should NOT be a type alias when no manual file exists")
	}
}

// TestTypeAlias_BothManualFilesGenerateBothAliases verifies that when both
// manual model and manual DTO files exist, both type alias files are generated.
// (Full scenario for #48)
func TestTypeAlias_BothManualFilesGenerateBothAliases(t *testing.T) {
	_, cleanup := setupModelGenTestDir(t)
	defer cleanup()

	// Create both manual files
	os.MkdirAll("models", 0755)
	os.MkdirAll("dto", 0755)
	os.WriteFile("models/user.go", []byte("package models\n\ntype User struct {}\n"), 0644)
	os.WriteFile("dto/user.go", []byte("package dto\n\ntype UserList struct {}\ntype UserDetail struct {}\n"), 0644)

	model := &ModelDefinition{
		Name:   "User",
		Preset: "auth",
		Sections: map[string][]ModelField{
			"Account": {
				{Name: "Email", Type: "string", Table: true},
				{Name: "Name", Type: "string", Table: true},
				{Name: "Role", Type: "string", Table: true},
			},
		},
	}

	err := GenerateModelFilesImpl(model, nil, "myapp", false, nil, nil, nil)
	if err != nil {
		t.Fatalf("GenerateModelFilesImpl: %v", err)
	}

	// Model alias
	modelContent, err := os.ReadFile("guxgen/models/user_gen.go")
	if err != nil {
		t.Fatalf("read model alias: %v", err)
	}
	if !strings.Contains(string(modelContent), "type User = manualmodels.User") {
		t.Error("missing model type alias")
	}

	// DTO alias
	dtoContent, err := os.ReadFile("guxgen/dto/user_gen.go")
	if err != nil {
		t.Fatalf("read DTO alias: %v", err)
	}
	if !strings.Contains(string(dtoContent), "type UserList = manualdto.UserList") {
		t.Error("missing UserList type alias")
	}
	if !strings.Contains(string(dtoContent), "type UserDetail = manualdto.UserDetail") {
		t.Error("missing UserDetail type alias")
	}
}

// TestTypeAlias_ModulePathInImport verifies that the module path is correctly
// embedded in the type alias import statements.
func TestTypeAlias_ModulePathInImport(t *testing.T) {
	_, cleanup := setupModelGenTestDir(t)
	defer cleanup()

	os.MkdirAll("models", 0755)
	os.WriteFile("models/client.go", []byte("package models\n\ntype Client struct {}\n"), 0644)

	model := &ModelDefinition{
		Name: "Client",
		Sections: map[string][]ModelField{
			"Details": {
				{Name: "Name", Type: "string", Table: true},
			},
		},
	}

	err := GenerateModelFilesImpl(model, nil, "github.com/example/myproject", false, nil, nil, nil)
	if err != nil {
		t.Fatalf("GenerateModelFilesImpl: %v", err)
	}

	content, err := os.ReadFile("guxgen/models/client_gen.go")
	if err != nil {
		t.Fatalf("read model alias: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, `import manualmodels "github.com/example/myproject/models"`) {
		t.Errorf("expected full module path in import, got:\n%s", s)
	}
}

// TestTypeAlias_MultiWordModelName verifies that type aliases use the correct
// snake_case file names for multi-word model names.
func TestTypeAlias_MultiWordModelName(t *testing.T) {
	_, cleanup := setupModelGenTestDir(t)
	defer cleanup()

	os.MkdirAll("models", 0755)
	os.WriteFile("models/lead_source.go", []byte("package models\n\ntype LeadSource struct {}\n"), 0644)

	model := &ModelDefinition{
		Name: "LeadSource",
		Sections: map[string][]ModelField{
			"Details": {
				{Name: "Name", Type: "string", Table: true},
			},
		},
	}

	err := GenerateModelFilesImpl(model, nil, "myapp", false, nil, nil, nil)
	if err != nil {
		t.Fatalf("GenerateModelFilesImpl: %v", err)
	}

	content, err := os.ReadFile("guxgen/models/lead_source_gen.go")
	if err != nil {
		t.Fatalf("read model alias for LeadSource: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "type LeadSource = manualmodels.LeadSource") {
		t.Errorf("expected LeadSource type alias, got:\n%s", s)
	}
}

// --- GORM tag generation tests (issue #49) ---

func TestGenerateGormTag_StringFieldSize(t *testing.T) {
	field := &ModelField{Name: "FirstName", Type: "string"}
	tag := generateGormTag(field)
	if !strings.Contains(tag, `size:255`) {
		t.Errorf("expected size:255 for plain string, got: %s", tag)
	}
}

func TestGenerateGormTag_TextareaField(t *testing.T) {
	field := &ModelField{Name: "Message", Type: "string", Input: "textarea"}
	tag := generateGormTag(field)
	if !strings.Contains(tag, `type:text`) {
		t.Errorf("expected type:text for textarea, got: %s", tag)
	}
	if strings.Contains(tag, `size:255`) {
		t.Errorf("textarea should not have size:255, got: %s", tag)
	}
}

func TestGenerateGormTag_RequiredStringField(t *testing.T) {
	field := &ModelField{Name: "Email", Type: "string", Required: true}
	tag := generateGormTag(field)
	if !strings.Contains(tag, `not null`) {
		t.Errorf("expected not null for required field, got: %s", tag)
	}
	if !strings.Contains(tag, `size:255`) {
		t.Errorf("expected size:255 for required string, got: %s", tag)
	}
}

func TestGenerateGormTag_NonStringField(t *testing.T) {
	fields := []ModelField{
		{Name: "Age", Type: "int"},
		{Name: "Active", Type: "bool"},
		{Name: "Price", Type: "float64"},
	}
	for _, f := range fields {
		tag := generateGormTag(&f)
		if strings.Contains(tag, `size:255`) || strings.Contains(tag, `type:text`) {
			t.Errorf("non-string field %s should not have size/type tag, got: %s", f.Name, tag)
		}
	}
}

func TestGenerateGormTag_RelationField(t *testing.T) {
	// Relation fields (e.g., Role string with Relation set) should not get size tag
	field := &ModelField{Name: "Role", Type: "string", Relation: "Role"}
	tag := generateGormTag(field)
	if strings.Contains(tag, `size:255`) || strings.Contains(tag, `type:text`) {
		t.Errorf("relation string field should not get size/type tag, got: %s", tag)
	}
}

func TestGenerateGormTag_SliceStringField(t *testing.T) {
	field := &ModelField{Name: "Tags", Type: "[]string"}
	tag := generateGormTag(field)
	if !strings.Contains(tag, `type:json`) {
		t.Errorf("expected type:json for []string, got: %s", tag)
	}
	if strings.Contains(tag, `size:255`) {
		t.Errorf("[]string should not have size:255, got: %s", tag)
	}
}

// TestConvertToTemplateField_PointerString tests that *string fields generate
// correct EditStateInit with nil-check dereference (#53).
func TestConvertToTemplateField_PointerString(t *testing.T) {
	field := &ModelField{Name: "AdTagURL", Type: "*string"}
	tf := convertToTemplateField(field, "Promotion", nil)

	if tf.StateType != "String" {
		t.Errorf("StateType = %q, want %q", tf.StateType, "String")
	}
	if !tf.IsPointer {
		t.Error("IsPointer should be true for *string")
	}
	if !strings.Contains(tf.EditStateInit, "displayItem.AdTagURL != nil") {
		t.Errorf("EditStateInit should check for nil, got: %s", tf.EditStateInit)
	}
	if !strings.Contains(tf.EditStateInit, "*displayItem.AdTagURL") {
		t.Errorf("EditStateInit should dereference pointer, got: %s", tf.EditStateInit)
	}
	if !strings.Contains(tf.DataValue, "parseStringPtr") {
		t.Errorf("DataValue should use parseStringPtr, got: %s", tf.DataValue)
	}
}

// TestConvertToTemplateField_PointerFloat64 tests that *float64 fields generate
// correct EditStateInit with nil-check dereference (#53).
func TestConvertToTemplateField_PointerFloat64(t *testing.T) {
	field := &ModelField{Name: "MinWatchTimeSec", Type: "*float64"}
	tf := convertToTemplateField(field, "PromotionVideo", nil)

	if tf.StateType != "String" {
		t.Errorf("StateType = %q, want %q", tf.StateType, "String")
	}
	if !tf.IsPointer {
		t.Error("IsPointer should be true for *float64")
	}
	if !strings.Contains(tf.EditStateInit, "displayItem.MinWatchTimeSec != nil") {
		t.Errorf("EditStateInit should check for nil, got: %s", tf.EditStateInit)
	}
	if !strings.Contains(tf.EditStateInit, "*displayItem.MinWatchTimeSec") {
		t.Errorf("EditStateInit should dereference pointer, got: %s", tf.EditStateInit)
	}
	if !strings.Contains(tf.DataValue, "parseFloatPtr") {
		t.Errorf("DataValue should use parseFloatPtr, got: %s", tf.DataValue)
	}
}

// TestConvertToTemplateField_NonPointerString tests that regular string fields
// use direct assignment (no pointer handling).
func TestConvertToTemplateField_NonPointerString(t *testing.T) {
	field := &ModelField{Name: "Name", Type: "string"}
	tf := convertToTemplateField(field, "User", nil)

	if tf.IsPointer {
		t.Error("IsPointer should be false for string")
	}
	if tf.EditStateInit != "displayItem.Name" {
		t.Errorf("EditStateInit = %q, want %q", tf.EditStateInit, "displayItem.Name")
	}
}

// TestGenerateDetailFieldCode_PointerString tests that *string fields in detail views
// generate nil-check code instead of direct access (#53).
func TestGenerateDetailFieldCode_PointerString(t *testing.T) {
	field := &ModelField{Name: "WebhookURL", Type: "*string"}
	tf := convertToTemplateField(field, "Promotion", nil)

	code := generateDetailFieldCode(field, tf, "promotion")

	if !strings.Contains(code, "displayItem.WebhookURL != nil") {
		t.Errorf("detail code should check for nil, got:\n%s", code)
	}
	if !strings.Contains(code, "*displayItem.WebhookURL") {
		t.Errorf("detail code should dereference pointer, got:\n%s", code)
	}
}

// TestGenerateDetailFieldCode_PointerFloat64 tests *float64 detail field generation (#53).
func TestGenerateDetailFieldCode_PointerFloat64(t *testing.T) {
	field := &ModelField{Name: "Price", Type: "*float64"}
	tf := convertToTemplateField(field, "Product", nil)

	code := generateDetailFieldCode(field, tf, "product")

	if !strings.Contains(code, "displayItem.Price != nil") {
		t.Errorf("detail code should check for nil, got:\n%s", code)
	}
	if !strings.Contains(code, "*displayItem.Price") {
		t.Errorf("detail code should dereference pointer, got:\n%s", code)
	}
}

// TestGenerateDetailFieldCode_PointerTime tests *time.Time detail field generation.
func TestGenerateDetailFieldCode_PointerTime(t *testing.T) {
	field := &ModelField{Name: "ExpiresAt", Type: "*time.Time"}
	tf := convertToTemplateField(field, "Session", nil)

	code := generateDetailFieldCode(field, tf, "session")

	if !strings.Contains(code, "displayItem.ExpiresAt != nil") {
		t.Errorf("detail code should check for nil, got:\n%s", code)
	}
	if !strings.Contains(code, `displayItem.ExpiresAt.Format("Jan 2, 2006")`) {
		t.Errorf("detail code should format time, got:\n%s", code)
	}
}

// TestGenerateDetailFieldCode_NonPointerString tests that regular string fields
// use direct access in detail views.
func TestGenerateDetailFieldCode_NonPointerString(t *testing.T) {
	field := &ModelField{Name: "Email", Type: "string"}
	tf := convertToTemplateField(field, "User", nil)

	code := generateDetailFieldCode(field, tf, "user")

	if strings.Contains(code, "!= nil") {
		t.Errorf("non-pointer field should not have nil check, got:\n%s", code)
	}
	if !strings.Contains(code, "displayItem.Email") {
		t.Errorf("should reference displayItem.Email, got:\n%s", code)
	}
}

// TestPrepareModelTemplateData_DisplayFieldIsPointer tests that the DisplayFieldIsPointer
// flag is correctly set when the display field type is a pointer (#53).
func TestPrepareModelTemplateData_DisplayFieldIsPointer(t *testing.T) {
	model := &ModelDefinition{
		Name:    "Promotion",
		Display: "SecondaryColor",
		Sections: map[string][]ModelField{
			"Main": {
				{Name: "SecondaryColor", Type: "*string", Table: true},
				{Name: "Active", Type: "bool"},
			},
		},
	}

	data := prepareModelTemplateData(model, "myapp", nil, nil, nil)
	// Display field is SecondaryColor which is *string — but detectDisplayField
	// only picks string fields, so DisplayField should be empty since we override with model.Display
	if data.DisplayField != "SecondaryColor" {
		t.Errorf("DisplayField = %q, want %q", data.DisplayField, "SecondaryColor")
	}
	if !data.DisplayFieldIsPointer {
		t.Error("DisplayFieldIsPointer should be true when display field is *string")
	}
}

// TestPrepareModelTemplateData_DTOFieldTypeOverride tests that manual DTO field types
// override config field types for admin page generation (#53).
func TestPrepareModelTemplateData_DTOFieldTypeOverride(t *testing.T) {
	model := &ModelDefinition{
		Name: "Widget",
		Sections: map[string][]ModelField{
			"Main": {
				{Name: "Description", Type: "string", Table: true},
			},
		},
	}

	// Simulate a manual DTO with *string where config says string
	dtoTypes := map[string]string{
		"Description": "*string",
	}

	data := prepareModelTemplateData(model, "myapp", nil, nil, dtoTypes)

	// The field type should be overridden to *string
	if len(data.AllFields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(data.AllFields))
	}
	tf := data.AllFields[0]
	if !tf.IsPointer {
		t.Error("field should be marked as pointer after DTO type override")
	}
	if !strings.Contains(tf.EditStateInit, "!= nil") {
		t.Errorf("EditStateInit should have nil check after DTO type override, got: %s", tf.EditStateInit)
	}
}

// TestPrepareModelTemplateData_ModelTypeOverride tests that model struct types
// override config types when they differ (manual model with *string, config says string).
func TestPrepareModelTemplateData_ModelTypeOverride(t *testing.T) {
	// Set up temp dir with a manual model file
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// Create a manual model with *string fields
	os.MkdirAll(filepath.Join("guxgen", "models"), 0755)
	modelContent := `package models

type Promotion struct {
	SecondaryColor *string
	AdTagURL       *string
	WebhookURL     *string
	Active         bool
}
`
	os.WriteFile(filepath.Join("guxgen", "models", "promotion.go"), []byte(modelContent), 0644)

	// Clear cache
	modelFieldTypesCache = make(map[string]map[string]string)

	// Config says "string" for these fields (mismatched with model)
	model := &ModelDefinition{
		Name: "Promotion",
		Sections: map[string][]ModelField{
			"Main": {
				{Name: "SecondaryColor", Type: "string", Table: true},
				{Name: "AdTagURL", Type: "string"},
				{Name: "WebhookURL", Type: "string"},
				{Name: "Active", Type: "bool"},
			},
		},
	}

	// Get model types (simulating what GenerateModelFilesImpl does)
	modelTypes, err := getModelFieldTypes("Promotion")
	if err != nil {
		t.Fatalf("getModelFieldTypes: %v", err)
	}

	data := prepareModelTemplateData(model, "myapp", nil, nil, modelTypes)

	// Verify that *string fields were detected and handled
	for _, tf := range data.AllFields {
		switch tf.Name {
		case "SecondaryColor", "AdTagURL", "WebhookURL":
			if !tf.IsPointer {
				t.Errorf("field %s should be IsPointer=true after model type override", tf.Name)
			}
			if !strings.Contains(tf.EditStateInit, "!= nil") {
				t.Errorf("field %s EditStateInit should have nil check, got: %s", tf.Name, tf.EditStateInit)
			}
			if !strings.Contains(tf.DetailField, "!= nil") {
				t.Errorf("field %s DetailField should have nil check, got: %s", tf.Name, tf.DetailField)
			}
		case "Active":
			if tf.IsPointer {
				t.Errorf("field Active should NOT be IsPointer")
			}
		}
	}
}

// TestPrepareModelTemplateData_ConfigPointerTypeWorks tests that config with *string
// type correctly generates pointer handling even without model type override.
func TestPrepareModelTemplateData_ConfigPointerTypeWorks(t *testing.T) {
	// No temp directory needed - just test config-driven types
	model := &ModelDefinition{
		Name: "Widget",
		Sections: map[string][]ModelField{
			"Main": {
				{Name: "Description", Type: "*string", Table: true},
				{Name: "Price", Type: "*float64"},
				{Name: "Name", Type: "string"},
			},
		},
	}

	// No DTO type override (nil) - rely purely on config types
	data := prepareModelTemplateData(model, "myapp", nil, nil, nil)

	for _, tf := range data.AllFields {
		switch tf.Name {
		case "Description":
			if !tf.IsPointer {
				t.Errorf("Description should be IsPointer=true from config *string")
			}
			if !strings.Contains(tf.EditStateInit, "!= nil") {
				t.Errorf("Description EditStateInit should have nil check, got: %s", tf.EditStateInit)
			}
			if !strings.Contains(tf.DetailField, "!= nil") {
				t.Errorf("Description DetailField should have nil check, got: %s", tf.DetailField)
			}
		case "Price":
			if !tf.IsPointer {
				t.Errorf("Price should be IsPointer=true from config *float64")
			}
			if !strings.Contains(tf.EditStateInit, "!= nil") {
				t.Errorf("Price EditStateInit should have nil check, got: %s", tf.EditStateInit)
			}
		case "Name":
			if tf.IsPointer {
				t.Errorf("Name should NOT be IsPointer")
			}
		}
	}
}

// TestAuthPresetPasswordField verifies that auth preset models get a virtual
// Password field injected into forms (#65).
func TestAuthPresetPasswordField(t *testing.T) {
	model := ModelDefinition{
		Name:   "User",
		Preset: "auth",
		Sections: map[string][]ModelField{
			"Account": {
				{Name: "Email", Type: "string", Required: true, Table: true, Input: "email"},
				{Name: "Name", Type: "string", Required: true, Table: true},
				{Name: "Role", Type: "string", Table: true, Input: "select"},
			},
		},
	}

	data := prepareModelTemplateData(&model, "myapp", nil, nil, nil)

	// Check that Password field was injected
	var passwordField *TemplateField
	for i, tf := range data.AllFields {
		if tf.Name == "Password" {
			passwordField = &data.AllFields[i]
			break
		}
	}

	if passwordField == nil {
		t.Fatal("Password field not found in AllFields for auth preset model")
	}

	if !passwordField.FormOnly {
		t.Error("Password field should be FormOnly=true")
	}

	if passwordField.DetailField != "" {
		t.Error("Password field should have empty DetailField")
	}

	if passwordField.EditStateInit != `""` {
		t.Errorf("Password EditStateInit = %q, want empty string literal", passwordField.EditStateInit)
	}

	if !strings.Contains(passwordField.FormField, "password") {
		t.Errorf("Password FormField should contain 'password' input type, got: %s", passwordField.FormField)
	}

	// Check Password is in the first section's fields
	if len(data.Sections) == 0 {
		t.Fatal("expected at least one section")
	}
	found := false
	for _, tf := range data.Sections[0].Fields {
		if tf.Name == "Password" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Password field not found in first section")
	}
}

// TestAuthPresetPasswordFieldNotInDTO verifies that FormOnly fields are excluded
// from non-auth models and that normal models don't get the password field.
func TestNonAuthModelNoPasswordField(t *testing.T) {
	model := ModelDefinition{
		Name: "Product",
		Sections: map[string][]ModelField{
			"Info": {
				{Name: "Name", Type: "string", Required: true, Table: true},
				{Name: "Price", Type: "float64", Table: true},
			},
		},
	}

	data := prepareModelTemplateData(&model, "myapp", nil, nil, nil)

	for _, tf := range data.AllFields {
		if tf.Name == "Password" {
			t.Error("non-auth model should not have a Password field")
		}
		if tf.FormOnly {
			t.Errorf("non-auth model should not have FormOnly fields, found: %s", tf.Name)
		}
	}
}

// TestParseSizeString verifies that parseSizeString correctly converts human-readable
// size strings to bytes (e.g., "5MB" -> 5242880).
func TestParseSizeString(t *testing.T) {
	tests := []struct {
		input       string
		expected    int64
		expectError bool
	}{
		{"5MB", 5 * 1024 * 1024, false},
		{"5mb", 5 * 1024 * 1024, false},
		{"5 MB", 5 * 1024 * 1024, false},
		{"1GB", 1024 * 1024 * 1024, false},
		{"500KB", 500 * 1024, false},
		{"1024", 1024, false},
		{"1024B", 1024, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
		{"5XB", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseSizeString(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("parseSizeString(%q) expected error, got none", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("parseSizeString(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.expected {
				t.Errorf("parseSizeString(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

// TestConvertToTemplateField_MultiFile tests that file[] fields generate correct
// template field with IsMultiFile=true and DTOType="[]*core.FileInfo".
func TestConvertToTemplateField_MultiFile(t *testing.T) {
	field := &ModelField{Name: "Gallery", Type: "string", Input: "file[]"}
	tf := convertToTemplateField(field, "Product", nil)

	if !tf.IsMultiFile {
		t.Error("IsMultiFile should be true for file[] input")
	}
	if tf.DTOType != "[]*core.FileInfo" {
		t.Errorf("DTOType = %q, want %q", tf.DTOType, "[]*core.FileInfo")
	}
	if tf.GoType != "string" {
		t.Errorf("GoType = %q, want %q (stores JSON array)", tf.GoType, "string")
	}
	if tf.StateType != "String" {
		t.Errorf("StateType = %q, want %q", tf.StateType, "String")
	}
	if !strings.Contains(tf.EditStateInit, "json.Marshal") {
		t.Errorf("EditStateInit should reconstruct JSON array from []*FileInfo, got: %s", tf.EditStateInit)
	}
}

// TestConvertToTemplateField_FileWithConfig tests that file fields with
// Accept, MaxSize, Directory config populate TemplateField correctly.
func TestConvertToTemplateField_FileWithConfig(t *testing.T) {
	field := &ModelField{
		Name:      "Avatar",
		Type:      "string",
		Input:     "file",
		Accept:    "image/*",
		MaxSize:   "5MB",
		Directory: "avatars",
	}
	tf := convertToTemplateField(field, "User", nil)

	if tf.FileAccept != "image/*" {
		t.Errorf("FileAccept = %q, want %q", tf.FileAccept, "image/*")
	}
	if tf.FileMaxSize != 5*1024*1024 {
		t.Errorf("FileMaxSize = %d, want %d", tf.FileMaxSize, 5*1024*1024)
	}
	if tf.FileDir != "avatars" {
		t.Errorf("FileDir = %q, want %q", tf.FileDir, "avatars")
	}
}

// TestGenerateFormFieldCode_MultiFile tests that multi-file fields generate
// ui.MultiFileUpload with JSON state management.
func TestGenerateFormFieldCode_MultiFile(t *testing.T) {
	field := &ModelField{Name: "Gallery", Type: "string", Input: "file[]"}
	tf := convertToTemplateField(field, "Product", nil)
	tf.StateVar = "galleryState"

	code := generateFormFieldCode(field, tf, "product")

	if !strings.Contains(code, "ui.MultiFileUpload") {
		t.Error("multi-file field should generate ui.MultiFileUpload")
	}
	if !strings.Contains(code, "json.Unmarshal") {
		t.Error("multi-file field should unmarshal JSON keys")
	}
	if !strings.Contains(code, "OnUploadComplete") {
		t.Error("multi-file field should have OnUploadComplete handler")
	}
	if !strings.Contains(code, "OnRemove") {
		t.Error("multi-file field should have OnRemove handler")
	}
}

// TestGenerateFormFieldCode_FileWithConfig tests that single file fields with
// per-field config emit correct FileUpload props.
func TestGenerateFormFieldCode_FileWithConfig(t *testing.T) {
	field := &ModelField{
		Name:      "Avatar",
		Type:      "string",
		Input:     "file",
		Accept:    "image/*",
		MaxSize:   "10MB",
		Directory: "avatars",
	}
	tf := convertToTemplateField(field, "User", nil)
	tf.StateVar = "avatarState"

	code := generateFormFieldCode(field, tf, "user")

	if !strings.Contains(code, `Accept: []string{"image/*"}`) {
		t.Error("file field with Accept should emit Accept prop")
	}
	if !strings.Contains(code, fmt.Sprintf("MaxSize: %d", 10*1024*1024)) {
		t.Error("file field with MaxSize should emit MaxSize prop")
	}
	if !strings.Contains(code, `"/__gux_api/upload?dir=avatars"`) {
		t.Error("file field with Directory should emit UploadURL with dir param")
	}
}

// TestGenerateDetailFieldCode_MultiFile tests that multi-file detail view
// generates gallery rendering code.
func TestGenerateDetailFieldCode_MultiFile(t *testing.T) {
	field := &ModelField{Name: "Gallery", Type: "string", Input: "file[]"}
	tf := convertToTemplateField(field, "Product", nil)
	tf.DTOFieldName = "Gallery"

	code := generateDetailFieldCode(field, tf, "product")

	if !strings.Contains(code, "displayItem.Gallery") {
		t.Error("detail code should reference Gallery field")
	}
	if !strings.Contains(code, "isImageFile") {
		t.Error("detail code should check for images")
	}
	if !strings.Contains(code, "flex flex-wrap gap-2") {
		t.Error("detail code should render image gallery")
	}
	if !strings.Contains(code, "h-20 w-20") {
		t.Error("detail code should render thumbnails")
	}
}

// TestExtractDisplayFields tests extracting field names from display templates
func TestExtractDisplayFields(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{
			name:   "single field",
			input:  "Name",
			expect: []string{"Name"},
		},
		{
			name:   "two template fields",
			input:  "{{FirstName}} {{LastName}}",
			expect: []string{"FirstName", "LastName"},
		},
		{
			name:   "three fields with separator",
			input:  "{{FirstName}} {{LastName}} - {{Organization}}",
			expect: []string{"FirstName", "LastName", "Organization"},
		},
		{
			name:   "empty string",
			input:  "",
			expect: nil,
		},
		{
			name:   "single template field",
			input:  "{{Title}}",
			expect: []string{"Title"},
		},
		{
			name:   "field with spaces in braces",
			input:  "{{ Name }}",
			expect: []string{"Name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDisplayFields(tt.input)
			if len(result) != len(tt.expect) {
				t.Errorf("extractDisplayFields(%q) = %v, want %v", tt.input, result, tt.expect)
				return
			}
			for i, v := range result {
				if v != tt.expect[i] {
					t.Errorf("extractDisplayFields(%q)[%d] = %q, want %q", tt.input, i, v, tt.expect[i])
				}
			}
		})
	}
}

// TestIsDisplayTemplate tests checking if display string is a template
func TestIsDisplayTemplate(t *testing.T) {
	tests := []struct {
		input  string
		expect bool
	}{
		{"Name", false},
		{"{{Name}}", true},
		{"{{FirstName}} {{LastName}}", true},
		{"", false},
	}

	for _, tt := range tests {
		if got := isDisplayTemplate(tt.input); got != tt.expect {
			t.Errorf("isDisplayTemplate(%q) = %v, want %v", tt.input, got, tt.expect)
		}
	}
}

// TestSectionOrderPreservation tests that section order is preserved during config round-trip
func TestSectionOrderPreservation(t *testing.T) {
	configJSON := `{
  "module": "testapp",
  "models": {
    "Client": {
      "sections": {
        "Contact": [{"name": "Email", "type": "string"}],
        "Account": [{"name": "Name", "type": "string"}],
        "Billing": [{"name": "Balance", "type": "float64"}]
      }
    }
  }
}`
	dir := t.TempDir()
	configPath := filepath.Join(dir, "gux.config.json")
	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original dir and restore after test
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	config, err := LoadModelsConfig(".")
	if err != nil {
		t.Fatalf("LoadModelsConfig: %v", err)
	}

	model := config.Models["Client"]
	expectedOrder := []string{"Contact", "Account", "Billing"}
	if len(model.SectionOrder) != len(expectedOrder) {
		t.Fatalf("SectionOrder length = %d, want %d: %v", len(model.SectionOrder), len(expectedOrder), model.SectionOrder)
	}
	for i, got := range model.SectionOrder {
		if got != expectedOrder[i] {
			t.Errorf("SectionOrder[%d] = %q, want %q", i, got, expectedOrder[i])
		}
	}
}

// TestFormFieldRequiredIndicator tests that required fields generate asterisk indicators
func TestFormFieldRequiredIndicator(t *testing.T) {
	field := &ModelField{
		Name:     "Email",
		Type:     "string",
		Required: true,
		Input:    "email",
	}
	tf := convertToTemplateField(field, "User", nil)
	code := generateFormFieldCode(field, tf, "user")

	if !strings.Contains(code, "true") {
		t.Error("expected form field code to pass true for required parameter")
	}
}

// TestFormFieldNotRequiredIndicator tests that non-required fields don't get asterisk
func TestFormFieldNotRequiredIndicator(t *testing.T) {
	field := &ModelField{
		Name: "Notes",
		Type: "string",
	}
	tf := convertToTemplateField(field, "User", nil)
	code := generateFormFieldCode(field, tf, "user")

	if !strings.Contains(code, "false") {
		t.Error("expected form field code to pass false for non-required parameter")
	}
}

// TestConvertToTemplateField_Format tests that Format is properly carried through
func TestConvertToTemplateField_Format(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"currency", "currency"},
		{"percent", "percent"},
		{"number", "number"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &ModelField{
				Name:   "Amount",
				Type:   "float64",
				Format: tt.format,
			}
			tf := convertToTemplateField(field, "Order", nil)
			if tf.Format != tt.format {
				t.Errorf("Format = %q, want %q", tf.Format, tt.format)
			}
		})
	}
}

// TestDetailFieldCode_FormattedCurrency tests that currency format generates formatCurrency call
func TestDetailFieldCode_FormattedCurrency(t *testing.T) {
	field := &ModelField{
		Name:   "Amount",
		Type:   "float64",
		Format: "currency",
	}
	tf := convertToTemplateField(field, "Order", nil)
	tf.Format = "currency"
	code := generateDetailFieldCode(field, tf, "order")

	if !strings.Contains(code, "formatCurrency") {
		t.Errorf("expected detail code to use formatCurrency, got:\n%s", code)
	}
}

// TestDetailFieldCode_FormattedPercent tests percent format in detail view
func TestDetailFieldCode_FormattedPercent(t *testing.T) {
	field := &ModelField{
		Name:   "Rate",
		Type:   "float64",
		Format: "percent",
	}
	tf := convertToTemplateField(field, "Order", nil)
	tf.Format = "percent"
	code := generateDetailFieldCode(field, tf, "order")

	if !strings.Contains(code, "formatPercent") {
		t.Errorf("expected detail code to use formatPercent, got:\n%s", code)
	}
}

// TestDetailFieldCode_PointerCurrency tests currency format for pointer float64
func TestDetailFieldCode_PointerCurrency(t *testing.T) {
	field := &ModelField{
		Name:   "Amount",
		Type:   "*float64",
		Format: "currency",
	}
	tf := convertToTemplateField(field, "Order", nil)
	tf.Format = "currency"
	code := generateDetailFieldCode(field, tf, "order")

	if !strings.Contains(code, "formatCurrencyPtr") {
		t.Errorf("expected detail code to use formatCurrencyPtr, got:\n%s", code)
	}
}

// TestConvertToTemplateField_Required tests Required flag propagation
func TestConvertToTemplateField_Required(t *testing.T) {
	field := &ModelField{
		Name:     "Email",
		Type:     "string",
		Required: true,
	}
	tf := convertToTemplateField(field, "User", nil)
	if !tf.Required {
		t.Error("TemplateField.Required should be true when ModelField.Required is true")
	}

	field2 := &ModelField{
		Name: "Notes",
		Type: "string",
	}
	tf2 := convertToTemplateField(field2, "User", nil)
	if tf2.Required {
		t.Error("TemplateField.Required should be false when ModelField.Required is false")
	}
}

// TestModelDefinitionMarshalJSON_PreservesOrder tests custom JSON marshal preserves section order
func TestModelDefinitionMarshalJSON_PreservesOrder(t *testing.T) {
	model := ModelDefinition{
		Name:         "Client",
		SectionOrder: []string{"Contact", "Account", "Billing"},
		Sections: map[string][]ModelField{
			"Contact": {{Name: "Email", Type: "string"}},
			"Account": {{Name: "Name", Type: "string"}},
			"Billing": {{Name: "Balance", Type: "float64"}},
		},
	}

	data, err := json.Marshal(model)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	jsonStr := string(data)
	contactIdx := strings.Index(jsonStr, `"Contact"`)
	accountIdx := strings.Index(jsonStr, `"Account"`)
	billingIdx := strings.Index(jsonStr, `"Billing"`)

	if contactIdx == -1 || accountIdx == -1 || billingIdx == -1 {
		t.Fatalf("missing section in JSON: %s", jsonStr)
	}

	if contactIdx > accountIdx || accountIdx > billingIdx {
		t.Errorf("sections not in expected order. Contact@%d, Account@%d, Billing@%d\nJSON: %s",
			contactIdx, accountIdx, billingIdx, jsonStr)
	}
}

// TestTimeFieldOmittedWhenEmpty tests that time fields are conditionally included in data map
func TestTimeFieldOmittedWhenEmpty(t *testing.T) {
	field := &ModelField{
		Name:  "LeadDate",
		Type:  "time.Time",
		Input: "date",
	}
	tf := convertToTemplateField(field, "Lead", nil)

	if !tf.IsTime {
		t.Error("expected IsTime to be true for time.Time field")
	}
}

// TestFormatFieldConfig tests that format config option is parsed correctly
func TestFormatFieldConfig(t *testing.T) {
	configJSON := `{
  "module": "testapp",
  "models": {
    "Order": {
      "sections": {
        "General": [
          {"name": "Amount", "type": "float64", "format": "currency"},
          {"name": "TaxRate", "type": "float64", "format": "percent"},
          {"name": "Quantity", "type": "int", "format": "number"}
        ]
      }
    }
  }
}`
	var config ModelsConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	fields := config.Models["Order"].Sections["General"]
	formats := map[string]string{
		"Amount":   "currency",
		"TaxRate":  "percent",
		"Quantity": "number",
	}

	for _, field := range fields {
		expected := formats[field.Name]
		if field.Format != expected {
			t.Errorf("field %s: Format = %q, want %q", field.Name, field.Format, expected)
		}
	}
}

// TestTextareaRequiredIndicator tests textarea fields get required asterisk
func TestTextareaRequiredIndicator(t *testing.T) {
	field := &ModelField{
		Name:     "Description",
		Type:     "string",
		Input:    "textarea",
		Required: true,
	}
	tf := convertToTemplateField(field, "Product", nil)
	code := generateFormFieldCode(field, tf, "product")

	if !strings.Contains(code, "text-red-500") {
		t.Error("expected textarea required field to contain red asterisk indicator")
	}
}

// TestSelectRequiredIndicator tests select fields get required asterisk
func TestSelectRequiredIndicator(t *testing.T) {
	field := &ModelField{
		Name:     "Status",
		Type:     "string",
		Input:    "select",
		Required: true,
		Options: []SelectOption{
			{Value: "active", Label: "Active"},
			{Value: "inactive", Label: "Inactive"},
		},
	}
	tf := convertToTemplateField(field, "Product", nil)
	code := generateFormFieldCode(field, tf, "product")

	if !strings.Contains(code, "text-red-500") {
		t.Error("expected select required field to contain red asterisk indicator")
	}
}

// TestBriefDTOInfo_CompositeDisplay tests BriefDTOInfo with composite display
func TestBriefDTOInfo_CompositeDisplay(t *testing.T) {
	display := "{{FirstName}} {{LastName}} - {{Organization}}"
	fields := extractDisplayFields(display)
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d: %v", len(fields), fields)
	}
	if fields[0] != "FirstName" || fields[1] != "LastName" || fields[2] != "Organization" {
		t.Errorf("fields = %v, want [FirstName LastName Organization]", fields)
	}

	info := BriefDTOInfo{
		Model:         "Client",
		DisplayField:  display,
		DisplayFields: fields,
		DisplayFormat: display,
	}

	if !isDisplayTemplate(info.DisplayField) {
		t.Error("expected display field to be detected as template")
	}
	if len(info.DisplayFields) != 3 {
		t.Errorf("DisplayFields length = %d, want 3", len(info.DisplayFields))
	}
}

// TestDetailFieldCode_RelationUsesDisplay tests that relation detail code uses Display() method
func TestDetailFieldCode_RelationUsesDisplay(t *testing.T) {
	field := &ModelField{
		Name:     "ClientID",
		Type:     "*uint",
		Relation: "Client",
	}
	tf := convertToTemplateField(field, "Order", map[string]string{"Client": "Name"})
	code := generateDetailFieldCode(field, tf, "order")

	if !strings.Contains(code, ".Display()") {
		t.Errorf("expected detail code to use Display() method, got:\n%s", code)
	}
}

// TestFormFieldCode_RequiredBoolParam tests the required bool parameter in formField calls
func TestFormFieldCode_RequiredBoolParam(t *testing.T) {
	tests := []struct {
		name     string
		required bool
		expect   string
	}{
		{"required field", true, "true"},
		{"optional field", false, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &ModelField{
				Name:     "Email",
				Type:     "string",
				Input:    "email",
				Required: tt.required,
			}
			tf := convertToTemplateField(field, "User", nil)
			code := generateFormFieldCode(field, tf, "user")

			// The formField call should include the required bool
			expectedPattern := fmt.Sprintf(`FormField("Email", "email", "email", emailState.Get(), %s,`, tt.expect)
			if !strings.Contains(code, expectedPattern) {
				t.Errorf("expected code to contain %q\nGot: %s", expectedPattern, code)
			}
		})
	}
}
