//go:build js && wasm

package components

import (
	"syscall/js"
)

// ThemeMode represents light or dark mode
type ThemeMode string

const (
	ThemeLight  ThemeMode = "light"
	ThemeDark   ThemeMode = "dark"
	ThemeSystem ThemeMode = "system"
)

// ThemeColors defines the color palette for a theme
type ThemeColors struct {
	// Background colors
	Background      string
	BackgroundAlt   string
	BackgroundHover string

	// Text colors
	Text        string
	TextMuted   string
	TextInverse string

	// Primary colors
	Primary      string
	PrimaryHover string
	PrimaryText  string

	// Secondary colors
	Secondary      string
	SecondaryHover string
	SecondaryText  string

	// Accent colors
	Accent      string
	AccentHover string
	AccentText  string

	// Status colors
	Success string
	Warning string
	Error   string
	Info    string

	// Border colors
	Border      string
	BorderFocus string

	// Shadow
	Shadow string
}

// DefaultLightColors provides default light theme colors
var DefaultLightColors = ThemeColors{
	Background:      "#ffffff",
	BackgroundAlt:   "#f9fafb",
	BackgroundHover: "#f3f4f6",

	Text:        "#111827",
	TextMuted:   "#6b7280",
	TextInverse: "#ffffff",

	Primary:      "#3b82f6",
	PrimaryHover: "#2563eb",
	PrimaryText:  "#ffffff",

	Secondary:      "#6b7280",
	SecondaryHover: "#4b5563",
	SecondaryText:  "#ffffff",

	Accent:      "#8b5cf6",
	AccentHover: "#7c3aed",
	AccentText:  "#ffffff",

	Success: "#22c55e",
	Warning: "#f59e0b",
	Error:   "#ef4444",
	Info:    "#3b82f6",

	Border:      "#e5e7eb",
	BorderFocus: "#3b82f6",

	Shadow: "rgba(0, 0, 0, 0.1)",
}

// DefaultDarkColors provides default dark theme colors
var DefaultDarkColors = ThemeColors{
	Background:      "#111827",
	BackgroundAlt:   "#1f2937",
	BackgroundHover: "#374151",

	Text:        "#f9fafb",
	TextMuted:   "#9ca3af",
	TextInverse: "#111827",

	Primary:      "#3b82f6",
	PrimaryHover: "#60a5fa",
	PrimaryText:  "#ffffff",

	Secondary:      "#9ca3af",
	SecondaryHover: "#d1d5db",
	SecondaryText:  "#111827",

	Accent:      "#a78bfa",
	AccentHover: "#c4b5fd",
	AccentText:  "#111827",

	Success: "#4ade80",
	Warning: "#fbbf24",
	Error:   "#f87171",
	Info:    "#60a5fa",

	Border:      "#374151",
	BorderFocus: "#60a5fa",

	Shadow: "rgba(0, 0, 0, 0.3)",
}

// ThemeManager handles dark/light mode switching with CSS variables
type ThemeManager struct {
	current      ThemeMode
	lightColors  ThemeColors
	darkColors   ThemeColors
	styleElement js.Value
	subscribers  []func(ThemeMode)
}

var globalThemeManager *ThemeManager

// InitTheme initializes the global theme manager
func InitTheme() *ThemeManager {
	return InitThemeWithColors(DefaultLightColors, DefaultDarkColors)
}

// InitThemeWithColors initializes with custom colors
func InitThemeWithColors(lightColors, darkColors ThemeColors) *ThemeManager {
	if globalThemeManager != nil {
		return globalThemeManager
	}

	globalThemeManager = &ThemeManager{
		current:     ThemeSystem,
		lightColors: lightColors,
		darkColors:  darkColors,
	}

	document := js.Global().Get("document")

	// Create style element for CSS variables
	globalThemeManager.styleElement = document.Call("createElement", "style")
	globalThemeManager.styleElement.Set("id", "gux-theme")
	document.Get("head").Call("appendChild", globalThemeManager.styleElement)

	// Check for saved preference
	localStorage := js.Global().Get("localStorage")
	if !localStorage.IsUndefined() && !localStorage.IsNull() {
		saved := localStorage.Call("getItem", "gux-theme")
		if !saved.IsNull() && !saved.IsUndefined() {
			switch saved.String() {
			case "light":
				globalThemeManager.current = ThemeLight
			case "dark":
				globalThemeManager.current = ThemeDark
			default:
				globalThemeManager.current = ThemeSystem
			}
		}
	}

	globalThemeManager.apply()

	// Listen for system preference changes
	mediaQuery := js.Global().Call("matchMedia", "(prefers-color-scheme: dark)")
	mediaQuery.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) any {
		if globalThemeManager.current == ThemeSystem {
			globalThemeManager.apply()
			globalThemeManager.notify()
		}
		return nil
	}))

	return globalThemeManager
}

// GetThemeManager returns the theme manager instance
func GetThemeManager() *ThemeManager {
	if globalThemeManager == nil {
		InitTheme()
	}
	return globalThemeManager
}

