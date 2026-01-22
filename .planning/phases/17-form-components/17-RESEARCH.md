# Phase 17: Form Components - Research

**Researched:** 2026-01-22
**Domain:** Go/WASM Form Components with core.Node and Tailwind CSS
**Confidence:** HIGH

## Summary

This research establishes patterns for building form components on top of Gux's `core.Node` system, extending the Phase 16 component foundation. The goal is to create reusable, accessible form components (Input, Textarea, Select, Checkbox, Radio, Switch, Form) that work identically in SSR and WASM environments.

The codebase already has two form component systems: the **old `components/` package** (WASM-only, uses `js.Value` directly with struct-based stateful components like `*Input`, `*Select`) and the **new `core/` system** (functional components returning `core.Node`). The `examples/minimal/admin/` directory demonstrates the current pattern with a `formField()` helper function that composes `core.Label` and `core.Input`.

The key insight is that form components in the new system should be **pure functions returning `core.Node`**, following the Phase 16 Button/Card patterns. State management happens at the page level using `r.StateString()` etc., not inside components. This enables SSR compatibility and clean separation of concerns.

**Primary recommendation:** Create form components in the `ui/` package following the established Props struct pattern. Components handle presentation (labels, styling, error display) while pages handle state and validation. Use compound patterns for complex components (RadioGroup + RadioOption).

## Standard Stack

The established libraries/tools for this domain:

### Core (Already in codebase)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| core.Node | - | Abstract UI tree | Enables SSR + WASM rendering |
| core.Input | - | Raw input element | Base for Input component |
| core.Select | - | Raw select element | Base for Select component |
| core.Textarea | - | Raw textarea element | Base for Textarea component |
| core.Label | - | Label element | Accessibility association |
| core.Attrs | - | Element attributes | OnChange, OnEnter handlers |
| ui.MergeClasses | - | Class concatenation | From Phase 16 |
| ui.ConditionalClass | - | Conditional classes | From Phase 16 |

### Supporting (From Phase 16)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| ui.Button | - | Form submit buttons | Form actions |
| ui.VStack | - | Vertical layout | Form field stacking |
| ui.HStack | - | Horizontal layout | Inline fields, button groups |

### Not Needed
| Library | Why Not |
|---------|---------|
| Form validation library | Use simple Go functions, validation shown in examples |
| CSS form library | Tailwind provides all needed utilities |
| Stateful form manager | State lives at page level via Router |

**Installation:**
No new dependencies required. All components will use existing `core` and `ui` packages.

## Component Architecture

### Component Inventory

Based on requirements, the following components are needed:

| Component | Type | Props Pattern | Notes |
|-----------|------|---------------|-------|
| Input | Simple | InputProps | Text, email, password, number, etc. |
| Textarea | Simple | TextareaProps | Multi-line text with resize control |
| Select | Compound | SelectProps + SelectOption | Dropdown with options |
| Checkbox | Simple | CheckboxProps | Boolean toggle with label |
| RadioGroup | Compound | RadioGroupProps + RadioOption | Mutually exclusive options |
| Switch | Simple | SwitchProps | Toggle switch (styled checkbox) |
| FormField | Wrapper | FormFieldProps | Label + input + error wrapper |
| FormMessage | Simple | FormMessageProps | Success/error messages |

### Recommended Project Structure

```
ui/
├── utils.go           # MergeClasses, ConditionalClass (Phase 16)
├── button.go          # Button component (Phase 16)
├── card.go            # Card components (Phase 16)
├── layout.go          # Layout components (Phase 16)
├── input.go           # Input, InputProps
├── textarea.go        # Textarea, TextareaProps
├── select.go          # Select, SelectProps, SelectOption
├── checkbox.go        # Checkbox, CheckboxProps
├── radio.go           # RadioGroup, RadioGroupProps, RadioOption, RadioOptionProps
├── switch.go          # Switch, SwitchProps
├── form.go            # FormField, FormFieldProps, FormMessage
└── *_test.go          # Tests for each component
```

## Props Patterns

### Pattern 1: Simple Form Input

**What:** Single input field with label support
**When to use:** Text inputs, email, password, number

