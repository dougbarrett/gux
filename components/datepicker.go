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
	calendarID  string    // unique ID for aria-controls
	displayed   time.Time // currently displayed month
	selected    time.Time
	isOpen      bool
	props       DatePickerProps
	focusedDay  int        // currently focused day (1-31)
	keyHandler  js.Func    // keyboard navigation handler
	dayButtons  []js.Value // day button references for focus management
}

// NewDatePicker creates a new DatePicker component
func NewDatePicker(props DatePickerProps) *DatePicker {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "relative mb-4")

	// Generate unique IDs for ARIA associations
	inputID := "datepicker-input-" + js.Global().Get("crypto").Call("randomUUID").String()
	calendarID := "datepicker-calendar-" + js.Global().Get("crypto").Call("randomUUID").String()

	dp := &DatePicker{
		container:  container,
		calendarID: calendarID,
		displayed:  time.Now(),
		selected:   props.Value,
		props:      props,
	}

	if !props.Value.IsZero() {
		dp.displayed = props.Value
	}

	// Label
	if props.Label != "" {
		label := document.Call("createElement", "label")
		label.Set("className", "block text-sm font-medium text-gray-700 mb-1")
		label.Set("textContent", props.Label)
		label.Call("setAttribute", "for", inputID)
		container.Call("appendChild", label)
	}

	// Input field
	inputWrapper := document.Call("createElement", "div")
	inputWrapper.Set("className", "relative")

	input := document.Call("createElement", "input")
	input.Set("type", "text")
	input.Set("id", inputID)
	input.Set("readOnly", true)
	input.Set("className", "w-full px-3 py-2 pr-10 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 cursor-pointer")

	placeholder := props.Placeholder
	if placeholder == "" {
		placeholder = "Select date"
	}
	input.Set("placeholder", placeholder)

	// ARIA combobox pattern
	input.Call("setAttribute", "role", "combobox")
	input.Call("setAttribute", "aria-haspopup", "dialog")
	input.Call("setAttribute", "aria-expanded", "false")
	input.Call("setAttribute", "aria-controls", calendarID)
	if props.Label == "" {
		// Provide aria-label if no visible label
		input.Call("setAttribute", "aria-label", placeholder)
	}

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

	// Calendar dropdown with dialog role
	calendar := document.Call("createElement", "div")
	calendar.Set("id", calendarID)
	calendar.Set("className", "absolute z-50 mt-1 bg-white border border-gray-200 rounded-lg shadow-lg p-4 hidden")
	calendar.Call("setAttribute", "role", "dialog")
	calendar.Call("setAttribute", "aria-modal", "false")
	calendar.Call("setAttribute", "aria-label", "Choose date")
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
	prevBtn.Call("setAttribute", "aria-label", "Previous month")
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
	monthYear.Call("setAttribute", "aria-live", "polite")
	monthYear.Call("setAttribute", "aria-atomic", "true")

	nextBtn := document.Call("createElement", "button")
	nextBtn.Set("type", "button")
	nextBtn.Set("className", "p-1 hover:bg-gray-100 rounded cursor-pointer")
	nextBtn.Call("setAttribute", "aria-label", "Next month")
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

	// Days grid with ARIA grid pattern
	daysGrid := document.Call("createElement", "table")
	daysGrid.Set("className", "w-full")
	daysGrid.Call("setAttribute", "role", "grid")
	daysGrid.Call("setAttribute", "aria-label", fmt.Sprintf("%s %d", monthNames[dp.displayed.Month()-1], dp.displayed.Year()))

	// Day names header row
	thead := document.Call("createElement", "thead")
	dayNamesRow := document.Call("createElement", "tr")
	dayNamesRow.Call("setAttribute", "role", "row")
	for _, day := range dayNames {
		th := document.Call("createElement", "th")
		th.Set("className", "text-center text-xs text-gray-500 font-medium py-1 w-8")
		th.Set("textContent", day)
		th.Call("setAttribute", "role", "columnheader")
		th.Call("setAttribute", "abbr", day)
		dayNamesRow.Call("appendChild", th)
	}
	thead.Call("appendChild", dayNamesRow)
	daysGrid.Call("appendChild", thead)

	// Days body - organized into weeks (rows)
	tbody := document.Call("createElement", "tbody")

	// Get first day of month and number of days
	firstOfMonth := time.Date(dp.displayed.Year(), dp.displayed.Month(), 1, 0, 0, 0, 0, time.Local)
	startWeekday := int(firstOfMonth.Weekday())
	daysInMonth := time.Date(dp.displayed.Year(), dp.displayed.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()

	today := time.Now()

	// Build calendar as rows of 7 days each
	cellIndex := 0
	var currentRow js.Value

	// Helper to create a new row
	createRow := func() js.Value {
		row := document.Call("createElement", "tr")
		row.Call("setAttribute", "role", "row")
		return row
	}

	// Helper to create an empty cell
	createEmptyCell := func() js.Value {
		td := document.Call("createElement", "td")
		td.Set("className", "w-8 h-8")
		td.Call("setAttribute", "role", "gridcell")
		return td
	}

	currentRow = createRow()

	// Empty cells for days before first of month
	for i := 0; i < startWeekday; i++ {
		currentRow.Call("appendChild", createEmptyCell())
		cellIndex++
	}

	// Initialize dayButtons slice for focus management
	dp.dayButtons = make([]js.Value, daysInMonth)

	// Day cells
	for day := 1; day <= daysInMonth; day++ {
		// Start new row if needed
		if cellIndex%7 == 0 && cellIndex > 0 {
			tbody.Call("appendChild", currentRow)
			currentRow = createRow()
		}

		dayDate := time.Date(dp.displayed.Year(), dp.displayed.Month(), day, 0, 0, 0, 0, time.Local)

		// Create cell
		td := document.Call("createElement", "td")
		td.Call("setAttribute", "role", "gridcell")

		// Create button
		dayBtn := document.Call("createElement", "button")
		dayBtn.Set("type", "button")

		className := "w-8 h-8 rounded-full text-sm hover:bg-gray-100 cursor-pointer focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset"

		// Determine states
		isSelected := !dp.selected.IsZero() && dp.selected.Year() == dayDate.Year() && dp.selected.Month() == dayDate.Month() && dp.selected.Day() == dayDate.Day()
		isToday := today.Year() == dayDate.Year() && today.Month() == dayDate.Month() && today.Day() == dayDate.Day()
		disabled := false

		if !dp.props.MinDate.IsZero() && dayDate.Before(dp.props.MinDate) {
			disabled = true
		}
		if !dp.props.MaxDate.IsZero() && dayDate.After(dp.props.MaxDate) {
			disabled = true
		}

		// Apply visual styles
		if isSelected {
			className = "w-8 h-8 rounded-full text-sm bg-blue-500 text-white cursor-pointer focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset"
		} else if isToday {
			className = "w-8 h-8 rounded-full text-sm border border-blue-500 text-blue-500 cursor-pointer focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset"
		}
		if disabled {
			className = "w-8 h-8 rounded-full text-sm text-gray-300 cursor-not-allowed focus:outline-none"
		}

		dayBtn.Set("className", className)
		dayBtn.Set("textContent", fmt.Sprintf("%d", day))

		// Roving tabindex - focused day gets tabindex=0, others get -1
		if day == dp.focusedDay {
			dayBtn.Call("setAttribute", "tabindex", "0")
		} else {
			dayBtn.Call("setAttribute", "tabindex", "-1")
		}

		// Store button reference for focus management
		dp.dayButtons[day-1] = dayBtn

		// ARIA attributes for gridcell
		if isSelected {
			td.Call("setAttribute", "aria-selected", "true")
		}
		if disabled {
			td.Call("setAttribute", "aria-disabled", "true")
		}
		if isToday {
			td.Call("setAttribute", "aria-current", "date")
		}

		// Click handler
		if !disabled {
			capturedDay := day
			dayBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
				args[0].Call("stopPropagation")
				dp.selectDate(time.Date(dp.displayed.Year(), dp.displayed.Month(), capturedDay, 0, 0, 0, 0, time.Local))
				return nil
			}))
		}

		td.Call("appendChild", dayBtn)
		currentRow.Call("appendChild", td)
		cellIndex++
	}

	// Add remaining empty cells to complete last row
	for cellIndex%7 != 0 {
		currentRow.Call("appendChild", createEmptyCell())
		cellIndex++
	}

	// Append final row
	tbody.Call("appendChild", currentRow)
	daysGrid.Call("appendChild", tbody)
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
	dp.input.Call("setAttribute", "aria-expanded", "true")

	// Initialize focusedDay: prefer selected date, then today if in displayed month, else 1
	if !dp.selected.IsZero() && dp.selected.Year() == dp.displayed.Year() && dp.selected.Month() == dp.displayed.Month() {
		dp.focusedDay = dp.selected.Day()
	} else {
		today := time.Now()
		if today.Year() == dp.displayed.Year() && today.Month() == dp.displayed.Month() {
			dp.focusedDay = today.Day()
		} else {
			dp.focusedDay = 1
		}
	}

	dp.renderCalendar()

	// Focus the focused day button
	if dp.focusedDay > 0 && dp.focusedDay <= len(dp.dayButtons) {
		dp.dayButtons[dp.focusedDay-1].Call("focus")
	}

	// Set up keyboard handler for arrow navigation
	dp.keyHandler = js.FuncOf(func(this js.Value, args []js.Value) any {
		event := args[0]
		key := event.Get("key").String()

		switch key {
		case "ArrowRight":
			event.Call("preventDefault")
			dp.moveFocusBy(1)
		case "ArrowLeft":
			event.Call("preventDefault")
			dp.moveFocusBy(-1)
		case "ArrowDown":
			event.Call("preventDefault")
			dp.moveFocusBy(7)
		case "ArrowUp":
			event.Call("preventDefault")
			dp.moveFocusBy(-7)
		case "Enter", " ":
			event.Call("preventDefault")
			// Select the focused date if not disabled
			focusedDate := time.Date(dp.displayed.Year(), dp.displayed.Month(), dp.focusedDay, 0, 0, 0, 0, time.Local)
			if !dp.isDateDisabled(focusedDate) {
				dp.selectDate(focusedDate)
			}
		case "Escape":
			event.Call("preventDefault")
			dp.close()
			dp.input.Call("focus") // Return focus to input
		}
		return nil
	})
	dp.calendar.Call("addEventListener", "keydown", dp.keyHandler)
}

