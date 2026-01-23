# Phase 22: SaaS Dashboard Example - Research

**Researched:** 2026-01-23
**Domain:** SaaS Dashboard UI patterns with Go WASM
**Confidence:** HIGH

## Summary

Phase 22 builds a SaaS dashboard example demonstrating authenticated dashboard patterns, CRUD operations, settings/profile management, and data visualization. This research examined the existing codebase extensively to identify available components, established patterns, and gaps.

The Gux framework already has comprehensive UI components (`ui/` package) and proven patterns from `examples/minimal/admin/` that directly apply to dashboard development. The primary work is composing existing components into dashboard-specific layouts and pages rather than building new components.

**Primary recommendation:** Follow the `examples/minimal/admin/` patterns exactly - use horizontal nav (not sidebar), existing Card/Grid/DataTable components, and established CRUD API patterns. Skip charts for v2.0 MVP; use stat cards instead.

## Key Findings

### 1. Dashboard Layout: Use Horizontal Nav (Not Sidebar)

**Finding:** The existing `examples/minimal/admin/` uses a horizontal navigation pattern, not a sidebar.

**Evidence from `admin/nav.go`:**
```go
func Nav() core.Node {
    return core.Div(core.Class("bg-gray-800 border-b border-gray-700"),
        core.Div(core.Class("max-w-6xl mx-auto px-4"),
            core.Div(core.Class("flex justify-between items-center py-4"),
                // Logo + nav links horizontally
                core.Div(core.Class("flex items-center gap-6"),
                    core.Span(core.Class("text-xl font-bold text-white"), core.Text("Admin Panel")),
                    core.A(core.Attrs{Href: "/admin", ...}, core.Text("Dashboard")),
                    // ... more links
                ),
            ),
        ),
    )
}
```

**Implication:** The SaaS Dashboard should follow this same horizontal nav pattern for consistency. A sidebar component exists in `components/sidebar.go` but it's part of the old component library (WASM-only, `//go:build js && wasm`) and not compatible with the SSR+hydration pattern.

**Recommendation:** Use horizontal nav like `examples/minimal/admin/`. This keeps the example simpler and consistent with existing patterns.

### 2. Available UI Components (Comprehensive)

**HIGH Confidence - Direct from `ui/` package source code:**

| Component | File | Key Props | Use In Dashboard |
|-----------|------|-----------|------------------|
| `Card`, `CardHeader`, `CardContent`, `CardFooter` | `card.go` | `Class`, `Children` | Stat cards, content sections |
| `Container` | `layout.go` | `MaxWidth` ("sm"/"md"/"lg"/"xl"/"2xl"/"7xl") | Page width constraint |
| `VStack`, `HStack` | `layout.go` | `Gap`, `Align`, `Justify` | Layout composition |
| `Grid` | `layout.go` | `Cols` ("1"/"2"/"3"/"4"/"6"/"12"), `Gap` | Stat cards grid, form layouts |
| `DataTable[T]` | `datatable.go` | `Data`, `Columns`, `OnRowClick`, `Striped`, `Hoverable` | Resource list |
| `Table`, `Thead`, `Tbody`, `Tr`, `Th`, `Td` | `table.go` | Standard props | Custom tables |
| `FormField` | `form.go` | `Label`, `Required`, `Error`, `Description` | All form inputs |
| `Input` | `input.go` | `Type`, `Size`, `OnChange`, `OnEnter`, `Error` | Text inputs |
| `Textarea` | `textarea.go` | `Rows`, `OnChange` | Long text |
| `Select` | `select.go` | `Options`, `OnChange` | Dropdowns |
| `Checkbox`, `Radio` | `checkbox.go`, `radio.go` | `Checked`, `Label` | Boolean/choice fields |
| `Switch` | `switch.go` | `Checked`, `Label` | Toggle settings |
| `Button` | `button.go` | `Variant`, `Size`, `OnClick`, `Disabled` | Actions |
| `Badge` | `badge.go` | `Text`, `Variant` (default/success/warning/error) | Status indicators |
| `Avatar` | `avatar.go` | `Src`, `Name`, `Size` | Profile display |
| `Alert` | `alert.go` | `Message`, `Variant`, `Title`, `OnClose` | Feedback messages |
| `Modal`, `ModalContent`, `ModalFooter` | `modal.go` | `Open`, `OnClose`, `Title`, `Size` | Confirmation dialogs |
| `Tabs`, `TabList`, `Tab`, `TabPanel` | `tabs.go` | `Active`, `OnSelect` | Settings sections |
| `Pagination` | `pagination.go` | `CurrentPage`, `TotalItems`, `PageSize`, `OnPageChange` | Table pagination |

