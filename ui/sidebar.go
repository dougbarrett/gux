package ui

import "github.com/dougbarrett/gux/core"

// SidebarProps configures the main Sidebar container.
type SidebarProps struct {
	Collapsed     bool        // Desktop: w-64 (false) or w-16 (true) icon-only
	MobileOpen    bool        // Mobile: visible (true) or hidden (false)
	OnCloseMobile func()      // Callback when clicking overlay or close button
	Class         string      // Additional classes
	Children      []core.Node // SidebarHeader, SidebarNav, SidebarFooter
}

// Sidebar creates a responsive sidebar navigation container.
//
// The sidebar is fixed on mobile (slides in/out) and static on desktop.
// It supports collapsed mode (icon-only) on desktop.
//
// Example:
//
//	collapsed := r.StateBool("collapsed", false)
//	mobileOpen := r.StateBool("mobileOpen", false)
//	Sidebar(SidebarProps{
//	    Collapsed:     collapsed.Get(),
//	    MobileOpen:    mobileOpen.Get(),
//	    OnCloseMobile: func() { mobileOpen.Set(false) },
//	    Children: []core.Node{
//	        SidebarHeader(SidebarHeaderProps{...}),
//	        SidebarNav(SidebarNavProps{...}),
//	        SidebarFooter(SidebarFooterProps{...}),
//	    },
//	})
func Sidebar(props SidebarProps) core.Node {
	// Determine width class based on collapsed state
	widthClass := "w-64"
	if props.Collapsed {
		widthClass = "w-16"
	}

	// Mobile transform class
	mobileTransform := "-translate-x-full"
	if props.MobileOpen {
		mobileTransform = "translate-x-0"
	}

	// Sidebar element classes
	sidebarClass := MergeClasses(
		// Base: fixed on mobile, static on desktop
		"fixed md:static inset-y-0 left-0 z-50",
		// Width transitions
		widthClass,
		// Mobile slide in/out
		"transform", mobileTransform, "md:translate-x-0",
		// Appearance
		"bg-gray-800 text-white flex flex-col h-screen",
		// Smooth transitions
		"transition-all duration-300 ease-in-out",
		props.Class,
	)

	// Build content
	children := []core.Node{}

	// Mobile overlay (hidden on desktop)
	if props.OnCloseMobile != nil {
		overlayClass := MergeClasses(
			"fixed inset-0 bg-black/50 z-40 md:hidden",
			ConditionalClass(!props.MobileOpen, "hidden"),
		)
		children = append(children,
			core.Div(core.Attrs{
				Class:   overlayClass,
				OnClick: props.OnCloseMobile,
			}),
		)
	}

	// Sidebar container
	children = append(children,
		core.El("aside", core.Attrs{
			Class: sidebarClass,
			Extra: map[string]string{
				"role": "complementary",
			},
		}, props.Children...),
	)

	return core.Frag(children...)
}

// SidebarHeaderProps configures the SidebarHeader component.
type SidebarHeaderProps struct {
	Title            string      // Brand/app title
	Logo             core.Node   // Optional logo node
	Collapsed        bool        // Hide title when collapsed
	OnToggleCollapse func()      // Desktop collapse toggle callback
	OnCloseMobile    func()      // Mobile close button callback
	Class            string      // Additional classes
}

// SidebarHeader creates the header area with brand and toggle buttons.
//
// Example:
//
//	SidebarHeader(SidebarHeaderProps{
//	    Title:            "My App",
//	    Collapsed:        collapsed.Get(),
//	    OnToggleCollapse: func() { collapsed.Set(!collapsed.Get()) },
//	    OnCloseMobile:    func() { mobileOpen.Set(false) },
//	})
func SidebarHeader(props SidebarHeaderProps) core.Node {
	// Header container
	headerClass := MergeClasses(
		"border-b border-gray-700 flex items-center justify-between",
		ConditionalClass(props.Collapsed, "p-2 flex-col gap-2"),
		ConditionalClass(!props.Collapsed, "p-4"),
		props.Class,
	)

	var content []core.Node

	// Logo/title area
	if props.Logo != nil || props.Title != "" {
		titleClass := MergeClasses(
			"font-bold whitespace-nowrap overflow-hidden",
			ConditionalClass(props.Collapsed, "hidden"),
			ConditionalClass(!props.Collapsed, "text-xl"),
		)

		if props.Logo != nil {
			content = append(content, props.Logo)
		}
		if props.Title != "" && !props.Collapsed {
			content = append(content,
				core.El("h1", core.Class(titleClass), core.Text(props.Title)),
			)
		}
	}

	// Button container
	btnContainerClass := "flex items-center gap-1"
	var buttons []core.Node

	// Desktop collapse toggle
	if props.OnToggleCollapse != nil {
		iconName := "chevron-left"
		ariaLabel := "Collapse sidebar"
		if props.Collapsed {
			iconName = "chevron-right"
			ariaLabel = "Expand sidebar"
		}

		buttons = append(buttons,
			core.Button(core.Attrs{
				Class:   "hidden md:block p-2 text-gray-400 hover:text-white rounded-lg hover:bg-gray-700 transition-colors cursor-pointer",
				OnClick: props.OnToggleCollapse,
				Extra: map[string]string{
					"aria-label":    ariaLabel,
					"aria-expanded": boolToAttr(!props.Collapsed),
				},
			}, Icon(IconProps{Name: iconName, Size: IconMD})),
		)
	}

	// Mobile close button
	if props.OnCloseMobile != nil {
		buttons = append(buttons,
			core.Button(core.Attrs{
				Class:   "md:hidden p-2 text-gray-400 hover:text-white rounded-lg hover:bg-gray-700 transition-colors cursor-pointer",
				OnClick: props.OnCloseMobile,
				Extra:   map[string]string{"aria-label": "Close sidebar"},
			}, Icon(IconProps{Name: "x", Size: IconLG})),
		)
	}

	if len(buttons) > 0 {
		content = append(content,
			core.Div(core.Class(btnContainerClass), buttons...),
		)
	}

	return core.Div(core.Class(headerClass), content...)
}

