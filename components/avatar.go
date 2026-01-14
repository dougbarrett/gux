//go:build js && wasm

package components

import (
	"strings"
	"syscall/js"
)

// AvatarSize defines avatar sizes
type AvatarSize string

const (
	AvatarXS AvatarSize = "xs" // 24px
	AvatarSM AvatarSize = "sm" // 32px
	AvatarMD AvatarSize = "md" // 40px
	AvatarLG AvatarSize = "lg" // 48px
	AvatarXL AvatarSize = "xl" // 64px
)

var avatarSizes = map[AvatarSize]string{
	AvatarXS: "w-6 h-6 text-xs",
	AvatarSM: "w-8 h-8 text-sm",
	AvatarMD: "w-10 h-10 text-base",
	AvatarLG: "w-12 h-12 text-lg",
	AvatarXL: "w-16 h-16 text-xl",
}

// AvatarProps configures an Avatar
type AvatarProps struct {
	Src      string     // image URL
	Alt      string     // alt text
	Name     string     // for initials fallback
	Size     AvatarSize
	Rounded  bool       // square with rounded corners vs full circle
	Status   string     // "online", "offline", "away", "busy"
	OnClick  func()
}

// Avatar creates a user avatar component
func Avatar(props AvatarProps) js.Value {
	document := js.Global().Get("document")

	size := props.Size
	if size == "" {
		size = AvatarMD
	}

	container := document.Call("createElement", "div")
	container.Set("className", "relative inline-block")

	avatar := document.Call("createElement", "div")
	roundedClass := "rounded-full"
	if props.Rounded {
		roundedClass = "rounded-lg"
	}

	baseClass := avatarSizes[size] + " " + roundedClass + " overflow-hidden flex items-center justify-center bg-gray-200 text-gray-600 font-medium"

	if props.OnClick != nil {
		baseClass += " cursor-pointer hover:opacity-80 transition-opacity"
	}

	avatar.Set("className", baseClass)

	// Image or initials
	if props.Src != "" {
		img := document.Call("createElement", "img")
		img.Set("src", props.Src)
		img.Set("alt", props.Alt)
		img.Set("className", "w-full h-full object-cover")

		// Fallback to initials on error
		img.Call("addEventListener", "error", js.FuncOf(func(this js.Value, args []js.Value) any {
			avatar.Set("innerHTML", "")
			initials := document.Call("createElement", "span")
			initials.Set("textContent", getInitials(props.Name))
			avatar.Call("appendChild", initials)
			return nil
		}))

		avatar.Call("appendChild", img)
	} else {
		initials := document.Call("createElement", "span")
		initials.Set("textContent", getInitials(props.Name))
		avatar.Call("appendChild", initials)
	}

	if props.OnClick != nil {
		avatar.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			props.OnClick()
			return nil
		}))
	}

	container.Call("appendChild", avatar)

	// Status indicator
	if props.Status != "" {
		status := document.Call("createElement", "span")

		var statusColor string
		switch props.Status {
		case "online":
			statusColor = "bg-green-500"
		case "offline":
			statusColor = "bg-gray-400"
		case "away":
			statusColor = "bg-yellow-500"
		case "busy":
			statusColor = "bg-red-500"
		default:
			statusColor = "bg-gray-400"
		}

		status.Set("className", "absolute bottom-0 right-0 w-3 h-3 "+statusColor+" border-2 border-white rounded-full")
		container.Call("appendChild", status)
	}

	return container
}

func getInitials(name string) string {
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

// AvatarGroup creates a stacked group of avatars
func AvatarGroup(avatars []AvatarProps, max int) js.Value {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "flex -space-x-2")

	count := len(avatars)
	if max > 0 && count > max {
		count = max
	}

	for i := 0; i < count; i++ {
		av := Avatar(avatars[i])
		av.Get("firstChild").Get("classList").Call("add", "ring-2", "ring-white")
		container.Call("appendChild", av)
	}

	// Show overflow count
	if max > 0 && len(avatars) > max {
		overflow := document.Call("createElement", "div")
		overflow.Set("className", "flex items-center justify-center w-10 h-10 rounded-full bg-gray-200 text-gray-600 text-sm font-medium ring-2 ring-white")

		remaining := len(avatars) - max
		overflow.Set("textContent", "+"+itoa(remaining))
		container.Call("appendChild", overflow)
	}

	return container
}