### 3. Charts: Skip for v2.0, Use Stat Cards

**Finding:** Charts exist in `components/charts.go` but are WASM-only (incompatible with SSR):
```go
//go:build js && wasm
package components
```

**Recommendation:** For the SaaS Dashboard MVP, use stat cards (Card + number display) instead of charts. This matches what `examples/minimal/admin/dashboard.go` already does:

```go
// Stats cards pattern from admin/dashboard.go
core.Div(core.Class("grid grid-cols-1 md:grid-cols-3 gap-6"),
    // Users card
    core.Div(core.Class("bg-gray-800 rounded-lg p-6"),
        core.Div(core.Class("text-gray-400 text-sm"), core.Text("Total Users")),
        core.Div(core.Class("text-3xl font-bold text-white mt-2"), core.Text(formatInt(users.Get()))),
    ),
    // ... more cards
)
```

**Future consideration:** If charts are needed, build pure SVG chart components using `core.Node` (like `ui/` components) or investigate server-side SVG generation.

### 4. CRUD Integration Pattern

**HIGH Confidence - Directly from existing code:**

**Data Loading Pattern (`admin/users.go`):**
```go
func Users(r *core.Router) func() core.Node {
    var users []dto.UserList

    r.OnLoad(func() {
        api.Users.List(func(result []dto.UserList, err error) {
            if err == nil {
                users = result
            }
        })
    })

    return func() core.Node {
        // Serialize for hydration
        usersJSON, _ := json.Marshal(users)
        usersState := r.StateString("users", string(usersJSON))

        // Deserialize for rendering
        var displayUsers []dto.UserList
        json.Unmarshal([]byte(usersState.Get()), &displayUsers)
        // ... render
    }
}
```

**Form Pattern (`admin/user_new.go`):**
```go
func UserNew(r *core.Router) func() core.Node {
    return func() core.Node {
        nameState := r.StateString("name", "")
        emailState := r.StateString("email", "")
        errorState := r.StateString("error", "")
        successState := r.StateString("success", "")

        save := func() {
            data := map[string]interface{}{
                "name": nameState.Get(),
                "email": emailState.Get(),
            }
            api.Users.Create(data, func(result *dto.UserDetail, err error) {
                if err != nil {
                    errorState.Set(err.Error())
                } else {
                    r.Navigate("/admin/users")
                }
            })
        }
        // ... render form
    }
}
```

**API Client Pattern (generated):**
```go
// List, Get, Create, Update, Delete operations
api.Users.List(func(result []dto.UserList, err error) { ... })
api.Users.Get(id, func(result *dto.UserDetail, err error) { ... })
api.Users.Create(data, func(result *dto.UserDetail, err error) { ... })
api.Users.Update(id, data, func(result *dto.UserDetail, err error) { ... })
api.Users.Delete(id, func(err error) { ... })
```

### 5. Settings Page Pattern

**Finding:** The `admin/account.go` provides a settings/profile page pattern:

