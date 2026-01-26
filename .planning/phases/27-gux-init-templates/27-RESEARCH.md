# Phase 27: gux init Templates - Research

**Researched:** 2026-01-26
**Domain:** Go CLI scaffolding, text/template, GORM models, Gux framework patterns
**Confidence:** HIGH

## Summary

This phase modernizes the `gux init` scaffolding to generate a functional application using the current Gux patterns (core.New(), SetDB(), CRUD(), Routes(), Run()). The existing scaffold.go already has the infrastructure (embed.FS, template rendering, go mod tidy execution) - the work is primarily about creating new template content and updating the file list.

The current templates follow an outdated client-server architecture (separate cmd/app and cmd/server binaries with internal/api). The new templates should produce a single app.go file that uses the hybrid SSR+WASM pattern demonstrated in examples/minimal, with models/ and pages/ directories.

**Primary recommendation:** Create templates that mirror the examples/minimal structure but simplified to a single CRUD model (Item), running `gux gen` after file creation to generate the .gux/api/ directory.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| text/template | Go stdlib | Template rendering | Built-in, no dependencies |
| embed | Go stdlib | Embed templates in binary | Built-in since Go 1.16 |
| gorm.io/gorm | v1.31.1 | ORM for database | Project standard |
| gorm.io/driver/sqlite | v1.6.0 | SQLite driver | Simple default database |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/dougbarrett/gux/core | latest | Framework core | All templates |
| github.com/dougbarrett/gux/ui | latest | UI components | Page templates |

### No Alternatives Needed
This phase uses established project patterns - no library decisions required.

**go.mod Template Dependencies:**
```go
require (
    github.com/dougbarrett/gux {{.GuxVersion}}
    gorm.io/driver/sqlite v1.6.0
    gorm.io/gorm v1.31.1
)
```

## Architecture Patterns

### Target Project Structure (After gux init)
```
myapp/
  app.go             # Main application (NEW template)
  go.mod             # Module file (UPDATED template)
  models/
    item.go          # Example GORM model (NEW template)
  pages/
    home.go          # Landing page (NEW template)
    items.go         # Items list (NEW template)
    item_new.go      # Item creation form (NEW template)
```

After `gux gen` runs automatically:
```
myapp/
  ...
  .gux/
    api/
      client.go      # Generated API client (WASM)
      client_server.go # Generated API client (Server)
    wasm/
      main.go        # Generated WASM entry point
  assets_gen.go      # Generated asset embedding
```

### Pattern 1: Single-File App Entry Point
**What:** All app setup in one app.go file
**When to use:** All scaffolded projects
**Example:**
```go
// Source: examples/minimal/app.go (verified)
package main

import (
    "log"

    "github.com/dougbarrett/gux/core"
    "myapp/.gux/api"
    "myapp/models"
    "myapp/pages"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func main() {
    // Database setup
    db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect database:", err)
    }
    db.AutoMigrate(&models.Item{})

    // Initialize API
    api.Init(db)

    // App setup
    app := core.New()
    app.SetDB(db)
    app.SetTitle("My App")

    // Register CRUD
    app.CRUD(models.Item{})

    // Routes
    app.Routes().
        Hybrid("/", pages.Home).
        Hybrid("/items", pages.Items).
        Hybrid("/items/new", pages.ItemNew)

    app.Run(":8080")
}
```

### Pattern 2: GORM Model with gorm.Model
**What:** Models embed gorm.Model for standard fields
**When to use:** All model templates
**Example:**
```go
// Source: examples/minimal/models/counter.go (verified)
package models

import "gorm.io/gorm"

type Item struct {
    gorm.Model
    Name        string `json:"name"`
    Description string `json:"description"`
}
```

### Pattern 3: Page Function Signature
**What:** PageFunc returns a loader that returns a component
**When to use:** All page templates
**Example:**
```go
// Source: examples/minimal/pages/home.go (verified)
func Home(r *core.Router) func() core.Node {
    // Loader (runs on server/navigation)
    var items []api.Item

    r.OnLoad(func() {
        api.Items.List(func(result []api.Item, err error) {
            if err == nil {
                items = result
            }
        })
    })

    // Component (reactive)
    return func() core.Node {
        return core.Div(core.Class("..."), ...)
    }
}
```