```go
// InputType defines the HTML input type
type InputType string

const (
    InputText     InputType = "text"
    InputEmail    InputType = "email"
    InputPassword InputType = "password"
    InputNumber   InputType = "number"
    InputSearch   InputType = "search"
    InputTel      InputType = "tel"
    InputURL      InputType = "url"
)

// InputSize matches ButtonSize for consistency
type InputSize string

const (
    InputSM InputSize = "sm"
    InputMD InputSize = "md"
    InputLG InputSize = "lg"
)

// InputProps configures the Input component
type InputProps struct {
    Type        InputType // Input type (default: text)
    Size        InputSize // Size variant (default: md)
    Name        string    // Input name attribute
    Value       string    // Current value
    Placeholder string    // Placeholder text
    Class       string    // Additional classes
    Disabled    bool      // Disabled state
    Required    bool      // Required indicator
    Error       string    // Error message to display
    OnChange    func(string) // Value change handler (WASM only)
    OnEnter     func()    // Enter key handler (WASM only)
}

func Input(props InputProps) core.Node {
    inputType := props.Type
    if inputType == "" {
        inputType = InputText
    }
    size := props.Size
    if size == "" {
        size = InputMD
    }

    class := MergeClasses(
        inputBaseClasses,
        inputSizeClasses[size],
        ConditionalClass(props.Error != "", inputErrorClasses),
        ConditionalClass(props.Disabled, inputDisabledClasses),
        props.Class,
    )

    attrs := core.Attrs{
        Type:     string(inputType),
        Name:     props.Name,
        Value:    props.Value,
        Class:    class,
        OnChange: props.OnChange,
        OnEnter:  props.OnEnter,
    }

    if props.Placeholder != "" {
        attrs.Extra = map[string]string{"placeholder": props.Placeholder}
    }
    if props.Disabled {
        if attrs.Extra == nil {
            attrs.Extra = make(map[string]string)
        }
        attrs.Extra["disabled"] = "disabled"
    }
    if props.Required {
        if attrs.Extra == nil {
            attrs.Extra = make(map[string]string)
        }
        attrs.Extra["required"] = "required"
        attrs.Extra["aria-required"] = "true"
    }
    if props.Error != "" {
        if attrs.Extra == nil {
            attrs.Extra = make(map[string]string)
        }
        attrs.Extra["aria-invalid"] = "true"
    }

    return core.Input(attrs)
}
```

### Pattern 2: FormField Wrapper

**What:** Combines label, input, and error display
**When to use:** Standard form layouts with consistent structure

```go
// FormFieldProps configures a form field wrapper
type FormFieldProps struct {
    Label       string      // Label text
    LabelFor    string      // ID to associate label with input
    Required    bool        // Show required indicator
    Error       string      // Error message
    Description string      // Help text below input
    Class       string      // Additional wrapper classes
    Children    []core.Node // Input component(s)
}

func FormField(props FormFieldProps) core.Node {
    children := []core.Node{}

    // Label
    if props.Label != "" {
        labelClass := MergeClasses(
            "block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1",
        )
        labelChildren := []core.Node{core.Text(props.Label)}
        if props.Required {
            labelChildren = append(labelChildren,
                core.Span(core.Class("text-red-500 ml-1"), core.Text("*")),
            )
        }

        attrs := core.Attrs{Class: labelClass}
        if props.LabelFor != "" {
            attrs.Extra = map[string]string{"for": props.LabelFor}
        }
        children = append(children, core.Label(attrs, labelChildren...))
    }

    // Input(s)
    children = append(children, props.Children...)

    // Error message
    if props.Error != "" {
        children = append(children,
            core.P(core.Attrs{
                Class: "text-sm text-red-600 dark:text-red-400 mt-1",
                Extra: map[string]string{"role": "alert"},
            }, core.Text(props.Error)),
        )
    }

    // Description
    if props.Description != "" && props.Error == "" {
        children = append(children,
            core.P(core.Class("text-sm text-gray-500 dark:text-gray-400 mt-1"),
                core.Text(props.Description),
            ),
        )
    }

    class := MergeClasses("mb-4", props.Class)
    return core.Div(core.Class(class), children...)
}
```

### Pattern 3: Compound Component (Select)

**What:** Parent component with typed option children
**When to use:** Select dropdowns, radio groups

