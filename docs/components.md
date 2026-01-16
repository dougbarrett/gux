# Components

Gux includes 45+ production-ready UI components built for Go WebAssembly. All components use Tailwind CSS for styling and follow a props-based configuration pattern.

## Getting Started

```go
import "gux/components"

func main() {
    // Load Tailwind CSS
    components.LoadTailwind()

    // Initialize toast notifications
    components.InitToasts()

    // Create your app
    app := components.NewApp("#app")
    app.Mount(/* your components */)
}
```

## Form Components

### Button

```go
// Basic button
btn := components.Button(components.ButtonProps{
    Text:    "Click Me",
    OnClick: func() { fmt.Println("clicked") },
})

// With variants and sizes
btn := components.Button(components.ButtonProps{
    Text:      "Save",
    Variant:   components.ButtonSuccess,
    Size:      components.ButtonLG,
    ClassName: "w-full",
    OnClick:   handleSave,
})

// Convenience constructors
components.PrimaryButton("Submit", handleSubmit)
components.SecondaryButton("Cancel", handleCancel)
components.DangerButton("Delete", handleDelete)
```

**Variants:** `ButtonPrimary`, `ButtonSecondary`, `ButtonSuccess`, `ButtonWarning`, `ButtonDanger`, `ButtonInfo`, `ButtonGhost`

**Sizes:** `ButtonSM`, `ButtonMD`, `ButtonLG`

### Input

```go
input := components.Input(components.InputProps{
    Label:       "Email",
    Type:        components.InputEmail,
    Placeholder: "you@example.com",
    Value:       "initial@value.com",
    OnChange:    func(value string) { /* handle change */ },
})

// Access value
email := input.Value()

// Set value programmatically
input.SetValue("new@email.com")

// Focus and clear
input.Focus()
input.Clear()
```

**Types:** `InputText`, `InputEmail`, `InputPassword`, `InputNumber`, `InputURL`

### TextArea

```go
textarea := components.TextArea(components.TextAreaProps{
    Label:       "Description",
    Placeholder: "Enter description...",
    Rows:        5,
    Value:       "",
    OnChange:    func(value string) { /* handle */ },
})
```

### Select

```go
// Simple select
sel := components.SimpleSelect(
    "Country",
    []components.SelectOption{
        {Label: "United States", Value: "us"},
        {Label: "Canada", Value: "ca"},
        {Label: "Mexico", Value: "mx"},
    },
    func(value string) { /* handle */ },
)

// With placeholder
sel := components.SimpleSelectWithPlaceholder(
    "Country",
    "Select a country...",
    options,
    handleChange,
)

// Full props
sel := components.Select(components.SelectProps{
    Label:       "Country",
    Options:     options,
    Value:       "us",
    Placeholder: "Select...",
    OnChange:    handleChange,
})
```

### Checkbox

```go
cb := components.Checkbox(components.CheckboxProps{
    Label:    "Accept terms",
    Checked:  false,
    OnChange: func(checked bool) { /* handle */ },
})

// Access state
isChecked := cb.IsChecked()
cb.SetChecked(true)
```

### Toggle

```go
toggle := components.Toggle(components.ToggleProps{
    Label:    "Enable notifications",
    Checked:  true,
    OnChange: func(enabled bool) { /* handle */ },
})

// Simple toggle
components.SimpleToggle("Dark Mode", isDark, func(on bool) {
    // handle toggle
})
```

### DatePicker

```go
picker := components.DatePicker(components.DatePickerProps{
    Label:       "Start Date",
    Placeholder: "Select date...",
    OnChange:    func(date time.Time) { /* handle */ },
})
```

### Combobox

Searchable dropdown with descriptions:

```go
combo := components.Combobox(components.ComboboxProps{
    Label:       "Assign to",
    Placeholder: "Search users...",
    Options: []components.ComboboxOption{
        {Label: "John Doe", Value: "1", Description: "Engineering"},
        {Label: "Jane Smith", Value: "2", Description: "Design"},
    },
    OnChange: func(value string) { /* handle */ },
})
```

