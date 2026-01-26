package main

import (
	"fmt"
	"os"
	"sort"
	"text/template"

	"golang.org/x/mod/modfile"
)

// Pattern holds metadata and template for a help pattern.
type Pattern struct {
	Name        string // e.g., "page", "page:list"
	Description string // e.g., "Basic page template with OnLoad pattern"
	FilePath    string // e.g., "pages/example.go"
	Template    string // The actual template code
}

// patterns is the registry of all available help patterns.
var patterns = map[string]Pattern{
	"page": {
		Name:        "page",
		Description: "Basic page template with Card layout",
		FilePath:    "pages/example.go",
		Template: `// pages/example.go
package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Example renders the example page.
func Example(r *core.Router) func() core.Node {
	return func() core.Node {
		return core.Div(core.Class("min-h-screen bg-gray-100 dark:bg-gray-900"),
			core.Div(core.Class("max-w-4xl mx-auto px-4 py-8"),
				ui.Card(ui.CardProps{
					Children: []core.Node{
						ui.CardHeader(ui.CardHeaderProps{
							Children: []core.Node{
								core.H1(core.Class("text-3xl font-bold text-gray-900 dark:text-white"),
									core.Text("Page Title"),
								),
							},
						}),
						ui.CardContent(ui.CardContentProps{
							Children: []core.Node{
								core.P(core.Class("text-gray-600 dark:text-gray-400 mb-4"),
									core.Text("Page content goes here."),
								),
							},
						}),
					},
				}),
			),
		)
	}
}
`,
	},
	"page:list": {
		Name:        "page:list",
		Description: "Page with data fetching and table display",
		FilePath:    "pages/items.go",
		Template: `// pages/items.go
package pages

import (
	"encoding/json"
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
	"{{.ModulePath}}/.gux/api"
)

// Items displays a list of items.
func Items(r *core.Router) func() core.Node {
	// Loader - runs on server, fetches data
	var items []api.Item
	r.OnLoad(func() {
		api.Items.List(func(result []api.Item, err error) {
			if err == nil {
				items = result
			}
		})
	})

	// Component - reactive rendering
	return func() core.Node {
		// State for hydration - serialize items to JSON string
		itemsJSON, _ := json.Marshal(items)
		itemsState := r.StateString("items", string(itemsJSON))

		// Deserialize items from state
		var currentItems []api.Item
		if itemsState.Get() != "" {
			json.Unmarshal([]byte(itemsState.Get()), &currentItems)
		}

		return core.Div(core.Class("min-h-screen bg-gray-100 dark:bg-gray-900"),
			core.Div(core.Class("max-w-6xl mx-auto px-4 py-8"),
				// Header
				core.Div(core.Class("flex justify-between items-center mb-6"),
					core.H1(core.Class("text-3xl font-bold text-gray-900 dark:text-white"),
						core.Text("Items"),
					),
					core.A(
						core.Attrs{
							Href:  "/items/new",
							Class: "px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition",
						},
						core.Text("Add Item"),
					),
				),

				// Table
				ui.Table(ui.TableProps{
					Children: []core.Node{
						ui.Thead(ui.TheadProps{
							Children: []core.Node{
								ui.Tr(ui.TrProps{
									Children: []core.Node{
										ui.Th(ui.ThProps{Children: []core.Node{core.Text("ID")}}),
										ui.Th(ui.ThProps{Children: []core.Node{core.Text("Name")}}),
										ui.Th(ui.ThProps{Children: []core.Node{core.Text("Description")}}),
									},
								}),
							},
						}),
						ui.Tbody(ui.TbodyProps{
							Children: itemRows(currentItems),
						}),
					},
				}),
			),
		)
	}
}

// itemRows converts items to table rows.
func itemRows(items []api.Item) []core.Node {
	if len(items) == 0 {
		return []core.Node{
			ui.Tr(ui.TrProps{
				Children: []core.Node{
					ui.Td(ui.TdProps{
						Class: "text-center text-gray-500 dark:text-gray-400 py-8",
						Attrs: core.Attrs{
							Colspan: "3",
						},
						Children: []core.Node{core.Text("No items yet")},
					}),
				},
			}),
		}
	}

	rows := make([]core.Node, len(items))
	for i, item := range items {
		rows[i] = ui.Tr(ui.TrProps{
			Children: []core.Node{
				ui.Td(ui.TdProps{
					Children: []core.Node{
						core.Text(fmt.Sprintf("%d", item.ID)),
					},
				}),
				ui.Td(ui.TdProps{
					Children: []core.Node{
						core.Text(item.Name),
					},
				}),
				ui.Td(ui.TdProps{
					Children: []core.Node{
						core.Text(item.Description),
					},
				}),
			},
		})
	}
	return rows
}
`,
	},
	"page:form": {
		Name:        "page:form",
		Description: "Page with form state, validation, and API submission",
		FilePath:    "pages/item_new.go",
		Template: `// pages/item_new.go
package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
	"{{.ModulePath}}/.gux/api"
)

// ItemNew is the form for creating a new item.
func ItemNew(r *core.Router) func() core.Node {
	return func() core.Node {
		// Form state
		nameState := r.StateString("name", "")
		descState := r.StateString("description", "")
		errorState := r.StateString("error", "")

		// Save function
		save := func() {
			// Validate
			if nameState.Get() == "" {
				errorState.Set("Name is required")
				return
			}

			// Create item via generated API
			data := map[string]interface{}{
				"name":        nameState.Get(),
				"description": descState.Get(),
			}
			api.Items.Create(data, func(result *api.Item, err error) {
				if err != nil {
					errorState.Set(err.Error())
				} else {
					// Clear error and navigate to items list
					errorState.Set("")
					r.Navigate("/items")
				}
			})
		}

		// Error message node (conditional)
		var errorNode core.Node = core.Frag()
		if errorState.Get() != "" {
			errorNode = ui.Alert(ui.AlertProps{
				Variant: "destructive",
				Children: []core.Node{
					core.Text(errorState.Get()),
				},
			})
		}

		return core.Div(core.Class("min-h-screen bg-gray-100 dark:bg-gray-900"),
			core.Div(core.Class("max-w-2xl mx-auto px-4 py-8"),
				// Back link
				core.A(
					core.Attrs{
						Href:  "/items",
						Class: "text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 mb-4 inline-block",
					},
					core.Text("Back to Items"),
				),

				// Page title
				core.H1(core.Class("text-3xl font-bold text-gray-900 dark:text-white mb-6"),
					core.Text("Create New Item"),
				),

				// Error message
				errorNode,

				// Form card
				ui.Card(ui.CardProps{
					Children: []core.Node{
						ui.CardContent(ui.CardContentProps{
							Children: []core.Node{
								// Name field
								ui.FormField(ui.FormFieldProps{
									Label:    "Name",
									Required: true,
									Children: []core.Node{
										ui.Input(ui.InputProps{
											Type:        "text",
											Placeholder: "Enter item name",
											Value:       nameState.Get(),
											OnChange: func(v string) {
												nameState.Set(v)
												// Clear error when user starts typing
												if errorState.Get() != "" {
													errorState.Set("")
												}
											},
										}),
									},
								}),

								// Description field
								ui.FormField(ui.FormFieldProps{
									Label: "Description",
									Children: []core.Node{
										ui.Textarea(ui.TextareaProps{
											Placeholder: "Enter item description (optional)",
											Value:       descState.Get(),
											OnChange: func(v string) {
												descState.Set(v)
											},
										}),
									},
								}),

								// Submit button
								core.Div(core.Class("mt-6"),
									ui.Button(ui.ButtonProps{
										Variant: "default",
										OnClick: save,
										Children: []core.Node{
											core.Text("Create Item"),
										},
									}),
								),
							},
						}),
					},
				}),
			),
		)
	}
}
`,
	},
	"model": {
		Name:        "model",
		Description: "GORM model with common fields",
		FilePath:    "models/item.go",
		Template: `// models/item.go
package models

import "gorm.io/gorm"

// Item is a GORM model example.
type Item struct {
	gorm.Model
	Name        string ` + "`" + `json:"name"` + "`" + `
	Description string ` + "`" + `json:"description"` + "`" + `
}
`,
	},
	"dto": {
		Name:        "dto",
		Description: "DTO structs with gux tags for field mapping",
		FilePath:    "dto/item.go",
		Template: `// dto/item.go
package dto

import "time"

// ItemList is the DTO for item list responses.
// The gux tags map DTO fields to model fields for automatic code generation.
type ItemList struct {
	ID        uint      ` + "`" + `json:"id" gux:"Item.ID"` + "`" + `
	Name      string    ` + "`" + `json:"name" gux:"Item.Name"` + "`" + `
	CreatedAt time.Time ` + "`" + `json:"created_at" gux:"Item.CreatedAt"` + "`" + `
}

// ItemDetail is the DTO for single item responses.
// Use preload tag to include related data.
type ItemDetail struct {
	ID          uint      ` + "`" + `json:"id" gux:"Item.ID"` + "`" + `
	Name        string    ` + "`" + `json:"name" gux:"Item.Name"` + "`" + `
	Description string    ` + "`" + `json:"description" gux:"Item.Description"` + "`" + `
	CreatedAt   time.Time ` + "`" + `json:"created_at" gux:"Item.CreatedAt"` + "`" + `
	UpdatedAt   time.Time ` + "`" + `json:"updated_at" gux:"Item.UpdatedAt"` + "`" + `
}
`,
	},
	"routes": {
		Name:        "routes",
		Description: "Route registration with Hybrid and RouteGroup",
		FilePath:    "app.go (routes section)",
		Template: `// app.go (routes section)
// Import your pages: "{{.ModulePath}}/pages"

// Register routes
app.Routes().
	Hybrid("/", pages.Home).
	Hybrid("/items", pages.Items).
	Hybrid("/items/new", pages.ItemNew).
	Hybrid("/items/:id", pages.ItemDetail)

// Route groups with separate WASM bundle
app.RouteGroup("/admin", core.WithBundle("admin")).
	Hybrid("/", admin.Dashboard).
	Hybrid("/users", admin.Users)
`,
	},
	"app": {
		Name:        "app",
		Description: "Complete app.go with database, CRUD, and routes",
		FilePath:    "app.go",
		Template: `// app.go
package main

import (
	"log"

	"github.com/dougbarrett/gux/core"
	"{{.ModulePath}}/.gux/api"
	"{{.ModulePath}}/models"
	"{{.ModulePath}}/pages"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Set up database
	db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// Auto-migrate models
	db.AutoMigrate(&models.Item{})

	// Initialize API with database for server-side queries
	api.Init(db)

	app := core.New()
	app.SetDB(db)
	app.SetTitle("My App")

	// Register CRUD for Item model
	// Creates: GET/POST /__gux_api/crud/items
	//          GET/PUT/DELETE /__gux_api/crud/items/:id
	app.CRUD(models.Item{})

	// Register routes
	app.Routes().
		Hybrid("/", pages.Home).
		Hybrid("/items", pages.Items).
		Hybrid("/items/new", pages.ItemNew)

	app.Run(":8080")
}
`,
	},
}