// GetTheme returns the current theme mode
func GetTheme() ThemeMode {
	if globalThemeManager == nil {
		InitTheme()
	}
	return globalThemeManager.current
}

// SetTheme changes the theme
func SetTheme(theme ThemeMode) {
	if globalThemeManager == nil {
		InitTheme()
	}
	globalThemeManager.current = theme

	localStorage := js.Global().Get("localStorage")
	if !localStorage.IsUndefined() && !localStorage.IsNull() {
		localStorage.Call("setItem", "gux-theme", string(theme))
	}

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
func OnThemeChange(fn func(ThemeMode)) func() {
	if globalThemeManager == nil {
		InitTheme()
	}

	globalThemeManager.subscribers = append(globalThemeManager.subscribers, fn)
	idx := len(globalThemeManager.subscribers) - 1

	return func() {
		if idx < len(globalThemeManager.subscribers) {
			globalThemeManager.subscribers = append(
				globalThemeManager.subscribers[:idx],
				globalThemeManager.subscribers[idx+1:]...,
			)
		}
	}
}

// IsDarkMode returns true if currently in dark mode
func IsDarkMode() bool {
	if globalThemeManager == nil {
		InitTheme()
	}
	return globalThemeManager.isDark()
}

// GetActiveColors returns the currently active color palette
func GetActiveColors() ThemeColors {
	if globalThemeManager == nil {
		InitTheme()
	}
	if globalThemeManager.isDark() {
		return globalThemeManager.darkColors
	}
	return globalThemeManager.lightColors
}

func (tm *ThemeManager) isDark() bool {
	if tm.current == ThemeSystem {
		return js.Global().Call("matchMedia", "(prefers-color-scheme: dark)").Get("matches").Bool()
	}
	return tm.current == ThemeDark
}

func (tm *ThemeManager) apply() {
	colors := tm.lightColors
	if tm.isDark() {
		colors = tm.darkColors
	}

	css := `:root {
		--bg: ` + colors.Background + `;
		--bg-alt: ` + colors.BackgroundAlt + `;
		--bg-hover: ` + colors.BackgroundHover + `;
		--text: ` + colors.Text + `;
		--text-muted: ` + colors.TextMuted + `;
		--text-inverse: ` + colors.TextInverse + `;
		--primary: ` + colors.Primary + `;
		--primary-hover: ` + colors.PrimaryHover + `;
		--primary-text: ` + colors.PrimaryText + `;
		--secondary: ` + colors.Secondary + `;
		--secondary-hover: ` + colors.SecondaryHover + `;
		--secondary-text: ` + colors.SecondaryText + `;
		--accent: ` + colors.Accent + `;
		--accent-hover: ` + colors.AccentHover + `;
		--accent-text: ` + colors.AccentText + `;
		--success: ` + colors.Success + `;
		--warning: ` + colors.Warning + `;
		--error: ` + colors.Error + `;
		--info: ` + colors.Info + `;
		--border: ` + colors.Border + `;
		--border-focus: ` + colors.BorderFocus + `;
		--shadow: ` + colors.Shadow + `;
	}

	body {
		background-color: var(--bg);
		color: var(--text);
		transition: background-color 0.3s ease, color 0.3s ease;
	}

	/* Theme-aware utility classes */
	.bg-theme { background-color: var(--bg); }
	.bg-theme-alt { background-color: var(--bg-alt); }
	.bg-theme-hover:hover { background-color: var(--bg-hover); }
	.text-theme { color: var(--text); }
	.text-theme-muted { color: var(--text-muted); }
	.border-theme { border-color: var(--border); }
	.shadow-theme { box-shadow: 0 1px 3px var(--shadow); }

	/* Theme-aware buttons */
	.btn-primary-theme {
		background-color: var(--primary);
		color: var(--primary-text);
		transition: background-color 0.2s ease;
	}
	.btn-primary-theme:hover {
		background-color: var(--primary-hover);
	}

	.btn-secondary-theme {
		background-color: var(--secondary);
		color: var(--secondary-text);
		transition: background-color 0.2s ease;
	}
	.btn-secondary-theme:hover {
		background-color: var(--secondary-hover);
	}

	.btn-accent-theme {
		background-color: var(--accent);
		color: var(--accent-text);
		transition: background-color 0.2s ease;
	}
	.btn-accent-theme:hover {
		background-color: var(--accent-hover);
	}

	/* Theme-aware card */
	.card-theme {
		background-color: var(--bg);
		border: 1px solid var(--border);
		box-shadow: 0 1px 3px var(--shadow);
	}

	/* Theme-aware input */
	.input-theme {
		background-color: var(--bg);
		color: var(--text);
		border: 1px solid var(--border);
		transition: border-color 0.2s ease, box-shadow 0.2s ease;
	}
	.input-theme:focus {
		border-color: var(--border-focus);
		outline: none;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	/* Status colors */
	.text-success { color: var(--success); }
	.text-warning { color: var(--warning); }
	.text-error { color: var(--error); }
	.text-info { color: var(--info); }
	.bg-success { background-color: var(--success); }
	.bg-warning { background-color: var(--warning); }
	.bg-error { background-color: var(--error); }
	.bg-info { background-color: var(--info); }
	`

	tm.styleElement.Set("textContent", css)

	// Update document classes
	document := js.Global().Get("document")
	html := document.Get("documentElement")
	body := document.Get("body")

	html.Get("classList").Call("remove", "dark", "light")
	body.Get("classList").Call("remove", "theme-dark", "theme-light")

	if tm.isDark() {
		html.Get("classList").Call("add", "dark")
		body.Get("classList").Call("add", "theme-dark")
	} else {
		html.Get("classList").Call("add", "light")
		body.Get("classList").Call("add", "theme-light")
	}
}

func (tm *ThemeManager) notify() {
	for _, fn := range tm.subscribers {
		fn(tm.current)
	}
}

// ThemeToggleProps configures the theme toggle button
type ThemeToggleProps struct {
	ClassName string
	ShowLabel bool
}

// ThemeToggle creates a theme toggle button
func ThemeToggle(props ...ThemeToggleProps) js.Value {
	document := js.Global().Get("document")

	var p ThemeToggleProps
	if len(props) > 0 {
		p = props[0]
	}

	btn := document.Call("createElement", "button")
	className := "inline-flex items-center justify-center p-2 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors cursor-pointer"
	if p.ClassName != "" {
		className = p.ClassName
	}
	btn.Set("className", className)
	btn.Set("type", "button")
	btn.Set("title", "Toggle theme")

	updateIcon := func() {
		if IsDarkMode() {
			// Show sun icon for switching to light
			html := `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>`
			if p.ShowLabel {
				html += `<span class="ml-2">Light Mode</span>`
			}
			btn.Set("innerHTML", html)
		} else {
			// Show moon icon for switching to dark
			html := `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"></path></svg>`
			if p.ShowLabel {
				html += `<span class="ml-2">Dark Mode</span>`
			}
			btn.Set("innerHTML", html)
		}
	}

	updateIcon()

	btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		ToggleTheme()
		updateIcon()
		return nil
	}))

	// Subscribe to theme changes from elsewhere
	OnThemeChange(func(mode ThemeMode) {
		updateIcon()
	})

	return btn
}

