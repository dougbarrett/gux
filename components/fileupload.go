//go:build js && wasm

package components

import (
	"fmt"
	"syscall/js"
)

// FileInfo represents information about an uploaded file
type FileInfo struct {
	Name     string
	Size     int64
	Type     string
	File     js.Value // The actual File object
	DataURL  string   // Base64 data URL (for images)
	Progress int      // Upload progress 0-100
}

// FileUploadProps configures a FileUpload component
type FileUploadProps struct {
	Label       string
	Accept      string // File types to accept (e.g., "image/*", ".pdf,.doc")
	Multiple    bool   // Allow multiple files
	MaxSize     int64  // Max file size in bytes (0 = no limit)
	MaxFiles    int    // Max number of files (0 = no limit)
	ShowPreview bool   // Show image previews
	OnSelect    func(files []FileInfo)
	OnError     func(err string)
}

// FileUpload creates a file upload component with drag & drop
type FileUpload struct {
	container js.Value
	dropzone  js.Value
	input     js.Value
	preview   js.Value
	files     []FileInfo
	props     FileUploadProps
}

// NewFileUpload creates a new FileUpload component
func NewFileUpload(props FileUploadProps) *FileUpload {
	document := js.Global().Get("document")

	f := &FileUpload{props: props}

	container := document.Call("createElement", "div")
	container.Set("className", "space-y-2")

	// Label
	if props.Label != "" {
		label := document.Call("createElement", "label")
		label.Set("className", "block text-sm font-medium text-gray-700")
		label.Set("textContent", props.Label)
		container.Call("appendChild", label)
	}

	// Dropzone
	dropzone := document.Call("createElement", "div")
	dropzone.Set("className", "border-2 border-dashed border-gray-300 rounded-lg p-6 text-center hover:border-blue-400 transition-colors cursor-pointer")

	// Icon
	icon := document.Call("createElement", "div")
	icon.Set("className", "text-4xl text-gray-400 mb-2")
	icon.Set("textContent", "📁")
	dropzone.Call("appendChild", icon)

	// Text
	text := document.Call("createElement", "div")
	text.Set("className", "text-sm text-gray-600")
	text.Set("innerHTML", "<span class='text-blue-500 font-medium'>Click to upload</span> or drag and drop")
	dropzone.Call("appendChild", text)

	// Hint
	hint := document.Call("createElement", "div")
	hint.Set("className", "text-xs text-gray-400 mt-1")
	hintText := ""
	if props.Accept != "" {
		hintText = props.Accept
	}
	if props.MaxSize > 0 {
		if hintText != "" {
			hintText += " • "
		}
		hintText += fmt.Sprintf("Max %s", formatFileSize(props.MaxSize))
	}
	if hintText != "" {
		hint.Set("textContent", hintText)
		dropzone.Call("appendChild", hint)
	}

	f.dropzone = dropzone
	container.Call("appendChild", dropzone)

	// Hidden file input
	input := document.Call("createElement", "input")
	input.Set("type", "file")
	input.Set("className", "hidden")
	if props.Accept != "" {
		input.Set("accept", props.Accept)
	}
	if props.Multiple {
		input.Set("multiple", true)
	}
	f.input = input
	container.Call("appendChild", input)

	// Preview container
	preview := document.Call("createElement", "div")
	preview.Set("className", "grid grid-cols-2 md:grid-cols-4 gap-2 mt-2")
	f.preview = preview
	container.Call("appendChild", preview)

	f.container = container

	// Event handlers
	dropzone.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		input.Call("click")
		return nil
	}))

	input.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) any {
		files := input.Get("files")
		f.handleFiles(files)
		return nil
	}))

	// Drag and drop
	dropzone.Call("addEventListener", "dragover", js.FuncOf(func(this js.Value, args []js.Value) any {
		args[0].Call("preventDefault")
		dropzone.Set("className", "border-2 border-dashed border-blue-500 rounded-lg p-6 text-center bg-blue-50 cursor-pointer")
		return nil
	}))

	dropzone.Call("addEventListener", "dragleave", js.FuncOf(func(this js.Value, args []js.Value) any {
		args[0].Call("preventDefault")
		dropzone.Set("className", "border-2 border-dashed border-gray-300 rounded-lg p-6 text-center hover:border-blue-400 transition-colors cursor-pointer")
		return nil
	}))

	dropzone.Call("addEventListener", "drop", js.FuncOf(func(this js.Value, args []js.Value) any {
		args[0].Call("preventDefault")
		dropzone.Set("className", "border-2 border-dashed border-gray-300 rounded-lg p-6 text-center hover:border-blue-400 transition-colors cursor-pointer")
		files := args[0].Get("dataTransfer").Get("files")
		f.handleFiles(files)
		return nil
	}))

	return f
}

