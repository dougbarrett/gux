//go:build js && wasm

package core

import (
	"fmt"
	"syscall/js"
)

var (
	rerenderScheduled = false
	scheduledRouter   *Router
	inChangeEvent     = false // When true, ScheduleRerender is suppressed
)

func init() {
	// Enable state debug logging in WASM mode
	StateDebugLog = func(action, key string, value any, router *Router) {
		// Get router address for debugging
		routerAddr := fmt.Sprintf("%p", router)
		stateMapLen := len(router.state)
		js.Global().Get("console").Call("log",
			fmt.Sprintf("[Gux State] %s key=%q value=%v router=%s state_size=%d",
				action, key, value, routerAddr, stateMapLen))
	}
}

// SetInChangeEvent sets the change event flag.
// When true, ScheduleRerender calls are ignored to prevent
// re-renders from interrupting click events.
func SetInChangeEvent(v bool) {
	inChangeEvent = v
}

// ScheduleRerender schedules a re-render after the current event loop completes.
// This batches multiple state changes into a single re-render and prevents
// re-renders from interrupting event handlers (like click handlers).
//
// When inChangeEvent is true (during input change/blur events), re-renders are
// completely suppressed. This prevents the DOM from being replaced while a
// button click is being processed. The state is still updated, so when the
// click handler runs, it reads the correct values.
func ScheduleRerender(r *Router) {
	// Suppress re-renders during change events to allow button clicks to work
	if inChangeEvent {
		return
	}

	if rerenderScheduled {
		return // Already scheduled
	}
	rerenderScheduled = true
	scheduledRouter = r

	// Use setTimeout(0) to defer re-render to after current event loop completes.
	js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
		rerenderScheduled = false
		if scheduledRouter != nil && scheduledRouter.rerender != nil {
			scheduledRouter.rerender()
		}
		return nil
	}), 0)
}
