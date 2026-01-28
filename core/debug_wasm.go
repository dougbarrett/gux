//go:build js && wasm

package core

import (
	"fmt"
	"syscall/js"
)

// debugStateLog logs state operations directly to the browser console.
// This bypasses any callback mechanism to ensure logging works.
func debugStateLog(action, key string, value any, router *Router) {
	routerAddr := fmt.Sprintf("%p", router)
	stateSize := 0
	if router != nil && router.state != nil {
		stateSize = len(router.state)
	}
	js.Global().Get("console").Call("log",
		fmt.Sprintf("[Gux State] %s key=%q value=%v router=%s state_size=%d",
			action, key, value, routerAddr, stateSize))
}
