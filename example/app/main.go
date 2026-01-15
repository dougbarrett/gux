//go:build js && wasm

package main

import (
	"syscall/js"
	"time"

	"gux/components"
	"gux/example/api"
	"gux/state"
)

var (
	app            *components.App
	layout         *components.Layout
	display        *components.DataDisplay
	router         *components.Router
	posts          *api.PostsClient
	modal          *components.Modal
	postsStore     *state.AsyncStore[[]api.Post]
	commandPalette *components.CommandPalette
)

func main() {
	// Initialize app (loads Tailwind, clears #app)
	app = components.NewApp("app")

	// Initialize toast notifications
	components.InitToasts()

	// Initialize API client
	posts = api.NewPostsClient()

	// Create router
	router = components.NewRouter()
	components.SetGlobalRouter(router)

	// Initialize async posts store
	postsStore = state.NewAsync[[]api.Post]()

	router.Register("/", showDashboard)
	router.Register("/api-test", showAPITest)
	router.Register("/create-post", showCreatePost)
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

	// Create UserMenu with sample data
	userMenu := components.NewUserMenu(components.UserMenuProps{
		Name:  "Admin User",
		Email: "admin@example.com",
		OnProfile: func() {
			components.Toast("Profile clicked", components.ToastInfo)
		},
		OnSettings: func() {
			router.Navigate("/settings")
		},
		OnLogout: func() {
			components.Toast("Logged out", components.ToastSuccess)
		},
	})

	// Create NotificationCenter with sample notifications
	notifications := []components.Notification{
		{ID: "1", Title: "New Post Created", Message: "A new post 'Hello World' was published", Time: "2 min ago", Type: "success"},
		{ID: "2", Title: "System Update", Message: "Version 2.0 is now available", Time: "1 hour ago", Type: "info"},
		{ID: "3", Title: "Warning", Message: "Storage is almost full", Time: "3 hours ago", Type: "warning", Read: true},
	}

	notificationCenter := components.NewNotificationCenter(components.NotificationCenterProps{
		Notifications: notifications,
		OnMarkRead: func(id string) {
			components.Toast("Marked notification "+id+" as read", components.ToastInfo)
		},
		OnMarkAllRead: func() {
			components.Toast("All notifications marked as read", components.ToastSuccess)
		},
		OnClear: func() {
			components.Toast("Notifications cleared", components.ToastInfo)
		},
		OnNotificationClick: func(id string) {
			components.Toast("Clicked notification "+id, components.ToastInfo)
		},
	})

	// Create layout
	layout = components.NewLayout(components.LayoutProps{
		Sidebar: components.SidebarProps{
			Title: "Admin Panel",
			Items: []components.NavItem{
				{Label: "Dashboard", Icon: "📊", Path: "/"},
				{Label: "API Test", Icon: "🔌", Path: "/api-test"},
				{Label: "Create Post", Icon: "✏️", Path: "/create-post"},
				{Label: "Components", Icon: "🧩", Path: "/components"},
				{Label: "Settings", Icon: "⚙️", Path: "/settings"},
			},
		},
		Header: components.HeaderProps{
			Title:              "Dashboard",
			NotificationCenter: notificationCenter,
			UserMenu:           userMenu,
			Actions: []components.HeaderAction{
				{Label: "Refresh", OnClick: func() { js.Global().Get("location").Call("reload") }},
			},
		},
	})

	app.Mount(layout.Element())

	// Register keyboard shortcut for sidebar collapse (Cmd/Ctrl+B)
	layout.Sidebar().RegisterKeyboardShortcut()

	// Create command palette with navigation and action commands
	commandPalette = components.NewCommandPalette(components.CommandPaletteProps{
		Placeholder:  "Search commands...",
		EmptyMessage: "No commands found",
		Commands:     getCommandPaletteCommands(),
	})

	// Mount command palette to document body
	js.Global().Get("document").Get("body").Call("appendChild", commandPalette.Element())

	// Register global Cmd+K / Ctrl+K shortcut
	commandPalette.RegisterKeyboardShortcut()

	// Start router
	router.Start()

	// Keep Go running
	app.Run()
}

