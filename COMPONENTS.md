# GoQuery Components API Reference

## Table of Contents

- [Core Components](#core-components)
  - [Button](#button)
  - [Input](#input)
  - [Select](#select)
  - [Textarea](#textarea)
  - [Checkbox](#checkbox)
- [Layout Components](#layout-components)
  - [Layout](#layout)
  - [Card](#card)
  - [Stack](#stack)
  - [Grid](#grid)
- [Navigation](#navigation)
  - [Router](#router)
  - [Tabs](#tabs)
  - [Pagination](#pagination)
- [Feedback Components](#feedback-components)
  - [Modal](#modal)
  - [Drawer](#drawer)
  - [Toast](#toast)
  - [Alert](#alert)
  - [Progress](#progress)
  - [Spinner](#spinner)
  - [Skeleton](#skeleton)
- [Data Display](#data-display)
  - [Table](#table)
  - [Badge](#badge)
  - [Avatar](#avatar)
  - [Tooltip](#tooltip)
- [Form Components](#form-components)
  - [Form](#form)
  - [FormBuilder](#formbuilder)
  - [Combobox](#combobox)
  - [Toggle](#toggle)
  - [FileUpload](#fileupload)
- [Charts](#charts)
  - [BarChart](#barchart)
  - [LineChart](#linechart)
  - [PieChart](#piechart)
  - [Sparkline](#sparkline)
- [Advanced Components](#advanced-components)
  - [VirtualList](#virtuallist)
  - [Dropdown](#dropdown)
  - [Inspector](#inspector)
- [Utilities](#utilities)
  - [Theme](#theme)
  - [Animation](#animation)
  - [FocusTrap](#focustrap)
  - [SkipLinks](#skiplinks)
- [State Management](#state-management)
  - [Store](#store)
  - [QueryCache](#querycache)
  - [Storage](#storage)
  - [WebSocket](#websocket)

---

## Core Components

### Button

A customizable button component.

```go
import "goquery/components"

btn := components.Button(components.ButtonProps{
    Text:      "Click Me",
    Variant:   components.ButtonPrimary, // Primary, Secondary, Success, Danger, Warning
    Size:      components.SizeMD,        // SM, MD, LG
    Disabled:  false,
    ClassName: "my-custom-class",
    OnClick: func() {
        // Handle click
    },
})
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| Text | string | "" | Button text |
| Variant | ButtonVariant | Primary | Button style variant |
| Size | Size | MD | Button size |
| Disabled | bool | false | Disable the button |
| ClassName | string | "" | Additional CSS classes |
| OnClick | func() | nil | Click handler |

### Input

A text input component with validation support.

```go
input := components.NewInput(components.InputProps{
    Type:        components.InputText,
    Label:       "Email",
    Placeholder: "Enter your email",
    Value:       "",
    Required:    true,
    OnChange: func(value string) {
        fmt.Println("Value:", value)
    },
})

// Methods
value := input.Value()
input.SetValue("new value")
input.SetError("Invalid input")
input.ClearError()
input.Focus()
```

**Input Types:** `InputText`, `InputEmail`, `InputPassword`, `InputNumber`, `InputTel`, `InputURL`, `InputSearch`, `InputDate`, `InputTime`

### Select

A dropdown select component.

```go
sel := components.NewSelect(components.SelectProps{
    Label:       "Country",
    Placeholder: "Select a country",
    Options: []components.SelectOption{
        {Label: "United States", Value: "us"},
        {Label: "Canada", Value: "ca"},
        {Label: "Mexico", Value: "mx"},
    },
    Value: "us",
    OnChange: func(value string) {
        fmt.Println("Selected:", value)
    },
})
```

---

## Layout Components

### Layout

A full-page layout with sidebar, header, and content area.

```go
layout := components.NewLayout(components.LayoutProps{
    Sidebar: components.SidebarProps{
        Title: "My App",
        Items: []components.NavItem{
            {Label: "Home", Icon: "home", Path: "/"},
            {Label: "Settings", Icon: "settings", Path: "/settings"},
        },
    },
    Header: components.HeaderProps{
        Title: "Dashboard",
    },
})

layout.SetContent(myContent)
```

### Card

A container component with optional header and footer.

```go
card := components.Card(components.CardProps{
    Title:     "Card Title",
    ClassName: "max-w-md",
    Children:  []js.Value{content},
})
```

### Stack

A vertical or horizontal stack layout.

```go
stack := components.Stack(components.StackProps{
    Direction: "vertical", // or "horizontal"
    Gap:       "4",        // Tailwind gap size
    Children:  []js.Value{child1, child2, child3},
})
```

### Grid

A CSS grid layout component.

```go
grid := components.Grid(components.GridProps{
    Cols:      3,
    Gap:       "4",
    ClassName: "p-4",
    Children:  []js.Value{item1, item2, item3},
})
```

---

## Navigation

### Router

Client-side routing for SPAs.

```go
router := components.NewRouter()
components.SetGlobalRouter(router)

router.Register("/", showHome)
router.Register("/users/:id", showUser)
router.Register("/posts/:slug", showPost)

router.Start()

// Navigate programmatically
components.Navigate("/users/123")

// Get current route params
params := components.GetRouteParams() // map[string]string{"id": "123"}
```

### Tabs

A tabbed interface component.

```go
tabs := components.NewTabs(components.TabsProps{
    Tabs: []components.Tab{
        {Label: "Overview", Content: overviewContent},
        {Label: "Details", Content: detailsContent},
        {Label: "Settings", Content: settingsContent},
    },
    DefaultTab: 0,
    OnChange: func(index int) {
        fmt.Println("Tab changed to:", index)
    },
})
```

### Pagination

A pagination component for lists.

```go
pagination := components.NewPagination(components.PaginationProps{
    CurrentPage: 1,
    TotalPages:  10,
    OnPageChange: func(page int) {
        loadPage(page)
    },
    ShowFirstLast: true,
    MaxVisible:    5,
})

// Update current page
pagination.SetPage(3)
```

---

## Feedback Components

### Modal

A modal dialog component.

```go
modal := components.NewModal(components.ModalProps{
    Title:       "Confirm Action",
    Content:     modalContent,
    Size:        "md", // sm, md, lg, xl, full
    CloseOnOverlayClick: true,
    OnClose: func() {
        // Handle close
    },
})

// Control modal
modal.Open()
modal.Close()
isOpen := modal.IsOpen()
```

### Drawer

A slide-out panel component.

```go
drawer := components.NewDrawer(components.DrawerProps{
    Title:     "Menu",
    Position:  "left", // left, right, top, bottom
    Size:      "md",
    Content:   drawerContent,
    OnClose: func() {
        // Handle close
    },
})

drawer.Open()
drawer.Close()
```

### Toast

Notification toast messages.

```go
// Show toasts
components.ShowToast(components.ToastProps{
    Message:  "Operation successful!",
    Type:     components.ToastSuccess, // Success, Error, Warning, Info
    Duration: 3000, // milliseconds
    Position: "top-right",
})

// Convenience functions
components.ToastSuccess("Saved!")
components.ToastError("Failed to save")
components.ToastWarning("Are you sure?")
components.ToastInfo("Processing...")
```

### Progress

A progress bar component.

```go
progress := components.NewProgress(components.ProgressProps{
    Value:     75,
    Max:       100,
    ShowLabel: true,
    Color:     "blue",
    Size:      "md",
})

// Update progress
progress.SetValue(90)
```

### Spinner

Loading spinner component.

```go
spinner := components.Spinner(components.SpinnerProps{
    Size:  "md",  // sm, md, lg
    Color: "blue",
})
```

### Skeleton

Loading placeholder component.

```go
skeleton := components.Skeleton(components.SkeletonProps{
    Width:   "200px",
    Height:  "20px",
    Rounded: true,
})
```

---

## Data Display

### Table

A data table component.

```go
table := components.NewTable(components.TableProps{
    Columns: []components.TableColumn{
        {Key: "id", Label: "ID", Sortable: true},
        {Key: "name", Label: "Name", Sortable: true},
        {Key: "email", Label: "Email"},
        {Key: "actions", Label: "Actions", Render: renderActions},
    },
    Data: []map[string]any{
        {"id": 1, "name": "John", "email": "john@example.com"},
        {"id": 2, "name": "Jane", "email": "jane@example.com"},
    },
    Sortable:   true,
    Selectable: true,
    OnSelect: func(selected []int) {
        fmt.Println("Selected rows:", selected)
    },
})
```

### Badge

A badge/tag component.

```go
badge := components.Badge(components.BadgeProps{
    Text:    "New",
    Variant: "success", // primary, secondary, success, danger, warning
    Size:    "sm",
})
```

### Avatar

User avatar component.

```go
avatar := components.Avatar(components.AvatarProps{
    Src:      "https://example.com/avatar.jpg",
    Alt:      "John Doe",
    Size:     "md",
    Fallback: "JD",
})
```

### Tooltip

Hover tooltip component.

```go
tooltip := components.NewTooltip(components.TooltipProps{
    Content:   "This is a tooltip",
    Position:  "top", // top, bottom, left, right
    Trigger:   triggerElement,
})
```

---

## Form Components

### Form

A validated form component.

```go
form := components.NewForm(components.FormProps{
    Fields: []components.FormField{
        {Name: "email", Label: "Email", Type: components.InputEmail, Rules: []components.ValidationRule{components.Required, components.Email}},
        {Name: "password", Label: "Password", Type: components.InputPassword, Rules: []components.ValidationRule{components.Required, components.MinLength(8)}},
    },
    SubmitLabel: "Login",
    OnSubmit: func(values map[string]string) {
        // Handle submission
    },
})
```

### FormBuilder

Dynamic form generation from configuration.

```go
fb := components.NewFormBuilder(components.FormBuilderProps{
    Fields: []components.BuilderField{
        {Name: "name", Type: components.BuilderFieldText, Label: "Name", Rules: []components.ValidationRule{components.Required}},
        {Name: "email", Type: components.BuilderFieldEmail, Label: "Email", Rules: []components.ValidationRule{components.Required, components.Email}},
        {Name: "role", Type: components.BuilderFieldSelect, Label: "Role", Options: []components.SelectOption{
            {Label: "Admin", Value: "admin"},
            {Label: "User", Value: "user"},
        }},
        {Name: "subscribe", Type: components.BuilderFieldCheckbox, Label: "Subscribe to newsletter"},
    },
    SubmitText: "Save",
    OnSubmit: func(values map[string]any) error {
        // Handle submission
        return nil
    },
})

// Methods
values := fb.GetValues()
fb.SetFormValue("name", "John")
fb.Reset()
isValid := fb.ValidateForm()
```

### Combobox

A searchable select/autocomplete component.

```go
combo := components.NewCombobox(components.ComboboxProps{
    Label:       "Select User",
    Placeholder: "Search users...",
    Options: []components.ComboboxOption{
        {Label: "John Doe", Value: "john"},
        {Label: "Jane Smith", Value: "jane"},
    },
    OnChange: func(value string) {
        fmt.Println("Selected:", value)
    },
})
```

### Toggle

A toggle switch component.

```go
toggle := components.NewToggle(components.ToggleProps{
    Label:   "Enable notifications",
    Checked: true,
    OnChange: func(checked bool) {
        fmt.Println("Toggled:", checked)
    },
})
```

### FileUpload

A drag-and-drop file upload component.

```go
upload := components.NewFileUpload(components.FileUploadProps{
    Label:         "Upload Files",
    Accept:        ".jpg,.png,.pdf",
    Multiple:      true,
    MaxSize:       5 * 1024 * 1024, // 5MB
    OnFilesSelect: func(files []components.UploadFile) {
        for _, f := range files {
            fmt.Println("File:", f.Name, f.Size)
        }
    },
})
```

---

## Charts

### BarChart

A bar chart component.

```go
chart := components.BarChart(components.BarChartProps{
    Data: []components.ChartDataPoint{
        {Label: "Jan", Value: 100},
        {Label: "Feb", Value: 150},
        {Label: "Mar", Value: 120},
    },
    Width:      "400px",
    Height:     "300px",
    Color:      "#3b82f6",
    ShowLabels: true,
    ShowValues: true,
})
```

### LineChart

A line chart component.

```go
chart := components.LineChart(components.LineChartProps{
    Data: []components.ChartDataPoint{
        {Label: "Jan", Value: 100},
        {Label: "Feb", Value: 150},
        {Label: "Mar", Value: 120},
    },
    Width:     "400px",
    Height:    "300px",
    Color:     "#3b82f6",
    Fill:      true,
    Smooth:    true,
    ShowDots:  true,
})
```

### PieChart

A pie/donut chart component.

```go
chart := components.PieChart(components.PieChartProps{
    Data: []components.ChartDataPoint{
        {Label: "Desktop", Value: 60, Color: "#3b82f6"},
        {Label: "Mobile", Value: 30, Color: "#10b981"},
        {Label: "Tablet", Value: 10, Color: "#f59e0b"},
    },
    Width:      "300px",
    Height:     "300px",
    Donut:      true,
    ShowLabels: true,
    ShowLegend: true,
})
```

### Sparkline

Inline mini charts.

```go
spark := components.Sparkline(components.SparklineProps{
    Data:    []float64{10, 20, 15, 30, 25, 40, 35},
    Type:    components.SparklineLine, // Line, Bar, Area
    Width:   "100px",
    Height:  "24px",
    Color:   "#3b82f6",
    ShowMin: true,
    ShowMax: true,
})

// Convenience functions
components.LineSparkline(data)
components.BarSparkline(data)
components.AreaSparkline(data)
components.TrendSparkline(data) // Shows min/max points
```

---

## Advanced Components

### VirtualList

A virtualized list for rendering large datasets efficiently.

```go
list := components.NewVirtualList(components.VirtualListProps{
    ItemCount:  10000,
    ItemHeight: 50,
    Height:     "400px",
    RenderItem: func(index int) js.Value {
        return renderListItem(index)
    },
})
```

### Dropdown

A dropdown menu component.

```go
dropdown := components.NewDropdown(components.DropdownProps{
    Trigger: triggerButton,
    Items: []components.DropdownItem{
        {Label: "Edit", Icon: "edit", OnClick: handleEdit},
        {Label: "Delete", Icon: "delete", OnClick: handleDelete, Danger: true},
        {Divider: true},
        {Label: "Settings", Icon: "settings", OnClick: handleSettings},
    },
    Position: "bottom-start",
})
```

### Inspector

A developer tool for inspecting the component tree.

```go
// Initialize in development
components.InitInspector()

// Or with custom props
inspector := components.NewInspector(components.InspectorProps{
    Position:  "bottom-right",
    Width:     "400px",
    Height:    "300px",
    Collapsed: true,
})
```

---

## Utilities

### Theme

Dark mode and theming support.

```go
// Initialize theme system
components.InitTheme()

// Toggle theme
components.ToggleTheme()

// Set specific theme
components.SetTheme(components.ThemeLight)
components.SetTheme(components.ThemeDark)
components.SetTheme(components.ThemeSystem)

// Check current theme
isDark := components.IsDarkMode()

// Subscribe to changes
unsubscribe := components.OnThemeChange(func(mode components.ThemeMode) {
    fmt.Println("Theme changed to:", mode)
})

// Theme toggle button
toggle := components.ThemeToggle()

// Theme selector dropdown
selector := components.ThemeSelector()
```

### Animation

Animation utilities and helpers.

```go
// Initialize animation CSS
components.InitAnimations()

// Apply animations
components.FadeIn(element, 300, onComplete)
components.FadeOut(element, 300, onComplete)
components.SlideIn(element, "right", 300, onComplete)
components.SlideOut(element, "left", 300, onComplete)
components.ScaleIn(element, 300, onComplete)
components.ScaleOut(element, 300, onComplete)

// Attention animations
components.Bounce(element)
components.Shake(element)
components.Pulse(element, 3) // iterations
components.Spin(element)
components.Wiggle(element)
components.Flash(element, 3)

// Staggered animations
components.Stagger(elements, components.Animation{
    Name:     "fadeIn",
    Duration: 300,
}, 100) // stagger delay

// Custom animation
components.Animate(components.AnimateProps{
    Element: element,
    Animation: components.Animation{
        Name:       "fadeIn",
        Duration:   300,
        Timing:     components.TimingEaseInOut,
        Delay:      100,
        Iterations: 1,
        Direction:  components.DirectionNormal,
    },
    OnComplete: func() {
        fmt.Println("Animation complete")
    },
})
```

### FocusTrap

Trap keyboard focus within a container (for modals, dialogs).

```go
trap := components.NewFocusTrap(container)

// Activate/deactivate
trap.Activate()
trap.Deactivate()
```

### SkipLinks

Accessibility skip navigation links.

```go
skipLinks := components.NewSkipLinks([]components.SkipLinkTarget{
    {ID: "main", Label: "Skip to main content"},
    {ID: "nav", Label: "Skip to navigation"},
})
```

---

## State Management

### Store

A generic reactive state container.

```go
import "goquery/state"

// Create a store
type AppState struct {
    User  *User
    Count int
}

store := state.New(AppState{Count: 0})

// Get state
currentState := store.Get()

// Set entire state
store.Set(AppState{Count: 5})

// Update state
store.Update(func(s *AppState) {
    s.Count++
})

// Subscribe to changes
unsubscribe := store.Subscribe(func(s AppState) {
    fmt.Println("State changed:", s)
})

// Derived stores
countStore := state.Derived(store, func(s AppState) int {
    return s.Count
})
```

### QueryCache

SWR-like data fetching with caching.

```go
import "goquery/state"

// Use the global cache
result := state.UseQuery("users", func() (any, error) {
    return fetchUsers()
}, state.QueryOptions{
    StaleTime:  5 * time.Minute,
    CacheTime:  10 * time.Minute,
    RetryCount: 3,
})

// Access result
if result.IsLoading {
    // Show loading
}
if result.IsError {
    // Show error
}
if result.IsSuccess {
    users := result.Data.([]User)
}

// Manual refetch
result.Refetch()

// Cache operations
cache := state.GetQueryCache()
cache.Invalidate("users")
cache.InvalidateAll()
cache.SetData("users", cachedUsers)
cache.Prefetch("posts", fetchPosts)
```

### Storage

localStorage and sessionStorage helpers.

```go
import "goquery/state"

// Persistent store (localStorage)
userStore := state.NewPersistentStore("user", User{Name: "Guest"})

// Session store (sessionStorage)
sessionStore := state.NewSessionStore("session", SessionData{})

// Manual storage operations
state.SetItem("key", "value")
value := state.GetItem("key")
state.RemoveItem("key")
```

### WebSocket

WebSocket support for real-time communication.

```go
import "goquery/state"

// Basic WebSocket
ws := state.NewWebSocket(state.WebSocketConfig{
    URL:               "wss://example.com/ws",
    ReconnectInterval: 3 * time.Second,
    MaxReconnects:     5,
    OnOpen: func() {
        fmt.Println("Connected")
    },
    OnMessage: func(data []byte) {
        fmt.Println("Received:", string(data))
    },
    OnClose: func(code int, reason string) {
        fmt.Println("Closed:", code, reason)
    },
    OnError: func(err string) {
        fmt.Println("Error:", err)
    },
})

ws.Connect()
ws.Send([]byte("Hello"))
ws.SendJSON(map[string]any{"type": "ping"})
ws.Close()

// Typed message handlers
ws.On("chat", func(data []byte) {
    var msg ChatMessage
    json.Unmarshal(data, &msg)
})

// WebSocket with integrated state
wss := state.NewWebSocketStore(state.WebSocketConfig{URL: "wss://example.com/ws"})

wss.Connect()
wss.Subscribe(func(s state.WSStoreState) {
    if s.Connected {
        fmt.Println("Connected!")
    }
})

// Access state
state := wss.State()
fmt.Println("Connected:", state.Connected)
fmt.Println("Messages:", state.MessageCount)
```

---

## Building & Running

```bash
# Build WASM
GOOS=js GOARCH=wasm go build -o main.wasm ./app

# With TinyGo (smaller output)
tinygo build -o main.wasm -target wasm ./app

# Run development server
cd example && make dev
```

## License

MIT