### Pattern 4: Form with State and Navigation
**What:** Form pages use StateString, validation, API calls, and r.Navigate()
**When to use:** item_new.go template
**Example:**
```go
// Source: examples/minimal/admin/user_new.go (verified pattern)
func ItemNew(r *core.Router) func() core.Node {
    return func() core.Node {
        nameState := r.StateString("name", "")
        descState := r.StateString("description", "")
        errorState := r.StateString("error", "")

        save := func() {
            if nameState.Get() == "" {
                errorState.Set("Name is required")
                return
            }
            data := map[string]interface{}{
                "name": nameState.Get(),
                "description": descState.Get(),
            }
            api.Items.Create(data, func(result *api.Item, err error) {
                if err != nil {
                    errorState.Set(err.Error())
                } else {
                    r.Navigate("/items")
                }
            })
        }

        // Return form UI...
    }
}
```

### Anti-Patterns to Avoid
- **Separate cmd/app and cmd/server:** Old pattern - use single app.go with hybrid routes
- **internal/api interfaces:** Old pattern - CRUD auto-generates API from models
- **Manual API client generation:** Use `gux gen` to auto-generate from CRUD registrations

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| API client | Custom fetch code | `gux gen` auto-generation | Handles CSRF, callbacks, error handling |
| WASM entry points | Manual wasm/main.go | `gux gen` auto-generation | Handles route registration, hydration |
| Form validation | Custom validators | StateString + conditional logic | Sufficient for templates |
| Table styling | Custom CSS | ui.Table components | Consistent, dark mode support |

**Key insight:** The scaffold should produce a minimal working app that uses the framework's code generation, not hand-written boilerplate.

## Common Pitfalls

### Pitfall 1: Not Running gux gen After File Creation
**What goes wrong:** App won't compile - missing .gux/api/ directory
**Why it happens:** CRUD models need code generation for API client
**How to avoid:** scaffold.go must run `gux gen` after creating files
**Warning signs:** "package not found" errors for .gux/api

### Pitfall 2: Template Variable Escaping
**What goes wrong:** {{.}} in template Go code gets interpreted as template variable
**Why it happens:** text/template parses all {{...}} as actions
**How to avoid:** Use `{{"{{"}}` escape sequence OR store code in variables
**Warning signs:** Missing code in generated files, template parse errors

### Pitfall 3: Missing api.Init(db) Call
**What goes wrong:** API calls return nil/error on server
**Why it happens:** Generated API server code needs DB reference
**How to avoid:** app.go template must call api.Init(db) before routes
**Warning signs:** "database not configured" errors, nil results

### Pitfall 4: gux gen Before models/ Exists
**What goes wrong:** gux gen fails or generates incomplete client
**Why it happens:** gux gen parses app.go and models/ to generate API
**How to avoid:** Create all source files before running gux gen
**Warning signs:** Empty or missing API client methods

### Pitfall 5: Import Path Mismatch
**What goes wrong:** Imports don't resolve after go mod tidy
**Why it happens:** Template uses wrong module path for .gux/api
**How to avoid:** Use {{.ModulePath}}/.gux/api in imports
**Warning signs:** "package not found" after generation

## Code Examples

### app.go Template
```go
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

    // Initialize API with database
    api.Init(db)

    // Create app
    app := core.New()
    app.SetDB(db)
    app.SetTitle("{{.AppName}}")

    // Register CRUD for Item model
    app.CRUD(models.Item{})

    // Register routes
    app.Routes().
        Hybrid("/", pages.Home).
        Hybrid("/items", pages.Items).
        Hybrid("/items/new", pages.ItemNew)

    app.Run(":8080")
}
```

### go.mod Template
```
module {{.ModulePath}}

go 1.24

require (
    {{.GuxModule}} {{.GuxVersion}}
    gorm.io/driver/sqlite v1.6.0
    gorm.io/gorm v1.31.1
)
```

### models/item.go Template
```go
package models

import "gorm.io/gorm"

// Item represents an example model with name and description.
type Item struct {
    gorm.Model
    Name        string `json:"name"`
    Description string `json:"description"`
}
```

