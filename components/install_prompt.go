//go:build js && wasm

package components

import (
	"syscall/js"
	"time"
)

// InstallPromptPosition defines where the install prompt banner appears
type InstallPromptPosition string

const (
	InstallPromptBottomLeft  InstallPromptPosition = "bottom-left"
	InstallPromptBottomRight InstallPromptPosition = "bottom-right"
	InstallPromptTopCenter   InstallPromptPosition = "top-center"
)

var installPromptPositionClasses = map[InstallPromptPosition]string{
	InstallPromptBottomLeft:  "fixed bottom-4 left-4",
	InstallPromptBottomRight: "fixed bottom-4 right-4",
	InstallPromptTopCenter:   "fixed top-4 left-1/2 -translate-x-1/2",
}

const (
	installPromptDismissKey     = "gux-install-prompt-dismissed"
	installPromptDismissDays    = 7
	installPromptDismissDaysStr = "7"
)

// InstallPromptProps configures the InstallPrompt component
type InstallPromptProps struct {
	Position   InstallPromptPosition
	OnDismiss  func()
	OnInstall  func()
	AppName    string // Optional app name, defaults to "this app"
	AppIconURL string // Optional icon URL
}

// InstallPrompt creates a PWA install prompt banner
type InstallPrompt struct {
	element js.Value
	manager *InstallPromptManager
	visible bool
}

// NewInstallPrompt creates a new InstallPrompt component
func NewInstallPrompt(props InstallPromptProps, manager *InstallPromptManager) *InstallPrompt {
	document := js.Global().Get("document")

	position := props.Position
	if position == "" {
		position = InstallPromptBottomRight
	}

	appName := props.AppName
	if appName == "" {
		appName = "Gux"
	}

	// Container
	container := document.Call("createElement", "div")
	posClass := installPromptPositionClasses[position]
	container.Set("className", posClass+" z-50 max-w-sm bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 p-4 transform transition-all duration-300 translate-y-2 opacity-0 pointer-events-none")
	container.Set("id", "install-prompt-banner")

	// Content wrapper
	content := document.Call("createElement", "div")
	content.Set("className", "flex items-start gap-3")

	// App icon
	icon := document.Call("createElement", "div")
	icon.Set("className", "w-12 h-12 flex-shrink-0 rounded-lg bg-blue-100 dark:bg-blue-900 flex items-center justify-center")
	if props.AppIconURL != "" {
		img := document.Call("createElement", "img")
		img.Set("src", props.AppIconURL)
		img.Set("alt", appName)
		img.Set("className", "w-8 h-8")
		icon.Call("appendChild", img)
	} else {
		iconText := document.Call("createElement", "span")
		iconText.Set("className", "text-2xl")
		iconText.Set("textContent", "📱")
		icon.Call("appendChild", iconText)
	}
	content.Call("appendChild", icon)

	// Text container
	textContainer := document.Call("createElement", "div")
	textContainer.Set("className", "flex-1 min-w-0")

	// Title
	title := document.Call("createElement", "h4")
	title.Set("className", "font-semibold text-gray-900 dark:text-white text-sm")
	title.Set("textContent", "Install "+appName)
	textContainer.Call("appendChild", title)

	// Description
	desc := document.Call("createElement", "p")
	desc.Set("className", "text-xs text-gray-600 dark:text-gray-400 mt-0.5")
	desc.Set("textContent", "Install for faster access and offline use")
	textContainer.Call("appendChild", desc)

	content.Call("appendChild", textContainer)
	container.Call("appendChild", content)

	// Buttons
	buttons := document.Call("createElement", "div")
	buttons.Set("className", "flex gap-2 mt-3")

	ip := &InstallPrompt{
		element: container,
		manager: manager,
		visible: false,
	}

	// Not now button
	notNowBtn := document.Call("createElement", "button")
	notNowBtn.Set("className", "flex-1 px-3 py-1.5 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 transition-colors cursor-pointer")
	notNowBtn.Set("textContent", "Not now")
	notNowBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		ip.dismiss()
		if props.OnDismiss != nil {
			props.OnDismiss()
		}
		return nil
	}))
	buttons.Call("appendChild", notNowBtn)

	// Install button
	installBtn := document.Call("createElement", "button")
	installBtn.Set("className", "flex-1 px-3 py-1.5 text-sm bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors cursor-pointer font-medium")
	installBtn.Set("textContent", "Install")
	installBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		if manager != nil {
			manager.ShowPrompt()
		}
		if props.OnInstall != nil {
			props.OnInstall()
		}
		return nil
	}))
	buttons.Call("appendChild", installBtn)

	container.Call("appendChild", buttons)

	return ip
}

// Element returns the DOM element
func (ip *InstallPrompt) Element() js.Value {
	return ip.element
}

// Show displays the install prompt banner with animation
func (ip *InstallPrompt) Show() {
	if ip.visible {
		return
	}

	// Check if recently dismissed
	if ip.wasRecentlyDismissed() {
		return
	}

	ip.visible = true
	ip.element.Set("className", ip.element.Get("className").String()[:len(ip.element.Get("className").String())-len(" translate-y-2 opacity-0 pointer-events-none")]+" translate-y-0 opacity-100 pointer-events-auto")

	// Force reflow and update classes for animation
	_ = ip.element.Get("offsetHeight")
	currentClass := ip.element.Get("className").String()
	// Remove hidden classes
	newClass := ""
	for _, c := range splitClasses(currentClass) {
		if c != "translate-y-2" && c != "opacity-0" && c != "pointer-events-none" {
			if newClass != "" {
				newClass += " "
			}
			newClass += c
		}
	}
	newClass += " translate-y-0 opacity-100 pointer-events-auto"
	ip.element.Set("className", newClass)
}

