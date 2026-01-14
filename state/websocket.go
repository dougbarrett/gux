//go:build js && wasm

package state

import (
	"encoding/json"
	"syscall/js"
	"time"
)

// WebSocketState represents the connection state
type WebSocketState int

const (
	WSConnecting WebSocketState = iota
	WSOpen
	WSClosing
	WSClosed
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// WebSocketConfig configures the WebSocket connection
type WebSocketConfig struct {
	URL               string
	Protocols         []string
	ReconnectInterval time.Duration
	MaxReconnects     int
	OnOpen            func()
	OnClose           func(code int, reason string)
	OnError           func(err string)
	OnMessage         func(data []byte)
	OnStateChange     func(state WebSocketState)
}

// WebSocket wraps the JavaScript WebSocket API
type WebSocket struct {
	config       WebSocketConfig
	ws           js.Value
	state        WebSocketState
	reconnects   int
	handlers     map[string][]func([]byte)
	shouldClose  bool
}

// NewWebSocket creates a new WebSocket connection
func NewWebSocket(config WebSocketConfig) *WebSocket {
	if config.ReconnectInterval == 0 {
		config.ReconnectInterval = 3 * time.Second
	}
	if config.MaxReconnects == 0 {
		config.MaxReconnects = 5
	}

	ws := &WebSocket{
		config:   config,
		state:    WSClosed,
		handlers: make(map[string][]func([]byte)),
	}

	return ws
}

// Connect establishes the WebSocket connection
func (w *WebSocket) Connect() {
	w.shouldClose = false
	w.connect()
}

func (w *WebSocket) connect() {
	if w.shouldClose {
		return
	}

	w.setState(WSConnecting)

	var ws js.Value
	if len(w.config.Protocols) > 0 {
		protocols := make([]any, len(w.config.Protocols))
		for i, p := range w.config.Protocols {
			protocols[i] = p
		}
		ws = js.Global().Get("WebSocket").New(w.config.URL, protocols)
	} else {
		ws = js.Global().Get("WebSocket").New(w.config.URL)
	}

	w.ws = ws

	// onopen
	ws.Set("onopen", js.FuncOf(func(this js.Value, args []js.Value) any {
		w.setState(WSOpen)
		w.reconnects = 0
		if w.config.OnOpen != nil {
			w.config.OnOpen()
		}
		return nil
	}))

	// onclose
	ws.Set("onclose", js.FuncOf(func(this js.Value, args []js.Value) any {
		event := args[0]
		code := event.Get("code").Int()
		reason := event.Get("reason").String()

		w.setState(WSClosed)

		if w.config.OnClose != nil {
			w.config.OnClose(code, reason)
		}

		// Auto reconnect if not intentionally closed
		if !w.shouldClose && w.reconnects < w.config.MaxReconnects {
			w.reconnects++
			time.AfterFunc(w.config.ReconnectInterval, func() {
				w.connect()
			})
		}

		return nil
	}))

	// onerror
	ws.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) any {
		if w.config.OnError != nil {
			w.config.OnError("WebSocket error")
		}
		return nil
	}))

	// onmessage
	ws.Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) any {
		event := args[0]
		data := event.Get("data").String()

		if w.config.OnMessage != nil {
			w.config.OnMessage([]byte(data))
		}

		// Try to parse as typed message and dispatch to handlers
		var msg WSMessage
		if err := json.Unmarshal([]byte(data), &msg); err == nil && msg.Type != "" {
			if handlers, ok := w.handlers[msg.Type]; ok {
				for _, h := range handlers {
					h(msg.Data)
				}
			}
		}

		return nil
	}))
}

func (w *WebSocket) setState(state WebSocketState) {
	w.state = state
	if w.config.OnStateChange != nil {
		w.config.OnStateChange(state)
	}
}

// Close closes the WebSocket connection
func (w *WebSocket) Close() {
	w.shouldClose = true
	if !w.ws.IsUndefined() && !w.ws.IsNull() {
		w.setState(WSClosing)
		w.ws.Call("close")
	}
}

// CloseWithCode closes with a specific code and reason
func (w *WebSocket) CloseWithCode(code int, reason string) {
	w.shouldClose = true
	if !w.ws.IsUndefined() && !w.ws.IsNull() {
		w.setState(WSClosing)
		w.ws.Call("close", code, reason)
	}
}

// Send sends raw data
func (w *WebSocket) Send(data []byte) error {
	if w.state != WSOpen {
		return nil // Silently fail if not connected
	}
	w.ws.Call("send", string(data))
	return nil
}

// SendText sends a text message
func (w *WebSocket) SendText(text string) error {
	return w.Send([]byte(text))
}

// SendJSON sends a JSON-encoded message
func (w *WebSocket) SendJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return w.Send(data)
}

// SendTyped sends a typed message
func (w *WebSocket) SendTyped(msgType string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	msg := WSMessage{
		Type: msgType,
		Data: payload,
	}
	return w.SendJSON(msg)
}

// On registers a handler for a specific message type
func (w *WebSocket) On(msgType string, handler func([]byte)) {
	w.handlers[msgType] = append(w.handlers[msgType], handler)
}

