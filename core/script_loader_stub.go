//go:build !js || !wasm

package core

// LoadScript is a no-op on non-WASM builds.
// On WASM, it dynamically loads an external JavaScript file with deduplication.
func LoadScript(url string, callback func(error)) {
	if callback != nil {
		callback(nil)
	}
}

// LoadCSS is a no-op on non-WASM builds.
// On WASM, it dynamically loads an external CSS stylesheet with deduplication.
func LoadCSS(url string, callback func(error)) {
	if callback != nil {
		callback(nil)
	}
}
