package main

import (
	"crypto/md5"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

// pluralize converts a singular word to its plural form.
// Handles common English pluralization rules.
func pluralize(word string) string {
	if word == "" {
		return word
	}

	lower := strings.ToLower(word)

	// Words ending in s, x, z, ch, sh → add "es"
	if strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") ||
		strings.HasSuffix(lower, "z") || strings.HasSuffix(lower, "ch") ||
		strings.HasSuffix(lower, "sh") {
		return word + "es"
	}

	// Words ending in consonant + y → change y to ies
	if strings.HasSuffix(lower, "y") && len(word) > 1 {
		prev := lower[len(lower)-2]
		// Check if previous char is a consonant (not a, e, i, o, u)
		if prev != 'a' && prev != 'e' && prev != 'i' && prev != 'o' && prev != 'u' {
			return word[:len(word)-1] + "ies"
		}
	}

	// Default: add "s"
	return word + "s"
}

// getModulePath uses go list to get the module path for the current directory
func getModulePath() (string, error) {
	cmd := exec.Command("go", "list", "-m")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get module path: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// getCurrentPackagePath gets the import path for the current directory
func getCurrentPackagePath() (string, error) {
	cmd := exec.Command("go", "list", ".")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get package path: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// findMainAppFile finds the main app file (app.go or main.go with core.New())
func findMainAppFile() (string, error) {
	// Check app.go first
	if _, err := os.Stat("app.go"); err == nil {
		return "app.go", nil
	}
	// Check main.go
	if _, err := os.Stat("main.go"); err == nil {
		return "main.go", nil
	}
	return "", fmt.Errorf("no app.go or main.go found in current directory")
}

// PageRoute represents a route with its page handler
type PageRoute struct {
	Path     string
	Handler  string // e.g., "pages.Home"
	IsHybrid bool
	Bundle   string // Bundle name (empty = default "app")
}

// BundleInfo represents a WASM bundle with its routes and imports
type BundleInfo struct {
	Name    string      // Bundle name (e.g., "admin")
	Routes  []PageRoute // Routes in this bundle
	Imports []string    // Package imports needed (e.g., "github.com/.../admin")
}

// CRUDModel represents a registered CRUD model
type CRUDModel struct {
	Name       string // e.g., "Counter"
	PluralName string // e.g., "Counters"
	Path       string // e.g., "counters"
	ListDTO    string // e.g., "UserList" - DTO type for list responses
	DetailDTO  string // e.g., "UserDetail" - DTO type for detail responses
	DTOPackage string // e.g., "dto" - package name for DTOs
	ListDTOInfo   *DTOInfo // Parsed DTO info for list responses
	DetailDTOInfo *DTOInfo // Parsed DTO info for detail responses
}

// DTOInfo holds parsed information about a DTO struct
type DTOInfo struct {
	Name      string            // e.g., "UserList"
	ModelName string            // e.g., "User" (extracted from first field's gux tag)
	Fields    []DTOFieldMapping // Field mappings
	Preloads  []string          // Preload directives from tags
}

// DTOFieldMapping represents a single field mapping from DTO to model
type DTOFieldMapping struct {
	DTOField    string // Field name in DTO, e.g., "Email"
	DTOType     string // Field type in DTO, e.g., "string"
	ModelField  string // Field name in model, e.g., "Email" (from gux:"User.Email")
	IsSlice     bool   // Whether this is a slice field (for relationships)
	SliceDTO    string // For slice fields, the element DTO type, e.g., "PostBrief"
	Preload     string // Preload directive if any
	IsNestedDTO bool   // Whether this is a nested DTO type (e.g., UserBrief)
}

// APIEndpointInfo represents a parsed API endpoint for code generation
type APIEndpointInfo struct {
	Method       string   // HTTP method: GET, POST, PUT, PATCH, DELETE
	Path         string   // URL path: /api/login, /api/users/:id
	FuncName     string   // Generated function name: Login, GetUser, DeleteUser
	RequestType  string   // Request body type name (empty for GET/DELETE)
	ResponseType string   // Response type name (empty for DELETE)
	Package      string   // Package where types are defined (e.g., "dto")
	PathParams   []string // Path parameter names (e.g., ["id"])
}

// isPrimitiveType checks if a type string represents a primitive Go type
func isPrimitiveType(t string) bool {
	// Strip pointer prefix for checking
	t = strings.TrimPrefix(t, "*")
	primitives := map[string]bool{
		"string": true, "int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true, "bool": true, "byte": true, "rune": true,
		"time.Time": true, // Common type that should be treated as primitive
	}
	return primitives[t]
}

// modelFieldTypesCache caches parsed model field types
var modelFieldTypesCache = make(map[string]map[string]string)

// ModelFieldInfo holds information about a model field for code generation
type ModelFieldInfo struct {
	Name     string
	Type     string
	JSONName string
}

// getModelFields returns the fields of a model (excluding gorm.Model embedded fields)
func getModelFields(modelName string) ([]ModelFieldInfo, error) {
	fieldTypes, err := getModelFieldTypes(modelName)
	if err != nil {
		return nil, err
	}

	// Fields to exclude (from gorm.Model)
	excludedFields := map[string]bool{
		"ID":        true,
		"CreatedAt": true,
		"UpdatedAt": true,
		"DeletedAt": true,
		"Model":     true, // embedded gorm.Model
	}

	var fields []ModelFieldInfo
	for name, typ := range fieldTypes {
		if excludedFields[name] {
			continue
		}
		// Skip embedded structs and relation fields (non-primitive pointer types)
		if strings.Contains(typ, ".") || (strings.HasPrefix(typ, "*") && !isPrimitiveType(strings.TrimPrefix(typ, "*"))) {
			// Check if it's a simple FK pointer like *uint - those we keep
			trimmed := strings.TrimPrefix(typ, "*")
			if trimmed != "uint" && trimmed != "int" && trimmed != "string" {
				continue
			}
		}
		// Skip slice types (M2M relations)
		if strings.HasPrefix(typ, "[]") {
			continue
		}

		fields = append(fields, ModelFieldInfo{
			Name:     name,
			Type:     typ,
			JSONName: toSnakeCase(name),
		})
	}

	return fields, nil
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// generateDTOStructCode generates a DTO struct definition from model fields
func generateDTOStructCode(modelName string, fields []ModelFieldInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`
// %s is the API DTO for %s.
type %s struct {
	ID          uint   `+"`json:\"id\"`"+`
	CreatedAt   string `+"`json:\"created_at,omitempty\"`"+`
	UpdatedAt   string `+"`json:\"updated_at,omitempty\"`"+`
`, modelName, modelName, modelName))

	for _, f := range fields {
		// Map Go types to JSON-safe types for the DTO
		dtoType := f.Type
		switch f.Type {
		case "int", "int64", "int32":
			dtoType = "int"
		case "uint", "uint64", "uint32":
			dtoType = "uint"
		case "*uint", "*int":
			dtoType = "*uint"
		case "float64", "float32":
			dtoType = "float64"
		case "bool":
			dtoType = "bool"
		case "time.Time", "*time.Time":
			dtoType = "string" // Times are serialized as strings
		default:
			dtoType = "string"
		}
		sb.WriteString(fmt.Sprintf("\t%s %s `json:\"%s,omitempty\"`\n", f.Name, dtoType, f.JSONName))
	}

	sb.WriteString("}\n")
	return sb.String()
}

// generateFieldAssignments generates field assignment code for DTO conversion
func generateFieldAssignments(fields []ModelFieldInfo, sourceVar string, indent string) string {
	var sb strings.Builder
	for _, f := range fields {
		switch f.Type {
		case "time.Time":
			sb.WriteString(fmt.Sprintf("%s%s: %s.%s.Format(time.RFC3339),\n", indent, f.Name, sourceVar, f.Name))
		case "*time.Time":
			sb.WriteString(fmt.Sprintf("%sif %s.%s != nil { result.%s = %s.%s.Format(time.RFC3339) }\n", indent, sourceVar, f.Name, f.Name, sourceVar, f.Name))
		default:
			sb.WriteString(fmt.Sprintf("%s%s: %s.%s,\n", indent, f.Name, sourceVar, f.Name))
		}
	}
	return sb.String()
}

// generateModelAssignments generates field assignment code for model update
func generateModelAssignments(fields []ModelFieldInfo, sourceVar, targetVar string, indent string) string {
	var sb strings.Builder
	for _, f := range fields {
		sb.WriteString(fmt.Sprintf("%s%s.%s = %s.%s\n", indent, targetVar, f.Name, sourceVar, f.Name))
	}
	return sb.String()
}

// generateServerAPICode generates server-side API code for models without DTOs
// using the actual model fields instead of assuming Name/Description
func generateServerAPICode(modelName, pluralName string) string {
	// Get actual fields from the model
	fields, err := getModelFields(modelName)
	if err != nil || len(fields) == 0 {
		// Fallback: generate minimal API with just ID
		return generateMinimalServerAPICode(modelName, pluralName)
	}

	var sb strings.Builder

	// Generate field assignments for model-to-DTO conversion
	var listFieldAssignments strings.Builder
	var getFieldAssignments strings.Builder
	var createModelAssignments strings.Builder
	var createResultAssignments strings.Builder
	var updateModelAssignments strings.Builder
	var updateResultAssignments strings.Builder

	for _, f := range fields {
		// For List/Get: DTO field = model field
		listFieldAssignments.WriteString(fmt.Sprintf("\t\t\t%s: item.%s,\n", f.Name, f.Name))
		getFieldAssignments.WriteString(fmt.Sprintf("\t\t%s: item.%s,\n", f.Name, f.Name))
		createResultAssignments.WriteString(fmt.Sprintf("\t\t%s: model.%s,\n", f.Name, f.Name))
		updateResultAssignments.WriteString(fmt.Sprintf("\t\t%s: model.%s,\n", f.Name, f.Name))

		// For Create: model field = DTO field
		createModelAssignments.WriteString(fmt.Sprintf("\t\t%s: item.%s,\n", f.Name, f.Name))

		// For Update: model.field = DTO.field
		updateModelAssignments.WriteString(fmt.Sprintf("\tmodel.%s = item.%s\n", f.Name, f.Name))
	}

	sb.WriteString(fmt.Sprintf(`
// %sAPI provides CRUD operations for %s.
type %sAPI struct{}

// %s is the API client for %s operations.
var %s = &%sAPI{}

// List returns all %s records.
func (a *%sAPI) List(callback func([]%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	var items []models.%s
	if err := db.Find(&items).Error; err != nil {
		callback(nil, err)
		return
	}
	// Convert to DTOs
	result := make([]%s, len(items))
	for i, item := range items {
		result[i] = %s{
			ID: item.ID,
%s		}
	}
	callback(result, nil)
}

// Get returns a single %s by ID.
func (a *%sAPI) Get(id uint, callback func(*%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	var item models.%s
	if err := db.First(&item, id).Error; err != nil {
		callback(nil, err)
		return
	}
	result := &%s{
		ID: item.ID,
%s	}
	callback(result, nil)
}

// Create creates a new %s.
func (a *%sAPI) Create(item *%s, callback func(*%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	model := models.%s{
%s	}
	if err := db.Create(&model).Error; err != nil {
		callback(nil, err)
		return
	}
	result := &%s{
		ID: model.ID,
%s	}
	callback(result, nil)
}

// Update updates an existing %s.
func (a *%sAPI) Update(item *%s, callback func(*%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	var model models.%s
	if err := db.First(&model, item.ID).Error; err != nil {
		callback(nil, err)
		return
	}
%s	if err := db.Save(&model).Error; err != nil {
		callback(nil, err)
		return
	}
	result := &%s{
		ID: model.ID,
%s	}
	callback(result, nil)
}

// Delete deletes a %s by ID.
func (a *%sAPI) Delete(id uint, callback func(error)) {
	if db == nil {
		callback(nil)
		return
	}
	callback(db.Delete(&models.%s{}, id).Error)
}
`,
		pluralName, modelName, // type comment
		pluralName,            // type name
		pluralName, modelName, // var comment
		pluralName, pluralName, // var declaration
		modelName,                  // List comment
		pluralName, modelName,      // List signature
		modelName,                  // List query
		modelName, modelName,       // List convert type
		listFieldAssignments.String(), // List field assignments
		modelName,                  // Get comment
		pluralName, modelName,      // Get signature
		modelName,                  // Get query
		modelName,                  // Get result type
		getFieldAssignments.String(), // Get field assignments
		modelName,                      // Create comment
		pluralName, modelName, modelName, // Create signature
		modelName,                      // Create model type
		createModelAssignments.String(), // Create model assignments
		modelName,                       // Create result type
		createResultAssignments.String(), // Create result assignments
		modelName,                        // Update comment
		pluralName, modelName, modelName, // Update signature
		modelName,                       // Update query
		updateModelAssignments.String(), // Update model assignments
		modelName,                        // Update result type
		updateResultAssignments.String(), // Update result assignments
		modelName,                        // Delete comment
		pluralName,                       // Delete signature
		modelName,                        // Delete model
	))

	return sb.String()
}

// generateMinimalServerAPICode generates server-side API code with only ID field
// Used as fallback when model fields cannot be parsed
func generateMinimalServerAPICode(modelName, pluralName string) string {
	return fmt.Sprintf(`
// %sAPI provides CRUD operations for %s.
type %sAPI struct{}

// %s is the API client for %s operations.
var %s = &%sAPI{}

// List returns all %s records.
func (a *%sAPI) List(callback func([]%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	var items []models.%s
	if err := db.Find(&items).Error; err != nil {
		callback(nil, err)
		return
	}
	result := make([]%s, len(items))
	for i, item := range items {
		result[i] = %s{ID: item.ID}
	}
	callback(result, nil)
}

// Get returns a single %s by ID.
func (a *%sAPI) Get(id uint, callback func(*%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	var item models.%s
	if err := db.First(&item, id).Error; err != nil {
		callback(nil, err)
		return
	}
	callback(&%s{ID: item.ID}, nil)
}

// Create creates a new %s.
func (a *%sAPI) Create(item *%s, callback func(*%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	model := models.%s{}
	if err := db.Create(&model).Error; err != nil {
		callback(nil, err)
		return
	}
	callback(&%s{ID: model.ID}, nil)
}

// Update updates an existing %s.
func (a *%sAPI) Update(item *%s, callback func(*%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	var model models.%s
	if err := db.First(&model, item.ID).Error; err != nil {
		callback(nil, err)
		return
	}
	if err := db.Save(&model).Error; err != nil {
		callback(nil, err)
		return
	}
	callback(&%s{ID: model.ID}, nil)
}

// Delete deletes a %s by ID.
func (a *%sAPI) Delete(id uint, callback func(error)) {
	if db == nil {
		callback(nil)
		return
	}
	callback(db.Delete(&models.%s{}, id).Error)
}
`,
		pluralName, modelName, // type comment
		pluralName,            // type name
		pluralName, modelName, // var comment
		pluralName, pluralName, // var declaration
		modelName,             // List comment
		pluralName, modelName, // List signature
		modelName,             // List query
		modelName, modelName,  // List convert
		modelName,             // Get comment
		pluralName, modelName, // Get signature
		modelName,             // Get query
		modelName,             // Get result
		modelName,                    // Create comment
		pluralName, modelName, modelName, // Create signature
		modelName, // Create model
		modelName, // Create result
		modelName,                    // Update comment
		pluralName, modelName, modelName, // Update signature
		modelName, // Update query
		modelName, // Update result
		modelName, // Delete comment
		pluralName,
		modelName, // Delete model
	)
}

// getModelFieldTypes parses a model file and returns a map of field name -> field type
// This is used to determine if foreign key fields are pointers
func getModelFieldTypes(modelName string) (map[string]string, error) {
	// Check cache first
	if cached, ok := modelFieldTypesCache[modelName]; ok {
		return cached, nil
	}

	// Parse models directory (guxgen/models for generated, models for manual)
	entries, err := os.ReadDir(filepath.Join("guxgen", "models"))
	if err != nil {
		// Fall back to models/ for backwards compatibility
		entries, err = os.ReadDir("models")
		if err != nil {
			return nil, fmt.Errorf("read models dir: %w", err)
		}
	}

	modelsDir := filepath.Join("guxgen", "models")
	if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
		modelsDir = "models"
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		fset := token.NewFileSet()
		filename := filepath.Join(modelsDir, entry.Name())
		node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		// Look for the target struct
		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != modelName {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				fieldTypes := make(map[string]string)
				for _, field := range structType.Fields.List {
					if len(field.Names) == 0 {
						continue
					}
					fieldName := field.Names[0].Name
					fieldTypes[fieldName] = formatType(field.Type)
				}

				// Cache and return
				modelFieldTypesCache[modelName] = fieldTypes
				return fieldTypes, nil
			}
		}
	}

	return nil, fmt.Errorf("model %s not found", modelName)
}

