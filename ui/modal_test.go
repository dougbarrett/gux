package ui

import (
	"strings"
	"testing"

	"github.com/dougbarrett/gux/core"
)

func TestModal_DefaultClosed(t *testing.T) {
	modal := Modal(ModalProps{
		Open: false,
		Children: []core.Node{
			ModalContent(ModalContentProps{
				Children: []core.Node{core.Text("Content")},
			}),
		},
	})
	html := renderHTML(modal)

	// Should have hidden class when closed
	if !strings.Contains(html, "hidden") {
		t.Errorf("expected hidden class when Open=false, got: %s", html)
	}
}

func TestModal_Open(t *testing.T) {
	modal := Modal(ModalProps{
		Open: true,
		Children: []core.Node{
			ModalContent(ModalContentProps{
				Children: []core.Node{core.Text("Content")},
			}),
		},
	})
	html := renderHTML(modal)

	// Should NOT have hidden class when open
	// The overlay class should not contain "hidden"
	if strings.Contains(html, `class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 hidden"`) {
		t.Errorf("expected no hidden class when Open=true, got: %s", html)
	}

	// Should have base overlay classes
	if !strings.Contains(html, "fixed inset-0") {
		t.Errorf("expected overlay classes, got: %s", html)
	}
}

func TestModal_AllSizes(t *testing.T) {
	tests := []struct {
		size          ModalSize
		expectedClass string
	}{
		{ModalSM, "max-w-sm"},
		{ModalMD, "max-w-md"},
		{ModalLG, "max-w-lg"},
		{ModalXL, "max-w-xl"},
		{ModalFull, "max-w-4xl"},
	}

	for _, tt := range tests {
		t.Run(string(tt.size), func(t *testing.T) {
			modal := Modal(ModalProps{
				Open: true,
				Size: tt.size,
				Children: []core.Node{
					core.Text("Content"),
				},
			})
			html := renderHTML(modal)

			if !strings.Contains(html, tt.expectedClass) {
				t.Errorf("size %s: expected class %q in: %s", tt.size, tt.expectedClass, html)
			}
		})
	}
}

func TestModal_DefaultSize(t *testing.T) {
	modal := Modal(ModalProps{
		Open:     true,
		Children: []core.Node{core.Text("Content")},
	})
	html := renderHTML(modal)

	// Default size should be MD
	if !strings.Contains(html, "max-w-md") {
		t.Errorf("expected default size max-w-md, got: %s", html)
	}
}

func TestModal_AriaAttributes(t *testing.T) {
	modal := Modal(ModalProps{
		Open:     true,
		Children: []core.Node{core.Text("Content")},
	})
	html := renderHTML(modal)

	// Should have role="dialog"
	if !strings.Contains(html, `role="dialog"`) {
		t.Errorf("expected role=dialog, got: %s", html)
	}

	// Should have aria-modal="true"
	if !strings.Contains(html, `aria-modal="true"`) {
		t.Errorf("expected aria-modal=true, got: %s", html)
	}
}

func TestModal_WithTitle(t *testing.T) {
	modal := Modal(ModalProps{
		Open:     true,
		Title:    "My Modal Title",
		Children: []core.Node{core.Text("Content")},
	})
	html := renderHTML(modal)

	// Should have H3 with id="modal-title"
	if !strings.Contains(html, `id="modal-title"`) {
		t.Errorf("expected id=modal-title, got: %s", html)
	}

	// Should have aria-labelledby
	if !strings.Contains(html, `aria-labelledby="modal-title"`) {
		t.Errorf("expected aria-labelledby=modal-title, got: %s", html)
	}

	// Should contain the title text
	if !strings.Contains(html, "My Modal Title") {
		t.Errorf("expected title text, got: %s", html)
	}

	// Should render H3 element
	if !strings.Contains(html, "<h3") {
		t.Errorf("expected h3 element, got: %s", html)
	}
}

