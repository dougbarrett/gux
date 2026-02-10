//go:build js && wasm

package core

// listenAndServe is a no-op on WASM — Run is never called in the browser.
func (a *App) listenAndServe(addr string) error {
	return nil
}
