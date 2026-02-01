package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSessionUser_NoAuth(t *testing.T) {
	app := New()
	req := httptest.NewRequest("GET", "/", nil)
	user := GetSessionUser(app, req)
	if user != nil {
		t.Error("expected nil user when auth not configured")
	}
}

func TestGetSessionUser_NoCookie(t *testing.T) {
	app := New()
	app.EnableAuth(AuthConfig{
		SessionStore: NewMemorySessionStore(),
	})
	req := httptest.NewRequest("GET", "/", nil)
	user := GetSessionUser(app, req)
	if user != nil {
		t.Error("expected nil user when no session cookie")
	}
}

func TestGetSessionUser_ValidSession(t *testing.T) {
	store := NewMemorySessionStore()
	app := New()
	app.EnableAuth(AuthConfig{
		SessionStore: store,
		CookieName:   "__test_session",
		CookieMaxAge: 3600,
	})

	// Create a session directly in the store
	testUser := &SessionUser{ID: "42", Email: "test@example.com", Roles: []string{"admin"}}
	store.Set("test-session-id", testUser, 3600_000_000_000)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "__test_session", Value: "test-session-id"})

	user := GetSessionUser(app, req)
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.ID != "42" {
		t.Errorf("expected ID=42, got %s", user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email=test@example.com, got %s", user.Email)
	}
}

func TestGetSessionUser_InvalidSession(t *testing.T) {
	store := NewMemorySessionStore()
	app := New()
	app.EnableAuth(AuthConfig{
		SessionStore: store,
		CookieName:   "__test_session",
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "__test_session", Value: "nonexistent-session"})

	user := GetSessionUser(app, req)
	if user != nil {
		t.Error("expected nil user for invalid session ID")
	}
}

func TestLoginFromHTTP_NoAuth(t *testing.T) {
	app := New()
	w := httptest.NewRecorder()
	err := LoginFromHTTP(app, w, &SessionUser{ID: "1"})
	if err == nil {
		t.Error("expected error when auth not configured")
	}
}

func TestLoginFromHTTP_Success(t *testing.T) {
	store := NewMemorySessionStore()
	app := New()
	app.EnableAuth(AuthConfig{
		SessionStore: store,
		CookieName:   "__test_session",
		CookieMaxAge: 3600,
		CookiePath:   "/",
	})

	w := httptest.NewRecorder()
	testUser := &SessionUser{ID: "99", Email: "new@example.com"}

	err := LoginFromHTTP(app, w, testUser)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify cookie was set
	resp := w.Result()
	cookies := resp.Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "__test_session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected session cookie to be set")
	}
	if sessionCookie.Value == "" {
		t.Error("expected non-empty session cookie value")
	}

	// Verify session can be retrieved from store
	user, err := store.Get(sessionCookie.Value)
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	if user == nil {
		t.Fatal("expected user in store, got nil")
	}
	if user.ID != "99" {
		t.Errorf("expected ID=99, got %s", user.ID)
	}
	if user.Email != "new@example.com" {
		t.Errorf("expected email=new@example.com, got %s", user.Email)
	}
}

func TestLoginFromHTTP_DefaultMaxAge(t *testing.T) {
	store := NewMemorySessionStore()
	app := New()
	app.EnableAuth(AuthConfig{
		SessionStore: store,
		CookieName:   "__test_session",
		// CookieMaxAge not set — should use DefaultSessionMaxAge
	})

	w := httptest.NewRecorder()
	err := LoginFromHTTP(app, w, &SessionUser{ID: "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify cookie was set (default max age should work)
	resp := w.Result()
	cookies := resp.Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "__test_session" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected session cookie with default max age")
	}
}

func TestLogoutFromHTTP_NoAuth(t *testing.T) {
	app := New()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	// Should not panic
	LogoutFromHTTP(app, w, req)
}

func TestLogoutFromHTTP_Success(t *testing.T) {
	store := NewMemorySessionStore()
	app := New()
	app.EnableAuth(AuthConfig{
		SessionStore: store,
		CookieName:   "__test_session",
		CookieMaxAge: 3600,
		CookiePath:   "/",
	})

	// First, create a session
	testUser := &SessionUser{ID: "42", Email: "test@example.com"}
	store.Set("session-to-delete", testUser, 3600_000_000_000)

	// Verify it exists
	user, _ := store.Get("session-to-delete")
	if user == nil {
		t.Fatal("session should exist before logout")
	}

	// Logout
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "__test_session", Value: "session-to-delete"})

	LogoutFromHTTP(app, w, req)

	// Verify session was deleted from store
	user, _ = store.Get("session-to-delete")
	if user != nil {
		t.Error("expected session to be deleted after logout")
	}

	// Verify cookie was cleared (MaxAge = -1)
	resp := w.Result()
	for _, c := range resp.Cookies() {
		if c.Name == "__test_session" {
			if c.MaxAge != -1 {
				t.Errorf("expected MaxAge=-1 to clear cookie, got %d", c.MaxAge)
			}
			return
		}
	}
	t.Error("expected clear cookie to be set")
}

func TestLoginThenGetSessionUser_RoundTrip(t *testing.T) {
	store := NewMemorySessionStore()
	app := New()
	app.EnableAuth(AuthConfig{
		SessionStore: store,
		CookieName:   "__test_session",
		CookieMaxAge: 3600,
		CookiePath:   "/",
	})

	// Login
	w := httptest.NewRecorder()
	testUser := &SessionUser{
		ID:    "7",
		Email: "roundtrip@example.com",
		Roles: []string{"editor", "viewer"},
	}
	err := LoginFromHTTP(app, w, testUser)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Extract session cookie
	resp := w.Result()
	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "__test_session" {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("no session cookie after login")
	}

	// Use cookie in new request to get session user
	req := httptest.NewRequest("GET", "/api/me", nil)
	req.AddCookie(sessionCookie)

	user := GetSessionUser(app, req)
	if user == nil {
		t.Fatal("expected user from session, got nil")
	}
	if user.ID != "7" {
		t.Errorf("expected ID=7, got %s", user.ID)
	}
	if user.Email != "roundtrip@example.com" {
		t.Errorf("expected email=roundtrip@example.com, got %s", user.Email)
	}
	if len(user.Roles) != 2 || user.Roles[0] != "editor" {
		t.Errorf("expected roles [editor, viewer], got %v", user.Roles)
	}
}

func TestLoginLogoutGetUser_FullCycle(t *testing.T) {
	store := NewMemorySessionStore()
	app := New()
	app.EnableAuth(AuthConfig{
		SessionStore: store,
		CookieName:   "__test_session",
		CookieMaxAge: 3600,
		CookiePath:   "/",
	})

	// Login
	loginW := httptest.NewRecorder()
	LoginFromHTTP(app, loginW, &SessionUser{ID: "1", Email: "user@test.com"})
	loginResp := loginW.Result()
	var cookie *http.Cookie
	for _, c := range loginResp.Cookies() {
		if c.Name == "__test_session" {
			cookie = c
		}
	}

	// Verify logged in
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(cookie)
	user := GetSessionUser(app, req)
	if user == nil || user.ID != "1" {
		t.Fatal("expected logged-in user")
	}

	// Logout
	logoutW := httptest.NewRecorder()
	logoutReq := httptest.NewRequest("POST", "/logout", nil)
	logoutReq.AddCookie(cookie)
	LogoutFromHTTP(app, logoutW, logoutReq)

	// Verify logged out
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.AddCookie(cookie) // same cookie, but session deleted
	user2 := GetSessionUser(app, req2)
	if user2 != nil {
		t.Error("expected nil user after logout")
	}
}