### FileUpload

```go
upload := components.FileUpload(components.FileUploadProps{
    Label:    "Upload Image",
    Multiple: false,
    OnChange: func(files []components.FileInfo) {
        for _, f := range files {
            fmt.Println(f.Name, f.Size, f.Type)
        }
    },
})

// Image upload with preview
components.ImageUpload("Profile Photo", func(files []components.FileInfo) {
    // handle upload
})
```

### FormBuilder

Dynamic form generation from configuration:

```go
form := components.NewFormBuilder(components.FormBuilderProps{
    Fields: []components.BuilderField{
        {
            Name:        "email",
            Type:        components.BuilderFieldEmail,
            Label:       "Email",
            Placeholder: "you@example.com",
            Rules:       []components.ValidationRule{components.Required, components.Email},
        },
        {
            Name:        "password",
            Type:        components.BuilderFieldPassword,
            Label:       "Password",
            Placeholder: "Enter password",
            Rules:       []components.ValidationRule{components.Required, components.MinLength(8)},
        },
        {
            Name:    "role",
            Type:    components.BuilderFieldSelect,
            Label:   "Role",
            Options: []components.SelectOption{
                {Label: "Admin", Value: "admin"},
                {Label: "User", Value: "user"},
            },
        },
    },
    SubmitText: "Create Account",
    ShowCancel: true,
    CancelText: "Back",
    OnSubmit: func(values map[string]string) {
        // values["email"], values["password"], values["role"]
    },
    OnCancel: func() { /* handle cancel */ },
})
```

**Field Types:** `BuilderFieldText`, `BuilderFieldEmail`, `BuilderFieldPassword`, `BuilderFieldNumber`, `BuilderFieldSelect`, `BuilderFieldTextarea`, `BuilderFieldCheckbox`

**Validation Rules:** `Required`, `Email`, `MinLength(n)`, `MaxLength(n)`, `Pattern(regex)`

## Layout Components

### Layout

Main application layout with sidebar and header:

```go
layout := components.Layout(components.LayoutProps{
    Sidebar: components.SidebarProps{
        Title: "My App",
        Items: []components.NavItem{
            {Label: "Dashboard", Path: "/", Icon: "home"},
            {Label: "Posts", Path: "/posts", Icon: "file-text"},
            {Label: "Settings", Path: "/settings", Icon: "settings"},
        },
    },
    Header: components.HeaderProps{
        Title: "Dashboard",
        Actions: []js.Value{
            components.Button(components.ButtonProps{Text: "New"}),
        },
    },
})

// Update content
layout.SetContent(myContent)

// Update with header title
layout.SetPageWithHeader("Posts", postsContent)

// Access parts
sidebar := layout.Sidebar()
header := layout.Header()
```

### Sidebar

```go
sidebar := components.Sidebar(components.SidebarProps{
    Title: "Admin",
    Items: []components.NavItem{
        {Label: "Home", Path: "/", Icon: "home"},
        {Label: "Users", Path: "/users", Icon: "users"},
    },
})

// Update active item
sidebar.SetActive("/users")
```

### Header

```go
header := components.Header(components.HeaderProps{
    Title: "Dashboard",
    Actions: []js.Value{
        components.Button(components.ButtonProps{Text: "New Post"}),
        components.ThemeToggle(),
    },
})

header.SetTitle("New Title")
```

### Card

```go
// Basic card
card := components.Card(components.CardProps{
    ClassName: "max-w-md",
}, content...)

// Titled card
card := components.TitledCard("Card Title", content...)

// Section card
card := components.SectionCard(
    "Section Title",
    "Optional description",
    content...,
)
```

### Tabs

```go
tabs := components.Tabs(components.TabsProps{
    Tabs: []components.Tab{
        {Label: "Profile", Content: profileContent},
        {Label: "Settings", Content: settingsContent},
        {Label: "Security", Content: securityContent},
    },
})
```

