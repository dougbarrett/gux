//go:build js && wasm

package components

import (
	"syscall/js"
)

// CodeLanguage defines supported programming languages
type CodeLanguage string

const (
	LangGo         CodeLanguage = "go"
	LangJavaScript CodeLanguage = "javascript"
	LangTypeScript CodeLanguage = "typescript"
	LangHTML       CodeLanguage = "html"
	LangCSS        CodeLanguage = "css"
	LangJSON       CodeLanguage = "json"
	LangMarkdown   CodeLanguage = "markdown"
	LangSQL        CodeLanguage = "sql"
	LangPython     CodeLanguage = "python"
	LangRust       CodeLanguage = "rust"
)

// CodeTheme defines the editor theme
type CodeTheme string

const (
	CodeThemeLight CodeTheme = "light"
	CodeThemeDark  CodeTheme = "dark"
	CodeThemeAuto  CodeTheme = "auto" // Follow Gux theme
)

// IsCodeMirrorAvailable checks if CodeMirror 6 is loaded
func IsCodeMirrorAvailable() bool {
	return !js.Global().Get("CM").IsUndefined()
}

// CodeEditorProps configures a CodeEditor component
type CodeEditorProps struct {
	Label       string
	Value       string
	Language    CodeLanguage
	Theme       CodeTheme // Default: auto (follows Gux theme)
	Height      string    // CSS height (default "300px")
	LineNumbers bool      // Default: true (included in basicSetup)
	ReadOnly    bool
	ClassName   string
	OnChange    func(value string)
}

// CodeEditor is a syntax-highlighted code editor using CodeMirror 6
type CodeEditor struct {
	container      js.Value
	editorEl       js.Value
	label          js.Value
	view           js.Value // EditorView instance
	editorID       string
	updateListener js.Func  // Stored for cleanup
	themeUnsub     func()   // Theme change unsubscribe
	props          CodeEditorProps
}

// NewCodeEditor creates a new code editor
// Note: CodeMirror 6 must be pre-loaded in HTML for this to work
func NewCodeEditor(props CodeEditorProps) *CodeEditor {
	if !IsCodeMirrorAvailable() {
		js.Global().Get("console").Call("error", "CodeMirror is not loaded. Add the CodeMirror module script to your HTML.")
		return nil
	}

	document := js.Global().Get("document")
	crypto := js.Global().Get("crypto")
	cm := js.Global().Get("CM")

	container := document.Call("createElement", "div")
	className := "mb-4"
	if props.ClassName != "" {
		className += " " + props.ClassName
	}
	container.Set("className", className)

	editorID := "code-" + crypto.Call("randomUUID").String()

	ce := &CodeEditor{
		container: container,
		editorID:  editorID,
		props:     props,
	}

	// Label
	if props.Label != "" {
		label := document.Call("createElement", "label")
		label.Set("className", "block text-sm font-medium text-secondary mb-1")
		label.Set("textContent", props.Label)
		container.Call("appendChild", label)
		ce.label = label
	}

	// Editor container
	editorEl := document.Call("createElement", "div")
	editorEl.Set("id", editorID)
	editorEl.Set("className", "border border-subtle rounded-md overflow-hidden")

	height := props.Height
	if height == "" {
		height = "300px"
	}
	editorEl.Get("style").Set("height", height)

	// ARIA accessibility
	editorEl.Call("setAttribute", "role", "textbox")
	editorEl.Call("setAttribute", "aria-multiline", "true")
	if props.Label != "" {
		editorEl.Call("setAttribute", "aria-label", props.Label)
	}

	container.Call("appendChild", editorEl)
	ce.editorEl = editorEl

	// Build extensions array using JavaScript
	extensions := ce.buildExtensions(props)

	// Create editor state with extensions
	stateConfig := js.Global().Get("Object").New()
	stateConfig.Set("doc", props.Value)
	if !extensions.IsUndefined() && extensions.Length() > 0 {
		stateConfig.Set("extensions", extensions)
	}

	state := cm.Get("EditorState").Call("create", stateConfig)

	// Create editor view
	viewConfig := js.Global().Get("Object").New()
	viewConfig.Set("state", state)
	viewConfig.Set("parent", editorEl)

	view := cm.Get("EditorView").New(viewConfig)
	ce.view = view

	return ce
}

// buildExtensions creates the CodeMirror extensions array
func (ce *CodeEditor) buildExtensions(props CodeEditorProps) js.Value {
	cm := js.Global().Get("CM")
	extensions := js.Global().Get("Array").New()

	// Add basicSetup (provides line numbers, folding, etc.)
	basicSetup := cm.Get("basicSetup")
	if !basicSetup.IsUndefined() {
		// basicSetup is an array, spread it into our extensions
		extensions = extensions.Call("concat", basicSetup)
	}

	// Add language extension
	langExt := ce.getLanguageExtension(props.Language)
	if !langExt.IsUndefined() {
		// Language functions return LanguageSupport, call it to get the extension
		langSupport := langExt.Invoke()
		if !langSupport.IsUndefined() {
			extensions.Call("push", langSupport)
		}
	}

	// Add theme based on props or auto-detect from Gux theme
	theme := props.Theme
	if theme == "" || theme == CodeThemeAuto {
		// Check if dark mode is active
		if isDarkMode() {
			theme = CodeThemeDark
		} else {
			theme = CodeThemeLight
		}
	}

	if theme == CodeThemeDark {
		oneDark := cm.Get("themes").Get("oneDark")
		if !oneDark.IsUndefined() {
			extensions.Call("push", oneDark)
		}
	}

	return extensions
}