func TestModal_WithoutTitle(t *testing.T) {
	modal := Modal(ModalProps{
		Open:     true,
		Children: []core.Node{core.Text("Content")},
	})
	html := renderHTML(modal)

	// Should NOT have aria-labelledby
	if strings.Contains(html, "aria-labelledby") {
		t.Errorf("expected no aria-labelledby without title, got: %s", html)
	}

	// Should NOT have H3 with id="modal-title"
	if strings.Contains(html, `id="modal-title"`) {
		t.Errorf("expected no id=modal-title without title, got: %s", html)
	}
}

func TestModal_WithCloseButton(t *testing.T) {
	closeClicked := false
	modal := Modal(ModalProps{
		Open:     true,
		OnClose:  func() { closeClicked = true },
		Children: []core.Node{core.Text("Content")},
	})
	html := renderHTML(modal)

	// Should have close button with aria-label
	if !strings.Contains(html, `aria-label="Close"`) {
		t.Errorf("expected close button with aria-label, got: %s", html)
	}

	// Should have header with close button (spacer + button)
	if !strings.Contains(html, "border-b") {
		t.Errorf("expected header border, got: %s", html)
	}

	// closeClicked should still be false (we're just rendering)
	if closeClicked {
		t.Error("OnClose should not be called during render")
	}
}

func TestModal_CustomClass(t *testing.T) {
	modal := Modal(ModalProps{
		Open:     true,
		Class:    "my-custom-class extra-class",
		Children: []core.Node{core.Text("Content")},
	})
	html := renderHTML(modal)

	// Should include custom classes
	if !strings.Contains(html, "my-custom-class") {
		t.Errorf("expected my-custom-class, got: %s", html)
	}
	if !strings.Contains(html, "extra-class") {
		t.Errorf("expected extra-class, got: %s", html)
	}

	// Should still have base classes
	if !strings.Contains(html, "rounded-lg") {
		t.Errorf("expected base classes, got: %s", html)
	}
}

func TestModal_Children(t *testing.T) {
	modal := Modal(ModalProps{
		Open: true,
		Children: []core.Node{
			core.Div(core.Class("test-child"), core.Text("Child 1")),
			core.Span(core.Class("test-span"), core.Text("Child 2")),
		},
	})
	html := renderHTML(modal)

	// Should contain children
	if !strings.Contains(html, "test-child") {
		t.Errorf("expected test-child class, got: %s", html)
	}
	if !strings.Contains(html, "Child 1") {
		t.Errorf("expected Child 1 text, got: %s", html)
	}
	if !strings.Contains(html, "test-span") {
		t.Errorf("expected test-span class, got: %s", html)
	}
	if !strings.Contains(html, "Child 2") {
		t.Errorf("expected Child 2 text, got: %s", html)
	}
}

func TestModalContent_Renders(t *testing.T) {
	content := ModalContent(ModalContentProps{
		Children: []core.Node{core.Text("Body content")},
	})
	html := renderHTML(content)

	// Should have padding classes
	if !strings.Contains(html, "px-6") {
		t.Errorf("expected px-6 class, got: %s", html)
	}
	if !strings.Contains(html, "py-4") {
		t.Errorf("expected py-4 class, got: %s", html)
	}

	// Should have overflow scroll
	if !strings.Contains(html, "overflow-y-auto") {
		t.Errorf("expected overflow-y-auto class, got: %s", html)
	}

	// Should have flex-1 for flexible height
	if !strings.Contains(html, "flex-1") {
		t.Errorf("expected flex-1 class, got: %s", html)
	}

	// Should contain content
	if !strings.Contains(html, "Body content") {
		t.Errorf("expected content text, got: %s", html)
	}
}

func TestModalContent_CustomClass(t *testing.T) {
	content := ModalContent(ModalContentProps{
		Class:    "custom-content-class",
		Children: []core.Node{core.Text("Content")},
	})
	html := renderHTML(content)

	// Should include custom class
	if !strings.Contains(html, "custom-content-class") {
		t.Errorf("expected custom-content-class, got: %s", html)
	}

	// Should still have base classes
	if !strings.Contains(html, "px-6") {
		t.Errorf("expected base px-6 class, got: %s", html)
	}
}

