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

// DocumentDetail shows a single document with edit/delete options.
func DocumentDetail(r *core.Router) func() core.Node {
	id := r.Param("id")
	var document *dto.DocumentDetail

	r.OnLoad(func() {
		var docID uint
		fmt.Sscanf(id, "%d", &docID)
		if docID > 0 {
			api.Documents.Get(docID, func(result *dto.DocumentDetail, err error) {
				if err == nil {
					document = result
				}
			})
		}
	})

	return func() core.Node {
		docJSON, _ := json.Marshal(document)
		docState := r.StateString("document", string(docJSON))
		deleting := r.StateBool("deleting", false)

		var currentDoc *dto.DocumentDetail
		json.Unmarshal([]byte(docState.Get()), &currentDoc)

		if currentDoc == nil || currentDoc.ID == 0 {
			return pages.AdminLayout(r, "Document",
				core.Div(core.Class("text-center text-gray-400 py-12"),
					core.Text("Document not found"),
				),
			)
		}

		statusColor := "bg-gray-700 text-gray-300"
		if currentDoc.Status == "published" {
			statusColor = "bg-green-900/50 text-green-200"
		} else if currentDoc.Status == "archived" {
			statusColor = "bg-gray-600 text-gray-400"
		}

		return pages.AdminLayout(r, "Document Detail",
			core.Div(core.Class("space-y-6"),
				core.Div(core.Class("flex justify-between items-start"),
					core.Div(core.Attrs{},
						core.H1(core.Class("text-2xl font-bold text-white mb-2"),
							core.Text(currentDoc.Title),
						),
						core.Span(core.Class("px-2 py-1 text-xs rounded font-medium "+statusColor),
							core.Text(currentDoc.Status),
						),
					),
					core.Div(core.Class("flex space-x-3"),
						core.A(
							core.Attrs{
								Href:     fmt.Sprintf("/admin/documents/%d/edit", currentDoc.ID),
								Class:    "px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition",
								External: true,
							},
							core.Text("Edit"),
						),
						ui.Button(ui.ButtonProps{
							Variant:  ui.ButtonDestructive,
							Disabled: deleting.Get(),
							Children: []core.Node{
								func() core.Node {
									if deleting.Get() {
										return core.Text("Deleting...")
									}
									return core.Text("Delete")
								}(),
							},
							OnClick: func() {
								if !deleting.Get() {
									deleting.Set(true)
									var docID uint
									fmt.Sscanf(id, "%d", &docID)
									api.Documents.Delete(docID, func(err error) {
										deleting.Set(false)
										if err == nil {
											r.Navigate("/admin/documents")
										}
									})
								}
							},
						}),
					),
				),

				ui.Card(ui.CardProps{
					Class: "bg-gray-800 border border-gray-700",
					Children: []core.Node{
						ui.CardContent(ui.CardContentProps{
							Children: []core.Node{
								detailRow("ID", fmt.Sprintf("%d", currentDoc.ID)),
								detailRow("Title", currentDoc.Title),
								core.Div(core.Class("py-3 border-b border-gray-700"),
									core.Span(core.Class("block text-sm font-medium text-gray-400 mb-2"),
										core.Text("Content"),
									),
									core.Div(core.Class("text-gray-300 whitespace-pre-wrap bg-gray-900/50 rounded p-4 border border-gray-700"),
										core.Text(currentDoc.Content),
									),
								),
								detailRow("Status", currentDoc.Status),
								detailRow("Author ID", fmt.Sprintf("%d", currentDoc.AuthorID)),
								detailRow("Created", currentDoc.CreatedAt.Format("January 2, 2006 at 3:04 PM")),
								detailRow("Updated", currentDoc.UpdatedAt.Format("January 2, 2006 at 3:04 PM")),
							},
						}),
					},
				}),
			),
		)
	}
}

func detailRow(label, value string) core.Node {
	return core.Div(core.Class("flex py-3 border-b border-gray-700 last:border-0"),
		core.Span(core.Class("w-32 text-gray-400 font-medium"),
			core.Text(label),
		),
		core.Span(core.Class("text-white"),
			core.Text(value),
		),
	)
}