### Accordion

```go
accordion := components.Accordion(components.AccordionProps{
    AllowMultiple: true,
    Items: []components.AccordionItem{
        {Title: "Section 1", Content: content1},
        {Title: "Section 2", Content: content2},
        {Title: "Section 3", Content: content3},
    },
})
```

### Drawer

Side panel that slides in:

```go
drawer := components.RightDrawer(components.DrawerProps{
    Title:   "Details",
    Content: detailsContent,
})

// Open/close
drawer.Open()
drawer.Close()

// Other positions
components.TopDrawer(props)
components.BottomDrawer(props)
components.LeftDrawer(props)
```

## Header Components

### UserMenu

User profile dropdown with avatar trigger:

```go
menu := components.NewUserMenu(components.UserMenuProps{
    Name:       "John Doe",
    Email:      "john@example.com",
    AvatarSrc:  "/avatar.jpg", // Optional - falls back to initials
    OnProfile:  func() { router.Navigate("/profile") },
    OnSettings: func() { router.Navigate("/settings") },
    OnLogout:   func() { handleLogout() },
})

// Add to header
header.Call("appendChild", menu.Element())

// Programmatic control
menu.Open()
menu.Close()
menu.Destroy() // Clean up event listeners
```

**Props:**
- `Name` - User's display name
- `Email` - User's email address
- `AvatarSrc` - Optional avatar image URL (falls back to initials)
- `OnProfile` - Callback when Profile is clicked
- `OnSettings` - Callback when Settings is clicked
- `OnLogout` - Callback when Logout is clicked

**Methods:**
- `Element()` - Returns the DOM element
- `Open()` - Opens the dropdown menu
- `Close()` - Closes the dropdown menu
- `Destroy()` - Cleans up event listeners

**Note:** Uses Avatar component internally and extends Dropdown for the menu.

### NotificationCenter

Notification bell with dropdown list and unread badge:

```go
nc := components.NewNotificationCenter(components.NotificationCenterProps{
    Notifications: []components.Notification{
        {
            ID:      "1",
            Title:   "New message",
            Message: "You have a new message from Alice",
            Time:    "2 min ago",
            Read:    false,
            Type:    "info", // "info", "success", "warning", "error"
        },
        {
            ID:      "2",
            Title:   "Task completed",
            Message: "Build process finished successfully",
            Time:    "1 hour ago",
            Read:    true,
            Type:    "success",
        },
    },
    OnMarkRead:          func(id string) { markNotificationRead(id) },
    OnMarkAllRead:       func() { markAllRead() },
    OnClear:             func() { clearNotifications() },
    OnNotificationClick: func(id string) { openNotification(id) },
})

// Add to header
header.Call("appendChild", nc.Element())

// Update notifications dynamically
nc.SetNotifications(newNotifications)

// Get unread count
count := nc.UnreadCount()

// Programmatic control
nc.Open()
nc.Close()
nc.Destroy()
```

**Notification struct:**
- `ID` - Unique identifier
- `Title` - Notification title
- `Message` - Notification body text
- `Time` - Time string (e.g., "2 min ago")
- `Read` - Whether the notification has been read
- `Type` - Type indicator: `"info"`, `"success"`, `"warning"`, `"error"`

**Props:**
- `Notifications` - Initial list of notifications
- `OnMarkRead` - Callback when a notification is marked as read
- `OnMarkAllRead` - Callback when "Mark all read" is clicked
- `OnClear` - Callback when "Clear all" is clicked
- `OnNotificationClick` - Callback when a notification is clicked

**Methods:**
- `Element()` - Returns the DOM element
- `SetNotifications([]Notification)` - Updates the notification list
- `UnreadCount()` - Returns number of unread notifications
- `Open()` - Opens the dropdown
- `Close()` - Closes the dropdown
- `Destroy()` - Cleans up event listeners

**Note:** Shows unread badge count on the bell icon. Notification list is scrollable.

