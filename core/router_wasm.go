//go:build js && wasm

package core

import (
	"syscall/js"
)

var (
	rerenderScheduled = false
	scheduledRouter   *Router
)

// ScheduleRerender schedules a re-render on the next animation frame.
// This batches multiple state changes into a single re-render and prevents
// re-renders from interrupting event handlers (like click handlers).
func ScheduleRerender(r *Router) {
	if rerenderScheduled {
		return // Already scheduled
	}
	rerenderScheduled = true
	scheduledRouter = r

	// Use requestAnimationFrame to defer re-render to next frame
	js.Global().Call("requestAnimationFrame", js.FuncOf(func(this js.Value, args []js.Value) any {
		rerenderScheduled = false
		if scheduledRouter != nil && scheduledRouter.rerender != nil {
			scheduledRouter.rerender()
		}
		return nil
	}))
}
