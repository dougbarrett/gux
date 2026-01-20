//go:build js && wasm

package components

import (
	"syscall/js"
)

// ToolbarPreset defines preset toolbar configurations
type ToolbarPreset string

const (
	ToolbarMinimal  ToolbarPreset = "minimal"
	ToolbarStandard ToolbarPreset = "standard"
	ToolbarFull     ToolbarPreset = "full"
)

var toolbarPresets = map[ToolbarPreset][]any{
	ToolbarMinimal: {
		[]any{"bold", "italic", "underline"},
		[]any{"link"},
	},
	ToolbarStandard: {
		[]any{"bold", "italic", "underline", "strike"},
		[]any{"blockquote", "code-block"},
		[]any{map[string]any{"list": "ordered"}, map[string]any{"list": "bullet"}},
		[]any{"link", "image"},
		[]any{"clean"},
	},
	ToolbarFull: {
		[]any{map[string]any{"header": []any{1, 2, 3, 4, 5, 6, false}}},
		[]any{"bold", "italic", "underline", "strike"},
		[]any{map[string]any{"color": []any{}}, map[string]any{"background": []any{}}},
		[]any{"blockquote", "code-block"},
		[]any{map[string]any{"list": "ordered"}, map[string]any{"list": "bullet"}},
		[]any{map[string]any{"indent": "-1"}, map[string]any{"indent": "+1"}},
		[]any{map[string]any{"align": []any{}}},
		[]any{"link", "image", "video"},
		[]any{"clean"},
	},
}

// IsQuillAvailable checks if Quill.js is loaded
func IsQuillAvailable() bool {
	return !js.Global().Get("Quill").IsUndefined()
}

var quillDarkModeStylesInjected = false

// injectQuillDarkModeStyles injects CSS for Quill dark mode support
func injectQuillDarkModeStyles() {
	if quillDarkModeStylesInjected {
		return
	}

	document := js.Global().Get("document")
	style := document.Call("createElement", "style")
	style.Set("id", "quill-dark-mode-styles")
	style.Set("textContent", `
		/* Quill Dark Mode Styles */
		.dark .ql-toolbar.ql-snow {
			background-color: rgb(31, 41, 55);
			border-color: rgb(55, 65, 81);
		}
		.dark .ql-container.ql-snow {
			background-color: rgb(17, 24, 39);
			border-color: rgb(55, 65, 81);
			color: rgb(229, 231, 235);
		}
		.dark .ql-editor {
			color: rgb(229, 231, 235);
		}
		.dark .ql-editor.ql-blank::before {
			color: rgb(156, 163, 175);
		}
		.dark .ql-snow .ql-stroke {
			stroke: rgb(209, 213, 219);
		}
		.dark .ql-snow .ql-fill {
			fill: rgb(209, 213, 219);
		}
		.dark .ql-snow .ql-picker {
			color: rgb(209, 213, 219);
		}
		.dark .ql-snow .ql-picker-options {
			background-color: rgb(31, 41, 55);
			border-color: rgb(55, 65, 81);
		}
		.dark .ql-snow .ql-picker-item:hover {
			color: rgb(96, 165, 250);
		}
		.dark .ql-snow button:hover .ql-stroke,
		.dark .ql-snow .ql-picker-label:hover .ql-stroke {
			stroke: rgb(96, 165, 250);
		}
		.dark .ql-snow button:hover .ql-fill,
		.dark .ql-snow .ql-picker-label:hover .ql-fill {
			fill: rgb(96, 165, 250);
		}
		.dark .ql-snow button.ql-active .ql-stroke {
			stroke: rgb(96, 165, 250);
		}
		.dark .ql-snow button.ql-active .ql-fill {
			fill: rgb(96, 165, 250);
		}
		.dark .ql-toolbar.ql-snow .ql-picker.ql-expanded .ql-picker-label {
			border-color: rgb(55, 65, 81);
		}
		.dark .ql-snow a {
			color: rgb(96, 165, 250);
		}
		.dark .ql-snow .ql-tooltip {
			background-color: rgb(31, 41, 55);
			border-color: rgb(55, 65, 81);
			color: rgb(229, 231, 235);
			box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3);
		}
		.dark .ql-snow .ql-tooltip input[type=text] {
			background-color: rgb(17, 24, 39);
			border-color: rgb(55, 65, 81);
			color: rgb(229, 231, 235);
		}
	`)
	document.Get("head").Call("appendChild", style)
	quillDarkModeStylesInjected = true
}

// RichTextEditorProps configures a RichTextEditor component
type RichTextEditorProps struct {
	Label         string
	Placeholder   string
	Value         string        // Initial HTML content
	Toolbar       ToolbarPreset // Preset toolbar configuration
	CustomToolbar [][]any       // Custom toolbar config (overrides Toolbar)
	Height        string        // CSS height (default "200px")
	ClassName     string
	Disabled      bool
	ReadOnly      bool
	OnChange      func(html string) // Called with HTML content
	OnTextChange  func(text string) // Called with plain text content
}

// RichTextEditor is a WYSIWYG rich text editor using Quill.js
type RichTextEditor struct {
	container      js.Value
	editorEl       js.Value
	label          js.Value
	quill          js.Value // Quill instance
	editorID       string
	textChangeFunc js.Func // Stored for cleanup
	props          RichTextEditorProps
}

