//go:build js && wasm

package components

import (
	"encoding/json"
	"syscall/js"
)

// DataDisplay shows loading states, errors, and formatted JSON data
type DataDisplay struct {
	element js.Value
}

// NewDataDisplay creates a new DataDisplay component
func NewDataDisplay() *DataDisplay {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "mt-4 p-4 bg-gray-100 rounded min-h-[100px]")
	container.Set("innerHTML", "<p class=\"text-gray-500\">Click a button to fetch data...</p>")

	return &DataDisplay{element: container}
}

// Element returns the underlying DOM element
func (d *DataDisplay) Element() js.Value {
	return d.element
}

// ShowLoading displays a loading message
func (d *DataDisplay) ShowLoading(message string) {
	d.element.Set("innerHTML", `<p class="text-blue-500">`+message+`</p>`)
}

// ShowError displays an error message
func (d *DataDisplay) ShowError(message string) {
	d.element.Set("innerHTML", `<p class="text-red-500">`+message+`</p>`)
}

// ShowJSON displays formatted JSON data
func (d *DataDisplay) ShowJSON(data any) {
	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		d.ShowError("Error formatting JSON: " + err.Error())
		return
	}

	d.element.Set("innerHTML", `<pre class="text-sm overflow-auto">`+string(formatted)+`</pre>`)
}
