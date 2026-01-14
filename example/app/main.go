//go:build js && wasm

package main

import (
	"context"
	"syscall/js"

	"goquery/components"
	"goquery/example/api"
)

var (
	app     *components.App
	layout  *components.Layout
	display *components.DataDisplay
	router  *components.Router
	posts   *api.PostsClient
	modal   *components.Modal
)

func main() {
	// Initialize app (loads Tailwind, clears #app)
	app = components.NewApp("app")

	// Initialize API client
	posts = api.NewPostsClient()

	// Create router
	router = components.NewRouter()
	components.SetGlobalRouter(router)

	router.Register("/", showDashboard)
	router.Register("/api-test", showAPITest)
	router.Register("/components", showComponents)
	router.Register("/settings", showSettings)

	// Update sidebar active state on navigation
	router.OnNavigate(func(path string) {
		if layout != nil {
			layout.Sidebar().SetActive(path)
		}
	})

	// Create data display (used by API Test page)
	display = components.NewDataDisplay()

	// Create layout
	layout = components.NewLayout(components.LayoutProps{
		Sidebar: components.SidebarProps{
			Title: "Admin Panel",
			Items: []components.NavItem{
				{Label: "Dashboard", Icon: "📊", Path: "/"},
				{Label: "API Test", Icon: "🔌", Path: "/api-test"},
				{Label: "Components", Icon: "🧩", Path: "/components"},
				{Label: "Settings", Icon: "⚙️", Path: "/settings"},
			},
		},
		Header: components.HeaderProps{
			Title: "Dashboard",
			Actions: []components.HeaderAction{
				{Label: "Refresh", OnClick: func() { js.Global().Get("location").Call("reload") }},
			},
		},
	})

	app.Mount(layout.Element())

	// Start router
	router.Start()

	// Keep Go running
	app.Run()
}

func showDashboard() {
	layout.SetContent(
		components.Div("space-y-4",
			components.Card(
				components.H2("Welcome to the Admin Dashboard"),
				components.Text("This is a proof of concept admin dashboard built entirely with Go WASM. Select an option from the sidebar to explore."),
			),
			components.Div("grid grid-cols-1 md:grid-cols-3 gap-4",
				statCard("Total Posts", "3", components.BadgePrimary),
				statCard("Active Users", "42", components.BadgeSuccess),
				statCard("Pending Tasks", "7", components.BadgeWarning),
			),
		),
	)
}

func statCard(label, value string, variant components.BadgeVariant) js.Value {
	return components.Card(
		components.Div("flex justify-between items-center",
			components.Text(label),
			components.Badge(components.BadgeProps{Text: value, Variant: variant, Rounded: true}),
		),
		components.HeadingWithClass(3, value, "text-3xl font-bold text-gray-800 mt-2"),
	)
}

func showAPITest() {
	layout.SetContent(
		components.Card(
			components.H2("API Test"),
			components.Div("flex gap-2 mb-4",
				components.Button(components.ButtonProps{
					Text: "Fetch Post #1",
					OnClick: func() {
						go fetchSinglePost()
					},
				}),
				components.Button(components.ButtonProps{
					Text:      "Fetch All Posts",
					ClassName: "px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600 cursor-pointer transition-colors",
					OnClick: func() {
						go fetchAllPosts()
					},
				}),
			),
			display.Element(),
		),
	)
}

func showComponents() {
	// Create tabs for component showcase
	tabs := components.NewTabs(components.TabsProps{
		Tabs: []components.Tab{
			{Label: "Forms", Content: formsDemo()},
			{Label: "Feedback", Content: feedbackDemo()},
			{Label: "Data", Content: dataDemo()},
		},
	})

	layout.SetContent(
		components.Card(
			components.H2("Component Showcase"),
			components.Text("Explore the available UI components."),
			components.Div("mt-4", tabs.Element()),
		),
	)
}

func formsDemo() js.Value {
	nameInput := components.NewInput(components.InputProps{
		Label:       "Name",
		Placeholder: "Enter your name",
	})

	emailInput := components.NewInput(components.InputProps{
		Type:        components.InputEmail,
		Label:       "Email",
		Placeholder: "you@example.com",
	})

	bioTextArea := components.NewTextArea(components.TextAreaProps{
		Label:       "Bio",
		Placeholder: "Tell us about yourself...",
		Rows:        3,
	})

	roleSelect := components.NewSelect(components.SelectProps{
		Label:       "Role",
		Placeholder: "Select a role",
		Options: []components.SelectOption{
			{Label: "Admin", Value: "admin"},
			{Label: "Editor", Value: "editor"},
			{Label: "Viewer", Value: "viewer"},
		},
	})

	notifyCheckbox := components.NewCheckbox(components.CheckboxProps{
		Label:   "Receive email notifications",
		Checked: true,
	})

	return components.Div("space-y-4",
		nameInput.Element(),
		emailInput.Element(),
		bioTextArea.Element(),
		roleSelect.Element(),
		notifyCheckbox.Element(),
		components.Button(components.ButtonProps{
			Text: "Submit Form",
			OnClick: func() {
				js.Global().Call("alert", "Name: "+nameInput.Value()+"\nEmail: "+emailInput.Value())
			},
		}),
	)
}

