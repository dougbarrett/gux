//go:build !js || !wasm

package ui

// createMultiFileUploadInit is a no-op on non-WASM builds.
// On WASM, it returns an OnMount callback that initializes multi-file upload interactivity.
func createMultiFileUploadInit(id string, props MultiFileUploadProps) func(el any) {
	return nil
}