// SidebarNavProps configures the SidebarNav component.
type SidebarNavProps struct {
	Class    string      // Additional classes
	Children []core.Node // SidebarSection or SidebarItem elements
}

// SidebarNav creates a navigation wrapper with proper ARIA attributes.
//
// Example:
//
//	SidebarNav(SidebarNavProps{
//	    Children: []core.Node{
//	        SidebarItem(SidebarItemProps{...}),
//	        SidebarSection(SidebarSectionProps{...}),
//	    },
//	})
func SidebarNav(props SidebarNavProps) core.Node {
	class := MergeClasses("flex-1 p-2 space-y-1 overflow-y-auto", props.Class)

	return core.El("nav", core.Attrs{
		Class: class,
		Extra: map[string]string{
			"role":       "navigation",
			"aria-label": "Main navigation",
		},
	}, props.Children...)
}

// SidebarSectionProps configures the SidebarSection component.
type SidebarSectionProps struct {
	Title     string      // Optional section title
	Collapsed bool        // Hide title when sidebar is collapsed
	Class     string      // Additional classes
	Children  []core.Node // SidebarItem elements
}

// SidebarSection groups related navigation items with an optional title.
//
// Example:
//
//	SidebarSection(SidebarSectionProps{
//	    Title:     "Management",
//	    Collapsed: collapsed.Get(),
//	    Children: []core.Node{
//	        SidebarItem(SidebarItemProps{...}),
//	    },
//	})
func SidebarSection(props SidebarSectionProps) core.Node {
	class := MergeClasses("space-y-1", props.Class)

	var content []core.Node

	// Section title (hidden when collapsed)
	if props.Title != "" && !props.Collapsed {
		content = append(content,
			core.Div(core.Class("px-3 py-2 text-xs font-semibold text-gray-400 uppercase tracking-wider"),
				core.Text(props.Title),
			),
		)
	}

	content = append(content, props.Children...)

	return core.Div(core.Class(class), content...)
}

// SidebarItemProps configures the SidebarItem component.
type SidebarItemProps struct {
	Label     string // Text label
	Href      string // Navigation URL
	Icon      string // Icon name (from ui.Icon)
	Active    bool   // Is this the current page
	Collapsed bool   // Show icon-only mode
	Badge     string // Optional badge text (e.g., notification count)
	External  bool   // Cross-bundle link (uses External: true)
	Class     string // Additional classes
}

// sidebarItemBaseClasses are common classes for all nav items.
const sidebarItemBaseClasses = "flex items-center rounded-lg transition-colors cursor-pointer"

// sidebarItemExpandedClasses are for expanded (full-width) mode.
const sidebarItemExpandedClasses = "gap-3 px-3 py-2"

// sidebarItemCollapsedClasses are for collapsed (icon-only) mode.
const sidebarItemCollapsedClasses = "justify-center px-2 py-2"

// sidebarItemActiveClasses are for the active/current page.
const sidebarItemActiveClasses = "bg-gray-700 text-white"

// sidebarItemInactiveClasses are for non-active items.
const sidebarItemInactiveClasses = "text-gray-300 hover:bg-gray-700 hover:text-white"