## Data Display Components

### Table

```go
table := components.Table(components.TableProps{
    Columns: []components.TableColumn{
        {Header: "ID", Key: "id", Width: "w-16"},
        {Header: "Name", Key: "name"},
        {Header: "Status", Key: "status", Render: func(row map[string]any) js.Value {
            status := row["status"].(string)
            variant := components.BadgeInfo
            if status == "active" {
                variant = components.BadgeSuccess
            }
            return components.Badge(components.BadgeProps{
                Text:    status,
                Variant: variant,
            })
        }},
        {Header: "Actions", Key: "actions", Render: func(row map[string]any) js.Value {
            return components.Button(components.ButtonProps{
                Text:    "Edit",
                Size:    components.ButtonSM,
                OnClick: func() { editRow(row["id"].(int)) },
            })
        }},
    },
    Data: []map[string]any{
        {"id": 1, "name": "John", "status": "active"},
        {"id": 2, "name": "Jane", "status": "pending"},
    },
    Striped:    true,
    Hoverable:  true,
    Bordered:   false,
    Compact:    false,
    OnRowClick: func(row map[string]any) { /* handle click */ },
})

// Update data
table.UpdateData(newData)
```

### Badge

```go
badge := components.Badge(components.BadgeProps{
    Text:    "Active",
    Variant: components.BadgeSuccess,
    Rounded: true,
})

// Convenience functions
components.BadgePrimaryText("Primary")
components.BadgeSuccessText("Success")
components.BadgeWarningText("Warning")
components.BadgeErrorText("Error")
```

**Variants:** `BadgePrimary`, `BadgeSecondary`, `BadgeSuccess`, `BadgeWarning`, `BadgeError`, `BadgeInfo`

### Avatar

```go
avatar := components.Avatar(components.AvatarProps{
    Name:   "John Doe",
    Size:   components.AvatarLG,
    Status: "online", // online, away, offline, busy
})

// Avatar group
group := components.AvatarGroup([]components.AvatarProps{
    {Name: "Alice"},
    {Name: "Bob"},
    {Name: "Charlie"},
})
```

**Sizes:** `AvatarSM`, `AvatarMD`, `AvatarLG`

### Breadcrumbs

```go
breadcrumbs := components.Breadcrumbs(components.BreadcrumbsProps{
    Items: []components.BreadcrumbItem{
        {Label: "Home", Path: "/"},
        {Label: "Users", Path: "/users"},
        {Label: "John Doe", Path: ""}, // Current page (no link)
    },
})
```

### Pagination

```go
pagination := components.Pagination(components.PaginationProps{
    CurrentPage:  1,
    TotalPages:   10,
    TotalItems:   100,
    ItemsPerPage: 10,
    ShowInfo:     true,
    OnPageChange: func(page int) {
        // Load page data
    },
})
```

### VirtualList

Efficient rendering for large lists:

```go
list := components.VirtualList(components.VirtualListProps{
    Items:      largeDataset, // []any
    ItemHeight: 48,
    OnRender: func(item any, index int) js.Value {
        data := item.(MyType)
        return components.Div("p-2", components.Text(data.Name))
    },
})
```

## Data Export

### ExportCSV

Export data to CSV file with browser download:

```go
data := []map[string]any{
    {"id": 1, "name": "John", "email": "john@example.com"},
    {"id": 2, "name": "Jane", "email": "jane@example.com"},
}

columns := []string{"id", "name", "email"}

components.ExportCSV(data, columns, "users.csv")
```

**Parameters:**
- `data` - Slice of maps containing row data
- `columns` - Column order and which fields to include
- `filename` - Output filename (.csv extension added automatically)

**Note:** Handles proper CSV escaping for quotes, commas, and newlines.

### ExportJSON

Export data to JSON file with browser download:

```go
data := []map[string]any{
    {"id": 1, "name": "John"},
    {"id": 2, "name": "Jane"},
}

components.ExportJSON(data, "users.json")
```

