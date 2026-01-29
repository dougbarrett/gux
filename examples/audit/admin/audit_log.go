package admin

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/audit/dto"
	"github.com/dougbarrett/gux/examples/audit/guxgen/api"
	"github.com/dougbarrett/gux/examples/audit/pages"
	"github.com/dougbarrett/gux/ui"
)

// AuditLog is the audit trail page showing all audit entries with filtering.
func AuditLog(r *core.Router) func() core.Node {
	var entries []dto.AuditEntryList

	r.OnLoad(func() {
		api.AuditEntries.List(func(result []dto.AuditEntryList, err error) {
			if err == nil {
				entries = result
			}
		})
	})

	return func() core.Node {
		entriesJSON, _ := json.Marshal(entries)
		entriesState := r.StateString("entries", string(entriesJSON))

		// Filters
		actionFilter := r.StateString("actionFilter", "")
		entityFilter := r.StateString("entityFilter", "")

		var displayEntries []dto.AuditEntryList
		json.Unmarshal([]byte(entriesState.Get()), &displayEntries)

		// Apply filters
		filtered := filterAuditEntries(displayEntries, actionFilter.Get(), entityFilter.Get())

		return pages.AdminLayout(r, "Audit Log",
			core.H1(core.Class("text-3xl font-bold text-white mb-6"),
				core.Text("Audit Log"),
			),
			core.P(core.Class("text-gray-400 mb-6"),
				core.Text("Complete audit trail of all changes made through the system. Every create, update, and delete operation is logged with user information and field-level diffs."),
			),

			// Filter controls
			ui.Card(ui.CardProps{
				Class: "bg-gray-800 border border-gray-700 mb-6",
				Children: []core.Node{
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							ui.HStack(ui.StackProps{
								Gap: "4",
								Children: []core.Node{
									core.Div(core.Attrs{},
										core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
											core.Text("Filter by Action"),
										),
										ui.Select(ui.SelectProps{
											Name:  "actionFilter",
											Value: actionFilter.Get(),
											Options: []ui.SelectOption{
												{Value: "", Label: "All Actions"},
												{Value: "create", Label: "Create"},
												{Value: "update", Label: "Update"},
												{Value: "delete", Label: "Delete"},
											},
											OnChange: func(v string) { actionFilter.Set(v) },
											Class:    "bg-gray-700 border-gray-600 text-white min-w-[150px]",
										}),
									),
									core.Div(core.Attrs{},
										core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
											core.Text("Filter by Entity"),
										),
										ui.Select(ui.SelectProps{
											Name:  "entityFilter",
											Value: entityFilter.Get(),
											Options: []ui.SelectOption{
												{Value: "", Label: "All Entities"},
												{Value: "User", Label: "User"},
												{Value: "Document", Label: "Document"},
											},
											OnChange: func(v string) { entityFilter.Set(v) },
											Class:    "bg-gray-700 border-gray-600 text-white min-w-[150px]",
										}),
									),
								},
							}),
						},
					}),
				},
			}),

			// Audit entries table
			func() core.Node {
				if len(filtered) == 0 {
					return ui.Card(ui.CardProps{
						Class: "bg-gray-800 border border-gray-700",
						Children: []core.Node{
							ui.CardContent(ui.CardContentProps{
								Children: []core.Node{
									core.Div(core.Class("text-center py-8 text-gray-400"),
										core.Text("No audit entries found"),
									),
								},
							}),
						},
					})
				}

				return ui.Card(ui.CardProps{
					Class: "bg-gray-800 border border-gray-700",
					Children: []core.Node{
						ui.CardContent(ui.CardContentProps{
							Class: "p-0",
							Children: []core.Node{
								ui.DataTable(ui.DataTableProps[dto.AuditEntryList]{
									Data:      filtered,
									Striped:   true,
									Hoverable: true,
									Class:     "bg-gray-800",
									Columns: []ui.ColumnDef[dto.AuditEntryList]{
										{
											Header: "Time",
											Render: func(e dto.AuditEntryList) core.Node {
												return core.Span(core.Class("text-gray-300"),
													core.Text(formatAuditTime(e.CreatedAt)),
												)
											},
										},
										{
											Header: "Action",
											Render: func(e dto.AuditEntryList) core.Node {
												return auditActionBadge(e.Action)
											},
										},
										{
											Header: "Entity",
											Render: func(e dto.AuditEntryList) core.Node {
												return core.Span(core.Class("text-gray-300 font-medium"),
													core.Text(fmt.Sprintf("%s #%d", e.EntityType, e.EntityID)),
												)
											},
										},
										{
											Header: "User",
											Render: func(e dto.AuditEntryList) core.Node {
												if e.UserEmail == "" {
													return core.Span(core.Class("text-gray-500 italic"),
														core.Text("System"),
													)
												}
												return core.Span(core.Class("text-gray-300"),
													core.Text(e.UserEmail),
												)
											},
										},
										{
											Header: "IP Address",
											Render: func(e dto.AuditEntryList) core.Node {
												return core.Span(core.Class("text-gray-400 font-mono text-sm"),
													core.Text(e.IPAddress),
												)
											},
										},
										{
											Header: "Changes",
											Render: func(e dto.AuditEntryList) core.Node {
												return formatChanges(e.Changes)
											},
										},
									},
								}),
							},
						}),
					},
				})
			}(),

			core.Div(core.Class("mt-4 text-right text-gray-500 text-sm"),
				core.Text(fmt.Sprintf("Showing %d entries", len(filtered))),
			),
		)
	}
}