// splitParamRoute splits a parameterized route into prefix and suffix around the parameter.
// "/admin/users/:id" -> ("/admin/users/", "")
// "/admin/users/:id/posts/new" -> ("/admin/users/", "/posts/new")
func splitParamRoute(route string) (prefix, suffix string) {
	parts := strings.Split(route, "/")
	var prefixParts, suffixParts []string
	foundParam := false

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			foundParam = true
			continue
		}
		if foundParam {
			suffixParts = append(suffixParts, part)
		} else {
			prefixParts = append(prefixParts, part)
		}
	}

	prefix = strings.Join(prefixParts, "/")
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	if len(suffixParts) > 0 {
		suffix = "/" + strings.Join(suffixParts, "/")
	}

	return prefix, suffix
}

// getParamName extracts the parameter name from a parameterized route.
// "/admin/users/:id" -> "id"
// "/admin/users/:userId/posts/:postId" -> "userId" (returns first param only)
func getParamName(route string) string {
	parts := strings.Split(route, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			return strings.TrimPrefix(part, ":")
		}
	}
	return "id" // default fallback
}

// parseDTOFile parses a DTO file and extracts field mappings from gux tags
func parseDTOFile(dtoDir string, dtoName string) (*DTOInfo, error) {
	// Find and parse all .go files in the dto directory
	entries, err := os.ReadDir(dtoDir)
	if err != nil {
		return nil, fmt.Errorf("read dto dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		fset := token.NewFileSet()
		filename := filepath.Join(dtoDir, entry.Name())
		node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		// Look for the target struct
		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != dtoName {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				info := &DTOInfo{
					Name:   dtoName,
					Fields: []DTOFieldMapping{},
				}

				// Parse each field
				for _, field := range structType.Fields.List {
					if len(field.Names) == 0 || field.Tag == nil {
						continue
					}

					fieldName := field.Names[0].Name
					tagValue := field.Tag.Value

					// Parse the gux tag: `gux:"User.Email"`
					guxTag := extractTag(tagValue, "gux")
					if guxTag == "" {
						continue
					}

					// Parse Model.Field format
					parts := strings.SplitN(guxTag, ".", 2)
					if len(parts) != 2 {
						continue
					}

					modelName := parts[0]
					modelField := parts[1]

					// Set the model name from first field
					if info.ModelName == "" {
						info.ModelName = modelName
					}

					mapping := DTOFieldMapping{
						DTOField:   fieldName,
						DTOType:    formatType(field.Type),
						ModelField: modelField,
					}

					// Check for preload tag
					preloadTag := extractTag(tagValue, "preload")
					if preloadTag != "" {
						mapping.Preload = preloadTag
						info.Preloads = append(info.Preloads, preloadTag)
					}

					// Check if it's a slice type (for relationships)
					if arrayType, ok := field.Type.(*ast.ArrayType); ok {
						mapping.IsSlice = true
						mapping.SliceDTO = formatType(arrayType.Elt)
						// Auto-detect preload from model field name if not explicitly set
						if preloadTag == "" && modelField != "" {
							mapping.Preload = modelField
							info.Preloads = append(info.Preloads, modelField)
						}
					} else if !isPrimitiveType(mapping.DTOType) {
						// Non-primitive, non-slice type is a nested DTO
						mapping.IsNestedDTO = true
						// Auto-detect preload from model field name if not explicitly set
						if preloadTag == "" && modelField != "" {
							mapping.Preload = modelField
							info.Preloads = append(info.Preloads, modelField)
						}
					}

					info.Fields = append(info.Fields, mapping)
				}

				return info, nil
			}
		}
	}

	return nil, fmt.Errorf("DTO %s not found in %s", dtoName, dtoDir)
}

// extractTag extracts a specific tag value from a struct tag string
func extractTag(tagStr, tagName string) string {
	// Remove backticks
	tagStr = strings.Trim(tagStr, "`")

	// Find the tag
	for _, part := range strings.Split(tagStr, " ") {
		if strings.HasPrefix(part, tagName+":") {
			value := strings.TrimPrefix(part, tagName+":")
			return strings.Trim(value, `"`)
		}
	}
	return ""
}

// formatType converts an ast type expression to a string
func formatType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return formatType(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + formatType(t.X)
	case *ast.ArrayType:
		return "[]" + formatType(t.Elt)
	default:
		return "interface{}"
	}
}

// parseCRUDModels parses the main app file to find CRUD model registrations
func parseCRUDModels(filename string) ([]CRUDModel, string, string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, "", "", fmt.Errorf("parse %s: %w", filename, err)
	}

	var models []CRUDModel
	var modelsImport string
	var dtoImport string

	// Find models and dto imports
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if strings.HasSuffix(path, "/models") || strings.Contains(path, "/models") {
			modelsImport = path
		}
		if strings.HasSuffix(path, "/dto") || strings.Contains(path, "/dto") {
			dtoImport = path
		}
	}

	// Find app.CRUD() calls
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Look for .CRUD( calls
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "CRUD" {
			return true
		}

		if len(call.Args) >= 1 {
			var model CRUDModel

			// Get model name from first argument: models.Counter{}
			switch arg := call.Args[0].(type) {
			case *ast.CompositeLit:
				// models.Counter{}
				if sel, ok := arg.Type.(*ast.SelectorExpr); ok {
					model.Name = sel.Sel.Name
				}
			case *ast.SelectorExpr:
				// models.Counter
				model.Name = arg.Sel.Name
			}

			if model.Name == "" {
				return true
			}

			model.PluralName = pluralize(model.Name)
			model.Path = strings.ToLower(pluralize(model.Name))

			// Parse DTO options from remaining arguments
			for i := 1; i < len(call.Args); i++ {
				optCall, ok := call.Args[i].(*ast.CallExpr)
				if !ok {
					continue
				}
				// Look for core.WithListDTO or core.WithDetailDTO
				optSel, ok := optCall.Fun.(*ast.SelectorExpr)
				if !ok {
					continue
				}

				dtoName := ""
				if len(optCall.Args) >= 1 {
					// Get DTO type name: dto.UserList{}
					switch dtoArg := optCall.Args[0].(type) {
					case *ast.CompositeLit:
						if dtoSel, ok := dtoArg.Type.(*ast.SelectorExpr); ok {
							dtoName = dtoSel.Sel.Name
							if pkgIdent, ok := dtoSel.X.(*ast.Ident); ok {
								model.DTOPackage = pkgIdent.Name
							}
						}
					case *ast.SelectorExpr:
						dtoName = dtoArg.Sel.Name
						if pkgIdent, ok := dtoArg.X.(*ast.Ident); ok {
							model.DTOPackage = pkgIdent.Name
						}
					}
				}

				switch optSel.Sel.Name {
				case "WithListDTO":
					model.ListDTO = dtoName
				case "WithDetailDTO":
					model.DetailDTO = dtoName
				case "WithDTO":
					model.ListDTO = dtoName
					model.DetailDTO = dtoName
				}
			}

			models = append(models, model)
		}
		return true
	})

	return models, modelsImport, dtoImport, nil
}

// parseAPIEndpoints parses the main app file to find core.API, core.APIGet, and core.APIDelete calls
func parseAPIEndpoints(filename string) ([]APIEndpointInfo, string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, "", fmt.Errorf("parse %s: %w", filename, err)
	}

	var endpoints []APIEndpointInfo
	var dtoImport string

	// Find dto import
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if strings.HasSuffix(path, "/dto") || strings.Contains(path, "/dto") {
			dtoImport = path
			break
		}
	}

	// Find core.API, core.APIGet, core.APIDelete calls
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Look for core.API, core.APIGet, core.APIDelete calls
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// Must be from core package
		ident, ok := sel.X.(*ast.Ident)
		if !ok || ident.Name != "core" {
			return true
		}

		var endpoint APIEndpointInfo

		switch sel.Sel.Name {
		case "API":
			// core.API(app, "POST", "/api/login", handler)
			if len(call.Args) >= 4 {
				endpoint = parseAPICall(call.Args)
			}
		case "APIGet":
			// core.APIGet(app, "/api/users/:id", handler)
			if len(call.Args) >= 3 {
				endpoint = parseAPIGetCall(call.Args)
			}
		case "APIDelete":
			// core.APIDelete(app, "/api/users/:id", handler)
			if len(call.Args) >= 3 {
				endpoint = parseAPIDeleteCall(call.Args)
			}
		default:
			return true
		}

		if endpoint.Path != "" {
			endpoints = append(endpoints, endpoint)
		}

		return true
	})

	return endpoints, dtoImport, nil
}

// parseAPICall parses core.API(app, "POST", "/api/login", handler) call
func parseAPICall(args []ast.Expr) APIEndpointInfo {
	var ep APIEndpointInfo

	// arg[1] = method string
	if lit, ok := args[1].(*ast.BasicLit); ok {
		ep.Method = strings.Trim(lit.Value, `"`)
	}

	// arg[2] = path string
	if lit, ok := args[2].(*ast.BasicLit); ok {
		ep.Path = strings.Trim(lit.Value, `"`)
		ep.PathParams = extractPathParams(ep.Path)
		ep.FuncName = generateFuncName(ep.Method, ep.Path)
	}

	// arg[3] = handler function - extract request/response types from signature
	if funcLit, ok := args[3].(*ast.FuncLit); ok {
		// Handler: func(ctx *core.APIContext, req LoginRequest) (LoginResponse, error)
		if funcLit.Type.Params != nil && len(funcLit.Type.Params.List) >= 2 {
			// Second param is the request type
			reqParam := funcLit.Type.Params.List[1]
			ep.RequestType, ep.Package = extractTypeName(reqParam.Type)
		}
		if funcLit.Type.Results != nil && len(funcLit.Type.Results.List) >= 1 {
			// First result is the response type
			var respPkg string
			ep.ResponseType, respPkg = extractTypeName(funcLit.Type.Results.List[0].Type)
			// Use response package if request package is empty (e.g., struct{} input)
			if ep.Package == "" && respPkg != "" {
				ep.Package = respPkg
			}
		}
	}

	return ep
}

// parseAPIGetCall parses core.APIGet(app, "/api/users/:id", handler) call
func parseAPIGetCall(args []ast.Expr) APIEndpointInfo {
	var ep APIEndpointInfo
	ep.Method = "GET"

	// arg[1] = path string
	if lit, ok := args[1].(*ast.BasicLit); ok {
		ep.Path = strings.Trim(lit.Value, `"`)
		ep.PathParams = extractPathParams(ep.Path)
		ep.FuncName = generateFuncName(ep.Method, ep.Path)
	}

	// arg[2] = handler function - extract response type from signature
	if funcLit, ok := args[2].(*ast.FuncLit); ok {
		// Handler: func(ctx *core.APIContext) (dto.UserDetail, error)
		if funcLit.Type.Results != nil && len(funcLit.Type.Results.List) >= 1 {
			ep.ResponseType, ep.Package = extractTypeName(funcLit.Type.Results.List[0].Type)
		}
	}

	return ep
}

// parseAPIDeleteCall parses core.APIDelete(app, "/api/users/:id", handler) call
func parseAPIDeleteCall(args []ast.Expr) APIEndpointInfo {
	var ep APIEndpointInfo
	ep.Method = "DELETE"

	// arg[1] = path string
	if lit, ok := args[1].(*ast.BasicLit); ok {
		ep.Path = strings.Trim(lit.Value, `"`)
		ep.PathParams = extractPathParams(ep.Path)
		ep.FuncName = generateFuncName(ep.Method, ep.Path)
	}

	// DELETE has no request body and typically no response body
	return ep
}

