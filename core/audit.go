//go:build !js || !wasm

package core

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// AuditAction represents the type of CRUD operation performed.
type AuditAction string

const (
	AuditActionCreate AuditAction = "create"
	AuditActionUpdate AuditAction = "update"
	AuditActionDelete AuditAction = "delete"
)

// AuditEntry represents a single audit log record stored in the database.
// Entries are immutable and append-only — no UpdatedAt or DeletedAt.
type AuditEntry struct {
	ID         uint        `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time   `json:"created_at"`
	Action     AuditAction `gorm:"type:varchar(10);index" json:"action"`
	EntityType string      `gorm:"type:varchar(100);index" json:"entity_type"`
	EntityID   uint        `gorm:"index" json:"entity_id"`
	UserID     string      `gorm:"type:varchar(100);index" json:"user_id"`
	UserEmail  string      `gorm:"type:varchar(255)" json:"user_email"`
	IPAddress  string      `gorm:"type:varchar(45)" json:"ip_address"`
	Changes    string      `gorm:"type:text" json:"changes"`
}

// AuditConfig holds per-model audit logging configuration.
type AuditConfig struct {
	Enabled      bool
	IgnoreFields map[string]bool // Go struct field names to exclude from diffs
}

// WithAuditLog enables automatic audit logging for a CRUD model.
// Optionally specify Go struct field names to exclude from change diffs
// (e.g., "PasswordHash"). UpdatedAt and DeletedAt are always excluded.
func WithAuditLog(ignoreFields ...string) CRUDOption {
	return func(m *CRUDModel) {
		ignore := map[string]bool{
			"UpdatedAt": true,
			"DeletedAt": true,
		}
		for _, f := range ignoreFields {
			ignore[f] = true
		}
		m.AuditConfig = &AuditConfig{
			Enabled:      true,
			IgnoreFields: ignore,
		}
	}
}

// writeAuditLog writes an audit entry to the database.
// Failures are logged to stderr but never cause request errors.
func (a *App) writeAuditLog(entry AuditEntry) {
	if a.db == nil {
		return
	}
	dbVal := reflect.ValueOf(a.db)
	createMethod := dbVal.MethodByName("Create")
	if !createMethod.IsValid() {
		return
	}
	ret := createMethod.Call([]reflect.Value{reflect.ValueOf(&entry)})
	if len(ret) > 0 {
		errField := ret[0].MethodByName("Error")
		if errField.IsValid() {
			errVal := errField.Call(nil)
			if len(errVal) > 0 && !errVal[0].IsNil() {
				log.Printf("[audit] failed to write audit entry: %v", errVal[0].Interface())
			}
		}
	}
}

// autoMigrateAudit creates the audit_entries table if it doesn't exist.
func (a *App) autoMigrateAudit() {
	if a.db == nil {
		return
	}
	dbVal := reflect.ValueOf(a.db)
	migrateMethod := dbVal.MethodByName("AutoMigrate")
	if migrateMethod.IsValid() {
		migrateMethod.Call([]reflect.Value{reflect.ValueOf(&AuditEntry{})})
	}
}

// computeFieldDiffFromMaps compares before and after maps, returning a JSON string
// of changed fields. Fields in the ignore set (PascalCase) are excluded.
// The maps use snake_case keys from JSON serialization.
func computeFieldDiffFromMaps(before, after map[string]interface{}, ignore map[string]bool) string {
	diff := make(map[string]interface{})
	for key, newVal := range after {
		if shouldIgnoreField(key, ignore) {
			continue
		}
		oldVal, exists := before[key]
		if !exists || !jsonEqual(oldVal, newVal) {
			diff[key] = map[string]interface{}{
				"old": oldVal,
				"new": newVal,
			}
		}
	}
	if len(diff) == 0 {
		return "{}"
	}
	data, _ := json.Marshal(diff)
	return string(data)
}

// shouldIgnoreField checks if a JSON key (snake_case) matches any ignored
// Go field name (PascalCase).
func shouldIgnoreField(jsonKey string, ignore map[string]bool) bool {
	for field := range ignore {
		if toSnakeCaseAudit(field) == jsonKey {
			return true
		}
	}
	return false
}

// jsonEqual compares two values that came from JSON deserialization.
func jsonEqual(a, b interface{}) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return string(aj) == string(bj)
}

// structToMapAudit converts a struct (or pointer to struct) to a map
// using JSON serialization for consistent type normalization.
func structToMapAudit(v interface{}) map[string]interface{} {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	data, err := json.Marshal(val.Interface())
	if err != nil {
		return nil
	}
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

// snapshotForAudit creates a JSON snapshot of an entity, excluding ignored fields.
func snapshotForAudit(item interface{}, ignore map[string]bool) string {
	m := structToMapAudit(item)
	if m == nil {
		return "{}"
	}
	for field := range ignore {
		key := toSnakeCaseAudit(field)
		delete(m, key)
	}
	data, _ := json.Marshal(m)
	return string(data)
}

// extractEntityID extracts the ID field from a model instance.
func extractEntityID(item interface{}) uint {
	val := reflect.ValueOf(item)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	idField := val.FieldByName("ID")
	if idField.IsValid() && idField.CanUint() {
		return uint(idField.Uint())
	}
	return 0
}

// getClientIP extracts the client IP address from the request,
// respecting X-Forwarded-For and X-Real-IP headers.
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

// toSnakeCaseAudit converts PascalCase to snake_case.
func toSnakeCaseAudit(s string) string {
	var result []byte
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, byte(c)+32) // toLower
		} else {
			result = append(result, byte(c))
		}
	}
	return string(result)
}
