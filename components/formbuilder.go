//go:build js && wasm

package components

import (
	"fmt"
	"syscall/js"
)

// BuilderFieldType defines the type of form field for builder
type BuilderFieldType string

const (
	BuilderFieldText     BuilderFieldType = "text"
	BuilderFieldEmail    BuilderFieldType = "email"
	BuilderFieldPassword BuilderFieldType = "password"
	BuilderFieldNumber   BuilderFieldType = "number"
	BuilderFieldTextarea BuilderFieldType = "textarea"
	BuilderFieldSelect   BuilderFieldType = "select"
	BuilderFieldCheckbox BuilderFieldType = "checkbox"
	BuilderFieldRadio    BuilderFieldType = "radio"
	BuilderFieldDate     BuilderFieldType = "date"
	BuilderFieldTime     BuilderFieldType = "time"
	BuilderFieldFile     BuilderFieldType = "file"
	BuilderFieldHidden   BuilderFieldType = "hidden"
	BuilderFieldCustom   BuilderFieldType = "custom"
)

// BuilderField defines a single form field configuration
type BuilderField struct {
	Name         string
	Type         BuilderFieldType
	Label        string
	Placeholder  string
	DefaultValue any
	Options      []SelectOption // For select, radio
	Rules        []ValidationRule
	ClassName    string
	Disabled     bool
	ReadOnly     bool
	Multiple     bool   // For file/select
	Accept       string // For file input
	Rows         int    // For textarea
	Min          string // For number/date
	Max          string // For number/date
	Step         string // For number
	CustomRender func(field BuilderField, value any, onChange func(any)) js.Value
}

// BuilderSection groups fields together
type BuilderSection struct {
	Title       string
	Description string
	Fields      []BuilderField
	ClassName   string
}

// FormBuilderProps configures the FormBuilder
type FormBuilderProps struct {
	Sections       []BuilderSection
	Fields         []BuilderField // Use directly if no sections
	SubmitText     string
	CancelText     string
	ShowCancel     bool
	OnSubmit       func(values map[string]any) error
	OnCancel       func()
	OnChange       func(name string, value any)
	ClassName      string
	FieldClassName string
	Inline         bool // Render fields inline
}

// FormBuilder creates dynamic forms from configuration
type FormBuilder struct {
	props    FormBuilderProps
	values   map[string]any
	errors   map[string]string
	touched  map[string]bool
	form     js.Value
	onChange []func(string, any)
}

// NewFormBuilder creates a new form builder instance
func NewFormBuilder(props FormBuilderProps) *FormBuilder {
	if props.SubmitText == "" {
		props.SubmitText = "Submit"
	}

	fb := &FormBuilder{
		props:   props,
		values:  make(map[string]any),
		errors:  make(map[string]string),
		touched: make(map[string]bool),
	}

	// Initialize default values
	allFields := fb.getAllFields()
	for _, field := range allFields {
		if field.DefaultValue != nil {
			fb.values[field.Name] = field.DefaultValue
		} else {
			switch field.Type {
			case BuilderFieldCheckbox:
				fb.values[field.Name] = false
			default:
				fb.values[field.Name] = ""
			}
		}
	}

	fb.form = fb.render()
	return fb
}

func (fb *FormBuilder) getAllFields() []BuilderField {
	if len(fb.props.Sections) > 0 {
		var fields []BuilderField
		for _, section := range fb.props.Sections {
			fields = append(fields, section.Fields...)
		}
		return fields
	}
	return fb.props.Fields
}