```go
// SelectOption represents a single option
type SelectOption struct {
    Value    string
    Label    string
    Disabled bool
}

// SelectProps configures the Select component
type SelectProps struct {
    Name        string         // Select name attribute
    Value       string         // Currently selected value
    Options     []SelectOption // Available options
    Placeholder string         // Placeholder option text
    Size        InputSize      // Size variant (default: md)
    Class       string         // Additional classes
    Disabled    bool           // Disabled state
    Required    bool           // Required indicator
    Error       string         // Error message
    OnChange    func(string)   // Selection change handler (WASM only)
}

func Select(props SelectProps) core.Node {
    size := props.Size
    if size == "" {
        size = InputMD
    }

    class := MergeClasses(
        selectBaseClasses,
        inputSizeClasses[size],
        ConditionalClass(props.Error != "", inputErrorClasses),
        ConditionalClass(props.Disabled, inputDisabledClasses),
        props.Class,
    )

    attrs := core.Attrs{
        Name:     props.Name,
        Class:    class,
        OnChange: props.OnChange,
    }

    if props.Disabled {
        attrs.Extra = map[string]string{"disabled": "disabled"}
    }
    if props.Required {
        if attrs.Extra == nil {
            attrs.Extra = make(map[string]string)
        }
        attrs.Extra["required"] = "required"
    }
    if props.Error != "" {
        if attrs.Extra == nil {
            attrs.Extra = make(map[string]string)
        }
        attrs.Extra["aria-invalid"] = "true"
    }

    // Build option nodes
    var options []core.Node

    // Placeholder option
    if props.Placeholder != "" {
        options = append(options, core.Option(
            core.Attrs{
                Value: "",
                Extra: map[string]string{
                    "disabled": "disabled",
                    "selected": boolToSelected(props.Value == ""),
                },
            },
            core.Text(props.Placeholder),
        ))
    }

    // Regular options
    for _, opt := range props.Options {
        optAttrs := core.Attrs{Value: opt.Value}
        if opt.Value == props.Value {
            optAttrs.Extra = map[string]string{"selected": "selected"}
        }
        if opt.Disabled {
            if optAttrs.Extra == nil {
                optAttrs.Extra = make(map[string]string)
            }
            optAttrs.Extra["disabled"] = "disabled"
        }
        options = append(options, core.Option(optAttrs, core.Text(opt.Label)))
    }

    return core.Select(attrs, options...)
}
```

### Pattern 4: RadioGroup with Compound Children

**What:** Group of mutually exclusive options
**When to use:** Single selection from small option set

```go
// RadioOptionProps configures a single radio option
type RadioOptionProps struct {
    Value       string // Option value
    Label       string // Display label
    Description string // Optional description text
    Disabled    bool   // Disabled state
}

// RadioGroupProps configures the RadioGroup component
type RadioGroupProps struct {
    Name      string             // Group name (required)
    Value     string             // Currently selected value
    Options   []RadioOptionProps // Available options
    Inline    bool               // Display horizontally
    Class     string             // Additional wrapper classes
    Disabled  bool               // Disable all options
    Required  bool               // Required indicator
    Error     string             // Error message
    OnChange  func(string)       // Selection change handler (WASM only)
}

func RadioGroup(props RadioGroupProps) core.Node {
    wrapperClass := MergeClasses(
        ConditionalClass(props.Inline, "flex flex-row gap-4"),
        ConditionalClass(!props.Inline, "flex flex-col gap-2"),
        props.Class,
    )

    var children []core.Node
    for _, opt := range props.Options {
        isChecked := opt.Value == props.Value
        isDisabled := props.Disabled || opt.Disabled

        radioClass := MergeClasses(
            "flex items-start gap-2",
            ConditionalClass(isDisabled, "opacity-50 cursor-not-allowed"),
        )

        inputAttrs := core.Attrs{
            Type:  "radio",
            Name:  props.Name,
            Value: opt.Value,
            Class: "mt-1 h-4 w-4 text-blue-600 border-gray-300 focus:ring-blue-500",
        }
        if isChecked {
            inputAttrs.Extra = map[string]string{"checked": "checked"}
        }
        if isDisabled {
            if inputAttrs.Extra == nil {
                inputAttrs.Extra = make(map[string]string)
            }
            inputAttrs.Extra["disabled"] = "disabled"
        }
        // OnChange for radio is typically handled by the group
        // Core Attrs.OnChange works on change event

        labelContent := []core.Node{
            core.Span(core.Class("text-sm font-medium text-gray-700 dark:text-gray-300"),
                core.Text(opt.Label),
            ),
        }
        if opt.Description != "" {
            labelContent = append(labelContent,
                core.Span(core.Class("block text-sm text-gray-500 dark:text-gray-400"),
                    core.Text(opt.Description),
                ),
            )
        }

        children = append(children,
            core.Label(core.Class(radioClass),
                core.Input(inputAttrs),
                core.Div(core.Attrs{}, labelContent...),
            ),
        )
    }

    return core.Div(core.Class(wrapperClass), children...)
}
```

