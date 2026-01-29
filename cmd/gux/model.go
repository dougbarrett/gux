package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// SelectOption represents an option in a select/multiselect field
type SelectOption struct {
	Value string `json:"value"`
	Label string `json:"label,omitempty"`
}

// BadgeConfig configures how a boolean field displays as a badge
type BadgeConfig struct {
	TrueVariant  string `json:"true,omitempty"`  // Badge variant when true (e.g., "success")
	FalseVariant string `json:"false,omitempty"` // Badge variant when false (e.g., "secondary")
	TrueLabel    string `json:"trueLabel,omitempty"`
	FalseLabel   string `json:"falseLabel,omitempty"`
}

// ModelField represents a field in the model definition
type ModelField struct {
	Name       string         `json:"name"`                 // Field name (PascalCase)
	Type       string         `json:"type"`                 // Go type (string, int, *uint, []string, etc.)
	Required   bool           `json:"required,omitempty"`   // Is required
	Input      string         `json:"input,omitempty"`      // Input type override (date, email, tel, textarea, select, multiselect)
	Relation   string         `json:"relation,omitempty"`   // FK relation model name
	Label      string         `json:"label,omitempty"`      // Display label (defaults to field name)
	Table      bool           `json:"table,omitempty"`      // Show in table list view
	Priority   int            `json:"priority,omitempty"`   // Table column priority (1=high, shown first)
	Sortable   bool           `json:"sortable,omitempty"`   // Column is sortable
	Filterable bool           `json:"filterable,omitempty"` // Can filter by this field
	Editable   bool           `json:"editable,omitempty"`   // Inline editable in table
	Options    []SelectOption `json:"options,omitempty"`    // Static select options
	OptionsRef string         `json:"optionsRef,omitempty"` // Reference to optionSet
	Default    string         `json:"default,omitempty"`    // Default value
	Badge      *BadgeConfig   `json:"badge,omitempty"`      // Badge display config for booleans
	Many2Many  string         `json:"many2many,omitempty"`  // Join table name for M2M relations
	FullWidth  bool           `json:"fullWidth,omitempty"`  // Span full width in grid layout
	Section    string         `json:"-"`                    // Form section (set during parsing)
}

// ModelDefinition represents a complete model configuration
type ModelDefinition struct {
	Name       string                  `json:"-"`                      // Model name (from map key)
	Preset     string                  `json:"preset,omitempty"`       // Preset: "auth" for authentication models
	Display    string                  `json:"display,omitempty"`      // Display field name (for Brief DTOs and relation labels)
	Sections   map[string][]ModelField `json:"sections"`               // Fields grouped by section
	Preloads   []string                `json:"preloads,omitempty"`     // Relations to preload
	Public     bool                    `json:"public,omitempty"`       // CRUD is public (no auth)
	Roles      []string                `json:"roles,omitempty"`        // Required roles for CRUD
	FormLayout string                  `json:"formLayout,omitempty"`   // Form layout: "stacked" (default) or "grid"
}

// OptionSet defines a reusable set of select options
type OptionSet map[string]string // value -> label

// ModelsConfig extends GuxConfig with model definitions
type ModelsConfig struct {
	Models     map[string]ModelDefinition `json:"models,omitempty"`
	OptionSets map[string]OptionSet       `json:"optionSets,omitempty"`
	Roles      []SelectOption             `json:"roles,omitempty"` // Top-level roles config
}

// LoadModelsConfig reads model definitions from gux.config.json
func LoadModelsConfig(dir string) (*ModelsConfig, error) {
	configPath := filepath.Join(dir, configFileName)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config ModelsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set model names from map keys
	for name, model := range config.Models {
		model.Name = name
		config.Models[name] = model
	}

	return &config, nil
}