// Hide hides the install prompt banner
func (ip *InstallPrompt) Hide() {
	if !ip.visible {
		return
	}
	ip.visible = false

	currentClass := ip.element.Get("className").String()
	newClass := ""
	for _, c := range splitClasses(currentClass) {
		if c != "translate-y-0" && c != "opacity-100" && c != "pointer-events-auto" {
			if newClass != "" {
				newClass += " "
			}
			newClass += c
		}
	}
	newClass += " translate-y-2 opacity-0 pointer-events-none"
	ip.element.Set("className", newClass)
}

// dismiss hides the banner and stores dismissal time
func (ip *InstallPrompt) dismiss() {
	ip.Hide()
	localStorage := js.Global().Get("localStorage")
	dismissTime := time.Now().UnixMilli()
	localStorage.Call("setItem", installPromptDismissKey, js.ValueOf(dismissTime))
}

// wasRecentlyDismissed checks if the prompt was dismissed within the last N days
func (ip *InstallPrompt) wasRecentlyDismissed() bool {
	localStorage := js.Global().Get("localStorage")
	dismissedValue := localStorage.Call("getItem", installPromptDismissKey)
	if dismissedValue.IsNull() || dismissedValue.IsUndefined() {
		return false
	}

	dismissedStr := dismissedValue.String()
	if dismissedStr == "" {
		return false
	}

	// Parse the timestamp
	dismissedTime := js.Global().Call("parseInt", dismissedStr, 10).Int()
	if dismissedTime == 0 {
		return false
	}

	// Check if within the dismiss period
	now := time.Now().UnixMilli()
	dismissPeriod := int64(installPromptDismissDays * 24 * 60 * 60 * 1000) // days in milliseconds
	return (now - int64(dismissedTime)) < dismissPeriod
}

// Helper to split class string
func splitClasses(classes string) []string {
	var result []string
	current := ""
	for _, c := range classes {
		if c == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// InstallPromptManager manages the beforeinstallprompt event lifecycle
type InstallPromptManager struct {
	deferredPrompt js.Value
	canInstall     bool
	callbacks      []func()
	installed      bool
}

// NewInstallPromptManager creates a new manager for PWA install prompts
func NewInstallPromptManager() *InstallPromptManager {
	manager := &InstallPromptManager{
		deferredPrompt: js.Null(),
		canInstall:     false,
		callbacks:      make([]func(), 0),
		installed:      false,
	}

	window := js.Global()

	// Listen for beforeinstallprompt
	beforeInstallHandler := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			event := args[0]
			// Prevent the mini-infobar from appearing
			event.Call("preventDefault")
			// Store the event for later use
			manager.deferredPrompt = event
			manager.canInstall = true
			// Notify all registered callbacks
			for _, cb := range manager.callbacks {
				cb()
			}
			js.Global().Get("console").Call("log", "[InstallPrompt] beforeinstallprompt captured")
		}
		return nil
	})
	window.Call("addEventListener", "beforeinstallprompt", beforeInstallHandler)

	// Listen for appinstalled
	appInstalledHandler := js.FuncOf(func(this js.Value, args []js.Value) any {
		manager.installed = true
		manager.canInstall = false
		manager.deferredPrompt = js.Null()
		js.Global().Get("console").Call("log", "[InstallPrompt] App was installed")
		return nil
	})
	window.Call("addEventListener", "appinstalled", appInstalledHandler)

	return manager
}

// CanInstall returns true if the app can be installed
func (m *InstallPromptManager) CanInstall() bool {
	return m.canInstall && !m.installed
}

// IsInstalled returns true if the app is already installed
func (m *InstallPromptManager) IsInstalled() bool {
	return m.installed
}

// ShowPrompt triggers the native install prompt
func (m *InstallPromptManager) ShowPrompt() {
	if m.deferredPrompt.IsNull() || m.deferredPrompt.IsUndefined() {
		js.Global().Get("console").Call("warn", "[InstallPrompt] No deferred prompt available")
		return
	}

	// Show the native prompt
	m.deferredPrompt.Call("prompt")

	// Wait for user response
	go func() {
		result := m.deferredPrompt.Call("userChoice")
		// userChoice returns a promise, need to handle async
		result.Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
			if len(args) > 0 {
				outcome := args[0].Get("outcome").String()
				js.Global().Get("console").Call("log", "[InstallPrompt] User choice: "+outcome)
				if outcome == "accepted" {
					m.installed = true
					m.canInstall = false
				}
			}
			// Clear the deferred prompt - can only be used once
			m.deferredPrompt = js.Null()
			return nil
		}))
	}()
}

// OnCanInstall registers a callback to be called when install becomes available
func (m *InstallPromptManager) OnCanInstall(callback func()) {
	m.callbacks = append(m.callbacks, callback)
	// If already installable, call immediately
	if m.canInstall {
		callback()
	}
}
