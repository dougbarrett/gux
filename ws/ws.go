//go:build js && wasm

// Package ws provides a type-safe WebSocket client for Go WASM applications.
// It mirrors the ergonomic API style of the fetch package.
package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"syscall/js"
)

// Common errors
var (
	ErrNotConnected    = errors.New("websocket not connected")
	ErrAlreadyConnected = errors.New("websocket already connected")
	ErrConnectionFailed = errors.New("websocket connection failed")
	ErrSendFailed      = errors.New("failed to send message")
)

// State represents WebSocket connection state
type State int

const (
	StateConnecting State = iota
	StateOpen
	StateClosing
	StateClosed
)

// Message represents a typed WebSocket message
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	ID      string          `json:"id,omitempty"` // For request/response correlation
}

// Client is a type-safe WebSocket client
type Client struct {
	url        string
	ws         js.Value
	state      State
	mu         sync.RWMutex
	handlers   map[string][]func(json.RawMessage)
	handlersMu sync.RWMutex

	// Pending requests waiting for responses
	pendingReqs   map[string]chan Message
	pendingReqsMu sync.RWMutex

	// Callbacks
	onOpen    func()
	onClose   func(code int, reason string)
	onError   func(err error)
	onMessage func(Message)

	// For cleanup
	openFunc    js.Func
	closeFunc   js.Func
	errorFunc   js.Func
	messageFunc js.Func
}

// Option configures a Client
type Option func(*Client)

// WithOnOpen sets the connection open callback
func WithOnOpen(fn func()) Option {
	return func(c *Client) {
		c.onOpen = fn
	}
}

// WithOnClose sets the connection close callback
func WithOnClose(fn func(code int, reason string)) Option {
	return func(c *Client) {
		c.onClose = fn
	}
}

// WithOnError sets the error callback
func WithOnError(fn func(err error)) Option {
	return func(c *Client) {
		c.onError = fn
	}
}

// WithOnMessage sets the raw message callback
func WithOnMessage(fn func(Message)) Option {
	return func(c *Client) {
		c.onMessage = fn
	}
}

// NewClient creates a new WebSocket client
func NewClient(url string, opts ...Option) *Client {
	c := &Client{
		url:         url,
		state:       StateClosed,
		handlers:    make(map[string][]func(json.RawMessage)),
		pendingReqs: make(map[string]chan Message),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Connect establishes the WebSocket connection
func (c *Client) Connect() error {
	c.mu.Lock()
	if c.state == StateOpen || c.state == StateConnecting {
		c.mu.Unlock()
		return ErrAlreadyConnected
	}
	c.state = StateConnecting
	c.mu.Unlock()

	done := make(chan error, 1)

	// Create WebSocket
	c.ws = js.Global().Get("WebSocket").New(c.url)

	// Setup event handlers
	c.openFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		c.mu.Lock()
		c.state = StateOpen
		c.mu.Unlock()

		if c.onOpen != nil {
			c.onOpen()
		}
		done <- nil
		return nil
	})

	c.closeFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		c.mu.Lock()
		c.state = StateClosed
		c.mu.Unlock()

		code := 1000
		reason := ""
		if len(args) > 0 {
			code = args[0].Get("code").Int()
			reason = args[0].Get("reason").String()
		}

		if c.onClose != nil {
			c.onClose(code, reason)
		}
		return nil
	})

	c.errorFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		err := ErrConnectionFailed
		if len(args) > 0 && args[0].Get("message").Truthy() {
			err = errors.New(args[0].Get("message").String())
		}

		c.mu.Lock()
		if c.state == StateConnecting {
			c.state = StateClosed
			done <- err
		}
		c.mu.Unlock()

		if c.onError != nil {
			c.onError(err)
		}
		return nil
	})

	c.messageFunc = js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}

		data := args[0].Get("data").String()

		var msg Message
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			// Try to handle as raw message
			if c.onMessage != nil {
				c.onMessage(Message{Payload: json.RawMessage(data)})
			}
			return nil
		}

		// Check for pending request/response correlation
		if msg.ID != "" {
			c.pendingReqsMu.RLock()
			ch, ok := c.pendingReqs[msg.ID]
			c.pendingReqsMu.RUnlock()
			if ok {
				ch <- msg
				return nil
			}
		}

		// Call type-specific handlers
		c.handlersMu.RLock()
		handlers := c.handlers[msg.Type]
		c.handlersMu.RUnlock()

		for _, handler := range handlers {
			handler(msg.Payload)
		}

		// Call generic message handler
		if c.onMessage != nil {
			c.onMessage(msg)
		}

		return nil
	})

	c.ws.Set("onopen", c.openFunc)
	c.ws.Set("onclose", c.closeFunc)
	c.ws.Set("onerror", c.errorFunc)
	c.ws.Set("onmessage", c.messageFunc)

	// Wait for connection or error
	return <-done
}