## State Management

### Form State Pattern

Forms in Gux use page-level state, not component-level state:

```go
func UserForm(r *core.Router) func() core.Node {
    return func() core.Node {
        // State at page level
        nameState := r.StateString("name", "")
        emailState := r.StateString("email", "")
        errorState := r.StateString("error", "")

        // Validation at page level
        validate := func() bool {
            if nameState.Get() == "" {
                errorState.Set("Name is required")
                return false
            }
            errorState.Set("")
            return true
        }

        // Submit handler at page level
        submit := func() {
            if validate() {
                // API call
            }
        }

        // Components just render
        return ui.VStack(ui.StackProps{
            Children: []core.Node{
                ui.FormField(ui.FormFieldProps{
                    Label:    "Name",
                    Required: true,
                    Error:    errorState.Get(),
                    Children: []core.Node{
                        ui.Input(ui.InputProps{
                            Name:     "name",
                            Value:    nameState.Get(),
                            OnChange: func(v string) { nameState.Set(v) },
                            OnEnter:  submit,
                            Error:    errorState.Get(),
                        }),
                    },
                }),
                ui.Button(ui.ButtonProps{
                    Type:     "submit",
                    OnClick:  submit,
                    Children: []core.Node{core.Text("Submit")},
                }),
            },
        })
    }
}
```

### Controlled vs Uncontrolled

All form components are **controlled** - they receive `Value` prop and call `OnChange`. This matches React's pattern and ensures SSR/WASM consistency:

- **SSR:** Value rendered in HTML, OnChange ignored
- **WASM:** Value sets initial state, OnChange updates state

The `core.Attrs.OnChange` handler fires on the `change` event (blur), not `input` event (every keystroke). This prevents excessive re-renders. The `OnEnter` handler captures Enter keypress for form submission.

## Validation Patterns

### Client-Side Validation

Keep validation simple and at the page level:

```go
// Validation functions in the page
func validateEmail(email string) string {
    if email == "" {
        return "Email is required"
    }
    if !strings.Contains(email, "@") {
        return "Invalid email format"
    }
    return ""
}

// In component usage
emailError := validateEmail(emailState.Get())
ui.FormField(ui.FormFieldProps{
    Label: "Email",
    Error: emailError,
    Children: []core.Node{
        ui.Input(ui.InputProps{
            Type:  ui.InputEmail,
            Value: emailState.Get(),
            Error: emailError, // For aria-invalid
            OnChange: func(v string) { emailState.Set(v) },
        }),
    },
})
```

### Server-Side Validation

Handle API errors at the page level:

```go
api.Users.Create(data, func(result *dto.UserDetail, err error) {
    if err != nil {
        // Parse field-specific errors if available
        errorState.Set(err.Error())
    } else {
        r.Navigate("/users")
    }
})
```

## Accessibility Requirements

### ARIA Attributes

| Attribute | When to Use | Component |
|-----------|-------------|-----------|
| `aria-required="true"` | Required fields | Input, Select, Textarea |
| `aria-invalid="true"` | Field has error | Input, Select, Textarea |
| `aria-describedby` | Link to error/description | Input, Select, Textarea |
| `role="alert"` | Error messages | Error text element |
| `role="radiogroup"` | Radio button container | RadioGroup |
| `aria-checked` | Switch state | Switch |

### Label Association

Always associate labels with inputs via `for`/`id`:

