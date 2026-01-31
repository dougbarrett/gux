//go:build !js || !wasm

package ui

// createVideoJSInitializer is a no-op on non-WASM builds.
// On WASM, it returns an OnMount callback that initializes VideoJS.
func createVideoJSInitializer(videoID string, props VideoPlayerProps) func(el any) {
	return nil
}

// createNativeVideoHandlers is a no-op on non-WASM builds.
// On WASM, it returns an OnMount callback that attaches HTML5 video event listeners.
func createNativeVideoHandlers(props VideoPlayerProps) func(el any) {
	return nil
}
