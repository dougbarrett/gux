//go:build js && wasm

package components

import (
	"encoding/json"
	"strings"
	"syscall/js"
)

// triggerDownload creates a file download in the browser
func triggerDownload(data []byte, filename, mimeType string) {
	document := js.Global().Get("document")
	URL := js.Global().Get("URL")

	// Create Uint8Array from Go []byte
	uint8Array := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(uint8Array, data)

	// Create Blob with proper MIME type
	blobOptions := js.Global().Get("Object").New()
	blobOptions.Set("type", mimeType)
	blob := js.Global().Get("Blob").New(
		js.Global().Get("Array").New(uint8Array),
		blobOptions,
	)

	// Create object URL
	objectURL := URL.Call("createObjectURL", blob)

	// Create anchor element for download
	anchor := document.Call("createElement", "a")
	anchor.Set("href", objectURL)
	anchor.Set("download", filename)
	anchor.Get("style").Set("display", "none")

	// Append to body, click, then cleanup
	document.Get("body").Call("appendChild", anchor)
	anchor.Call("click")
	document.Get("body").Call("removeChild", anchor)

	// Revoke object URL to free memory
	URL.Call("revokeObjectURL", objectURL)
}

// escapeCSVField escapes a field for CSV output
// Handles quotes, commas, and newlines
func escapeCSVField(value string) string {
	needsQuotes := false
	if strings.ContainsAny(value, ",\"\n\r") {
		needsQuotes = true
	}

	if needsQuotes {
		// Escape quotes by doubling them
		escaped := strings.ReplaceAll(value, "\"", "\"\"")
		return "\"" + escaped + "\""
	}
	return value
}

// ExportCSV exports data to a CSV file and triggers browser download
// columns determines the order and which fields to include
func ExportCSV(data []map[string]any, columns []string, filename string) {
	if len(data) == 0 {
		return
	}

	var builder strings.Builder

	// Write header row
	for i, col := range columns {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(escapeCSVField(col))
	}
	builder.WriteString("\n")

	// Write data rows
	for _, row := range data {
		for i, col := range columns {
			if i > 0 {
				builder.WriteString(",")
			}
			value := row[col]
			if value != nil {
				builder.WriteString(escapeCSVField(toString(value)))
			}
		}
		builder.WriteString("\n")
	}

	// Ensure filename has .csv extension
	if !strings.HasSuffix(filename, ".csv") {
		filename += ".csv"
	}

	triggerDownload([]byte(builder.String()), filename, "text/csv;charset=utf-8")
}

// ExportJSON exports data to a JSON file and triggers browser download
func ExportJSON(data []map[string]any, filename string) {
	if len(data) == 0 {
		return
	}

	// Marshal with indentation for readability
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		// Fallback to non-indented if indentation fails
		jsonBytes, err = json.Marshal(data)
		if err != nil {
			return
		}
	}

	// Ensure filename has .json extension
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	triggerDownload(jsonBytes, filename, "application/json")
}
