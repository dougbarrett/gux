//go:build js && wasm

package core

import (
	"fmt"
	"net/url"
	"strings"
	"syscall/js"
)

var (
	rerenderScheduled = false
	scheduledRouter   *Router
	inChangeEvent     = false // When true, ScheduleRerender is suppressed
)

// debugRouter holds reference to the current router for debugging
var debugRouter *Router

// SetDebugRouter sets the router for debugging (called from generated WASM code)
func SetDebugRouter(r *Router) {
	println("[Gux] SetDebugRouter called")
	debugRouter = r
	// Expose debug function to JavaScript console
	js.Global().Set("__guxDebugState", js.FuncOf(func(this js.Value, args []js.Value) any {
		if debugRouter == nil {
			println("[Gux Debug] No router set")
			return nil
		}
		println("[Gux Debug] Router:", fmt.Sprintf("%p", debugRouter))
		println("[Gux Debug] State keys:")
		for k, v := range debugRouter.state {
			println("  ", k, "=", fmt.Sprintf("%v", v))
		}
		return nil
	}))
	println("[Gux] Debug enabled - call __guxDebugState() to inspect state")
}

func init() {
	println("[Gux] router_wasm.go init() called")
	// Enable state debug logging in WASM mode
	StateDebugLog = func(action, key string, value any, router *Router) {
		// Get router address for debugging - use println for TinyGo compatibility
		println(fmt.Sprintf("[Gux State] %s key=%q value=%v router=%p state_size=%d",
			action, key, value, router, len(router.state)))
	}
}

// wasmQuery reads a query parameter from the browser's window.location.search.
func wasmQuery(name string) string {
	search := js.Global().Get("location").Get("search").String()
	if search == "" {
		return ""
	}
	vals, err := url.ParseQuery(strings.TrimPrefix(search, "?"))
	if err != nil {
		return ""
	}
	return vals.Get(name)
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
