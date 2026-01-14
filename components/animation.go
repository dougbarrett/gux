//go:build js && wasm

package components

import (
	"fmt"
	"syscall/js"
)

// AnimationTiming defines animation timing functions
type AnimationTiming string

const (
	TimingLinear    AnimationTiming = "linear"
	TimingEase      AnimationTiming = "ease"
	TimingEaseIn    AnimationTiming = "ease-in"
	TimingEaseOut   AnimationTiming = "ease-out"
	TimingEaseInOut AnimationTiming = "ease-in-out"
)

// AnimationDirection defines animation direction
type AnimationDirection string

const (
	DirectionNormal    AnimationDirection = "normal"
	DirectionReverse   AnimationDirection = "reverse"
	DirectionAlternate AnimationDirection = "alternate"
)

// Animation represents a CSS animation configuration
type Animation struct {
	Name       string
	Duration   int // milliseconds
	Timing     AnimationTiming
	Delay      int // milliseconds
	Iterations int // 0 = infinite
	Direction  AnimationDirection
	FillMode   string
}

// AnimateProps configures an animation on an element
type AnimateProps struct {
	Element    js.Value
	Animation  Animation
	OnComplete func()
}

// predefined animations CSS
var animationsCSS = `
@keyframes fadeIn {
	from { opacity: 0; }
	to { opacity: 1; }
}

@keyframes fadeOut {
	from { opacity: 1; }
	to { opacity: 0; }
}

@keyframes slideInLeft {
	from { transform: translateX(-100%); opacity: 0; }
	to { transform: translateX(0); opacity: 1; }
}

@keyframes slideInRight {
	from { transform: translateX(100%); opacity: 0; }
	to { transform: translateX(0); opacity: 1; }
}

@keyframes slideInUp {
	from { transform: translateY(100%); opacity: 0; }
	to { transform: translateY(0); opacity: 1; }
}

@keyframes slideInDown {
	from { transform: translateY(-100%); opacity: 0; }
	to { transform: translateY(0); opacity: 1; }
}

@keyframes slideOutLeft {
	from { transform: translateX(0); opacity: 1; }
	to { transform: translateX(-100%); opacity: 0; }
}

@keyframes slideOutRight {
	from { transform: translateX(0); opacity: 1; }
	to { transform: translateX(100%); opacity: 0; }
}

@keyframes slideOutUp {
	from { transform: translateY(0); opacity: 1; }
	to { transform: translateY(-100%); opacity: 0; }
}

@keyframes slideOutDown {
	from { transform: translateY(0); opacity: 1; }
	to { transform: translateY(100%); opacity: 0; }
}

@keyframes scaleIn {
	from { transform: scale(0); opacity: 0; }
	to { transform: scale(1); opacity: 1; }
}

@keyframes scaleOut {
	from { transform: scale(1); opacity: 1; }
	to { transform: scale(0); opacity: 0; }
}

@keyframes bounce {
	0%, 20%, 53%, 100% { transform: translateY(0); }
	40% { transform: translateY(-30px); }
	43% { transform: translateY(-15px); }
	70% { transform: translateY(-4px); }
}

@keyframes shake {
	0%, 100% { transform: translateX(0); }
	10%, 30%, 50%, 70%, 90% { transform: translateX(-10px); }
	20%, 40%, 60%, 80% { transform: translateX(10px); }
}

@keyframes pulse {
	0%, 100% { opacity: 1; }
	50% { opacity: 0.5; }
}

@keyframes spin {
	from { transform: rotate(0deg); }
	to { transform: rotate(360deg); }
}

@keyframes ping {
	75%, 100% { transform: scale(2); opacity: 0; }
}

@keyframes wiggle {
	0%, 100% { transform: rotate(0deg); }
	25% { transform: rotate(-10deg); }
	75% { transform: rotate(10deg); }
}

@keyframes flash {
	0%, 50%, 100% { opacity: 1; }
	25%, 75% { opacity: 0; }
}

/* Transition utility classes */
.transition-all { transition: all 0.3s ease; }
.transition-opacity { transition: opacity 0.3s ease; }
.transition-transform { transition: transform 0.3s ease; }
.transition-colors { transition: background-color 0.3s ease, border-color 0.3s ease, color 0.3s ease; }
.transition-shadow { transition: box-shadow 0.3s ease; }

.duration-150 { transition-duration: 150ms; }
.duration-200 { transition-duration: 200ms; }
.duration-300 { transition-duration: 300ms; }
.duration-500 { transition-duration: 500ms; }
.duration-700 { transition-duration: 700ms; }
.duration-1000 { transition-duration: 1000ms; }
`

var animationsInitialized = false

// InitAnimations adds the animation keyframes CSS to the document
func InitAnimations() {
	if animationsInitialized {
		return
	}

	document := js.Global().Get("document")

	style := document.Call("createElement", "style")
	style.Set("id", "goquery-animations")
	style.Set("textContent", animationsCSS)

	document.Get("head").Call("appendChild", style)
	animationsInitialized = true
}

// Animate applies an animation to an element
func Animate(props AnimateProps) {
	InitAnimations()

	el := props.Element
	anim := props.Animation

	if anim.Duration == 0 {
		anim.Duration = 300
	}
	if anim.Timing == "" {
		anim.Timing = TimingEase
	}
	if anim.Direction == "" {
		anim.Direction = DirectionNormal
	}
	if anim.FillMode == "" {
		anim.FillMode = "forwards"
	}

	iterations := "1"
	if anim.Iterations == 0 {
		iterations = "infinite"
	} else {
		iterations = fmt.Sprintf("%d", anim.Iterations)
	}

	animationValue := fmt.Sprintf("%s %dms %s %dms %s %s %s",
		anim.Name,
		anim.Duration,
		anim.Timing,
		anim.Delay,
		iterations,
		anim.Direction,
		anim.FillMode,
	)

	el.Get("style").Set("animation", animationValue)

	if props.OnComplete != nil {
		totalDuration := anim.Duration + anim.Delay
		js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
			props.OnComplete()
			return nil
		}), totalDuration)
	}
}

