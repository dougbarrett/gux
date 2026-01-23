# Phase 23: Admin Panel Example - Research

**Researched:** 2026-01-23
**Domain:** Admin panel patterns, user management, activity logging
**Confidence:** HIGH

## Summary

This phase builds a standalone Admin Panel example demonstrating user management, activity logs, and system settings. Research confirms established patterns from existing examples (especially examples/minimal/admin and examples/saas) should be followed.

**Key findings:**
- Horizontal navigation pattern is established (examples/minimal/admin, examples/saas)
- DataTable component supports generic typed data display but lacks built-in bulk selection
- User model and DTO patterns from examples/minimal provide the template
- Activity logs require a new model; no existing logging infrastructure
- Settings page pattern from examples/saas/pages/settings.go demonstrates tabbed interface

**Primary recommendation:** Use horizontal nav layout matching examples/minimal/admin and examples/saas. Implement bulk selection manually with checkboxes and local state. Create ActivityLog model for audit trail. Use tabbed settings matching examples/saas pattern.

## Layout Pattern

### Established Pattern: Horizontal Navigation

Both examples/minimal/admin and examples/saas use horizontal navigation:

```go
// Pattern from examples/minimal/admin/nav.go
func Nav() core.Node {
    return core.Div(core.Class("bg-gray-800 border-b border-gray-700"),
        core.Div(core.Class("max-w-6xl mx-auto px-4"),
            core.Div(core.Class("flex justify-between items-center py-4"),
                // Left: brand + nav links
                core.Div(core.Class("flex items-center gap-6"),
                    core.Span(core.Class("text-xl font-bold text-white"),
                        core.Text("Admin Panel"),
                    ),
                    // Nav links...
                ),
                // Right: back to site / user avatar
            ),
        ),
    )
}
```

### Layout Component

```go
// Pattern from examples/saas/pages/layout.go
func AdminLayout(children ...core.Node) core.Node {
    return core.Div(core.Class("min-h-screen bg-gray-900"),
        Nav(),
        core.Main(core.Class("max-w-6xl mx-auto px-4 py-8"),
            core.Frag(children...),
        ),
    )
}
```

**Decision:** Use horizontal navigation pattern for consistency with existing examples.

**Confidence:** HIGH - verified from examples/minimal/admin/nav.go and examples/saas/pages/layout.go

## User Management

### Model Structure

From examples/minimal/models/user.go:

```go
type User struct {
    gorm.Model
    Email        string `json:"email" gorm:"uniqueIndex"`
    Name         string `json:"name"`
    PasswordHash string `json:"password_hash"` // Never exposed
    Role         string `json:"role"`
    Posts        []Post `json:"posts,omitempty" gorm:"foreignKey:UserID"`
}
```

For admin panel, extend with additional fields:

```go
type User struct {
    gorm.Model
    Email        string    `json:"email" gorm:"uniqueIndex"`
    Name         string    `json:"name"`
    PasswordHash string    `json:"-"` // Use json:"-" to NEVER expose
    Role         string    `json:"role" gorm:"default:user"` // admin, user, viewer
    Status       string    `json:"status" gorm:"default:active"` // active, suspended, pending
    LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}
```

### DTO Structure

From examples/minimal/dto/user.go pattern:

```go
// UserList - for list views (never includes PasswordHash)
type UserList struct {
    ID          uint       `json:"id" gux:"User.ID"`
    Email       string     `json:"email" gux:"User.Email"`
    Name        string     `json:"name" gux:"User.Name"`
    Role        string     `json:"role" gux:"User.Role"`
    Status      string     `json:"status" gux:"User.Status"`
    LastLoginAt *time.Time `json:"last_login_at" gux:"User.LastLoginAt"`
    CreatedAt   time.Time  `json:"created_at" gux:"User.CreatedAt"`
}

// UserDetail - for detail/edit views
type UserDetail struct {
    ID          uint       `json:"id" gux:"User.ID"`
    Email       string     `json:"email" gux:"User.Email"`
    Name        string     `json:"name" gux:"User.Name"`
    Role        string     `json:"role" gux:"User.Role"`
    Status      string     `json:"status" gux:"User.Status"`
    LastLoginAt *time.Time `json:"last_login_at" gux:"User.LastLoginAt"`
    CreatedAt   time.Time  `json:"created_at" gux:"User.CreatedAt"`
    UpdatedAt   time.Time  `json:"updated_at" gux:"User.UpdatedAt"`
}
```

### Search and Filter Implementation

No built-in search component exists. Implement with:

1. **Search input** using `ui.Input` with `InputSearch` type
2. **Filter state** managed via `r.StateString`
3. **Client-side filtering** of loaded data

```go
// Pattern for search/filter
searchQuery := r.StateString("search", "")
roleFilter := r.StateString("roleFilter", "")

// Filter users client-side
var filtered []dto.UserList
for _, u := range displayUsers {
    if matchesSearch(u, searchQuery.Get()) && matchesRole(u, roleFilter.Get()) {
        filtered = append(filtered, u)
    }
}
```

**Confidence:** HIGH - verified from ui/input.go (InputSearch type exists)

## Activity Logs

### New Model Required

No existing activity log model. Create:

```go
type ActivityLog struct {
    gorm.Model
    UserID      *uint   `json:"user_id"`       // nil for system events
    User        *User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Action      string  `json:"action"`        // created, updated, deleted, login, logout
    Entity      string  `json:"entity"`        // user, project, setting
    EntityID    *uint   `json:"entity_id"`     // nil for non-entity actions
    Description string  `json:"description"`   // Human-readable description
    IPAddress   string  `json:"ip_address,omitempty"`
    Metadata    string  `json:"metadata,omitempty"` // JSON for extra data
}
```

### DTO Structure

```go
type ActivityLogList struct {
    ID          uint       `json:"id" gux:"ActivityLog.ID"`
    UserID      *uint      `json:"user_id" gux:"ActivityLog.UserID"`
    UserName    string     `json:"user_name"` // Derived via preload
    Action      string     `json:"action" gux:"ActivityLog.Action"`
    Entity      string     `json:"entity" gux:"ActivityLog.Entity"`
    Description string     `json:"description" gux:"ActivityLog.Description"`
    CreatedAt   time.Time  `json:"created_at" gux:"ActivityLog.CreatedAt"`
}
```

### Display Pattern

Use DataTable for activity logs list:

```go
ui.DataTable(ui.DataTableProps[dto.ActivityLogList]{
    Data: logs,
    Columns: []ui.ColumnDef[dto.ActivityLogList]{
        {Header: "Time", Render: func(l dto.ActivityLogList) core.Node {
            return core.Text(l.CreatedAt.Format("Jan 2, 15:04"))
        }},
        {Header: "User", Render: func(l dto.ActivityLogList) core.Node {
            return core.Text(l.UserName)
        }},
        {Header: "Action", Render: func(l dto.ActivityLogList) core.Node {
            return actionBadge(l.Action)
        }},
        {Header: "Description", Render: func(l dto.ActivityLogList) core.Node {
            return core.Text(l.Description)
        }},
    },
    Striped:   true,
    Hoverable: true,
})
```

**Confidence:** MEDIUM - new model, pattern based on existing DataTable usage

## Role-Based UI

### Display-Only Approach (As Specified)

Requirements state "display only, defer actual RBAC". Implement as conditional rendering:

```go
// Mock current user (hardcoded for example)
currentUser := struct {
    ID   uint
    Name string
    Role string
}{ID: 1, Name: "Admin User", Role: "admin"}

// Conditional rendering based on role
func showAdminActions(role string) bool {
    return role == "admin"
}

// Usage in component
func() core.Node {
    if showAdminActions(currentUser.Role) {
        return ui.Button(ui.ButtonProps{
            Variant:  ui.ButtonDestructive,
            Children: []core.Node{core.Text("Delete User")},
        })
    }
    return core.Frag()
}()
```

### Role Badge Helper

```go
func roleBadge(role string) core.Node {
    var variant ui.BadgeVariant
    switch role {
    case "admin":
        variant = ui.BadgeError // Red for admin
    case "user":
        variant = ui.BadgeSuccess // Green for regular users
    case "viewer":
        variant = ui.BadgeDefault // Gray for viewers
    default:
        variant = ui.BadgeDefault
    }
    return ui.Badge(ui.BadgeProps{
        Text:    role,
        Variant: variant,
    })
}
```

**Confidence:** HIGH - simple conditional rendering, no framework changes needed

## Bulk Actions

### Current DataTable Limitation

The DataTable component (ui/datatable.go) does NOT have built-in selection support:

```go
// DataTableProps - no Selection field exists
type DataTableProps[T any] struct {
    Data       []T
    Columns    []ColumnDef[T]
    RowKey     func(T) string
    OnRowClick func(T)
    Striped    bool
    Hoverable  bool
    Class      string
}
```

### Implementation Pattern

Implement bulk selection with manual checkbox column and state:

```go
func UsersPage(r *core.Router) func() core.Node {
    // ...loader...

    return func() core.Node {
        // Track selected IDs in state
        selectedJSON := r.StateString("selected", "[]")

        // Parse selected IDs
        var selected []uint
        json.Unmarshal([]byte(selectedJSON.Get()), &selected)

        // Toggle selection helper
        toggleSelect := func(id uint) {
            var newSelected []uint
            found := false
            for _, s := range selected {
                if s == id {
                    found = true
                } else {
                    newSelected = append(newSelected, s)
                }
            }
            if !found {
                newSelected = append(newSelected, id)
            }
            b, _ := json.Marshal(newSelected)
            selectedJSON.Set(string(b))
        }

        // Check if ID is selected
        isSelected := func(id uint) bool {
            for _, s := range selected {
                if s == id {
                    return true
                }
            }
            return false
        }

        // Bulk action bar (shown when items selected)
        bulkActionBar := func() core.Node {
            if len(selected) == 0 {
                return core.Frag()
            }
            return core.Div(core.Class("bg-blue-900/30 border border-blue-700 rounded-lg p-4 mb-4 flex justify-between items-center"),
                core.Text(fmt.Sprintf("%d users selected", len(selected))),
                ui.HStack(ui.StackProps{
                    Gap: "2",
                    Children: []core.Node{
                        ui.Button(ui.ButtonProps{
                            Variant:  ui.ButtonDestructive,
                            Children: []core.Node{core.Text("Delete Selected")},
                            OnClick: func() {
                                // Open confirmation modal
                            },
                        }),
                    },
                }),
            )
        }()

        // Manual table with checkbox column
        // (Cannot use DataTable directly for selection)
    }
}
```

### Alternative: Custom Table with Selection

For bulk actions, use manual `ui.Table` components with checkbox column:

```go
ui.Table(ui.TableProps{
    Children: []core.Node{
        ui.Thead(ui.TheadProps{
            Children: []core.Node{
                ui.Tr(ui.TrProps{
                    Children: []core.Node{
                        ui.Th(ui.ThProps{
                            Class: "w-8",
                            Children: []core.Node{
                                ui.Checkbox(ui.CheckboxProps{
                                    Checked: len(selected) == len(users),
                                    // Toggle all
                                }),
                            },
                        }),
                        ui.Th(ui.ThProps{Children: []core.Node{core.Text("Name")}}),
                        // ...other headers
                    },
                }),
            },
        }),
        ui.Tbody(ui.TbodyProps{
            Children: userRows(users, selected, toggleSelect),
        }),
    },
})
```

**Confidence:** HIGH - verified DataTable lacks selection; manual implementation required

## Settings

### Tabbed Interface Pattern

From examples/saas/pages/settings.go:

```go
func Settings(r *core.Router) func() core.Node {
    return func() core.Node {
        activeTab := r.StateInt("activeTab", 0)

        return AdminLayout(
            core.H1(core.Class("text-3xl font-bold text-white mb-8"),
                core.Text("Settings"),
            ),

            ui.Card(ui.CardProps{
                Class: "bg-gray-800 border border-gray-700",
                Children: []core.Node{
                    ui.CardContent(ui.CardContentProps{
                        Children: []core.Node{
                            ui.Tabs(ui.TabsProps{
                                Children: []core.Node{
                                    ui.TabList(ui.TabListProps{
                                        Class: "border-gray-700",
                                        Children: []core.Node{
                                            ui.Tab(ui.TabProps{
                                                Active:   activeTab.Get() == 0,
                                                OnSelect: func() { activeTab.Set(0) },
                                                Children: []core.Node{core.Text("General")},
                                            }),
                                            ui.Tab(ui.TabProps{
                                                Active:   activeTab.Get() == 1,
                                                OnSelect: func() { activeTab.Set(1) },
                                                Children: []core.Node{core.Text("Security")},
                                            }),
                                            ui.Tab(ui.TabProps{
                                                Active:   activeTab.Get() == 2,
                                                OnSelect: func() { activeTab.Set(2) },
                                                Children: []core.Node{core.Text("Email")},
                                            }),
                                        },
                                    }),
                                    ui.TabPanel(ui.TabPanelProps{Active: activeTab.Get() == 0, Children: generalTab(r)}),
                                    ui.TabPanel(ui.TabPanelProps{Active: activeTab.Get() == 1, Children: securityTab(r)}),
                                    ui.TabPanel(ui.TabPanelProps{Active: activeTab.Get() == 2, Children: emailTab(r)}),
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

### Admin Settings Categories

**General Settings:**
- Site name
- Site description
- Default timezone
- Default language

**Security Settings:**
- Session timeout
- Password requirements
- Two-factor settings (display only)

**Email Settings:**
- SMTP configuration (display only)
- Email templates preview

**Note:** Settings are display/mock only for this example - no actual persistence needed.

**Confidence:** HIGH - directly from examples/saas/pages/settings.go

## Technical Decisions

### Port Assignment

Following established pattern:
- examples/minimal: 8080
- examples/auth: 8081
- examples/saas: 8082
- examples/marketing: 8083 (assumed)
- **examples/admin: 8084**

### Single WASM Bundle

Per Phase 22 decisions: "Single WASM bundle (no separate admin bundle) for simplicity"

```go
app.Routes().
    Hybrid("/", pages.Dashboard).
    Hybrid("/users", pages.Users).
    Hybrid("/users/:id", pages.UserDetail).
    Hybrid("/users/:id/edit", pages.UserEdit).
    Hybrid("/users/new", pages.UserNew).
    Hybrid("/activity", pages.ActivityLogs).
    Hybrid("/settings", pages.Settings)