// ThemeSelectorProps configures the theme selector dropdown
type ThemeSelectorProps struct {
	ClassName string
	Label     string
}

// ThemeSelector creates a dropdown to select theme mode
func ThemeSelector(props ...ThemeSelectorProps) js.Value {
	document := js.Global().Get("document")

	var p ThemeSelectorProps
	if len(props) > 0 {
		p = props[0]
	}
	if p.Label == "" {
		p.Label = "Theme"
	}

	container := document.Call("createElement", "div")
	container.Set("className", "relative inline-block "+p.ClassName)

	label := document.Call("createElement", "label")
	label.Set("className", "block text-sm font-medium mb-1 text-theme-muted")
	label.Set("textContent", p.Label)
	container.Call("appendChild", label)

	selectEl := document.Call("createElement", "select")
	selectEl.Set("className", "input-theme block w-full px-3 py-2 rounded-md text-sm cursor-pointer")

	options := []struct {
		value string
		label string
	}{
		{"system", "System"},
		{"light", "Light"},
		{"dark", "Dark"},
	}

	current := GetTheme()
	for _, opt := range options {
		option := document.Call("createElement", "option")
		option.Set("value", opt.value)
		option.Set("textContent", opt.label)
		if string(current) == opt.value {
			option.Set("selected", true)
		}
		selectEl.Call("appendChild", option)
	}

	selectEl.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) any {
		value := selectEl.Get("value").String()
		switch value {
		case "light":
			SetTheme(ThemeLight)
		case "dark":
			SetTheme(ThemeDark)
		default:
			SetTheme(ThemeSystem)
		}
		return nil
	}))

	container.Call("appendChild", selectEl)
	return container
}

// ThemedCard creates a card with theme-aware styling
func ThemedCard(children ...js.Value) js.Value {
	document := js.Global().Get("document")

	card := document.Call("createElement", "div")
	card.Set("className", "card-theme rounded-lg p-4")

	for _, child := range children {
		card.Call("appendChild", child)
	}

	return card
}

// ThemedButton creates a button with theme-aware styling
func ThemedButton(text string, variant string, onClick func()) js.Value {
	document := js.Global().Get("document")

	btn := document.Call("createElement", "button")

	className := "px-4 py-2 rounded-md font-medium transition-colors "
	switch variant {
	case "secondary":
		className += "btn-secondary-theme"
	case "accent":
		className += "btn-accent-theme"
	default:
		className += "btn-primary-theme"
	}

	btn.Set("className", className)
	btn.Set("type", "button")
	btn.Set("textContent", text)

	if onClick != nil {
		btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			onClick()
			return nil
		}))
	}

	return btn
}

// ThemedInput creates an input with theme-aware styling
func ThemedInput(placeholder string) js.Value {
	document := js.Global().Get("document")

	input := document.Call("createElement", "input")
	input.Set("className", "input-theme px-3 py-2 rounded-md w-full")
	input.Set("type", "text")
	input.Set("placeholder", placeholder)

	return input
}