func (f *FileUpload) handleFiles(fileList js.Value) {
	count := fileList.Length()

	if f.props.MaxFiles > 0 && count > f.props.MaxFiles {
		if f.props.OnError != nil {
			f.props.OnError(fmt.Sprintf("Maximum %d files allowed", f.props.MaxFiles))
		}
		return
	}

	f.files = nil
	f.preview.Set("innerHTML", "")

	for i := 0; i < count; i++ {
		file := fileList.Index(i)
		name := file.Get("name").String()
		size := int64(file.Get("size").Int())
		fileType := file.Get("type").String()

		// Check file size
		if f.props.MaxSize > 0 && size > f.props.MaxSize {
			if f.props.OnError != nil {
				f.props.OnError(fmt.Sprintf("File %s exceeds maximum size of %s", name, formatFileSize(f.props.MaxSize)))
			}
			continue
		}

		info := FileInfo{
			Name: name,
			Size: size,
			Type: fileType,
			File: file,
		}

		// Generate preview for images
		if f.props.ShowPreview && isImageType(fileType) {
			f.createImagePreview(file, &info)
		} else {
			f.createFilePreview(info)
		}

		f.files = append(f.files, info)
	}

	if f.props.OnSelect != nil && len(f.files) > 0 {
		f.props.OnSelect(f.files)
	}
}

func (f *FileUpload) createImagePreview(file js.Value, info *FileInfo) {
	document := js.Global().Get("document")

	reader := js.Global().Get("FileReader").New()

	reader.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) any {
		dataURL := reader.Get("result").String()
		info.DataURL = dataURL

		card := document.Call("createElement", "div")
		card.Set("className", "relative group")

		img := document.Call("createElement", "img")
		img.Set("src", dataURL)
		img.Set("className", "w-full h-24 object-cover rounded")
		card.Call("appendChild", img)

		// Overlay with file name
		overlay := document.Call("createElement", "div")
		overlay.Set("className", "absolute inset-0 bg-black bg-opacity-50 opacity-0 group-hover:opacity-100 transition-opacity rounded flex items-end")

		nameEl := document.Call("createElement", "div")
		nameEl.Set("className", "text-white text-xs p-1 truncate w-full")
		nameEl.Set("textContent", info.Name)
		overlay.Call("appendChild", nameEl)

		card.Call("appendChild", overlay)

		// Remove button
		removeBtn := f.createRemoveButton(info.Name)
		card.Call("appendChild", removeBtn)

		f.preview.Call("appendChild", card)
		return nil
	}))

	reader.Call("readAsDataURL", file)
}

func (f *FileUpload) createFilePreview(info FileInfo) {
	document := js.Global().Get("document")

	card := document.Call("createElement", "div")
	card.Set("className", "relative bg-gray-100 rounded p-3")

	// File icon
	icon := document.Call("createElement", "div")
	icon.Set("className", "text-2xl text-gray-500 text-center")
	icon.Set("textContent", getFileIcon(info.Type))
	card.Call("appendChild", icon)

	// File name
	name := document.Call("createElement", "div")
	name.Set("className", "text-xs text-gray-700 truncate mt-1")
	name.Set("textContent", info.Name)
	card.Call("appendChild", name)

	// File size
	size := document.Call("createElement", "div")
	size.Set("className", "text-xs text-gray-400")
	size.Set("textContent", formatFileSize(info.Size))
	card.Call("appendChild", size)

	// Remove button
	removeBtn := f.createRemoveButton(info.Name)
	card.Call("appendChild", removeBtn)

	f.preview.Call("appendChild", card)
}

