//go:build js && wasm

package storage

import (
	"encoding/json"
	"syscall/js"
)

// Local provides access to browser localStorage
var Local = &localStorage{}

type localStorage struct{}

// Set stores a string value
func (l *localStorage) Set(key, value string) {
	js.Global().Get("localStorage").Call("setItem", key, value)
}

// Get retrieves a string value
func (l *localStorage) Get(key string) string {
	val := js.Global().Get("localStorage").Call("getItem", key)
	if val.IsNull() || val.IsUndefined() {
		return ""
	}
	return val.String()
}

// Remove deletes a key
func (l *localStorage) Remove(key string) {
	js.Global().Get("localStorage").Call("removeItem", key)
}

// Clear removes all keys
func (l *localStorage) Clear() {
	js.Global().Get("localStorage").Call("clear")
}

// SetJSON stores a JSON-serializable value
func (l *localStorage) SetJSON(key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	l.Set(key, string(data))
	return nil
}

// GetJSON retrieves and unmarshals a JSON value
func (l *localStorage) GetJSON(key string, dest any) error {
	val := l.Get(key)
	if val == "" {
		return nil
	}
	return json.Unmarshal([]byte(val), dest)
}

// Session provides access to browser sessionStorage
var Session = &sessionStorage{}

type sessionStorage struct{}

// Set stores a string value
func (s *sessionStorage) Set(key, value string) {
	js.Global().Get("sessionStorage").Call("setItem", key, value)
}

// Get retrieves a string value
func (s *sessionStorage) Get(key string) string {
	val := js.Global().Get("sessionStorage").Call("getItem", key)
	if val.IsNull() || val.IsUndefined() {
		return ""
	}
	return val.String()
}

// Remove deletes a key
func (s *sessionStorage) Remove(key string) {
	js.Global().Get("sessionStorage").Call("removeItem", key)
}

// Clear removes all keys
func (s *sessionStorage) Clear() {
	js.Global().Get("sessionStorage").Call("clear")
}

// SetJSON stores a JSON-serializable value
func (s *sessionStorage) SetJSON(key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.Set(key, string(data))
	return nil
}

// GetJSON retrieves and unmarshals a JSON value
func (s *sessionStorage) GetJSON(key string, dest any) error {
	val := s.Get(key)
	if val == "" {
		return nil
	}
	return json.Unmarshal([]byte(val), dest)
}