// extractTypeName extracts the type name from an AST expression
func extractTypeName(expr ast.Expr) (typeName, pkg string) {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name, ""
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return t.Sel.Name, ident.Name
		}
		return t.Sel.Name, ""
	case *ast.StarExpr:
		return extractTypeName(t.X)
	case *ast.ArrayType:
		elemType, elemPkg := extractTypeName(t.Elt)
		return "[]" + elemType, elemPkg
	}
	return "", ""
}

// extractPathParams extracts path parameter names from a path pattern
// e.g., "/api/users/:id/posts/:postId" -> ["id", "postId"]
func extractPathParams(path string) []string {
	var params []string
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			params = append(params, strings.TrimPrefix(part, ":"))
		}
	}
	return params
}

// generateFuncName generates a function name from method and path
// e.g., POST /api/login -> Login, GET /api/users/:id -> GetUser, DELETE /api/users/:id -> DeleteUser
func generateFuncName(method, path string) string {
	// Remove /api prefix if present
	path = strings.TrimPrefix(path, "/api/")
	path = strings.TrimPrefix(path, "/")

	// Split path and take meaningful parts
	parts := strings.Split(path, "/")
	var nameParts []string

	for _, part := range parts {
		if part == "" {
			continue
		}
		// Skip path parameters
		if strings.HasPrefix(part, ":") {
			continue
		}
		// Capitalize first letter
		if len(part) > 0 {
			nameParts = append(nameParts, strings.ToUpper(part[:1])+part[1:])
		}
	}

	name := strings.Join(nameParts, "")

	// Add method prefix for non-POST methods
	switch method {
	case "GET":
		if !strings.HasPrefix(name, "Get") {
			name = "Get" + name
		}
	case "DELETE":
		if !strings.HasPrefix(name, "Delete") {
			name = "Delete" + name
		}
	case "PUT":
		if !strings.HasPrefix(name, "Update") {
			name = "Update" + name
		}
	case "PATCH":
		if !strings.HasPrefix(name, "Patch") {
			name = "Patch" + name
		}
	}

	// Handle singular for endpoints with :id
	if len(nameParts) > 0 && strings.HasSuffix(nameParts[len(nameParts)-1], "s") {
		// Check if path has an :id parameter after this resource
		if strings.Contains(path, "/:") {
			// Make singular: users/:id -> User
			lastPart := nameParts[len(nameParts)-1]
			if strings.HasSuffix(lastPart, "s") {
				singular := lastPart[:len(lastPart)-1]
				name = strings.TrimSuffix(name, lastPart) + singular
			}
		}
	}

	if name == "" {
		name = "Root"
	}

	return name
}

// parseRoutes parses the main app file to find Hybrid routes
func parseRoutes(filename string) ([]PageRoute, string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, "", fmt.Errorf("parse %s: %w", filename, err)
	}

	var routes []PageRoute
	var pagesImport string

	// Find pages import
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if strings.HasSuffix(path, "/pages") || strings.Contains(path, "/pages") {
			pagesImport = path
			break
		}
	}

	// Find Hybrid() calls - simplified parsing
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Look for .Hybrid( calls
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Hybrid" {
			return true
		}

		if len(call.Args) >= 2 {
			// Get path (first arg)
			if lit, ok := call.Args[0].(*ast.BasicLit); ok {
				path := strings.Trim(lit.Value, `"`)

				// Get handler (second arg)
				handler := ""
				switch h := call.Args[1].(type) {
				case *ast.SelectorExpr:
					if ident, ok := h.X.(*ast.Ident); ok {
						handler = ident.Name + "." + h.Sel.Name
					}
				case *ast.Ident:
					handler = h.Name
				}

				if handler != "" {
					routes = append(routes, PageRoute{
						Path:     path,
						Handler:  handler,
						IsHybrid: true,
					})
				}
			}
		}
		return true
	})

	return routes, pagesImport, nil
}

// parseRoutesAndBundles parses the main app file to find all routes and bundles
func parseRoutesAndBundles(filename string) (map[string]*BundleInfo, map[string]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("parse %s: %w", filename, err)
	}

	// Bundle name -> BundleInfo
	bundles := make(map[string]*BundleInfo)
	bundles["app"] = &BundleInfo{Name: "app"} // Default bundle

	// Package alias -> import path (e.g., "pages" -> "github.com/.../pages")
	imports := make(map[string]string)

	// Find all imports
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		var alias string
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			parts := strings.Split(path, "/")
			alias = parts[len(parts)-1]
		}
		imports[alias] = path
	}

	// Helper to extract RouteGroup info from a call expression
	extractRouteGroupInfo := func(call *ast.CallExpr) (prefix string, bundle string, found bool) {
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "RouteGroup" {
			return "", "", false
		}

		// Get prefix (first argument)
		if len(call.Args) >= 1 {
			if lit, ok := call.Args[0].(*ast.BasicLit); ok {
				prefix = strings.Trim(lit.Value, `"`)
			}
		}

		// Check for WithBundle option in remaining arguments
		for i := 1; i < len(call.Args); i++ {
			if callArg, ok := call.Args[i].(*ast.CallExpr); ok {
				if selArg, ok := callArg.Fun.(*ast.SelectorExpr); ok {
					if selArg.Sel.Name == "WithBundle" && len(callArg.Args) >= 1 {
						if lit, ok := callArg.Args[0].(*ast.BasicLit); ok {
							bundle = strings.Trim(lit.Value, `"`)
						}
					}
				}
			}
		}

		return prefix, bundle, true
	}

	// Helper to walk up the call chain to find RouteGroup
	findRouteGroupInChain := func(call *ast.CallExpr) (prefix string, bundle string, found bool) {
		current := call
		for {
			sel, ok := current.Fun.(*ast.SelectorExpr)
			if !ok {
				return "", "", false
			}

			// Check if the receiver is a call expression
			chainedCall, ok := sel.X.(*ast.CallExpr)
			if !ok {
				return "", "", false
			}

			// Check if this call is RouteGroup
			if prefix, bundle, found := extractRouteGroupInfo(chainedCall); found {
				return prefix, bundle, true
			}

			// Check if the chained call is a method call (Hybrid, GET, etc.)
			chainedSel, ok := chainedCall.Fun.(*ast.SelectorExpr)
			if !ok {
				return "", "", false
			}

			// If it's another Hybrid/GET/POST/Protected/etc., keep walking up
			switch chainedSel.Sel.Name {
			case "Hybrid", "GET", "POST", "PUT", "PATCH", "DELETE", "Protected", "RequireRole":
				current = chainedCall
				continue
			}

			return "", "", false
		}
	}

	// Find all Hybrid calls
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// Only process Hybrid calls
		if sel.Sel.Name != "Hybrid" || len(call.Args) < 2 {
			return true
		}

		// Get path
		path := ""
		if lit, ok := call.Args[0].(*ast.BasicLit); ok {
			path = strings.Trim(lit.Value, `"`)
		}

		// Get handler - handles direct handlers (pages.Home) and wrapped handlers (pages.Layout(pages.Home))
		handler := ""
		pkgAlias := ""
		switch h := call.Args[1].(type) {
		case *ast.SelectorExpr:
			if ident, ok := h.X.(*ast.Ident); ok {
				pkgAlias = ident.Name
				handler = ident.Name + "." + h.Sel.Name
			}
		case *ast.Ident:
			handler = h.Name
		case *ast.CallExpr:
			// Wrapped handler like pages.AdminLayout(pages.Dashboard)
			// Extract the outer function: pages.AdminLayout
			if sel, ok := h.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					pkgAlias = ident.Name
					handler = ident.Name + "." + sel.Sel.Name
					// Add the inner call as argument: (pages.Dashboard)
					if len(h.Args) > 0 {
						if innerSel, ok := h.Args[0].(*ast.SelectorExpr); ok {
							if innerIdent, ok := innerSel.X.(*ast.Ident); ok {
								handler += "(" + innerIdent.Name + "." + innerSel.Sel.Name + ")"
							}
						}
					}
				}
			}
		}

		if handler == "" {
			return true
		}

		// Determine bundle by walking up the call chain
		bundleName := "app"
		prefix, bundle, found := findRouteGroupInChain(call)
		if found {
			if bundle != "" {
				bundleName = bundle
				// Ensure bundle exists
				if _, exists := bundles[bundleName]; !exists {
					bundles[bundleName] = &BundleInfo{Name: bundleName}
				}
			}
			// Apply prefix to path
			if path == "/" {
				path = prefix
			} else {
				path = prefix + path
			}
		}

		route := PageRoute{
			Path:     path,
			Handler:  handler,
			IsHybrid: true,
			Bundle:   bundleName,
		}

		// Check for duplicate path in this bundle (skip if already exists)
		isDuplicate := false
		for _, existing := range bundles[bundleName].Routes {
			if existing.Path == path {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			bundles[bundleName].Routes = append(bundles[bundleName].Routes, route)
		}

		// Track import needed for this bundle
		if pkgAlias != "" {
			if importPath, ok := imports[pkgAlias]; ok {
				found := false
				for _, imp := range bundles[bundleName].Imports {
					if imp == importPath {
						found = true
						break
					}
				}
				if !found {
					bundles[bundleName].Imports = append(bundles[bundleName].Imports, importPath)
				}
			}
		}

		return true
	})

	return bundles, imports, nil
}

// generateWasmEntryPoint generates guxgen/wasm/main.go
func generateWasmEntryPoint(modulePath, pagesImport string, routes []PageRoute) error {
	if err := os.MkdirAll("guxgen/wasm", 0755); err != nil {
		return err
	}

	// Build the router/page matching code (same logic as bundle version)
	// Note: 'path' is already declared in loadPage(), so we just use it here
	var routeCode strings.Builder
	if len(routes) == 1 && !strings.Contains(routes[0].Path, ":") {
		// Single non-parameterized route - simple case
		routeCode.WriteString(fmt.Sprintf("\t\tcomponent := %s(router)\n", routes[0].Handler))
	} else {
		// Multiple routes or parameterized routes - need pattern matching
		// Note: 'path' is already declared in loadPage()
		routeCode.WriteString("\t\tvar component func() core.Node\n")

		// Separate exact routes from parameterized routes
		var exactRoutes, paramRoutes []PageRoute
		for _, route := range routes {
			if strings.Contains(route.Path, ":") {
				paramRoutes = append(paramRoutes, route)
			} else {
				exactRoutes = append(exactRoutes, route)
			}
		}

		// Generate switch for exact matches first
		if len(exactRoutes) > 0 {
			routeCode.WriteString("\t\tswitch path {\n")
			for _, route := range exactRoutes {
				routeCode.WriteString(fmt.Sprintf("\t\tcase %q:\n", route.Path))
				routeCode.WriteString("\t\t\trouter.SetRouteParams(map[string]string{\"__path\": path})\n")
				routeCode.WriteString(fmt.Sprintf("\t\t\tcomponent = %s(router)\n", route.Handler))
			}
			routeCode.WriteString("\t\t}\n")
		}

		// Generate if-else for parameterized routes
		if len(paramRoutes) > 0 {
			routeCode.WriteString("\t\tif component == nil {\n")
			for i, route := range paramRoutes {
				prefix, suffix := splitParamRoute(route.Path)
				paramName := getParamName(route.Path)
				if i == 0 {
					routeCode.WriteString(fmt.Sprintf("\t\t\tif matchRoute(path, %q, %q) {\n", prefix, suffix))
				} else {
					routeCode.WriteString(fmt.Sprintf("\t\t\t} else if matchRoute(path, %q, %q) {\n", prefix, suffix))
				}
				// Extract and set route params before rendering (include __path for Path() method)
				routeCode.WriteString(fmt.Sprintf("\t\t\t\trouter.SetRouteParams(map[string]string{\"__path\": path, %q: extractRouteParam(path, %q, %q)})\n", paramName, prefix, suffix))
				routeCode.WriteString(fmt.Sprintf("\t\t\t\tcomponent = %s(router)\n", route.Handler))
			}
			routeCode.WriteString("\t\t\t}\n")
			routeCode.WriteString("\t\t}\n")
		}

		// Default fallback
		routeCode.WriteString("\t\tif component == nil {\n")
		defaultHandler := ""
		for _, r := range routes {
			if r.Path == "/" {
				defaultHandler = r.Handler
				break
			}
		}
		if defaultHandler == "" && len(routes) > 0 {
			defaultHandler = routes[0].Handler
		}
		if defaultHandler != "" {
			routeCode.WriteString(fmt.Sprintf("\t\t\tcomponent = %s(router)\n", defaultHandler))
		}
		routeCode.WriteString("\t\t}\n")
	}

	code := fmt.Sprintf(`//go:build js && wasm

package main

import (
	"encoding/json"
	"strings"
	"syscall/js"

	"github.com/dougbarrett/gux/core"
	"%s"
)

// matchRoute checks if a path matches a parameterized route pattern.
func matchRoute(path, prefix, suffix string) bool {
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	rest := path[len(prefix):]
	if suffix == "" {
		return len(rest) > 0 && !strings.Contains(rest, "/")
	}
	if !strings.HasSuffix(rest, suffix) {
		return false
	}
	paramValue := rest[:len(rest)-len(suffix)]
	return len(paramValue) > 0 && !strings.Contains(paramValue, "/")
}

// extractRouteParam extracts the parameter value from a path given prefix and suffix.
func extractRouteParam(path, prefix, suffix string) string {
	rest := path[len(prefix):]
	if suffix == "" {
		return rest
	}
	return rest[:len(rest)-len(suffix)]
}

// fetchLoader fetches page state from loader endpoint
func fetchLoader(path string, callback func(map[string]any)) {
	loaderPath := "/__gux_api/pages" + path
	if path == "/" {
		loaderPath = "/__gux_api/pages/index"
	}

	promise := js.Global().Call("fetch", loaderPath)
	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
		resp := args[0]
		if resp.Get("ok").Bool() {
			resp.Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
				// Parse JSON response properly to preserve all types (arrays, objects, etc.)
				jsonStr := args[0].String()
				var state map[string]any
				if err := json.Unmarshal([]byte(jsonStr), &state); err == nil {
					callback(state)
				} else {
					callback(nil)
				}
				return nil
			}))
		} else {
			callback(nil)
		}
		return nil
	}))
}

func main() {
	// Debug: use println which TinyGo reliably outputs to console
	println("[Gux WASM] main() started")

	document := js.Global().Get("document")
	window := js.Global()
	container := document.Call("getElementById", "app")

	var router *core.Router
	var render func()

	// Cached page component - persists across re-renders within same page
	// This ensures closures (like OnChange handlers) reference stable state objects
	var currentComponent func() core.Node
	var currentPath string

	// loadPage creates the page component for the current path
	// Called on navigation and initial load, NOT on state-triggered re-renders
	loadPage := func() {
		path := js.Global().Get("location").Get("pathname").String()
		// Only reload component if path changed (handles state-triggered re-renders)
		if currentComponent != nil && path == currentPath {
			return
		}
		currentPath = path
%s
		currentComponent = component
	}

	render = func() {
		// Save focus state before re-render
		activeElement := document.Get("activeElement")
		var focusName string
		if !activeElement.IsNull() && !activeElement.IsUndefined() {
			focusName = activeElement.Call("getAttribute", "name").String()
		}

		container.Set("innerHTML", "")

		// Use cached component - closures remain stable across re-renders
		if currentComponent != nil {
			node := currentComponent()
			result := node.Render(core.DOM())
			if domVal := result.DOMValue(); domVal != nil {
				container.Call("appendChild", domVal.(js.Value))
			}
		}

		// Intercept link clicks for client-side navigation
		links := document.Call("querySelectorAll", "a[href]")
		for i := 0; i < links.Get("length").Int(); i++ {
			link := links.Call("item", i)
			href := link.Get("href").String()
			origin := window.Get("location").Get("origin").String()

			// Skip external links (marked with data-gux-external)
			if link.Call("getAttribute", "data-gux-external").String() == "true" {
				continue
			}

			// Only intercept internal links
			if len(href) >= len(origin) && href[:len(origin)] == origin {
				link.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
					args[0].Call("preventDefault")
					path := this.Get("pathname").String()
					router.Navigate(path)
					return nil
				}))
			}
		}

		// Restore focus after re-render
		if focusName != "" {
			newElement := document.Call("querySelector", "[name=\""+focusName+"\"]")
			if !newElement.IsNull() && !newElement.IsUndefined() {
				newElement.Call("focus")
			}
		}
	}

	// Navigate fetches page data then renders
	navigate := func(path string) {
		// Clear current component to force reload for new path
		currentComponent = nil
		// Clear page-specific state before navigating to avoid stale data
		router.ClearState()
		window.Get("history").Call("pushState", nil, "", path)
		fetchLoader(path, func(state map[string]any) {
			if state != nil {
				router.Hydrate(state)
			}
			loadPage()
			render()
		})
	}

	router = core.NewRouter(render)
	router.SetNavigate(navigate)
	println("[Gux WASM] Calling SetDebugRouter")
	core.SetDebugRouter(router) // Enable state debugging
	println("[Gux WASM] SetDebugRouter completed")

	// Hydrate state from SSR
	stateEl := document.Call("getElementById", "__gux_state")
	if !stateEl.IsNull() && !stateEl.IsUndefined() {
		stateJSON := stateEl.Get("textContent").String()
		var state map[string]any
		if err := json.Unmarshal([]byte(stateJSON), &state); err == nil {
			router.Hydrate(state)
		}
	}

	// Handle browser back/forward
	window.Call("addEventListener", "popstate", js.FuncOf(func(this js.Value, args []js.Value) any {
		currentComponent = nil // Force reload for history navigation
		// Clear page-specific state for browser history navigation
		router.ClearState()
		// Fetch fresh data for the new/previous page
		path := window.Get("location").Get("pathname").String()
		fetchLoader(path, func(state map[string]any) {
			if state != nil {
				router.Hydrate(state)
			}
			loadPage()
			render()
		})
		return nil
	}))

	// Initial load
	loadPage()
	render()

	select {}
}
`, pagesImport, routeCode.String())

	return os.WriteFile("guxgen/wasm/main.go", []byte(code), 0644)
}

