//go:build js && wasm

package components

import (
	"syscall/js"

	"gux/state"
)

// ConnectionStatusVariant defines the display variant
type ConnectionStatusVariant string

const (
	ConnectionStatusDotVariant   ConnectionStatusVariant = "dot"
	ConnectionStatusBadgeVariant ConnectionStatusVariant = "badge"
	ConnectionStatusTextVariant  ConnectionStatusVariant = "text"
	ConnectionStatusFullVariant  ConnectionStatusVariant = "full"
)

// ConnectionStatusSize defines the component size
type ConnectionStatusSize string

const (
	ConnectionStatusSM ConnectionStatusSize = "sm"
	ConnectionStatusMD ConnectionStatusSize = "md"
	ConnectionStatusLG ConnectionStatusSize = "lg"
)

// Default labels for each state
var defaultStatusLabels = map[state.WebSocketState]string{
	state.WSConnecting: "Connecting",
	state.WSOpen:       "Connected",
	state.WSClosing:    "Disconnecting",
	state.WSClosed:     "Disconnected",
}

// Status colors for each state
var statusDotColors = map[state.WebSocketState]string{
	state.WSConnecting: "bg-yellow-400 dark:bg-yellow-500",
	state.WSOpen:       "bg-green-500 dark:bg-green-400",
	state.WSClosing:    "bg-yellow-500 dark:bg-yellow-400",
	state.WSClosed:     "bg-red-500 dark:bg-red-400",
}

// Badge variant colors for each state
var statusBadgeColors = map[state.WebSocketState]string{
	state.WSConnecting: "bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200",
	state.WSOpen:       "bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200",
	state.WSClosing:    "bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200",
	state.WSClosed:     "bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200",
}

// Dot sizes
var dotSizes = map[ConnectionStatusSize]string{
	ConnectionStatusSM: "w-2 h-2",
	ConnectionStatusMD: "w-3 h-3",
	ConnectionStatusLG: "w-4 h-4",
}

// ConnectionStatusProps configures a ConnectionStatus component
type ConnectionStatusProps struct {
	Variant   ConnectionStatusVariant
	Size      ConnectionStatusSize
	Labels    map[state.WebSocketState]string // Custom labels for states
	ShowLabel bool                            // Show text label next to dot (for dot variant)
}

// ConnectionStatus displays WebSocket connection state
type ConnectionStatus struct {
	element     js.Value
	dotEl       js.Value
	labelEl     js.Value
	props       ConnectionStatusProps
	state       state.WebSocketState
	unsubscribe func()
}

// NewConnectionStatus creates a new ConnectionStatus component
func NewConnectionStatus(props ConnectionStatusProps) *ConnectionStatus {
	// Set defaults
	if props.Variant == "" {
		props.Variant = ConnectionStatusDotVariant
	}
	if props.Size == "" {
		props.Size = ConnectionStatusMD
	}
	if props.Labels == nil {
		props.Labels = defaultStatusLabels
	}

	cs := &ConnectionStatus{
		props: props,
		state: state.WSClosed,
	}

	cs.render()
	return cs
}

func (cs *ConnectionStatus) render() {
	document := js.Global().Get("document")

	switch cs.props.Variant {
	case ConnectionStatusDotVariant:
		cs.renderDot(document)
	case ConnectionStatusBadgeVariant:
		cs.renderBadge(document)
	case ConnectionStatusTextVariant:
		cs.renderText(document)
	case ConnectionStatusFullVariant:
		cs.renderFull(document)
	default:
		cs.renderDot(document)
	}
}

func (cs *ConnectionStatus) renderDot(document js.Value) {
	container := document.Call("createElement", "div")
	container.Set("className", "flex items-center gap-2")
	container.Set("role", "status")
	container.Set("ariaLive", "polite")

	// Create dot
	dot := document.Call("createElement", "span")
	cs.dotEl = dot
	cs.updateDotStyle()

	container.Call("appendChild", dot)

	// Add label if ShowLabel is true
	if cs.props.ShowLabel {
		label := document.Call("createElement", "span")
		label.Set("className", "text-sm text-gray-600 dark:text-gray-400")
		label.Set("textContent", cs.getLabel())
		cs.labelEl = label
		container.Call("appendChild", label)
	}

	// Add tooltip with state info
	wrapper := WithTooltip(container, TooltipProps{
		Text:     cs.getTooltipText(),
		Position: TooltipBottom,
	})

	cs.element = wrapper
}

func (cs *ConnectionStatus) renderBadge(document js.Value) {
	badge := document.Call("createElement", "span")
	badge.Set("role", "status")
	badge.Set("ariaLive", "polite")
	cs.labelEl = badge
	cs.updateBadgeStyle()

	cs.element = badge
}

func (cs *ConnectionStatus) renderText(document js.Value) {
	container := document.Call("createElement", "span")
	container.Set("role", "status")
	container.Set("ariaLive", "polite")
	container.Set("className", "text-sm font-medium")
	cs.labelEl = container
	cs.updateTextStyle()

	cs.element = container
}

func (cs *ConnectionStatus) renderFull(document js.Value) {
	container := document.Call("createElement", "div")
	container.Set("className", "flex items-center gap-2 px-3 py-2 bg-gray-100 dark:bg-gray-700 rounded-lg")
	container.Set("role", "status")
	container.Set("ariaLive", "polite")

	// Dot indicator
	dot := document.Call("createElement", "span")
	cs.dotEl = dot
	cs.updateDotStyle()
	container.Call("appendChild", dot)

	// Label
	label := document.Call("createElement", "span")
	label.Set("className", "text-sm font-medium text-gray-700 dark:text-gray-300")
	label.Set("textContent", cs.getLabel())
	cs.labelEl = label
	container.Call("appendChild", label)

	cs.element = container
}

