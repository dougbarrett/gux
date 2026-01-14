//go:build js && wasm

// Package fetch provides a TinyGo-compatible HTTP client using the browser's fetch API
package fetch

import (
	"errors"
	"syscall/js"
)

// Response represents an HTTP response
type Response struct {
	Status     int
	StatusText string
	OK         bool
	Body       string
	Headers    map[string]string
}

// Options configures a fetch request
type Options struct {
	Method  string
	Headers map[string]string
	Body    string
}

// Error types
var (
	ErrFetchFailed  = errors.New("fetch failed")
	ErrNetworkError = errors.New("network error")
)

// Fetch performs an HTTP request using the browser's fetch API
// This is synchronous and blocks until the request completes
func Fetch(url string, opts *Options) (*Response, error) {
	done := make(chan struct{})
	var response *Response
	var fetchErr error

	// Build fetch options
	jsOpts := js.Global().Get("Object").New()

	if opts != nil {
		if opts.Method != "" {
			jsOpts.Set("method", opts.Method)
		}

		if len(opts.Headers) > 0 {
			headers := js.Global().Get("Object").New()
			for k, v := range opts.Headers {
				headers.Set(k, v)
			}
			jsOpts.Set("headers", headers)
		}

		if opts.Body != "" {
			jsOpts.Set("body", opts.Body)
		}
	}

	// Success handler
	thenFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		resp := args[0]

		// Get response body as text
		resp.Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
			bodyText := args[0].String()

			response = &Response{
				Status:     resp.Get("status").Int(),
				StatusText: resp.Get("statusText").String(),
				OK:         resp.Get("ok").Bool(),
				Body:       bodyText,
				Headers:    make(map[string]string),
			}

			close(done)
			return nil
		}))

		return nil
	})

	// Error handler
	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		fetchErr = errors.New(args[0].Get("message").String())
		close(done)
		return nil
	})

	// Execute fetch
	js.Global().Call("fetch", url, jsOpts).Call("then", thenFunc).Call("catch", catchFunc)

	// Wait for completion
	<-done

	// Clean up
	thenFunc.Release()
	catchFunc.Release()

	if fetchErr != nil {
		return nil, fetchErr
	}

	return response, nil
}

// Get performs a GET request
func Get(url string, headers map[string]string) (*Response, error) {
	return Fetch(url, &Options{
		Method:  "GET",
		Headers: headers,
	})
}

// Post performs a POST request with a JSON body
func Post(url string, body string, headers map[string]string) (*Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json"
	}

	return Fetch(url, &Options{
		Method:  "POST",
		Headers: headers,
		Body:    body,
	})
}

// Put performs a PUT request with a JSON body
func Put(url string, body string, headers map[string]string) (*Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json"
	}

	return Fetch(url, &Options{
		Method:  "PUT",
		Headers: headers,
		Body:    body,
	})
}

// Delete performs a DELETE request
func Delete(url string, headers map[string]string) (*Response, error) {
	return Fetch(url, &Options{
		Method:  "DELETE",
		Headers: headers,
	})
}
