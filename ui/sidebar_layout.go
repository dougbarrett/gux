package ui

import "github.com/dougbarrett/gux/core"

// SidebarLayoutProps configures the SidebarLayout component.
type SidebarLayoutProps struct {
	Sidebar       core.Node   // The Sidebar component
	MobileOpen    bool        // For overlay rendering (passed to overlay)
	OnCloseMobile func()      // Overlay click handler
	Class         string      // Additional classes for the layout container
	Children      []core.Node // Main content (typically SidebarMain)
}

// SidebarLayout creates a responsive layout with sidebar and main content.
//
// On desktop, the sidebar is positioned statically alongside the main content.
// On mobile, the sidebar overlays the content when open.
//
// Example:
//
//	SidebarLayout(SidebarLayoutProps{
//	    Sidebar: Sidebar(SidebarProps{...}),
//	    Children: []core.Node{
//	        SidebarMain(SidebarMainProps{
//	            Children: []core.Node{...},
//	        }),
//	    },
//	})
func SidebarLayout(props SidebarLayoutProps) core.Node {
	class := MergeClasses(
		"flex min-h-screen bg-gray-50 dark:bg-gray-900",
		props.Class,
	)

	var content []core.Node

	// Add sidebar
	if props.Sidebar != nil {
		content = append(content, props.Sidebar)
	}

	// Add main content
	content = append(content, props.Children...)

	return core.Div(core.Class(class), content...)
}

// SidebarMainProps configures the SidebarMain component.
type SidebarMainProps struct {
	Collapsed     bool        // Adjust margin for collapsed sidebar
	MobileOpen    bool        // Is mobile sidebar open (for accessibility)
	OnOpenMobile  func()      // Callback to open mobile sidebar
	ShowHeader    bool        // Show header with mobile hamburger (default: true)
	HeaderTitle   string      // Optional title in the header
	Class         string      // Additional classes
	ContentClass  string      // Additional classes for the content area
	Children      []core.Node // Main content
}

// SidebarMain creates the main content area that adjusts for the sidebar.
//
// Includes a responsive header with mobile hamburger menu toggle.
//
// Example:
//
//	SidebarMain(SidebarMainProps{
//	    Collapsed:    collapsed.Get(),
//	    OnOpenMobile: func() { mobileOpen.Set(true) },
//	    Children: []core.Node{
//	        core.H1(core.Class("text-2xl"), core.Text("Dashboard")),
//	    },
//	})
func SidebarMain(props SidebarMainProps) core.Node {
	// Main wrapper - fills remaining space
	mainClass := MergeClasses(
		"flex-1 flex flex-col min-w-0",
		props.Class,
	)

	var content []core.Node

	// Mobile header with hamburger menu (shown on mobile only)
	showHeader := props.ShowHeader
	if props.OnOpenMobile != nil {
		// Default to showing header if callback is provided
		showHeader = true
	}

	if showHeader && props.OnOpenMobile != nil {
		headerClass := "md:hidden flex items-center gap-4 p-4 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700"

		headerContent := []core.Node{
			// Hamburger button
			core.Button(core.Attrs{
				Class:   "p-2 text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors cursor-pointer",
				OnClick: props.OnOpenMobile,
				Extra: map[string]string{
					"aria-label":    "Open sidebar",
					"aria-expanded": boolToAttr(props.MobileOpen),
				},
			}, Icon(IconProps{Name: "menu", Size: IconLG})),
		}

		// Optional title
		if props.HeaderTitle != "" {
			headerContent = append(headerContent,
				core.Span(core.Class("font-semibold text-gray-900 dark:text-white"),
					core.Text(props.HeaderTitle),
				),
			)
		}

		content = append(content,
			core.Div(core.Class(headerClass), headerContent...),
		)
	}

	// Content area
	contentClass := MergeClasses(
		"flex-1 p-4 md:p-8 overflow-auto",
		props.ContentClass,
	)

	content = append(content,
		core.El("main", core.Class(contentClass), props.Children...),
	)

	return core.Div(core.Class(mainClass), content...)
}

// MobileMenuButton creates a standalone hamburger button for custom layouts.
//
// Example:
//
//	MobileMenuButton(MobileMenuButtonProps{
//	    OnClick:  func() { mobileOpen.Set(true) },
//	    IsOpen:   mobileOpen.Get(),
//	})
type MobileMenuButtonProps struct {
	OnClick func() // Click handler to open mobile sidebar
	IsOpen  bool   // Current state of mobile sidebar
	Class   string // Additional classes
}

// MobileMenuButton creates a hamburger menu button for mobile.
func MobileMenuButton(props MobileMenuButtonProps) core.Node {
	class := MergeClasses(
		"md:hidden p-2 text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors cursor-pointer",
		props.Class,
	)

	return core.Button(core.Attrs{
		Class:   class,
		OnClick: props.OnClick,
		Extra: map[string]string{
			"aria-label":    "Open sidebar",
			"aria-expanded": boolToAttr(props.IsOpen),
		},
	}, Icon(IconProps{Name: "menu", Size: IconLG}))
}
