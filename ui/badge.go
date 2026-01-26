package ui

import "github.com/dougbarrett/gux/core"

// BadgeVariant defines the visual style of a badge.
type BadgeVariant string

const (
	BadgeDefault BadgeVariant = "default"
	BadgeSuccess BadgeVariant = "success"
	BadgeWarning BadgeVariant = "warning"
	BadgeError   BadgeVariant = "error"
)

// BadgeProps configures a Badge component.
type BadgeProps struct {
	Text    string       // Badge text content
	Variant BadgeVariant // Visual style (default: default)
	Class   string       // Additional classes (override defaults)
}

// badgeVariantClasses maps variants to their Tailwind classes.
var badgeVariantClasses = map[BadgeVariant]string{
	BadgeDefault: "bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200",
	BadgeSuccess: "bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200",
	BadgeWarning: "bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200",
	BadgeError:   "bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200",
}

// badgeBaseClasses are the common classes applied to all badges.
const badgeBaseClasses = "inline-flex items-center px-3 py-1 rounded-full text-xs font-medium"

// Badge creates a badge component for status indicators.
//
// Example:
//
//	Badge(BadgeProps{
//	    Text:    "Active",
//	    Variant: BadgeSuccess,
//	})
func Badge(props BadgeProps) core.Node {
	// Apply defaults
	variant := props.Variant
	if variant == "" {
		variant = BadgeDefault
	}

	// Build class string
	class := MergeClasses(
		badgeBaseClasses,
		badgeVariantClasses[variant],
		props.Class,
	)

	return core.Span(core.Class(class), core.Text(props.Text))
}
