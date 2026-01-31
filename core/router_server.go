//go:build !js || !wasm

package core

// ScheduleRerender on server-side just calls rerender immediately.
// Server-side rendering doesn't need deferred re-renders since it's a single pass.
func ScheduleRerender(r *Router) {
	if r.rerender != nil {
		r.rerender()
	}
}

// SetInterval is a no-op on server — timers only run in WASM.
func (r *Router) SetInterval(callback func(), ms int) *TimerHandle {
	return &TimerHandle{cleared: true}
}

// SetTimeout is a no-op on server — timers only run in WASM.
func (r *Router) SetTimeout(callback func(), ms int) *TimerHandle {
	return &TimerHandle{cleared: true}
}

func clearTimer(h *TimerHandle) {}

// wasmQuery is a no-op on server — query params are read from r.request.
func wasmQuery(name string) string {
	return ""
}