// generateBundleWasmEntryPoint generates a WASM entry point for a specific bundle
func generateBundleWasmEntryPoint(bundleName string, bundle *BundleInfo) error {
	wasmDir := "guxgen/wasm_" + bundleName
	if bundleName == "app" {
		wasmDir = "guxgen/wasm"
	}

	if err := os.MkdirAll(wasmDir, 0755); err != nil {
		return err
	}

	routes := bundle.Routes
	imports := bundle.Imports

	// Build the router/page matching code
	// Note: 'path' is already declared in loadPage(), so we just use it here
	var routeCode strings.Builder
	if len(routes) == 1 && !strings.Contains(routes[0].Path, ":") {
		// Single non-parameterized route - simple case
		routeCode.WriteString(fmt.Sprintf("\t\tcomponent := %s(router)\n", routes[0].Handler))
	} else {
		// Multiple routes or parameterized routes - need pattern matching
		// Note: 'path' is already declared in loadPage()
		routeCode.WriteString("\t\tvar component func() core.Node\n")

		// Separate exact routes from parameterized routes
		var exactRoutes, paramRoutes []PageRoute
		for _, route := range routes {
			if strings.Contains(route.Path, ":") {
				paramRoutes = append(paramRoutes, route)
			} else {
				exactRoutes = append(exactRoutes, route)
			}
		}

		// Generate switch for exact matches first
		if len(exactRoutes) > 0 {
			routeCode.WriteString("\t\tswitch path {\n")
			for _, route := range exactRoutes {
				routeCode.WriteString(fmt.Sprintf("\t\tcase %q:\n", route.Path))
				routeCode.WriteString("\t\t\trouter.SetRouteParams(map[string]string{\"__path\": path})\n")
				routeCode.WriteString(fmt.Sprintf("\t\t\tcomponent = %s(router)\n", route.Handler))
			}
			routeCode.WriteString("\t\t}\n")
		}

		// Generate if-else for parameterized routes
		if len(paramRoutes) > 0 {
			routeCode.WriteString("\t\tif component == nil {\n")
			for i, route := range paramRoutes {
				// Generate matching code for parameterized route
				prefix, suffix := splitParamRoute(route.Path)
				paramName := getParamName(route.Path)
				if i == 0 {
					routeCode.WriteString(fmt.Sprintf("\t\t\tif matchRoute(path, %q, %q) {\n", prefix, suffix))
				} else {
					routeCode.WriteString(fmt.Sprintf("\t\t\t} else if matchRoute(path, %q, %q) {\n", prefix, suffix))
				}
				// Extract and set route params before rendering (include __path for Path() method)
				routeCode.WriteString(fmt.Sprintf("\t\t\t\trouter.SetRouteParams(map[string]string{\"__path\": path, %q: extractRouteParam(path, %q, %q)})\n", paramName, prefix, suffix))
				routeCode.WriteString(fmt.Sprintf("\t\t\t\tcomponent = %s(router)\n", route.Handler))
			}
			routeCode.WriteString("\t\t\t}\n")
			routeCode.WriteString("\t\t}\n")
		}

		// Default fallback
		routeCode.WriteString("\t\tif component == nil {\n")
		defaultHandler := ""
		for _, r := range routes {
			if r.Path == "/" || strings.HasSuffix(r.Path, "/") {
				defaultHandler = r.Handler
				break
			}
		}
		if defaultHandler == "" && len(routes) > 0 {
			defaultHandler = routes[0].Handler
		}
		if defaultHandler != "" {
			routeCode.WriteString(fmt.Sprintf("\t\t\tcomponent = %s(router)\n", defaultHandler))
		}
		routeCode.WriteString("\t\t}\n")
	}

	// Build imports section
	var importSection strings.Builder
	for _, imp := range imports {
		parts := strings.Split(imp, "/")
		alias := parts[len(parts)-1]
		importSection.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
		_ = alias // Alias inferred from path
	}

	// Build route patterns for cross-bundle navigation detection
	var bundleRoutesCode strings.Builder
	bundleRoutesCode.WriteString("// bundleRoutePatterns contains all route patterns in this bundle\n")
	bundleRoutesCode.WriteString("var bundleRoutePatterns = []struct {\n")
	bundleRoutesCode.WriteString("\texact  string // Exact path match (empty if parameterized)\n")
	bundleRoutesCode.WriteString("\tprefix string // Prefix for parameterized routes\n")
	bundleRoutesCode.WriteString("\tsuffix string // Suffix for parameterized routes\n")
	bundleRoutesCode.WriteString("}{\n")
	for _, route := range routes {
		if strings.Contains(route.Path, ":") {
			prefix, suffix := splitParamRoute(route.Path)
			bundleRoutesCode.WriteString(fmt.Sprintf("\t{prefix: %q, suffix: %q},\n", prefix, suffix))
		} else {
			bundleRoutesCode.WriteString(fmt.Sprintf("\t{exact: %q},\n", route.Path))
		}
	}
	bundleRoutesCode.WriteString("}\n")

	code := fmt.Sprintf(`//go:build js && wasm

package main

import (
	"encoding/json"
	"strings"
	"syscall/js"

	"github.com/dougbarrett/gux/core"
%s)

// matchRoute checks if a path matches a parameterized route pattern.
// prefix is the part before the param, suffix is the part after.
// e.g., matchRoute("/admin/users/123", "/admin/users/", "") returns true
// e.g., matchRoute("/admin/users/123/posts/new", "/admin/users/", "/posts/new") returns true
func matchRoute(path, prefix, suffix string) bool {
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	rest := path[len(prefix):]
	if suffix == "" {
		// No suffix - just need something after the prefix (the param value)
		return len(rest) > 0 && !strings.Contains(rest, "/")
	}
	// Has suffix - need to match suffix at the end
	if !strings.HasSuffix(rest, suffix) {
		return false
	}
	// Extract the param value (between prefix and suffix)
	paramValue := rest[:len(rest)-len(suffix)]
	return len(paramValue) > 0 && !strings.Contains(paramValue, "/")
}

// extractRouteParam extracts the parameter value from a path given prefix and suffix.
// e.g., extractRouteParam("/admin/users/123", "/admin/users/", "") returns "123"
// e.g., extractRouteParam("/admin/users/123/posts/new", "/admin/users/", "/posts/new") returns "123"
func extractRouteParam(path, prefix, suffix string) string {
	rest := path[len(prefix):]
	if suffix == "" {
		return rest
	}
	return rest[:len(rest)-len(suffix)]
}

%s

// isRouteInBundle checks if a path matches any route pattern in this bundle.
// Returns true if the path should be handled by this bundle's router.
func isRouteInBundle(path string) bool {
	for _, pattern := range bundleRoutePatterns {
		if pattern.exact != "" {
			// Exact match
			if path == pattern.exact {
				return true
			}
		} else {
			// Parameterized match
			if matchRoute(path, pattern.prefix, pattern.suffix) {
				return true
			}
		}
	}
	return false
}

// fetchLoader fetches page state from loader endpoint
func fetchLoader(path string, callback func(map[string]any)) {
	loaderPath := "/__gux_api/pages" + path
	if path == "/" {
		loaderPath = "/__gux_api/pages/index"
	}

	promise := js.Global().Call("fetch", loaderPath)
	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
		resp := args[0]
		if resp.Get("ok").Bool() {
			resp.Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
				// Parse JSON response properly to preserve all types (arrays, objects, etc.)
				jsonStr := args[0].String()
				var state map[string]any
				if err := json.Unmarshal([]byte(jsonStr), &state); err == nil {
					callback(state)
				} else {
					callback(nil)
				}
				return nil
			}))
		} else {
			callback(nil)
		}
		return nil
	}))
}

func main() {
	// Debug: use println which TinyGo reliably outputs to console
	println("[Gux WASM] main() started")

	document := js.Global().Get("document")
	window := js.Global()
	container := document.Call("getElementById", "app")

	var router *core.Router
	var render func()

	// Cached page component - persists across re-renders within same page
	// This ensures closures (like OnChange handlers) reference stable state objects
	var currentComponent func() core.Node
	var currentPath string

	// loadPage creates the page component for the current path
	// Called on navigation and initial load, NOT on state-triggered re-renders
	loadPage := func() {
		path := js.Global().Get("location").Get("pathname").String()
		// Only reload component if path changed (handles state-triggered re-renders)
		if currentComponent != nil && path == currentPath {
			return
		}
		currentPath = path
%s
		currentComponent = component
	}

	render = func() {
		// Save focus state before re-render
		activeElement := document.Get("activeElement")
		var focusName string
		if !activeElement.IsNull() && !activeElement.IsUndefined() {
			focusName = activeElement.Call("getAttribute", "name").String()
		}

		container.Set("innerHTML", "")

		// Use cached component - closures remain stable across re-renders
		if currentComponent != nil {
			node := currentComponent()
			result := node.Render(core.DOM())
			if domVal := result.DOMValue(); domVal != nil {
				container.Call("appendChild", domVal.(js.Value))
			}
		}

		// Intercept link clicks for client-side navigation
		links := document.Call("querySelectorAll", "a[href]")
		for i := 0; i < links.Get("length").Int(); i++ {
			link := links.Call("item", i)
			href := link.Get("href").String()
			origin := window.Get("location").Get("origin").String()

			// Skip external links (marked with data-gux-external)
			if link.Call("getAttribute", "data-gux-external").String() == "true" {
				continue
			}

			// Only intercept internal links
			if len(href) >= len(origin) && href[:len(origin)] == origin {
				link.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
					args[0].Call("preventDefault")
					path := this.Get("pathname").String()
					router.Navigate(path)
					return nil
				}))
			}
		}

		// Restore focus after re-render
		if focusName != "" {
			newElement := document.Call("querySelector", "[name=\""+focusName+"\"]")
			if !newElement.IsNull() && !newElement.IsUndefined() {
				newElement.Call("focus")
			}
		}
	}

	// Navigate fetches page data then renders
	// For cross-bundle navigation, do a full page redirect
	navigate := func(path string) {
		// Check if this path belongs to this bundle
		if !isRouteInBundle(path) {
			// Cross-bundle navigation - do full page redirect
			window.Get("location").Set("href", path)
			return
		}
		// Same-bundle navigation - clear component to force reload
		currentComponent = nil
		// Clear page-specific state before navigating to avoid stale data
		router.ClearState()
		window.Get("history").Call("pushState", nil, "", path)
		fetchLoader(path, func(state map[string]any) {
			if state != nil {
				router.Hydrate(state)
			}
			loadPage()
			render()
		})
	}

	router = core.NewRouter(render)
	router.SetNavigate(navigate)
	println("[Gux WASM] Calling SetDebugRouter")
	core.SetDebugRouter(router) // Enable state debugging
	println("[Gux WASM] SetDebugRouter completed")

	// Hydrate state from SSR
	stateEl := document.Call("getElementById", "__gux_state")
	if !stateEl.IsNull() && !stateEl.IsUndefined() {
		stateJSON := stateEl.Get("textContent").String()
		var state map[string]any
		if err := json.Unmarshal([]byte(stateJSON), &state); err == nil {
			router.Hydrate(state)
		}
	}

	// Handle browser back/forward
	window.Call("addEventListener", "popstate", js.FuncOf(func(this js.Value, args []js.Value) any {
		currentComponent = nil // Force reload for history navigation
		// Clear page-specific state for browser history navigation
		router.ClearState()
		// Fetch fresh data for the new/previous page
		path := window.Get("location").Get("pathname").String()
		fetchLoader(path, func(state map[string]any) {
			if state != nil {
				router.Hydrate(state)
			}
			loadPage()
			render()
		})
		return nil
	}))

	// Initial load
	loadPage()
	render()

	select {}
}
`, importSection.String(), bundleRoutesCode.String(), routeCode.String())

	return os.WriteFile(wasmDir+"/main.go", []byte(code), 0644)
}

