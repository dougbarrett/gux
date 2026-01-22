package pages

// Router provides page context and state management.
// In the future, this will be generated with per-target implementations.
type Router struct {
	state    map[string]any
	rerender func()
}

// NewRouter creates a new router instance.
func NewRouter(rerender func()) *Router {
	return &Router{
		state:    make(map[string]any),
		rerender: rerender,
	}
}

// State returns a reactive state value.
// The key ensures stable identity across re-renders.
type State[T any] struct {
	key      string
	router   *Router
	initial  T
}

func (r *Router) State(key string, initial int) *State[int] {
	if _, ok := r.state[key]; !ok {
		r.state[key] = initial
	}
	return &State[int]{key: key, router: r, initial: initial}
}

func (s *State[T]) Get() T {
	if val, ok := s.router.state[s.key]; ok {
		return val.(T)
	}
	return s.initial
}

func (s *State[T]) Set(val T) {
	s.router.state[s.key] = val
	if s.router.rerender != nil {
		s.router.rerender()
	}
}