// getModulePathForHelp reads go.mod and extracts the module path.
// Uses modfile.ModulePath for tolerant parsing (doesn't require valid go.mod).
// Returns a placeholder and prints warning to stderr if go.mod is not found.
func getModulePathForHelp() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: Could not read go.mod. Using placeholder import path.")
		fmt.Fprintln(os.Stderr, "Run from a Go module directory or replace 'your-module-path' manually.")
		fmt.Fprintln(os.Stderr, "")
		return "your-module-path"
	}

	path := modfile.ModulePath(data)
	if path == "" {
		fmt.Fprintln(os.Stderr, "Warning: Could not parse module path from go.mod. Using placeholder.")
		fmt.Fprintln(os.Stderr, "Replace 'your-module-path' with your actual module path.")
		fmt.Fprintln(os.Stderr, "")
		return "your-module-path"
	}

	return path
}

// printPatternList prints all available patterns sorted alphabetically.
func printPatternList() {
	fmt.Println("Available patterns:")

	if len(patterns) == 0 {
		fmt.Println("  (no patterns registered yet)")
		fmt.Println("")
		fmt.Println("Patterns will be added in a future update.")
		return
	}

	// Sort pattern names alphabetically
	names := make([]string, 0, len(patterns))
	for name := range patterns {
		names = append(names, name)
	}
	sort.Strings(names)

	// Print each pattern
	for _, name := range names {
		p := patterns[name]
		fmt.Printf("  %-20s %s\n", name, p.Description)
	}

	fmt.Println("")
	fmt.Println("Usage: gux help <pattern>")
	fmt.Println("Example: gux help page > pages/mypage.go")
}

// runHelpPattern handles the `gux help <pattern>` command.
// If name is empty, lists all patterns.
// If pattern exists, renders it with module path substitution.
// If pattern not found, prints error and lists available patterns.
func runHelpPattern(name string) {
	if name == "" {
		printPatternList()
		return
	}

	pattern, exists := patterns[name]
	if !exists {
		fmt.Fprintf(os.Stderr, "Unknown pattern: %s\n\n", name)
		printPatternList()
		os.Exit(1)
	}

	// Get module path for template substitution
	modulePath := getModulePathForHelp()

	// Template data
	data := struct {
		ModulePath string
	}{
		ModulePath: modulePath,
	}

	// Parse and execute template
	tmpl, err := template.New(name).Parse(pattern.Template)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing pattern template: %v\n", err)
		os.Exit(1)
	}

	// Output to stdout for piping
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing pattern template: %v\n", err)
		os.Exit(1)
	}
}