```go
// Generate unique ID if needed
inputID := props.ID
if inputID == "" && props.Name != "" {
    inputID = "input-" + props.Name
}

// In FormField
core.Label(core.Attrs{
    Extra: map[string]string{"for": inputID},
}, core.Text(props.Label))

// In Input
core.Input(core.Attrs{
    ID: inputID,
    // ...
})
```

### Keyboard Navigation

- **Tab:** Move between form fields
- **Enter:** Submit form (via OnEnter handler)
- **Space:** Toggle checkbox/switch, select radio
- **Arrow keys:** Navigate radio group options

## Tailwind Styling

### Base Input Classes

```go
const inputBaseClasses = "w-full border rounded-md shadow-sm " +
    "bg-white dark:bg-gray-800 " +
    "text-gray-900 dark:text-gray-100 " +
    "border-gray-300 dark:border-gray-600 " +
    "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 " +
    "placeholder:text-gray-400 dark:placeholder:text-gray-500"

var inputSizeClasses = map[InputSize]string{
    InputSM: "px-2 py-1 text-sm",
    InputMD: "px-3 py-2 text-base",
    InputLG: "px-4 py-3 text-lg",
}

const inputErrorClasses = "border-red-500 dark:border-red-400 " +
    "focus:ring-red-500 focus:border-red-500"

const inputDisabledClasses = "bg-gray-100 dark:bg-gray-700 " +
    "cursor-not-allowed opacity-50"
```

### Select Classes

```go
const selectBaseClasses = inputBaseClasses + " appearance-none " +
    "bg-no-repeat bg-right " +
    "pr-10" // Space for dropdown arrow

// Dropdown arrow using background image or pseudo-element
```

### Checkbox/Radio Classes

```go
const checkboxBaseClasses = "h-4 w-4 rounded " +
    "text-blue-600 " +
    "border-gray-300 dark:border-gray-600 " +
    "focus:ring-blue-500 focus:ring-offset-0 " +
    "bg-white dark:bg-gray-800"

const radioBaseClasses = "h-4 w-4 " +
    "text-blue-600 " +
    "border-gray-300 dark:border-gray-600 " +
    "focus:ring-blue-500 focus:ring-offset-0 " +
    "bg-white dark:bg-gray-800"
```

### Switch Classes

```go
// Track (background)
const switchTrackClasses = "relative inline-flex h-6 w-11 items-center rounded-full " +
    "transition-colors duration-200 ease-in-out " +
    "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"

const switchTrackOnClasses = "bg-blue-600"
const switchTrackOffClasses = "bg-gray-300 dark:bg-gray-600"

// Knob (moving circle)
const switchKnobClasses = "inline-block h-4 w-4 rounded-full bg-white " +
    "transform transition-transform duration-200 ease-in-out shadow"

const switchKnobOnClasses = "translate-x-6"
const switchKnobOffClasses = "translate-x-1"
```

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Form field wrapper | Inline label/input/error | `FormField` component | Consistency, accessibility |
| Unique IDs | Manual ID generation | Auto-generate in component | Avoids collisions |
| Error styling | Conditional class logic | Props pattern with error | Single source of truth |
| Radio state | Manual checked logic | RadioGroup with value | Handles exclusivity |
| Select options | Manual option building | SelectOption slice | Type safety |

**Key insight:** Form components should compose `core.*` primitives (Input, Select, Label) with styling and accessibility, not replace them. This keeps the abstraction thin and predictable.

## Common Pitfalls

### Pitfall 1: SSR Checkbox/Radio "checked" Attribute

**What goes wrong:** Checked state not rendering correctly in SSR
**Why it happens:** Boolean attributes in HTML are presence-based, not value-based
**How to avoid:** Use `Extra` map with "checked" key only when true

```go
// WRONG
attrs.Extra = map[string]string{"checked": fmt.Sprint(isChecked)}

// CORRECT
if isChecked {
    attrs.Extra = map[string]string{"checked": "checked"}
}
```

**Warning signs:** Checkboxes always appear unchecked on initial SSR load

### Pitfall 2: Select "selected" Option Hydration

**What goes wrong:** Selected option mismatches between SSR and WASM
**Why it happens:** Default browser selection differs from value prop
**How to avoid:** Always explicitly set selected attribute on correct option

