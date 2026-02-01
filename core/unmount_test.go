package core

import (
	"strings"
	"testing"
)

func TestOnUnmount_IgnoredInSSR(t *testing.T) {
	// OnUnmount should have no effect on HTML rendering (SSR)
	called := false
	node := El("div", Attrs{
		Class: "test",
		OnUnmount: func() {
			called = true
		},
	}, Text("hello"))

	html := node.Render(HTML()).HTML()

	if called {
		t.Error("OnUnmount should not be called during SSR rendering")
	}
	if !strings.Contains(html, `<div class="test">hello</div>`) {
		t.Errorf("unexpected HTML output: %s", html)
	}
}

func TestOnUnmount_DoesNotAffectHTMLOutput(t *testing.T) {
	// An element with OnUnmount should produce identical HTML to one without
	withUnmount := El("div", Attrs{
		ID:    "test",
		Class: "container",
		OnUnmount: func() {
			// cleanup
		},
	}, Text("content"))

	withoutUnmount := El("div", Attrs{
		ID:    "test",
		Class: "container",
	}, Text("content"))

	htmlWith := withUnmount.Render(HTML()).HTML()
	htmlWithout := withoutUnmount.Render(HTML()).HTML()

	if htmlWith != htmlWithout {
		t.Errorf("OnUnmount should not affect HTML output.\nwith:    %s\nwithout: %s", htmlWith, htmlWithout)
	}
}

func TestOnUnmount_WithOnMount_BothIgnoredInSSR(t *testing.T) {
	mountCalled := false
	unmountCalled := false

	node := El("video", Attrs{
		ID:       "player-1",
		Controls: true,
		OnMount: func(el any) {
			mountCalled = true
		},
		OnUnmount: func() {
			unmountCalled = true
		},
	})

	html := node.Render(HTML()).HTML()

	if mountCalled {
		t.Error("OnMount should not be called during SSR")
	}
	if unmountCalled {
		t.Error("OnUnmount should not be called during SSR")
	}
	if !strings.Contains(html, "<video") {
		t.Errorf("expected video element, got: %s", html)
	}
}

func TestFireUnmountCallbacks_NoOp(t *testing.T) {
	// On server builds, FireUnmountCallbacks is a no-op stub.
	// Should not panic when called.
	FireUnmountCallbacks()
}

func TestRegisterUnmount_NoOp(t *testing.T) {
	// On server builds, RegisterUnmount is a no-op stub.
	// Should not panic when called with a function or nil.
	RegisterUnmount(func() {})
	RegisterUnmount(nil)
}

func TestRegisterAndFireUnmount_ServerStubs(t *testing.T) {
	// Register several callbacks and fire — all no-ops on server
	called := 0
	RegisterUnmount(func() { called++ })
	RegisterUnmount(func() { called++ })
	RegisterUnmount(func() { called++ })
	FireUnmountCallbacks()

	// On server builds these are no-ops, so called should remain 0
	if called != 0 {
		t.Errorf("expected 0 callbacks fired on server, got: %d", called)
	}
}

func TestOnUnmount_NestedElements(t *testing.T) {
	// OnUnmount on nested elements should not affect HTML output
	node := El("div", Attrs{Class: "outer", OnUnmount: func() {}},
		El("div", Attrs{Class: "inner", OnUnmount: func() {}},
			Text("nested"),
		),
	)

	html := node.Render(HTML()).HTML()

	if !strings.Contains(html, `class="outer"`) {
		t.Errorf("expected outer class, got: %s", html)
	}
	if !strings.Contains(html, `class="inner"`) {
		t.Errorf("expected inner class, got: %s", html)
	}
	if !strings.Contains(html, "nested") {
		t.Errorf("expected nested text, got: %s", html)
	}
}

func TestOnUnmount_NilDoesNotPanic(t *testing.T) {
	// Setting OnUnmount to nil should be fine
	node := El("div", Attrs{
		OnUnmount: nil,
	}, Text("ok"))

	html := node.Render(HTML()).HTML()
	if !strings.Contains(html, "ok") {
		t.Errorf("expected text content, got: %s", html)
	}
}