### pages/home.go Template
```go
package pages

import (
    "github.com/dougbarrett/gux/core"
    "github.com/dougbarrett/gux/ui"
)

// Home is the landing page.
func Home(r *core.Router) func() core.Node {
    return func() core.Node {
        return core.Div(core.Class("min-h-screen bg-gray-100 dark:bg-gray-900"),
            core.Div(core.Class("max-w-4xl mx-auto px-4 py-8"),
                ui.Card(ui.CardProps{
                    Children: []core.Node{
                        ui.CardHeader(ui.CardHeaderProps{
                            Children: []core.Node{
                                core.H1(core.Class("text-2xl font-bold text-gray-900 dark:text-white"),
                                    core.Text("Welcome to {{.AppName}}"),
                                ),
                            },
                        }),
                        ui.CardContent(ui.CardContentProps{
                            Children: []core.Node{
                                core.P(core.Class("text-gray-600 dark:text-gray-300 mb-4"),
                                    core.Text("Your Gux application is ready. Start building!"),
                                ),
                                core.A(
                                    core.Attrs{
                                        Href:  "/items",
                                        Class: "text-blue-600 dark:text-blue-400 hover:underline",
                                    },
                                    core.Text("View Items"),
                                ),
                            },
                        }),
                    },
                }),
            ),
        )
    }
}
```

### pages/items.go Template
```go
package pages

import (
    "encoding/json"
    "fmt"

    "github.com/dougbarrett/gux/core"
    "github.com/dougbarrett/gux/ui"
    "{{.ModulePath}}/.gux/api"
)

// Items shows the list of items.
func Items(r *core.Router) func() core.Node {
    var items []api.Item

    r.OnLoad(func() {
        api.Items.List(func(result []api.Item, err error) {
            if err == nil {
                items = result
            }
        })
    })

    return func() core.Node {
        // Store items in state for hydration
        itemsJSON, _ := json.Marshal(items)
        itemsState := r.StateString("items", string(itemsJSON))

        // Parse items from state
        var displayItems []api.Item
        json.Unmarshal([]byte(itemsState.Get()), &displayItems)

        return core.Div(core.Class("min-h-screen bg-gray-100 dark:bg-gray-900"),
            core.Div(core.Class("max-w-4xl mx-auto px-4 py-8"),
                core.Div(core.Class("flex justify-between items-center mb-6"),
                    core.H1(core.Class("text-2xl font-bold text-gray-900 dark:text-white"),
                        core.Text("Items"),
                    ),
                    core.A(
                        core.Attrs{
                            Href:  "/items/new",
                            Class: "px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700",
                        },
                        core.Text("+ Add Item"),
                    ),
                ),
                ui.Card(ui.CardProps{
                    Children: []core.Node{
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
                                    Children: itemRows(displayItems),
                                }),
                            },
                        }),
                    },
                }),
            ),
        )
    }
}

func itemRows(items []api.Item) []core.Node {
    if len(items) == 0 {
        return []core.Node{
            ui.Tr(ui.TrProps{
                Children: []core.Node{
                    ui.Td(ui.TdProps{
                        Class:    "text-center text-gray-500",
                        Children: []core.Node{core.Text("No items yet")},
                    }),
                },
            }),
        }
    }

    var rows []core.Node
    for _, item := range items {
        rows = append(rows, ui.Tr(ui.TrProps{
            Children: []core.Node{
                ui.Td(ui.TdProps{Children: []core.Node{core.Text(fmt.Sprintf("%d", item.ID))}}),
                ui.Td(ui.TdProps{Children: []core.Node{core.Text(item.Name)}}),
                ui.Td(ui.TdProps{Children: []core.Node{core.Text(item.Description)}}),
            },
        }))
    }
    return rows
}
```