**Parameters:**
- `data` - Slice of maps to export
- `filename` - Output filename (.json extension added automatically)

**Note:** JSON is formatted with indentation for readability.

### ExportPDF

Export data to PDF table with browser download:

```go
data := []map[string]any{
    {"id": 1, "name": "John", "status": "active"},
    {"id": 2, "name": "Jane", "status": "pending"},
}

headers := []string{"ID", "Name", "Status"}
keys := []string{"id", "name", "status"}

components.ExportPDF(data, headers, keys, "report.pdf", components.PDFExportOptions{
    Title:       "User Report",
    Orientation: "landscape", // "portrait" or "landscape"
    PageSize:    "a4",        // "a4" or "letter"
})
```

**Parameters:**
- `data` - Slice of maps containing row data
- `headers` - Column display names
- `keys` - Data field keys (order matches headers)
- `filename` - Output filename (.pdf extension added automatically)
- `options` - PDFExportOptions struct

**PDFExportOptions:**
- `Title` - Title shown at top of PDF
- `Orientation` - `"portrait"` or `"landscape"` (default: portrait)
- `PageSize` - `"a4"` or `"letter"` (default: a4)

**Note:** Requires jsPDF and jsPDF-AutoTable libraries. Table component has built-in export dropdown when `Exportable: true`.

## Feedback Components

### Modal

```go
modal := components.Modal(components.ModalProps{
    Title:      "Confirm Action",
    CloseOnEsc: true,
    Width:      "max-w-lg",
    Content:    components.Text("Are you sure you want to proceed?"),
    Footer: components.Div("flex gap-2 justify-end",
        components.Button(components.ButtonProps{
            Text:    "Cancel",
            Variant: components.ButtonSecondary,
            OnClick: func() { modal.Close() },
        }),
        components.Button(components.ButtonProps{
            Text:    "Confirm",
            Variant: components.ButtonPrimary,
            OnClick: func() {
                // Handle confirm
                modal.Close()
            },
        }),
    ),
})

// Show modal
modal.Open()

// Close modal
modal.Close()
```

### Toast

```go
// Initialize once at app start
components.InitToasts()

// Show toasts
components.Toast("Operation successful!", components.ToastSuccess)
components.Toast("Something went wrong", components.ToastError)
components.Toast("Please note...", components.ToastInfo)
components.Toast("Be careful!", components.ToastWarning)
```

**Variants:** `ToastSuccess`, `ToastError`, `ToastInfo`, `ToastWarning`

### Alert

```go
alert := components.Alert(components.AlertProps{
    Variant: components.AlertWarning,
    Message: "Your session will expire soon.",
})

// Convenience functions
components.AlertInfoMsg("Information message")
components.AlertSuccessMsg("Success message")
components.AlertWarningMsg("Warning message")
components.AlertErrorMsg("Error message")
```

**Variants:** `AlertInfo`, `AlertSuccess`, `AlertWarning`, `AlertError`

### Progress

```go
progress := components.NewProgress(components.ProgressProps{
    Value:    75,
    Variant:  components.BadgeSuccess,
    Striped:  true,
    Animated: true,
    Label:    "75% Complete",
})
```

### Spinner

```go
spinner := components.Spinner(components.SpinnerProps{
    Size:  components.SpinnerLG,
    Label: "Loading...",
    Color: "text-blue-500",
})
```

**Sizes:** `SpinnerSM`, `SpinnerMD`, `SpinnerLG`

### Skeleton

Loading placeholders:

```go
// Text skeleton
components.SkeletonText(3) // 3 lines

// Avatar skeleton
components.SkeletonAvatar()

// Card skeleton
components.SkeletonCard()
```

### Tooltip

```go
// Wrap element with tooltip
buttonWithTooltip := components.WithTooltip(
    components.Button(components.ButtonProps{Text: "Hover me"}),
    "This is a tooltip",
    components.TooltipTop, // Top, Bottom, Left, Right
)
```

