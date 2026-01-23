package pages

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/admin/.gux/api"
	"github.com/dougbarrett/gux/examples/admin/dto"
	"github.com/dougbarrett/gux/ui"
)

// Users is the user list page with search, filter, and bulk actions.
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
		// Store users in state for hydration
		usersJSON, _ := json.Marshal(users)
		usersState := r.StateString("users", string(usersJSON))

		// Filter state
		searchQuery := r.StateString("searchQuery", "")
		roleFilter := r.StateString("roleFilter", "")
		statusFilter := r.StateString("statusFilter", "")

		// Selected users state (stored as JSON array of IDs)
		selectedJSON := r.StateString("selectedJSON", "[]")

		// Modal state
		showBulkDeleteModal := r.StateBool("showBulkDeleteModal", false)

		// Parse users from state
		var displayUsers []dto.UserList
		json.Unmarshal([]byte(usersState.Get()), &displayUsers)

		// Parse selected IDs
		var selectedIDs []uint
		json.Unmarshal([]byte(selectedJSON.Get()), &selectedIDs)

		// Apply filters
		filteredUsers := filterUsers(displayUsers, searchQuery.Get(), roleFilter.Get(), statusFilter.Get())

		// Check if all filtered users are selected
		allSelected := len(filteredUsers) > 0 && len(selectedIDs) == len(filteredUsers)

		return AdminLayout(
			core.H1(core.Class("text-3xl font-bold text-white mb-8"),
				core.Text("Users"),
			),

			// Add User button
			core.Div(core.Class("mb-6"),
				core.A(
					core.Attrs{Href: "/users/new", Class: "inline-flex items-center px-4 py-2 bg-green-600 hover:bg-green-700 text-white font-medium rounded-lg transition"},
					core.Text("+ Add User"),
				),
			),

			// Search and filter controls
			ui.Card(ui.CardProps{
				Class: "bg-gray-800 border border-gray-700 mb-6",
				Children: []core.Node{
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							ui.HStack(ui.StackProps{
								Gap: "4",
								Children: []core.Node{
									// Search input
									core.Div(core.Class("flex-1"),
										ui.Input(ui.InputProps{
											Type:        ui.InputSearch,
											Name:        "search",
											Value:       searchQuery.Get(),
											Placeholder: "Search by name or email...",
											OnChange:    func(v string) { searchQuery.Set(v) },
											Class:       "bg-gray-700 border-gray-600 text-white",
										}),
									),

									// Role filter
									core.Div(core.Attrs{},
										ui.Select(ui.SelectProps{
											Name:        "roleFilter",
											Value:       roleFilter.Get(),
											Placeholder: "All Roles",
											Options: []ui.SelectOption{
												{Value: "", Label: "All Roles"},
												{Value: "admin", Label: "Admin"},
												{Value: "user", Label: "User"},
												{Value: "editor", Label: "Editor"},
											},
											OnChange: func(v string) { roleFilter.Set(v) },
											Class:    "bg-gray-700 border-gray-600 text-white min-w-[150px]",
										}),
									),

									// Status filter
									core.Div(core.Attrs{},
										ui.Select(ui.SelectProps{
											Name:        "statusFilter",
											Value:       statusFilter.Get(),
											Placeholder: "All Status",
											Options: []ui.SelectOption{
												{Value: "", Label: "All Status"},
												{Value: "active", Label: "Active"},
												{Value: "suspended", Label: "Suspended"},
												{Value: "pending", Label: "Pending"},
											},
											OnChange: func(v string) { statusFilter.Set(v) },
											Class:    "bg-gray-700 border-gray-600 text-white min-w-[150px]",
										}),
									),
								},
							}),
						},
					}),
				},
			}),

			// Bulk action bar (shown when users are selected)
			bulkActionBar(len(selectedIDs), func() {
				// Deselect all
				selectedJSON.Set("[]")
			}, func() {
				// Open bulk delete modal
				showBulkDeleteModal.Set(true)
			}),

			// Users table
			ui.Card(ui.CardProps{
				Class: "bg-gray-800 border border-gray-700",
				Children: []core.Node{
					usersTable(r, filteredUsers, selectedIDs, selectedJSON, allSelected, usersState),
				},
			}),

			// Bulk delete confirmation modal
			bulkDeleteModal(r, len(selectedIDs), showBulkDeleteModal, selectedIDs, selectedJSON, usersState),
		)
	}
}