// isDarkMode checks if the current Gux theme is dark
func isDarkMode() bool {
	document := js.Global().Get("document")
	html := document.Get("documentElement")
	classList := html.Get("classList")
	return classList.Call("contains", "dark").Bool()
}

func (ce *CodeEditor) getLanguageExtension(lang CodeLanguage) js.Value {
	cm := js.Global().Get("CM")
	langs := cm.Get("languages")

	switch lang {
	case LangGo:
		return langs.Get("go")
	case LangJavaScript, LangTypeScript:
		return langs.Get("javascript")
	case LangHTML:
		return langs.Get("html")
	case LangCSS:
		return langs.Get("css")
	case LangJSON:
		return langs.Get("json")
	case LangMarkdown:
		return langs.Get("markdown")
	case LangSQL:
		return langs.Get("sql")
	case LangPython:
		return langs.Get("python")
	case LangRust:
		return langs.Get("rust")
	default:
		return js.Undefined()
	}
}

// recreateWithTheme recreates the editor with the current theme
func (ce *CodeEditor) recreateWithTheme() {
	currentValue := ce.Value()
	ce.view.Call("destroy")

	cm := js.Global().Get("CM")

	// Rebuild extensions with potentially new theme
	extensions := ce.buildExtensions(ce.props)

	stateConfig := js.Global().Get("Object").New()
	stateConfig.Set("doc", currentValue)
	if !extensions.IsUndefined() && extensions.Length() > 0 {
		stateConfig.Set("extensions", extensions)
	}

	state := cm.Get("EditorState").Call("create", stateConfig)

	viewConfig := js.Global().Get("Object").New()
	viewConfig.Set("state", state)
	viewConfig.Set("parent", ce.editorEl)

	ce.view = cm.Get("EditorView").New(viewConfig)
}

// Element returns the container DOM element
func (ce *CodeEditor) Element() js.Value {
	return ce.container
}

// Value returns the current code content
func (ce *CodeEditor) Value() string {
	return ce.view.Get("state").Get("doc").Call("toString").String()
}

// SetValue replaces the entire content
func (ce *CodeEditor) SetValue(value string) {
	ce.view.Call("dispatch", js.ValueOf(map[string]any{
		"changes": map[string]any{
			"from":   0,
			"to":     ce.view.Get("state").Get("doc").Get("length").Int(),
			"insert": value,
		},
	}))
}

// Focus sets focus on the editor
func (ce *CodeEditor) Focus() {
	ce.view.Call("focus")
}

// Destroy cleans up resources
func (ce *CodeEditor) Destroy() {
	if ce.updateListener.Truthy() {
		ce.updateListener.Release()
	}
	if ce.themeUnsub != nil {
		ce.themeUnsub()
	}
	ce.view.Call("destroy")
}

// GoEditor creates a Go code editor
func GoEditor(label string) *CodeEditor {
	return NewCodeEditor(CodeEditorProps{
		Label:    label,
		Language: LangGo,
	})
}

// JavaScriptEditor creates a JavaScript code editor
func JavaScriptEditor(label string) *CodeEditor {
	return NewCodeEditor(CodeEditorProps{
		Label:    label,
		Language: LangJavaScript,
	})
}

// JSONEditor creates a JSON code editor
func JSONEditor(label string) *CodeEditor {
	return NewCodeEditor(CodeEditorProps{
		Label:    label,
		Language: LangJSON,
	})
}

// SQLEditor creates a SQL code editor
func SQLEditor(label string) *CodeEditor {
	return NewCodeEditor(CodeEditorProps{
		Label:    label,
		Language: LangSQL,
	})
}

// PythonEditor creates a Python code editor
func PythonEditor(label string) *CodeEditor {
	return NewCodeEditor(CodeEditorProps{
		Label:    label,
		Language: LangPython,
	})
}

// MarkdownEditor creates a Markdown code editor
func MarkdownEditor(label string) *CodeEditor {
	return NewCodeEditor(CodeEditorProps{
		Label:    label,
		Language: LangMarkdown,
	})
}

// HTMLEditor creates an HTML code editor
func HTMLEditor(label string) *CodeEditor {
	return NewCodeEditor(CodeEditorProps{
		Label:    label,
		Language: LangHTML,
	})
}

// CSSEditor creates a CSS code editor
func CSSEditor(label string) *CodeEditor {
	return NewCodeEditor(CodeEditorProps{
		Label:    label,
		Language: LangCSS,
	})
}