### ConnectionStatus

Display WebSocket connection state with multiple variants:

```go
// Basic dot indicator
status := components.NewConnectionStatus(components.ConnectionStatusProps{
    Variant:   components.ConnectionStatusDotVariant, // "dot", "badge", "text", "full"
    Size:      components.ConnectionStatusMD,         // "sm", "md", "lg"
    ShowLabel: true,                                  // Show text label next to dot
})

// Bind to WebSocket store for automatic updates
status.BindToWebSocket(wsStore)

// Manual state control
status.SetState(state.WSOpen)      // Connected
status.SetState(state.WSConnecting) // Connecting
status.SetState(state.WSClosed)     // Disconnected

// Get current state
currentState := status.State()

// Cleanup
status.Unbind()
```

**Convenience constructors:**

```go
components.ConnectionStatusDot(components.ConnectionStatusMD)  // Minimal dot
components.ConnectionStatusBadge()                             // Badge with text
components.ConnectionStatusText()                              // Text only
components.ConnectionStatusFull()                              // Icon + message
```

**Variants:**
- `ConnectionStatusDotVariant` - Colored dot indicator
- `ConnectionStatusBadgeVariant` - Pill badge with status text
- `ConnectionStatusTextVariant` - Text-only status
- `ConnectionStatusFullVariant` - Full indicator with dot and label

**Sizes:** `ConnectionStatusSM`, `ConnectionStatusMD`, `ConnectionStatusLG`

**Props:**
- `Variant` - Display variant (default: dot)
- `Size` - Component size (default: md)
- `Labels` - Custom labels for each state
- `ShowLabel` - Show text label next to dot

**Methods:**
- `Element()` - Returns the DOM element
- `SetState(state)` - Manually set connection state
- `State()` - Get current state
- `BindToWebSocket(store)` - Subscribe to WebSocketStore changes
- `Unbind()` - Remove WebSocket subscription

**Note:** Uses ARIA live region for accessibility announcements when state changes.

### EmptyState

Friendly empty state messages with optional action:

```go
empty := components.NewEmptyState(components.EmptyStateProps{
    Icon:        "📭",              // Emoji icon
    Title:       "No messages",
    Description: "You don't have any messages yet.",
    ActionLabel: "Compose",          // Optional button
    OnAction:    func() { compose() },
    Compact:     false,              // Smaller variant
})

container.Call("appendChild", empty.Element())
```

**Convenience constructors:**

```go
// Generic "no data" state
components.NoData("No items")

// "No results" with clear filter button
components.NoResults(func() { clearFilters() })

// "Nothing selected" state
components.NoSelection()
```

**Props:**
- `Icon` - Emoji or text icon (default: "📭")
- `Title` - Main heading (required)
- `Description` - Explanatory text (optional)
- `ActionLabel` - Button text (optional)
- `OnAction` - Button click handler (optional)
- `Compact` - Smaller variant for inline use

**Methods:**
- `Element()` - Returns the container DOM element

**Note:** Centered layout with large emoji icon and optional primary action button.

### ConfirmDialog

Confirmation dialog for destructive or important actions:

```go
dialog := components.NewConfirmDialog(components.ConfirmDialogProps{
    Title:       "Delete Item",
    Message:     "Are you sure you want to delete this item? This action cannot be undone.",
    ConfirmText: "Delete",           // Default: "Confirm"
    CancelText:  "Cancel",           // Default: "Cancel"
    Variant:     components.ConfirmVariantDanger, // Button style
    OnConfirm:   func() { deleteItem() },
    OnCancel:    func() { /* optional */ },
})

document.Get("body").Call("appendChild", dialog.Element())

// Show dialog
dialog.Open()

// Close programmatically
dialog.Close()

// Check state
if dialog.IsOpen() { /* ... */ }
```

**Convenience constructors:**