// NewRichTextEditor creates a new rich text editor
// Note: Quill.js must be pre-loaded in HTML for this to work
func NewRichTextEditor(props RichTextEditorProps) *RichTextEditor {
	if !IsQuillAvailable() {
		js.Global().Get("console").Call("error", "Quill.js is not loaded. Add <script src=\"https://cdn.quilljs.com/1.3.7/quill.min.js\"></script> to your HTML.")
		return nil
	}

	// Inject dark mode styles on first use
	injectQuillDarkModeStyles()

	document := js.Global().Get("document")
	crypto := js.Global().Get("crypto")

	container := document.Call("createElement", "div")
	className := "mb-4"
	if props.ClassName != "" {
		className += " " + props.ClassName
	}
	container.Set("className", className)

	editorID := "rte-" + crypto.Call("randomUUID").String()

	rte := &RichTextEditor{
		container: container,
		editorID:  editorID,
		props:     props,
	}

	// Label
	if props.Label != "" {
		label := document.Call("createElement", "label")
		label.Set("className", "block text-sm font-medium text-secondary mb-1")
		label.Set("textContent", props.Label)
		label.Set("htmlFor", editorID)
		container.Call("appendChild", label)
		rte.label = label
	}

	// Editor container
	editorEl := document.Call("createElement", "div")
	editorEl.Set("id", editorID)
	height := props.Height
	if height == "" {
		height = "200px"
	}
	editorEl.Get("style").Set("minHeight", height)

	container.Call("appendChild", editorEl)
	rte.editorEl = editorEl

	// Determine toolbar config
	var toolbar any
	if props.CustomToolbar != nil {
		toolbar = props.CustomToolbar
	} else {
		preset := props.Toolbar
		if preset == "" {
			preset = ToolbarStandard
		}
		toolbar = toolbarPresets[preset]
	}

	// Create Quill options
	quillOpts := map[string]any{
		"theme": "snow",
		"modules": map[string]any{
			"toolbar": toolbar,
		},
		"placeholder": props.Placeholder,
		"readOnly":    props.ReadOnly || props.Disabled,
	}

	// Create Quill instance (pass element directly, not selector, since element isn't in DOM yet)
	quill := js.Global().Get("Quill").New(editorEl, js.ValueOf(quillOpts))
	rte.quill = quill

	// Set initial value
	if props.Value != "" {
		quill.Get("root").Set("innerHTML", props.Value)
	}

	// ARIA accessibility on the contenteditable element
	root := quill.Get("root")
	root.Call("setAttribute", "role", "textbox")
	root.Call("setAttribute", "aria-multiline", "true")
	if props.Label != "" {
		root.Call("setAttribute", "aria-label", props.Label)
	}

	// Set up change handlers
	if props.OnChange != nil || props.OnTextChange != nil {
		rte.textChangeFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
			if props.OnChange != nil {
				html := quill.Get("root").Get("innerHTML").String()
				props.OnChange(html)
			}
			if props.OnTextChange != nil {
				text := quill.Call("getText").String()
				props.OnTextChange(text)
			}
			return nil
		})
		quill.Call("on", "text-change", rte.textChangeFunc)
	}

	return rte
}

// Element returns the container DOM element
func (rte *RichTextEditor) Element() js.Value {
	return rte.container
}

// Value returns the current HTML content
func (rte *RichTextEditor) Value() string {
	return rte.quill.Get("root").Get("innerHTML").String()
}

// SetValue sets the HTML content
func (rte *RichTextEditor) SetValue(html string) {
	rte.quill.Get("root").Set("innerHTML", html)
}

// GetHTML returns the HTML content (alias for Value)
func (rte *RichTextEditor) GetHTML() string {
	return rte.Value()
}

// GetText returns the plain text content
func (rte *RichTextEditor) GetText() string {
	return rte.quill.Call("getText").String()
}

// Focus sets focus on the editor
func (rte *RichTextEditor) Focus() {
	rte.quill.Call("focus")
}

// Enable enables the editor
func (rte *RichTextEditor) Enable() {
	rte.quill.Call("enable", true)
}

// Disable disables the editor
func (rte *RichTextEditor) Disable() {
	rte.quill.Call("enable", false)
}

// Destroy cleans up resources
func (rte *RichTextEditor) Destroy() {
	if rte.textChangeFunc.Truthy() {
		rte.quill.Call("off", "text-change", rte.textChangeFunc)
		rte.textChangeFunc.Release()
	}
}

// SimpleRichTextEditor creates a basic rich text editor with standard toolbar
func SimpleRichTextEditor(label, placeholder string) *RichTextEditor {
	return NewRichTextEditor(RichTextEditorProps{
		Label:       label,
		Placeholder: placeholder,
		Toolbar:     ToolbarStandard,
	})
}

// MinimalRichTextEditor creates a minimal rich text editor
func MinimalRichTextEditor(placeholder string) *RichTextEditor {
	return NewRichTextEditor(RichTextEditorProps{
		Placeholder: placeholder,
		Toolbar:     ToolbarMinimal,
	})
}

// FullRichTextEditor creates a full-featured rich text editor
func FullRichTextEditor(label, placeholder string) *RichTextEditor {
	return NewRichTextEditor(RichTextEditorProps{
		Label:       label,
		Placeholder: placeholder,
		Toolbar:     ToolbarFull,
	})
}