func (fb *FormBuilder) render() js.Value {
	document := js.Global().Get("document")

	form := document.Call("createElement", "form")
	form.Set("className", "space-y-6 "+fb.props.ClassName)

	// Prevent default form submission
	form.Call("addEventListener", "submit", js.FuncOf(func(this js.Value, args []js.Value) any {
		args[0].Call("preventDefault")
		fb.handleSubmit()
		return nil
	}))

	if len(fb.props.Sections) > 0 {
		for _, section := range fb.props.Sections {
			sectionEl := fb.renderSection(section)
			form.Call("appendChild", sectionEl)
		}
	} else {
		fieldsContainer := document.Call("createElement", "div")
		className := "space-y-4"
		if fb.props.Inline {
			className = "flex flex-wrap gap-4"
		}
		fieldsContainer.Set("className", className)

		for _, field := range fb.props.Fields {
			fieldEl := fb.renderField(field)
			fieldsContainer.Call("appendChild", fieldEl)
		}
		form.Call("appendChild", fieldsContainer)
	}

	// Buttons
	buttonContainer := document.Call("createElement", "div")
	buttonContainer.Set("className", "flex gap-3 pt-4")

	submitBtn := document.Call("createElement", "button")
	submitBtn.Set("type", "submit")
	submitBtn.Set("className", "px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors")
	submitBtn.Set("textContent", fb.props.SubmitText)
	buttonContainer.Call("appendChild", submitBtn)

	if fb.props.ShowCancel {
		cancelBtn := document.Call("createElement", "button")
		cancelBtn.Set("type", "button")
		cancelBtn.Set("className", "px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition-colors")
		cancelBtn.Set("textContent", fb.props.CancelText)
		if fb.props.CancelText == "" {
			cancelBtn.Set("textContent", "Cancel")
		}
		cancelBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			if fb.props.OnCancel != nil {
				fb.props.OnCancel()
			}
			return nil
		}))
		buttonContainer.Call("appendChild", cancelBtn)
	}

	form.Call("appendChild", buttonContainer)

	return form
}

func (fb *FormBuilder) renderSection(section BuilderSection) js.Value {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "space-y-4 "+section.ClassName)

	if section.Title != "" {
		title := document.Call("createElement", "h3")
		title.Set("className", "text-lg font-semibold text-gray-900")
		title.Set("textContent", section.Title)
		container.Call("appendChild", title)
	}

	if section.Description != "" {
		desc := document.Call("createElement", "p")
		desc.Set("className", "text-sm text-gray-500 mb-4")
		desc.Set("textContent", section.Description)
		container.Call("appendChild", desc)
	}

	fieldsContainer := document.Call("createElement", "div")
	className := "space-y-4"
	if fb.props.Inline {
		className = "flex flex-wrap gap-4"
	}
	fieldsContainer.Set("className", className)

	for _, field := range section.Fields {
		fieldEl := fb.renderField(field)
		fieldsContainer.Call("appendChild", fieldEl)
	}

	container.Call("appendChild", fieldsContainer)
	return container
}

func (fb *FormBuilder) renderField(field BuilderField) js.Value {
	document := js.Global().Get("document")

	// Custom render
	if field.Type == BuilderFieldCustom && field.CustomRender != nil {
		return field.CustomRender(field, fb.values[field.Name], func(val any) {
			fb.setValue(field.Name, val)
		})
	}

	container := document.Call("createElement", "div")
	className := fb.props.FieldClassName
	if className == "" {
		className = "space-y-1"
	}
	if field.ClassName != "" {
		className += " " + field.ClassName
	}
	container.Set("className", className)

	// Label (except for checkbox which has inline label)
	if field.Label != "" && field.Type != BuilderFieldHidden && field.Type != BuilderFieldCheckbox {
		label := document.Call("createElement", "label")
		label.Set("className", "block text-sm font-medium text-gray-700")
		label.Set("htmlFor", field.Name)
		label.Set("textContent", field.Label)

		// Add required indicator
		for _, rule := range field.Rules {
			if rule.Message == Required.Message || rule.Message == "This field is required" {
				span := document.Call("createElement", "span")
				span.Set("className", "text-red-500 ml-1")
				span.Set("textContent", "*")
				label.Call("appendChild", span)
				break
			}
		}

		container.Call("appendChild", label)
	}

	// Input element
	var input js.Value
	switch field.Type {
	case BuilderFieldTextarea:
		input = fb.renderTextarea(field)
	case BuilderFieldSelect:
		input = fb.renderSelect(field)
	case BuilderFieldCheckbox:
		input = fb.renderCheckbox(field)
	case BuilderFieldRadio:
		input = fb.renderRadioGroup(field)
	default:
		input = fb.renderInput(field)
	}

	container.Call("appendChild", input)

	// Error message
	if field.Type != BuilderFieldHidden {
		errorEl := document.Call("createElement", "p")
		errorEl.Set("className", "text-sm text-red-500 hidden")
		errorEl.Set("id", field.Name+"-error")
		errorEl.Call("setAttribute", "role", "alert")
		container.Call("appendChild", errorEl)
	}

	return container
}