// buildWasmBundle builds a specific WASM bundle
func buildWasmBundle(bundleName string, tinygo bool) error {
	wasmDir := "guxgen/wasm_" + bundleName
	outputFile := "guxgen/dist/" + bundleName + ".wasm"

	if bundleName == "app" {
		wasmDir = "guxgen/wasm"
		outputFile = "guxgen/dist/app.wasm"
	}

	if err := os.MkdirAll("guxgen/dist", 0755); err != nil {
		return err
	}

	fmt.Printf("Building WASM bundle: %s...\n", bundleName)

	var cmd *exec.Cmd
	if tinygo {
		cmd = exec.Command("tinygo", "build", "-o", outputFile, "-target", "wasm", "./"+wasmDir)
	} else {
		cmd = exec.Command("go", "build", "-o", outputFile, "./"+wasmDir)
		cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("WASM build failed for bundle %s: %w", bundleName, err)
	}

	// Get size for display
	info, err := os.Stat(outputFile)
	if err != nil {
		return err
	}
	sizeMB := float64(info.Size()) / 1024 / 1024
	compiler := "TinyGo"
	if !tinygo {
		compiler = "Go"
	}
	fmt.Printf("Built %s (%.2f MB) with %s\n", outputFile, sizeMB, compiler)

	return nil
}

// generateServerDTOCode generates server-side API code for models with parsed DTO info
func generateServerDTOCode(m CRUDModel) string {
	var sb strings.Builder

	// Determine which DTO to use for list and detail
	listDTO := m.ListDTO
	detailDTO := m.DetailDTO
	if detailDTO == "" {
		detailDTO = listDTO
	}

	// Get the model name from the DTO info
	modelName := m.Name
	if m.ListDTOInfo != nil && m.ListDTOInfo.ModelName != "" {
		modelName = m.ListDTOInfo.ModelName
	}

	// Build preload chain for list
	listPreloads := ""
	if m.ListDTOInfo != nil {
		for _, p := range m.ListDTOInfo.Preloads {
			listPreloads += fmt.Sprintf(".Preload(\"%s\")", p)
		}
	}

	// Build preload chain for detail
	detailPreloads := ""
	if m.DetailDTOInfo != nil {
		for _, p := range m.DetailDTOInfo.Preloads {
			detailPreloads += fmt.Sprintf(".Preload(\"%s\")", p)
		}
	}

	// Generate field mapping for list DTO
	listFieldMapping := generateFieldMapping(m.ListDTOInfo, "item")
	listNestedMapping := generateNestedDTOMapping(m.ListDTOInfo, "result[i]", "item", true)

	// Generate field mapping for detail DTO
	detailFieldMapping := generateFieldMapping(m.DetailDTOInfo, "item")
	detailNestedMapping := generateNestedDTOMapping(m.DetailDTOInfo, "result", "item", false)

	// Check if detail DTO has preloads (relationships) - if so, try FromModel pattern
	hasRelationships := m.DetailDTOInfo != nil && len(m.DetailDTOInfo.Preloads) > 0

	// Generate the Get method body based on whether we have relationships
	var getMethodBody string
	if hasRelationships {
		// Use FromModel pattern for DTOs with relationships (if implemented)
		// Uses type assertion to check if FromModel exists at runtime
		getMethodBody = fmt.Sprintf(`// Get returns a single %s by ID.
func (a *%sAPI) Get(id uint, callback func(*dto.%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	var item models.%s
	if err := db%s.First(&item, id).Error; err != nil {
		callback(nil, err)
		return
	}
	// Try DTOMapper.FromModel if implemented (for complex relationships)
	empty := &dto.%s{}
	if mapper, ok := interface{}(empty).(interface{ FromModel(interface{}) interface{} }); ok {
		mapped := mapper.FromModel(item)
		if result, ok := mapped.(dto.%s); ok {
			callback(&result, nil)
			return
		}
	}
	// Fallback to basic mapping
	result := &dto.%s{
%s	}
%s	callback(result, nil)
}`,
			m.Name, m.PluralName, detailDTO, modelName, detailPreloads,
			detailDTO, detailDTO, detailDTO, detailFieldMapping, detailNestedMapping)
	} else {
		// Use simple field mapping for DTOs without relationships
		getMethodBody = fmt.Sprintf(`// Get returns a single %s by ID.
func (a *%sAPI) Get(id uint, callback func(*dto.%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	var item models.%s
	if err := db%s.First(&item, id).Error; err != nil {
		callback(nil, err)
		return
	}
	result := &dto.%s{
%s	}
%s	callback(result, nil)
}`,
			m.Name, m.PluralName, detailDTO, modelName, detailPreloads, detailDTO, detailFieldMapping, detailNestedMapping)
	}

	sb.WriteString(fmt.Sprintf(`
// %sAPI provides CRUD operations for %s.
type %sAPI struct{}

// %s is the API client for %s operations.
var %s = &%sAPI{}

// List returns all %s records.
func (a *%sAPI) List(callback func([]dto.%s, error)) {
	if db == nil {
		callback(nil, nil)
		return
	}
	var items []models.%s
	if err := db%s.Find(&items).Error; err != nil {
		callback(nil, err)
		return
	}
	// Convert to DTOs using field mappings from gux tags
	result := make([]dto.%s, len(items))
	for i, item := range items {
		result[i] = dto.%s{
%s		}
%s	}
	callback(result, nil)
}

%s

// Create creates a new %s (server-side stub - actual creation handled by CRUD endpoint).
func (a *%sAPI) Create(data map[string]interface{}, callback func(*dto.%s, error)) {
	// Server-side: this is typically not called directly
	// The CRUD endpoint handles creation with hooks
	callback(nil, nil)
}

// Update updates an existing %s (server-side stub).
func (a *%sAPI) Update(id uint, data map[string]interface{}, callback func(*dto.%s, error)) {
	// Server-side: this is typically not called directly
	callback(nil, nil)
}

// Delete deletes a %s by ID (server-side stub).
func (a *%sAPI) Delete(id uint, callback func(error)) {
	// Server-side: this is typically not called directly
	callback(nil)
}
`,
		m.PluralName, m.Name,                    // type comment
		m.PluralName,                            // type name
		m.PluralName, m.Name,                    // var comment
		m.PluralName, m.PluralName,              // var declaration
		m.Name,                                  // List comment
		m.PluralName, listDTO,                   // List signature
		modelName,                               // List query model
		listPreloads,                            // List preloads
		listDTO,                                 // List result make
		listDTO,                                 // List DTO type
		listFieldMapping,                        // List field mapping
		listNestedMapping,                       // List nested DTO mapping
		getMethodBody,                           // Get method (generated above)
		m.Name,                                  // Create comment
		m.PluralName, detailDTO,                 // Create signature
		m.Name,                                  // Update comment
		m.PluralName, detailDTO,                 // Update signature
		m.Name,                                  // Delete comment
		m.PluralName,                            // Delete signature
	))

	return sb.String()
}

// generateFieldMapping generates the field assignment code from parsed DTO info
func generateFieldMapping(info *DTOInfo, varName string) string {
	if info == nil {
		return ""
	}

	var sb strings.Builder
	for _, f := range info.Fields {
		if f.IsSlice {
			// For slice fields (relationships), we need to map each element
			sb.WriteString(fmt.Sprintf("\t\t\t// %s mapped from %s.%s (slice)\n",
				f.DTOField, info.ModelName, f.ModelField))
			// Skip slice mapping in simple generation - requires nested loop
			continue
		}
		if f.IsNestedDTO {
			// Skip nested DTOs here - they are handled by generateNestedDTOMapping
			continue
		}
		sb.WriteString(fmt.Sprintf("\t\t\t%s: %s.%s,\n", f.DTOField, varName, f.ModelField))
	}
	return sb.String()
}

// generateNestedDTOMapping generates code to map nested DTO fields after the main struct literal
// Returns empty string if no nested DTOs, otherwise returns the mapping code
func generateNestedDTOMapping(info *DTOInfo, resultVar, itemVar string, forList bool) string {
	if info == nil {
		return ""
	}

	// Get model field types to determine if FK fields are pointers
	modelFieldTypes, _ := getModelFieldTypes(info.ModelName)

	var sb strings.Builder
	for _, f := range info.Fields {
		if !f.IsNestedDTO {
			continue
		}

		// Handle pointer types - strip * prefix for parsing, track for assignment
		dtoTypeName := strings.TrimPrefix(f.DTOType, "*")
		isPointer := strings.HasPrefix(f.DTOType, "*")

		// Determine the relation field name in the model
		// Use Preload if set, otherwise derive from ModelField by removing "ID" suffix
		relationField := f.Preload
		if relationField == "" {
			relationField = strings.TrimSuffix(f.ModelField, "ID")
			if relationField == "" {
				relationField = f.ModelField
			}
		}

		// Build the foreign key field name - this is what we check for nil/zero
		fkField := f.ModelField
		if !strings.HasSuffix(fkField, "ID") {
			fkField = f.ModelField + "ID"
		}

		// Check if the FK field is a pointer type in the model
		fkIsPointer := false
		if modelFieldTypes != nil {
			if fkType, ok := modelFieldTypes[fkField]; ok {
				fkIsPointer = strings.HasPrefix(fkType, "*")
			}
		}

		// Generate the appropriate nil/zero check based on FK field type
		var fkCheck string
		if fkIsPointer {
			fkCheck = fmt.Sprintf("%s.%s != nil", itemVar, fkField)
		} else {
			fkCheck = fmt.Sprintf("%s.%s != 0", itemVar, fkField)
		}

		// Parse nested DTO type to get field mappings (e.g., UserBrief -> ID, Name)
		nestedInfo, err := parseDTOFile("guxgen/dto", dtoTypeName)
		if err != nil {
			// Fallback to dto/ for backwards compatibility
			nestedInfo, err = parseDTOFile("dto", dtoTypeName)
		}
		if err != nil {
			// Can't parse nested DTO - generate simple direct mapping using relation field
			// This handles cases where the nested DTO isn't in the dto/ directory
			dtoAssign := fmt.Sprintf("dto.%s{ID: %s.%s.ID}", dtoTypeName, itemVar, relationField)
			if isPointer {
				dtoAssign = fmt.Sprintf("&dto.%s{ID: %s.%s.ID}", dtoTypeName, itemVar, relationField)
			}
			if forList {
				sb.WriteString(fmt.Sprintf("\t\tif %s {\n", fkCheck))
				sb.WriteString(fmt.Sprintf("\t\t\t%s.%s = %s\n", resultVar, f.DTOField, dtoAssign))
				sb.WriteString("\t\t}\n")
			} else {
				sb.WriteString(fmt.Sprintf("\tif %s {\n", fkCheck))
				sb.WriteString(fmt.Sprintf("\t\t%s.%s = %s\n", resultVar, f.DTOField, dtoAssign))
				sb.WriteString("\t}\n")
			}
			continue
		}

		// Generate check for loaded relationship
		dtoPrefix := "dto."
		if isPointer {
			dtoPrefix = "&dto."
		}
		if forList {
			// For list: result[i].Salesperson = dto.UserBrief{...}
			sb.WriteString(fmt.Sprintf("\t\tif %s {\n", fkCheck))
			sb.WriteString(fmt.Sprintf("\t\t\t%s.%s = %s%s{\n", resultVar, f.DTOField, dtoPrefix, dtoTypeName))
		} else {
			// For single: result.Salesperson = dto.UserBrief{...}
			sb.WriteString(fmt.Sprintf("\tif %s {\n", fkCheck))
			sb.WriteString(fmt.Sprintf("\t\t%s.%s = %s%s{\n", resultVar, f.DTOField, dtoPrefix, dtoTypeName))
		}

		// Add nested field mappings - use relationField to access the relation object
		for _, nf := range nestedInfo.Fields {
			if nf.IsSlice || nf.IsNestedDTO {
				continue // Skip complex nested fields for now
			}
			if forList {
				sb.WriteString(fmt.Sprintf("\t\t\t\t%s: %s.%s.%s,\n", nf.DTOField, itemVar, relationField, nf.ModelField))
			} else {
				sb.WriteString(fmt.Sprintf("\t\t\t%s: %s.%s.%s,\n", nf.DTOField, itemVar, relationField, nf.ModelField))
			}
		}

		if forList {
			sb.WriteString("\t\t\t}\n")
			sb.WriteString("\t\t}\n")
		} else {
			sb.WriteString("\t\t}\n")
			sb.WriteString("\t}\n")
		}
	}
	return sb.String()
}