func (f *FileUpload) createRemoveButton(fileName string) js.Value {
	document := js.Global().Get("document")

	btn := document.Call("createElement", "button")
	btn.Set("className", "absolute top-1 right-1 w-5 h-5 bg-red-500 text-white rounded-full text-xs hover:bg-red-600")
	btn.Set("textContent", "×")

	btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		args[0].Call("stopPropagation")
		f.removeFile(fileName)
		return nil
	}))

	return btn
}

func (f *FileUpload) removeFile(name string) {
	// Remove from files slice
	newFiles := make([]FileInfo, 0, len(f.files)-1)
	for _, file := range f.files {
		if file.Name != name {
			newFiles = append(newFiles, file)
		}
	}
	f.files = newFiles

	// Re-render previews
	f.preview.Set("innerHTML", "")
	for _, file := range f.files {
		if f.props.ShowPreview && isImageType(file.Type) && file.DataURL != "" {
			// Re-create image preview with stored dataURL
			f.createImagePreviewFromURL(file)
		} else {
			f.createFilePreview(file)
		}
	}

	if f.props.OnSelect != nil {
		f.props.OnSelect(f.files)
	}
}

func (f *FileUpload) createImagePreviewFromURL(info FileInfo) {
	document := js.Global().Get("document")

	card := document.Call("createElement", "div")
	card.Set("className", "relative group")

	img := document.Call("createElement", "img")
	img.Set("src", info.DataURL)
	img.Set("className", "w-full h-24 object-cover rounded")
	card.Call("appendChild", img)

	overlay := document.Call("createElement", "div")
	overlay.Set("className", "absolute inset-0 bg-black bg-opacity-50 opacity-0 group-hover:opacity-100 transition-opacity rounded flex items-end")

	nameEl := document.Call("createElement", "div")
	nameEl.Set("className", "text-white text-xs p-1 truncate w-full")
	nameEl.Set("textContent", info.Name)
	overlay.Call("appendChild", nameEl)

	card.Call("appendChild", overlay)

	removeBtn := f.createRemoveButton(info.Name)
	card.Call("appendChild", removeBtn)

	f.preview.Call("appendChild", card)
}

// Element returns the container DOM element
func (f *FileUpload) Element() js.Value {
	return f.container
}

// Files returns the selected files
func (f *FileUpload) Files() []FileInfo {
	return f.files
}

// Clear clears all selected files
func (f *FileUpload) Clear() {
	f.files = nil
	f.preview.Set("innerHTML", "")
	f.input.Set("value", "")
}

// Helper functions
func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func isImageType(mimeType string) bool {
	return len(mimeType) >= 6 && mimeType[:6] == "image/"
}

func getFileIcon(mimeType string) string {
	switch {
	case isImageType(mimeType):
		return "🖼️"
	case mimeType == "application/pdf":
		return "📄"
	case mimeType == "application/zip" || mimeType == "application/x-zip-compressed":
		return "📦"
	case len(mimeType) >= 5 && mimeType[:5] == "video":
		return "🎬"
	case len(mimeType) >= 5 && mimeType[:5] == "audio":
		return "🎵"
	default:
		return "📎"
	}
}

// ImageUpload creates a file upload for images with preview
func ImageUpload(label string, onSelect func([]FileInfo)) *FileUpload {
	return NewFileUpload(FileUploadProps{
		Label:       label,
		Accept:      "image/*",
		Multiple:    true,
		ShowPreview: true,
		OnSelect:    onSelect,
	})
}

// SingleFileUpload creates a single file upload
func SingleFileUpload(label, accept string, onSelect func([]FileInfo)) *FileUpload {
	return NewFileUpload(FileUploadProps{
		Label:    label,
		Accept:   accept,
		Multiple: false,
		OnSelect: onSelect,
	})
}

// DocumentUpload creates a document upload (PDF, Word, etc.)
func DocumentUpload(label string, onSelect func([]FileInfo)) *FileUpload {
	return NewFileUpload(FileUploadProps{
		Label:    label,
		Accept:   ".pdf,.doc,.docx,.xls,.xlsx,.txt",
		Multiple: true,
		OnSelect: onSelect,
	})
}