### pages/item_new.go Template
```go
package pages

import (
    "github.com/dougbarrett/gux/core"
    "github.com/dougbarrett/gux/ui"
    "{{.ModulePath}}/.gux/api"
)

// ItemNew is the form for creating a new item.
func ItemNew(r *core.Router) func() core.Node {
    return func() core.Node {
        nameState := r.StateString("name", "")
        descState := r.StateString("description", "")
        errorState := r.StateString("error", "")

        var errorNode core.Node = core.Frag()
        if errorState.Get() != "" {
            errorNode = ui.Alert(ui.AlertProps{
                Variant:  ui.AlertError,
                Children: []core.Node{core.Text(errorState.Get())},
            })
        }

        save := func() {
            if nameState.Get() == "" {
                errorState.Set("Name is required")
                return
            }
            data := map[string]interface{}{
                "name":        nameState.Get(),
                "description": descState.Get(),
            }
            api.Items.Create(data, func(result *api.Item, err error) {
                if err != nil {
                    errorState.Set(err.Error())
                } else {
                    r.Navigate("/items")
                }
            })
        }

        return core.Div(core.Class("min-h-screen bg-gray-100 dark:bg-gray-900"),
            core.Div(core.Class("max-w-2xl mx-auto px-4 py-8"),
                core.A(
                    core.Attrs{Href: "/items", Class: "text-blue-600 dark:text-blue-400 hover:underline mb-4 inline-block"},
                    core.Text("Back to Items"),
                ),
                core.H1(core.Class("text-2xl font-bold text-gray-900 dark:text-white mb-6"),
                    core.Text("Create New Item"),
                ),
                errorNode,
                ui.Card(ui.CardProps{
                    Children: []core.Node{
                        ui.CardContent(ui.CardContentProps{
                            Children: []core.Node{
                                ui.FormField(ui.FormFieldProps{
                                    Label:    "Name",
                                    Required: true,
                                    Children: []core.Node{
                                        ui.Input(ui.InputProps{
                                            Name:        "name",
                                            Value:       nameState.Get(),
                                            Placeholder: "Enter item name",
                                            OnChange:    func(v string) { nameState.Set(v) },
                                            OnEnter:     save,
                                        }),
                                    },
                                }),
                                ui.FormField(ui.FormFieldProps{
                                    Label: "Description",
                                    Children: []core.Node{
                                        ui.Textarea(ui.TextareaProps{
                                            Name:        "description",
                                            Value:       descState.Get(),
                                            Placeholder: "Enter item description",
                                            OnChange:    func(v string) { descState.Set(v) },
                                        }),
                                    },
                                }),
                                ui.Button(ui.ButtonProps{
                                    Text:    "Create Item",
                                    Variant: ui.ButtonPrimary,
                                    OnClick: save,
                                }),
                            },
                        }),
                    },
                }),
            ),
        )
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| cmd/app/main.go + cmd/server/main.go | Single app.go with hybrid routes | v2.0 | Simpler architecture |
| internal/api interfaces + @client annotations | app.CRUD() + auto-generation | v2.0 | Less boilerplate |
| Separate WASM and server binaries | Single binary with embedded WASM | v2.0 | Simpler deployment |

**Deprecated/outdated:**
- cmd/app/main.go.tmpl: Replaced by app.go
- cmd/server/main.go.tmpl: Replaced by app.go
- internal/api/*.tmpl: Replaced by models/ + CRUD auto-generation

## Open Questions

All technical questions resolved. No blockers.

## Sources

### Primary (HIGH confidence)
- `/Users/dougbarrett/projects/dbb1dev/goquery/cmd/gux/scaffold.go` - Current scaffold implementation
- `/Users/dougbarrett/projects/dbb1dev/goquery/examples/minimal/app.go` - Reference app.go pattern
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/app.go` - App/Router/Routes API
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/crud.go` - CRUD API
- `/Users/dougbarrett/projects/dbb1dev/goquery/cmd/gux/generate.go` - gux gen implementation
- `/Users/dougbarrett/projects/dbb1dev/goquery/ui/*.go` - UI component patterns

### Secondary (MEDIUM confidence)
- `/Users/dougbarrett/projects/dbb1dev/goquery/examples/minimal/` - Full example implementation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Using existing project dependencies
- Architecture: HIGH - Verified against examples/minimal
- Pitfalls: HIGH - Based on actual codebase analysis

**Research date:** 2026-01-26
**Valid until:** 60 days (stable internal patterns)