func (dp *DatePicker) close() {
	dp.isOpen = false
	dp.calendar.Get("classList").Call("add", "hidden")
	dp.input.Call("setAttribute", "aria-expanded", "false")

	// Clean up keyboard handler
	if dp.keyHandler.Truthy() {
		dp.calendar.Call("removeEventListener", "keydown", dp.keyHandler)
		dp.keyHandler.Release()
	}

	// Clear dayButtons slice
	dp.dayButtons = nil
}

// isDateDisabled checks if a date is outside the allowed range
func (dp *DatePicker) isDateDisabled(date time.Time) bool {
	if !dp.props.MinDate.IsZero() && date.Before(dp.props.MinDate) {
		return true
	}
	if !dp.props.MaxDate.IsZero() && date.After(dp.props.MaxDate) {
		return true
	}
	return false
}

// moveFocusBy moves the focused day by the specified number of days
// Handles month boundary transitions automatically
func (dp *DatePicker) moveFocusBy(days int) {
	// Calculate current focused date
	currentDate := time.Date(dp.displayed.Year(), dp.displayed.Month(), dp.focusedDay, 0, 0, 0, 0, time.Local)

	// Calculate new date
	newDate := currentDate.AddDate(0, 0, days)

	// Check if we've crossed a month boundary
	if newDate.Month() != dp.displayed.Month() || newDate.Year() != dp.displayed.Year() {
		// Update displayed month and re-render
		dp.displayed = time.Date(newDate.Year(), newDate.Month(), 1, 0, 0, 0, 0, time.Local)
		dp.focusedDay = newDate.Day()
		dp.renderCalendar()
	} else {
		// Same month, just update focus
		dp.focusedDay = newDate.Day()
	}

	// Focus the new day button
	if dp.focusedDay > 0 && dp.focusedDay <= len(dp.dayButtons) {
		// Update tabindex for roving tabindex pattern
		for i, btn := range dp.dayButtons {
			if i+1 == dp.focusedDay {
				btn.Call("setAttribute", "tabindex", "0")
				btn.Call("focus")
			} else {
				btn.Call("setAttribute", "tabindex", "-1")
			}
		}
	}
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