```

### Database: SQLite

Following pattern from other examples:

```go
db, err := gorm.Open(sqlite.Open("admin.db"), &gorm.Config{})
db.AutoMigrate(&models.User{}, &models.ActivityLog{})
```

### Mock Data Seeding

Seed sample users and activity logs on first run:

```go
var count int64
db.Model(&models.User{}).Count(&count)
if count == 0 {
    // Seed admin user
    admin := models.User{
        Email:  "admin@example.com",
        Name:   "Admin User",
        Role:   "admin",
        Status: "active",
    }
    admin.SetPassword("admin123")
    db.Create(&admin)

    // Seed regular users
    // Seed activity logs
}
```

**Confidence:** HIGH - follows established example patterns

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Table display | Custom HTML table | ui.Table or ui.DataTable | Consistent styling, dark mode support |
| Modal dialogs | Custom overlay | ui.Modal | Accessibility, backdrop handling |
| Form inputs | Raw HTML inputs | ui.Input, ui.Select | Consistent styling, validation states |
| Badges/Status | Custom spans | ui.Badge | Consistent color variants |
| Tabs | Custom tab switching | ui.Tabs | ARIA attributes, keyboard nav |
| Pagination | Custom page links | ui.Pagination | Page calculation, disabled states |
| Password hashing | Custom crypto | bcrypt (golang.org/x/crypto/bcrypt) | Security best practices |

## Common Pitfalls

### Pitfall 1: Exposing PasswordHash
**What goes wrong:** PasswordHash field appears in API responses
**Why it happens:** Forgetting to use DTOs or json:"-" tag
**How to avoid:**
- Use `json:"-"` on model: `PasswordHash string \`json:"-"\``
- Use DTOs that don't include PasswordHash
**Warning signs:** PasswordHash visible in network tab

### Pitfall 2: State Hydration with Complex Types
**What goes wrong:** Arrays/objects don't hydrate correctly
**Why it happens:** Must serialize to JSON string for state
**How to avoid:**
```go
// Correct pattern
usersJSON, _ := json.Marshal(users)
usersState := r.StateString("users", string(usersJSON))
var displayUsers []dto.UserList
json.Unmarshal([]byte(usersState.Get()), &displayUsers)
```
**Warning signs:** Empty arrays after WASM hydration

### Pitfall 3: Checkbox onChange Limitation
**What goes wrong:** Checkbox onChange doesn't work as expected
**Why it happens:** core.Attrs.OnChange returns string, not bool
**How to avoid:** Use row-level OnClick instead of checkbox OnChange
**Warning signs:** Checkbox clicks don't update state

### Pitfall 4: Modal Click Propagation
**What goes wrong:** Clicking modal content closes the modal
**Why it happens:** Click propagates to backdrop
**How to avoid:** Modal component already handles this - don't add extra OnClick to modal content
**Warning signs:** Modal closes when clicking inside content

## Code Examples

### Dashboard with Stats Cards

```go
// From examples/saas/pages/dashboard.go pattern
func statCard(title, value, colorClass string) core.Node {
    return ui.Card(ui.CardProps{
        Class: "bg-gray-800 border border-gray-700",
        Children: []core.Node{
            ui.CardContent(ui.CardContentProps{
                Children: []core.Node{
                    core.Div(core.Class("text-gray-400 text-sm"),
                        core.Text(title),
                    ),
                    core.Div(core.Class("text-3xl font-bold mt-2 "+colorClass),
                        core.Text(value),
                    ),
                },
            }),
        },
    })
}
```

### Delete Confirmation Modal

