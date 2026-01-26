//go:build js && wasm

package main

import (
	"encoding/json"
	"strings"
	"syscall/js"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/marketing/pages"
)

// matchRoute checks if a path matches a parameterized route pattern.
// prefix is the part before the param, suffix is the part after.
// e.g., matchRoute("/admin/users/123", "/admin/users/", "") returns true
// e.g., matchRoute("/admin/users/123/posts/new", "/admin/users/", "/posts/new") returns true
func matchRoute(path, prefix, suffix string) bool {
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	rest := path[len(prefix):]
	if suffix == "" {
		// No suffix - just need something after the prefix (the param value)
		return len(rest) > 0 && !strings.Contains(rest, "/")
	}
	// Has suffix - need to match suffix at the end
	if !strings.HasSuffix(rest, suffix) {
		return false
	}
	// Extract the param value (between prefix and suffix)
	paramValue := rest[:len(rest)-len(suffix)]
	return len(paramValue) > 0 && !strings.Contains(paramValue, "/")
}

// extractRouteParam extracts the parameter value from a path given prefix and suffix.
// e.g., extractRouteParam("/admin/users/123", "/admin/users/", "") returns "123"
// e.g., extractRouteParam("/admin/users/123/posts/new", "/admin/users/", "/posts/new") returns "123"
func extractRouteParam(path, prefix, suffix string) string {
	rest := path[len(prefix):]
	if suffix == "" {
		return rest
	}
	return rest[:len(rest)-len(suffix)]
}

// fetchLoader fetches page state from loader endpoint
func fetchLoader(path string, callback func(map[string]any)) {
	loaderPath := "/__gux_api/pages" + path
	if path == "/" {
		loaderPath = "/__gux_api/pages/index"
	}

	promise := js.Global().Call("fetch", loaderPath)
	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
		resp := args[0]
		if resp.Get("ok").Bool() {
			resp.Call("json").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
				// Convert JS object to Go map
				state := make(map[string]any)
				jsObj := args[0]
				keys := js.Global().Get("Object").Call("keys", jsObj)
				for i := 0; i < keys.Get("length").Int(); i++ {
					key := keys.Index(i).String()
					val := jsObj.Get(key)
					switch val.Type() {
					case js.TypeNumber:
						state[key] = val.Float()
					case js.TypeBoolean:
						state[key] = val.Bool()
					case js.TypeString:
						state[key] = val.String()
					}
				}
				callback(state)
				return nil
			}))
		} else {
			callback(nil)
		}
		return nil
	}))
}

func main() {
	document := js.Global().Get("document")
	window := js.Global()
	container := document.Call("getElementById", "app")

	var router *core.Router
	var render func()

	render = func() {
		// Save focus state before re-render
		activeElement := document.Get("activeElement")
		var focusName string
		if !activeElement.IsNull() && !activeElement.IsUndefined() {
			focusName = activeElement.Call("getAttribute", "name").String()
		}

		container.Set("innerHTML", "")
		path := js.Global().Get("location").Get("pathname").String()
		var component func() core.Node
		switch path {
		case "/contact":
			component = pages.Contact(router)
		case "/about":
			component = pages.About(router)
		case "/pricing":
			component = pages.Pricing(router)
		case "/features":
			component = pages.Features(router)
		case "/":
			component = pages.Home(router)
		}
		if component == nil {
			component = pages.Home(router)
		}

		node := component()
		result := node.Render(core.DOM())
		if domVal := result.DOMValue(); domVal != nil {
			container.Call("appendChild", domVal.(js.Value))
		}

		// Intercept link clicks for client-side navigation
		links := document.Call("querySelectorAll", "a[href]")
		for i := 0; i < links.Get("length").Int(); i++ {
			link := links.Call("item", i)
			href := link.Get("href").String()
			origin := window.Get("location").Get("origin").String()

			// Skip external links (marked with data-gux-external)
			if link.Call("getAttribute", "data-gux-external").String() == "true" {
				continue
			}

			// Only intercept internal links
			if len(href) >= len(origin) && href[:len(origin)] == origin {
				link.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
					args[0].Call("preventDefault")
					path := this.Get("pathname").String()
					router.Navigate(path)
					return nil
				}))
			}
		}

		// Restore focus after re-render
		if focusName != "" {
			newElement := document.Call("querySelector", "[name=\""+focusName+"\"]")
			if !newElement.IsNull() && !newElement.IsUndefined() {
				newElement.Call("focus")
			}
		}
	}

	// Navigate fetches page data then renders
	navigate := func(path string) {
		window.Get("history").Call("pushState", nil, "", path)
		fetchLoader(path, func(state map[string]any) {
			if state != nil {
				router.Hydrate(state)
			}
			render()
		})
	}

	router = core.NewRouter(render)
	router.SetNavigate(navigate)

	// Hydrate state from SSR
	stateEl := document.Call("getElementById", "__gux_state")
	if !stateEl.IsNull() && !stateEl.IsUndefined() {
		stateJSON := stateEl.Get("textContent").String()
		var state map[string]any
		if err := json.Unmarshal([]byte(stateJSON), &state); err == nil {
			router.Hydrate(state)
		}
	}

	// Handle browser back/forward
	window.Call("addEventListener", "popstate", js.FuncOf(func(this js.Value, args []js.Value) any {
		render()
		return nil
	}))

	render()

	select {}
}
