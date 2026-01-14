//go:build js && wasm

package state

import "sync"

// Store is a generic reactive state container
type Store[T any] struct {
	mu          sync.RWMutex
	state       T
	subscribers []func(T)
	nextID      int
	subIDs      map[int]int // maps subscription ID to index
}

// New creates a new store with initial state
func New[T any](initial T) *Store[T] {
	return &Store[T]{
		state:  initial,
		subIDs: make(map[int]int),
	}
}

// Get returns the current state (read-only snapshot)
func (s *Store[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// Set replaces the entire state
func (s *Store[T]) Set(newState T) {
	s.mu.Lock()
	s.state = newState
	subs := make([]func(T), len(s.subscribers))
	copy(subs, s.subscribers)
	s.mu.Unlock()

	for _, sub := range subs {
		sub(newState)
	}
}

// Update applies a mutation function to the state
func (s *Store[T]) Update(fn func(*T)) {
	s.mu.Lock()
	fn(&s.state)
	newState := s.state
	subs := make([]func(T), len(s.subscribers))
	copy(subs, s.subscribers)
	s.mu.Unlock()

	for _, sub := range subs {
		sub(newState)
	}
}

// Subscribe registers a callback for state changes, returns unsubscribe function
func (s *Store[T]) Subscribe(fn func(T)) func() {
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	s.subIDs[id] = len(s.subscribers)
	s.subscribers = append(s.subscribers, fn)
	s.mu.Unlock()

	// Return unsubscribe function
	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		idx, ok := s.subIDs[id]
		if !ok {
			return
		}

		// Remove subscriber
		s.subscribers = append(s.subscribers[:idx], s.subscribers[idx+1:]...)
		delete(s.subIDs, id)

		// Update indices for remaining subscribers
		for subID, subIdx := range s.subIDs {
			if subIdx > idx {
				s.subIDs[subID] = subIdx - 1
			}
		}
	}
}

// Derived creates a derived store that transforms the parent state
func Derived[T, U any](parent *Store[T], transform func(T) U) *Store[U] {
	derived := New(transform(parent.Get()))

	parent.Subscribe(func(state T) {
		derived.Set(transform(state))
	})

	return derived
}
