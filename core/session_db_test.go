package core

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	return db
}

func TestDBSessionStore_SetAndGet(t *testing.T) {
	db := newTestDB(t)
	store := NewDBSessionStore(db)

	user := &SessionUser{
		ID:    "user-1",
		Email: "test@example.com",
		Name:  "Test User",
		Roles: []string{"admin"},
	}

	if err := store.Set("session-1", user, 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := store.Get("session-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected session user, got nil")
	}
	if got.ID != "user-1" {
		t.Errorf("expected ID user-1, got %s", got.ID)
	}
	if got.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", got.Email)
	}
	if len(got.Roles) != 1 || got.Roles[0] != "admin" {
		t.Errorf("expected roles [admin], got %v", got.Roles)
	}
}

func TestDBSessionStore_GetExpired(t *testing.T) {
	db := newTestDB(t)
	store := NewDBSessionStore(db)

	user := &SessionUser{ID: "user-1", Email: "test@example.com"}

	// Set with very short expiry
	if err := store.Set("session-expired", user, 1*time.Millisecond); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Wait for expiry
	time.Sleep(10 * time.Millisecond)

	got, err := store.Get("session-expired")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for expired session, got %+v", got)
	}
}

func TestDBSessionStore_GetNonExistent(t *testing.T) {
	db := newTestDB(t)
	store := NewDBSessionStore(db)

	got, err := store.Get("does-not-exist")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for non-existent session, got %+v", got)
	}
}

func TestDBSessionStore_Delete(t *testing.T) {
	db := newTestDB(t)
	store := NewDBSessionStore(db)

	user := &SessionUser{ID: "user-1", Email: "test@example.com"}
	if err := store.Set("session-del", user, 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	if err := store.Delete("session-del"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	got, err := store.Get("session-del")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil after delete, got %+v", got)
	}
}

func TestDBSessionStore_Overwrite(t *testing.T) {
	db := newTestDB(t)
	store := NewDBSessionStore(db)

	user1 := &SessionUser{ID: "user-1", Email: "first@example.com"}
	user2 := &SessionUser{ID: "user-2", Email: "second@example.com"}

	if err := store.Set("session-ow", user1, 1*time.Hour); err != nil {
		t.Fatalf("Set 1 failed: %v", err)
	}
	if err := store.Set("session-ow", user2, 1*time.Hour); err != nil {
		t.Fatalf("Set 2 failed: %v", err)
	}

	got, err := store.Get("session-ow")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected session user, got nil")
	}
	if got.ID != "user-2" {
		t.Errorf("expected overwritten user ID user-2, got %s", got.ID)
	}
	if got.Email != "second@example.com" {
		t.Errorf("expected overwritten email second@example.com, got %s", got.Email)
	}
}

func TestDBSessionStore_Cleanup(t *testing.T) {
	db := newTestDB(t)
	store := NewDBSessionStore(db)

	user := &SessionUser{ID: "user-1", Email: "test@example.com"}

	// Create an expired session
	if err := store.Set("session-cleanup", user, 1*time.Millisecond); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	// Create a valid session
	if err := store.Set("session-valid", user, 1*time.Hour); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	store.cleanup()

	// Expired session should be gone
	got, _ := store.Get("session-cleanup")
	if got != nil {
		t.Errorf("expected expired session to be cleaned up, got %+v", got)
	}

	// Valid session should remain
	got, err := store.Get("session-valid")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got == nil {
		t.Error("expected valid session to survive cleanup")
	}
}
