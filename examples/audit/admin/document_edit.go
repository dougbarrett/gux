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

// DocumentEdit is the edit document form.
func DocumentEdit(r *core.Router) func() core.Node {
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
		title := r.StateString("title", "")
		content := r.StateString("content", "")
		status := r.StateString("status", "")
		saving := r.StateBool("saving", false)
		errorMsg := r.StateString("error", "")
		initialized := r.StateBool("initialized", false)

		var currentDoc *dto.DocumentDetail
		json.Unmarshal([]byte(docState.Get()), &currentDoc)

		// Initialize form fields from loaded document
		if currentDoc != nil && currentDoc.ID != 0 && !initialized.Get() {
			title.Set(currentDoc.Title)
			content.Set(currentDoc.Content)
			status.Set(currentDoc.Status)
			initialized.Set(true)
		}

		if currentDoc == nil || currentDoc.ID == 0 {
			return pages.AdminLayout(r, "Edit Document",
				core.Div(core.Class("text-center text-gray-400 py-12"),
					core.Text("Document not found"),
				),
			)
		}

		handleSubmit := func() {
			if saving.Get() {
				return
			}
			if title.Get() == "" {
				errorMsg.Set("Title is required")
				return
			}
			saving.Set(true)
			errorMsg.Set("")

			data := map[string]interface{}{
				"title":   title.Get(),
				"content": content.Get(),
				"status":  status.Get(),
			}

			var docID uint
			fmt.Sscanf(id, "%d", &docID)
			api.Documents.Update(docID, data, func(result *dto.DocumentDetail, err error) {
				saving.Set(false)
				if err != nil {
					errorMsg.Set("Failed to update document")
					return
				}
				r.Navigate(fmt.Sprintf("/admin/documents/%s", id))
			})
		}

		return pages.AdminLayout(r, "Edit Document",
			core.Div(core.Class("space-y-6"),
				core.Div(core.Class("flex justify-between items-center"),
					core.H1(core.Class("text-2xl font-bold text-white"),
						core.Text(fmt.Sprintf("Edit Document #%d", currentDoc.ID)),
					),
					ui.Button(ui.ButtonProps{
						Variant:  ui.ButtonSecondary,
						Children: []core.Node{core.Text("Cancel")},
						OnClick: func() {
							r.Navigate(fmt.Sprintf("/admin/documents/%s", id))
						},
					}),
				),
				core.If(errorMsg.Get() != "",
					core.Div(core.Class("bg-red-900/50 border border-red-700 text-red-200 px-4 py-3 rounded"),
						core.Text(errorMsg.Get()),
					),
				),
				ui.Card(ui.CardProps{
					Class: "bg-gray-800 border border-gray-700",
					Children: []core.Node{
						ui.CardContent(ui.CardContentProps{
							Children: []core.Node{
								core.Div(core.Class("space-y-6"),
									core.Div(core.Attrs{},
										core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
											core.Text("Title"),
										),
										ui.Input(ui.InputProps{
											ID:          "title",
											Placeholder: "Enter document title",
											Value:       title.Get(),
											OnChange:    func(val string) { title.Set(val) },
											Disabled:    saving.Get(),
										}),
									),
									core.Div(core.Attrs{},
										core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
											core.Text("Content"),
										),
										ui.Input(ui.InputProps{
											ID:          "content",
											Placeholder: "Enter document content",
											Value:       content.Get(),
											OnChange:    func(val string) { content.Set(val) },
											Disabled:    saving.Get(),
										}),
									),
									core.Div(core.Attrs{},
										core.Label(core.Class("block text-sm font-medium text-gray-300 mb-2"),
											core.Text("Status"),
										),
										ui.Select(ui.SelectProps{
											ID:    "status",
											Value: status.Get(),
											OnChange: func(val string) { status.Set(val) },
											Options: []ui.SelectOption{
												{Value: "draft", Label: "Draft"},
												{Value: "published", Label: "Published"},
												{Value: "archived", Label: "Archived"},
											},
											Disabled: saving.Get(),
										}),
									),
									core.Div(core.Class("flex justify-end pt-4"),
										ui.Button(ui.ButtonProps{
											Variant:  ui.ButtonPrimary,
											Disabled: saving.Get(),
											Children: []core.Node{
												func() core.Node {
													if saving.Get() {
														return core.Text("Saving...")
													}
													return core.Text("Save Changes")
												}(),
											},
											OnClick: handleSubmit,
										}),
									),
								),
							},
						}),
					},
				}),
			),
		)
	}
}