// generateAPIClient generates guxgen/api/client.go with WASM-compatible DTOs
func generateAPIClient(modelsImport string, dtoImport string, models []CRUDModel) error {
	if len(models) == 0 {
		return nil // No CRUD models, skip API client generation
	}

	if err := os.MkdirAll("guxgen/api", 0755); err != nil {
		return err
	}

	// Generate DTO structs for models WITHOUT custom DTOs
	var dtoCode strings.Builder
	for _, m := range models {
		// Skip models with custom DTOs
		if m.ListDTO != "" || m.DetailDTO != "" {
			continue
		}
		// Parse actual model fields
		fields, err := getModelFields(m.Name)
		if err != nil || len(fields) == 0 {
			// Fallback to minimal DTO with just ID
			dtoCode.WriteString(fmt.Sprintf(`
// %s is a WASM-compatible DTO for %s.
type %s struct {
	ID          uint   `+"`json:\"id\"`"+`
	CreatedAt   string `+"`json:\"created_at,omitempty\"`"+`
	UpdatedAt   string `+"`json:\"updated_at,omitempty\"`"+`
}
`, m.Name, m.Name, m.Name))
		} else {
			// Generate DTO with actual fields
			dtoCode.WriteString(generateDTOStructCode(m.Name, fields))
		}
	}

	// Generate model-specific API code
	var modelCode strings.Builder
	for _, m := range models {
		// Models with DTOs get specialized read-only API
		if m.ListDTO != "" || m.DetailDTO != "" {
			listType := m.ListDTO
			if listType == "" {
				listType = m.Name
			}
			detailType := m.DetailDTO
			if detailType == "" {
				detailType = m.Name
			}
			dtoPkg := m.DTOPackage
			if dtoPkg == "" {
				dtoPkg = "dto"
			}

			modelCode.WriteString(fmt.Sprintf(`
// %sAPI provides read operations for %s with DTO responses.
type %sAPI struct {
	baseURL string
}

// %s is the API client for %s operations.
var %s = &%sAPI{baseURL: "/__gux_api/crud/%s"}

// List returns all %s records using %s DTO.
func (a *%sAPI) List(callback func([]%s.%s, error)) {
	go func() {
		resp, err := fetch("GET", a.baseURL, nil)
		if err != nil {
			callback(nil, err)
			return
		}
		var items []%s.%s
		if err := json.Unmarshal(resp, &items); err != nil {
			callback(nil, err)
			return
		}
		callback(items, nil)
	}()
}

// Get returns a single %s by ID using %s DTO.
func (a *%sAPI) Get(id uint, callback func(*%s.%s, error)) {
	go func() {
		resp, err := fetch("GET", fmt.Sprintf("%%s/%%d", a.baseURL, id), nil)
		if err != nil {
			callback(nil, err)
			return
		}
		var item %s.%s
		if err := json.Unmarshal(resp, &item); err != nil {
			callback(nil, err)
			return
		}
		callback(&item, nil)
	}()
}

// Create creates a new %s with the given data.
func (a *%sAPI) Create(data map[string]interface{}, callback func(*%s.%s, error)) {
	go func() {
		body, _ := json.Marshal(data)
		resp, err := fetch("POST", a.baseURL, body)
		if err != nil {
			callback(nil, err)
			return
		}
		var item %s.%s
		if err := json.Unmarshal(resp, &item); err != nil {
			callback(nil, err)
			return
		}
		callback(&item, nil)
	}()
}

// Update updates an existing %s.
func (a *%sAPI) Update(id uint, data map[string]interface{}, callback func(*%s.%s, error)) {
	go func() {
		body, _ := json.Marshal(data)
		resp, err := fetch("PUT", fmt.Sprintf("%%s/%%d", a.baseURL, id), body)
		if err != nil {
			callback(nil, err)
			return
		}
		var item %s.%s
		if err := json.Unmarshal(resp, &item); err != nil {
			callback(nil, err)
			return
		}
		callback(&item, nil)
	}()
}

// Delete deletes a %s by ID.
func (a *%sAPI) Delete(id uint, callback func(error)) {
	go func() {
		_, err := fetch("DELETE", fmt.Sprintf("%%s/%%d", a.baseURL, id), nil)
		callback(err)
	}()
}
`,
				m.PluralName, m.Name,               // type comment
				m.PluralName,                       // type name
				m.PluralName, m.Name,               // var comment
				m.PluralName, m.PluralName, m.Path, // var declaration
				m.Name, listType,                   // List comment
				m.PluralName, dtoPkg, listType,     // List signature
				dtoPkg, listType,                   // List body
				m.Name, detailType,                 // Get comment
				m.PluralName, dtoPkg, detailType,   // Get signature
				dtoPkg, detailType,                 // Get body
				m.Name,                             // Create comment
				m.PluralName, dtoPkg, detailType,   // Create signature
				dtoPkg, detailType,                 // Create body
				m.Name,                             // Update comment
				m.PluralName, dtoPkg, detailType,   // Update signature
				dtoPkg, detailType,                 // Update body
				m.Name,                             // Delete comment
				m.PluralName,                       // Delete signature
			))
			continue
		}

		// Models without DTOs get full CRUD with inline DTO
		modelCode.WriteString(fmt.Sprintf(`
// %sAPI provides CRUD operations for %s.
type %sAPI struct {
	baseURL string
}

// %s is the API client for %s operations.
var %s = &%sAPI{baseURL: "/__gux_api/crud/%s"}

// List returns all %s records.
func (a *%sAPI) List(callback func([]%s, error)) {
	go func() {
		resp, err := fetch("GET", a.baseURL, nil)
		if err != nil {
			callback(nil, err)
			return
		}
		var items []%s
		if err := json.Unmarshal(resp, &items); err != nil {
			callback(nil, err)
			return
		}
		callback(items, nil)
	}()
}

// Get returns a single %s by ID.
func (a *%sAPI) Get(id uint, callback func(*%s, error)) {
	go func() {
		resp, err := fetch("GET", fmt.Sprintf("%%s/%%d", a.baseURL, id), nil)
		if err != nil {
			callback(nil, err)
			return
		}
		var item %s
		if err := json.Unmarshal(resp, &item); err != nil {
			callback(nil, err)
			return
		}
		callback(&item, nil)
	}()
}

// Create creates a new %s.
func (a *%sAPI) Create(item *%s, callback func(*%s, error)) {
	go func() {
		body, _ := json.Marshal(item)
		resp, err := fetch("POST", a.baseURL, body)
		if err != nil {
			callback(nil, err)
			return
		}
		var result %s
		if err := json.Unmarshal(resp, &result); err != nil {
			callback(nil, err)
			return
		}
		callback(&result, nil)
	}()
}

// Update updates an existing %s.
func (a *%sAPI) Update(item *%s, callback func(*%s, error)) {
	go func() {
		body, _ := json.Marshal(item)
		resp, err := fetch("PUT", fmt.Sprintf("%%s/%%d", a.baseURL, item.ID), body)
		if err != nil {
			callback(nil, err)
			return
		}
		var result %s
		if err := json.Unmarshal(resp, &result); err != nil {
			callback(nil, err)
			return
		}
		callback(&result, nil)
	}()
}

// Delete deletes a %s by ID.
func (a *%sAPI) Delete(id uint, callback func(error)) {
	go func() {
		_, err := fetch("DELETE", fmt.Sprintf("%%s/%%d", a.baseURL, id), nil)
		callback(err)
	}()
}
`,
			m.PluralName, m.Name,               // type comment
			m.PluralName,                       // type name
			m.PluralName, m.Name,               // var comment
			m.PluralName, m.PluralName, m.Path, // var declaration
			m.Name,                             // List comment
			m.PluralName, m.Name,               // List signature
			m.Name,                             // List body
			m.Name,                             // Get comment
			m.PluralName, m.Name,               // Get signature
			m.Name,                             // Get body
			m.Name,                             // Create comment
			m.PluralName, m.Name, m.Name,       // Create signature
			m.Name,                             // Create body
			m.Name,                             // Update comment
			m.PluralName, m.Name, m.Name,       // Update signature
			m.Name,                             // Update body
			m.Name,                             // Delete comment
			m.PluralName,                       // Delete signature
		))
	}

	// Note: modelsImport is unused since we generate our own DTOs
	_ = modelsImport

	// Build imports list
	imports := `"encoding/json"
	"errors"
	"fmt"
	"syscall/js"`

	// Add dto import if any model uses DTOs
	if dtoImport != "" {
		for _, m := range models {
			if m.ListDTO != "" || m.DetailDTO != "" {
				imports += fmt.Sprintf("\n\n\t\"%s\"", dtoImport)
				break
			}
		}
	}

	code := fmt.Sprintf(`//go:build js && wasm

// Code generated by gux; DO NOT EDIT.
package api

import (
	%s
)

// Init is a no-op in WASM (database not available).
func Init(db interface{}) {}

// getCSRFToken reads the CSRF token from the meta tag.
func getCSRFToken() string {
	doc := js.Global().Get("document")
	meta := doc.Call("querySelector", "meta[name=\"csrf-token\"]")
	if meta.IsNull() || meta.IsUndefined() {
		return ""
	}
	return meta.Get("content").String()
}

// fetch makes an HTTP request and returns the response body.
func fetch(method, url string, body []byte) ([]byte, error) {
	done := make(chan struct{})
	var result []byte
	var fetchErr error

	// Create JS object for headers
	jsOpts := js.Global().Get("Object").New()
	jsOpts.Set("method", method)
	headers := js.Global().Get("Object").New()
	headers.Set("Content-Type", "application/json")

	// Add CSRF token for mutating requests
	if method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE" {
		if token := getCSRFToken(); token != "" {
			headers.Set("X-CSRF-Token", token)
		}
	}

	jsOpts.Set("headers", headers)
	if body != nil {
		jsOpts.Set("body", string(body))
	}

	// Call fetch
	promise := js.Global().Call("fetch", url, jsOpts)

	// Handle response
	var thenFunc, catchFunc js.Func
	thenFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resp := args[0]
		if !resp.Get("ok").Bool() {
			fetchErr = errors.New("request failed: " + resp.Get("statusText").String())
			close(done)
			return nil
		}
		// Get response text
		resp.Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			result = []byte(args[0].String())
			close(done)
			return nil
		}))
		return nil
	})
	catchFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fetchErr = errors.New("fetch error: " + args[0].Get("message").String())
		close(done)
		return nil
	})

	promise.Call("then", thenFunc).Call("catch", catchFunc)

	<-done
	return result, fetchErr
}

// Post makes a POST request to the given URL with JSON body.
// Use this for custom API endpoints like authentication.
func Post[T any](url string, data any, callback func(T, error)) {
	go func() {
		body, err := json.Marshal(data)
		if err != nil {
			var zero T
			callback(zero, err)
			return
		}
		resp, err := fetch("POST", url, body)
		if err != nil {
			var zero T
			callback(zero, err)
			return
		}
		var result T
		if err := json.Unmarshal(resp, &result); err != nil {
			var zero T
			callback(zero, err)
			return
		}
		callback(result, nil)
	}()
}
%s%s
`, imports, dtoCode.String(), modelCode.String())

	if err := os.WriteFile("guxgen/api/client.go", []byte(code), 0644); err != nil {
		return err
	}

	// Generate server-side implementation (real DB queries)
	var stubDtoCode strings.Builder
	for _, m := range models {
		// Skip models with DTOs - they import from the dto package
		if m.ListDTO != "" || m.DetailDTO != "" {
			continue
		}
		// Parse actual model fields instead of hardcoding Name/Description
		fields, err := getModelFields(m.Name)
		if err != nil {
			// If we can't parse the model, generate empty DTO
			stubDtoCode.WriteString(fmt.Sprintf(`
// %s is the API DTO for %s.
type %s struct {
	ID          uint   `+"`json:\"id\"`"+`
	CreatedAt   string `+"`json:\"created_at,omitempty\"`"+`
	UpdatedAt   string `+"`json:\"updated_at,omitempty\"`"+`
}
`, m.Name, m.Name, m.Name))
			continue
		}
		stubDtoCode.WriteString(generateDTOStructCode(m.Name, fields))
	}

	var stubCode strings.Builder
	for _, m := range models {
		// Models WITH parsed DTO info get field-mapped code
		if m.ListDTOInfo != nil || m.DetailDTOInfo != nil {
			stubCode.WriteString(generateServerDTOCode(m))
			continue
		}

		// Models WITHOUT DTOs get dynamically generated code based on actual model fields
		if m.ListDTO != "" || m.DetailDTO != "" {
			continue // Skip - no parsed info available
		}
		stubCode.WriteString(generateServerAPICode(m.Name, m.PluralName))
	}

	// Build server-side imports
	serverImports := fmt.Sprintf(`"gorm.io/gorm"

	"%s"`, modelsImport)

	// Add dto import if any models use DTOs with parsed info
	for _, m := range models {
		if m.ListDTOInfo != nil || m.DetailDTOInfo != nil {
			serverImports += fmt.Sprintf(`

	"%s"`, dtoImport)
			break
		}
	}

	stubFile := fmt.Sprintf(`//go:build !js || !wasm

// Code generated by gux; DO NOT EDIT.
// Server-side implementation - queries database directly.
package api

import (
	%s
)

var db *gorm.DB

// Init initializes the API with a database connection.
// Called automatically by the gux framework.
func Init(database *gorm.DB) {
	db = database
}

// Post is a stub for server-side builds.
// The actual POST request happens client-side via WASM.
func Post[T any](url string, data any, callback func(T, error)) {
	// Server-side: this function is only called from WASM
	// It's included here for compilation compatibility
	var zero T
	callback(zero, nil)
}
%s%s
`, serverImports, stubDtoCode.String(), stubCode.String())

	return os.WriteFile("guxgen/api/client_server.go", []byte(stubFile), 0644)
}

