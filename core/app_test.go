package core

import (
	"net/http"
	"syscall"
	"testing"
	"time"
)

func TestRunGracefulShutdown(t *testing.T) {
	app := New()
	app.SetTitle("Test App")

	// Register a simple route so Handler() works
	app.Routes().GET("/", func(r *Router) func() Node {
		return func() Node { return Text("ok") }
	})

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run(":0") // :0 picks a random free port
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Server should be set
	srv := app.Server()
	if srv == nil {
		t.Fatal("Server() returned nil after Run was called")
	}

	// Send SIGINT to trigger graceful shutdown
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	// Wait for Run to return
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return within 5 seconds after SIGINT")
	}
}

func TestRunDrainsInFlightRequests(t *testing.T) {
	app := New()
	app.SetTitle("Drain Test")

	handlerStarted := make(chan struct{})
	handlerDone := make(chan struct{})

	// Register a slow handler
	app.Routes().GET("/slow", func(r *Router) func() Node {
		close(handlerStarted)
		// Simulate a slow request
		time.Sleep(500 * time.Millisecond)
		close(handlerDone)
		return func() Node { return Text("done") }
	})

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run(":18921")
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Start a slow request
	go http.Get("http://localhost:18921/slow")

	// Wait for handler to start processing
	<-handlerStarted

	// Send SIGINT while request is in-flight
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	// The in-flight request should complete before shutdown
	select {
	case <-handlerDone:
		// Good - handler completed before shutdown
	case <-time.After(5 * time.Second):
		t.Fatal("In-flight request was not drained before shutdown")
	}

	// Run should return cleanly
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return within 5 seconds")
	}
}

func TestServerReturnsNilBeforeRun(t *testing.T) {
	app := New()
	if srv := app.Server(); srv != nil {
		t.Fatalf("Server() should return nil before Run, got %v", srv)
	}
}
