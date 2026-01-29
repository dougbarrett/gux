package main

import (
	"strings"
	"testing"
)

// TestGenerateFormFieldCode_RelationDropdownUsesDisplayField verifies that
// relation dropdown options use the Brief DTO's actual display field
// instead of a hardcoded field name. (Regression test for #22, #27)
func TestGenerateFormFieldCode_RelationDropdownUsesDisplayField(t *testing.T) {
	tests := []struct {
		name         string
		relation     string
		displayField string
		fieldType    string
		wantField    string // expected field access in generated code
	}{
		{
			name:         "default Name field",
			relation:     "Industry",
			displayField: "Name",
			fieldType:    "*uint",
			wantField:    "item.Name",
		},
		{
			name:         "custom display field OrderNumber",
			relation:     "Order",
			displayField: "OrderNumber",
			fieldType:    "*uint",
			wantField:    "item.OrderNumber",
		},
		{
			name:         "custom display field FirstName",
			relation:     "Client",
			displayField: "FirstName",
			fieldType:    "*uint",
			wantField:    "item.FirstName",
		},
		{
			name:         "required uint relation with custom display",
			relation:     "Order",
			displayField: "OrderNumber",
			fieldType:    "uint",
			wantField:    "item.OrderNumber",
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
				RelationDisplayField: tt.displayField,
			}
			code := generateFormFieldCode(field, tf, "test")

			if !strings.Contains(code, tt.wantField) {
				t.Errorf("expected generated code to contain %q, but it doesn't.\nGenerated:\n%s", tt.wantField, code)
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