// SaveModelsConfig saves model definitions to gux.config.json
func SaveModelsConfig(dir string, modelsConfig *ModelsConfig) error {
	configPath := filepath.Join(dir, configFileName)

	// Load existing config first
	existingData, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var fullConfig map[string]any
	if len(existingData) > 0 {
		if err := json.Unmarshal(existingData, &fullConfig); err != nil {
			return err
		}
	} else {
		fullConfig = make(map[string]any)
	}

	// Update models and optionSets
	if modelsConfig.Models != nil {
		fullConfig["models"] = modelsConfig.Models
	}
	if modelsConfig.OptionSets != nil {
		fullConfig["optionSets"] = modelsConfig.OptionSets
	}

	data, err := json.MarshalIndent(fullConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// GetAllFields returns all fields from all sections in order
func (m *ModelDefinition) GetAllFields() []ModelField {
	var fields []ModelField
	for sectionName, sectionFields := range m.Sections {
		for _, field := range sectionFields {
			field.Section = sectionName
			fields = append(fields, field)
		}
	}
	return fields
}

// GetTableFields returns fields marked for table display, sorted by priority
func (m *ModelDefinition) GetTableFields() []ModelField {
	var tableFields []ModelField
	for sectionName, sectionFields := range m.Sections {
		for _, field := range sectionFields {
			if field.Table {
				field.Section = sectionName
				tableFields = append(tableFields, field)
			}
		}
	}

	// Sort by priority (lower = first), then by name
	for i := 0; i < len(tableFields)-1; i++ {
		for j := i + 1; j < len(tableFields); j++ {
			// If priority is 0, treat as lowest priority (highest number)
			pi := tableFields[i].Priority
			pj := tableFields[j].Priority
			if pi == 0 {
				pi = 999
			}
			if pj == 0 {
				pj = 999
			}
			if pi > pj {
				tableFields[i], tableFields[j] = tableFields[j], tableFields[i]
			}
		}
	}

	return tableFields
}

// GetRelationFields returns fields that have relations (FK or M2M)
func (m *ModelDefinition) GetRelationFields() []ModelField {
	var relFields []ModelField
	for _, field := range m.GetAllFields() {
		if field.Relation != "" {
			relFields = append(relFields, field)
		}
	}
	return relFields
}

// ToSnakeCase converts PascalCase to snake_case
// Handles common abbreviations like ID, URL, API, etc.
func ToSnakeCase(s string) string {
	var result strings.Builder
	n := len(s)
	for i := 0; i < n; i++ {
		r := rune(s[i])
		if i > 0 && r >= 'A' && r <= 'Z' {
			// Check if this is part of an abbreviation (consecutive uppercase)
			// Don't add underscore if previous char was uppercase and next char is uppercase or end
			prevUpper := s[i-1] >= 'A' && s[i-1] <= 'Z'
			nextUpper := i+1 < n && s[i+1] >= 'A' && s[i+1] <= 'Z'
			atEnd := i+1 == n

			if !prevUpper || (!nextUpper && !atEnd) {
				result.WriteRune('_')
			}
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// ToKebabCase converts PascalCase to kebab-case (for URL routes)
func ToKebabCase(s string) string {
	snake := ToSnakeCase(s)
	return strings.ReplaceAll(snake, "_", "-")
}

// ToPlural returns a simple English plural form
func ToPlural(s string) string {
	if s == "" {
		return s
	}
	lower := strings.ToLower(s)

	// Irregular plurals
	irregulars := map[string]string{
		"person": "people",
		"child":  "children",
		"man":    "men",
		"woman":  "women",
		"foot":   "feet",
		"tooth":  "teeth",
		"goose":  "geese",
		"mouse":  "mice",
	}
	if plural, ok := irregulars[lower]; ok {
		// Preserve original case
		if s[0] >= 'A' && s[0] <= 'Z' {
			return strings.ToUpper(string(plural[0])) + plural[1:]
		}
		return plural
	}

	// Words ending in consonant + y -> ies
	if len(s) > 1 && s[len(s)-1] == 'y' {
		prev := s[len(s)-2]
		if prev != 'a' && prev != 'e' && prev != 'i' && prev != 'o' && prev != 'u' {
			return s[:len(s)-1] + "ies"
		}
	}

	// Words ending in s, x, z, ch, sh -> es
	if strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") ||
		strings.HasSuffix(lower, "z") || strings.HasSuffix(lower, "ch") ||
		strings.HasSuffix(lower, "sh") {
		return s + "es"
	}

	// Default: add s
	return s + "s"
}

// GoTypeToInputType returns the default HTML input type for a Go type
func GoTypeToInputType(goType string) string {
	switch goType {
	case "time.Time", "*time.Time":
		return "date"
	case "int", "int32", "int64", "uint", "uint32", "uint64":
		return "number"
	case "float32", "float64":
		return "number"
	case "bool":
		return "checkbox"
	default:
		return "text"
	}
}

// IsPointerType returns true if the type is a pointer
func IsPointerType(goType string) bool {
	return strings.HasPrefix(goType, "*")
}

// IsSliceType returns true if the type is a slice
func IsSliceType(goType string) bool {
	return strings.HasPrefix(goType, "[]")
}

// GetBaseType returns the base type without pointer or slice prefix
func GetBaseType(goType string) string {
	goType = strings.TrimPrefix(goType, "*")
	goType = strings.TrimPrefix(goType, "[]")
	return goType
}

// runModelCommand handles the gux model subcommands
func runModelCommand(args []string) {
	if len(args) == 0 {
		printModelUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "add":
		runModelAdd(args[1:])
	case "list":
		runModelList()
	case "regen":
		if len(args) < 2 {
			fmt.Println("Error: model name required")
			fmt.Println("Usage: gux model regen <ModelName>")
			os.Exit(1)
		}
		runModelRegen(args[1])
	case "export":
		if len(args) < 2 {
			fmt.Println("Error: model name required")
			fmt.Println("Usage: gux model export <ModelName>")
			os.Exit(1)
		}
		runModelExport(args[1])
	case "help", "-h", "--help":
		printModelUsage()
	default:
		fmt.Printf("Unknown model subcommand: %s\n", args[0])
		printModelUsage()
		os.Exit(1)
	}
}

func printModelUsage() {
	fmt.Println(`gux model - Scaffold models, DTOs, and admin pages

Usage:
    gux model add                    Interactive model builder
    gux model add <Name>             Start with model name pre-filled
    gux model add --from-config      Generate all models from gux.config.json
    gux model list                   List models defined in config
    gux model regen <Name>           Regenerate files for a model
    gux model export <Name>          Export existing model to config

Examples:
    gux model add                    # Interactive wizard
    gux model add Product            # Pre-fill model name
    gux model add --from-config      # Generate from config
    gux model regen Client           # Regenerate Client files

Generated files:
    models/{name}_gen.go             GORM model
    dto/{name}_gen.go                List and Detail DTOs
    admin/{names}_gen.go             List page
    admin/{name}_new_gen.go          Create page
    admin/{name}_detail_gen.go       View/Edit page`)
}

func runModelAdd(args []string) {
	// Check for --from-config flag
	if len(args) > 0 && args[0] == "--from-config" {
		runModelAddFromConfig()
		return
	}

	// Get optional pre-filled model name
	var modelName string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		modelName = args[0]
	}

	// Run interactive builder
	RunInteractiveModelAdd(modelName)
}

func runModelAddFromConfig() {
	config, err := LoadModelsConfig(".")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if len(config.Models) == 0 {
		fmt.Println("No models defined in gux.config.json")
		fmt.Println("Use 'gux model add' to create models interactively.")
		os.Exit(1)
	}

	fmt.Printf("Generating %d model(s) from config...\n\n", len(config.Models))

	for name, model := range config.Models {
		model.Name = name
		fmt.Printf("Generating %s...\n", name)
		if err := GenerateModelFiles(&model, config.OptionSets); err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  Done\n")
		}
	}

	fmt.Println("\nAll models generated. Run 'gux gen' to update API client.")
}

func runModelList() {
	config, err := LoadModelsConfig(".")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No gux.config.json found.")
			fmt.Println("Use 'gux model add' to create models interactively.")
			return
		}
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if len(config.Models) == 0 {
		fmt.Println("No models defined in gux.config.json")
		fmt.Println("Use 'gux model add' to create models interactively.")
		return
	}

	fmt.Println("Models defined in gux.config.json:")
	for name, model := range config.Models {
		fieldCount := 0
		for _, fields := range model.Sections {
			fieldCount += len(fields)
		}
		fmt.Printf("  %-20s %d fields, %d sections\n", name, fieldCount, len(model.Sections))
	}
}

func runModelRegen(name string) {
	config, err := LoadModelsConfig(".")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	model, ok := config.Models[name]
	if !ok {
		fmt.Printf("Model '%s' not found in gux.config.json\n", name)
		fmt.Println("Available models:")
		for n := range config.Models {
			fmt.Printf("  %s\n", n)
		}
		os.Exit(1)
	}

	// If model uses auth preset and top-level roles are defined, apply them to Role field
	if model.Preset == "auth" && len(config.Roles) > 0 {
		applyRolesToAuthModel(&model, config.Roles)
	}

	model.Name = name
	fmt.Printf("Regenerating %s...\n", name)
	if err := GenerateModelFiles(&model, config.OptionSets); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done. Run 'gux gen' to update API client.")
}

// applyRolesToAuthModel updates the Role field options from top-level roles config
func applyRolesToAuthModel(model *ModelDefinition, roles []SelectOption) {
	for sectionName, fields := range model.Sections {
		for i, field := range fields {
			if field.Name == "Role" && field.Input == "select" {
				model.Sections[sectionName][i].Options = roles
			}
		}
	}
}

func runModelExport(name string) {
	// TODO: Parse existing model file and export to config
	fmt.Printf("Exporting model '%s' to gux.config.json...\n", name)
	fmt.Println("This feature is not yet implemented.")
	fmt.Println("For now, define models manually in gux.config.json or use 'gux model add'.")
}

// RunInteractiveModelAdd runs the interactive model builder
func RunInteractiveModelAdd(prefilledName string) {
	var modelName string = prefilledName
	var fields []ModelField
	var sections []string

	// Step 1: Model name (if not pre-filled)
	if modelName == "" {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Model name").
					Description("PascalCase, singular (e.g., Client, Product, Order)").
					Placeholder("Client").
					Value(&modelName).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("model name is required")
						}
						if s[0] < 'A' || s[0] > 'Z' {
							return fmt.Errorf("must start with uppercase letter")
						}
						return nil
					}),
			),
		)

		if err := form.Run(); err != nil {
			if err == huh.ErrUserAborted {
				fmt.Println("\nAborted.")
				return
			}
			fmt.Printf("Error: %v\n", err)
			return
		}
	}

	fmt.Printf("\nDefining model: %s\n\n", modelName)

	// Step 2: Add fields interactively
	continueAdding := true
	for continueAdding {
		field, err := promptForField(sections)
		if err != nil {
			if err == huh.ErrUserAborted {
				break
			}
			fmt.Printf("Error: %v\n", err)
			break
		}

		// Track new sections
		if field.Section != "" && !slices.Contains(sections, field.Section) {
			sections = append(sections, field.Section)
		}

		fields = append(fields, *field)

		// Field summary
		tags := []string{}
		if field.Required {
			tags = append(tags, "required")
		}
		if field.Table {
			tags = append(tags, "table")
		}
		if field.Relation != "" {
			tags = append(tags, "->"+field.Relation)
		}
		tagStr := ""
		if len(tags) > 0 {
			tagStr = " [" + strings.Join(tags, ", ") + "]"
		}
		fmt.Printf("  Added: %s (%s)%s\n", field.Name, field.Type, tagStr)

		// Continue?
		continueForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Add another field?").
					Value(&continueAdding),
			),
		)
		if err := continueForm.Run(); err != nil {
			break
		}
	}

	if len(fields) == 0 {
		fmt.Println("\nNo fields added. Exiting.")
		return
	}

	// Summary
	fmt.Printf("\nModel Summary: %s\n", modelName)
	fmt.Println(strings.Repeat("-", 50))
	for _, f := range fields {
		tags := []string{}
		if f.Required {
			tags = append(tags, "required")
		}
		if f.Table {
			tags = append(tags, "table")
		}
		if f.Relation != "" {
			tags = append(tags, "->"+f.Relation)
		}
		tagStr := ""
		if len(tags) > 0 {
			tagStr = " [" + strings.Join(tags, ", ") + "]"
		}
		fmt.Printf("  %s: %s%s\n", f.Name, f.Type, tagStr)
	}

	// Generate?
	var doGenerate bool
	var saveToConfig bool

	genForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Generate model, DTOs, and admin pages?").
				Value(&doGenerate),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Save model definition to gux.config.json?").
				Description("Allows regeneration with 'gux model regen'").
				Value(&saveToConfig),
		),
	)
	if err := genForm.Run(); err != nil {
		return
	}

	// Build model definition
	model := &ModelDefinition{
		Name:     modelName,
		Sections: make(map[string][]ModelField),
	}

	// Group fields by section
	for _, field := range fields {
		section := field.Section
		if section == "" {
			section = "General"
		}
		model.Sections[section] = append(model.Sections[section], field)
	}

	// Collect preloads from relation fields
	for _, field := range fields {
		if field.Relation != "" {
			// Preload name is the field name without ID suffix
			preloadName := strings.TrimSuffix(field.Name, "ID")
			if preloadName == field.Name {
				preloadName = field.Relation
			}
			model.Preloads = append(model.Preloads, preloadName)
		}
	}

	if saveToConfig {
		config, _ := LoadModelsConfig(".")
		if config == nil {
			config = &ModelsConfig{
				Models: make(map[string]ModelDefinition),
			}
		}
		if config.Models == nil {
			config.Models = make(map[string]ModelDefinition)
		}
		config.Models[modelName] = *model
		if err := SaveModelsConfig(".", config); err != nil {
			fmt.Printf("Warning: could not save to config: %v\n", err)
		} else {
			fmt.Println("Saved to gux.config.json")
		}
	}

	if doGenerate {
		fmt.Printf("\nGenerating files for %s...\n", modelName)
		if err := GenerateModelFiles(model, nil); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\nDone! Run 'gux gen' to update API client, then 'gux dev' to see your pages.")
	}
}

