---
phase: 27-gux-init-templates
verified: 2026-01-26T17:20:00Z
status: passed
score: 10/10 must-haves verified
---

# Phase 27: gux init Templates Verification Report

**Phase Goal:** Update `gux init` to scaffold working apps using the modern `core/` framework instead of the deleted `components/` package.

**Verified:** 2026-01-26T17:20:00Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | app.go.tmpl renders valid Go code with core.New(), SetDB(), CRUD(), Routes(), Run() | ✓ VERIFIED | Lines 27-42: All five methods present and properly sequenced |
| 2 | go.mod.tmpl includes gux, gorm, and sqlite dependencies | ✓ VERIFIED | Lines 5-8: {{.GuxModule}}, gorm.io/gorm v1.31.1, gorm.io/driver/sqlite v1.6.0 |
| 3 | models/item.go.tmpl defines GORM model with ID, Name, Description, timestamps | ✓ VERIFIED | Lines 6-9: gorm.Model (provides ID, timestamps), Name, Description fields |
| 4 | home.go.tmpl renders landing page with ui.Card and link to /items | ✓ VERIFIED | Lines 13-38: ui.Card, ui.CardHeader, ui.CardContent, link href="/items" |
| 5 | items.go.tmpl loads items via OnLoad and displays them in ui.Table | ✓ VERIFIED | Lines 14-22: r.OnLoad with api.Items.List; Lines 53-70: ui.Table rendering |
| 6 | item_new.go.tmpl provides form with state, validation, and API submission | ✓ VERIFIED | Lines 13-15: StateString for name/desc/error; Lines 19-23: validation; Lines 26-38: api.Items.Create |
| 7 | gux init creates all new template files (app.go, models/item.go, pages/*.go) | ✓ VERIFIED | scaffold.go lines 104-115: filesToCreate includes all 7 files |
| 8 | gux init runs gux gen after file creation | ✓ VERIFIED | scaffold.go lines 140-151: exec.Command("gux", "gen") after go mod tidy |
| 9 | gux init runs go mod tidy to download dependencies | ✓ VERIFIED | scaffold.go lines 127-138: exec.Command("go", "mod", "tidy") |
| 10 | Old templates (cmd/app, cmd/server, internal/api) are removed | ✓ VERIFIED | cmd/gux/templates/cmd/ and internal/ directories do not exist |

**Score:** 10/10 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `cmd/gux/templates/app.go.tmpl` | Main app template using core/ framework | ✓ VERIFIED | 43 lines, contains core.New(), SetDB(), CRUD(), Routes(), Run() |
| `cmd/gux/templates/go.mod.tmpl` | Module file template with dependencies | ✓ VERIFIED | 9 lines, includes gux, gorm, sqlite with proper version placeholders |
| `cmd/gux/templates/models/item.go.tmpl` | Example GORM model template | ✓ VERIFIED | 10 lines, embeds gorm.Model, has Name/Description with JSON tags |
| `cmd/gux/templates/pages/home.go.tmpl` | Landing page template | ✓ VERIFIED | 43 lines, uses ui.Card, ui.CardHeader, ui.CardContent, link to /items |
| `cmd/gux/templates/pages/items.go.tmpl` | Items list template with data loading | ✓ VERIFIED | 117 lines, has OnLoad, api.Items.List, ui.Table, empty state handling |
| `cmd/gux/templates/pages/item_new.go.tmpl` | Item creation form template | ✓ VERIFIED | 127 lines, has StateString, validation, api.Items.Create, error handling |
| `cmd/gux/scaffold.go` | Updated scaffold with new file list and gux gen call | ✓ VERIFIED | Modified, references all new templates, calls gux gen after go mod tidy |

**Artifact Status Summary:** 7/7 artifacts exist, substantive, and wired

**Level 2 (Substantive) Details:**
- app.go.tmpl: 43 lines, no stubs, exports main()
- go.mod.tmpl: 9 lines, no stubs, proper module definition
- models/item.go.tmpl: 10 lines, no stubs, exports Item struct
- home.go.tmpl: 43 lines, no stubs, exports Home function
- items.go.tmpl: 117 lines, no stubs, exports Items function
- item_new.go.tmpl: 127 lines, 2 "placeholder" occurrences are user-facing input placeholders (not code stubs)

**Level 3 (Wired) Details:**
- All templates imported by scaffold.go via filesToCreate slice (lines 104-115)
- app.go.tmpl imports models package and registers CRUD: app.CRUD(models.Item{})
- items.go.tmpl imports {{.ModulePath}}/.gux/api and calls api.Items.List
- item_new.go.tmpl imports {{.ModulePath}}/.gux/api and calls api.Items.Create

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| app.go.tmpl | models/item.go | import and CRUD registration | ✓ WIRED | Line 8: import "{{.ModulePath}}/models"; Line 34: app.CRUD(models.Item{}) |
| items.go.tmpl | api.Items | OnLoad API call | ✓ WIRED | Line 9: import api; Line 17: api.Items.List(callback) |
| item_new.go.tmpl | api.Items | Create API call | ✓ WIRED | Line 6: import api; Line 30: api.Items.Create(data, callback) |
| scaffold.go | templates/*.tmpl | filesToCreate slice | ✓ WIRED | Lines 104-115: All 7 templates referenced with correct paths |
| app.go.tmpl | .gux/api | api.Init() call | ✓ WIRED | Line 7: import api; Line 25: api.Init(db) |

**All key links verified and operational.**

### Requirements Coverage

No explicit REQUIREMENTS.md entries for Phase 27. Phase requirements derived from v2.2 milestone goal.

### Anti-Patterns Found

None detected.

**Scan Results:**
- No TODO/FIXME/XXX comments found
- No "placeholder" or "coming soon" in code (only in user-facing input placeholder text)
- No empty implementations (return null, return {}, console.log-only handlers)
- All handlers have substantive implementations

### Human Verification Required

None. All verification completed programmatically.

### Gaps Summary

No gaps found. Phase 27 goal fully achieved.

---

**Detailed Evidence:**

## Truth 1: app.go.tmpl has core.New(), SetDB(), CRUD(), Routes(), Run()

```go
// Lines 27-42 of app.go.tmpl
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
```

All five required methods present: core.New(), SetDB(), CRUD(), Routes(), Run()

## Truth 2: go.mod.tmpl dependencies

```
module {{.ModulePath}}

go 1.24

require (
    {{.GuxModule}} {{.GuxVersion}}
    gorm.io/driver/sqlite v1.6.0
    gorm.io/gorm v1.31.1
)
```

Includes gux (with placeholder), gorm v1.31.1, sqlite driver v1.6.0

## Truth 3: models/item.go.tmpl GORM model

```go
type Item struct {
    gorm.Model
    Name        string `json:"name"`
    Description string `json:"description"`
}
```

Embeds gorm.Model (provides ID, CreatedAt, UpdatedAt, DeletedAt), has Name and Description fields with JSON tags

## Truth 4: home.go.tmpl ui.Card and link

```go
ui.Card(ui.CardProps{
    Children: []core.Node{
        ui.CardHeader(ui.CardHeaderProps{
            Children: []core.Node{
                core.H1(core.Class("text-3xl font-bold text-gray-900 dark:text-white"),
                    core.Text("Welcome to {{.AppName}}"),
                ),
            },
        }),
        ui.CardContent(ui.CardContentProps{
            Children: []core.Node{
                // ...
                core.A(
                    core.Attrs{
                        Href:  "/items",
                        Class: "inline-block px-6 py-3 bg-blue-600 hover:bg-blue-700...",
                    },
                    core.Text("View Items"),
                ),
            },
        }),
    },
}),
```

Uses ui.Card compound components, link to /items present

## Truth 5: items.go.tmpl OnLoad and ui.Table

```go
var items []api.Item
r.OnLoad(func() {
    api.Items.List(func(result []api.Item, err error) {
        if err == nil {
            items = result
        }
    })
})

// Later in component:
ui.Table(ui.TableProps{
    Children: []core.Node{
        ui.Thead(ui.TheadProps{ /* ... */ }),
        ui.Tbody(ui.TbodyProps{
            Children: itemRows(currentItems),
        }),
    },
}),
```

OnLoad fetches from api.Items.List, renders with ui.Table

## Truth 6: item_new.go.tmpl state, validation, API submission

```go
// State
nameState := r.StateString("name", "")
descState := r.StateString("description", "")
errorState := r.StateString("error", "")