// filterUsers applies search, role, and status filters.
func filterUsers(users []dto.UserList, query, role, status string) []dto.UserList {
	if query == "" && role == "" && status == "" {
		return users
	}

	query = strings.ToLower(query)
	var filtered []dto.UserList
	for _, u := range users {
		matchesSearch := query == "" ||
			strings.Contains(strings.ToLower(u.Name), query) ||
			strings.Contains(strings.ToLower(u.Email), query)
		matchesRole := role == "" || u.Role == role
		matchesStatus := status == "" || u.Status == status

		if matchesSearch && matchesRole && matchesStatus {
			filtered = append(filtered, u)
		}
	}
	return filtered
}

// bulkActionBar renders the bulk action bar when users are selected.
func bulkActionBar(count int, onDeselect, onDelete func()) core.Node {
	if count == 0 {
		return core.Frag()
	}

	return core.Div(core.Class("mb-4 p-4 bg-blue-900 border border-blue-700 rounded-lg flex items-center justify-between"),
		core.Span(core.Class("text-white font-medium"),
			core.Text(fmt.Sprintf("%d user(s) selected", count)),
		),
		core.Div(core.Class("flex gap-2"),
			core.Button(
				core.Attrs{
					Type:    "button",
					Class:   "px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white font-medium rounded-lg transition",
					OnClick: onDeselect,
				},
				core.Text("Deselect All"),
			),
			core.Button(
				core.Attrs{
					Type:    "button",
					Class:   "px-4 py-2 bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg transition",
					OnClick: onDelete,
				},
				core.Text("Delete Selected"),
			),
		),
	)
}

// usersTable renders the users table with checkboxes.
func usersTable(r *core.Router, users []dto.UserList, selectedIDs []uint, selectedJSON *core.State[string], allSelected bool, usersState *core.State[string]) core.Node {
	if len(users) == 0 {
		return ui.CardContent(ui.CardContentProps{
			Children: []core.Node{
				core.Div(core.Class("text-gray-400 py-8 text-center"),
					core.Text("No users found"),
				),
			},
		})
	}

	// Build header row
	headerRow := ui.Tr(ui.TrProps{
		Children: []core.Node{
			// Select all checkbox
			ui.Th(ui.ThProps{
				Class: "w-12",
				Children: []core.Node{
					core.Input(core.Attrs{
						Type:  "checkbox",
						Class: "w-4 h-4 rounded border-gray-600 bg-gray-700 text-blue-500 focus:ring-blue-500",
						Extra: func() map[string]string {
							if allSelected {
								return map[string]string{"checked": "checked"}
							}
							return nil
						}(),
						OnClick: func() {
							if allSelected {
								// Deselect all
								selectedJSON.Set("[]")
							} else {
								// Select all filtered users
								var ids []uint
								for _, u := range users {
									ids = append(ids, u.ID)
								}
								data, _ := json.Marshal(ids)
								selectedJSON.Set(string(data))
							}
						},
					}),
				},
			}),
			ui.Th(ui.ThProps{Children: []core.Node{core.Text("Name")}}),
			ui.Th(ui.ThProps{Children: []core.Node{core.Text("Email")}}),
			ui.Th(ui.ThProps{Children: []core.Node{core.Text("Role")}}),
			ui.Th(ui.ThProps{Children: []core.Node{core.Text("Status")}}),
			ui.Th(ui.ThProps{Children: []core.Node{core.Text("Last Login")}}),
			ui.Th(ui.ThProps{Children: []core.Node{core.Text("Actions")}}),
		},
	})

	// Build data rows
	var rows []core.Node
	for _, user := range users {
		capturedUser := user
		isSelected := containsID(selectedIDs, capturedUser.ID)

		rows = append(rows, ui.Tr(ui.TrProps{
			Class: "hover:bg-gray-750",
			Children: []core.Node{
				// Checkbox
				ui.Td(ui.TdProps{
					Children: []core.Node{
						core.Input(core.Attrs{
							Type:  "checkbox",
							Class: "w-4 h-4 rounded border-gray-600 bg-gray-700 text-blue-500 focus:ring-blue-500",
							Extra: func() map[string]string {
								if isSelected {
									return map[string]string{"checked": "checked"}
								}
								return nil
							}(),
							OnClick: func() {
								toggleSelect(selectedJSON, capturedUser.ID)
							},
						}),
					},
				}),
				// Name (as link)
				ui.Td(ui.TdProps{
					Children: []core.Node{
						core.A(
							core.Attrs{
								Href:  fmt.Sprintf("/users/%d", capturedUser.ID),
								Class: "text-blue-400 hover:text-blue-300 font-medium",
							},
							core.Text(capturedUser.Name),
						),
					},
				}),
				// Email
				ui.Td(ui.TdProps{
					Children: []core.Node{
						core.Text(capturedUser.Email),
					},
				}),
				// Role badge
				ui.Td(ui.TdProps{
					Children: []core.Node{
						roleBadge(capturedUser.Role),
					},
				}),
				// Status badge
				ui.Td(ui.TdProps{
					Children: []core.Node{
						userStatusBadge(capturedUser.Status),
					},
				}),
				// Last Login
				ui.Td(ui.TdProps{
					Children: []core.Node{
						core.Text(formatLastLogin(capturedUser.LastLoginAt)),
					},
				}),
				// Actions
				ui.Td(ui.TdProps{
					Children: []core.Node{
						core.A(
							core.Attrs{
								Href:  fmt.Sprintf("/users/%d", capturedUser.ID),
								Class: "text-blue-400 hover:text-blue-300",
							},
							core.Text("View"),
						),
					},
				}),
			},
		}))
	}

	return ui.Table(ui.TableProps{
		Class: "bg-gray-800",
		Children: []core.Node{
			ui.Thead(ui.TheadProps{
				Class:    "bg-gray-700",
				Children: []core.Node{headerRow},
			}),
			ui.Tbody(ui.TbodyProps{
				Class:    "divide-gray-700",
				Children: rows,
			}),
		},
	})
}

