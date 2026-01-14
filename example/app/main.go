//go:build js && wasm

package main

import (
	"context"
	"syscall/js"
	"time"

	"goquery/components"
	"goquery/example/api"
	"goquery/state"
)

var (
	app        *components.App
	layout     *components.Layout
	display    *components.DataDisplay
	router     *components.Router
	posts      *api.PostsClient
	modal      *components.Modal
	postsStore *state.AsyncStore[[]api.Post]
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
			{Title: "What is GoQuery?", Content: components.Text("GoQuery is a Go WebAssembly framework for building web applications entirely in Go.")},
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
		components.Section("Clipboard", components.CopyableText("npm install goquery")),
		components.Section("Date Picker", datePicker.Element()),
		components.Section("Accordion", accordion.Element()),
		components.Section("Stepper", stepper.Element()),
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
				_, err := posts.Create(context.Background(), api.CreatePostRequest{
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

	post, err := posts.GetByID(context.Background(), 1)
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

	allPosts, err := posts.GetAll(context.Background())
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