func (fb *FormBuilder) renderInput(field BuilderField) js.Value {
	document := js.Global().Get("document")

	input := document.Call("createElement", "input")
	input.Set("type", string(field.Type))
	input.Set("name", field.Name)
	input.Set("id", field.Name)
	input.Set("className", "w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors")

	if field.Placeholder != "" {
		input.Set("placeholder", field.Placeholder)
	}
	if field.Disabled {
		input.Set("disabled", true)
	}
	if field.ReadOnly {
		input.Set("readOnly", true)
	}
	if field.Min != "" {
		input.Set("min", field.Min)
	}
	if field.Max != "" {
		input.Set("max", field.Max)
	}
	if field.Step != "" {
		input.Set("step", field.Step)
	}
	if field.Type == BuilderFieldFile {
		if field.Accept != "" {
			input.Set("accept", field.Accept)
		}
		if field.Multiple {
			input.Set("multiple", true)
		}
	}

	// Set initial value
	if val, ok := fb.values[field.Name]; ok {
		input.Set("value", fmt.Sprintf("%v", val))
	}

	// Change handler
	fieldName := field.Name
	fieldType := field.Type
	input.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		var value any
		if fieldType == BuilderFieldNumber {
			value = input.Get("valueAsNumber").Float()
		} else if fieldType == BuilderFieldFile {
			value = input.Get("files")
		} else {
			value = input.Get("value").String()
		}
		fb.setValue(fieldName, value)
		return nil
	}))

	// Blur handler for validation
	input.Call("addEventListener", "blur", js.FuncOf(func(this js.Value, args []js.Value) any {
		fb.touched[fieldName] = true
		fb.validateField(field)
		return nil
	}))

	return input
}

func (fb *FormBuilder) renderTextarea(field BuilderField) js.Value {
	document := js.Global().Get("document")

	textarea := document.Call("createElement", "textarea")
	textarea.Set("name", field.Name)
	textarea.Set("id", field.Name)
	textarea.Set("className", "w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors resize-y")

	rows := field.Rows
	if rows == 0 {
		rows = 4
	}
	textarea.Set("rows", rows)

	if field.Placeholder != "" {
		textarea.Set("placeholder", field.Placeholder)
	}
	if field.Disabled {
		textarea.Set("disabled", true)
	}
	if field.ReadOnly {
		textarea.Set("readOnly", true)
	}

	if val, ok := fb.values[field.Name]; ok {
		textarea.Set("value", fmt.Sprintf("%v", val))
	}

	fieldName := field.Name
	textarea.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		fb.setValue(fieldName, textarea.Get("value").String())
		return nil
	}))

	textarea.Call("addEventListener", "blur", js.FuncOf(func(this js.Value, args []js.Value) any {
		fb.touched[fieldName] = true
		fb.validateField(field)
		return nil
	}))

	return textarea
}

func (fb *FormBuilder) renderSelect(field BuilderField) js.Value {
	document := js.Global().Get("document")

	selectEl := document.Call("createElement", "select")
	selectEl.Set("name", field.Name)
	selectEl.Set("id", field.Name)
	selectEl.Set("className", "w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors bg-white")

	if field.Disabled {
		selectEl.Set("disabled", true)
	}
	if field.Multiple {
		selectEl.Set("multiple", true)
	}

	// Add placeholder option
	if field.Placeholder != "" {
		placeholder := document.Call("createElement", "option")
		placeholder.Set("value", "")
		placeholder.Set("textContent", field.Placeholder)
		placeholder.Set("disabled", true)
		placeholder.Set("selected", true)
		selectEl.Call("appendChild", placeholder)
	}

	currentVal := fmt.Sprintf("%v", fb.values[field.Name])
	for _, opt := range field.Options {
		option := document.Call("createElement", "option")
		option.Set("value", opt.Value)
		option.Set("textContent", opt.Label)
		if opt.Disabled {
			option.Set("disabled", true)
		}
		if opt.Value == currentVal {
			option.Set("selected", true)
		}
		selectEl.Call("appendChild", option)
	}

	fieldName := field.Name
	selectEl.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) any {
		fb.setValue(fieldName, selectEl.Get("value").String())
		return nil
	}))

	selectEl.Call("addEventListener", "blur", js.FuncOf(func(this js.Value, args []js.Value) any {
		fb.touched[fieldName] = true
		fb.validateField(field)
		return nil
	}))

	return selectEl
}