// Off removes all handlers for a message type
func (w *WebSocket) Off(msgType string) {
	delete(w.handlers, msgType)
}

// State returns the current connection state
func (w *WebSocket) State() WebSocketState {
	return w.state
}

// IsConnected returns true if the WebSocket is open
func (w *WebSocket) IsConnected() bool {
	return w.state == WSOpen
}

// Reconnect forces a reconnection
func (w *WebSocket) Reconnect() {
	w.Close()
	w.reconnects = 0
	w.Connect()
}

// WSStoreState represents the WebSocket store state
type WSStoreState struct {
	Connected    bool
	Connecting   bool
	LastMessage  string
	MessageCount int
	Error        string
}

// WebSocketStore combines WebSocket with reactive state
type WebSocketStore struct {
	ws       *WebSocket
	store    *Store[WSStoreState]
	messages []string
}

// NewWebSocketStore creates a WebSocket with integrated state management
func NewWebSocketStore(config WebSocketConfig) *WebSocketStore {
	store := New(WSStoreState{
		Connected:    false,
		Connecting:   false,
		LastMessage:  "",
		MessageCount: 0,
		Error:        "",
	})

	wss := &WebSocketStore{
		store:    store,
		messages: make([]string, 0),
	}

	// Wrap callbacks to update store
	originalOnOpen := config.OnOpen
	config.OnOpen = func() {
		store.Update(func(s *WSStoreState) {
			s.Connected = true
			s.Connecting = false
			s.Error = ""
		})
		if originalOnOpen != nil {
			originalOnOpen()
		}
	}

	originalOnClose := config.OnClose
	config.OnClose = func(code int, reason string) {
		store.Update(func(s *WSStoreState) {
			s.Connected = false
			s.Connecting = false
		})
		if originalOnClose != nil {
			originalOnClose(code, reason)
		}
	}

	originalOnError := config.OnError
	config.OnError = func(err string) {
		store.Update(func(s *WSStoreState) {
			s.Error = err
		})
		if originalOnError != nil {
			originalOnError(err)
		}
	}

	originalOnMessage := config.OnMessage
	config.OnMessage = func(data []byte) {
		msg := string(data)

		// Keep last 100 messages
		if len(wss.messages) >= 100 {
			wss.messages = wss.messages[1:]
		}
		wss.messages = append(wss.messages, msg)

		store.Update(func(s *WSStoreState) {
			s.MessageCount++
			s.LastMessage = msg
		})

		if originalOnMessage != nil {
			originalOnMessage(data)
		}
	}

	config.OnStateChange = func(state WebSocketState) {
		store.Update(func(s *WSStoreState) {
			s.Connecting = state == WSConnecting
		})
	}

	wss.ws = NewWebSocket(config)
	return wss
}

// Connect connects the WebSocket
func (wss *WebSocketStore) Connect() {
	wss.store.Update(func(s *WSStoreState) {
		s.Connecting = true
	})
	wss.ws.Connect()
}

// Close closes the WebSocket
func (wss *WebSocketStore) Close() {
	wss.ws.Close()
}

// Send sends data through the WebSocket
func (wss *WebSocketStore) Send(data []byte) error {
	return wss.ws.Send(data)
}

// SendJSON sends JSON data
func (wss *WebSocketStore) SendJSON(v any) error {
	return wss.ws.SendJSON(v)
}

// SendTyped sends a typed message
func (wss *WebSocketStore) SendTyped(msgType string, data any) error {
	return wss.ws.SendTyped(msgType, data)
}

// On registers a message handler
func (wss *WebSocketStore) On(msgType string, handler func([]byte)) {
	wss.ws.On(msgType, handler)
}

// Store returns the underlying state store
func (wss *WebSocketStore) Store() *Store[WSStoreState] {
	return wss.store
}

// State returns the current store state
func (wss *WebSocketStore) State() WSStoreState {
	return wss.store.Get()
}

// Subscribe subscribes to store changes
func (wss *WebSocketStore) Subscribe(fn func(WSStoreState)) func() {
	return wss.store.Subscribe(fn)
}

// IsConnected returns connection status
func (wss *WebSocketStore) IsConnected() bool {
	return wss.ws.IsConnected()
}

// Messages returns the message history
func (wss *WebSocketStore) Messages() []string {
	return wss.messages
}

// ClearMessages clears the message history
func (wss *WebSocketStore) ClearMessages() {
	wss.messages = make([]string, 0)
	wss.store.Update(func(s *WSStoreState) {
		s.MessageCount = 0
	})
}

// UseWebSocket is a convenience function for creating a WebSocket connection
func UseWebSocket(url string, onMessage func([]byte)) *WebSocket {
	return NewWebSocket(WebSocketConfig{
		URL:       url,
		OnMessage: onMessage,
	})
}

// UseRealtimeStore creates a WebSocket store for real-time data
func UseRealtimeStore(url string) *WebSocketStore {
	return NewWebSocketStore(WebSocketConfig{URL: url})
}
