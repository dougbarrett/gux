//go:build js && wasm

package components

import (
	"regexp"
	"syscall/js"
)

// ValidationRule defines a validation check
type ValidationRule struct {
	Validate func(value string) bool
	Message  string
}

// Common validation rules
var (
	Required = ValidationRule{
		Validate: func(v string) bool { return v != "" },
		Message:  "This field is required",
	}

	Email = ValidationRule{
		Validate: func(v string) bool {
			if v == "" {
				return true // Use Required for empty check
			}
			re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
			return re.MatchString(v)
		},
		Message: "Please enter a valid email address",
	}
)

// MinLength creates a minimum length validation rule
func MinLength(n int) ValidationRule {
	return ValidationRule{
		Validate: func(v string) bool { return len(v) >= n },
		Message:  "Must be at least " + itoa(n) + " characters",
	}
}

// MaxLength creates a maximum length validation rule
func MaxLength(n int) ValidationRule {
	return ValidationRule{
		Validate: func(v string) bool { return len(v) <= n },
		Message:  "Must be at most " + itoa(n) + " characters",
	}
}

// Pattern creates a regex pattern validation rule
func Pattern(pattern, message string) ValidationRule {
	re := regexp.MustCompile(pattern)
	return ValidationRule{
		Validate: func(v string) bool {
			if v == "" {
				return true
			}
			return re.MatchString(v)
		},
		Message: message,
	}
}

// Simple int to string for validation messages
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// FormField defines a field in a form
type FormField struct {
	Name        string
	Label       string
	Type        InputType
	Placeholder string
	Value       string
	Rules       []ValidationRule
}

// FormProps configures a Form component
type FormProps struct {
	Fields      []FormField
	SubmitLabel string
	OnSubmit    func(values map[string]string)
	OnCancel    func()
	CancelLabel string
}

// Form is a validated form component
type Form struct {
	element js.Value
	fields  map[string]*formFieldInstance
}

type formFieldInstance struct {
	input      *Input
	errorEl    js.Value
	rules      []ValidationRule
	errorShown bool
}

// NewForm creates a new Form component
func NewForm(props FormProps) *Form {
	document := js.Global().Get("document")

	form := document.Call("createElement", "form")
	form.Set("className", "space-y-4")

	f := &Form{
		element: form,
		fields:  make(map[string]*formFieldInstance),
	}

	// Create fields
	for _, field := range props.Fields {
		fieldContainer := document.Call("createElement", "div")

		input := NewInput(InputProps{
			Type:        field.Type,
			Label:       field.Label,
			Placeholder: field.Placeholder,
			Value:       field.Value,
		})

		fieldContainer.Call("appendChild", input.Element())

		// Error message element
		errorEl := document.Call("createElement", "p")
		errorEl.Set("className", "text-red-500 text-sm mt-1 hidden")
		fieldContainer.Call("appendChild", errorEl)

		form.Call("appendChild", fieldContainer)

		f.fields[field.Name] = &formFieldInstance{
			input:   input,
			errorEl: errorEl,
			rules:   field.Rules,
		}
	}

	// Button container
	buttonContainer := document.Call("createElement", "div")
	buttonContainer.Set("className", "flex gap-2 pt-4")

	// Submit button
	submitLabel := props.SubmitLabel
	if submitLabel == "" {
		submitLabel = "Submit"
	}

	submitBtn := Button(ButtonProps{
		Text: submitLabel,
		OnClick: func() {
			if f.Validate() && props.OnSubmit != nil {
				props.OnSubmit(f.Values())
			}
		},
	})
	buttonContainer.Call("appendChild", submitBtn)

	// Cancel button
	if props.OnCancel != nil {
		cancelLabel := props.CancelLabel
		if cancelLabel == "" {
			cancelLabel = "Cancel"
		}
		cancelBtn := Button(ButtonProps{
			Text:      cancelLabel,
			ClassName: "px-4 py-2 bg-gray-200 text-gray-800 rounded-md hover:bg-gray-300 cursor-pointer transition-colors",
			OnClick:   props.OnCancel,
		})
		buttonContainer.Call("appendChild", cancelBtn)
	}

	form.Call("appendChild", buttonContainer)

	// Prevent default form submission
	form.Call("addEventListener", "submit", js.FuncOf(func(this js.Value, args []js.Value) any {
		args[0].Call("preventDefault")
		return nil
	}))

	return f
}

// Element returns the form DOM element
func (f *Form) Element() js.Value {
	return f.element
}

// Values returns all field values
func (f *Form) Values() map[string]string {
	values := make(map[string]string)
	for name, field := range f.fields {
		values[name] = field.input.Value()
	}
	return values
}

// Value returns a single field value
func (f *Form) Value(name string) string {
	if field, ok := f.fields[name]; ok {
		return field.input.Value()
	}
	return ""
}

// SetValue sets a field value
func (f *Form) SetValue(name, value string) {
	if field, ok := f.fields[name]; ok {
		field.input.SetValue(value)
	}
}

// Validate runs validation on all fields
func (f *Form) Validate() bool {
	valid := true
	for _, field := range f.fields {
		if !f.validateField(field) {
			valid = false
		}
	}
	return valid
}

func (f *Form) validateField(field *formFieldInstance) bool {
	value := field.input.Value()

	for _, rule := range field.rules {
		if !rule.Validate(value) {
			field.input.SetError(rule.Message)
			field.errorEl.Set("textContent", rule.Message)
			field.errorEl.Get("classList").Call("remove", "hidden")
			field.errorShown = true
			return false
		}
	}

	// Clear any previous error
	if field.errorShown {
		field.input.ClearError()
		field.errorEl.Get("classList").Call("add", "hidden")
		field.errorShown = false
	}

	return true
}

// Reset clears all fields and errors
func (f *Form) Reset() {
	for _, field := range f.fields {
		field.input.SetValue("")
		field.input.ClearError()
		field.errorEl.Get("classList").Call("add", "hidden")
		field.errorShown = false
	}
}

// SetFieldError manually sets an error on a field (e.g., from server validation)
func (f *Form) SetFieldError(name, message string) {
	if field, ok := f.fields[name]; ok {
		field.input.SetError(message)
		field.errorEl.Set("textContent", message)
		field.errorEl.Get("classList").Call("remove", "hidden")
		field.errorShown = true
	}
}
