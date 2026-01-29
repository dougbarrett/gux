//go:build !js || !wasm

package core

// ScheduleRerender on server-side just calls rerender immediately.
// Server-side rendering doesn't need deferred re-renders since it's a single pass.
func ScheduleRerender(r *Router) {
	if r.rerender != nil {
		r.rerender()
	}
}

// wasmQuery is a no-op on server — query params are read from r.request.
func wasmQuery(name string) string {
	return ""
}