// promptForField prompts the user for a single field definition
func promptForField(existingSections []string) (*ModelField, error) {
	field := &ModelField{}

	// Field name and type
	var fieldType string
	basicForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Field name").
				Description("PascalCase (e.g., FirstName, LeadDate)").
				Value(&field.Name).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("field name is required")
					}
					if s[0] < 'A' || s[0] > 'Z' {
						return fmt.Errorf("must start with uppercase letter")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Field type").
				Options(
					huh.NewOption("string", "string"),
					huh.NewOption("int", "int"),
					huh.NewOption("uint", "uint"),
					huh.NewOption("*uint (nullable foreign key)", "*uint"),
					huh.NewOption("bool", "bool"),
					huh.NewOption("float64", "float64"),
					huh.NewOption("time.Time", "time.Time"),
					huh.NewOption("*time.Time (nullable date)", "*time.Time"),
					huh.NewOption("[]string (tags/multi-select)", "[]string"),
				).
				Value(&fieldType),
		),
	)

	if err := basicForm.Run(); err != nil {
		return nil, err
	}
	field.Type = fieldType

	// Section selection
	sectionOptions := []huh.Option[string]{
		huh.NewOption("(Create new section)", "_new"),
	}
	for _, s := range existingSections {
		sectionOptions = append(sectionOptions, huh.NewOption(s, s))
	}
	if len(existingSections) == 0 {
		sectionOptions = append(sectionOptions, huh.NewOption("General", "General"))
	}

	var selectedSection string
	sectionForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Form section").
				Description("Group related fields together").
				Options(sectionOptions...).
				Value(&selectedSection),
		),
	)

	if err := sectionForm.Run(); err != nil {
		return nil, err
	}

	if selectedSection == "_new" {
		var newSectionName string
		newForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("New section name").
					Placeholder("Lead Info").
					Value(&newSectionName),
			),
		)
		if err := newForm.Run(); err != nil {
			return nil, err
		}
		field.Section = newSectionName
	} else {
		field.Section = selectedSection
	}

	// Type-specific options
	switch field.Type {
	case "*uint":
		// Foreign key - ask for relation
		if err := promptForRelation(field); err != nil {
			return nil, err
		}

	case "string":
		// String input type
		if err := promptForStringInput(field); err != nil {
			return nil, err
		}

	case "[]string":
		// Multi-select options
		if err := promptForMultiSelectOptions(field); err != nil {
			return nil, err
		}

	case "time.Time", "*time.Time":
		field.Input = "date"

	case "bool":
		// Ask about badge display
		if err := promptForBooleanDisplay(field); err != nil {
			return nil, err
		}
	}

	// Common attributes
	if err := promptForAttributes(field); err != nil {
		return nil, err
	}

	return field, nil
}