```go
for _, opt := range props.Options {
    optAttrs := core.Attrs{Value: opt.Value}
    if opt.Value == props.Value {
        optAttrs.Extra = map[string]string{"selected": "selected"}
    }
    options = append(options, core.Option(optAttrs, core.Text(opt.Label)))
}
```

**Warning signs:** Wrong option appears selected after page load

### Pitfall 3: Placeholder vs Required

**What goes wrong:** Required validation fails when placeholder is selected
**Why it happens:** Placeholder option has empty value, validation sees empty string
**How to avoid:** Placeholder option should be disabled, use `required` attribute

```go
// Placeholder option
core.Option(
    core.Attrs{
        Value: "",
        Extra: map[string]string{
            "disabled": "disabled", // Can't submit with this selected
            "selected": boolToSelected(props.Value == ""),
        },
    },
    core.Text(props.Placeholder),
)
```

**Warning signs:** Form submits with "Select..." placeholder value

### Pitfall 4: OnChange Event Timing

**What goes wrong:** State updates but UI doesn't reflect changes
**Why it happens:** `core.Attrs.OnChange` fires on `change` (blur), not `input`
**How to avoid:** This is intentional for performance. Use OnEnter for immediate actions.

```go
// This updates on blur, which is correct
OnChange: func(v string) { nameState.Set(v) },

// For immediate action (like search), consider debouncing at page level
// or accept the limitation
```

**Warning signs:** Input value seems "laggy" or delayed (this is expected behavior)

### Pitfall 5: Textarea Value vs Children

**What goes wrong:** Textarea content not rendering in SSR
**Why it happens:** `core.Textarea` expects children for initial content in SSR
**How to avoid:** Pass value as both Value attr and child Text node

```go
// SSR needs children, WASM uses value attribute
core.Textarea(
    core.Attrs{
        Name:     props.Name,
        Value:    props.Value, // For WASM
        OnChange: props.OnChange,
    },
    core.Text(props.Value), // For SSR
)
```

**Warning signs:** Textarea empty on initial SSR, populated after hydration

## Code Examples

### Complete Input Component

```go
package ui

import "github.com/dougbarrett/gux/core"

// InputType defines the HTML input type.
type InputType string

const (
    InputText     InputType = "text"
    InputEmail    InputType = "email"
    InputPassword InputType = "password"
    InputNumber   InputType = "number"
    InputSearch   InputType = "search"
    InputTel      InputType = "tel"
    InputURL      InputType = "url"
)

// InputSize defines the size of an input.
type InputSize string

const (
    InputSM InputSize = "sm"
    InputMD InputSize = "md"
    InputLG InputSize = "lg"
)

var inputSizeClasses = map[InputSize]string{
    InputSM: "px-2 py-1.5 text-sm",
    InputMD: "px-3 py-2 text-base",
    InputLG: "px-4 py-3 text-lg",
}

const inputBaseClasses = "w-full border rounded-md shadow-sm " +
    "bg-white dark:bg-gray-800 " +
    "text-gray-900 dark:text-gray-100 " +
    "border-gray-300 dark:border-gray-600 " +
    "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 " +
    "placeholder:text-gray-400 dark:placeholder:text-gray-500"

const inputErrorClasses = "border-red-500 dark:border-red-400 " +
    "focus:ring-red-500 focus:border-red-500"

const inputDisabledClasses = "bg-gray-100 dark:bg-gray-700 cursor-not-allowed opacity-50"

// InputProps configures the Input component.
type InputProps struct {
    Type        InputType    // Input type (default: text)
    Size        InputSize    // Size variant (default: md)
    ID          string       // Element ID for label association
    Name        string       // Input name attribute
    Value       string       // Current value
    Placeholder string       // Placeholder text
    Class       string       // Additional classes
    Disabled    bool         // Disabled state
    Required    bool         // Required indicator
    Error       string       // Error message (adds aria-invalid)
    OnChange    func(string) // Value change handler (WASM only)
    OnEnter     func()       // Enter key handler (WASM only)
}

// Input creates a styled text input component.
//
// Example:
//
//     Input(InputProps{
//         Type:        InputEmail,
//         Name:        "email",
//         Value:       emailState.Get(),
//         Placeholder: "you@example.com",
//         OnChange:    func(v string) { emailState.Set(v) },
//     })
func Input(props InputProps) core.Node {
    inputType := props.Type
    if inputType == "" {
        inputType = InputText
    }

    size := props.Size
    if size == "" {
        size = InputMD
    }

    class := MergeClasses(
        inputBaseClasses,
        inputSizeClasses[size],
        ConditionalClass(props.Error != "", inputErrorClasses),
        ConditionalClass(props.Disabled, inputDisabledClasses),
        props.Class,
    )

    attrs := core.Attrs{
        ID:       props.ID,
        Type:     string(inputType),
        Name:     props.Name,
        Value:    props.Value,
        Class:    class,
        OnChange: props.OnChange,
        OnEnter:  props.OnEnter,
    }

    // Build extra attributes
    extra := make(map[string]string)
    if props.Placeholder != "" {
        extra["placeholder"] = props.Placeholder
    }
    if props.Disabled {
        extra["disabled"] = "disabled"
    }
    if props.Required {
        extra["required"] = "required"
        extra["aria-required"] = "true"
    }
    if props.Error != "" {
        extra["aria-invalid"] = "true"
    }
    if len(extra) > 0 {
        attrs.Extra = extra
    }

    return core.Input(attrs)
}
```

