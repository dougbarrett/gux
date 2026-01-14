//go:build js && wasm

package components

import (
	"fmt"
	"syscall/js"
)

// Default heading classes by level
var headingClasses = map[int]string{
	1: "text-3xl font-bold mb-6 text-gray-900 dark:text-gray-100",
	2: "text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100",
	3: "text-base font-semibold mb-3 text-gray-900 dark:text-gray-100",
	4: "text-sm font-semibold mb-2 text-gray-900 dark:text-gray-100",
	5: "text-xs font-semibold mb-2 text-gray-900 dark:text-gray-100",
	6: "text-xs font-medium mb-1 text-gray-900 dark:text-gray-100",
}

// Heading creates a heading element (h1-h6) with appropriate styling.
// Level should be 1-6, defaults to 2 if out of range.
func Heading(level int, content string) js.Value {
	if level < 1 || level > 6 {
		level = 2
	}

	className := headingClasses[level]
	tag := fmt.Sprintf("h%d", level)

	el := El(tag, className)
	el.Set("textContent", content)
	return el
}

// HeadingWithClass creates a heading with custom classes.
func HeadingWithClass(level int, content string, className string) js.Value {
	if level < 1 || level > 6 {
		level = 2
	}

	tag := fmt.Sprintf("h%d", level)
	el := El(tag, className)
	el.Set("textContent", content)
	return el
}

// H1 creates an h1 heading
func H1(content string) js.Value { return Heading(1, content) }

// H2 creates an h2 heading
func H2(content string) js.Value { return Heading(2, content) }

// H3 creates an h3 heading
func H3(content string) js.Value { return Heading(3, content) }

// H4 creates an h4 heading
func H4(content string) js.Value { return Heading(4, content) }

// H5 creates an h5 heading
func H5(content string) js.Value { return Heading(5, content) }

// H6 creates an h6 heading
func H6(content string) js.Value { return Heading(6, content) }