func promptForRelation(field *ModelField) error {
	relForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Related model").
				Description("Model this FK points to (e.g., User, Category)").
				Placeholder("User").
				Value(&field.Relation),
			huh.NewInput().
				Title("Display label").
				Description("How to label this in forms (e.g., 'Salesperson')").
				Placeholder(field.Name).
				Value(&field.Label),
		),
	)
	return relForm.Run()
}

func promptForStringInput(field *ModelField) error {
	var inputType string
	inputForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Input type").
				Options(
					huh.NewOption("Text (default)", "text"),
					huh.NewOption("Email", "email"),
					huh.NewOption("Phone", "tel"),
					huh.NewOption("URL", "url"),
					huh.NewOption("Textarea (multiline)", "textarea"),
					huh.NewOption("Select (dropdown)", "select"),
				).
				Value(&inputType),
		),
	)

	if err := inputForm.Run(); err != nil {
		return err
	}

	if inputType != "text" {
		field.Input = inputType
	}

	// If select, prompt for options
	if inputType == "select" {
		return promptForSelectOptions(field)
	}

	return nil
}

func promptForSelectOptions(field *ModelField) error {
	var optionsText string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title("Options").
				Description("One per line: value or value:Label").
				Placeholder("draft:Draft\nactive:Active\narchived:Archived").
				Value(&optionsText),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Default value").
				Description("Optional - leave empty for no default").
				Value(&field.Default),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	// Parse options
	for _, line := range strings.Split(optionsText, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		opt := SelectOption{Value: parts[0]}
		if len(parts) == 2 {
			opt.Label = parts[1]
		} else {
			// Title case the value for label
			caser := cases.Title(language.English)
			opt.Label = caser.String(parts[0])
		}
		field.Options = append(field.Options, opt)
	}

	return nil
}