func TestModalFooter_Renders(t *testing.T) {
	footer := ModalFooter(ModalFooterProps{
		Children: []core.Node{core.Text("Footer content")},
	})
	html := renderHTML(footer)

	// Should have border-t
	if !strings.Contains(html, "border-t") {
		t.Errorf("expected border-t class, got: %s", html)
	}

	// Should have flex classes
	if !strings.Contains(html, "flex") {
		t.Errorf("expected flex class, got: %s", html)
	}
	if !strings.Contains(html, "justify-end") {
		t.Errorf("expected justify-end class, got: %s", html)
	}
	if !strings.Contains(html, "gap-2") {
		t.Errorf("expected gap-2 class, got: %s", html)
	}

	// Should have padding
	if !strings.Contains(html, "px-6") {
		t.Errorf("expected px-6 class, got: %s", html)
	}
	if !strings.Contains(html, "py-4") {
		t.Errorf("expected py-4 class, got: %s", html)
	}

	// Should contain content
	if !strings.Contains(html, "Footer content") {
		t.Errorf("expected footer content text, got: %s", html)
	}
}

func TestModalFooter_CustomClass(t *testing.T) {
	footer := ModalFooter(ModalFooterProps{
		Class:    "custom-footer-class",
		Children: []core.Node{core.Text("Footer")},
	})
	html := renderHTML(footer)

	// Should include custom class
	if !strings.Contains(html, "custom-footer-class") {
		t.Errorf("expected custom-footer-class, got: %s", html)
	}

	// Should still have base classes
	if !strings.Contains(html, "border-t") {
		t.Errorf("expected base border-t class, got: %s", html)
	}
}

func TestModal_CompleteExample(t *testing.T) {
	// Test a complete modal with all compound components
	modal := Modal(ModalProps{
		Open:  true,
		Title: "Confirm Delete",
		Size:  ModalLG,
		Class: "test-modal",
		Children: []core.Node{
			ModalContent(ModalContentProps{
				Class:    "test-content",
				Children: []core.Node{core.P(core.Attrs{}, core.Text("Are you sure you want to delete this item?"))},
			}),
			ModalFooter(ModalFooterProps{
				Class: "test-footer",
				Children: []core.Node{
					Button(ButtonProps{
						Variant:  ButtonSecondary,
						Children: []core.Node{core.Text("Cancel")},
					}),
					Button(ButtonProps{
						Variant:  ButtonDestructive,
						Children: []core.Node{core.Text("Delete")},
					}),
				},
			}),
		},
	})
	html := renderHTML(modal)

	// Check modal structure
	if !strings.Contains(html, `role="dialog"`) {
		t.Errorf("expected role=dialog, got: %s", html)
	}
	if !strings.Contains(html, "max-w-lg") {
		t.Errorf("expected large size, got: %s", html)
	}
	if !strings.Contains(html, "test-modal") {
		t.Errorf("expected test-modal class, got: %s", html)
	}
	if !strings.Contains(html, "Confirm Delete") {
		t.Errorf("expected title, got: %s", html)
	}

	// Check content
	if !strings.Contains(html, "test-content") {
		t.Errorf("expected test-content class, got: %s", html)
	}
	if !strings.Contains(html, "Are you sure") {
		t.Errorf("expected content text, got: %s", html)
	}

	// Check footer
	if !strings.Contains(html, "test-footer") {
		t.Errorf("expected test-footer class, got: %s", html)
	}
	if !strings.Contains(html, "Cancel") {
		t.Errorf("expected Cancel button, got: %s", html)
	}
	if !strings.Contains(html, "Delete") {
		t.Errorf("expected Delete button, got: %s", html)
	}
}