### Complete FormField Component

```go
package ui

import "github.com/dougbarrett/gux/core"

// FormFieldProps configures the FormField wrapper component.
type FormFieldProps struct {
    Label       string      // Label text
    LabelFor    string      // ID to associate label with input
    Required    bool        // Show required indicator
    Error       string      // Error message to display
    Description string      // Help text below input
    Class       string      // Additional wrapper classes
    Children    []core.Node // Input component(s)
}

// FormField creates a form field wrapper with label, input, and error display.
//
// Example:
//
//     FormField(FormFieldProps{
//         Label:    "Email",
//         LabelFor: "email-input",
//         Required: true,
//         Error:    emailError,
//         Children: []core.Node{
//             Input(InputProps{
//                 ID:   "email-input",
//                 Type: InputEmail,
//                 Name: "email",
//             }),
//         },
//     })
func FormField(props FormFieldProps) core.Node {
    var children []core.Node

    // Label
    if props.Label != "" {
        labelClass := "block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"

        labelChildren := []core.Node{core.Text(props.Label)}
        if props.Required {
            labelChildren = append(labelChildren,
                core.Span(core.Class("text-red-500 ml-1"), core.Text("*")),
            )
        }

        labelAttrs := core.Attrs{Class: labelClass}
        if props.LabelFor != "" {
            labelAttrs.Extra = map[string]string{"for": props.LabelFor}
        }

        children = append(children, core.Label(labelAttrs, labelChildren...))
    }

    // Input(s)
    children = append(children, props.Children...)

    // Error message
    if props.Error != "" {
        children = append(children,
            core.P(
                core.Attrs{
                    Class: "text-sm text-red-600 dark:text-red-400 mt-1",
                    Extra: map[string]string{"role": "alert"},
                },
                core.Text(props.Error),
            ),
        )
    } else if props.Description != "" {
        // Description (only shown when no error)
        children = append(children,
            core.P(
                core.Class("text-sm text-gray-500 dark:text-gray-400 mt-1"),
                core.Text(props.Description),
            ),
        )
    }

    class := MergeClasses("mb-4", props.Class)
    return core.Div(core.Class(class), children...)
}
```

### Complete Checkbox Component

