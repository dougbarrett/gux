//go:build !js || !wasm

package core

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// listenAndServe starts the HTTP server with graceful shutdown on SIGTERM/SIGINT.
func (a *App) listenAndServe(addr string) error {
	srv := &http.Server{Addr: addr, Handler: a.Handler()}
	a.server = srv

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		<-sigCh
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