// Validation
save := func() {
    if nameState.Get() == "" {
        errorState.Set("Name is required")
        return
    }

    // API submission
    data := map[string]interface{}{
        "name":        nameState.Get(),
        "description": descState.Get(),
    }
    api.Items.Create(data, func(result *api.Item, err error) {
        if err != nil {
            errorState.Set(err.Error())
        } else {
            errorState.Set("")
            r.Navigate("/items")
        }
    })
}
```

Has StateString for three fields, validates name is required, submits via api.Items.Create

## Truth 7: scaffold.go creates all files

```go
// Lines 104-115 of scaffold.go
filesToCreate := []struct {
    tmplPath string
    destPath string
}{
    {"templates/go.mod.tmpl", "go.mod"},
    {"templates/app.go.tmpl", "app.go"},
    {"templates/models/item.go.tmpl", "models/item.go"},
    {"templates/pages/home.go.tmpl", "pages/home.go"},
    {"templates/pages/items.go.tmpl", "pages/items.go"},
    {"templates/pages/item_new.go.tmpl", "pages/item_new.go"},
    {"templates/Dockerfile.tmpl", "Dockerfile"},
}
```

All 7 files defined in filesToCreate slice

## Truth 8: scaffold.go runs gux gen

```go
// Lines 140-151 of scaffold.go
// Run gux gen to generate API client
fmt.Println("\nRunning gux gen...")
genCmd := exec.Command("gux", "gen")
genCmd.Dir = targetDir
genCmd.Stdout = os.Stdout
genCmd.Stderr = os.Stderr
if err := genCmd.Run(); err != nil {
    fmt.Printf("Warning: gux gen failed: %v\n", err)
    fmt.Println("You may need to run 'gux gen' manually.")
} else {
    fmt.Println("  API client generated")
}
```

gux gen called automatically after go mod tidy, with graceful failure handling

## Truth 9: scaffold.go runs go mod tidy

```go
// Lines 127-138 of scaffold.go
fmt.Println("\nRunning go mod tidy...")
cmd := exec.Command("go", "mod", "tidy")
cmd.Dir = targetDir
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
if err := cmd.Run(); err != nil {
    fmt.Printf("Warning: go mod tidy failed: %v\n", err)
    fmt.Println("You may need to run 'go mod tidy' manually.")
} else {
    fmt.Println("  dependencies downloaded")
}
```

go mod tidy called before gux gen

## Truth 10: Old templates removed

```bash
$ ls cmd/gux/templates/cmd 2>&1
ls: /Users/dougbarrett/projects/dbb1dev/goquery/cmd/gux/templates/cmd: No such file or directory

$ ls cmd/gux/templates/internal 2>&1
ls: /Users/dougbarrett/projects/dbb1dev/goquery/cmd/gux/templates/internal: No such file or directory

$ ls cmd/gux/templates/
app.go.tmpl
claude
Dockerfile.tmpl
go.mod.tmpl
models
pages
```

Old cmd/ and internal/ template directories do not exist, only new templates present

---

_Verified: 2026-01-26T17:20:00Z_
_Verifier: Claude (gsd-verifier)_
