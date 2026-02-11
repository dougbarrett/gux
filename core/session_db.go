//go:build !js || !wasm

package core

import (
	"fmt"
	"log"
	"reflect"
	"time"
)

// DBSession is the GORM model for database-backed sessions.
type DBSession struct {
	ID        string    `gorm:"primarykey;type:varchar(64)"`
	Data      string    `gorm:"type:text"`
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName overrides the default table name.
func (DBSession) TableName() string {
	return "db_sessions"
}

// DBSessionStore is a database-backed session store using GORM.
// It persists sessions across server restarts, unlike MemorySessionStore.
type DBSessionStore struct {
	db any // GORM *gorm.DB
}

// NewDBSessionStore creates a new database-backed session store.
// The db parameter should be a *gorm.DB instance. It auto-migrates
// the db_sessions table and starts a background cleanup goroutine.
func NewDBSessionStore(db any) *DBSessionStore {
	store := &DBSessionStore{db: db}
	store.autoMigrate()

	// Start background cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			store.cleanup()
		}
	}()

	return store
}

// Get retrieves a session user by session ID.
func (s *DBSessionStore) Get(sessionID string) (*SessionUser, error) {
	var session DBSession

	// db.Where("id = ? AND expires_at > ?", sessionID, time.Now()).First(&session)
	dbVal := reflect.ValueOf(s.db)
	whereMethod := dbVal.MethodByName("Where")
	if !whereMethod.IsValid() {
		return nil, fmt.Errorf("db missing Where method")
	}
	result := whereMethod.Call([]reflect.Value{
		reflect.ValueOf("id = ? AND expires_at > ?"),
		reflect.ValueOf(sessionID),
		reflect.ValueOf(time.Now()),
	})
	if len(result) == 0 {
		return nil, fmt.Errorf("Where returned no result")
	}

	firstMethod := result[0].MethodByName("First")
	if !firstMethod.IsValid() {
		return nil, fmt.Errorf("db missing First method")
	}
	ret := firstMethod.Call([]reflect.Value{reflect.ValueOf(&session)})
	if err := extractReflectError(ret); err != nil {
		// Not found is not an error — just means no valid session
		return nil, nil
	}

	return sessionUserFromJSON(session.Data), nil
}

// Set stores a session user with the given session ID and max age.
func (s *DBSessionStore) Set(sessionID string, user *SessionUser, maxAge time.Duration) error {
	session := DBSession{
		ID:        sessionID,
		Data:      sessionUserToJSON(user),
		ExpiresAt: time.Now().Add(maxAge),
	}

	// db.Save(&session) — upserts (insert or update)
	dbVal := reflect.ValueOf(s.db)
	saveMethod := dbVal.MethodByName("Save")
	if !saveMethod.IsValid() {
		return fmt.Errorf("db missing Save method")
	}
	ret := saveMethod.Call([]reflect.Value{reflect.ValueOf(&session)})
	if err := extractReflectError(ret); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}

// Delete removes a session by ID.
func (s *DBSessionStore) Delete(sessionID string) error {
	// db.Where("id = ?", sessionID).Delete(&DBSession{})
	dbVal := reflect.ValueOf(s.db)
	whereMethod := dbVal.MethodByName("Where")
	if !whereMethod.IsValid() {
		return fmt.Errorf("db missing Where method")
	}
	result := whereMethod.Call([]reflect.Value{
		reflect.ValueOf("id = ?"),
		reflect.ValueOf(sessionID),
	})
	if len(result) == 0 {
		return fmt.Errorf("Where returned no result")
	}

	deleteMethod := result[0].MethodByName("Delete")
	if !deleteMethod.IsValid() {
		return fmt.Errorf("db missing Delete method")
	}
	ret := deleteMethod.Call([]reflect.Value{reflect.ValueOf(&DBSession{})})
	if err := extractReflectError(ret); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// autoMigrate creates the db_sessions table if it doesn't exist.
func (s *DBSessionStore) autoMigrate() {
	if s.db == nil {
		return
	}
	dbVal := reflect.ValueOf(s.db)
	migrateMethod := dbVal.MethodByName("AutoMigrate")
	if migrateMethod.IsValid() {
		migrateMethod.Call([]reflect.Value{reflect.ValueOf(&DBSession{})})
	}
}

// cleanup removes expired sessions from the database.
func (s *DBSessionStore) cleanup() {
	if s.db == nil {
		return
	}
	// db.Where("expires_at < ?", time.Now()).Delete(&DBSession{})
	dbVal := reflect.ValueOf(s.db)
	whereMethod := dbVal.MethodByName("Where")
	if !whereMethod.IsValid() {
		return
	}
	result := whereMethod.Call([]reflect.Value{
		reflect.ValueOf("expires_at < ?"),
		reflect.ValueOf(time.Now()),
	})
	if len(result) == 0 {
		return
	}

	deleteMethod := result[0].MethodByName("Delete")
	if !deleteMethod.IsValid() {
		return
	}
	ret := deleteMethod.Call([]reflect.Value{reflect.ValueOf(&DBSession{})})
	if err := extractReflectError(ret); err != nil {
		log.Printf("[session] failed to cleanup expired sessions: %v", err)
	}
}

// extractReflectError extracts an error from a GORM reflection result.
// GORM methods return *gorm.DB which has a public Error field.
// Returns nil if no error or if the result doesn't have an Error field.
func extractReflectError(ret []reflect.Value) error {
	if len(ret) == 0 {
		return nil
	}
	// GORM returns *gorm.DB — dereference pointer, then check Error field
	val := ret[0]
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	errField := val.FieldByName("Error")
	if !errField.IsValid() {
		return nil
	}
	if errField.IsNil() {
		return nil
	}
	if e, ok := errField.Interface().(error); ok {
		return e
	}
	return nil
}
