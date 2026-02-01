package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_ReturnsNonNil(t *testing.T) {
	app := New()
	h := app.Handler()
	if h == nil {
		t.Fatal("Handler() returned nil")
	}
}

func TestHandler_ServesRoutes(t *testing.T) {
	app := New()
	app.Routes().
		Hybrid("/", func(r *Router) func() Node {
			return func() Node {
				return Text("home page")
			}
		})

	server := httptest.NewServer(app.Handler())
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET / status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestHandler_CustomEndpoints(t *testing.T) {
	app := New()

	type PingResponse struct {
		Status string `json:"status"`
	}

	APIGet(app, "/api/ping", func(ctx *APIContext) (PingResponse, error) {
		return PingResponse{Status: "ok"}, nil
	}).Public()

	server := httptest.NewServer(app.Handler())
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/ping")
	if err != nil {
		t.Fatalf("GET /api/ping: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var result PingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Status != "ok" {
		t.Errorf("status = %q, want %q", result.Status, "ok")
	}
}

func TestHandler_CSRFMiddlewareApplied(t *testing.T) {
	app := New()

	type Req struct {
		Value string `json:"value"`
	}
	type Resp struct {
		OK bool `json:"ok"`
	}

	API(app, "POST", "/api/test", func(ctx *APIContext, req Req) (Resp, error) {
		return Resp{OK: true}, nil
	}).Public()

	server := httptest.NewServer(app.Handler())
	defer server.Close()

	// POST without CSRF token should be rejected (CSRF enabled by default)
	resp, err := http.Post(server.URL+"/api/test", "application/json", strings.NewReader(`{"value":"x"}`))
	if err != nil {
		t.Fatalf("POST /api/test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("POST without CSRF token: status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestHandler_CSRFDisabled(t *testing.T) {
	app := New()
	app.DisableCSRF()

	type Req struct {
		Value string `json:"value"`
	}
	type Resp struct {
		OK bool `json:"ok"`
	}

	API(app, "POST", "/api/test", func(ctx *APIContext, req Req) (Resp, error) {
		return Resp{OK: true}, nil
	}).Public()

	server := httptest.NewServer(app.Handler())
	defer server.Close()

	// POST without CSRF token should pass when CSRF is disabled
	resp, err := http.Post(server.URL+"/api/test", "application/json", strings.NewReader(`{"value":"x"}`))
	if err != nil {
		t.Fatalf("POST /api/test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("POST with CSRF disabled: status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestHandler_WithRecorder(t *testing.T) {
	app := New()
	app.DisableCSRF()
	app.Routes().
		Hybrid("/hello", func(r *Router) func() Node {
			return func() Node {
				return Text("hello world")
			}
		})

	handler := app.Handler()

	req := httptest.NewRequest("GET", "/hello", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "hello world") {
		t.Errorf("body does not contain 'hello world': %s", rr.Body.String())
	}
}
