//go:build js && wasm

package components

import "syscall/js"

// ConfirmDialogVariant defines the visual style of the dialog
type ConfirmDialogVariant string

const (
	ConfirmVariantDefault ConfirmDialogVariant = "default"
	ConfirmVariantDanger  ConfirmDialogVariant = "danger"
	ConfirmVariantWarning ConfirmDialogVariant = "warning"
)

// ConfirmDialogProps configures a ConfirmDialog
type ConfirmDialogProps struct {
	Title       string               // Dialog title
	Message     string               // Confirmation message
	ConfirmText string               // Confirm button text (default: "Confirm")
	CancelText  string               // Cancel button text (default: "Cancel")
	Variant     ConfirmDialogVariant // Visual style (affects confirm button)
	OnConfirm   func()               // Called when confirmed
	OnCancel    func()               // Called when cancelled (optional)
}

// ConfirmDialog wraps Modal for confirmation workflows
type ConfirmDialog struct {
	modal     *Modal
	props     ConfirmDialogProps
	messageID string // ARIA: unique ID for aria-describedby
}

// NewConfirmDialog creates a new confirmation dialog
func NewConfirmDialog(props ConfirmDialogProps) *ConfirmDialog {
	// Set defaults
	if props.ConfirmText == "" {
		props.ConfirmText = "Confirm"
	}
	if props.CancelText == "" {
		props.CancelText = "Cancel"
	}
	if props.Variant == "" {
		props.Variant = ConfirmVariantDefault
	}

	// Generate unique ID for aria-describedby
	document := js.Global().Get("document")
	messageID := "confirm-desc-" + js.Global().Get("crypto").Call("randomUUID").String()

	cd := &ConfirmDialog{
		props:     props,
		messageID: messageID,
	}

	// Determine confirm button variant
	var confirmVariant ButtonVariant
	switch props.Variant {
	case ConfirmVariantDanger:
		confirmVariant = ButtonDanger
	case ConfirmVariantWarning:
		confirmVariant = ButtonWarning
	default:
		confirmVariant = ButtonPrimary
	}

	// Build footer with Cancel and Confirm buttons
	footer := Div("flex justify-end gap-2",
		SecondaryButton(props.CancelText, func() {
			cd.modal.Close()
			if props.OnCancel != nil {
				props.OnCancel()
			}
		}),
		Button(ButtonProps{
			Text:    props.ConfirmText,
			Variant: confirmVariant,
			OnClick: func() {
				cd.modal.Close()
				if props.OnConfirm != nil {
					props.OnConfirm()
				}
			},
		}),
	)

	// Create message element with ID for aria-describedby
	messageEl := document.Call("createElement", "p")
	messageEl.Set("id", messageID)
	messageEl.Set("className", "text-gray-700 dark:text-gray-300")
	messageEl.Set("textContent", props.Message)

	// Create the modal
	cd.modal = NewModal(ModalProps{
		Title:      props.Title,
		Content:    messageEl,
		Footer:     footer,
		CloseOnEsc: true,
	})

	// Override to alertdialog role and add aria-describedby
	modalEl := cd.modal.ModalElement()
	modalEl.Call("setAttribute", "role", "alertdialog")
	modalEl.Call("setAttribute", "aria-describedby", messageID)

	return cd
}

// Element returns the dialog DOM element
func (cd *ConfirmDialog) Element() js.Value {
	return cd.modal.Element()
}

// Open shows the confirmation dialog
func (cd *ConfirmDialog) Open() {
	cd.modal.Open()
}

// Close hides the confirmation dialog
func (cd *ConfirmDialog) Close() {
	cd.modal.Close()
}

// IsOpen returns whether the dialog is currently open
func (cd *ConfirmDialog) IsOpen() bool {
	return cd.modal.IsOpen()
}

// Confirm is a shortcut for simple confirmation dialogs
func Confirm(title, message string, onConfirm func()) *ConfirmDialog {
	return NewConfirmDialog(ConfirmDialogProps{
		Title:     title,
		Message:   message,
		OnConfirm: onConfirm,
	})
}

// ConfirmDanger is a shortcut for dangerous action confirmation
func ConfirmDanger(title, message string, onConfirm func()) *ConfirmDialog {
	return NewConfirmDialog(ConfirmDialogProps{
		Title:       title,
		Message:     message,
		ConfirmText: "Delete",
		Variant:     ConfirmVariantDanger,
		OnConfirm:   onConfirm,
	})
}