func (fb *FormBuilder) renderCheckbox(field BuilderField) js.Value {
	document := js.Global().Get("document")

	wrapper := document.Call("createElement", "div")
	wrapper.Set("className", "flex items-center gap-2")

	input := document.Call("createElement", "input")
	input.Set("type", "checkbox")
	input.Set("name", field.Name)
	input.Set("id", field.Name)
	input.Set("className", "h-4 w-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500")

	if field.Disabled {
		input.Set("disabled", true)
	}

	if val, ok := fb.values[field.Name].(bool); ok && val {
		input.Set("checked", true)
	}

	fieldName := field.Name
	input.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) any {
		fb.setValue(fieldName, input.Get("checked").Bool())
		return nil
	}))

	wrapper.Call("appendChild", input)

	if field.Label != "" {
		label := document.Call("createElement", "label")
		label.Set("className", "text-sm text-gray-700 cursor-pointer")
		label.Set("htmlFor", field.Name)
		label.Set("textContent", field.Label)
		wrapper.Call("appendChild", label)
	}

	return wrapper
}

func (fb *FormBuilder) renderRadioGroup(field BuilderField) js.Value {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "space-y-2")

	currentVal := fmt.Sprintf("%v", fb.values[field.Name])

	for i, opt := range field.Options {
		wrapper := document.Call("createElement", "div")
		wrapper.Set("className", "flex items-center gap-2")

		input := document.Call("createElement", "input")
		input.Set("type", "radio")
		input.Set("name", field.Name)
		input.Set("id", fmt.Sprintf("%s-%d", field.Name, i))
		input.Set("value", opt.Value)
		input.Set("className", "h-4 w-4 text-blue-600 border-gray-300 focus:ring-blue-500")

		if field.Disabled || opt.Disabled {
			input.Set("disabled", true)
		}
		if opt.Value == currentVal {
			input.Set("checked", true)
		}

		optValue := opt.Value
		fieldName := field.Name
		input.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) any {
			fb.setValue(fieldName, optValue)
			return nil
		}))

		wrapper.Call("appendChild", input)

		label := document.Call("createElement", "label")
		label.Set("className", "text-sm text-gray-700 cursor-pointer")
		label.Set("htmlFor", fmt.Sprintf("%s-%d", field.Name, i))
		label.Set("textContent", opt.Label)
		wrapper.Call("appendChild", label)

		container.Call("appendChild", wrapper)
	}

	return container
}

func (fb *FormBuilder) setValue(name string, value any) {
	fb.values[name] = value

	if fb.props.OnChange != nil {
		fb.props.OnChange(name, value)
	}

	for _, fn := range fb.onChange {
		fn(name, value)
	}

	// Validate if touched
	if fb.touched[name] {
		for _, field := range fb.getAllFields() {
			if field.Name == name {
				fb.validateField(field)
				break
			}
		}
	}
}

func (fb *FormBuilder) validateField(field BuilderField) bool {
	value := fb.values[field.Name]
	strVal := fmt.Sprintf("%v", value)

	for _, rule := range field.Rules {
		// Use the existing ValidationRule which has a Validate function
		if !rule.Validate(strVal) {
			fb.errors[field.Name] = rule.Message
			fb.showError(field.Name, rule.Message)
			return false
		}
	}

	delete(fb.errors, field.Name)
	fb.hideError(field.Name)
	return true
}