```go
func Account(r *core.Router) func() core.Node {
    return func() core.Node {
        return core.Div(core.Class("min-h-screen bg-gray-900"),
            Nav(),
            core.Div(core.Class("max-w-2xl mx-auto px-4 py-8"),
                // Section with info display
                core.Div(core.Class("bg-gray-800 rounded-lg p-6 space-y-6"),
                    // Field groups
                    core.Div(core.Attrs{},
                        core.Label(...),
                        core.Div(core.Class("text-white text-lg"), ...),
                    ),
                    // Actions
                    core.Div(core.Class("pt-4 border-t border-gray-700"),
                        core.Button(...),
                    ),
                ),
                // Danger zone
                core.Div(core.Class("mt-8 bg-red-900/20 border border-red-800 rounded-lg p-6"),
                    ...
                ),
            ),
        )
    }
}
```

**For Settings with Tabs (recommended):**
```go
activeTab := r.StateInt("activeTab", 0)
ui.Tabs(ui.TabsProps{
    Children: []core.Node{
        ui.TabList(ui.TabListProps{
            Children: []core.Node{
                ui.Tab(ui.TabProps{Active: activeTab.Get() == 0, OnSelect: func() { activeTab.Set(0) }, Children: []core.Node{core.Text("General")}}),
                ui.Tab(ui.TabProps{Active: activeTab.Get() == 1, OnSelect: func() { activeTab.Set(1) }, Children: []core.Node{core.Text("Security")}}),
            },
        }),
        ui.TabPanel(ui.TabPanelProps{Active: activeTab.Get() == 0, Children: generalSettings}),
        ui.TabPanel(ui.TabPanelProps{Active: activeTab.Get() == 1, Children: securitySettings}),
    },
})
```

### 6. Authentication State Handling

**Finding:** The `examples/auth/` example does not implement actual auth state checking - it's focused on demonstrating auth UI flows. The dashboard page is a simple success view.

**Recommendation for SaaS Dashboard:** Follow the same pattern - this is a UI demonstration, not a full auth implementation. The dashboard can assume the user is "logged in" and display mock user data. For a real app, auth middleware and token checking would be added at the framework level.

**Pattern from auth dashboard:**
```go
func Dashboard(r *core.Router) func() core.Node {
    return func() core.Node {
        return PageLayout(
            // Content assumes user is authenticated
            core.Text("Welcome back!"),
            ...
        )
    }
}
```

## Architecture Patterns

### Recommended Project Structure

```
examples/saas/
├── app.go                 # Route setup, CRUD registration
├── models/
│   └── project.go         # Resource model
├── dto/
│   └── project.go         # DTOs for list/detail
├── pages/
│   ├── layout.go          # DashboardLayout, Nav
│   ├── dashboard.go       # Dashboard overview with stats
│   ├── projects.go        # Resource list (DataTable)
│   ├── project_detail.go  # Single resource view
│   ├── project_new.go     # Create form
│   ├── project_edit.go    # Edit form
│   ├── settings.go        # Settings with tabs
│   └── profile.go         # Profile view/edit
└── .gux/                   # Generated (api/, wasm/)
```

### Layout Pattern: DashboardLayout

```go
func DashboardLayout(children ...core.Node) core.Node {
    return core.Div(core.Class("min-h-screen bg-gray-900"),
        Nav(),
        core.Main(core.Class("max-w-6xl mx-auto px-4 py-8"),
            children...,
        ),
    )
}

func Nav() core.Node {
    return core.Div(core.Class("bg-gray-800 border-b border-gray-700"),
        core.Div(core.Class("max-w-6xl mx-auto px-4"),
            core.Div(core.Class("flex justify-between items-center py-4"),
                // Brand
                core.A(core.Attrs{Href: "/", Class: "text-xl font-bold text-white"},
                    core.Text("SaaS App"),
                ),
                // Nav links
                core.Div(core.Class("flex items-center gap-6"),
                    core.A(core.Attrs{Href: "/dashboard", Class: "text-gray-300 hover:text-white"}, core.Text("Dashboard")),
                    core.A(core.Attrs{Href: "/projects", Class: "text-gray-300 hover:text-white"}, core.Text("Projects")),
                    core.A(core.Attrs{Href: "/settings", Class: "text-gray-300 hover:text-white"}, core.Text("Settings")),
                ),
                // User menu (Avatar + name)
                core.Div(core.Class("flex items-center gap-2"),
                    ui.Avatar(ui.AvatarProps{Name: "John Doe", Size: ui.AvatarSM}),
                    core.A(core.Attrs{Href: "/profile", Class: "text-gray-300 hover:text-white"}, core.Text("Profile")),
                ),
            ),
        ),
    )
}
```

