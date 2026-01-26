# Phase 28: Help Pattern Commands - Research

**Researched:** 2026-01-26
**Domain:** Go CLI commands, go.mod parsing, template-based code output
**Confidence:** HIGH

## Summary

This phase implements `gux help <pattern>` commands that output copy-pasteable Go code boilerplate for common Gux patterns. The help system acts as a quick reference for developers, providing production-ready code snippets that use the correct module path from the project's go.mod file.

The implementation involves: (1) parsing go.mod to extract the module path, (2) storing pattern templates (similar to existing scaffold templates), and (3) a new CLI command that selects and renders the appropriate template. The architecture follows the same text/template approach already used in scaffold.go.

This is not complex infrastructure work - it's primarily about organizing code snippets from Phase 27's templates and exposing them via a CLI interface. The go.mod parsing uses the official `golang.org/x/mod/modfile` package.

**Primary recommendation:** Add a `help.go` file to cmd/gux with pattern templates as string constants (not embedded files) since they're small and read-only. Use `modfile.ModulePath()` for simple, tolerant go.mod parsing.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| golang.org/x/mod/modfile | latest | Parse go.mod to get module path | Official Go team package |
| text/template | Go stdlib | Render templates with module path substitution | Already used in scaffold.go |
| flag | Go stdlib | CLI argument parsing | Already used in main.go |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| fmt | Go stdlib | Print output to stdout | All pattern output |
| os | Go stdlib | Read go.mod file | go.mod integration |

### No Alternatives Needed
This phase uses established project patterns and the official Go module parsing package.

**Installation (for modfile):**
```bash
go get golang.org/x/mod/modfile
```

## Architecture Patterns

### Recommended File Structure
```
cmd/gux/
  main.go              # Add "help" case to switch
  help.go              # NEW: Pattern definitions and help command
  scaffold.go          # Existing (reference for template approach)
```

### Pattern 1: Help Command Handler in main.go
**What:** Add case for `gux help [pattern]` that delegates to help.go
**When to use:** Consistent with existing main.go switch structure
**Example:**
```go
// Source: Matches existing pattern in cmd/gux/main.go
case "help", "-h", "--help":
    if len(os.Args) > 2 {
        // gux help <pattern>
        runHelpPattern(os.Args[2])
    } else {
        // gux help (no pattern)
        printUsage()
    }
```

### Pattern 2: Module Path Extraction
**What:** Read go.mod and extract module path using modfile.ModulePath
**When to use:** Before rendering any pattern that needs imports
**Example:**
```go
// Source: golang.org/x/mod/modfile package docs
func getModulePath() (string, error) {
    data, err := os.ReadFile("go.mod")
    if err != nil {
        return "", err
    }
    path := modfile.ModulePath(data)
    if path == "" {
        return "", errors.New("could not parse module path from go.mod")
    }
    return path, nil
}
```

### Pattern 3: Pattern Templates as String Constants
**What:** Define pattern templates as raw string literals in help.go
**When to use:** For all help patterns (small, read-only, no need for embed.FS)
**Example:**
```go
const pageTemplate = `// {{.FilePath}}
package pages

import (
    "github.com/dougbarrett/gux/core"
)

// {{.FuncName}} renders the {{.Name}} page.
func {{.FuncName}}(r *core.Router) func() core.Node {
    return func() core.Node {
        return core.Div(core.Class("min-h-screen"),
            core.Text("Hello from {{.Name}}"),
        )
    }
}
`
```

### Pattern 4: Pattern Registry Map
**What:** Map pattern names to their templates and metadata
**When to use:** Clean dispatch for `gux help <pattern>`
**Example:**
```go
type Pattern struct {
    Name        string
    Description string
    Template    string
    DefaultFile string // e.g., "pages/example.go"
}

var patterns = map[string]Pattern{
    "page":      {Name: "page", Description: "Basic page template", ...},
    "page:list": {Name: "page:list", Description: "List page with API call", ...},
    // ...
}
```

### Anti-Patterns to Avoid
- **Separate template files for help patterns:** Unlike scaffold templates, help patterns are small and don't change per-project. Use string constants for simplicity.
- **Complex subcommand frameworks:** The existing flag.NewFlagSet pattern is sufficient. Don't add cobra/urfave/cli.
- **Hardcoded module paths:** Always use modfile.ModulePath() or show placeholder with warning.

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| go.mod parsing | Regex extraction | modfile.ModulePath() | Handles edge cases (comments, whitespace, encoding) |
| Template rendering | String concatenation | text/template | Already used, consistent, handles escaping |

**Key insight:** The `modfile.ModulePath()` function is specifically designed to be tolerant of malformed go.mod files, making it more robust than regex parsing.

## Common Pitfalls

### Pitfall 1: Failing on Missing go.mod
**What goes wrong:** Error and exit when go.mod doesn't exist
**Why it happens:** Running `gux help` outside a project
**How to avoid:** Fall back to placeholder with warning message
**Warning signs:** User runs `gux help page` in home directory

```go
// Correct handling:
modulePath, err := getModulePath()
if err != nil {
    modulePath = "your-module-path"
    fmt.Fprintln(os.Stderr, "Warning: Could not read go.mod. Using placeholder import path.")
    fmt.Fprintln(os.Stderr, "Run from a Go module directory or replace 'your-module-path' manually.")
    fmt.Fprintln(os.Stderr, "")
}
```

### Pitfall 2: Template Variable Collision
**What goes wrong:** `{{.ModulePath}}` in help output gets interpreted
**Why it happens:** Output contains template syntax that text/template processes
**How to avoid:** Use `{{"{{"}}` escaping or define as separate constant after template rendering
**Warning signs:** Missing `.ModulePath` in output

