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
	"{{.ModulePath}}/guxgen/api"
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
	"{{.ModulePath}}/guxgen/api"
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
	"{{.ModulePath}}/guxgen/api"
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
	"api": {
		Name:        "api",
		Description: "Typed API endpoint with request/response types",
		FilePath:    "app.go (api section)",
		Template: `// app.go (api section)
// Typed API endpoints provide compile-time type safety with automatic JSON handling.

// Define request and response types (typically in dto/ package)
type LoginRequest struct {
	Email    string ` + "`" + `json:"email"` + "`" + `
	Password string ` + "`" + `json:"password"` + "`" + `
}

type LoginResponse struct {
	Success  bool   ` + "`" + `json:"success"` + "`" + `
	Redirect string ` + "`" + `json:"redirect,omitempty"` + "`" + `
	Error    string ` + "`" + `json:"error,omitempty"` + "`" + `
}

// Register typed POST endpoint - request body is auto-decoded
core.API(app, "POST", "/api/login", func(ctx *core.APIContext, req LoginRequest) (LoginResponse, error) {
	db := ctx.DB().(*gorm.DB)

	// Find user
	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return LoginResponse{Error: "Invalid credentials"}, nil
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		return LoginResponse{Error: "Invalid credentials"}, nil
	}

	// Create session
	ctx.Login(&core.SessionUser{
		ID:    fmt.Sprintf("%d", user.ID),
		Email: user.Email,
		Name:  user.Name,
		Roles: []string{user.Role},
	})

	return LoginResponse{Success: true, Redirect: "/dashboard"}, nil
})

// Register typed GET endpoint (no request body)
core.APIGet(app, "/api/users/:id", func(ctx *core.APIContext) (dto.UserDetail, error) {
	id := ctx.ParamUint("id")
	db := ctx.DB().(*gorm.DB)

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		return dto.UserDetail{}, api.NotFound("User not found")
	}

	return dto.UserDetail{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}, nil
})

// Register typed DELETE endpoint
core.APIDelete(app, "/api/users/:id", func(ctx *core.APIContext) error {
	id := ctx.ParamUint("id")
	db := ctx.DB().(*gorm.DB)

	if err := db.Delete(&models.User{}, id).Error; err != nil {
		return api.InternalError("Failed to delete user")
	}
	return nil
})

// APIContext methods:
//   ctx.Param("id")      - Get path parameter as string
//   ctx.ParamInt("id")   - Get path parameter as int
//   ctx.ParamUint("id")  - Get path parameter as uint
//   ctx.Query("search")  - Get query parameter
//   ctx.Header("X-Key")  - Get request header
//   ctx.User()           - Get authenticated user (or nil)
//   ctx.DB()             - Get database connection
//   ctx.Login(user)      - Create session for user
//   ctx.Logout()         - Destroy current session
`,
	},
	"api:handler": {
		Name:        "api:handler",
		Description: "Extracted API handler function",
		FilePath:    "handlers/auth.go",
		Template: `// handlers/auth.go
// Extracted handlers for cleaner app.go organization
package handlers

import (
	"fmt"

	"{{.ModulePath}}/dto"
	"{{.ModulePath}}/models"

	"github.com/dougbarrett/gux/api"
	"github.com/dougbarrett/gux/core"
	"gorm.io/gorm"
)

// LoginHandler handles user authentication.
// Register in app.go: core.API(app, "POST", "/api/login", handlers.LoginHandler)
func LoginHandler(ctx *core.APIContext, req dto.LoginRequest) (dto.LoginResponse, error) {
	db := ctx.DB().(*gorm.DB)

	// Find user by email
	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return dto.LoginResponse{Error: "Invalid email or password"}, nil
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		return dto.LoginResponse{Error: "Invalid email or password"}, nil
	}

	// Create session
	if err := ctx.Login(&core.SessionUser{
		ID:    fmt.Sprintf("%d", user.ID),
		Email: user.Email,
		Name:  user.Name,
		Roles: []string{user.Role},
	}); err != nil {
		return dto.LoginResponse{}, api.InternalError("Failed to create session")
	}

	return dto.LoginResponse{Success: true, Redirect: "/dashboard"}, nil
}

// LogoutHandler handles user logout.
// Register in app.go: core.API(app, "POST", "/api/logout", handlers.LogoutHandler)
func LogoutHandler(ctx *core.APIContext, req struct{}) (dto.LogoutResponse, error) {
	ctx.Logout()
	return dto.LogoutResponse{Success: true, Redirect: "/"}, nil
}

// GetUserHandler returns user details.
// Register in app.go: core.APIGet(app, "/api/users/:id", handlers.GetUserHandler)
func GetUserHandler(ctx *core.APIContext) (dto.UserDetail, error) {
	id := ctx.ParamUint("id")
	db := ctx.DB().(*gorm.DB)

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		return dto.UserDetail{}, api.NotFound("User not found")
	}

	return dto.UserDetail{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}, nil
}
`,
	},
	"layout:admin": {
		Name:        "layout:admin",
		Description: "Complete admin layout with sidebar, header, and breadcrumbs",
		FilePath:    "pages/layout.go",
		Template: `// pages/layout.go
// Admin layout with collapsible sidebar, mobile menu, and page header with breadcrumbs
package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// AdminLayout wraps admin pages with sidebar navigation.
// Use this as a layout wrapper for all admin routes.
//
// Usage in route registration:
//   app.RouteGroup("/admin", core.WithBundle("admin")).
//       Hybrid("/", AdminLayout(Dashboard)).
//       Hybrid("/users", AdminLayout(Users)).
//       Hybrid("/users/:id", AdminLayout(UserDetail))
func AdminLayout(page func(*core.Router) func() core.Node) func(*core.Router) func() core.Node {
	return func(r *core.Router) func() core.Node {
		// Get the inner page component
		pageComponent := page(r)

		return func() core.Node {
			// Sidebar state
			collapsed := r.StateBool("collapsed", false)
			mobileOpen := r.StateBool("mobileOpen", false)
			currentPath := r.Path()

			// Navigation items - customize for your app
			navItems := []ui.SidebarItemProps{
				{Label: "Dashboard", Href: "/admin", Icon: "home", Active: currentPath == "/admin"},
				{Label: "Users", Href: "/admin/users", Icon: "users", Active: currentPath == "/admin/users" || currentPath[:12] == "/admin/users"},
				{Label: "Settings", Href: "/admin/settings", Icon: "settings", Active: currentPath == "/admin/settings"},
			}

			// Build sidebar items
			sidebarItems := make([]core.Node, len(navItems))
			for i, item := range navItems {
				item.Collapsed = collapsed.Get()
				sidebarItems[i] = ui.SidebarItem(item)
			}

			return ui.SidebarLayout(ui.SidebarLayoutProps{
				Sidebar: ui.Sidebar(ui.SidebarProps{
					Collapsed:     collapsed.Get(),
					MobileOpen:    mobileOpen.Get(),
					OnCloseMobile: func() { mobileOpen.Set(false) },
					Children: []core.Node{
						// Header with logo/title
						ui.SidebarHeader(ui.SidebarHeaderProps{
							Title:            "Admin Panel",
							Collapsed:        collapsed.Get(),
							OnToggleCollapse: func() { collapsed.Set(!collapsed.Get()) },
							OnCloseMobile:    func() { mobileOpen.Set(false) },
						}),
						// Navigation
						ui.SidebarNav(ui.SidebarNavProps{
							Children: []core.Node{
								ui.SidebarSection(ui.SidebarSectionProps{
									Title:     "Main",
									Collapsed: collapsed.Get(),
									Children:  sidebarItems,
								}),
							},
						}),
						// Footer with user info
						ui.SidebarFooter(ui.SidebarFooterProps{
							Collapsed: collapsed.Get(),
							Children: []core.Node{
								ui.SidebarUser(ui.SidebarUserProps{
									Name:      "Admin User",
									Email:     "admin@example.com",
									Href:      "/admin/profile",
									Collapsed: collapsed.Get(),
								}),
							},
						}),
					},
				}),
				Children: []core.Node{
					ui.SidebarMain(ui.SidebarMainProps{
						Collapsed:    collapsed.Get(),
						MobileOpen:   mobileOpen.Get(),
						OnOpenMobile: func() { mobileOpen.Set(true) },
						HeaderTitle:  "Admin",
						Children: []core.Node{
							// Render the actual page content
							pageComponent(),
						},
					}),
				},
			})
		}
	}
}

// Example: Dashboard page with breadcrumbs
func Dashboard(r *core.Router) func() core.Node {
	return func() core.Node {
		return core.Div(core.Attrs{},
			ui.PageHeader(ui.PageHeaderProps{
				Title:    "Dashboard",
				Subtitle: "Welcome to the admin panel",
				Breadcrumbs: []ui.BreadcrumbItem{
					{Label: "Admin", Href: "/admin"},
					{Label: "Dashboard"},
				},
			}),
			// Dashboard content here
			core.Div(core.Class("grid grid-cols-1 md:grid-cols-3 gap-6"),
				ui.Card(ui.CardProps{
					Children: []core.Node{
						ui.CardHeader(ui.CardHeaderProps{
							Children: []core.Node{core.Text("Total Users")},
						}),
						ui.CardContent(ui.CardContentProps{
							Children: []core.Node{
								core.Div(core.Class("text-3xl font-bold"), core.Text("1,234")),
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
	"layout:sidebar": {
		Name:        "layout:sidebar",
		Description: "Sidebar navigation setup with sections and items",
		FilePath:    "pages/sidebar.go",
		Template: `// pages/sidebar.go
// Sidebar navigation with sections, nested items, and user profile
package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// BuildSidebar creates the sidebar navigation component.
// Call this from your layout to render the sidebar.
func BuildSidebar(r *core.Router, collapsed, mobileOpen *core.StateBool) core.Node {
	currentPath := r.Path()

	return ui.Sidebar(ui.SidebarProps{
		Collapsed:     collapsed.Get(),
		MobileOpen:    mobileOpen.Get(),
		OnCloseMobile: func() { mobileOpen.Set(false) },
		Children: []core.Node{
			// Header
			ui.SidebarHeader(ui.SidebarHeaderProps{
				Title:            "My App",
				Collapsed:        collapsed.Get(),
				OnToggleCollapse: func() { collapsed.Set(!collapsed.Get()) },
				OnCloseMobile:    func() { mobileOpen.Set(false) },
			}),

			// Navigation with multiple sections
			ui.SidebarNav(ui.SidebarNavProps{
				Children: []core.Node{
					// Main section
					ui.SidebarSection(ui.SidebarSectionProps{
						Title:     "Main",
						Collapsed: collapsed.Get(),
						Children: []core.Node{
							ui.SidebarItem(ui.SidebarItemProps{
								Label:     "Dashboard",
								Href:      "/",
								Icon:      "home",
								Active:    currentPath == "/",
								Collapsed: collapsed.Get(),
							}),
							ui.SidebarItem(ui.SidebarItemProps{
								Label:     "Analytics",
								Href:      "/analytics",
								Icon:      "bar-chart-2",
								Active:    currentPath == "/analytics",
								Collapsed: collapsed.Get(),
								Badge:     "New", // Optional badge
							}),
						},
					}),

					// Management section
					ui.SidebarSection(ui.SidebarSectionProps{
						Title:     "Management",
						Collapsed: collapsed.Get(),
						Children: []core.Node{
							ui.SidebarItem(ui.SidebarItemProps{
								Label:     "Users",
								Href:      "/users",
								Icon:      "users",
								Active:    hasPrefix(currentPath, "/users"),
								Collapsed: collapsed.Get(),
							}),
							ui.SidebarItem(ui.SidebarItemProps{
								Label:     "Projects",
								Href:      "/projects",
								Icon:      "folder",
								Active:    hasPrefix(currentPath, "/projects"),
								Collapsed: collapsed.Get(),
							}),
							ui.SidebarItem(ui.SidebarItemProps{
								Label:     "Teams",
								Href:      "/teams",
								Icon:      "users",
								Active:    hasPrefix(currentPath, "/teams"),
								Collapsed: collapsed.Get(),
							}),
						},
					}),

					// Settings section
					ui.SidebarSection(ui.SidebarSectionProps{
						Title:     "System",
						Collapsed: collapsed.Get(),
						Children: []core.Node{
							ui.SidebarItem(ui.SidebarItemProps{
								Label:     "Settings",
								Href:      "/settings",
								Icon:      "settings",
								Active:    currentPath == "/settings",
								Collapsed: collapsed.Get(),
							}),
							ui.SidebarItem(ui.SidebarItemProps{
								Label:     "Help",
								Href:      "/help",
								Icon:      "help-circle",
								Active:    currentPath == "/help",
								Collapsed: collapsed.Get(),
							}),
						},
					}),
				},
			}),

			// Footer with user profile
			ui.SidebarFooter(ui.SidebarFooterProps{
				Collapsed: collapsed.Get(),
				Children: []core.Node{
					ui.SidebarUser(ui.SidebarUserProps{
						Name:      "John Doe",
						Email:     "john@example.com",
						Href:      "/profile",
						Collapsed: collapsed.Get(),
					}),
				},
			}),
		},
	})
}

// hasPrefix checks if path starts with prefix.
// Use for highlighting parent nav items when on child pages.
func hasPrefix(path, prefix string) bool {
	if len(path) < len(prefix) {
		return false
	}
	return path[:len(prefix)] == prefix
}

// Available icons (from ui.Icon):
// - home, users, settings, folder, file, search
// - bar-chart-2, pie-chart, trending-up
// - mail, bell, calendar, clock
// - plus, minus, x, check, alert-circle
// - chevron-left, chevron-right, chevron-down, chevron-up
// - menu, grid, list, layout
// - edit, trash, download, upload
// - eye, eye-off, lock, unlock
// - heart, star, bookmark
// - help-circle, info, external-link
`,
	},
	"breadcrumb": {
		Name:        "breadcrumb",
		Description: "Breadcrumb navigation with page header",
		FilePath:    "pages/example.go",
		Template: `// Breadcrumb navigation example
// Use ui.Breadcrumb for simple trails, ui.PageHeader for full page headers
package pages

import (
	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// Example page with breadcrumb navigation
func UserDetail(r *core.Router) func() core.Node {
	// Data loading
	userID := r.Param("id")
	userName := "John Doe" // Load from API

	return func() core.Node {
		return core.Div(core.Class("p-6"),
			// Option 1: Simple breadcrumb trail
			ui.Breadcrumb(ui.BreadcrumbProps{
				Items: []ui.BreadcrumbItem{
					{Label: "Home", Href: "/"},
					{Label: "Users", Href: "/users"},
					{Label: userName}, // Current page (no href)
				},
				HomeIcon: true, // Show home icon on first item
			}),

			// Page content
			core.H1(core.Class("text-2xl font-bold mt-4 mb-6"),
				core.Text(userName),
			),
		)
	}
}

// Example with full PageHeader (includes title, subtitle, actions)
func UserDetailFull(r *core.Router) func() core.Node {
	userID := r.Param("id")
	userName := "John Doe"

	return func() core.Node {
		return core.Div(core.Class("p-6"),
			// PageHeader combines breadcrumbs, title, subtitle, and actions
			ui.PageHeader(ui.PageHeaderProps{
				Breadcrumbs: []ui.BreadcrumbItem{
					{Label: "Home", Href: "/"},
					{Label: "Users", Href: "/users"},
					{Label: userName},
				},
				Title:    userName,
				Subtitle: "User profile and settings",
				Actions: []core.Node{
					ui.Button(ui.ButtonProps{
						Variant: "outline",
						Children: []core.Node{
							ui.Icon(ui.IconProps{Name: "edit", Size: ui.IconSM}),
							core.Text(" Edit"),
						},
					}),
					ui.Button(ui.ButtonProps{
						Variant: "destructive",
						Children: []core.Node{
							ui.Icon(ui.IconProps{Name: "trash", Size: ui.IconSM}),
							core.Text(" Delete"),
						},
					}),
				},
			}),

			// Page content
			ui.Card(ui.CardProps{
				Children: []core.Node{
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							core.Text("User details go here..."),
						},
					}),
				},
			}),
		)
	}
}

// BreadcrumbItem options:
//   Label    string  - Display text (required)
//   Href     string  - Link URL (empty for current page)
//   Icon     string  - Custom icon (optional)

// PageHeader options:
//   Title       string            - Page title
//   Subtitle    string            - Optional description
//   Breadcrumbs []BreadcrumbItem  - Breadcrumb trail
//   Actions     []core.Node       - Buttons on right side
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
