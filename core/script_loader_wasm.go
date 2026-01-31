//go:build js && wasm

package core

import (
	"fmt"
	"sync"
	"syscall/js"
)

var (
	loadedScripts  = make(map[string]bool)
	loadingScripts = make(map[string][]func(error))
	scriptMu       sync.Mutex
)

// LoadScript dynamically loads an external JavaScript file.
// Multiple calls with the same URL share the same load operation (deduplication).
// The callback is called with nil on success, or an error on failure.
func LoadScript(url string, callback func(error)) {
	scriptMu.Lock()

	if loadedScripts[url] {
		scriptMu.Unlock()
		if callback != nil {
			callback(nil)
		}
		return
	}

	if callbacks, loading := loadingScripts[url]; loading {
		loadingScripts[url] = append(callbacks, callback)
		scriptMu.Unlock()
		return
	}

	loadingScripts[url] = []func(error){callback}
	scriptMu.Unlock()

	doc := js.Global().Get("document")
	script := doc.Call("createElement", "script")
	script.Set("src", url)
	script.Set("async", true)

	script.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) any {
		scriptMu.Lock()
		loadedScripts[url] = true
		callbacks := loadingScripts[url]
		delete(loadingScripts, url)
		scriptMu.Unlock()

		for _, cb := range callbacks {
			if cb != nil {
				cb(nil)
			}
		}
		return nil
	}))

	script.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) any {
		scriptMu.Lock()
		callbacks := loadingScripts[url]
		delete(loadingScripts, url)
		scriptMu.Unlock()

		err := fmt.Errorf("failed to load script: %s", url)
		for _, cb := range callbacks {
			if cb != nil {
				cb(err)
			}
		}
		return nil
	}))

	doc.Get("head").Call("appendChild", script)
}

// LoadCSS dynamically loads an external CSS stylesheet.
// Multiple calls with the same URL share the same load operation (deduplication).
// The callback is called with nil on success, or an error on failure.
func LoadCSS(url string, callback func(error)) {
	scriptMu.Lock()

	if loadedScripts[url] {
		scriptMu.Unlock()
		if callback != nil {
			callback(nil)
		}
		return
	}

	if callbacks, loading := loadingScripts[url]; loading {
		loadingScripts[url] = append(callbacks, callback)
		scriptMu.Unlock()
		return
	}

	loadingScripts[url] = []func(error){callback}
	scriptMu.Unlock()

	doc := js.Global().Get("document")
	link := doc.Call("createElement", "link")
	link.Set("rel", "stylesheet")
	link.Set("href", url)

	link.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) any {
		scriptMu.Lock()
		loadedScripts[url] = true
		callbacks := loadingScripts[url]
		delete(loadingScripts, url)
		scriptMu.Unlock()

		for _, cb := range callbacks {
			if cb != nil {
				cb(nil)
			}
		}
		return nil
	}))

	link.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) any {
		scriptMu.Lock()
		callbacks := loadingScripts[url]
		delete(loadingScripts, url)
		scriptMu.Unlock()

		err := fmt.Errorf("failed to load CSS: %s", url)
		for _, cb := range callbacks {
			if cb != nil {
				cb(err)
			}
		}
		return nil
	}))

	doc.Get("head").Call("appendChild", link)
}
