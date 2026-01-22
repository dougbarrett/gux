package ui

import (
	"strings"

	"github.com/dougbarrett/gux/core"
)

// AvatarSize defines the size of an avatar.
type AvatarSize string

const (
	AvatarSM AvatarSize = "sm"
	AvatarMD AvatarSize = "md"
	AvatarLG AvatarSize = "lg"
)

// AvatarProps configures an Avatar component.
type AvatarProps struct {
	Src   string     // Image URL (if empty, shows initials)
	Alt   string     // Alt text for image
	Name  string     // Name for generating initials fallback
	Size  AvatarSize // Size (default: md)
	Class string     // Additional classes (override defaults)
}

// avatarSizeClasses maps sizes to their Tailwind classes.
var avatarSizeClasses = map[AvatarSize]string{
	AvatarSM: "w-8 h-8 text-xs",
	AvatarMD: "w-10 h-10 text-sm",
	AvatarLG: "w-12 h-12 text-base",
}

// avatarBaseClasses are the common classes applied to all avatars.
const avatarBaseClasses = "rounded-full overflow-hidden flex items-center justify-center"

// avatarInitialsClasses are applied when showing initials fallback.
const avatarInitialsClasses = "bg-gray-200 dark:bg-gray-600 text-gray-600 dark:text-gray-200 font-medium"

// getInitials extracts initials from a name.
// Returns "?" for empty or whitespace-only names.
func getInitials(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "?"
	}
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "?"
	}
	if len(parts) == 1 {
		if len(parts[0]) > 0 {
			return strings.ToUpper(string(parts[0][0]))
		}
		return "?"
	}
	first := string(parts[0][0])
	last := string(parts[len(parts)-1][0])
	return strings.ToUpper(first + last)
}

// Avatar creates an avatar component showing an image or initials fallback.
//
// Example with image:
//
//	Avatar(AvatarProps{
//	    Src:  "/images/user.jpg",
//	    Alt:  "John Doe",
//	    Size: AvatarMD,
//	})
//
// Example with initials:
//
//	Avatar(AvatarProps{
//	    Name: "John Doe",
//	    Size: AvatarLG,
//	})
func Avatar(props AvatarProps) core.Node {
	// Apply defaults
	size := props.Size
	if size == "" {
		size = AvatarMD
	}

	// Decide rendering at build time (SSR-compatible)
	if props.Src != "" {
		// Render image avatar
		class := MergeClasses(
			avatarBaseClasses,
			avatarSizeClasses[size],
			props.Class,
		)

		return core.Div(core.Class(class),
			core.Img(core.Attrs{
				Src:   props.Src,
				Alt:   props.Alt,
				Class: "w-full h-full object-cover",
			}),
		)
	}

	// Render initials fallback
	class := MergeClasses(
		avatarBaseClasses,
		avatarSizeClasses[size],
		avatarInitialsClasses,
		props.Class,
	)

	initials := getInitials(props.Name)
	return core.Div(core.Class(class), core.Text(initials))
}
