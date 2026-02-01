//go:build !js || !wasm

package ui

// createFileUploadInit is a no-op on non-WASM builds.
// On WASM, it returns an OnMount callback that initializes file upload interactivity.
func createFileUploadInit(id string, props FileUploadProps) func(el any) {
	return nil
}