func (fb *FormBuilder) showError(name, message string) {
	document := js.Global().Get("document")
	errorID := name + "-error"
	errorEl := document.Call("getElementById", errorID)
	if !errorEl.IsNull() && !errorEl.IsUndefined() {
		errorEl.Set("textContent", message)
		errorEl.Get("classList").Call("remove", "hidden")
	}

	// Add error styling and ARIA attributes to input
	input := document.Call("getElementById", name)
	if !input.IsNull() && !input.IsUndefined() {
		input.Get("classList").Call("add", "border-red-500")
		input.Get("classList").Call("remove", "border-gray-300")
		input.Call("setAttribute", "aria-invalid", "true")
		input.Call("setAttribute", "aria-describedby", errorID)
	}
}

func (fb *FormBuilder) hideError(name string) {
	document := js.Global().Get("document")
	errorEl := document.Call("getElementById", name+"-error")
	if !errorEl.IsNull() && !errorEl.IsUndefined() {
		errorEl.Get("classList").Call("add", "hidden")
	}

	// Remove error styling and ARIA attributes from input
	input := document.Call("getElementById", name)
	if !input.IsNull() && !input.IsUndefined() {
		input.Get("classList").Call("remove", "border-red-500")
		input.Get("classList").Call("add", "border-gray-300")
		input.Call("removeAttribute", "aria-invalid")
		input.Call("removeAttribute", "aria-describedby")
	}
}

func (fb *FormBuilder) handleSubmit() {
	// Mark all as touched
	for _, field := range fb.getAllFields() {
		fb.touched[field.Name] = true
	}

	// Validate all
	valid := true
	for _, field := range fb.getAllFields() {
		if !fb.validateField(field) {
			valid = false
		}
	}

	if !valid {
		return
	}

	if fb.props.OnSubmit != nil {
		err := fb.props.OnSubmit(fb.values)
		if err != nil {
			// Show general error
			console := js.Global().Get("console")
			console.Call("error", "Form submission error:", err.Error())
		}
	}
}

// Element returns the form element
func (fb *FormBuilder) Element() js.Value {
	return fb.form
}

// GetValues returns all current form values
func (fb *FormBuilder) GetValues() map[string]any {
	result := make(map[string]any)
	for k, v := range fb.values {
		result[k] = v
	}
	return result
}

// GetValue returns a specific field value
func (fb *FormBuilder) GetValue(name string) any {
	return fb.values[name]
}

// SetFormValue sets a field value programmatically
func (fb *FormBuilder) SetFormValue(name string, value any) {
	fb.setValue(name, value)

	// Update DOM
	document := js.Global().Get("document")
	input := document.Call("getElementById", name)
	if !input.IsNull() && !input.IsUndefined() {
		tagName := input.Get("tagName").String()
		inputType := input.Get("type").String()

		if inputType == "checkbox" {
			if boolVal, ok := value.(bool); ok {
				input.Set("checked", boolVal)
			}
		} else if tagName == "SELECT" || tagName == "INPUT" || tagName == "TEXTAREA" {
			input.Set("value", fmt.Sprintf("%v", value))
		}
	}
}

// Reset resets the form to initial values
func (fb *FormBuilder) Reset() {
	for _, field := range fb.getAllFields() {
		if field.DefaultValue != nil {
			fb.SetFormValue(field.Name, field.DefaultValue)
		} else {
			switch field.Type {
			case BuilderFieldCheckbox:
				fb.SetFormValue(field.Name, false)
			default:
				fb.SetFormValue(field.Name, "")
			}
		}
		fb.hideError(field.Name)
	}
	fb.touched = make(map[string]bool)
	fb.errors = make(map[string]string)
}

// ValidateForm validates all fields and returns whether the form is valid
func (fb *FormBuilder) ValidateForm() bool {
	valid := true
	for _, field := range fb.getAllFields() {
		fb.touched[field.Name] = true
		if !fb.validateField(field) {
			valid = false
		}
	}
	return valid
}

// GetErrors returns all current validation errors
func (fb *FormBuilder) GetErrors() map[string]string {
	result := make(map[string]string)
	for k, v := range fb.errors {
		result[k] = v
	}
	return result
}

// OnFormChange adds a change listener
func (fb *FormBuilder) OnFormChange(fn func(name string, value any)) {
	fb.onChange = append(fb.onChange, fn)
}

// BuildForm is a convenience function to create a form from field definitions
func BuildForm(props FormBuilderProps) js.Value {
	fb := NewFormBuilder(props)
	return fb.Element()
}