func filterAuditEntries(entries []dto.AuditEntryList, action, entity string) []dto.AuditEntryList {
	if action == "" && entity == "" {
		return entries
	}
	var filtered []dto.AuditEntryList
	for _, e := range entries {
		matchesAction := action == "" || e.Action == action
		matchesEntity := entity == "" || e.EntityType == entity
		if matchesAction && matchesEntity {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func formatAuditTime(t time.Time) string {
	return t.Format("Jan 2, 2006 15:04:05")
}

func formatChanges(changesJSON string) core.Node {
	if changesJSON == "" || changesJSON == "{}" {
		return core.Span(core.Class("text-gray-500 italic"),
			core.Text("No changes recorded"),
		)
	}

	// Try to parse and display nicely
	var changes map[string]interface{}
	if err := json.Unmarshal([]byte(changesJSON), &changes); err != nil {
		return core.Span(core.Class("text-gray-400 font-mono text-xs"),
			core.Text(truncate(changesJSON, 80)),
		)
	}

	var items []core.Node
	for key, val := range changes {
		items = append(items, formatChangeField(key, val))
	}

	if len(items) == 0 {
		return core.Span(core.Class("text-gray-500 italic"),
			core.Text("No changes"),
		)
	}

	return core.Div(core.Class("space-y-1"), items...)
}

func formatChangeField(key string, val interface{}) core.Node {
	// Check if it's a diff (has "old" and "new" keys)
	if m, ok := val.(map[string]interface{}); ok {
		if oldVal, hasOld := m["old"]; hasOld {
			newVal := m["new"]
			return core.Div(core.Class("text-xs"),
				core.Span(core.Class("font-medium text-gray-300"),
					core.Text(key+": "),
				),
				core.Span(core.Class("text-red-400 line-through"),
					core.Text(fmt.Sprintf("%v", oldVal)),
				),
				core.Span(core.Class("text-gray-500 mx-1"),
					core.Text("\u2192"),
				),
				core.Span(core.Class("text-green-400"),
					core.Text(fmt.Sprintf("%v", newVal)),
				),
			)
		}
	}

	// Simple value (from create snapshot)
	return core.Div(core.Class("text-xs"),
		core.Span(core.Class("font-medium text-gray-300"),
			core.Text(key+": "),
		),
		core.Span(core.Class("text-gray-400"),
			core.Text(fmt.Sprintf("%v", val)),
		),
	)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Ensure imports are used
var _ = strings.ToLower