// bulkDeleteModal renders the bulk delete confirmation modal.
func bulkDeleteModal(r *core.Router, count int, openState *core.State[bool], selectedIDs []uint, selectedJSON, usersState *core.State[string]) core.Node {
	return ui.Modal(ui.ModalProps{
		Open:  openState.Get(),
		Title: fmt.Sprintf("Delete %d User(s)?", count),
		Size:  ui.ModalSM,
		OnClose: func() {
			openState.Set(false)
		},
		Children: []core.Node{
			ui.ModalContent(ui.ModalContentProps{
				Children: []core.Node{
					core.P(core.Class("text-gray-600 dark:text-gray-300"),
						core.Text("Are you sure you want to delete the selected users? This action cannot be undone."),
					),
				},
			}),
			ui.ModalFooter(ui.ModalFooterProps{
				Children: []core.Node{
					core.Button(
						core.Attrs{
							Type:  "button",
							Class: "px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white font-medium rounded-lg transition",
							OnClick: func() {
								openState.Set(false)
							},
						},
						core.Text("Cancel"),
					),
					core.Button(
						core.Attrs{
							Type:  "button",
							Class: "px-4 py-2 bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg transition",
							OnClick: func() {
								// Delete all selected users
								for _, id := range selectedIDs {
									capturedID := id
									api.Users.Delete(capturedID, func(err error) {
										// Individual delete callback
									})
								}
								// Clear selection and close modal
								selectedJSON.Set("[]")
								openState.Set(false)
								// Refresh the user list
								api.Users.List(func(result []dto.UserList, err error) {
									if err == nil {
										data, _ := json.Marshal(result)
										usersState.Set(string(data))
									}
								})
							},
						},
						core.Text("Delete"),
					),
				},
			}),
		},
	})
}

// toggleSelect adds or removes an ID from the selected array.
func toggleSelect(selectedJSON *core.State[string], id uint) {
	var ids []uint
	json.Unmarshal([]byte(selectedJSON.Get()), &ids)

	if containsID(ids, id) {
		// Remove ID
		var newIDs []uint
		for _, existingID := range ids {
			if existingID != id {
				newIDs = append(newIDs, existingID)
			}
		}
		data, _ := json.Marshal(newIDs)
		selectedJSON.Set(string(data))
	} else {
		// Add ID
		ids = append(ids, id)
		data, _ := json.Marshal(ids)
		selectedJSON.Set(string(data))
	}
}

// containsID checks if an ID is in the slice.
func containsID(ids []uint, id uint) bool {
	for _, existingID := range ids {
		if existingID == id {
			return true
		}
	}
	return false
}

// roleBadge returns a badge for the user role.
func roleBadge(role string) core.Node {
	var variant ui.BadgeVariant
	switch role {
	case "admin":
		variant = ui.BadgeError
	case "editor":
		variant = ui.BadgeWarning
	case "user":
		variant = ui.BadgeSuccess
	default:
		variant = ui.BadgeDefault
	}
	return ui.Badge(ui.BadgeProps{
		Variant: variant,
		Text:    capitalizeFirst(role),
	})
}

// userStatusBadge returns a badge for the user status.
func userStatusBadge(status string) core.Node {
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
		Variant: variant,
		Text:    capitalizeFirst(status),
	})
}

// formatLastLogin formats the last login time or returns "Never".
func formatLastLogin(t *time.Time) string {
	if t == nil {
		return "Never"
	}
	return t.Format("Jan 2, 2006")
}