### Stats Card Pattern

```go
func StatCard(title string, value string, color string) core.Node {
    valueClass := "text-3xl font-bold mt-2 " + color
    return ui.Card(ui.CardProps{
        Class: "bg-gray-800 border-gray-700",
        Children: []core.Node{
            ui.CardContent(ui.CardContentProps{
                Children: []core.Node{
                    core.Div(core.Class("text-gray-400 text-sm"), core.Text(title)),
                    core.Div(core.Class(valueClass), core.Text(value)),
                },
            }),
        },
    })
}
```

### DataTable Integration Pattern

```go
ui.DataTable(ui.DataTableProps[dto.ProjectList]{
    Data: projects,
    Striped: true,
    Hoverable: true,
    OnRowClick: func(p dto.ProjectList) {
        r.Navigate(fmt.Sprintf("/projects/%d", p.ID))
    },
    Columns: []ui.ColumnDef[dto.ProjectList]{
        {Header: "Name", Render: func(p dto.ProjectList) core.Node {
            return core.A(core.Attrs{
                Href: fmt.Sprintf("/projects/%d", p.ID),
                Class: "text-blue-400 hover:text-blue-300",
            }, core.Text(p.Name))
        }},
        {Header: "Status", Render: func(p dto.ProjectList) core.Node {
            return ui.Badge(ui.BadgeProps{
                Text: p.Status,
                Variant: statusVariant(p.Status),
            })
        }},
        {Header: "Created", Render: func(p dto.ProjectList) core.Node {
            return core.Text(p.CreatedAt.Format("Jan 2, 2006"))
        }},
    },
})
```

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Data tables | Custom table rendering | `ui.DataTable[T]` | Type-safe, handles empty state, row clicks |
| Form layout | Manual label/error/input | `ui.FormField` | Consistent styling, accessibility |
| Stat cards | Raw div structure | `ui.Card` + pattern | Maintainable, consistent |
| Navigation state | Custom URL parsing | `r.Navigate()`, route params | Framework handles hydration |
| Data hydration | localStorage/custom | `r.StateString` + JSON | SSR-compatible |
| Modals | Custom overlay | `ui.Modal` | Focus trap, accessibility |
| Pagination | Custom page logic | `ui.Pagination` | Edge cases handled |

## Common Pitfalls

### Pitfall 1: Forgetting JSON Marshaling for Hydration
**What goes wrong:** Data loaded in `OnLoad` isn't available after WASM hydration
**Why it happens:** `OnLoad` runs on server, state must be serialized to transfer to client
**How to avoid:** Always marshal data to state string in render function:
```go
r.OnLoad(func() { items = fetchData() })
return func() core.Node {
    itemsJSON, _ := json.Marshal(items)
    itemsState := r.StateString("items", string(itemsJSON))
    json.Unmarshal([]byte(itemsState.Get()), &displayItems)
}
```
**Warning signs:** Page shows data on server render but blank after WASM loads

### Pitfall 2: Using Old Components Library
**What goes wrong:** Build errors or SSR failures
**Why it happens:** `components/` package uses `//go:build js && wasm` (WASM-only)
**How to avoid:** Use only `ui/` package components for SSR+hydration apps
**Warning signs:** Import from `components/` fails to compile for server

### Pitfall 3: Direct API Calls Without Callbacks
**What goes wrong:** Race conditions, UI doesn't update
**Why it happens:** WASM API calls are async via goroutines
**How to avoid:** Always use callback pattern and update state in callback:
```go
api.Projects.Delete(id, func(err error) {
    if err != nil { errorState.Set(err.Error()) }
    else { r.Navigate("/projects") }
})
```