func feedbackDemo() js.Value {
	// Create modal
	modal = components.NewModal(components.ModalProps{
		Title: "Confirm Action",
		Content: components.Div("",
			components.Text("Are you sure you want to proceed with this action?"),
		),
		Footer: components.Div("flex justify-end gap-2",
			components.Button(components.ButtonProps{
				Text:      "Cancel",
				ClassName: "px-4 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300 cursor-pointer transition-colors",
				OnClick: func() {
					modal.Close()
				},
			}),
			components.Button(components.ButtonProps{
				Text: "Confirm",
				OnClick: func() {
					modal.Close()
				},
			}),
		),
		CloseOnEsc: true,
	})

	return components.Div("space-y-4",
		components.H3("Alerts"),
		components.AlertInfoMsg("This is an informational message."),
		components.AlertSuccessMsg("Operation completed successfully!"),
		components.AlertWarningMsg("Please review before continuing."),
		components.AlertErrorMsg("An error occurred while processing."),

		components.H3("Badges"),
		components.Div("flex flex-wrap gap-2",
			components.BadgeText("Default"),
			components.BadgePrimaryText("Primary"),
			components.BadgeSuccessText("Success"),
			components.BadgeWarningText("Warning"),
			components.BadgeErrorText("Error"),
			components.BadgeInfoText("Info"),
		),

		components.H3("Spinner"),
		components.Div("flex items-center gap-4",
			components.Spinner(components.SpinnerProps{Size: components.SpinnerSM}),
			components.Spinner(components.SpinnerProps{Size: components.SpinnerMD}),
			components.Spinner(components.SpinnerProps{Size: components.SpinnerLG, Label: "Loading..."}),
		),

		components.H3("Modal"),
		components.Button(components.ButtonProps{
			Text: "Open Modal",
			OnClick: func() {
				modal.Open()
			},
		}),
		modal.Element(),
	)
}

func dataDemo() js.Value {
	// Sample table data
	tableData := []map[string]any{
		{"id": 1, "name": "John Doe", "email": "john@example.com", "status": "Active"},
		{"id": 2, "name": "Jane Smith", "email": "jane@example.com", "status": "Pending"},
		{"id": 3, "name": "Bob Wilson", "email": "bob@example.com", "status": "Inactive"},
	}

	table := components.NewTable(components.TableProps{
		Columns: []components.TableColumn{
			{Header: "ID", Key: "id", Width: "60px"},
			{Header: "Name", Key: "name"},
			{Header: "Email", Key: "email"},
			{Header: "Status", Key: "status", Render: func(row map[string]any, value any) js.Value {
				status := value.(string)
				var variant components.BadgeVariant
				switch status {
				case "Active":
					variant = components.BadgeSuccess
				case "Pending":
					variant = components.BadgeWarning
				case "Inactive":
					variant = components.BadgeError
				}
				return components.Badge(components.BadgeProps{Text: status, Variant: variant})
			}},
		},
		Data:      tableData,
		Striped:   true,
		Hoverable: true,
	})

	return components.Div("space-y-4",
		components.H3("Table"),
		table.Element(),
	)
}

func showSettings() {
	nameInput := components.NewInput(components.InputProps{
		Label: "Display Name",
		Value: "Admin User",
	})

	themeSelect := components.NewSelect(components.SelectProps{
		Label: "Theme",
		Value: "light",
		Options: []components.SelectOption{
			{Label: "Light", Value: "light"},
			{Label: "Dark", Value: "dark"},
			{Label: "System", Value: "system"},
		},
	})

	notifyCheckbox := components.NewCheckbox(components.CheckboxProps{
		Label:   "Enable notifications",
		Checked: true,
	})

	layout.SetContent(
		components.Card(
			components.H2("Settings"),
			components.Div("space-y-4 max-w-md",
				nameInput.Element(),
				themeSelect.Element(),
				notifyCheckbox.Element(),
				components.Div("pt-4",
					components.Button(components.ButtonProps{
						Text: "Save Settings",
						OnClick: func() {
							js.Global().Call("alert", "Settings saved!")
						},
					}),
				),
			),
		),
	)
}

func fetchSinglePost() {
	display.ShowLoading("Fetching post #1...")

	post, err := posts.GetByID(context.Background(), 1)
	if err != nil {
		display.ShowError("Error: " + err.Error())
		return
	}

	display.ShowJSON(post)
}

func fetchAllPosts() {
	display.ShowLoading("Fetching all posts...")

	allPosts, err := posts.GetAll(context.Background())
	if err != nil {
		display.ShowError("Error: " + err.Error())
		return
	}

	// Show first 5 posts
	if len(allPosts) > 5 {
		allPosts = allPosts[:5]
	}
	display.ShowJSON(allPosts)
}