```go
package ui

import "github.com/dougbarrett/gux/core"

// CheckboxProps configures the Checkbox component.
type CheckboxProps struct {
    ID          string     // Element ID for label association
    Name        string     // Input name attribute
    Checked     bool       // Checked state
    Label       string     // Label text (optional)
    Description string     // Description text (optional)
    Class       string     // Additional classes
    Disabled    bool       // Disabled state
    OnChange    func(bool) // State change handler (WASM only) - NOT SUPPORTED IN CORE YET
}

// Checkbox creates a checkbox input with optional label.
//
// Note: core.Attrs does not currently support OnChange for checkboxes
// returning bool. Use a wrapper with JavaScript for now.
//
// Example:
//
//     Checkbox(CheckboxProps{
//         ID:      "remember",
//         Name:    "remember",
//         Checked: rememberState.Get(),
//         Label:   "Remember me",
//     })
func Checkbox(props CheckboxProps) core.Node {
    checkboxClass := MergeClasses(
        "h-4 w-4 rounded",
        "text-blue-600",
        "border-gray-300 dark:border-gray-600",
        "focus:ring-blue-500 focus:ring-offset-0",
        "bg-white dark:bg-gray-800",
        ConditionalClass(props.Disabled, "cursor-not-allowed opacity-50"),
    )

    attrs := core.Attrs{
        ID:    props.ID,
        Type:  "checkbox",
        Name:  props.Name,
        Class: checkboxClass,
    }

    extra := make(map[string]string)
    if props.Checked {
        extra["checked"] = "checked"
    }
    if props.Disabled {
        extra["disabled"] = "disabled"
    }
    if len(extra) > 0 {
        attrs.Extra = extra
    }

    input := core.Input(attrs)

    // Without label, return just the input
    if props.Label == "" {
        return input
    }

    // With label, wrap in label element
    wrapperClass := MergeClasses(
        "flex items-start gap-2",
        ConditionalClass(props.Disabled, "cursor-not-allowed"),
        props.Class,
    )

    labelContent := []core.Node{
        core.Span(core.Class("text-sm font-medium text-gray-700 dark:text-gray-300"),
            core.Text(props.Label),
        ),
    }

    if props.Description != "" {
        labelContent = append(labelContent,
            core.Span(core.Class("block text-sm text-gray-500 dark:text-gray-400"),
                core.Text(props.Description),
            ),
        )
    }

    return core.Label(core.Class(wrapperClass),
        input,
        core.Div(core.Attrs{}, labelContent...),
    )
}
```

## Open Questions

Things that couldn't be fully resolved:

1. **Checkbox/Radio OnChange Handler**
   - What we know: `core.Attrs.OnChange` returns string, not bool
   - What's unclear: How to handle boolean change events cleanly
   - Recommendation: For Phase 17, rely on form submission to get checked state, or add OnChange to core.Attrs that supports checkbox/radio. Document limitation.

2. **ID Generation Strategy**
   - What we know: Labels need `for` attribute matching input `id`
   - What's unclear: Whether to auto-generate IDs or require explicit passing
   - Recommendation: Accept optional ID prop, auto-generate from name if not provided, document pattern

3. **Form Wrapper Component**
   - What we know: Examples use `core.Div` for form containers, not `core.Form`
   - What's unclear: Whether a `Form` component with OnSubmit is needed
   - Recommendation: Create simple Form wrapper that uses core.Form with OnSubmit, adds CSRF handling if needed

## Sources

### Primary (HIGH confidence)
- `/Users/dougbarrett/projects/dbb1dev/goquery/ui/button.go` - Established Props pattern
- `/Users/dougbarrett/projects/dbb1dev/goquery/ui/utils.go` - MergeClasses, ConditionalClass utilities
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/node.go` - Attrs struct definition
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/elements.go` - Input, Select, Textarea, Label helpers
- `/Users/dougbarrett/projects/dbb1dev/goquery/core/dom_renderer.go` - OnChange behavior
- `/Users/dougbarrett/projects/dbb1dev/goquery/examples/minimal/admin/user_new.go` - formField pattern
- `/Users/dougbarrett/projects/dbb1dev/goquery/examples/minimal/admin/user_edit.go` - selectOption pattern
- `/Users/dougbarrett/projects/dbb1dev/goquery/.planning/phases/16-component-foundation/16-RESEARCH.md` - Phase 16 patterns

### Secondary (MEDIUM confidence)
- Old `components/input.go` - Reference for feature list (not patterns)
- Old `components/form.go` - Reference for validation approach

### Tertiary (LOW confidence)
- WCAG 2.1 Form Guidelines - Accessibility requirements (general knowledge)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Based on existing codebase and Phase 16 patterns
- Architecture: HIGH - Verified against examples/minimal and ui/ package
- Accessibility: MEDIUM - Based on general ARIA knowledge, verify with testing
- Validation: HIGH - Pattern verified in existing codebase
- Styling: HIGH - Tailwind patterns consistent with Phase 16

**Research date:** 2026-01-22
**Valid until:** 2026-03-22 (60 days - stable domain)
