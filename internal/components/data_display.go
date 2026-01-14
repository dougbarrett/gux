//go:build js && wasm

package components

import (
	"encoding/json"
	"syscall/js"
)

type DataDisplay struct {
	element js.Value
}

const (
	dataDisplayClass = "bg-gray-100 p-4 rounded overflow-x-auto font-mono whitespace-pre-wrap text-sm"
	errorClass       = "text-red-500 p-2"
	loadingClass     = "text-gray-500 p-2 animate-pulse"
)

func NewDataDisplay() *DataDisplay {
	document := js.Global().Get("document")
	element := document.Call("createElement", "div")
	return &DataDisplay{element: element}
}

func (d *DataDisplay) Element() js.Value {
	return d.element
}

func (d *DataDisplay) ShowLoading(message string) {
	if message == "" {
		message = "Loading..."
	}
	d.element.Set("innerHTML", "")

	document := js.Global().Get("document")
	p := document.Call("createElement", "p")
	p.Set("textContent", message)
	p.Set("className", loadingClass)

	d.element.Call("appendChild", p)
}

func (d *DataDisplay) ShowError(message string) {
	d.element.Set("innerHTML", "")

	document := js.Global().Get("document")
	p := document.Call("createElement", "p")
	p.Set("textContent", message)
	p.Set("className", errorClass)

	d.element.Call("appendChild", p)
}

func (d *DataDisplay) ShowJSON(data any) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		d.ShowError("Error formatting JSON")
		return err
	}

	d.element.Set("innerHTML", "")

	document := js.Global().Get("document")
	pre := document.Call("createElement", "pre")
	pre.Set("textContent", string(jsonBytes))
	pre.Set("className", dataDisplayClass)

	d.element.Call("appendChild", pre)
	return nil
}
