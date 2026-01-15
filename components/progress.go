//go:build js && wasm

package components

import (
	"fmt"
	"syscall/js"
)

// ProgressVariant defines progress bar color
type ProgressVariant string

const (
	ProgressDefault ProgressVariant = "default"
	ProgressPrimary ProgressVariant = "primary"
	ProgressSuccess ProgressVariant = "success"
	ProgressWarning ProgressVariant = "warning"
	ProgressError   ProgressVariant = "error"
)

var progressColors = map[ProgressVariant]string{
	ProgressDefault: "bg-gray-600",
	ProgressPrimary: "bg-blue-600",
	ProgressSuccess: "bg-green-600",
	ProgressWarning: "bg-yellow-500",
	ProgressError:   "bg-red-600",
}

// ProgressProps configures a Progress bar
type ProgressProps struct {
	Value         int // 0-100
	Variant       ProgressVariant
	ShowLabel     bool
	Striped       bool
	Animated      bool
	Height        string // e.g., "h-2", "h-4"
	Label         string // custom label, defaults to percentage
	Indeterminate bool   // shows infinite animation
	AriaLabel     string // accessible label for context (e.g., "Uploading file")
}

// Progress is a progress bar component
type Progress struct {
	container js.Value
	track     js.Value
	bar       js.Value
	label     js.Value
	props     ProgressProps
}

// NewProgress creates a new Progress component
func NewProgress(props ProgressProps) *Progress {
	document := js.Global().Get("document")

	container := document.Call("createElement", "div")
	container.Set("className", "w-full")

	height := props.Height
	if height == "" {
		height = "h-2"
	}

	variant := props.Variant
	if variant == "" {
		variant = ProgressPrimary
	}

	// Track (ARIA progressbar widget)
	track := document.Call("createElement", "div")
	track.Set("className", "w-full "+height+" bg-gray-200 rounded-full overflow-hidden")
	track.Call("setAttribute", "role", "progressbar")
	track.Call("setAttribute", "aria-valuemin", "0")
	track.Call("setAttribute", "aria-valuemax", "100")
	if !props.Indeterminate {
		track.Call("setAttribute", "aria-valuenow", fmt.Sprintf("%d", props.Value))
	}
	if props.AriaLabel != "" {
		track.Call("setAttribute", "aria-label", props.AriaLabel)
	}

	// Bar
	bar := document.Call("createElement", "div")
	barClass := height + " rounded-full transition-all duration-300 " + progressColors[variant]

	if props.Striped {
		barClass += " bg-stripes"
	}
	if props.Animated && props.Striped {
		barClass += " animate-stripes"
	}
	if props.Indeterminate {
		barClass += " animate-indeterminate"
		bar.Get("style").Set("width", "30%")
	} else {
		bar.Get("style").Set("width", fmt.Sprintf("%d%%", props.Value))
	}

	bar.Set("className", barClass)
	track.Call("appendChild", bar)
	container.Call("appendChild", track)

	p := &Progress{
		container: container,
		track:     track,
		bar:       bar,
		props:     props,
	}

	// Label
	if props.ShowLabel {
		labelContainer := document.Call("createElement", "div")
		labelContainer.Set("className", "flex justify-between text-sm text-gray-600 mt-1")

		label := document.Call("createElement", "span")
		if props.Label != "" {
			label.Set("textContent", props.Label)
		}

		percent := document.Call("createElement", "span")
		percent.Set("textContent", fmt.Sprintf("%d%%", props.Value))

		labelContainer.Call("appendChild", label)
		labelContainer.Call("appendChild", percent)
		container.Call("appendChild", labelContainer)

		p.label = percent
	}

	// Add CSS for stripes if needed
	if props.Striped {
		addProgressStyles()
	}

	return p
}

// Element returns the DOM element
func (p *Progress) Element() js.Value {
	return p.container
}

// SetValue updates the progress value
func (p *Progress) SetValue(value int) {
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	p.props.Value = value
	p.bar.Get("style").Set("width", fmt.Sprintf("%d%%", value))
	// Update ARIA value for screen readers
	p.track.Call("setAttribute", "aria-valuenow", fmt.Sprintf("%d", value))
	if p.label.Truthy() {
		p.label.Set("textContent", fmt.Sprintf("%d%%", value))
	}
}

var progressStylesAdded = false

func addProgressStyles() {
	if progressStylesAdded {
		return
	}
	progressStylesAdded = true

	document := js.Global().Get("document")
	style := document.Call("createElement", "style")
	style.Set("textContent", `
		.bg-stripes {
			background-image: linear-gradient(
				45deg,
				rgba(255,255,255,0.15) 25%,
				transparent 25%,
				transparent 50%,
				rgba(255,255,255,0.15) 50%,
				rgba(255,255,255,0.15) 75%,
				transparent 75%,
				transparent
			);
			background-size: 1rem 1rem;
		}
		.animate-stripes {
			animation: stripes 1s linear infinite;
		}
		.animate-indeterminate {
			animation: indeterminate 1.5s ease-in-out infinite;
		}
		@keyframes stripes {
			from { background-position: 1rem 0; }
			to { background-position: 0 0; }
		}
		@keyframes indeterminate {
			0% { transform: translateX(-100%); }
			100% { transform: translateX(400%); }
		}
	`)
	document.Get("head").Call("appendChild", style)
}

// Quick progress bar creator
func ProgressBar(value int, variant ProgressVariant) js.Value {
	return NewProgress(ProgressProps{Value: value, Variant: variant}).Element()
}
