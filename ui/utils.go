// Package ui provides the Gux component library built on core's Node system.
package ui

import "strings"

// MergeClasses combines CSS class strings, filtering empty values.
// Custom classes should be passed last to allow overrides.
//
// Example:
//
//	MergeClasses("foo", "", "bar") // returns "foo bar"
//	MergeClasses("  foo  ", "bar") // returns "foo bar"
func MergeClasses(classes ...string) string {
	var result []string
	for _, c := range classes {
		c = strings.TrimSpace(c)
		if c != "" {
			result = append(result, c)
		}
	}
	return strings.Join(result, " ")
}

// ConditionalClass returns the class if condition is true, empty string otherwise.
// Useful for toggling classes based on component state.
//
// Example:
//
//	ConditionalClass(isActive, "bg-blue-500") // returns "bg-blue-500" if isActive
//	ConditionalClass(false, "hidden")         // returns ""
func ConditionalClass(condition bool, class string) string {
	if condition {
		return class
	}
	return ""
}
