//go:build js && wasm

package components

import (
	"fmt"
	"syscall/js"
	"time"
)

var monthNames = []string{
	"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

var dayNames = []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

// DatePickerProps configures a DatePicker
type DatePickerProps struct {
	Label       string
	Value       time.Time
	Placeholder string
	MinDate     time.Time
	MaxDate     time.Time
	OnChange    func(time.Time)
}

// DatePicker is a date selection component
type DatePicker struct {
	container   js.Value
	input       js.Value
	calendar    js.Value
	displayed   time.Time // currently displayed month
	selected    time.Time
	isOpen      bool
	props       DatePickerProps
}

// NewDatePicker creates a new DatePicker component
func NewDatePicker(props DatePickerProps) *DatePicker {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "relative mb-4")

	dp := &DatePicker{
		container: container,
		displayed: time.Now(),
		selected:  props.Value,
		props:     props,
	}

	if !props.Value.IsZero() {
		dp.displayed = props.Value
	}

	// Label
	if props.Label != "" {
		label := document.Call("createElement", "label")
		label.Set("className", "block text-sm font-medium text-gray-700 mb-1")
		label.Set("textContent", props.Label)
		container.Call("appendChild", label)
	}

	// Input field
	inputWrapper := document.Call("createElement", "div")
	inputWrapper.Set("className", "relative")

	input := document.Call("createElement", "input")
	input.Set("type", "text")
	input.Set("readOnly", true)
	input.Set("className", "w-full px-3 py-2 pr-10 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 cursor-pointer")

	placeholder := props.Placeholder
	if placeholder == "" {
		placeholder = "Select date"
	}
	input.Set("placeholder", placeholder)

	if !props.Value.IsZero() {
		input.Set("value", props.Value.Format("Jan 2, 2006"))
	}

	// Calendar icon
	icon := document.Call("createElement", "div")
	icon.Set("className", "absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none text-gray-400")
	icon.Set("innerHTML", `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path></svg>`)

	inputWrapper.Call("appendChild", input)
	inputWrapper.Call("appendChild", icon)
	container.Call("appendChild", inputWrapper)

	dp.input = input

	// Calendar dropdown
	calendar := document.Call("createElement", "div")
	calendar.Set("className", "absolute z-50 mt-1 bg-white border border-gray-200 rounded-lg shadow-lg p-4 hidden")
	container.Call("appendChild", calendar)
	dp.calendar = calendar

	dp.renderCalendar()

	// Toggle calendar on input click
	input.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		dp.toggle()
		return nil
	}))

	// Close on outside click
	js.Global().Get("document").Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		target := args[0].Get("target")
		if !container.Call("contains", target).Bool() {
			dp.close()
		}
		return nil
	}))

	return dp
}