// SidebarItem creates a navigation link with icon and optional badge.
//
// When collapsed, shows only the icon with a tooltip on hover.
//
// Example:
//
//	SidebarItem(SidebarItemProps{
//	    Label:     "Dashboard",
//	    Href:      "/",
//	    Icon:      "home",
//	    Active:    currentPath == "/",
//	    Collapsed: collapsed.Get(),
//	})
func SidebarItem(props SidebarItemProps) core.Node {
	// Build classes
	stateClass := sidebarItemInactiveClasses
	if props.Active {
		stateClass = sidebarItemActiveClasses
	}

	sizeClass := sidebarItemExpandedClasses
	if props.Collapsed {
		sizeClass = sidebarItemCollapsedClasses
	}

	class := MergeClasses(
		sidebarItemBaseClasses,
		sizeClass,
		stateClass,
		"group relative",
		props.Class,
	)

	// Build ARIA attributes
	extra := map[string]string{}
	if props.Active {
		extra["aria-current"] = "page"
	}

	// Build content
	var content []core.Node

	// Icon
	if props.Icon != "" {
		content = append(content,
			Icon(IconProps{Name: props.Icon, Size: IconMD, Class: "flex-shrink-0"}),
		)
	}

	// Label (hidden when collapsed)
	if !props.Collapsed {
		labelClass := "whitespace-nowrap overflow-hidden"
		content = append(content,
			core.Span(core.Class(labelClass), core.Text(props.Label)),
		)
	}

	// Badge (hidden when collapsed)
	if props.Badge != "" && !props.Collapsed {
		content = append(content,
			core.Span(core.Class("ml-auto bg-blue-600 text-white text-xs px-2 py-0.5 rounded-full"),
				core.Text(props.Badge),
			),
		)
	}

	// Tooltip for collapsed mode (CSS-driven visibility)
	if props.Collapsed {
		tooltipClass := MergeClasses(
			"absolute left-full ml-2 px-2 py-1",
			"bg-gray-900 text-white text-sm rounded shadow-lg whitespace-nowrap",
			"opacity-0 invisible group-hover:opacity-100 group-hover:visible",
			"transition-opacity duration-200 pointer-events-none z-50",
		)
		content = append(content,
			core.Div(core.Attrs{
				Class: tooltipClass,
				Extra: map[string]string{"role": "tooltip"},
			}, core.Text(props.Label)),
		)
	}

	// Link element
	attrs := core.Attrs{
		Href:     props.Href,
		Class:    class,
		External: props.External,
		Extra:    extra,
	}

	return core.A(attrs, content...)
}

// SidebarFooterProps configures the SidebarFooter component.
type SidebarFooterProps struct {
	Collapsed bool        // Adjust layout for collapsed mode
	Class     string      // Additional classes
	Children  []core.Node // Content (typically user info, avatar, etc.)
}

// SidebarFooter creates the footer area for user profile and actions.
//
// Example:
//
//	SidebarFooter(SidebarFooterProps{
//	    Collapsed: collapsed.Get(),
//	    Children: []core.Node{
//	        // User avatar and info
//	    },
//	})
func SidebarFooter(props SidebarFooterProps) core.Node {
	class := MergeClasses(
		"border-t border-gray-700 p-2",
		ConditionalClass(props.Collapsed, "flex justify-center"),
		props.Class,
	)

	return core.Div(core.Class(class), props.Children...)
}

// SidebarUserProps configures the SidebarUser component.
type SidebarUserProps struct {
	Name      string // User's display name
	Email     string // User's email (optional)
	AvatarSrc string // Avatar image URL (optional, uses initials if empty)
	Href      string // Link to profile page
	Collapsed bool   // Show avatar-only mode
	Class     string // Additional classes
}

// SidebarUser creates a user profile link for the sidebar footer.
//
// Example:
//
//	SidebarUser(SidebarUserProps{
//	    Name:      "John Doe",
//	    Email:     "john@example.com",
//	    Href:      "/profile",
//	    Collapsed: collapsed.Get(),
//	})
func SidebarUser(props SidebarUserProps) core.Node {
	class := MergeClasses(
		"flex items-center rounded-lg transition-colors cursor-pointer",
		"text-gray-300 hover:bg-gray-700 hover:text-white",
		ConditionalClass(props.Collapsed, "justify-center p-2"),
		ConditionalClass(!props.Collapsed, "gap-3 p-2"),
		"group relative",
		props.Class,
	)

	var content []core.Node

	// Avatar
	content = append(content,
		Avatar(AvatarProps{
			Src:   props.AvatarSrc,
			Name:  props.Name,
			Size:  AvatarMD,
			Class: "flex-shrink-0",
		}),
	)

	// User info (hidden when collapsed)
	if !props.Collapsed {
		infoContent := []core.Node{
			core.Div(core.Class("font-medium text-white text-sm truncate"),
				core.Text(props.Name),
			),
		}
		if props.Email != "" {
			infoContent = append(infoContent,
				core.Div(core.Class("text-xs text-gray-400 truncate"),
					core.Text(props.Email),
				),
			)
		}
		content = append(content,
			core.Div(core.Class("flex-1 min-w-0"), infoContent...),
		)
	}

	// Tooltip for collapsed mode
	if props.Collapsed {
		tooltipClass := MergeClasses(
			"absolute left-full ml-2 px-2 py-1",
			"bg-gray-900 text-white text-sm rounded shadow-lg whitespace-nowrap",
			"opacity-0 invisible group-hover:opacity-100 group-hover:visible",
			"transition-opacity duration-200 pointer-events-none z-50",
		)
		content = append(content,
			core.Div(core.Attrs{
				Class: tooltipClass,
				Extra: map[string]string{"role": "tooltip"},
			}, core.Text(props.Name)),
		)
	}

	return core.A(core.Attrs{
		Href:  props.Href,
		Class: class,
	}, content...)
}