// Close closes the WebSocket connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != StateOpen {
		return ErrNotConnected
	}

	c.state = StateClosing
	c.ws.Call("close")

	// Cleanup JS functions
	c.openFunc.Release()
	c.closeFunc.Release()
	c.errorFunc.Release()
	c.messageFunc.Release()

	return nil
}

// State returns the current connection state
func (c *Client) State() State {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// IsConnected returns true if the WebSocket is connected
func (c *Client) IsConnected() bool {
	return c.State() == StateOpen
}

// Send sends a typed message over the WebSocket
func (c *Client) Send(msgType string, payload any) error {
	c.mu.RLock()
	if c.state != StateOpen {
		c.mu.RUnlock()
		return ErrNotConnected
	}
	c.mu.RUnlock()

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	msg := Message{
		Type:    msgType,
		Payload: payloadBytes,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	c.ws.Call("send", string(msgBytes))
	return nil
}

// SendRaw sends a raw string message
func (c *Client) SendRaw(data string) error {
	c.mu.RLock()
	if c.state != StateOpen {
		c.mu.RUnlock()
		return ErrNotConnected
	}
	c.mu.RUnlock()

	c.ws.Call("send", data)
	return nil
}

// On registers a handler for a specific message type
func (c *Client) On(msgType string, handler func(json.RawMessage)) {
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()
	c.handlers[msgType] = append(c.handlers[msgType], handler)
}

// OnTyped registers a typed handler for a specific message type
func OnTyped[T any](c *Client, msgType string, handler func(T)) {
	c.On(msgType, func(data json.RawMessage) {
		var payload T
		if err := json.Unmarshal(data, &payload); err != nil {
			return
		}
		handler(payload)
	})
}

// Request sends a message and waits for a response with matching ID
func (c *Client) Request(msgType string, payload any) (json.RawMessage, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	// Generate a unique ID using timestamp + random component
	id := fmt.Sprintf("%d-%d", js.Global().Get("Date").Call("now").Int(), int(js.Global().Get("Math").Call("random").Float()*1000000))

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	msg := Message{
		Type:    msgType,
		Payload: payloadBytes,
		ID:      id,
	}

	// Setup response channel
	respCh := make(chan Message, 1)

	// Register pending request
	c.pendingReqsMu.Lock()
	c.pendingReqs[id] = respCh
	c.pendingReqsMu.Unlock()

	// Ensure cleanup
	defer func() {
		c.pendingReqsMu.Lock()
		delete(c.pendingReqs, id)
		c.pendingReqsMu.Unlock()
	}()

	// Send message
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal message: %w", err)
	}

	c.ws.Call("send", string(msgBytes))

	// Wait for response
	resp := <-respCh

	// Check if it's an error response
	if resp.Type == "error" {
		var errMsg struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(resp.Payload, &errMsg); err == nil {
			return nil, errors.New(errMsg.Message)
		}
		return nil, errors.New("unknown error")
	}

	return resp.Payload, nil
}

// RequestTyped sends a message and returns a typed response
func RequestTyped[Req any, Resp any](c *Client, msgType string, req Req) (*Resp, error) {
	respData, err := c.Request(msgType, req)
	if err != nil {
		return nil, err
	}

	var resp Resp
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &resp, nil
}
