# Keyboard Shortcuts

This document provides a complete reference for all keyboard navigation patterns implemented in Gux components. All components support full keyboard accessibility following WCAG 2.1 guidelines.

> **Platform Note:** Throughout this document, `Cmd` refers to the Command key on macOS and `Ctrl` refers to the Control key on Windows/Linux. Shortcuts like `Cmd/Ctrl+K` work on both platforms.

## Table of Contents

- [Global Shortcuts](#global-shortcuts)
- [Modal & Dialog](#modal--dialog)
- [Command Palette](#command-palette)
- [Tabs](#tabs)
- [DatePicker](#datepicker)
- [Dropdown & Menu](#dropdown--menu)
- [General Navigation](#general-navigation)
- [Accessibility Notes](#accessibility-notes)

## Global Shortcuts

Application-wide keyboard shortcuts available anywhere in the interface.

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl+K` | Open command palette |
| `Cmd/Ctrl+B` | Toggle sidebar visibility |

### Usage

```go
// Command palette registers its own global shortcut
palette := components.NewCommandPalette(props)
palette.RegisterKeyboardShortcut() // Registers Cmd/Ctrl+K

// Cleanup when done
palette.UnregisterKeyboardShortcut()
```

## Modal & Dialog

Keyboard navigation for modal dialogs, including focus trapping.

| Shortcut | Action |
|----------|--------|
| `Escape` | Close the modal |
| `Tab` | Move focus to next focusable element (within modal) |
| `Shift+Tab` | Move focus to previous focusable element (within modal) |

### Focus Trap Behavior

When a modal opens:
1. Focus automatically moves to the first focusable element inside the modal
2. Tab/Shift+Tab cycles only through elements within the modal
3. Focus cannot escape to elements behind the modal overlay
4. When modal closes, focus returns to the element that triggered it

### Usage

```go
modal := components.Modal(components.ModalProps{
    Title:      "Confirm Action",
    CloseOnEsc: true, // Enable Escape key to close
    Content:    myContent,
})

// FocusTrap is automatically applied
modal.Open()
```

See also: [Modal Component](components.md#modal), [ConfirmDialog Component](components.md#confirmdialog)

## Command Palette

Searchable command launcher for quick actions.

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl+K` | Open/close command palette |
| `↑` (Up Arrow) | Navigate to previous command |
| `↓` (Down Arrow) | Navigate to next command |
| `Enter` | Execute selected command |
| `Escape` | Close palette |

### Usage

```go
palette := components.NewCommandPalette(components.CommandPaletteProps{
    Commands: []components.Command{
        {ID: "new-post", Label: "Create New Post", OnExecute: createPost},
        {ID: "settings", Label: "Open Settings", OnExecute: openSettings},
    },
})

// Register global Cmd/Ctrl+K shortcut
palette.RegisterKeyboardShortcut()

// Mount to DOM
document.Get("body").Call("appendChild", palette.Element())
```

See also: [CommandPalette Component](components.md#commandpalette)

## Tabs

Keyboard navigation for tabbed interfaces using roving tabindex pattern.

| Shortcut | Action |
|----------|--------|
| `←` (Left Arrow) | Move to previous tab |
| `→` (Right Arrow) | Move to next tab |
| `Home` | Jump to first tab |
| `End` | Jump to last tab |

### Navigation Behavior

- Arrow keys move focus AND activate the tab
- Focus wraps: Right from last tab goes to first, Left from first goes to last
- Home/End provide quick access to first and last tabs

### Usage

```go
tabs := components.Tabs(components.TabsProps{
    Tabs: []components.Tab{
        {Label: "Profile", Content: profileContent},
        {Label: "Settings", Content: settingsContent},
        {Label: "Security", Content: securityContent},
    },
})
```

See also: [Tabs Component](components.md#tabs)

## DatePicker

Full keyboard navigation for the calendar grid.

| Shortcut | Action |
|----------|--------|
| `←` (Left Arrow) | Move to previous day |
| `→` (Right Arrow) | Move to next day |
| `↑` (Up Arrow) | Move to same day previous week |
| `↓` (Down Arrow) | Move to same day next week |
| `Enter` | Select focused date |
| `Space` | Select focused date |
| `Escape` | Close calendar popup |

### Grid Navigation

The calendar uses a grid pattern where:
- Arrow keys move through days in a logical manner
- Moving past month boundaries automatically navigates to adjacent months
- Enter/Space confirm the selection and close the picker

### Usage

```go
picker := components.DatePicker(components.DatePickerProps{
    Label:       "Start Date",
    Placeholder: "Select date...",
    OnChange:    func(date time.Time) { /* handle */ },
})
```

See also: [DatePicker Component](components.md#datepicker)

## Dropdown & Menu

Keyboard navigation for dropdown menus and select components.

| Shortcut | Action |
|----------|--------|
| `↑` (Up Arrow) | Navigate to previous item |
| `↓` (Down Arrow) | Navigate to next item |
| `Enter` | Select highlighted item |
| `Escape` | Close dropdown |

### Components Using This Pattern

- **Dropdown** - General dropdown menus
- **Select** - Form select inputs
- **Combobox** - Searchable dropdown with type-ahead
- **UserMenu** - User profile dropdown
- **NotificationCenter** - Notification dropdown

### Usage

```go
// Dropdown example
dropdown := components.Dropdown(components.DropdownProps{
    Trigger: triggerButton,
    Items: []components.DropdownItem{
        {Label: "Edit", OnClick: handleEdit},
        {Label: "Delete", OnClick: handleDelete},
    },
})

// Combobox with keyboard search
combo := components.Combobox(components.ComboboxProps{
    Label:       "Assign to",
    Placeholder: "Search users...",
    Options:     userOptions,
    OnChange:    handleAssign,
})
```

See also: [Select Component](components.md#select), [Combobox Component](components.md#combobox), [UserMenu Component](components.md#usermenu)

## General Navigation

Standard keyboard patterns for page navigation.

| Shortcut | Action |
|----------|--------|
| `Tab` | Move focus to next focusable element |
| `Shift+Tab` | Move focus to previous focusable element |
| `Enter` | Activate skip link (when focused) |

### Skip Links

Skip links allow keyboard users to bypass repetitive navigation and jump directly to main content.

```go
// Add skip links at the start of your app
skipLinks := components.SkipLinks()
document.Get("body").Call("prepend", skipLinks)

// Skip links target:
// - #main-content - Jump to main content area
```

### Focus Indicators

All interactive elements show visible focus indicators:
- Default ring style: `focus:ring-2 focus:ring-blue-500`
- Dark mode support: `dark:focus:ring-blue-400`

See also: [Accessibility Documentation](accessibility.md)

## Accessibility Notes

### Screen Reader Announcements

Keyboard actions trigger appropriate screen reader announcements via ARIA live regions:
- Command palette results update (`aria-live="polite"`)
- Tab selection changes announced
- Modal open/close announced (`role="dialog"`, `aria-modal="true"`)
- Connection status changes announced

### Focus Management

Gux components implement proper focus management:

1. **FocusTrap** - Modals and dialogs trap focus within their boundaries
2. **Focus Restoration** - When overlays close, focus returns to the trigger element
3. **Roving Tabindex** - Tabs, menus use roving tabindex for arrow key navigation
4. **Skip Links** - Allow bypassing navigation to reach main content

### Implementation Example

```go
// FocusTrap for custom modal content
trap := components.FocusTrap(modalContent)

// The trap ensures:
// - Focus stays within modalContent
// - Tab wraps from last to first element
// - Shift+Tab wraps from first to last element
```

### Testing Keyboard Navigation

Verify keyboard accessibility:
1. Unplug your mouse and navigate using only keyboard
2. Check that all interactive elements are reachable via Tab
3. Verify focus indicators are always visible
4. Test that Escape closes all overlays
5. Confirm arrow key navigation in menus and tabs

See also: [Accessibility Guide](accessibility.md)
