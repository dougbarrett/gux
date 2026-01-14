//go:build js && wasm

package components

import (
	"syscall/js"
)

// Theme represents light or dark mode
type Theme string

const (
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
	ThemeSystem Theme = "system"
)

// ThemeManager handles dark/light mode switching
type ThemeManager struct {
	current     Theme
	subscribers []func(Theme)
}

var globalThemeManager *ThemeManager

// InitTheme initializes the global theme manager
func InitTheme() *ThemeManager {
	if globalThemeManager != nil {
		return globalThemeManager
	}

	globalThemeManager = &ThemeManager{
		current: ThemeLight,
	}

	// Check for saved preference
	saved := js.Global().Get("localStorage").Call("getItem", "theme")
	if !saved.IsNull() && !saved.IsUndefined() {
		globalThemeManager.current = Theme(saved.String())
	}

	globalThemeManager.apply()

	// Listen for system preference changes
	mediaQuery := js.Global().Call("matchMedia", "(prefers-color-scheme: dark)")
	mediaQuery.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) any {
		if globalThemeManager.current == ThemeSystem {
			globalThemeManager.apply()
		}
		return nil
	}))

	return globalThemeManager
}

// GetTheme returns the current theme
func GetTheme() Theme {
	if globalThemeManager == nil {
		InitTheme()
	}
	return globalThemeManager.current
}

// SetTheme changes the theme
func SetTheme(theme Theme) {
	if globalThemeManager == nil {
		InitTheme()
	}
	globalThemeManager.current = theme
	js.Global().Get("localStorage").Call("setItem", "theme", string(theme))
	globalThemeManager.apply()
	globalThemeManager.notify()
}

// ToggleTheme switches between light and dark
func ToggleTheme() {
	if globalThemeManager == nil {
		InitTheme()
	}

	if globalThemeManager.isDark() {
		SetTheme(ThemeLight)
	} else {
		SetTheme(ThemeDark)
	}
}

// OnThemeChange subscribes to theme changes
func OnThemeChange(fn func(Theme)) func() {
	if globalThemeManager == nil {
		InitTheme()
	}

	globalThemeManager.subscribers = append(globalThemeManager.subscribers, fn)
	idx := len(globalThemeManager.subscribers) - 1

	return func() {
		globalThemeManager.subscribers = append(
			globalThemeManager.subscribers[:idx],
			globalThemeManager.subscribers[idx+1:]...,
		)
	}
}

// IsDarkMode returns true if currently in dark mode
func IsDarkMode() bool {
	if globalThemeManager == nil {
		InitTheme()
	}
	return globalThemeManager.isDark()
}

func (tm *ThemeManager) isDark() bool {
	if tm.current == ThemeSystem {
		return js.Global().Call("matchMedia", "(prefers-color-scheme: dark)").Get("matches").Bool()
	}
	return tm.current == ThemeDark
}

func (tm *ThemeManager) apply() {
	document := js.Global().Get("document")
	html := document.Get("documentElement")

	if tm.isDark() {
		html.Get("classList").Call("add", "dark")
	} else {
		html.Get("classList").Call("remove", "dark")
	}
}

func (tm *ThemeManager) notify() {
	for _, fn := range tm.subscribers {
		fn(tm.current)
	}
}

// ThemeToggle creates a theme toggle button
func ThemeToggle() js.Value {
	document := js.Global().Get("document")

	btn := document.Call("createElement", "button")
	btn.Set("className", "p-2 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors cursor-pointer")

	updateIcon := func() {
		if IsDarkMode() {
			btn.Set("innerHTML", `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>`)
		} else {
			btn.Set("innerHTML", `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"></path></svg>`)
		}
	}

	updateIcon()

	btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		ToggleTheme()
		updateIcon()
		return nil
	}))

	return btn
}
