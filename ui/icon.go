package ui

import "github.com/dougbarrett/gux/core"

// IconSize defines the size of an icon.
type IconSize string

const (
	IconXS IconSize = "xs"
	IconSM IconSize = "sm"
	IconMD IconSize = "md"
	IconLG IconSize = "lg"
	IconXL IconSize = "xl"
)

// IconProps configures an Icon component.
type IconProps struct {
	Name  string   // Icon name (e.g., "home", "settings")
	Size  IconSize // Size (default: md)
	Class string   // Additional classes
}

// iconSizeClasses maps sizes to Tailwind width/height classes.
var iconSizeClasses = map[IconSize]string{
	IconXS: "w-3 h-3",
	IconSM: "w-4 h-4",
	IconMD: "w-5 h-5",
	IconLG: "w-6 h-6",
	IconXL: "w-8 h-8",
}

// iconPaths contains SVG path data for each icon.
// Icons are from Heroicons (https://heroicons.com) - outline style.
var iconPaths = map[string]string{
	"home": "M2.25 12l8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25",

	"folder": "M2.25 12.75V12A2.25 2.25 0 014.5 9.75h15A2.25 2.25 0 0121.75 12v.75m-8.69-6.44l-2.12-2.12a1.5 1.5 0 00-1.061-.44H4.5A2.25 2.25 0 002.25 6v12a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9a2.25 2.25 0 00-2.25-2.25h-5.379a1.5 1.5 0 01-1.06-.44z",

	"settings": "M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z M15 12a3 3 0 11-6 0 3 3 0 016 0z",

	"users": "M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z",

	"activity": "M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z",

	"menu": "M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5",

	"x": "M6 18L18 6M6 6l12 12",

	"chevron-left": "M15.75 19.5L8.25 12l7.5-7.5",

	"chevron-right": "M8.25 4.5l7.5 7.5-7.5 7.5",

	"chevron-down": "M19.5 8.25l-7.5 7.5-7.5-7.5",

	"chevron-up": "M4.5 15.75l7.5-7.5 7.5 7.5",

	"bell": "M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0",

	"plus": "M12 4.5v15m7.5-7.5h-15",

	"search": "M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z",

	"user": "M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z",

	"logout": "M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15M12 9l-3 3m0 0l3 3m-3-3h12.75",

	"document": "M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z",

	"chart": "M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z",
}

// Icon creates an SVG icon component.
//
// Example:
//
//	Icon(IconProps{
//	    Name: "home",
//	    Size: IconMD,
//	})
func Icon(props IconProps) core.Node {
	// Apply defaults
	size := props.Size
	if size == "" {
		size = IconMD
	}

	// Get path data
	pathData, ok := iconPaths[props.Name]
	if !ok {
		// Return empty node for unknown icons
		return core.Frag()
	}

	// Build class string - include stroke classes for CSS-based coloring
	class := MergeClasses(
		iconSizeClasses[size],
		"flex-shrink-0",
		"stroke-current",
		"[stroke-width:1.5]",
		props.Class,
	)

	// Create SVG element with path
	// Use both CSS classes and SVG attributes for maximum compatibility
	return core.El("svg", core.Attrs{
		Class: class,
		Extra: map[string]string{
			"xmlns":       "http://www.w3.org/2000/svg",
			"fill":        "none",
			"viewBox":     "0 0 24 24",
			"aria-hidden": "true",
		},
	},
		core.El("path", core.Attrs{
			Extra: map[string]string{
				"stroke":          "currentColor",
				"stroke-width":    "1.5",
				"stroke-linecap":  "round",
				"stroke-linejoin": "round",
				"d":               pathData,
			},
		}),
	)
}

// IconNames returns all available icon names.
func IconNames() []string {
	names := make([]string, 0, len(iconPaths))
	for name := range iconPaths {
		names = append(names, name)
	}
	return names
}