func showDashboard() {
	layout.SetContent(
		components.Div("space-y-4",
			components.TitledCard("Welcome to the Admin Dashboard",
				"This is a proof of concept admin dashboard built entirely with Go WASM. Select an option from the sidebar to explore.",
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
	layout.SetPage("API Test", "",
		components.Div("flex gap-2 mb-4",
			components.PrimaryButton("Fetch Post #1", func() { go fetchSinglePost() }),
			components.SuccessButton("Fetch All Posts", func() { go fetchAllPosts() }),
		),
		display.Element(),
	)
}

func showComponents() {
	tabs := components.NewTabs(components.TabsProps{
		Tabs: []components.Tab{
			{Label: "Forms", Content: formsDemo()},
			{Label: "Feedback", Content: feedbackDemo()},
			{Label: "Data", Content: dataDemo()},
			{Label: "New", Content: newComponentsDemo()},
			{Label: "Advanced", Content: advancedDemo()},
			{Label: "Charts", Content: chartsDemo()},
			{Label: "Utilities", Content: utilitiesDemo()},
		},
	})

	layout.SetPage("Component Showcase", "Explore the available UI components.",
		components.Div("mt-4", tabs.Element()),
	)
}

func formsDemo() js.Value {
	nameInput := components.TextInput("Name", "Enter your name")
	emailInput := components.EmailInput("Email", "you@example.com")
	bioTextArea := components.NewTextArea(components.TextAreaProps{
		Label:       "Bio",
		Placeholder: "Tell us about yourself...",
		Rows:        3,
	})
	roleSelect := components.SimpleSelectWithPlaceholder("Role", "Select a role", "Admin", "Editor", "Viewer")
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
		components.PrimaryButton("Submit Form", func() {
			js.Global().Call("alert", "Name: "+nameInput.Value()+"\nEmail: "+emailInput.Value())
		}),
	)
}

func feedbackDemo() js.Value {
	// Create modal
	modal = components.NewModal(components.ModalProps{
		Title:   "Confirm Action",
		Content: components.Text("Are you sure you want to proceed with this action?"),
		Footer: components.Div("flex justify-end gap-2",
			components.SecondaryButton("Cancel", func() { modal.Close() }),
			components.PrimaryButton("Confirm", func() {
				modal.Close()
				components.Toast("Action confirmed!", components.ToastSuccess)
			}),
		),
		CloseOnEsc: true,
	})

	return components.Div("space-y-4",
		components.Section("Toasts",
			components.Div("flex flex-wrap gap-2",
				components.Button(components.ButtonProps{Text: "Info Toast", Variant: components.ButtonInfo, Size: components.ButtonSM, OnClick: func() {
					components.Toast("This is an info message", components.ToastInfo)
				}}),
				components.Button(components.ButtonProps{Text: "Success Toast", Variant: components.ButtonSuccess, Size: components.ButtonSM, OnClick: func() {
					components.Toast("Operation successful!", components.ToastSuccess)
				}}),
				components.Button(components.ButtonProps{Text: "Warning Toast", Variant: components.ButtonWarning, Size: components.ButtonSM, OnClick: func() {
					components.Toast("Please be careful!", components.ToastWarning)
				}}),
				components.Button(components.ButtonProps{Text: "Error Toast", Variant: components.ButtonDanger, Size: components.ButtonSM, OnClick: func() {
					components.Toast("Something went wrong!", components.ToastError)
				}}),
			),
		),

		components.Section("Alerts",
			components.AlertInfoMsg("This is an informational message."),
			components.AlertSuccessMsg("Operation completed successfully!"),
			components.AlertWarningMsg("Please review before continuing."),
			components.AlertErrorMsg("An error occurred while processing."),
		),

		components.Section("Badges",
			components.Div("flex flex-wrap gap-2",
				components.BadgeText("Default"),
				components.BadgePrimaryText("Primary"),
				components.BadgeSuccessText("Success"),
				components.BadgeWarningText("Warning"),
				components.BadgeErrorText("Error"),
				components.BadgeInfoText("Info"),
			),
		),

		components.Section("Spinner",
			components.Div("flex items-center gap-4",
				components.Spinner(components.SpinnerProps{Size: components.SpinnerSM}),
				components.Spinner(components.SpinnerProps{Size: components.SpinnerMD}),
				components.Spinner(components.SpinnerProps{Size: components.SpinnerLG, Label: "Loading..."}),
			),
		),

		components.Section("Modal",
			components.PrimaryButton("Open Modal", func() { modal.Open() }),
			modal.Element(),
		),
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

func newComponentsDemo() js.Value {
	progress := components.NewProgress(components.ProgressProps{
		Value: 65, Variant: components.ProgressPrimary, Label: "Upload Progress",
	})
	progressStriped := components.NewProgress(components.ProgressProps{
		Value: 45, Variant: components.ProgressSuccess, Striped: true, Animated: true,
	})

	avatarGroup := components.AvatarGroup([]components.AvatarProps{
		{Name: "John Doe", Status: "online"},
		{Name: "Jane Smith", Status: "away"},
		{Name: "Bob Wilson", Status: "offline"},
		{Name: "Alice Brown", Status: "busy"},
		{Name: "Charlie Davis"},
	}, 4)

	accordion := components.NewAccordion(components.AccordionProps{
		Items: []components.AccordionItem{
			{Title: "What is Gux?", Content: components.Text("Gux is a Go WebAssembly framework for building web applications entirely in Go.")},
			{Title: "How does it work?", Content: components.Text("It compiles Go code to WebAssembly and provides component APIs for DOM manipulation.")},
			{Title: "Is it production ready?", Content: components.Text("It's a proof of concept demonstrating the capabilities of Go WASM for web development.")},
		},
		AllowMultiple: true,
	})

	datePicker := components.NewDatePicker(components.DatePickerProps{
		Label: "Select Date", Placeholder: "Choose a date",
		OnChange: func(t time.Time) {
			components.Toast("Date selected: "+t.Format("Jan 2, 2006"), components.ToastInfo)
		},
	})

	stepper := components.NewStepper(components.StepperProps{
		Steps: []components.Step{
			{Title: "Account", Description: "Create account"},
			{Title: "Profile", Description: "Set up profile"},
			{Title: "Complete", Description: "All done!"},
		},
		CurrentStep: 1,
	})

	breadcrumbs := components.Breadcrumbs(components.BreadcrumbsProps{
		Items: []components.BreadcrumbItem{
			{Label: "Home", Path: "/"},
			{Label: "Components", Path: "/components"},
			{Label: "New"},
		},
	})

	tooltipBtn := components.WithTooltip(
		components.PrimaryButton("Hover me for tooltip", nil),
		components.TooltipProps{Text: "This is a helpful tooltip!", Position: components.TooltipTop},
	)

	return components.Div("space-y-6",
		components.Section("Breadcrumbs", breadcrumbs),
		components.Section("Progress Bars",
			components.Div("space-y-3", progress.Element(), progressStriped.Element()),
		),
		components.Section("Skeleton Loaders",
			components.Div("flex gap-4",
				components.SkeletonAvatar("w-12 h-12"),
				components.Div("flex-1", components.SkeletonText(2)),
			),
		),
		components.Section("Avatars",
			components.Div("flex items-center gap-4",
				components.Avatar(components.AvatarProps{Name: "John Doe", Size: components.AvatarLG, Status: "online"}),
				components.Avatar(components.AvatarProps{Name: "Jane Smith", Size: components.AvatarMD, Status: "away"}),
				avatarGroup,
			),
		),
		components.Section("Tooltip", tooltipBtn),
		components.Section("Clipboard", components.CopyableText("npm install gux")),
		components.Section("Date Picker", datePicker.Element()),
		components.Section("Accordion", accordion.Element()),
		components.Section("Stepper", stepper.Element()),
	)
}

func advancedDemo() js.Value {
	// Toggle
	toggle := components.SimpleToggle("Dark Mode", false, func(checked bool) {
		if checked {
			components.Toast("Dark mode enabled", components.ToastInfo)
		} else {
			components.Toast("Dark mode disabled", components.ToastInfo)
		}
	})

	// Dropdown
	dropdown := components.ActionDropdown("Actions", []components.DropdownItem{
		{Label: "Edit", Icon: "✏️", OnClick: func() { components.Toast("Edit clicked", components.ToastInfo) }},
		{Label: "Duplicate", Icon: "📋", OnClick: func() { components.Toast("Duplicate clicked", components.ToastInfo) }},
		{Divider: true},
		{Label: "Delete", Icon: "🗑️", OnClick: func() { components.Toast("Delete clicked", components.ToastError) }},
	})

	// Combobox
	combobox := components.NewCombobox(components.ComboboxProps{
		Label:       "Select Framework",
		Placeholder: "Search frameworks...",
		Options: []components.ComboboxOption{
			{Label: "React", Value: "react", Description: "A JavaScript library for building user interfaces"},
			{Label: "Vue", Value: "vue", Description: "The Progressive JavaScript Framework"},
			{Label: "Angular", Value: "angular", Description: "Platform for building mobile and desktop web apps"},
			{Label: "Svelte", Value: "svelte", Description: "Cybernetically enhanced web apps"},
			{Label: "Go WASM", Value: "gowasm", Description: "Build web apps with Go and WebAssembly"},
		},
		OnChange: func(value string) {
			components.Toast("Selected: "+value, components.ToastSuccess)
		},
	})

	// Pagination
	pagination := components.NewPagination(components.PaginationProps{
		CurrentPage:  3,
		TotalPages:   10,
		TotalItems:   97,
		ItemsPerPage: 10,
		ShowInfo:     true,
		OnPageChange: func(page int) {
			components.Toast("Page "+string(rune('0'+page))+" clicked", components.ToastInfo)
		},
	})

	// File Upload
	fileUpload := components.ImageUpload("Upload Images", func(files []components.FileInfo) {
		components.Toast("Uploaded "+string(rune('0'+len(files)))+" file(s)", components.ToastSuccess)
	})

	// Drawer button
	var drawer *components.Drawer
	drawerContent := components.Div("space-y-4",
		components.Text("This is a slide-out drawer panel. Great for settings, filters, or secondary navigation."),
		components.PrimaryButton("Close Drawer", func() {
			if drawer != nil {
				drawer.Close()
			}
		}),
	)
	drawer = components.RightDrawer("Settings Panel", drawerContent)

	return components.Div("space-y-6",
		components.Section("Toggle Switch",
			toggle.Element(),
		),
		components.Section("Dropdown Menu",
			dropdown.Element(),
		),
		components.Section("Combobox / Autocomplete",
			components.Div("max-w-sm", combobox.Element()),
		),
		components.Section("Pagination",
			pagination.Element(),
		),
		components.Section("File Upload",
			fileUpload.Element(),
		),
		components.Section("Drawer Panel",
			components.PrimaryButton("Open Drawer", func() { drawer.Open() }),
		),
	)
}

func utilitiesDemo() js.Value {
	// Theme toggle
	themeToggle := components.ThemeToggle(components.ThemeToggleProps{
		ShowLabel: true,
	})

	// Animation demo elements
	document := js.Global().Get("document")
	animBox := document.Call("createElement", "div")
	animBox.Set("className", "w-16 h-16 bg-blue-500 rounded-lg flex items-center justify-center text-white font-bold")
	animBox.Set("textContent", "Go")

	// WebSocket Echo Demo (raw WebSocket)
	wsStatusText := document.Call("createElement", "span")
	wsStatusText.Set("className", "text-gray-500")
	wsStatusText.Set("textContent", "Disconnected")

	wsMessageLog := document.Call("createElement", "div")
	wsMessageLog.Set("className", "h-32 overflow-y-auto bg-gray-100 dark:bg-gray-800 rounded p-2 text-sm font-mono text-gray-800 dark:text-gray-200 border border-gray-200 dark:border-gray-700")
	wsMessageLog.Set("textContent", "No messages yet...")

	wsInput := document.Call("createElement", "input")
	wsInput.Set("type", "text")
	wsInput.Set("placeholder", "Type a message to echo...")
	wsInput.Set("className", "flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-white rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500")

	var wsStore *state.WebSocketStore

	appendEchoMessage := func(msg string) {
		current := wsMessageLog.Get("innerHTML").String()
		if current == "No messages yet..." {
			wsMessageLog.Set("innerHTML", msg)
		} else {
			wsMessageLog.Set("innerHTML", current+"<br>"+msg)
		}
		wsMessageLog.Set("scrollTop", wsMessageLog.Get("scrollHeight"))
	}

	connectEchoWS := func() {
		wsStore = state.NewWebSocketStore(state.WebSocketConfig{
			URL: "wss://echo.websocket.org",
			OnOpen: func() {
				wsStatusText.Set("className", "text-green-500 font-medium")
				wsStatusText.Set("textContent", "Connected")
				appendEchoMessage("[System] Connected to echo server")
			},
			OnClose: func(code int, reason string) {
				wsStatusText.Set("className", "text-gray-500")
				wsStatusText.Set("textContent", "Disconnected")
				appendEchoMessage("[System] Disconnected")
			},
			OnMessage: func(data []byte) {
				appendEchoMessage("[Echo] " + string(data))
			},
			OnError: func(err string) {
				wsStatusText.Set("className", "text-red-500")
				wsStatusText.Set("textContent", "Error")
				appendEchoMessage("[Error] " + err)
			},
		})
		wsStore.Connect()
	}

	disconnectEchoWS := func() {
		if wsStore != nil {
			wsStore.Close()
			wsStore = nil
		}
	}

	sendEchoMessage := func() {
		if wsStore != nil && wsStore.IsConnected() {
			msg := wsInput.Get("value").String()
			if msg != "" {
				appendEchoMessage("[Sent] " + msg)
				wsStore.Send([]byte(msg))
				wsInput.Set("value", "")
			}
		} else {
			components.Toast("Not connected", components.ToastWarning)
		}
	}

	// Posts Subscribe Demo - First-party API (mirrors HTTP client pattern)
	subStatusText := document.Call("createElement", "span")
	subStatusText.Set("className", "text-gray-500")
	subStatusText.Set("textContent", "Not subscribed")

	subLog := document.Call("createElement", "div")
	subLog.Set("className", "h-32 overflow-y-auto bg-gray-100 dark:bg-gray-800 rounded p-2 text-sm font-mono text-gray-800 dark:text-gray-200 border border-gray-200 dark:border-gray-700")
	subLog.Set("textContent", "No events yet...")

	var postsSub *api.Subscription

	appendSubEvent := func(msg string) {
		current := subLog.Get("innerHTML").String()
		if current == "No events yet..." {
			subLog.Set("innerHTML", msg)
		} else {
			subLog.Set("innerHTML", current+"<br>"+msg)
		}
		subLog.Set("scrollTop", subLog.Get("scrollHeight"))
	}

	startSubscription := func() {
		var err error
		// Simple Subscribe API - just pass a handler function
		postsSub, err = posts.Subscribe(func(event api.PostEvent) {
			switch event.Type {
			case "created":
				appendSubEvent("[Created] " + event.Post.Title)
			case "updated":
				appendSubEvent("[Updated] " + event.Post.Title)
			case "deleted":
				appendSubEvent("[Deleted] Post #" + string(rune('0'+event.ID)))
			}
		})
		if err != nil {
			subStatusText.Set("className", "text-red-500")
			subStatusText.Set("textContent", "Failed")
			appendSubEvent("[Error] " + err.Error())
			components.Toast("Subscribe failed", components.ToastError)
			return
		}
		subStatusText.Set("className", "text-green-500 font-medium")
		subStatusText.Set("textContent", "Subscribed")
		appendSubEvent("[System] Listening for post events...")
		components.Toast("Subscribed to posts!", components.ToastSuccess)
	}

	stopSubscription := func() {
		if postsSub != nil {
			postsSub.Close()
			postsSub = nil
			subStatusText.Set("className", "text-gray-500")
			subStatusText.Set("textContent", "Not subscribed")
			appendSubEvent("[System] Unsubscribed")
		}
	}

	// FormBuilder demo
	formBuilder := components.NewFormBuilder(components.FormBuilderProps{
		Fields: []components.BuilderField{
			{Name: "username", Type: components.BuilderFieldText, Label: "Username", Placeholder: "Enter username", Rules: []components.ValidationRule{components.Required}},
			{Name: "email", Type: components.BuilderFieldEmail, Label: "Email", Placeholder: "you@example.com", Rules: []components.ValidationRule{components.Required, components.Email}},
			{Name: "password", Type: components.BuilderFieldPassword, Label: "Password", Placeholder: "Enter password", Rules: []components.ValidationRule{components.Required, components.MinLength(6)}},
			{Name: "role", Type: components.BuilderFieldSelect, Label: "Role", Placeholder: "Select role", Options: []components.SelectOption{
				{Label: "Admin", Value: "admin"},
				{Label: "Editor", Value: "editor"},
				{Label: "Viewer", Value: "viewer"},
			}},
			{Name: "bio", Type: components.BuilderFieldTextarea, Label: "Bio", Placeholder: "Tell us about yourself...", Rows: 3},
			{Name: "newsletter", Type: components.BuilderFieldCheckbox, Label: "Subscribe to newsletter"},
		},
		SubmitText: "Register",
		ShowCancel: true,
		CancelText: "Reset",
		OnSubmit: func(values map[string]any) error {
			components.Toast("Form submitted!", components.ToastSuccess)
			return nil
		},
		OnCancel: func() {
			components.Toast("Form reset", components.ToastInfo)
		},
	})

	return components.Div("space-y-6",
		components.Section("Theme System",
			components.Div("space-y-4",
				components.Text("Toggle between light and dark mode:"),
				components.Div("flex items-center gap-4",
					themeToggle,
					components.ThemeSelector(),
				),
			),
		),
		components.Section("Animations",
			components.Div("space-y-4",
				components.Div("flex flex-wrap gap-2",
					components.Button(components.ButtonProps{Text: "Bounce", Size: components.ButtonSM, OnClick: func() {
						components.Bounce(animBox)
					}}),
					components.Button(components.ButtonProps{Text: "Shake", Size: components.ButtonSM, OnClick: func() {
						components.Shake(animBox)
					}}),
					components.Button(components.ButtonProps{Text: "Pulse", Size: components.ButtonSM, OnClick: func() {
						components.Pulse(animBox, 3)
					}}),
					components.Button(components.ButtonProps{Text: "Spin", Size: components.ButtonSM, OnClick: func() {
						components.Spin(animBox)
					}}),
					components.Button(components.ButtonProps{Text: "Wiggle", Size: components.ButtonSM, OnClick: func() {
						components.Wiggle(animBox)
					}}),
					components.Button(components.ButtonProps{Text: "Flash", Size: components.ButtonSM, OnClick: func() {
						components.Flash(animBox, 3)
					}}),
				),
				components.Div("p-4 bg-gray-100 rounded-lg inline-block", animBox),
			),
		),
		components.Section("Form Builder",
			components.Div("max-w-md",
				components.Text("Dynamic form generated from configuration:"),
				components.Div("mt-4", formBuilder.Element()),
			),
		),
		components.Section("WebSocket (Raw)",
			components.Div("space-y-3",
				components.Text("Raw WebSocket with echo.websocket.org:"),
				components.Div("flex items-center gap-4",
					components.Text("Status:"),
					wsStatusText,
				),
				components.Div("flex gap-2",
					components.Button(components.ButtonProps{Text: "Connect", Variant: components.ButtonSuccess, Size: components.ButtonSM, OnClick: func() {
						connectEchoWS()
					}}),
					components.Button(components.ButtonProps{Text: "Disconnect", Variant: components.ButtonDanger, Size: components.ButtonSM, OnClick: func() {
						disconnectEchoWS()
					}}),
				),
				wsMessageLog,
				components.Div("flex gap-2",
					wsInput,
					components.Button(components.ButtonProps{Text: "Send", Size: components.ButtonSM, OnClick: func() {
						sendEchoMessage()
					}}),
				),
			),
		),
		components.Section("Posts Subscribe (First-Party API)",
			components.Div("space-y-3",
				components.Text("Type-safe WebSocket API - mirrors HTTP client pattern:"),
				components.Div("flex items-center gap-4",
					components.Text("Status:"),
					subStatusText,
				),
				components.Div("flex gap-2",
					components.Button(components.ButtonProps{Text: "Subscribe", Variant: components.ButtonSuccess, Size: components.ButtonSM, OnClick: func() {
						startSubscription()
					}}),
					components.Button(components.ButtonProps{Text: "Unsubscribe", Variant: components.ButtonDanger, Size: components.ButtonSM, OnClick: func() {
						stopSubscription()
					}}),
				),
				subLog,
				components.Text("Create posts via HTTP API to see real-time events here."),
			),
		),
		components.Section("Component Inspector",
			components.Div("space-y-2",
				components.Text("Debug tool for viewing component hierarchy:"),
				components.PrimaryButton("Open Inspector", func() {
					components.InitInspector()
					inspector := components.GetInspector()
					if inspector != nil {
						inspector.Open()
					}
				}),
			),
		),
	)
}

func chartsDemo() js.Value {
	// Sample data
	barData := []components.ChartData{
		{Label: "Jan", Value: 65},
		{Label: "Feb", Value: 59},
		{Label: "Mar", Value: 80},
		{Label: "Apr", Value: 81},
		{Label: "May", Value: 56},
		{Label: "Jun", Value: 55},
	}

	lineData := []components.ChartData{
		{Label: "Mon", Value: 12},
		{Label: "Tue", Value: 19},
		{Label: "Wed", Value: 15},
		{Label: "Thu", Value: 25},
		{Label: "Fri", Value: 22},
		{Label: "Sat", Value: 30},
		{Label: "Sun", Value: 28},
	}

	pieData := []components.ChartData{
		{Label: "Desktop", Value: 55},
		{Label: "Mobile", Value: 35},
		{Label: "Tablet", Value: 10},
	}

	sparklineData := []float64{5, 10, 5, 20, 8, 15, 12, 18, 14, 22, 16, 25}

	return components.Div("space-y-6",
		components.Section("Bar Chart",
			components.BarChart(components.BarChartProps{
				Data:       barData,
				Height:     "200px",
				ShowLabels: true,
				ShowValues: true,
				Horizontal: true,
			}),
		),
		components.Section("Line Chart",
			components.LineChart(components.LineChartProps{
				Data:       lineData,
				Height:     "200px",
				ShowLabels: true,
				ShowPoints: true,
				ShowGrid:   true,
				FillColor:  "#3b82f6",
			}),
		),
		components.Section("Pie Chart",
			components.PieChart(components.PieChartProps{
				Data:       pieData,
				ShowLegend: true,
			}),
		),
		components.Section("Donut Chart",
			components.DonutChart(pieData, 40),
		),
		components.Section("Sparklines",
			components.Div("space-y-2",
				components.Div("flex items-center gap-4",
					components.Text("Revenue:"),
					components.LineSparkline(sparklineData),
				),
				components.Div("flex items-center gap-4",
					components.Text("Users:"),
					components.BarSparkline(sparklineData),
				),
				components.Div("flex items-center gap-4",
					components.Text("Trend:"),
					components.TrendSparkline(sparklineData),
				),
			),
		),
	)
}

func showCreatePost() {
	form := components.NewForm(components.FormProps{
		Fields: []components.FormField{
			{Name: "title", Label: "Title", Placeholder: "Enter post title", Rules: []components.ValidationRule{components.Required, components.MinLength(3)}},
			{Name: "body", Label: "Body", Placeholder: "Write your post content...", Rules: []components.ValidationRule{components.Required, components.MinLength(10)}},
		},
		SubmitLabel: "Create Post",
		CancelLabel: "Cancel",
		OnSubmit: func(values map[string]string) {
			go func() {
				_, err := posts.Create(api.CreatePostRequest{
					UserID: 1, Title: values["title"], Body: values["body"],
				})
				if err != nil {
					components.Toast("Failed to create post: "+err.Error(), components.ToastError)
					return
				}
				components.Toast("Post created successfully!", components.ToastSuccess)
				router.Navigate("/api-test")
			}()
		},
		OnCancel: func() { router.Navigate("/") },
	})

	layout.SetPage("Create New Post", "Fill out the form below to create a new post.",
		components.Div("mt-4 max-w-lg", form.Element()),
	)
}

func showSettings() {
	nameInput := components.NewInput(components.InputProps{Label: "Display Name", Value: "Admin User"})
	themeSelect := components.SimpleSelect("Theme", "Light", "Dark", "System")
	notifyCheckbox := components.NewCheckbox(components.CheckboxProps{Label: "Enable notifications", Checked: true})

	layout.SetPage("Settings", "",
		components.Div("space-y-4 max-w-md",
			nameInput.Element(),
			themeSelect.Element(),
			notifyCheckbox.Element(),
			components.Div("pt-4",
				components.PrimaryButton("Save Settings", func() {
					components.Toast("Settings saved successfully!", components.ToastSuccess)
				}),
			),
		),
	)
}

func fetchSinglePost() {
	display.ShowLoading("Fetching post #1...")

	post, err := posts.GetByID(1)
	if err != nil {
		display.ShowError("Error: " + err.Error())
		components.Toast("Failed to fetch post", components.ToastError)
		return
	}

	display.ShowJSON(post)
	components.Toast("Post loaded successfully", components.ToastSuccess)
}

func fetchAllPosts() {
	display.ShowLoading("Fetching all posts...")

	allPosts, err := posts.GetAll()
	if err != nil {
		display.ShowError("Error: " + err.Error())
		components.Toast("Failed to fetch posts", components.ToastError)
		return
	}

	// Show first 5 posts
	if len(allPosts) > 5 {
		allPosts = allPosts[:5]
	}
	display.ShowJSON(allPosts)
	components.Toast("Posts loaded successfully", components.ToastSuccess)
}

// getCommandPaletteCommands returns the commands for the command palette
func getCommandPaletteCommands() []components.Command {
	return []components.Command{
		// Navigation commands
		{
			ID:          "nav-dashboard",
			Label:       "Dashboard",
			Description: "Go to the main dashboard",
			Icon:        "📊",
			Category:    "Navigation",
			OnExecute:   func() { router.Navigate("/") },
		},
		{
			ID:          "nav-api-test",
			Label:       "API Test",
			Description: "Test API endpoints",
			Icon:        "🔌",
			Category:    "Navigation",
			OnExecute:   func() { router.Navigate("/api-test") },
		},
		{
			ID:          "nav-create-post",
			Label:       "Create Post",
			Description: "Create a new blog post",
			Icon:        "✏️",
			Category:    "Navigation",
			OnExecute:   func() { router.Navigate("/create-post") },
		},
		{
			ID:          "nav-components",
			Label:       "Components",
			Description: "View component showcase",
			Icon:        "🧩",
			Category:    "Navigation",
			OnExecute:   func() { router.Navigate("/components") },
		},
		{
			ID:          "nav-settings",
			Label:       "Settings",
			Description: "Manage application settings",
			Icon:        "⚙️",
			Category:    "Navigation",
			OnExecute:   func() { router.Navigate("/settings") },
		},
		// Action commands
		{
			ID:          "action-toggle-sidebar",
			Label:       "Toggle Sidebar",
			Description: "Collapse or expand the sidebar",
			Icon:        "📐",
			Category:    "Actions",
			Shortcut:    "Ctrl+B",
			OnExecute:   func() { layout.Sidebar().ToggleCollapse() },
		},
		{
			ID:          "action-toggle-theme",
			Label:       "Toggle Dark Mode",
			Description: "Switch between light and dark theme",
			Icon:        "🌙",
			Category:    "Actions",
			OnExecute: func() {
				components.ToggleTheme()
				components.Toast("Theme toggled", components.ToastInfo)
			},
		},
		{
			ID:          "action-refresh",
			Label:       "Refresh Page",
			Description: "Reload the current page",
			Icon:        "🔄",
			Category:    "Actions",
			OnExecute: func() {
				js.Global().Get("location").Call("reload")
			},
		},
		{
			ID:          "action-clear-notifications",
			Label:       "Clear Notifications",
			Description: "Mark all notifications as read",
			Icon:        "🔔",
			Category:    "Actions",
			OnExecute: func() {
				components.Toast("Notifications cleared", components.ToastSuccess)
			},
		},
	}
}