### Pitfall 3: Incorrect Package Names in Examples
**What goes wrong:** Help output says `package main` for page code
**Why it happens:** Copy-paste from app.go patterns
**How to avoid:** Each pattern has correct package (pages, models, dto)
**Warning signs:** Code doesn't compile when pasted

### Pitfall 4: Output Without File Path Header
**What goes wrong:** User doesn't know where to save the code
**Why it happens:** Forgot to include file path comment
**How to avoid:** Every pattern template starts with `// {filepath}` comment
**Warning signs:** Requirement OUT-01 not met

### Pitfall 5: Mixing tabs and spaces
**What goes wrong:** Inconsistent indentation in output
**Why it happens:** Template literal uses different indentation than project
**How to avoid:** Use tabs consistently (Go gofmt standard)
**Warning signs:** `gofmt -d` shows changes

## Code Examples

Verified patterns from existing scaffold templates (Phase 27):

### Basic Page Pattern (from templates/pages/home.go.tmpl)
```go
// pages/example.go
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
								core.P(core.Class("text-gray-600 dark:text-gray-400"),
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
```

### List Page Pattern (from templates/pages/items.go.tmpl)
```go
// pages/items.go
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
		// State for hydration
		itemsJSON, _ := json.Marshal(items)
		itemsState := r.StateString("items", string(itemsJSON))

		var currentItems []api.Item
		if itemsState.Get() != "" {
			json.Unmarshal([]byte(itemsState.Get()), &currentItems)
		}

		return core.Div(core.Class("min-h-screen bg-gray-100 dark:bg-gray-900"),
			// ... table rendering
		)
	}
}
```

### Form Page Pattern (from templates/pages/item_new.go.tmpl)
```go
// pages/item_new.go
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
		errorState := r.StateString("error", "")

		save := func() {
			if nameState.Get() == "" {
				errorState.Set("Name is required")
				return
			}
			data := map[string]interface{}{
				"name": nameState.Get(),
			}
			api.Items.Create(data, func(result *api.Item, err error) {
				if err != nil {
					errorState.Set(err.Error())
				} else {
					r.Navigate("/items")
				}
			})
		}

		// ... form UI
	}
}
```

### GORM Model Pattern (from templates/models/item.go.tmpl)
```go
// models/item.go
package models

import "gorm.io/gorm"

// Item is an example model.
type Item struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
}
```

### DTO Pattern (from examples/minimal/dto/user.go)
```go
// dto/item.go
package dto

import "time"

// ItemList is the DTO for list responses.
type ItemList struct {
	ID        uint      `json:"id" gux:"Item.ID"`
	Name      string    `json:"name" gux:"Item.Name"`
	CreatedAt time.Time `json:"created_at" gux:"Item.CreatedAt"`
}

// ItemDetail is the DTO for single item responses.
type ItemDetail struct {
	ID          uint      `json:"id" gux:"Item.ID"`
	Name        string    `json:"name" gux:"Item.Name"`
	Description string    `json:"description" gux:"Item.Description"`
	CreatedAt   time.Time `json:"created_at" gux:"Item.CreatedAt"`
	UpdatedAt   time.Time `json:"updated_at" gux:"Item.UpdatedAt"`
}
```

### Routes Pattern (from templates/app.go.tmpl)
```go
// In app.go, after app := core.New()

// Register routes
app.Routes().
	Hybrid("/", pages.Home).
	Hybrid("/items", pages.Items).
	Hybrid("/items/new", pages.ItemNew).
	Hybrid("/items/:id", pages.ItemDetail)

// With route groups
app.RouteGroup("/admin", core.WithBundle("admin")).
	Hybrid("/", admin.Dashboard).
	Hybrid("/users", admin.Users)
```

### App Pattern (from templates/app.go.tmpl)
```go
// app.go
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
	app.SetTitle("My App")

	// Register CRUD
	app.CRUD(models.Item{})

	// Register routes
	app.Routes().
		Hybrid("/", pages.Home).
		Hybrid("/items", pages.Items)

	app.Run(":8080")
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| No help patterns | gux help <pattern> | Phase 28 | Quick reference for developers |
| Manual import path editing | Auto-detect from go.mod | Phase 28 | Copy-paste ready code |

**Current limitations:**
- Help patterns are static (no customization parameters like model name)
- Single pattern output (no batch generation)

These are acceptable for v1 help system.

## Open Questions

All technical questions resolved:

1. **Pattern naming convention**: Use `pattern` for simple, `pattern:variant` for specialized (e.g., `page`, `page:list`, `page:form`)
2. **go.mod fallback**: Use `your-module-path` placeholder with stderr warning
3. **Template storage**: String constants in help.go (not embed.FS)

## Sources

### Primary (HIGH confidence)
- `/Users/dougbarrett/projects/dbb1dev/goquery/cmd/gux/scaffold.go` - Template rendering pattern
- `/Users/dougbarrett/projects/dbb1dev/goquery/cmd/gux/main.go` - CLI structure pattern
- `/Users/dougbarrett/projects/dbb1dev/goquery/cmd/gux/templates/*.tmpl` - Existing templates (source of truth for patterns)
- [golang.org/x/mod/modfile](https://pkg.go.dev/golang.org/x/mod/modfile) - Official go.mod parsing
- `/Users/dougbarrett/projects/dbb1dev/goquery/examples/minimal/dto/` - DTO patterns

### Secondary (MEDIUM confidence)
- `/Users/dougbarrett/projects/dbb1dev/goquery/.planning/phases/27-gux-init-templates/27-RESEARCH.md` - Phase 27 context

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Uses existing project patterns + official Go package
- Architecture: HIGH - Simple addition to existing CLI structure
- Pitfalls: HIGH - Based on actual template work from Phase 27

**Research date:** 2026-01-26
**Valid until:** 60 days (stable internal patterns)