// generateEndpointClient generates guxgen/api/endpoints_gen.go for typed API endpoints
func generateEndpointClient(endpoints []APIEndpointInfo, dtoImport string) error {
	if len(endpoints) == 0 {
		return nil
	}

	if err := os.MkdirAll("guxgen/api", 0755); err != nil {
		return err
	}

	// Determine if we need fmt import (for path parameters)
	needsFmt := false
	for _, ep := range endpoints {
		if len(ep.PathParams) > 0 {
			needsFmt = true
			break
		}
	}

	// Build imports
	imports := `"encoding/json"
	"errors"

	guxfetch "github.com/dougbarrett/gux/fetch"`

	if needsFmt {
		imports = `"encoding/json"
	"errors"
	"fmt"

	guxfetch "github.com/dougbarrett/gux/fetch"`
	}

	// Check if we need dto import
	hasDTOTypes := false
	for _, ep := range endpoints {
		if ep.Package != "" && ep.Package != "main" {
			hasDTOTypes = true
			break
		}
	}

	if hasDTOTypes && dtoImport != "" {
		imports += fmt.Sprintf("\n\n\t\"%s\"", dtoImport)
	}

	// Generate endpoint functions
	var funcCode strings.Builder
	for _, ep := range endpoints {
		funcCode.WriteString(generateEndpointFunc(ep))
	}

	code := fmt.Sprintf(`//go:build js && wasm

// Code generated by gux; DO NOT EDIT.
// API endpoint client for typed API endpoints.
package api

import (
	%s
)

// apiEndpointFetch makes an HTTP request and returns the response body
func apiEndpointFetch(method, url string, body []byte) ([]byte, error) {
	opts := &guxfetch.Options{
		Method: method,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	if body != nil {
		opts.Body = string(body)
	}

	resp, err := guxfetch.Fetch(url, opts)
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.StatusText)
	}
	return []byte(resp.Body), nil
}
%s
`, imports, funcCode.String())

	if err := os.WriteFile("guxgen/api/endpoints_gen.go", []byte(code), 0644); err != nil {
		return err
	}

	// Generate server-side implementation that makes HTTP requests
	var serverFuncs strings.Builder
	needsDTOImport := false
	for _, ep := range endpoints {
		serverFuncs.WriteString(generateEndpointServerFunc(ep))
		// Track if we need dto import
		if ep.RequestType != "" || ep.ResponseType != "" {
			if ep.Package != "" && ep.Package != "main" {
				needsDTOImport = true
			}
		}
	}

	// Build server imports
	// fmt is always needed for endpointFetch error handling
	serverImports := `"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"`

	if needsDTOImport && dtoImport != "" {
		serverImports += fmt.Sprintf("\n\n\t\"%s\"", dtoImport)
	}

	serverCode := fmt.Sprintf(`//go:build !js || !wasm

// Code generated by gux; DO NOT EDIT.
// Server-side implementation for typed API endpoints.
// Makes HTTP requests to the API during SSR.
package api

import (
	%s
)

// endpointBaseURL is the base URL for API calls during SSR.
// Set via InitEndpoints() or defaults to http://localhost:PORT.
var endpointBaseURL = ""

// endpointSessionID stores the current session ID for SSR requests.
// This allows API calls during SSR to inherit the user's authentication.
var endpointSessionID = ""

// endpointSessionCookieName is the name of the session cookie.
var endpointSessionCookieName = "__gux_session"

// InitEndpoints sets the base URL for server-side API calls.
// Call this in your main() before starting the server.
// If not called, defaults to http://localhost:$PORT or http://localhost:8080.
func InitEndpoints(baseURL string) {
	endpointBaseURL = baseURL
}

// SetEndpointSession sets the session ID and cookie name for SSR API calls.
// This should be called before rendering pages that make API calls.
// The session is automatically propagated to API requests made during SSR.
func SetEndpointSession(sessionID, cookieName string) {
	endpointSessionID = sessionID
	if cookieName != "" {
		endpointSessionCookieName = cookieName
	}
}

// ClearEndpointSession clears the session after page rendering is complete.
func ClearEndpointSession() {
	endpointSessionID = ""
}

// getEndpointBaseURL returns the base URL for API calls.
func getEndpointBaseURL() string {
	if endpointBaseURL != "" {
		return endpointBaseURL
	}
	// Try to get from PORT environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return "http://localhost:" + port
}

// endpointFetch makes an HTTP request to the API endpoint.
// If a session ID is set, it will be passed as a cookie for authentication.
func endpointFetch(method, path string, body []byte) ([]byte, error) {
	url := getEndpointBaseURL() + path

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Pass session cookie if available (for SSR auth propagation)
	if endpointSessionID != "" {
		req.AddCookie(&http.Cookie{
			Name:  endpointSessionCookieName,
			Value: endpointSessionID,
		})
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %%s", string(respBody))
	}

	return respBody, nil
}
%s`, serverImports, serverFuncs.String())

	return os.WriteFile("guxgen/api/endpoints_server_gen.go", []byte(serverCode), 0644)
}

// generateEndpointFunc generates a single endpoint client function
func generateEndpointFunc(ep APIEndpointInfo) string {
	var sb strings.Builder

	// Build function parameters
	var params []string
	for _, p := range ep.PathParams {
		params = append(params, p+" uint")
	}

	// Add request body parameter if needed
	if ep.RequestType != "" {
		reqType := ep.RequestType
		if ep.Package != "" && ep.Package != "main" {
			reqType = ep.Package + "." + ep.RequestType
		}
		params = append(params, "req "+reqType)
	}

	// Build URL with path parameters
	urlExpr := `"` + ep.Path + `"`
	if len(ep.PathParams) > 0 {
		// Replace :param with %d for formatting
		urlPattern := ep.Path
		for _, p := range ep.PathParams {
			urlPattern = strings.Replace(urlPattern, ":"+p, "%d", 1)
		}
		urlExpr = `fmt.Sprintf("` + urlPattern + `", ` + strings.Join(ep.PathParams, ", ") + `)`
	}

	// Determine response handling
	if ep.ResponseType == "" {
		// DELETE - no response body
		paramStr := strings.Join(params, ", ")
		sb.WriteString(fmt.Sprintf(`
// %s calls %s %s
func %s(%s, callback func(error)) {
	go func() {
		_, err := apiEndpointFetch("%s", %s, nil)
		callback(err)
	}()
}
`, ep.FuncName, ep.Method, ep.Path, ep.FuncName, paramStr, ep.Method, urlExpr))
	} else {
		// Has response body
		respType := ep.ResponseType
		if ep.Package != "" && ep.Package != "main" {
			respType = ep.Package + "." + ep.ResponseType
		}

		// Add callback parameter
		params = append(params, fmt.Sprintf("callback func(%s, error)", respType))
		paramStr := strings.Join(params, ", ")

		if ep.RequestType != "" {
			// POST/PUT/PATCH with request body
			sb.WriteString(fmt.Sprintf(`
// %s calls %s %s
func %s(%s) {
	go func() {
		body, err := json.Marshal(req)
		if err != nil {
			var zero %s
			callback(zero, err)
			return
		}
		resp, err := apiEndpointFetch("%s", %s, body)
		if err != nil {
			var zero %s
			callback(zero, err)
			return
		}
		var result %s
		if err := json.Unmarshal(resp, &result); err != nil {
			var zero %s
			callback(zero, err)
			return
		}
		callback(result, nil)
	}()
}
`, ep.FuncName, ep.Method, ep.Path, ep.FuncName, paramStr, respType, ep.Method, urlExpr, respType, respType, respType))
		} else {
			// Request without explicit body type
			// POST/PUT/PATCH need an empty JSON object, GET needs nil
			bodyArg := "nil"
			if ep.Method == "POST" || ep.Method == "PUT" || ep.Method == "PATCH" {
				bodyArg = `[]byte("{}")`
			}
			sb.WriteString(fmt.Sprintf(`
// %s calls %s %s
func %s(%s) {
	go func() {
		resp, err := apiEndpointFetch("%s", %s, %s)
		if err != nil {
			var zero %s
			callback(zero, err)
			return
		}
		var result %s
		if err := json.Unmarshal(resp, &result); err != nil {
			var zero %s
			callback(zero, err)
			return
		}
		callback(result, nil)
	}()
}
`, ep.FuncName, ep.Method, ep.Path, ep.FuncName, paramStr, ep.Method, urlExpr, bodyArg, respType, respType, respType))
		}
	}

	return sb.String()
}

// generateEndpointServerFunc generates a server-side implementation for an endpoint
// that makes HTTP requests to the API endpoint during SSR.
func generateEndpointServerFunc(ep APIEndpointInfo) string {
	var sb strings.Builder

	// Build function parameters
	var params []string
	for _, p := range ep.PathParams {
		params = append(params, p+" uint")
	}

	// Add request body parameter if needed
	if ep.RequestType != "" {
		reqType := ep.RequestType
		if ep.Package != "" && ep.Package != "main" {
			reqType = ep.Package + "." + ep.RequestType
		}
		params = append(params, "req "+reqType)
	}

	// Build URL with path parameters
	urlExpr := `"` + ep.Path + `"`
	if len(ep.PathParams) > 0 {
		// Replace :param with %d for formatting
		urlPattern := ep.Path
		for _, p := range ep.PathParams {
			urlPattern = strings.Replace(urlPattern, ":"+p, "%d", 1)
		}
		urlExpr = `fmt.Sprintf("` + urlPattern + `", ` + strings.Join(ep.PathParams, ", ") + `)`
	}

	// Determine response handling
	if ep.ResponseType == "" {
		// DELETE - no response body
		paramStr := strings.Join(params, ", ")
		sb.WriteString(fmt.Sprintf(`
// %s calls %s %s
func %s(%s, callback func(error)) {
	_, err := endpointFetch("%s", %s, nil)
	callback(err)
}
`, ep.FuncName, ep.Method, ep.Path, ep.FuncName, paramStr, ep.Method, urlExpr))
	} else {
		// Has response body
		respType := ep.ResponseType
		if ep.Package != "" && ep.Package != "main" {
			respType = ep.Package + "." + ep.ResponseType
		}

		// Add callback parameter
		params = append(params, fmt.Sprintf("callback func(%s, error)", respType))
		paramStr := strings.Join(params, ", ")

		if ep.RequestType != "" {
			// POST/PUT/PATCH with request body
			sb.WriteString(fmt.Sprintf(`
// %s calls %s %s
func %s(%s) {
	body, err := json.Marshal(req)
	if err != nil {
		var zero %s
		callback(zero, err)
		return
	}
	resp, err := endpointFetch("%s", %s, body)
	if err != nil {
		var zero %s
		callback(zero, err)
		return
	}
	var result %s
	if err := json.Unmarshal(resp, &result); err != nil {
		var zero %s
		callback(zero, err)
		return
	}
	callback(result, nil)
}
`, ep.FuncName, ep.Method, ep.Path, ep.FuncName, paramStr, respType, ep.Method, urlExpr, respType, respType, respType))
		} else {
			// GET or POST/PUT without explicit request body
			bodyArg := "nil"
			if ep.Method == "POST" || ep.Method == "PUT" || ep.Method == "PATCH" {
				bodyArg = `[]byte("{}")`
			}
			sb.WriteString(fmt.Sprintf(`
// %s calls %s %s
func %s(%s) {
	resp, err := endpointFetch("%s", %s, %s)
	if err != nil {
		var zero %s
		callback(zero, err)
		return
	}
	var result %s
	if err := json.Unmarshal(resp, &result); err != nil {
		var zero %s
		callback(zero, err)
		return
	}
	callback(result, nil)
}
`, ep.FuncName, ep.Method, ep.Path, ep.FuncName, paramStr, ep.Method, urlExpr, bodyArg, respType, respType, respType))
		}
	}

	return sb.String()
}

// computeFileHash computes a short hash of a file's content for cache busting
func computeFileHash(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	hash := fmt.Sprintf("%x", md5.Sum(data))
	// Return first 8 characters of hash
	if len(hash) > 8 {
		return hash[:8]
	}
	return hash
}

// generateAssetsFile generates assets_gen.go with support for multiple bundles
func generateAssetsFile(modulePath string, bundles []string) error {
	var embedCode strings.Builder
	var initCode strings.Builder

	// Compute hashes for cache busting
	stylesHash := computeFileHash("guxgen/dist/styles.css")
	wasmHashes := make(map[string]string)
	wasmHashes["app"] = computeFileHash("guxgen/dist/app.wasm")
	for _, bundle := range bundles {
		if bundle != "app" {
			wasmHashes[bundle] = computeFileHash(fmt.Sprintf("guxgen/dist/%s.wasm", bundle))
		}
	}

	// Default bundle (app.wasm)
	embedCode.WriteString(`//go:embed guxgen/dist/app.wasm
var wasmBinary []byte

//go:embed guxgen/dist/wasm_exec.js
var wasmExecJS []byte

//go:embed guxgen/dist/styles.css
var stylesCSS []byte
`)

	// Additional bundles
	for _, bundle := range bundles {
		if bundle != "app" {
			embedCode.WriteString(fmt.Sprintf(`
//go:embed guxgen/dist/%s.wasm
var wasm%s []byte
`, bundle, strings.Title(bundle)))
		}
	}

	// Build hash map code
	var hashMapCode strings.Builder
	hashMapCode.WriteString("map[string]string{")
	for bundle, hash := range wasmHashes {
		hashMapCode.WriteString(fmt.Sprintf("%q: %q, ", bundle, hash))
	}
	hashMapCode.WriteString("}")

	// Init function
	initCode.WriteString("func init() {\n")
	initCode.WriteString("\tcore.SetDefaultAssets(wasmBinary, wasmExecJS, stylesCSS)\n")
	for _, bundle := range bundles {
		if bundle != "app" {
			initCode.WriteString(fmt.Sprintf("\tcore.SetDefaultBundle(%q, wasm%s)\n", bundle, strings.Title(bundle)))
		}
	}
	// Set asset hashes for cache busting
	initCode.WriteString(fmt.Sprintf("\tcore.SetDefaultAssetHashes(%q, %s)\n", stylesHash, hashMapCode.String()))
	initCode.WriteString("}\n")

	code := fmt.Sprintf(`// Code generated by gux; DO NOT EDIT.
package main

import (
	_ "embed"

	"github.com/dougbarrett/gux/core"
)

%s
%s`, embedCode.String(), initCode.String())

	return os.WriteFile("assets_gen.go", []byte(code), 0644)
}