```go
// From examples/saas/pages/project_detail.go pattern
func deleteConfirmationModal(r *core.Router, user *dto.UserDetail, openState *core.State[bool]) core.Node {
    return ui.Modal(ui.ModalProps{
        Open:  openState.Get(),
        Title: "Delete User",
        Size:  ui.ModalSM,
        OnClose: func() {
            openState.Set(false)
        },
        Children: []core.Node{
            ui.ModalContent(ui.ModalContentProps{
                Children: []core.Node{
                    core.P(core.Class("text-gray-600 dark:text-gray-300"),
                        core.Text(fmt.Sprintf("Are you sure you want to delete \"%s\"?", user.Name)),
                    ),
                },
            }),
            ui.ModalFooter(ui.ModalFooterProps{
                Children: []core.Node{
                    ui.Button(ui.ButtonProps{
                        Variant:  ui.ButtonSecondary,
                        OnClick:  func() { openState.Set(false) },
                        Children: []core.Node{core.Text("Cancel")},
                    }),
                    ui.Button(ui.ButtonProps{
                        Variant: ui.ButtonDestructive,
                        OnClick: func() {
                            api.Users.Delete(user.ID, func(err error) {
                                if err == nil {
                                    r.Navigate("/users")
                                }
                            })
                        },
                        Children: []core.Node{core.Text("Delete")},
                    }),
                },
            }),
        },
    })
}
```

### User Status Badge

```go
func statusBadge(status string) core.Node {
    var variant ui.BadgeVariant
    switch status {
    case "active":
        variant = ui.BadgeSuccess
    case "suspended":
        variant = ui.BadgeError
    case "pending":
        variant = ui.BadgeWarning
    default:
        variant = ui.BadgeDefault
    }
    return ui.Badge(ui.BadgeProps{
        Text:    status,
        Variant: variant,
    })
}
```

## Project Structure

```
examples/admin/
├── app.go                      # App setup, routes, CRUD
├── models/
│   ├── user.go                 # User model with Role, Status
│   └── activity_log.go         # ActivityLog model
├── dto/
│   ├── user.go                 # UserList, UserDetail DTOs
│   └── activity_log.go         # ActivityLogList DTO
├── pages/
│   ├── layout.go               # AdminLayout, Nav
│   ├── dashboard.go            # Dashboard with stats
│   ├── users.go                # User list with search/filter
│   ├── user_detail.go          # User detail view
│   ├── user_edit.go            # User edit form
│   ├── user_new.go             # Create user form
│   ├── activity.go             # Activity logs list
│   └── settings.go             # Tabbed settings page
└── .gux/                       # Generated (api/, wasm/)
```

## Open Questions

1. **Bulk delete confirmation:** Should bulk delete show list of users being deleted, or just count?
   - Recommendation: Show count with warning, matching single-delete modal pattern

2. **Activity log retention:** Should activity logs have a maximum age?
   - Recommendation: No retention for example; mention in settings as display-only option

3. **User avatar:** Should users have avatars?
   - Recommendation: Use ui.Avatar with initials (Name), no image upload for simplicity

## Sources

### Primary (HIGH confidence)
- examples/minimal/admin/nav.go - Horizontal nav pattern
- examples/minimal/admin/users.go - User list pattern
- examples/minimal/admin/user_detail.go - User detail pattern
- examples/minimal/models/user.go - User model structure
- examples/minimal/dto/user.go - DTO patterns
- examples/saas/pages/layout.go - DashboardLayout pattern
- examples/saas/pages/dashboard.go - statCard pattern
- examples/saas/pages/settings.go - Tabbed settings pattern
- examples/saas/pages/project_detail.go - Delete confirmation modal
- ui/datatable.go - DataTable API (no selection support)
- ui/modal.go - Modal component API
- ui/badge.go - Badge variants
- ui/tabs.go - Tabs component API
- ui/checkbox.go - Checkbox component API
- ui/input.go - Input types (InputSearch)
- ui/pagination.go - Pagination component API

### Secondary (MEDIUM confidence)
- Phase 22 decisions in STATE.md - Single bundle, horizontal nav, port pattern

## Metadata

**Confidence breakdown:**
- Layout pattern: HIGH - directly from existing examples
- User management: HIGH - directly from existing models/DTOs
- Activity logs: MEDIUM - new model, based on existing patterns
- Role-based UI: HIGH - simple conditional rendering
- Bulk actions: HIGH - verified DataTable limitation
- Settings: HIGH - directly from examples/saas/pages/settings.go

**Research date:** 2026-01-23
**Valid until:** 2026-02-23 (stable patterns, 30 days)