func (cs *ConnectionStatus) updateDotStyle() {
	if cs.dotEl.IsUndefined() || cs.dotEl.IsNull() {
		return
	}

	size := dotSizes[cs.props.Size]
	color := statusDotColors[cs.state]

	className := size + " " + color + " rounded-full inline-block"

	// Add pulse animation for connecting state
	if cs.state == state.WSConnecting {
		className += " animate-pulse"
	}

	cs.dotEl.Set("className", className)
}

func (cs *ConnectionStatus) updateBadgeStyle() {
	if cs.labelEl.IsUndefined() || cs.labelEl.IsNull() {
		return
	}

	color := statusBadgeColors[cs.state]
	className := "inline-flex items-center gap-1.5 px-2.5 py-0.5 text-xs font-medium rounded-full " + color

	// Add pulse for connecting
	if cs.state == state.WSConnecting {
		className += " animate-pulse"
	}

	cs.labelEl.Set("className", className)
	cs.labelEl.Set("textContent", cs.getLabel())
}

func (cs *ConnectionStatus) updateTextStyle() {
	if cs.labelEl.IsUndefined() || cs.labelEl.IsNull() {
		return
	}

	var textColor string
	switch cs.state {
	case state.WSConnecting:
		textColor = "text-yellow-600 dark:text-yellow-400"
	case state.WSOpen:
		textColor = "text-green-600 dark:text-green-400"
	case state.WSClosing:
		textColor = "text-yellow-600 dark:text-yellow-400"
	case state.WSClosed:
		textColor = "text-red-600 dark:text-red-400"
	}

	className := "text-sm font-medium " + textColor
	if cs.state == state.WSConnecting {
		className += " animate-pulse"
	}

	cs.labelEl.Set("className", className)
	cs.labelEl.Set("textContent", cs.getLabel())
}

func (cs *ConnectionStatus) getLabel() string {
	if label, ok := cs.props.Labels[cs.state]; ok {
		return label
	}
	return defaultStatusLabels[cs.state]
}

func (cs *ConnectionStatus) getTooltipText() string {
	return "Connection status: " + cs.getLabel()
}

// Element returns the underlying DOM element
func (cs *ConnectionStatus) Element() js.Value {
	return cs.element
}

// SetState manually sets the connection state
func (cs *ConnectionStatus) SetState(newState state.WebSocketState) {
	cs.state = newState
	cs.updateDisplay()
}

// State returns the current state
func (cs *ConnectionStatus) State() state.WebSocketState {
	return cs.state
}

func (cs *ConnectionStatus) updateDisplay() {
	cs.updateDotStyle()

	// Update label if present
	if !cs.labelEl.IsUndefined() && !cs.labelEl.IsNull() {
		switch cs.props.Variant {
		case ConnectionStatusBadgeVariant:
			cs.updateBadgeStyle()
		case ConnectionStatusTextVariant:
			cs.updateTextStyle()
		default:
			cs.labelEl.Set("textContent", cs.getLabel())
		}
	}

	// Update tooltip if wrapper has tooltip
	if !cs.element.IsUndefined() && !cs.element.IsNull() {
		// Find and update tooltip text
		tooltipEl := cs.element.Call("querySelector", "[role='tooltip']")
		if !tooltipEl.IsUndefined() && !tooltipEl.IsNull() {
			tooltipEl.Set("textContent", cs.getTooltipText())
		}
	}
}

// BindToWebSocket subscribes to WebSocketStore state changes
func (cs *ConnectionStatus) BindToWebSocket(store *state.WebSocketStore) {
	// Unsubscribe from any previous binding
	if cs.unsubscribe != nil {
		cs.unsubscribe()
	}

	// Set initial state based on store
	storeState := store.State()
	if storeState.Connected {
		cs.state = state.WSOpen
	} else if storeState.Connecting {
		cs.state = state.WSConnecting
	} else {
		cs.state = state.WSClosed
	}
	cs.updateDisplay()

	// Subscribe to store changes
	cs.unsubscribe = store.Subscribe(func(s state.WSStoreState) {
		var newState state.WebSocketState
		if s.Connected {
			newState = state.WSOpen
		} else if s.Connecting {
			newState = state.WSConnecting
		} else {
			newState = state.WSClosed
		}

		if newState != cs.state {
			cs.state = newState
			cs.updateDisplay()
		}
	})
}

// Unbind removes the WebSocket subscription
func (cs *ConnectionStatus) Unbind() {
	if cs.unsubscribe != nil {
		cs.unsubscribe()
		cs.unsubscribe = nil
	}
}

// Convenience constructors

// ConnectionStatusDot creates a minimal dot indicator
func ConnectionStatusDot(size ConnectionStatusSize) *ConnectionStatus {
	return NewConnectionStatus(ConnectionStatusProps{
		Variant: ConnectionStatusDotVariant,
		Size:    size,
	})
}

// ConnectionStatusBadge creates a badge with text
func ConnectionStatusBadge() *ConnectionStatus {
	return NewConnectionStatus(ConnectionStatusProps{
		Variant: ConnectionStatusBadgeVariant,
	})
}

// ConnectionStatusText creates a text-only indicator
func ConnectionStatusText() *ConnectionStatus {
	return NewConnectionStatus(ConnectionStatusProps{
		Variant: ConnectionStatusTextVariant,
	})
}

// ConnectionStatusFull creates a full status with icon and message
func ConnectionStatusFull() *ConnectionStatus {
	return NewConnectionStatus(ConnectionStatusProps{
		Variant: ConnectionStatusFullVariant,
	})
}