// buildWasmNew builds the WASM module using TinyGo
func buildWasmNew(tinygo bool) error {
	if err := os.MkdirAll("guxgen/dist", 0755); err != nil {
		return err
	}

	fmt.Println("Building WASM...")

	var cmd *exec.Cmd
	if tinygo {
		cmd = exec.Command("tinygo", "build", "-o", "guxgen/dist/app.wasm", "-target", "wasm", "./guxgen/wasm")
	} else {
		cmd = exec.Command("go", "build", "-o", "guxgen/dist/app.wasm", "./guxgen/wasm")
		cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("WASM build failed: %w", err)
	}

	// Get size for display
	info, err := os.Stat("guxgen/dist/app.wasm")
	if err != nil {
		return err
	}
	sizeMB := float64(info.Size()) / 1024 / 1024
	compiler := "TinyGo"
	if !tinygo {
		compiler = "Go"
	}
	fmt.Printf("Built guxgen/dist/app.wasm (%.2f MB) with %s\n", sizeMB, compiler)

	return nil
}

// copyWasmExec copies wasm_exec.js from TinyGo/Go installation
func copyWasmExec(tinygo bool) error {
	// Ensure dist directory exists
	if err := os.MkdirAll("guxgen/dist", 0755); err != nil {
		return err
	}

	var src string
	if tinygo {
		cmd := exec.Command("tinygo", "env", "TINYGOROOT")
		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("TinyGo not found: %w", err)
		}
		src = filepath.Join(strings.TrimSpace(string(out)), "targets", "wasm_exec.js")
	} else {
		cmd := exec.Command("go", "env", "GOROOT")
		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("Go not found: %w", err)
		}
		src = filepath.Join(strings.TrimSpace(string(out)), "lib", "wasm", "wasm_exec.js")
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read wasm_exec.js: %w", err)
	}

	return os.WriteFile("guxgen/dist/wasm_exec.js", data, 0644)
}

// getGuxModulePath finds the gux module in the Go module cache
func getGuxModulePath() (string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/dougbarrett/gux")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to find gux module: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// readSafelist reads the Tailwind safelist file from the gux module
func readSafelist(guxPath string) ([]string, error) {
	safelistPath := filepath.Join(guxPath, "ui", "tailwind-safelist.txt")
	data, err := os.ReadFile(safelistPath)
	if err != nil {
		return nil, err
	}

	var classes []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		classes = append(classes, line)
	}
	return classes, nil
}

// generateSafelistFile generates a file containing class names for Tailwind to scan
func generateSafelistFile(classes []string) error {
	// Create a file that Tailwind's @source can scan for class names
	// Format: class="classname" so Tailwind recognizes them
	var sb strings.Builder
	sb.WriteString("<!-- Auto-generated safelist for gux/ui components -->\n")
	sb.WriteString("<!-- Tailwind scans this file for class names -->\n")

	for _, class := range classes {
		sb.WriteString(fmt.Sprintf("<div class=\"%s\"></div>\n", class))
	}

	return os.WriteFile("guxgen/styles/safelist.html", []byte(sb.String()), 0644)
}

// generateTailwindConfig generates Tailwind config files in guxgen/
func generateTailwindConfig() error {
	if err := os.MkdirAll("guxgen/styles", 0755); err != nil {
		return err
	}

	// Try to find gux module and read safelist
	var safelistImport string
	guxPath, err := getGuxModulePath()
	if err == nil {
		classes, err := readSafelist(guxPath)
		if err == nil && len(classes) > 0 {
			if err := generateSafelistFile(classes); err == nil {
				safelistImport = "\n/* Include gux/ui component classes */\n@source \"./safelist.html\";\n"
				fmt.Printf("  Including %d classes from gux/ui safelist\n", len(classes))
			}
		}
	}

	// Generate input.css for Tailwind v4+
	// Uses @source directive to scan Go files for class names
	inputCSS := fmt.Sprintf(`@import "tailwindcss";

/* Enable class-based dark mode (requires class="dark" on html element) */
@custom-variant dark (&:where(.dark, .dark *));
%s
/* Scan Go files for Tailwind classes */
@source "./pages/**/*.go";
@source "./components/**/*.go";
@source "./*.go";
`, safelistImport)

	return os.WriteFile("guxgen/styles/input.css", []byte(inputCSS), 0644)
}

// buildTailwind builds Tailwind CSS using the CLI
func buildTailwind() error {
	fmt.Println("Building Tailwind CSS...")

	// Generate config files
	if err := generateTailwindConfig(); err != nil {
		return fmt.Errorf("failed to generate Tailwind config: %w", err)
	}

	if err := os.MkdirAll("guxgen/dist", 0755); err != nil {
		return err
	}

	// Try global tailwindcss first (works with npm global, ruby gem, standalone binary)
	// Then fall back to npx if global not found
	// Tailwind v4 doesn't need -c config flag, uses CSS directives
	args := []string{
		"-i", "guxgen/styles/input.css",
		"-o", "guxgen/dist/styles.css",
		"--minify",
	}

	cmd := exec.Command("tailwindcss", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Try npx as fallback (for local npm installs)
		cmd = exec.Command("npx", append([]string{"tailwindcss"}, args...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Tailwind CSS build failed (install with: npm install -D tailwindcss or gem install tailwindcss-ruby): %w", err)
		}
	}

	// Get size for display
	info, err := os.Stat("guxgen/dist/styles.css")
	if err != nil {
		return err
	}
	sizeKB := float64(info.Size()) / 1024
	fmt.Printf("Built guxgen/dist/styles.css (%.2f KB)\n", sizeKB)

	return nil
}

// buildBinary builds the final server binary
func buildBinary() error {
	fmt.Println("Building binary...")

	cmd := exec.Command("go", "build", "-o", "bin/app", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("binary build failed: %w", err)
	}

	info, err := os.Stat("bin/app")
	if err != nil {
		return err
	}
	sizeMB := float64(info.Size()) / 1024 / 1024
	fmt.Printf("Built bin/app (%.2f MB)\n", sizeMB)

	return nil
}

// runBuildNew is the new build command for the simplified architecture
func runBuildNew(tinygo bool) {
	// Get module path
	modulePath, err := getModulePath()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Find main app file
	appFile, err := findMainAppFile()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Parse routes and bundles from app file
	bundles, imports, err := parseRoutesAndBundles(appFile)
	if err != nil {
		fmt.Printf("Error parsing routes: %v\n", err)
		os.Exit(1)
	}

	// Count total routes
	totalRoutes := 0
	for _, bundle := range bundles {
		totalRoutes += len(bundle.Routes)
	}
	if totalRoutes == 0 {
		fmt.Println("Warning: no Hybrid routes found")
	}

	// Collect bundle names
	bundleNames := make([]string, 0, len(bundles))
	for name := range bundles {
		bundleNames = append(bundleNames, name)
	}

	// Parse CRUD models from app file
	crudModels, modelsImport, dtoImport, err := parseCRUDModels(appFile)
	if err != nil {
		fmt.Printf("Error parsing CRUD models: %v\n", err)
		os.Exit(1)
	}

	// Check if we have a gux.config.json - if so, use guxgen/ paths
	hasGuxConfig := false
	if config, err := LoadModelsConfig("."); err == nil && len(config.Models) > 0 {
		hasGuxConfig = true
	}

	// For projects with gux.config.json, always use guxgen/ paths
	// This ensures generated DTOs and models are correctly imported
	if hasGuxConfig {
		pkgPath, err := getCurrentPackagePath()
		if err != nil {
			fmt.Printf("Error getting package path: %v\n", err)
			os.Exit(1)
		}
		modelsImport = pkgPath + "/guxgen/models"
		dtoImport = pkgPath + "/guxgen/dto"
	} else {
		// Legacy: If no models import found, construct it from current package path
		if modelsImport == "" && len(crudModels) > 0 {
			pkgPath, err := getCurrentPackagePath()
			if err != nil {
				fmt.Printf("Error getting package path: %v\n", err)
				os.Exit(1)
			}
			modelsImport = pkgPath + "/guxgen/models"
		}

		// Legacy: If no dto import found but we have DTOs, construct it
		if dtoImport == "" {
			for _, m := range crudModels {
				if m.ListDTO != "" || m.DetailDTO != "" {
					pkgPath, err := getCurrentPackagePath()
					if err == nil {
						dtoImport = pkgPath + "/guxgen/dto"
					}
					break
				}
			}
		}
	}

	// Generate API client if there are CRUD models
	if len(crudModels) > 0 {
		fmt.Println("Generating API client...")

		// Parse DTO files to get field mappings (check guxgen/dto first, then dto)
		for i := range crudModels {
			m := &crudModels[i]
			if m.ListDTO != "" {
				info, err := parseDTOFile("guxgen/dto", m.ListDTO)
				if err != nil {
					// Fallback to dto/ for backwards compatibility
					info, err = parseDTOFile("dto", m.ListDTO)
				}
				if err != nil {
					fmt.Printf("Warning: could not parse DTO %s: %v\n", m.ListDTO, err)
				} else {
					m.ListDTOInfo = info
				}
			}
			if m.DetailDTO != "" {
				info, err := parseDTOFile("guxgen/dto", m.DetailDTO)
				if err != nil {
					// Fallback to dto/ for backwards compatibility
					info, err = parseDTOFile("dto", m.DetailDTO)
				}
				if err != nil {
					fmt.Printf("Warning: could not parse DTO %s: %v\n", m.DetailDTO, err)
				} else {
					m.DetailDTOInfo = info
				}
			}
		}

		if err := generateAPIClient(modelsImport, dtoImport, crudModels); err != nil {
			fmt.Printf("Error generating API client: %v\n", err)
			os.Exit(1)
		}
	}

	// Generate WASM entry points for each bundle
	fmt.Println("Generating WASM entry points...")
	for name, bundle := range bundles {
		if len(bundle.Routes) == 0 {
			continue // Skip bundles with no routes
		}
		if err := generateBundleWasmEntryPoint(name, bundle); err != nil {
			fmt.Printf("Error generating WASM entry for bundle %s: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("  Generated entry point for bundle: %s (%d routes)\n", name, len(bundle.Routes))
	}

	// Copy wasm_exec.js
	if err := copyWasmExec(tinygo); err != nil {
		fmt.Printf("Error copying wasm_exec.js: %v\n", err)
		os.Exit(1)
	}

	// Build WASM for each bundle
	for name, bundle := range bundles {
		if len(bundle.Routes) == 0 {
			continue // Skip bundles with no routes
		}
		if err := buildWasmBundle(name, tinygo); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Build Tailwind CSS
	if err := buildTailwind(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Generate assets_gen.go with all bundles
	fmt.Println("Generating assets...")
	if err := generateAssetsFile(modulePath, bundleNames); err != nil {
		fmt.Printf("Error generating assets: %v\n", err)
		os.Exit(1)
	}

	// Silence unused variable warning for imports (used for debugging)
	_ = imports

	// Create bin directory
	if err := os.MkdirAll("bin", 0755); err != nil {
		fmt.Printf("Error creating bin/: %v\n", err)
		os.Exit(1)
	}

	// Build final binary
	if err := buildBinary(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nBuild complete! Run with: ./bin/app")
}

// runClean removes all generated files
func runClean() {
	removed := []string{}

	if err := os.RemoveAll("bin"); err == nil {
		if _, err := os.Stat("bin"); os.IsNotExist(err) {
			removed = append(removed, "bin/")
		}
	} else {
		removed = append(removed, "bin/")
	}

	if err := os.RemoveAll("guxgen"); err == nil {
		removed = append(removed, "guxgen/")
	}

	if err := os.Remove("assets_gen.go"); err == nil {
		removed = append(removed, "assets_gen.go")
	}

	if len(removed) > 0 {
		fmt.Printf("Cleaned: %s\n", strings.Join(removed, ", "))
	} else {
		fmt.Println("Nothing to clean")
	}
}

// runDevNew builds and runs the server (does NOT clean up on exit to preserve guxgen files)
func runDevNew(tinygo bool, watch bool) {
	// Build first
	runBuildNew(tinygo)

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("\nStarting dev server...\n")
	if watch {
		fmt.Println("Hot reload enabled - watching for file changes...")
	}

	// Channel to signal rebuild needed
	rebuildChan := make(chan struct{}, 1)

	// Start file watcher if watch mode enabled
	if watch {
		go watchAndRegenerate(true, func() {
			select {
			case rebuildChan <- struct{}{}:
			default:
			}
		})
	}

	for {
		// Run the binary
		cmd := exec.Command("./bin/app")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		// Set hot reload env var if watch mode enabled
		if watch {
			cmd.Env = append(os.Environ(), "GUX_HOT_RELOAD=1")
		}

		if err := cmd.Start(); err != nil {
			fmt.Printf("Failed to start: %v\n", err)
			os.Exit(1)
		}

		// Wait for signal, rebuild request, or process exit
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-sigChan:
			fmt.Println("\nShutting down...")
			cmd.Process.Signal(os.Interrupt)
			<-done
			return // Exit without cleaning
		case <-rebuildChan:
			// Kill current server and rebuild
			fmt.Println("\nRebuilding...")
			cmd.Process.Signal(os.Interrupt)
			<-done
			runBuildNew(tinygo)
			fmt.Println("\nRestarting server...")
			continue // Restart loop
		case err := <-done:
			if err != nil {
				fmt.Printf("Server exited with error: %v\n", err)
			}
			return // Exit without cleaning
		}
	}
}