```go
// Standard confirmation
dialog := components.Confirm("Confirm Action", "Are you sure?", func() {
    // Handle confirm
})
dialog.Open()

// Dangerous action (red button)
dialog := components.ConfirmDanger("Delete Account", "This cannot be undone.", func() {
    // Handle delete
})
dialog.Open()
```

**Variants:**
- `ConfirmVariantDefault` - Primary button style
- `ConfirmVariantDanger` - Red/danger button style
- `ConfirmVariantWarning` - Warning button style

**Props:**
- `Title` - Dialog title
- `Message` - Confirmation message
- `ConfirmText` - Confirm button text (default: "Confirm")
- `CancelText` - Cancel button text (default: "Cancel")
- `Variant` - Visual style (affects confirm button)
- `OnConfirm` - Called when confirmed
- `OnCancel` - Called when cancelled (optional)

**Methods:**
- `Element()` - Returns the dialog DOM element
- `Open()` - Shows the confirmation dialog
- `Close()` - Hides the dialog
- `IsOpen()` - Returns whether dialog is open

**Note:** Uses `alertdialog` ARIA role for accessibility. Wraps Modal component with focus management.

## Navigation Components

### Router

```go
router := components.NewRouter()
components.SetGlobalRouter(router)

// Register routes
router.Register("/", showHome)
router.Register("/posts", showPosts)
router.Register("/posts/:id", showPost)
router.Register("/settings", showSettings)

// Start listening to URL changes
router.Start()

// Navigate programmatically
router.Navigate("/posts")

// Get current path
currentPath := router.CurrentPath()
```

### Link

```go
link := components.Link(components.LinkProps{
    Path:      "/posts",
    Text:      "View Posts",
    ClassName: "text-blue-500 hover:underline",
})
```

### Stepper

Multi-step progress indicator:

```go
stepper := components.Stepper(components.StepperProps{
    CurrentStep: 1, // 0-indexed
    Steps: []components.Step{
        {Title: "Account", Description: "Create your account"},
        {Title: "Profile", Description: "Set up your profile"},
        {Title: "Confirm", Description: "Review and confirm"},
    },
})
```

### CommandPalette

Searchable command palette with Cmd/Ctrl+K keyboard shortcut:

```go
palette := components.NewCommandPalette(components.CommandPaletteProps{
    Commands: []components.Command{
        {
            ID:          "new-post",
            Label:       "Create New Post",
            Description: "Start writing a new blog post",
            Icon:        "📝",
            Category:    "Actions",
            Shortcut:    "Ctrl+N",
            OnExecute:   func() { createNewPost() },
        },
        {
            ID:          "go-dashboard",
            Label:       "Go to Dashboard",
            Description: "Navigate to main dashboard",
            Icon:        "🏠",
            Category:    "Navigation",
            OnExecute:   func() { router.Navigate("/") },
        },
        {
            ID:          "toggle-theme",
            Label:       "Toggle Dark Mode",
            Icon:        "🌙",
            Category:    "Settings",
            OnExecute:   func() { toggleTheme() },
        },
    },
    Placeholder:  "Search commands...",
    EmptyMessage: "No commands found",
    OnClose:      func() { /* optional close callback */ },
})

// Mount the overlay element
document.Get("body").Call("appendChild", palette.Element())

// Register global keyboard shortcut (Cmd/Ctrl+K)
palette.RegisterKeyboardShortcut()

// Programmatic control
palette.Open()
palette.Close()
palette.IsOpen()

// Update commands dynamically
palette.SetCommands(newCommands)

// Cleanup
palette.UnregisterKeyboardShortcut()
palette.Destroy()
```

**Command struct:**
- `ID` - Unique identifier
- `Label` - Display text
- `Description` - Optional subtitle
- `Icon` - Optional emoji or icon
- `Category` - For grouping (e.g., "Navigation", "Actions")
- `OnExecute` - Function called when command is executed
- `Shortcut` - Optional keyboard hint (e.g., "Ctrl+B")

