//go:build js && wasm

package core

import (
	"fmt"
	"net/url"
	"strings"
	"syscall/js"
)

var (
	rerenderScheduled     = false
	scheduledRouter       *Router
	inChangeEvent         = false  // When true, ScheduleRerender is deferred
	changeEventSuppressed *Router // Router that was suppressed during change event
	inPointerDown         = false  // When true, re-renders are deferred until mouseup/touchend
	pointerDownSuppressed *Router // Router that was suppressed during pointer-down window
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

	// Suppress re-renders during the mousedown→click window (#112).
	//
	// When a user clicks a button after typing in a form field, the browser
	// fires: mousedown → blur → change → (gap) → mouseup → click.
	// The change handler updates state and schedules a re-render via
	// setTimeout(0). That setTimeout fires during the gap, replacing the
	// entire DOM (innerHTML=""), which destroys the original button element.
	// The browser then fires mouseup on a NEW button (different DOM node)
	// and does NOT fire click because the mousedown target was destroyed.
	//
	// By tracking mousedown/touchstart, we defer re-renders until after
	// mouseup/touchend. The deferred render's setTimeout(0) then fires
	// AFTER click (since mouseup and click are in the same macrotask),
	// so the click handler runs on the original, intact button.
	doc := js.Global().Get("document")
	doc.Call("addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) any {
		inPointerDown = true
		return nil
	}), true)
	doc.Call("addEventListener", "mouseup", js.FuncOf(func(this js.Value, args []js.Value) any {
		clearPointerDown()
		return nil
	}), true)
	doc.Call("addEventListener", "touchstart", js.FuncOf(func(this js.Value, args []js.Value) any {
		inPointerDown = true
		return nil
	}), true)
	doc.Call("addEventListener", "touchend", js.FuncOf(func(this js.Value, args []js.Value) any {
		clearPointerDown()
		return nil
	}), true)
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

// SetInterval calls the callback repeatedly at the given interval (milliseconds).
// The timer is automatically cleared when the user navigates to a different page.
// Returns a TimerHandle that can be used to cancel the interval manually.
func (r *Router) SetInterval(callback func(), ms int) *TimerHandle {
	id := js.Global().Call("setInterval", js.FuncOf(func(this js.Value, args []js.Value) any {
		callback()
		return nil
	}), ms).Int()
	h := &TimerHandle{id: id, interval: true, router: r}
	r.activeTimers = append(r.activeTimers, h)
	return h
}

// SetTimeout calls the callback once after the given delay (milliseconds).
// The timer is automatically cleared when the user navigates to a different page.
// Returns a TimerHandle that can be used to cancel the timeout manually.
func (r *Router) SetTimeout(callback func(), ms int) *TimerHandle {
	id := js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
		callback()
		return nil
	}), ms).Int()
	h := &TimerHandle{id: id, interval: false, router: r}
	r.activeTimers = append(r.activeTimers, h)
	return h
}

func clearTimer(h *TimerHandle) {
	if h.interval {
		js.Global().Call("clearInterval", h.id)
	} else {
		js.Global().Call("clearTimeout", h.id)
	}
}

// clearPointerDown clears the pointer-down flag and fires any suppressed re-render.
// Called from mouseup/touchend listeners.
func clearPointerDown() {
	inPointerDown = false
	if pointerDownSuppressed != nil {
		r := pointerDownSuppressed
		pointerDownSuppressed = nil
		ScheduleRerender(r)
	}
}

// SetInChangeEvent sets the change event flag.
// When true, ScheduleRerender calls are deferred to prevent
// re-renders from interrupting click events. When the flag is
// cleared, any suppressed re-render is scheduled.
func SetInChangeEvent(v bool) {
	inChangeEvent = v
	if !v && changeEventSuppressed != nil {
		// Change event finished — schedule the deferred re-render
		r := changeEventSuppressed
		changeEventSuppressed = nil
		ScheduleRerender(r)
	}
}

// ScheduleRerender schedules a re-render after the current event loop completes.
// This batches multiple state changes into a single re-render and prevents
// re-renders from interrupting event handlers (like click handlers).
//
// Re-renders are deferred in two scenarios:
//  1. inChangeEvent: during synchronous change handler execution
//  2. inPointerDown: during the mousedown→mouseup window, to prevent
//     setTimeout(0) from destroying click targets before click fires (#112)
func ScheduleRerender(r *Router) {
	// Defer re-renders during change events to allow button clicks to work.
	// The deferred re-render is scheduled when SetInChangeEvent(false) is called.
	if inChangeEvent {
		changeEventSuppressed = r
		return
	}

	// Defer re-renders during the mousedown→mouseup window (#112).
	// The change event fires synchronously during mousedown (via blur), and
	// SetInChangeEvent(false) calls ScheduleRerender. Without this check,
	// the setTimeout(0) would fire before mouseup, destroying the click target.
	if inPointerDown {
		pointerDownSuppressed = r
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

// Path returns the current request path.
// In WASM, reads from route params or window.location.pathname.
func (r *Router) Path() string {
	if path, ok := r.routeParams["__path"]; ok {
		return path
	}
	return js.Global().Get("location").Get("pathname").String()
}

// Query returns a URL query parameter value by name.
// In WASM, reads from window.location.search.
func (r *Router) Query(name string) string {
	return wasmQuery(name)
}

// Login is server-side only. In WASM, returns an error.
func (r *Router) Login(user *SessionUser) error {
	return fmt.Errorf("Login() is server-side only")
}

// Logout is server-side only. In WASM, returns an error.
func (r *Router) Logout() error {
	return fmt.Errorf("Logout() is server-side only")
}

// Redirect performs a full-page navigation.
// In WASM, uses the redirect callback or window.location.assign().
func (r *Router) Redirect(path string) {
	if r.redirect != nil {
		r.redirect(path)
		return
	}
	js.Global().Get("location").Call("assign", path)
}
