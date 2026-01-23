package pages

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/admin/.gux/api"
	"github.com/dougbarrett/gux/examples/admin/dto"
	"github.com/dougbarrett/gux/ui"
)

// Activity is the activity logs page with filtering capabilities.
func Activity(r *core.Router) func() core.Node {
	// Store logs as JSON string for state hydration
	var logsData []dto.ActivityLogList

	r.OnLoad(func() {
		api.ActivityLogs.List(func(result []dto.ActivityLogList, err error) {
			if err == nil {
				logsData = result
			}
		})
	})

	return func() core.Node {
		// State for logs (stored as JSON string for array)
		logsJSON := r.StateString("logsJSON", func() string {
			data, _ := json.Marshal(logsData)
			return string(data)
		}())

		// Filter state
		actionFilter := r.StateString("actionFilter", "")
		entityFilter := r.StateString("entityFilter", "")

		// Parse logs from JSON
		var logs []dto.ActivityLogList
		json.Unmarshal([]byte(logsJSON.Get()), &logs)

		// Apply filters
		filteredLogs := filterLogs(logs, actionFilter.Get(), entityFilter.Get())

		return AdminLayout(
			core.H1(core.Class("text-2xl font-bold text-white mb-6"),
				core.Text("Activity Logs"),
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
									// Action filter
									core.Div(core.Attrs{},
										core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
											core.Text("Filter by Action"),
										),
										ui.Select(ui.SelectProps{
											Name:        "actionFilter",
											Value:       actionFilter.Get(),
											Placeholder: "All Actions",
											Options: []ui.SelectOption{
												{Value: "", Label: "All Actions"},
												{Value: "created", Label: "Created"},
												{Value: "updated", Label: "Updated"},
												{Value: "deleted", Label: "Deleted"},
												{Value: "login", Label: "Login"},
												{Value: "logout", Label: "Logout"},
											},
											OnChange: func(v string) { actionFilter.Set(v) },
											Class:    "bg-gray-700 border-gray-600 text-white min-w-[150px]",
										}),
									),

									// Entity filter
									core.Div(core.Attrs{},
										core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
											core.Text("Filter by Entity"),
										),
										ui.Select(ui.SelectProps{
											Name:        "entityFilter",
											Value:       entityFilter.Get(),
											Placeholder: "All Entities",
											Options: []ui.SelectOption{
												{Value: "", Label: "All Entities"},
												{Value: "user", Label: "User"},
												{Value: "setting", Label: "Setting"},
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

			// Activity logs table or empty state
			func() core.Node {
				if len(filteredLogs) == 0 {
					return ui.Card(ui.CardProps{
						Class: "bg-gray-800 border border-gray-700",
						Children: []core.Node{
							ui.CardContent(ui.CardContentProps{
								Children: []core.Node{
									core.Div(core.Class("text-center py-8 text-gray-400"),
										core.Text("No activity logs found"),
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
								ui.DataTable(ui.DataTableProps[dto.ActivityLogList]{
									Data:      filteredLogs,
									Striped:   true,
									Hoverable: true,
									Class:     "bg-gray-800",
									Columns: []ui.ColumnDef[dto.ActivityLogList]{
										{
											Header: "Time",
											Render: func(log dto.ActivityLogList) core.Node {
												return core.Span(core.Class("text-gray-300"),
													core.Text(formatLogTime(log.CreatedAt)),
												)
											},
										},
										{
											Header: "User",
											Render: func(log dto.ActivityLogList) core.Node {
												userName := log.UserName
												if userName == "" {
													userName = "System"
												}
												return core.Span(core.Class("text-gray-300"),
													core.Text(userName),
												)
											},
										},
										{
											Header: "Action",
											Render: func(log dto.ActivityLogList) core.Node {
												return actionBadge(log.Action)
											},
										},
										{
											Header: "Entity",
											Render: func(log dto.ActivityLogList) core.Node {
												return core.Span(core.Class("text-gray-300"),
													core.Text(capitalizeFirst(log.Entity)),
												)
											},
										},
										{
											Header: "Description",
											Render: func(log dto.ActivityLogList) core.Node {
												return core.Span(core.Class("text-gray-400"),
													core.Text(log.Description),
												)
											},
										},
									},
								}),
							},
						}),
					},
				})
			}(),
		)
	}
}

// filterLogs applies action and entity filters to the logs slice.
func filterLogs(logs []dto.ActivityLogList, actionFilter, entityFilter string) []dto.ActivityLogList {
	if actionFilter == "" && entityFilter == "" {
		return logs
	}

	var filtered []dto.ActivityLogList
	for _, log := range logs {
		matchesAction := actionFilter == "" || log.Action == actionFilter
		matchesEntity := entityFilter == "" || log.Entity == entityFilter
		if matchesAction && matchesEntity {
			filtered = append(filtered, log)
		}
	}
	return filtered
}

// actionBadge returns a colored badge for the action type.
func actionBadge(action string) core.Node {
	var variant ui.BadgeVariant
	switch action {
	case "created":
		variant = ui.BadgeSuccess
	case "updated":
		variant = ui.BadgeDefault
	case "deleted":
		variant = ui.BadgeError
	case "login":
		variant = ui.BadgeSuccess
	case "logout":
		variant = ui.BadgeWarning
	default:
		variant = ui.BadgeDefault
	}

	return ui.Badge(ui.BadgeProps{
		Variant: variant,
		Text:    capitalizeFirst(action),
	})
}

// formatLogTime formats a time.Time as "Jan 2, 15:04".
func formatLogTime(t time.Time) string {
	return t.Format("Jan 2, 15:04")
}

// capitalizeFirst capitalizes the first letter of a string.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