### Pitfall 4: Select Elements Without Controlled Value
**What goes wrong:** Select doesn't reflect state after edit/create
**Why it happens:** Need to set `selected` attribute on matching option
**How to avoid:** Use the `selectOption` helper pattern:
```go
func selectOption(value, label, selectedValue string) core.Node {
    attrs := core.Attrs{Value: value}
    if value == selectedValue {
        attrs.Extra = map[string]string{"selected": "selected"}
    }
    return core.Option(attrs, core.Text(label))
}
```

### Pitfall 5: Missing Error State Handling
**What goes wrong:** Users don't see why operations failed
**Why it happens:** Error callbacks aren't wired to UI
**How to avoid:** Always include error/success state:
```go
errorState := r.StateString("error", "")
successState := r.StateString("success", "")
// Wire to UI and clear on new operation
```

## Code Examples

### Dashboard Overview Page
```go
// Source: Adapted from examples/minimal/admin/dashboard.go
func Dashboard(r *core.Router) func() core.Node {
    var stats DashboardStats

    r.OnLoad(func() {
        // Fetch stats
        stats = DashboardStats{
            TotalProjects: 42,
            ActiveUsers: 156,
            Revenue: "$12,345",
        }
    })

    return func() core.Node {
        projectCount := r.StateInt("projects", stats.TotalProjects)
        userCount := r.StateInt("users", stats.ActiveUsers)
        revenue := r.StateString("revenue", stats.Revenue)

        return DashboardLayout(
            core.H1(core.Class("text-3xl font-bold text-white mb-8"),
                core.Text("Dashboard"),
            ),
            // Stats grid
            ui.Grid(ui.GridProps{
                Cols: "3",
                Gap: "6",
                Class: "mb-8",
                Children: []core.Node{
                    StatCard("Total Projects", fmt.Sprintf("%d", projectCount.Get()), "text-white"),
                    StatCard("Active Users", fmt.Sprintf("%d", userCount.Get()), "text-white"),
                    StatCard("Revenue", revenue.Get(), "text-green-400"),
                },
            }),
            // Recent activity section
            ui.Card(ui.CardProps{
                Class: "bg-gray-800 border-gray-700",
                Children: []core.Node{
                    ui.CardHeader(ui.CardHeaderProps{
                        Children: []core.Node{
                            core.H2(core.Class("text-xl font-semibold text-white"),
                                core.Text("Recent Activity"),
                            ),
                        },
                    }),
                    ui.CardContent(ui.CardContentProps{
                        Children: []core.Node{
                            // Activity items...
                        },
                    }),
                },
            }),
        )
    }
}
```

### Settings Page with Tabs
```go
// Source: Pattern from ui/tabs.go + admin/account.go
func Settings(r *core.Router) func() core.Node {
    return func() core.Node {
        activeTab := r.StateInt("activeTab", 0)

        return DashboardLayout(
            core.H1(core.Class("text-3xl font-bold text-white mb-8"),
                core.Text("Settings"),
            ),
            ui.Card(ui.CardProps{
                Class: "bg-gray-800 border-gray-700",
                Children: []core.Node{
                    ui.CardContent(ui.CardContentProps{
                        Children: []core.Node{
                            ui.Tabs(ui.TabsProps{
                                Children: []core.Node{
                                    ui.TabList(ui.TabListProps{
                                        Children: []core.Node{
                                            ui.Tab(ui.TabProps{
                                                Active: activeTab.Get() == 0,
                                                OnSelect: func() { activeTab.Set(0) },
                                                Children: []core.Node{core.Text("General")},
                                            }),
                                            ui.Tab(ui.TabProps{
                                                Active: activeTab.Get() == 1,
                                                OnSelect: func() { activeTab.Set(1) },
                                                Children: []core.Node{core.Text("Notifications")},
                                            }),
                                            ui.Tab(ui.TabProps{
                                                Active: activeTab.Get() == 2,
                                                OnSelect: func() { activeTab.Set(2) },
                                                Children: []core.Node{core.Text("Security")},
                                            }),
                                        },
                                    }),
                                    ui.TabPanel(ui.TabPanelProps{
                                        Active: activeTab.Get() == 0,
                                        Children: []core.Node{GeneralSettingsForm(r)},
                                    }),
                                    ui.TabPanel(ui.TabPanelProps{
                                        Active: activeTab.Get() == 1,
                                        Children: []core.Node{NotificationSettingsForm(r)},
                                    }),
                                    ui.TabPanel(ui.TabPanelProps{
                                        Active: activeTab.Get() == 2,
                                        Children: []core.Node{SecuritySettingsForm(r)},
                                    }),
                                },
                            }),
                        },
                    }),
                },
            }),
        )
    }
}
```