func promptForMultiSelectOptions(field *ModelField) error {
	var sourceType string
	sourceForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Multi-select type").
				Options(
					huh.NewOption("Static options (stored as JSON array)", "static"),
					huh.NewOption("Many-to-many relationship", "m2m"),
				).
				Value(&sourceType),
		),
	)

	if err := sourceForm.Run(); err != nil {
		return err
	}

	if sourceType == "static" {
		field.Input = "multiselect"
		return promptForSelectOptions(field)
	}

	// M2M relationship
	var relatedModel string
	var joinTable string

	m2mForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Related model").
				Description("Model to link (e.g., Category, Tag)").
				Value(&relatedModel),
			huh.NewInput().
				Title("Join table name").
				Description("e.g., product_categories").
				Value(&joinTable),
		),
	)

	if err := m2mForm.Run(); err != nil {
		return err
	}

	field.Type = "[]" + relatedModel
	field.Relation = relatedModel
	field.Many2Many = joinTable
	field.Input = "multiselect"

	return nil
}

func promptForBooleanDisplay(field *ModelField) error {
	var useBadge bool
	badgeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Display as colored badge in table?").
				Value(&useBadge),
		),
	)

	if err := badgeForm.Run(); err != nil {
		return err
	}

	if useBadge {
		field.Badge = &BadgeConfig{
			TrueVariant:  "success",
			FalseVariant: "secondary",
			TrueLabel:    "Yes",
			FalseLabel:   "No",
		}

		// Custom labels
		var trueLabel, falseLabel string
		labelForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Label when true").
					Placeholder("Yes").
					Value(&trueLabel),
				huh.NewInput().
					Title("Label when false").
					Placeholder("No").
					Value(&falseLabel),
			),
		)

		if err := labelForm.Run(); err != nil {
			return err
		}

		if trueLabel != "" {
			field.Badge.TrueLabel = trueLabel
		}
		if falseLabel != "" {
			field.Badge.FalseLabel = falseLabel
		}
	}

	return nil
}

