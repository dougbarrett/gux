package admin

import (
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/audit/dto"
	"github.com/dougbarrett/gux/examples/audit/guxgen/api"
	"github.com/dougbarrett/gux/examples/audit/pages"
	"github.com/dougbarrett/gux/ui"
)

// DocumentNew is the create document form.
func DocumentNew(r *core.Router) func() core.Node {
	title := r.StateString("title", "")
	content := r.StateString("content", "")
	status := r.StateString("status", "draft")
	saving := r.StateBool("saving", false)
	errorMsg := r.StateString("error", "")

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

		user := r.User()
		authorID := uint(0)
		if user != nil && user.ID != "" {
			var id uint
			fmt.Sscanf(user.ID, "%d", &id)
			authorID = id
		}

		data := map[string]interface{}{
			"title":     title.Get(),
			"content":   content.Get(),
			"status":    status.Get(),
			"author_id": authorID,
		}

		api.Documents.Create(data, func(result *dto.DocumentDetail, err error) {
			saving.Set(false)
			if err != nil {
				errorMsg.Set("Failed to create document")
				return
			}
			if result != nil {
				r.Navigate(fmt.Sprintf("/admin/documents/%d", result.ID))
			} else {
				r.Navigate("/admin/documents")
			}
		})
	}

	return func() core.Node {
		return pages.AdminLayout(r, "New Document",
			core.Div(core.Class("space-y-6"),
				core.Div(core.Class("flex justify-between items-center"),
					core.H1(core.Class("text-2xl font-bold text-white"),
						core.Text("New Document"),
					),
					ui.Button(ui.ButtonProps{
						Variant:  ui.ButtonSecondary,
						Children: []core.Node{core.Text("Cancel")},
						OnClick: func() {
							r.Navigate("/admin/documents")
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
														return core.Text("Creating...")
													}
													return core.Text("Create Document")
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