### Profile Page with Avatar
```go
// Source: Pattern from ui/avatar.go + admin/account.go
func Profile(r *core.Router) func() core.Node {
    return func() core.Node {
        name := r.StateString("name", "John Doe")
        email := r.StateString("email", "john@example.com")

        return DashboardLayout(
            core.H1(core.Class("text-3xl font-bold text-white mb-8"),
                core.Text("Profile"),
            ),
            ui.Card(ui.CardProps{
                Class: "bg-gray-800 border-gray-700",
                Children: []core.Node{
                    ui.CardContent(ui.CardContentProps{
                        Children: []core.Node{
                            // Avatar + name header
                            core.Div(core.Class("flex items-center gap-4 mb-6"),
                                ui.Avatar(ui.AvatarProps{
                                    Name: name.Get(),
                                    Size: ui.AvatarLG,
                                }),
                                core.Div(core.Attrs{},
                                    core.H2(core.Class("text-xl font-bold text-white"),
                                        core.Text(name.Get()),
                                    ),
                                    core.P(core.Class("text-gray-400"),
                                        core.Text(email.Get()),
                                    ),
                                ),
                            ),
                            ui.Divider(ui.DividerProps{}),
                            // Profile form fields...
                        },
                    }),
                },
            }),
        )
    }
}
```

## State of the Art

| Old Approach | Current Approach | Impact |
|--------------|------------------|--------|
| WASM-only components (`components/`) | SSR+hydration components (`ui/`) | Use `ui/` package |
| Sidebar navigation | Horizontal nav in `examples/minimal/admin/` | Follow horizontal nav pattern |
| Manual table rendering | `ui.DataTable[T]` generic | Type-safe, less boilerplate |
| Custom form layout | `ui.FormField` wrapper | Consistent, accessible |

## Open Questions

1. **Resource Type:** The requirements say "Resource list" but don't specify what resource. Recommendation: Use "Projects" as the resource type (common SaaS pattern).

2. **Color Scheme:** The existing examples use dark theme (`bg-gray-900`). Should SaaS dashboard support both light/dark? Recommendation: Start with dark theme for consistency, light theme can be added via Tailwind dark: classes later.

3. **Avatar Upload:** Requirements mention "profile editing with avatar upload". File upload requires additional server-side handling not currently in CRUD. Recommendation: Defer avatar upload to future enhancement; use initials-based Avatar for MVP.

## Sources

### Primary (HIGH confidence)
- `ui/*.go` - All UI component source files reviewed
- `examples/minimal/admin/*.go` - Dashboard, CRUD, form patterns
- `examples/auth/pages/*.go` - Layout patterns
- `examples/marketing/pages/layout.go` - Additional layout patterns
- `core/app.go` - Router and state management

### Secondary (MEDIUM confidence)
- `components/sidebar.go` - Sidebar exists but WASM-only
- `components/charts.go` - Charts exist but WASM-only

## Metadata

**Confidence breakdown:**
- Available components: HIGH - Direct source code review
- Layout patterns: HIGH - Existing examples examined
- CRUD integration: HIGH - Generated API and usage patterns reviewed
- Charts recommendation: MEDIUM - Based on build constraints

**Research date:** 2026-01-23
**Valid until:** 2026-02-23 (stable codebase)