func promptForAttributes(field *ModelField) error {
	var attrChoices []string
	attrForm := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Field attributes").
				Description("Select all that apply").
				Options(
					huh.NewOption("Required", "required"),
					huh.NewOption("Show in table list", "table"),
					huh.NewOption("Sortable column", "sortable"),
					huh.NewOption("Filterable", "filterable"),
					huh.NewOption("Inline editable", "editable"),
				).
				Value(&attrChoices),
		),
	)

	if err := attrForm.Run(); err != nil {
		return err
	}

	for _, attr := range attrChoices {
		switch attr {
		case "required":
			field.Required = true
		case "table":
			field.Table = true
		case "sortable":
			field.Sortable = true
		case "filterable":
			field.Filterable = true
		case "editable":
			field.Editable = true
		}
	}

	// Table priority if shown in table
	if field.Table {
		var priority string
		priorityForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Column priority").
					Description("Lower = appears more to the left").
					Options(
						huh.NewOption("High (1) - Always visible", "1"),
						huh.NewOption("Medium (2) - Usually visible", "2"),
						huh.NewOption("Low (3) - Hidden on mobile", "3"),
					).
					Value(&priority),
			),
		)

		if err := priorityForm.Run(); err != nil {
			return err
		}

		switch priority {
		case "1":
			field.Priority = 1
		case "2":
			field.Priority = 2
		case "3":
			field.Priority = 3
		}
	}

	return nil
}

