// client_hooks.go - Custom hooks for Client admin pages
// This file is NOT regenerated and can be safely modified.
package admin

import (
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/dto"
	"github.com/dougbarrett/gux/ui"
)

func init() {
	// Add "Export CSV" button to list page header
	ClientListActions = func(ctx HookContext) []core.Node {
		return []core.Node{
			ui.Button(ui.ButtonProps{
				Variant:  ui.ButtonSecondary,
				Size:     ui.ButtonSM,
				Children: []core.Node{core.Text("Export CSV")},
				OnClick: func() {
					// TODO: Implement export logic
					fmt.Println("Export CSV clicked")
				},
			}),
		}
	}

	// Add "Quick Edit" button to each row
	ClientListRowActions = func(ctx HookContext, item dto.ClientList) []core.Node {
		return []core.Node{
			ui.Button(ui.ButtonProps{
				Variant:  ui.ButtonGhost,
				Size:     ui.ButtonSM,
				Children: []core.Node{core.Text("Quick Edit")},
				OnClick: func() {
					ctx.Router.Navigate(fmt.Sprintf("/admin/clients/%d/edit", item.ID))
				},
			}),
		}
	}

	// Add "Send Email" button to detail page
	ClientDetailActions = func(ctx HookContext, item dto.ClientDetail) []core.Node {
		return []core.Node{
			ui.Button(ui.ButtonProps{
				Variant:  ui.ButtonSecondary,
				Children: []core.Node{core.Text("Send Email")},
				OnClick: func() {
					// TODO: Open email dialog
					fmt.Printf("Send email to: %s\n", item.Email)
				},
			}),
		}
	}

	// Add custom "Activity Log" section to detail page
	ClientDetailSections = func(ctx HookContext, item dto.ClientDetail) []core.Node {
		return []core.Node{
			core.Div(core.Class("mt-6 pt-4 border-t border-gray-200 dark:border-gray-700"),
				core.H3(core.Class("text-lg font-semibold text-gray-900 dark:text-white mb-4"),
					core.Text("Activity Log"),
				),
				core.P(core.Class("text-gray-500 dark:text-gray-400"),
					core.Text("No activity recorded yet."),
				),
			),
		}
	}

	// Example: Validate data before save
	// ClientBeforeSave = func(ctx HookContext, data map[string]any, isEdit bool) error {
	// 	email, ok := data["email"].(string)
	// 	if ok && !strings.Contains(email, "@") {
	// 		return fmt.Errorf("invalid email format")
	// 	}
	// 	return nil
	// }

	// Example: Log after save
	// ClientAfterSave = func(ctx HookContext, id uint, isEdit bool) {
	// 	action := "created"
	// 	if isEdit {
	// 		action = "updated"
	// 	}
	// 	fmt.Printf("Client %d was %s\n", id, action)
	// }
}
