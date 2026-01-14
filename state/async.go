//go:build js && wasm

package state

// AsyncState represents loading state for async operations
type AsyncState[T any] struct {
	Data    T
	Loading bool
	Error   error
}

// AsyncStore is a store specialized for async data
type AsyncStore[T any] struct {
	*Store[AsyncState[T]]
}

// NewAsync creates a new async store
func NewAsync[T any]() *AsyncStore[T] {
	return &AsyncStore[T]{
		Store: New(AsyncState[T]{}),
	}
}

// NewAsyncWithDefault creates a new async store with default data
func NewAsyncWithDefault[T any](defaultData T) *AsyncStore[T] {
	return &AsyncStore[T]{
		Store: New(AsyncState[T]{Data: defaultData}),
	}
}

// Load starts a loading operation
func (s *AsyncStore[T]) Load(fn func() (T, error)) {
	s.Update(func(state *AsyncState[T]) {
		state.Loading = true
		state.Error = nil
	})

	go func() {
		data, err := fn()
		s.Update(func(state *AsyncState[T]) {
			state.Loading = false
			if err != nil {
				state.Error = err
			} else {
				state.Data = data
			}
		})
	}()
}

// SetData sets the data without loading state
func (s *AsyncStore[T]) SetData(data T) {
	s.Update(func(state *AsyncState[T]) {
		state.Data = data
		state.Loading = false
		state.Error = nil
	})
}

// SetError sets an error state
func (s *AsyncStore[T]) SetError(err error) {
	s.Update(func(state *AsyncState[T]) {
		state.Loading = false
		state.Error = err
	})
}

// IsLoading returns true if currently loading
func (s *AsyncStore[T]) IsLoading() bool {
	return s.Get().Loading
}

// HasError returns true if there's an error
func (s *AsyncStore[T]) HasError() bool {
	return s.Get().Error != nil
}

// Data returns the current data
func (s *AsyncStore[T]) Data() T {
	return s.Get().Data
}

// Err returns the current error
func (s *AsyncStore[T]) Err() error {
	return s.Get().Error
}
