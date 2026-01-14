//go:build js && wasm

package state

import (
	"encoding/json"
	"syscall/js"
)

// Storage provides a typed interface to localStorage/sessionStorage
type Storage struct {
	storage js.Value
}

// LocalStorage returns a Storage instance backed by localStorage
func LocalStorage() *Storage {
	return &Storage{
		storage: js.Global().Get("localStorage"),
	}
}

// SessionStorage returns a Storage instance backed by sessionStorage
func SessionStorage() *Storage {
	return &Storage{
		storage: js.Global().Get("sessionStorage"),
	}
}

// Get retrieves a string value from storage
func (s *Storage) Get(key string) string {
	val := s.storage.Call("getItem", key)
	if val.IsNull() || val.IsUndefined() {
		return ""
	}
	return val.String()
}

// Set stores a string value
func (s *Storage) Set(key, value string) {
	s.storage.Call("setItem", key, value)
}

// GetJSON retrieves and unmarshals a JSON value
func (s *Storage) GetJSON(key string, target any) error {
	val := s.Get(key)
	if val == "" {
		return nil
	}
	return json.Unmarshal([]byte(val), target)
}

// SetJSON marshals and stores a value as JSON
func (s *Storage) SetJSON(key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.Set(key, string(data))
	return nil
}

// Remove deletes a key from storage
func (s *Storage) Remove(key string) {
	s.storage.Call("removeItem", key)
}

// Clear removes all keys from storage
func (s *Storage) Clear() {
	s.storage.Call("clear")
}

// Has checks if a key exists
func (s *Storage) Has(key string) bool {
	val := s.storage.Call("getItem", key)
	return !val.IsNull() && !val.IsUndefined()
}

// Keys returns all keys in storage
func (s *Storage) Keys() []string {
	length := s.storage.Get("length").Int()
	keys := make([]string, length)
	for i := 0; i < length; i++ {
		keys[i] = s.storage.Call("key", i).String()
	}
	return keys
}

// Length returns the number of items in storage
func (s *Storage) Length() int {
	return s.storage.Get("length").Int()
}

// PersistentStore wraps a Store with localStorage persistence
type PersistentStore[T any] struct {
	*Store[T]
	key     string
	storage *Storage
}

// NewPersistentStore creates a store that persists to localStorage
func NewPersistentStore[T any](key string, initial T) *PersistentStore[T] {
	storage := LocalStorage()
	store := New(initial)

	ps := &PersistentStore[T]{
		Store:   store,
		key:     key,
		storage: storage,
	}

	// Load from storage
	var saved T
	if err := storage.GetJSON(key, &saved); err == nil && storage.Has(key) {
		store.Set(saved)
	}

	// Subscribe to changes and persist
	store.Subscribe(func(value T) {
		storage.SetJSON(key, value)
	})

	return ps
}

// Reset clears storage and resets to initial value
func (ps *PersistentStore[T]) Reset(initial T) {
	ps.storage.Remove(ps.key)
	ps.Store.Set(initial)
}

// SessionStore wraps a Store with sessionStorage persistence
type SessionStore[T any] struct {
	*Store[T]
	key     string
	storage *Storage
}

// NewSessionStore creates a store that persists to sessionStorage
func NewSessionStore[T any](key string, initial T) *SessionStore[T] {
	storage := SessionStorage()
	store := New(initial)

	ss := &SessionStore[T]{
		Store:   store,
		key:     key,
		storage: storage,
	}

	// Load from storage
	var saved T
	if err := storage.GetJSON(key, &saved); err == nil && storage.Has(key) {
		store.Set(saved)
	}

	// Subscribe to changes and persist
	store.Subscribe(func(value T) {
		storage.SetJSON(key, value)
	})

	return ss
}

// Convenience functions for common operations

// GetLocalString gets a string from localStorage
func GetLocalString(key string) string {
	return LocalStorage().Get(key)
}

// SetLocalString sets a string in localStorage
func SetLocalString(key, value string) {
	LocalStorage().Set(key, value)
}

// GetLocalJSON gets a JSON value from localStorage
func GetLocalJSON[T any](key string) (T, bool) {
	var result T
	storage := LocalStorage()
	if !storage.Has(key) {
		return result, false
	}
	err := storage.GetJSON(key, &result)
	return result, err == nil
}

// SetLocalJSON sets a JSON value in localStorage
func SetLocalJSON[T any](key string, value T) error {
	return LocalStorage().SetJSON(key, value)
}

// RemoveLocal removes a key from localStorage
func RemoveLocal(key string) {
	LocalStorage().Remove(key)
}

// GetSessionString gets a string from sessionStorage
func GetSessionString(key string) string {
	return SessionStorage().Get(key)
}

// SetSessionString sets a string in sessionStorage
func SetSessionString(key, value string) {
	SessionStorage().Set(key, value)
}

// GetSessionJSON gets a JSON value from sessionStorage
func GetSessionJSON[T any](key string) (T, bool) {
	var result T
	storage := SessionStorage()
	if !storage.Has(key) {
		return result, false
	}
	err := storage.GetJSON(key, &result)
	return result, err == nil
}

// SetSessionJSON sets a JSON value in sessionStorage
func SetSessionJSON[T any](key string, value T) error {
	return SessionStorage().SetJSON(key, value)
}

// RemoveSession removes a key from sessionStorage
func RemoveSession(key string) {
	SessionStorage().Remove(key)
}