func (dp *DatePicker) renderCalendar() {
	document := js.Global().Get("document")
	dp.calendar.Set("innerHTML", "")

	// Header with month/year and navigation
	header := document.Call("createElement", "div")
	header.Set("className", "flex items-center justify-between mb-4")

	prevBtn := document.Call("createElement", "button")
	prevBtn.Set("type", "button")
	prevBtn.Set("className", "p-1 hover:bg-gray-100 rounded cursor-pointer")
	prevBtn.Set("innerHTML", `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path></svg>`)
	prevBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		args[0].Call("stopPropagation")
		dp.displayed = dp.displayed.AddDate(0, -1, 0)
		dp.renderCalendar()
		return nil
	}))

	monthYear := document.Call("createElement", "span")
	monthYear.Set("className", "font-semibold text-gray-800")
	monthYear.Set("textContent", fmt.Sprintf("%s %d", monthNames[dp.displayed.Month()-1], dp.displayed.Year()))

	nextBtn := document.Call("createElement", "button")
	nextBtn.Set("type", "button")
	nextBtn.Set("className", "p-1 hover:bg-gray-100 rounded cursor-pointer")
	nextBtn.Set("innerHTML", `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path></svg>`)
	nextBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		args[0].Call("stopPropagation")
		dp.displayed = dp.displayed.AddDate(0, 1, 0)
		dp.renderCalendar()
		return nil
	}))

	header.Call("appendChild", prevBtn)
	header.Call("appendChild", monthYear)
	header.Call("appendChild", nextBtn)
	dp.calendar.Call("appendChild", header)

	// Day names
	dayNamesRow := document.Call("createElement", "div")
	dayNamesRow.Set("className", "grid grid-cols-7 gap-1 mb-2")
	for _, day := range dayNames {
		d := document.Call("createElement", "div")
		d.Set("className", "text-center text-xs text-gray-500 font-medium py-1")
		d.Set("textContent", day)
		dayNamesRow.Call("appendChild", d)
	}
	dp.calendar.Call("appendChild", dayNamesRow)

	// Days grid
	daysGrid := document.Call("createElement", "div")
	daysGrid.Set("className", "grid grid-cols-7 gap-1")

	// Get first day of month and number of days
	firstOfMonth := time.Date(dp.displayed.Year(), dp.displayed.Month(), 1, 0, 0, 0, 0, time.Local)
	startWeekday := int(firstOfMonth.Weekday())
	daysInMonth := time.Date(dp.displayed.Year(), dp.displayed.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()

	today := time.Now()

	// Empty cells for days before first of month
	for i := 0; i < startWeekday; i++ {
		empty := document.Call("createElement", "div")
		empty.Set("className", "w-8 h-8")
		daysGrid.Call("appendChild", empty)
	}

	// Day cells
	for day := 1; day <= daysInMonth; day++ {
		dayDate := time.Date(dp.displayed.Year(), dp.displayed.Month(), day, 0, 0, 0, 0, time.Local)

		dayBtn := document.Call("createElement", "button")
		dayBtn.Set("type", "button")

		className := "w-8 h-8 rounded-full text-sm hover:bg-gray-100 cursor-pointer"

		// Check if selected
		if !dp.selected.IsZero() && dp.selected.Year() == dayDate.Year() && dp.selected.Month() == dayDate.Month() && dp.selected.Day() == dayDate.Day() {
			className = "w-8 h-8 rounded-full text-sm bg-blue-500 text-white cursor-pointer"
		} else if today.Year() == dayDate.Year() && today.Month() == dayDate.Month() && today.Day() == dayDate.Day() {
			className = "w-8 h-8 rounded-full text-sm border border-blue-500 text-blue-500 cursor-pointer"
		}

		// Check if disabled by min/max
		disabled := false
		if !dp.props.MinDate.IsZero() && dayDate.Before(dp.props.MinDate) {
			disabled = true
		}
		if !dp.props.MaxDate.IsZero() && dayDate.After(dp.props.MaxDate) {
			disabled = true
		}

		if disabled {
			className = "w-8 h-8 rounded-full text-sm text-gray-300 cursor-not-allowed"
		}

		dayBtn.Set("className", className)
		dayBtn.Set("textContent", fmt.Sprintf("%d", day))

		if !disabled {
			capturedDay := day
			dayBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
				args[0].Call("stopPropagation")
				dp.selectDate(time.Date(dp.displayed.Year(), dp.displayed.Month(), capturedDay, 0, 0, 0, 0, time.Local))
				return nil
			}))
		}

		daysGrid.Call("appendChild", dayBtn)
	}

	dp.calendar.Call("appendChild", daysGrid)

	// Today button
	todayBtn := document.Call("createElement", "button")
	todayBtn.Set("type", "button")
	todayBtn.Set("className", "w-full mt-3 py-1 text-sm text-blue-600 hover:bg-blue-50 rounded cursor-pointer")
	todayBtn.Set("textContent", "Today")
	todayBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		args[0].Call("stopPropagation")
		dp.selectDate(time.Now())
		return nil
	}))
	dp.calendar.Call("appendChild", todayBtn)
}

func (dp *DatePicker) selectDate(date time.Time) {
	dp.selected = date
	dp.input.Set("value", date.Format("Jan 2, 2006"))
	dp.close()

	if dp.props.OnChange != nil {
		dp.props.OnChange(date)
	}
}

func (dp *DatePicker) toggle() {
	if dp.isOpen {
		dp.close()
	} else {
		dp.open()
	}
}

func (dp *DatePicker) open() {
	dp.isOpen = true
	dp.calendar.Get("classList").Call("remove", "hidden")
	dp.renderCalendar()
}

func (dp *DatePicker) close() {
	dp.isOpen = false
	dp.calendar.Get("classList").Call("add", "hidden")
}

// Element returns the DOM element
func (dp *DatePicker) Element() js.Value {
	return dp.container
}

// Value returns the selected date
func (dp *DatePicker) Value() time.Time {
	return dp.selected
}

// SetValue sets the selected date
func (dp *DatePicker) SetValue(date time.Time) {
	dp.selected = date
	dp.displayed = date
	dp.input.Set("value", date.Format("Jan 2, 2006"))
	dp.renderCalendar()
}

// Clear clears the selected date
func (dp *DatePicker) Clear() {
	dp.selected = time.Time{}
	dp.input.Set("value", "")
}