// GenerateModelFiles generates all files for a model definition
func GenerateModelFiles(model *ModelDefinition, optionSets map[string]OptionSet) error {
	// Get module path and admin config
	config, err := LoadConfig(".")
	if err != nil {
		return fmt.Errorf("load config for module path: %w (run from a gux project directory)", err)
	}

	// Build display fields map from all models in config
	displayFields := BuildDisplayFieldsMap(".")

	// Build set of models defined in gux.config.json (for detecting external relations)
	configModels := BuildConfigModelsSet(".")

	return GenerateModelFilesImpl(model, optionSets, config.Module, config.Admin, displayFields, configModels)
}

// BuildConfigModelsSet returns a set of model names defined in gux.config.json
func BuildConfigModelsSet(dir string) map[string]bool {
	configModels := make(map[string]bool)
	modelsConfig, err := LoadModelsConfig(dir)
	if err != nil {
		return configModels // Return empty map if config not found
	}
	for name := range modelsConfig.Models {
		configModels[name] = true
	}
	return configModels
}

// BuildDisplayFieldsMap creates a map of model name -> display field from the config
func BuildDisplayFieldsMap(dir string) map[string]string {
	displayFields := make(map[string]string)
	modelsConfig, err := LoadModelsConfig(dir)
	if err != nil {
		return displayFields // Return empty map if config not found
	}
	for name, model := range modelsConfig.Models {
		// Check explicit display field from config first
		if model.Display != "" {
			displayFields[name] = model.Display
			continue
		}

		// Auto-detect: check candidate fields
		candidates := []string{"Name", "Title", "FirstName", "Label"}
		found := false

		for _, candidate := range candidates {
			for _, fields := range model.Sections {
				for _, f := range fields {
					if f.Name == candidate && f.Type == "string" {
						displayFields[name] = candidate
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if found {
				break
			}
		}

		// Fallback to first string field
		if !found {
			for _, fields := range model.Sections {
				for _, field := range fields {
					if field.Type == "string" {
						displayFields[name] = field.Name
						found = true
						break
					}
				}
				if found {
					break
				}
			}
		}
	}
	return displayFields
}