**Props:**
- `Commands` - List of available commands
- `Placeholder` - Search input placeholder (default: "Search commands...")
- `EmptyMessage` - Message when no results (default: "No commands found")
- `OnClose` - Optional callback when palette closes

**Methods:**
- `Element()` - Returns the overlay DOM element
- `Open()` - Shows the command palette
- `Close()` - Hides the command palette
- `IsOpen()` - Returns whether palette is currently open
- `SetCommands([]Command)` - Updates available commands
- `RegisterKeyboardShortcut()` - Registers global Cmd/Ctrl+K listener
- `UnregisterKeyboardShortcut()` - Removes global keyboard listener
- `Destroy()` - Cleans up all resources

**Keyboard shortcuts:**
- `Cmd/Ctrl+K` - Toggle command palette
- `↑↓` - Navigate through commands
- `Enter` - Execute selected command
- `Esc` - Close palette

**Note:** Commands are automatically grouped by category. Uses FocusTrap for modal focus management.

## Chart Components

### BarChart

```go
chart := components.BarChart(components.ChartProps{
    Data: []components.ChartData{
        {Label: "Jan", Value: 100},
        {Label: "Feb", Value: 150},
        {Label: "Mar", Value: 120},
        {Label: "Apr", Value: 180},
    },
    Height:     200,
    ShowLabels: true,
    ShowValues: true,
})
```

### LineChart

```go
chart := components.LineChart(components.ChartProps{
    Data: []components.ChartData{
        {Label: "Mon", Value: 10},
        {Label: "Tue", Value: 25},
        {Label: "Wed", Value: 15},
        {Label: "Thu", Value: 30},
    },
    Height:     200,
    ShowLabels: true,
})
```

### PieChart / DonutChart

```go
pie := components.PieChart(components.ChartProps{
    Data: []components.ChartData{
        {Label: "Chrome", Value: 65},
        {Label: "Firefox", Value: 20},
        {Label: "Safari", Value: 15},
    },
    ShowLabels: true,
})

donut := components.DonutChart(components.ChartProps{
    Data:       data,
    ShowLabels: true,
})
```

### Sparkline

Inline mini charts:

```go
// Line sparkline
components.LineSparkline([]float64{10, 25, 15, 30, 20, 35})

// Bar sparkline
components.BarSparkline([]float64{10, 25, 15, 30, 20, 35})

// Trend indicator
components.TrendSparkline([]float64{10, 15, 12, 18, 25})
```

## Utility Components

### Theme

```go
// Theme toggle button
toggle := components.ThemeToggle()

// Theme selector dropdown
selector := components.ThemeSelector()
```

### Animation

```go
// Get DOM element reference
element := myComponent

// Apply animations
components.Bounce(element)
components.Shake(element)
components.Pulse(element)
components.Spin(element)
components.Wiggle(element)
components.Flash(element)
```

### Clipboard

```go
// Copyable text with button
copyable := components.CopyableText("npm install gux")
```

### DataDisplay

Debug component for showing formatted data:

```go
display := components.NewDataDisplay()

// Show states
display.ShowLoading()
display.ShowError(err)
display.ShowJSON(data)
```

### Inspector

Component hierarchy debugger:

```go
components.InitInspector()
inspector := components.GetInspector()
inspector.Open()
```

### Accessibility

```go
// Focus trap for modals
trap := components.FocusTrap(modalContent)

// Skip links for keyboard navigation
skipLinks := components.SkipLinks()
```

## Helper Functions

### Element Creation

```go
// Generic div with classes and children
div := components.Div("flex gap-4 p-2",
    child1,
    child2,
)

// Text elements
components.Text("Paragraph text")
components.H1("Heading 1")
components.H2("Heading 2")
components.H3("Heading 3")
// ... H4, H5, H6

// Section wrapper
section := components.Section("Section Title", content...)
```

## Dark Mode Support

All components automatically support dark mode when using the theme utilities:

```go
// Toggle dark mode
components.ThemeToggle()

// Components use dark: variants
// e.g., "bg-white dark:bg-gray-800"
```
