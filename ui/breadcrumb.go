package ui

import "github.com/dougbarrett/gux/core"

// BreadcrumbItem represents a single breadcrumb entry.
type BreadcrumbItem struct {
	Label string // Display text
	Href  string // Link URL (empty for current page)
	Icon  string // Optional icon name
}

// BreadcrumbProps configures the Breadcrumb component.
type BreadcrumbProps struct {
	Items    []BreadcrumbItem // Breadcrumb trail (last item is current page)
	HomeIcon bool             // Show home icon for first item (default: true)
	Class    string           // Additional classes for container
}

// Breadcrumb creates a navigation breadcrumb trail.
//
// The last item in the list is treated as the current page (non-clickable).
// All other items are rendered as links.
//
// Example:
//
//	Breadcrumb(BreadcrumbProps{
//	    Items: []BreadcrumbItem{
//	        {Label: "Home", Href: "/"},
//	        {Label: "Users", Href: "/admin/users"},
//	        {Label: "John Doe"}, // Current page (no href)
//	    },
//	})
func Breadcrumb(props BreadcrumbProps) core.Node {
	if len(props.Items) == 0 {
		return core.Frag()
	}

	containerClass := MergeClasses(
		"flex items-center gap-2 text-sm",
		props.Class,
	)

	var content []core.Node

	for i, item := range props.Items {
		isLast := i == len(props.Items)-1
		isFirst := i == 0

		// Separator (except before first item)
		if i > 0 {
			content = append(content,
				core.Span(core.Class("text-gray-400 dark:text-gray-500"),
					Icon(IconProps{Name: "chevron-right", Size: IconSM}),
				),
			)
		}

		// Build item content
		var itemContent []core.Node

		// Icon for first item (home) or custom icon
		if item.Icon != "" {
			itemContent = append(itemContent,
				Icon(IconProps{Name: item.Icon, Size: IconSM, Class: "mr-1.5"}),
			)
		} else if isFirst && props.HomeIcon {
			itemContent = append(itemContent,
				Icon(IconProps{Name: "home", Size: IconSM, Class: "mr-1.5"}),
			)
		}

		// Label
		itemContent = append(itemContent, core.Text(item.Label))

		if isLast || item.Href == "" {
			// Current page - non-clickable
			content = append(content,
				core.Span(core.Class("text-gray-900 dark:text-white font-medium flex items-center"),
					itemContent...,
				),
			)
		} else {
			// Link
			content = append(content,
				core.A(core.Attrs{
					Href:  item.Href,
					Class: "text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 flex items-center transition-colors",
				}, itemContent...),
			)
		}
	}

	return core.El("nav", core.Attrs{
		Class: containerClass,
		Extra: map[string]string{
			"aria-label": "Breadcrumb",
		},
	}, content...)
}

// BreadcrumbSeparator creates a custom separator element.
// Use this if you want a different separator style.
//
// Example:
//
//	BreadcrumbSeparator(BreadcrumbSeparatorProps{
//	    Icon: "arrow-right",
//	})
type BreadcrumbSeparatorProps struct {
	Icon  string // Icon name (default: chevron-right)
	Class string // Additional classes
}

func BreadcrumbSeparator(props BreadcrumbSeparatorProps) core.Node {
	icon := props.Icon
	if icon == "" {
		icon = "chevron-right"
	}

	class := MergeClasses(
		"text-gray-400 dark:text-gray-500",
		props.Class,
	)

	return core.Span(core.Class(class),
		Icon(IconProps{Name: icon, Size: IconSM}),
	)
}

// PageHeaderProps configures the PageHeader component.
type PageHeaderProps struct {
	Title       string           // Page title
	Subtitle    string           // Optional subtitle/description
	Breadcrumbs []BreadcrumbItem // Optional breadcrumb trail
	Actions     []core.Node      // Optional action buttons (right side)
	Class       string           // Additional classes
}

// PageHeader creates a consistent page header with optional breadcrumbs.
//
// This combines a breadcrumb trail, title, subtitle, and action buttons
// in a standard layout pattern for admin/dashboard pages.
//
// Example:
//
//	PageHeader(PageHeaderProps{
//	    Breadcrumbs: []BreadcrumbItem{
//	        {Label: "Home", Href: "/"},
//	        {Label: "Users", Href: "/admin/users"},
//	        {Label: "John Doe"},
//	    },
//	    Title:    "John Doe",
//	    Subtitle: "User Profile",
//	    Actions: []core.Node{
//	        ui.Button(ui.ButtonProps{
//	            Variant:  "default",
//	            Children: []core.Node{core.Text("Edit")},
//	        }),
//	    },
//	})
func PageHeader(props PageHeaderProps) core.Node {
	containerClass := MergeClasses(
		"mb-6",
		props.Class,
	)

	var content []core.Node

	// Breadcrumbs
	if len(props.Breadcrumbs) > 0 {
		content = append(content,
			core.Div(core.Class("mb-4"),
				Breadcrumb(BreadcrumbProps{
					Items:    props.Breadcrumbs,
					HomeIcon: true,
				}),
			),
		)
	}

	// Title row with optional actions
	titleRowClass := "flex items-start justify-between gap-4"
	var titleContent []core.Node

	// Title and subtitle
	titleBlock := []core.Node{
		core.H1(core.Class("text-2xl md:text-3xl font-bold text-gray-900 dark:text-white"),
			core.Text(props.Title),
		),
	}
	if props.Subtitle != "" {
		titleBlock = append(titleBlock,
			core.P(core.Class("mt-1 text-gray-600 dark:text-gray-400"),
				core.Text(props.Subtitle),
			),
		)
	}
	titleContent = append(titleContent,
		core.Div(core.Attrs{}, titleBlock...),
	)

	// Actions
	if len(props.Actions) > 0 {
		titleContent = append(titleContent,
			core.Div(core.Class("flex items-center gap-2 flex-shrink-0"),
				props.Actions...,
			),
		)
	}

	content = append(content,
		core.Div(core.Class(titleRowClass), titleContent...),
	)

	return core.Div(core.Class(containerClass), content...)
}
