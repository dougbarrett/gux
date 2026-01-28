//go:build !js || !wasm

package core

// debugStateLog is a no-op on server builds
func debugStateLog(action, key string, value any, router *Router) {
	// Server-side: no logging needed
}
