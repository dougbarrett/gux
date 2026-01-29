package admin

import (
	"encoding/json"
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/audit/dto"
	"github.com/dougbarrett/gux/examples/audit/guxgen/api"
	"github.com/dougbarrett/gux/examples/audit/pages"
	"github.com/dougbarrett/gux/ui"
)

// Documents lists all documents.
func Documents(r *core.Router) func() core.Node {
	var documents []dto.DocumentList

	r.OnLoad(func() {
		api.Documents.List(func(result []dto.DocumentList, err error) {
			if err == nil {
				documents = result
			}
		})
	})

	return func() core.Node {
		docsJSON, _ := json.Marshal(documents)
		docsState := r.StateString("documents", string(docsJSON))

		var displayDocs []dto.DocumentList
		json.Unmarshal([]byte(docsState.Get()), &displayDocs)

		return pages.AdminLayout(r, "Documents",
			core.Div(core.Class("flex justify-between items-center mb-6"),
				core.H1(core.Class("text-3xl font-bold text-white"),
					core.Text("Documents"),
				),
				core.A(
					core.Attrs{Href: "/admin/documents/new", Class: "inline-flex items-center px-4 py-2 bg-green-600 hover:bg-green-700 text-white font-medium rounded-lg transition", External: true},
					core.Text("+ New Document"),
				),
			),

			ui.Card(ui.CardProps{
				Class: "bg-gray-800 border border-gray-700",
				Children: []core.Node{
					ui.CardContent(ui.CardContentProps{
						Class: "p-0",
						Children: []core.Node{
							ui.DataTable(ui.DataTableProps[dto.DocumentList]{
								Data:      displayDocs,
								Striped:   true,
								Hoverable: true,
								Class:     "bg-gray-800",
								Columns: []ui.ColumnDef[dto.DocumentList]{
									{
										Header: "ID",
										Render: func(doc dto.DocumentList) core.Node {
											return core.Span(core.Class("text-gray-400"),
												core.Text(fmt.Sprintf("%d", doc.ID)),
											)
										},
									},
									{
										Header: "Title",
										Render: func(doc dto.DocumentList) core.Node {
											return core.A(
												core.Attrs{
													Href:     fmt.Sprintf("/admin/documents/%d", doc.ID),
													Class:    "text-blue-400 hover:text-blue-300 font-medium",
													External: true,
												},
												core.Text(doc.Title),
											)
										},
									},
									{
										Header: "Status",
										Render: func(doc dto.DocumentList) core.Node {
											return docStatusBadge(doc.Status)
										},
									},
									{
										Header: "Created",
										Render: func(doc dto.DocumentList) core.Node {
											return core.Span(core.Class("text-gray-400"),
												core.Text(doc.CreatedAt.Format("Jan 2, 2006")),
											)
										},
									},
								},
							}),
						},
					}),
				},
			}),
		)
	}
}

func docStatusBadge(status string) core.Node {
	var variant ui.BadgeVariant
	switch status {
	case "published":
		variant = ui.BadgeSuccess
	case "draft":
		variant = ui.BadgeWarning
	case "archived":
		variant = ui.BadgeDefault
	default:
		variant = ui.BadgeDefault
	}
	return ui.Badge(ui.BadgeProps{
		Variant: variant,
		Text:    capitalizeFirst(status),
	})
}
