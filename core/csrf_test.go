package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsExemptPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		exemptPaths []string
		want        bool
	}{
		{
			name:        "exact match",
			path:        "/api/login",
			exemptPaths: []string{"/api/login"},
			want:        true,
		},
		{
			name:        "no match",
			path:        "/api/users",
			exemptPaths: []string{"/api/login"},
			want:        false,
		},
		{
			name:        "prefix wildcard match",
			path:        "/api/invite/accept",
			exemptPaths: []string{"/api/invite/*"},
			want:        true,
		},
		{
			name:        "prefix wildcard no match",
			path:        "/api/users/1",
			exemptPaths: []string{"/api/invite/*"},
			want:        false,
		},
		{
			name:        "prefix wildcard exact prefix",
			path:        "/api/invite/",
			exemptPaths: []string{"/api/invite/*"},
			want:        true,
		},
		{
			name:        "multiple exempt paths",
			path:        "/api/webhook",
			exemptPaths: []string{"/api/login", "/api/webhook", "/api/health"},
			want:        true,
		},
		{
			name:        "empty exempt paths",
			path:        "/api/login",
			exemptPaths: nil,
			want:        false,
		},
		{
			name:        "trailing slash mismatch",
			path:        "/api/login/",
			exemptPaths: []string{"/api/login"},
			want:        false,
		},
		{
			name:        "wildcard covers trailing slash",
			path:        "/api/login/",
			exemptPaths: []string{"/api/login*"},
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExemptPath(tt.path, tt.exemptPaths)
			if got != tt.want {
				t.Errorf("isExemptPath(%q, %v) = %v, want %v", tt.path, tt.exemptPaths, got, tt.want)
			}
		})
	}
}

func TestCSRFMiddleware_ExemptPathSkipsValidation(t *testing.T) {
	config := CSRFConfig{
		Enabled:      true,
		CookiePath:   "/",
		CookieMaxAge: 3600,
		ExemptPaths:  []string{"/api/login", "/api/webhook/*"},
	}

	handler := CSRFMiddleware(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name       string
		method     string
		path       string
		csrfHeader string
		csrfCookie string
		wantStatus int
	}{
		{
			name:       "exempt exact path POST without token passes",
			method:     "POST",
			path:       "/api/login",
			wantStatus: http.StatusOK,
		},
		{
			name:       "exempt exact path POST with stale token passes",
			method:     "POST",
			path:       "/api/login",
			csrfHeader: "stale-token-value",
			csrfCookie: "different-cookie-value",
			wantStatus: http.StatusOK,
		},
		{
			name:       "exempt wildcard path POST without token passes",
			method:     "POST",
			path:       "/api/webhook/stripe",
			wantStatus: http.StatusOK,
		},
		{
			name:       "exempt wildcard path POST with mismatched token passes",
			method:     "POST",
			path:       "/api/webhook/stripe",
			csrfHeader: "old-token",
			csrfCookie: "new-token",
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-exempt path POST without token fails",
			method:     "POST",
			path:       "/api/users",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "non-exempt path POST with matching token passes",
			method:     "POST",
			path:       "/api/users",
			csrfHeader: "valid-token",
			csrfCookie: "valid-token",
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-exempt path POST with mismatched token fails",
			method:     "POST",
			path:       "/api/users",
			csrfHeader: "token-a",
			csrfCookie: "token-b",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "non-exempt path GET without token passes",
			method:     "GET",
			path:       "/api/users",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.csrfHeader != "" {
				req.Header.Set(CSRFHeaderName, tt.csrfHeader)
			}
			if tt.csrfCookie != "" {
				req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: tt.csrfCookie})
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestValidateCSRF(t *testing.T) {
	tests := []struct {
		name       string
		cookie     string
		header     string
		want       bool
	}{
		{
			name:   "matching tokens",
			cookie: "abc123",
			header: "abc123",
			want:   true,
		},
		{
			name:   "mismatched tokens",
			cookie: "abc123",
			header: "xyz789",
			want:   false,
		},
		{
			name:   "empty cookie",
			cookie: "",
			header: "abc123",
			want:   false,
		},
		{
			name:   "empty header",
			cookie: "abc123",
			header: "",
			want:   false,
		},
		{
			name:   "both empty",
			cookie: "",
			header: "",
			want:   false,
		},
		{
			name:   "different lengths",
			cookie: "short",
			header: "muchlonger",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", nil)
			if tt.cookie != "" {
				req.AddCookie(&http.Cookie{Name: CSRFCookieName, Value: tt.cookie})
			}
			if tt.header != "" {
				req.Header.Set(CSRFHeaderName, tt.header)
			}
			got := validateCSRF(req)
			if got != tt.want {
				t.Errorf("validateCSRF() = %v, want %v", got, tt.want)
			}
		})
	}
}