// FadeIn fades an element in
func FadeIn(el js.Value, duration int, onComplete func()) {
	Animate(AnimateProps{
		Element:    el,
		Animation:  Animation{Name: "fadeIn", Duration: duration},
		OnComplete: onComplete,
	})
}

// FadeOut fades an element out
func FadeOut(el js.Value, duration int, onComplete func()) {
	Animate(AnimateProps{
		Element:    el,
		Animation:  Animation{Name: "fadeOut", Duration: duration},
		OnComplete: onComplete,
	})
}

// SlideIn slides an element in from a direction
func SlideIn(el js.Value, direction string, duration int, onComplete func()) {
	animName := "slideInRight"
	switch direction {
	case "left":
		animName = "slideInLeft"
	case "right":
		animName = "slideInRight"
	case "up":
		animName = "slideInUp"
	case "down":
		animName = "slideInDown"
	}
	Animate(AnimateProps{
		Element:    el,
		Animation:  Animation{Name: animName, Duration: duration},
		OnComplete: onComplete,
	})
}

// SlideOut slides an element out in a direction
func SlideOut(el js.Value, direction string, duration int, onComplete func()) {
	animName := "slideOutRight"
	switch direction {
	case "left":
		animName = "slideOutLeft"
	case "right":
		animName = "slideOutRight"
	case "up":
		animName = "slideOutUp"
	case "down":
		animName = "slideOutDown"
	}
	Animate(AnimateProps{
		Element:    el,
		Animation:  Animation{Name: animName, Duration: duration},
		OnComplete: onComplete,
	})
}

// ScaleIn scales an element in
func ScaleIn(el js.Value, duration int, onComplete func()) {
	Animate(AnimateProps{
		Element:    el,
		Animation:  Animation{Name: "scaleIn", Duration: duration},
		OnComplete: onComplete,
	})
}

// ScaleOut scales an element out
func ScaleOut(el js.Value, duration int, onComplete func()) {
	Animate(AnimateProps{
		Element:    el,
		Animation:  Animation{Name: "scaleOut", Duration: duration},
		OnComplete: onComplete,
	})
}

// Bounce applies a bounce animation
func Bounce(el js.Value) {
	Animate(AnimateProps{
		Element:   el,
		Animation: Animation{Name: "bounce", Duration: 1000},
	})
}

// Shake applies a shake animation
func Shake(el js.Value) {
	Animate(AnimateProps{
		Element:   el,
		Animation: Animation{Name: "shake", Duration: 500},
	})
}

// Pulse applies a pulse animation
func Pulse(el js.Value, iterations int) {
	Animate(AnimateProps{
		Element:   el,
		Animation: Animation{Name: "pulse", Duration: 1000, Iterations: iterations},
	})
}

// Spin applies a spin animation
func Spin(el js.Value) {
	Animate(AnimateProps{
		Element:   el,
		Animation: Animation{Name: "spin", Duration: 1000, Iterations: 0},
	})
}

// Wiggle applies a wiggle animation
func Wiggle(el js.Value) {
	Animate(AnimateProps{
		Element:   el,
		Animation: Animation{Name: "wiggle", Duration: 500},
	})
}

// Flash applies a flash animation
func Flash(el js.Value, iterations int) {
	Animate(AnimateProps{
		Element:   el,
		Animation: Animation{Name: "flash", Duration: 1000, Iterations: iterations},
	})
}

// Transition applies a CSS transition to an element
type TransitionProps struct {
	Element    js.Value
	Property   string // "all", "opacity", "transform", etc.
	Duration   int    // milliseconds
	Timing     AnimationTiming
	Delay      int // milliseconds
}

// SetTransition sets up a CSS transition on an element
func SetTransition(props TransitionProps) {
	if props.Duration == 0 {
		props.Duration = 300
	}
	if props.Timing == "" {
		props.Timing = TimingEase
	}
	if props.Property == "" {
		props.Property = "all"
	}

	value := fmt.Sprintf("%s %dms %s %dms",
		props.Property,
		props.Duration,
		props.Timing,
		props.Delay,
	)

	props.Element.Get("style").Set("transition", value)
}

// RemoveTransition removes transition from an element
func RemoveTransition(el js.Value) {
	el.Get("style").Set("transition", "")
}

// Stagger applies staggered animations to multiple elements
func Stagger(elements []js.Value, animation Animation, staggerDelay int) {
	for i, el := range elements {
		anim := animation
		anim.Delay = i * staggerDelay
		Animate(AnimateProps{
			Element:   el,
			Animation: anim,
		})
	}
}

// AnimatedList wraps children with staggered fade-in animation
func AnimatedList(children ...js.Value) js.Value {
	document := js.Global().Get("document")
	container := document.Call("createElement", "div")

	for i, child := range children {
		wrapper := document.Call("createElement", "div")
		wrapper.Get("style").Set("opacity", "0")
		wrapper.Call("appendChild", child)
		container.Call("appendChild", wrapper)

		// Stagger the fade-in
		delay := i * 100
		js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
			FadeIn(wrapper, 300, nil)
			return nil
		}), delay)
	}

	return container
}
