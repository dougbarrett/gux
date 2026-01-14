//go:build js && wasm

package components

import "syscall/js"

// TextVariant defines text styling variants
type TextVariant string

const (
	TextDefault TextVariant = ""
	TextMuted   TextVariant = "muted"
	TextError   TextVariant = "error"
	TextSuccess TextVariant = "success"
)

var textVariantClasses = map[TextVariant]string{
	TextDefault: "text-gray-800",
	TextMuted:   "text-gray-600",
	TextError:   "text-red-500",
	TextSuccess: "text-green-500",
}

// Text creates a paragraph element with text content.
func Text(content string) js.Value {
	return TextWithVariant(content, TextMuted)
}

// TextWithVariant creates a paragraph with a specific style variant.
func TextWithVariant(content string, variant TextVariant) js.Value {
	className := textVariantClasses[variant]
	if className == "" {
		className = textVariantClasses[TextDefault]
	}

	el := El("p", className)
	el.Set("textContent", content)
	return el
}

// TextWithClass creates a paragraph with custom classes.
func TextWithClass(content string, className string) js.Value {
	el := El("p", className)
	el.Set("textContent", content)
	return el
}
