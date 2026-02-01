//go:build !js || !wasm

package core

// RegisterUnmount is a no-op on server — unmount callbacks only run in WASM.
func RegisterUnmount(fn func()) {}

// FireUnmountCallbacks is a no-op on server — unmount callbacks only run in WASM.
func FireUnmountCallbacks() {}
