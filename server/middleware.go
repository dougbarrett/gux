package server

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain combines multiple middleware into a single middleware
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Logger logs request method, path, and duration
func Logger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
		})
	}
}

// CORS adds Cross-Origin Resource Sharing headers
func CORS(opts CORSOptions) Middleware {
	if opts.AllowOrigin == "" {
		opts.AllowOrigin = "*"
	}
	if opts.AllowMethods == "" {
		opts.AllowMethods = "GET, POST, PUT, DELETE, OPTIONS"
	}
	if opts.AllowHeaders == "" {
		opts.AllowHeaders = "Content-Type, Authorization"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", opts.AllowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", opts.AllowMethods)
			w.Header().Set("Access-Control-Allow-Headers", opts.AllowHeaders)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type CORSOptions struct {
	AllowOrigin  string
	AllowMethods string
	AllowHeaders string
}

// Recover catches panics and returns 500
func Recover() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("panic: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID adds a unique request ID header
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		var counter uint64
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			counter++
			w.Header().Set("X-Request-ID", fmt.Sprintf("%d-%d", time.Now().UnixNano(), counter))
			next.ServeHTTP(w, r)
		})
	}
}

