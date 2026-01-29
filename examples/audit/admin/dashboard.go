package admin

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/audit/dto"
	"github.com/dougbarrett/gux/examples/audit/guxgen/api"
	"github.com/dougbarrett/gux/examples/audit/pages"
	"github.com/dougbarrett/gux/ui"
)

// Dashboard shows overview stats and recent audit entries.
func Dashboard(r *core.Router) func() core.Node {
	var documents []dto.DocumentList
	var auditEntries []dto.AuditEntryList

	r.OnLoad(func() {
		api.Documents.List(func(result []dto.DocumentList, err error) {
			if err == nil {
				documents = result
			}
		})
		api.AuditEntries.List(func(result []dto.AuditEntryList, err error) {
			if err == nil {
				auditEntries = result
			}
		})
	})

	return func() core.Node {
		docsJSON, _ := json.Marshal(documents)
		docsState := r.StateString("documents", string(docsJSON))
		auditJSON, _ := json.Marshal(auditEntries)
		auditState := r.StateString("auditEntries", string(auditJSON))

		var displayDocs []dto.DocumentList
		json.Unmarshal([]byte(docsState.Get()), &displayDocs)
		var displayAudit []dto.AuditEntryList
		json.Unmarshal([]byte(auditState.Get()), &displayAudit)

		// Stats
		totalDocs := len(displayDocs)
		published := 0
		drafts := 0
		for _, d := range displayDocs {
			switch d.Status {
			case "published":
				published++
			case "draft":
				drafts++
			}
		}

		totalDocsState := r.StateInt("totalDocs", totalDocs)
		publishedState := r.StateInt("published", published)
		draftsState := r.StateInt("drafts", drafts)
		totalAuditState := r.StateInt("totalAudit", len(displayAudit))

		// Recent audit entries (up to 5)
		recentAudit := displayAudit
		if len(recentAudit) > 5 {
			recentAudit = recentAudit[:5]
		}

		return pages.AdminLayout(r, "Dashboard",
			core.H1(core.Class("text-3xl font-bold text-white mb-8"),
				core.Text("Dashboard"),
			),

			// Stats grid
			ui.Grid(ui.GridProps{
				Cols: "4",
				Gap:  "6",
				Children: []core.Node{
					dashStatCard("Documents", fmt.Sprintf("%d", totalDocsState.Get()), "text-white"),
					dashStatCard("Published", fmt.Sprintf("%d", publishedState.Get()), "text-green-400"),
					dashStatCard("Drafts", fmt.Sprintf("%d", draftsState.Get()), "text-yellow-400"),
					dashStatCard("Audit Entries", fmt.Sprintf("%d", totalAuditState.Get()), "text-blue-400"),
				},
			}),

			// Recent audit activity
			ui.Card(ui.CardProps{
				Class: "mt-8 bg-gray-800 border border-gray-700",
				Children: []core.Node{
					ui.CardHeader(ui.CardHeaderProps{
						Class: "border-gray-700",
						Children: []core.Node{
							core.H2(core.Class("text-xl font-semibold text-white"),
								core.Text("Recent Audit Activity"),
							),
						},
					}),
					ui.CardContent(ui.CardContentProps{
						Children: []core.Node{
							recentAuditList(recentAudit),
						},
					}),
				},
			}),
		)
	}
}

func dashStatCard(title, value, colorClass string) core.Node {
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

func recentAuditList(entries []dto.AuditEntryList) core.Node {
	if len(entries) == 0 {
		return core.Div(core.Class("text-gray-400 py-4 text-center"),
			core.Text("No audit entries yet"),
		)
	}

	var items []core.Node
	for _, entry := range entries {
		items = append(items, core.Div(core.Class("flex justify-between items-center py-2 border-b border-gray-700 last:border-0"),
			core.Div(core.Class("flex items-center gap-3"),
				auditActionBadge(entry.Action),
				core.Span(core.Class("text-gray-300"),
					core.Text(fmt.Sprintf("%s #%d", entry.EntityType, entry.EntityID)),
				),
			),
			core.Div(core.Class("flex items-center gap-4"),
				core.Span(core.Class("text-gray-400 text-sm"),
					core.Text(entry.UserEmail),
				),
				core.Span(core.Class("text-gray-500 text-sm"),
					core.Text(entry.CreatedAt.Format("Jan 2, 15:04")),
				),
			),
		))
	}
	return core.Div(core.Class("space-y-1"), items...)
}

func auditActionBadge(action string) core.Node {
	var variant ui.BadgeVariant
	switch action {
	case "create":
		variant = ui.BadgeSuccess
	case "update":
		variant = ui.BadgeDefault
	case "delete":
		variant = ui.BadgeError
	default:
		variant = ui.BadgeDefault
	}
	return ui.Badge(ui.BadgeProps{
		Variant: variant,
		Text:    capitalizeFirst(action),
	})
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-32) + s[1:]
}

// suppress unused import
var _ = time.Now
