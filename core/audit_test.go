package core

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestComputeFieldDiffFromMaps(t *testing.T) {
	before := map[string]interface{}{
		"name":  "Alice",
		"email": "alice@old.com",
		"role":  "user",
	}
	after := map[string]interface{}{
		"name":  "Alice",
		"email": "alice@new.com",
		"role":  "admin",
	}
	result := computeFieldDiffFromMaps(before, after, nil)

	var diff map[string]interface{}
	if err := json.Unmarshal([]byte(result), &diff); err != nil {
		t.Fatalf("failed to parse diff JSON: %v", err)
	}

	if _, ok := diff["name"]; ok {
		t.Error("name should not appear in diff (unchanged)")
	}
	if _, ok := diff["email"]; !ok {
		t.Error("email should appear in diff")
	}
	if _, ok := diff["role"]; !ok {
		t.Error("role should appear in diff")
	}

	emailDiff := diff["email"].(map[string]interface{})
	if emailDiff["old"] != "alice@old.com" {
		t.Errorf("expected old email alice@old.com, got %v", emailDiff["old"])
	}
	if emailDiff["new"] != "alice@new.com" {
		t.Errorf("expected new email alice@new.com, got %v", emailDiff["new"])
	}
}

func TestComputeFieldDiffFromMaps_IgnoreFields(t *testing.T) {
	before := map[string]interface{}{
		"name":         "Alice",
		"password_hash": "old-hash",
		"updated_at":    "2024-01-01",
	}
	after := map[string]interface{}{
		"name":         "Bob",
		"password_hash": "new-hash",
		"updated_at":    "2024-06-01",
	}
	ignore := map[string]bool{
		"PasswordHash": true,
		"UpdatedAt":    true,
	}
	result := computeFieldDiffFromMaps(before, after, ignore)

	var diff map[string]interface{}
	json.Unmarshal([]byte(result), &diff)

	if _, ok := diff["password_hash"]; ok {
		t.Error("password_hash should be ignored")
	}
	if _, ok := diff["updated_at"]; ok {
		t.Error("updated_at should be ignored")
	}
	if _, ok := diff["name"]; !ok {
		t.Error("name should appear in diff")
	}
}

func TestComputeFieldDiffFromMaps_NoChanges(t *testing.T) {
	m := map[string]interface{}{"name": "Alice", "age": float64(30)}
	result := computeFieldDiffFromMaps(m, m, nil)
	if result != "{}" {
		t.Errorf("expected empty diff, got %s", result)
	}
}

type testModel struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Score float64 `json:"score"`
}

func TestSnapshotForAudit(t *testing.T) {
	item := &testModel{ID: 1, Name: "Alice", Email: "alice@test.com", Score: 95.5}
	ignore := map[string]bool{"Email": true}
	result := snapshotForAudit(item, ignore)

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(result), &m); err != nil {
		t.Fatalf("failed to parse snapshot: %v", err)
	}

	if _, ok := m["email"]; ok {
		t.Error("email should be excluded from snapshot")
	}
	if m["name"] != "Alice" {
		t.Errorf("expected name=Alice, got %v", m["name"])
	}
	if m["id"].(float64) != 1 {
		t.Errorf("expected id=1, got %v", m["id"])
	}
}

func TestSnapshotForAudit_NilIgnore(t *testing.T) {
	item := &testModel{ID: 1, Name: "Bob"}
	result := snapshotForAudit(item, nil)

	var m map[string]interface{}
	json.Unmarshal([]byte(result), &m)

	if m["name"] != "Bob" {
		t.Errorf("expected name=Bob, got %v", m["name"])
	}
}

func TestExtractEntityID(t *testing.T) {
	item := &testModel{ID: 42}
	id := extractEntityID(item)
	if id != 42 {
		t.Errorf("expected ID=42, got %d", id)
	}
}

func TestExtractEntityID_NonPointer(t *testing.T) {
	item := testModel{ID: 99}
	id := extractEntityID(item)
	if id != 99 {
		t.Errorf("expected ID=99, got %d", id)
	}
}

func TestExtractEntityID_NoIDField(t *testing.T) {
	type noID struct {
		Name string
	}
	id := extractEntityID(&noID{Name: "test"})
	if id != 0 {
		t.Errorf("expected 0 for struct without ID, got %d", id)
	}
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "203.0.113.50")
	r.RemoteAddr = "10.0.0.1:12345"

	ip := getClientIP(r)
	if ip != "203.0.113.50" {
		t.Errorf("expected 203.0.113.50, got %s", ip)
	}
}

func TestGetClientIP_XForwardedFor_Multiple(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18, 150.172.238.178")

	ip := getClientIP(r)
	if ip != "203.0.113.50" {
		t.Errorf("expected first IP 203.0.113.50, got %s", ip)
	}
}

func TestGetClientIP_XRealIP(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("X-Real-IP", "198.51.100.10")
	r.RemoteAddr = "10.0.0.1:12345"

	ip := getClientIP(r)
	if ip != "198.51.100.10" {
		t.Errorf("expected 198.51.100.10, got %s", ip)
	}
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.168.1.100:54321"

	ip := getClientIP(r)
	if ip != "192.168.1.100" {
		t.Errorf("expected 192.168.1.100, got %s", ip)
	}
}

func TestWithAuditLog_DefaultIgnores(t *testing.T) {
	m := &CRUDModel{}
	opt := WithAuditLog()
	opt(m)

	if m.AuditConfig == nil {
		t.Fatal("expected AuditConfig to be set")
	}
	if !m.AuditConfig.Enabled {
		t.Error("expected Enabled=true")
	}
	if !m.AuditConfig.IgnoreFields["UpdatedAt"] {
		t.Error("expected UpdatedAt in default ignore list")
	}
	if !m.AuditConfig.IgnoreFields["DeletedAt"] {
		t.Error("expected DeletedAt in default ignore list")
	}
}

func TestWithAuditLog_CustomIgnores(t *testing.T) {
	m := &CRUDModel{}
	opt := WithAuditLog("PasswordHash", "Secret")
	opt(m)

	if !m.AuditConfig.IgnoreFields["PasswordHash"] {
		t.Error("expected PasswordHash in ignore list")
	}
	if !m.AuditConfig.IgnoreFields["Secret"] {
		t.Error("expected Secret in ignore list")
	}
	// Default ignores should still be present
	if !m.AuditConfig.IgnoreFields["UpdatedAt"] {
		t.Error("expected UpdatedAt in ignore list")
	}
}

func TestToSnakeCaseAudit(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PasswordHash", "password_hash"},
		{"UpdatedAt", "updated_at"},
		{"ID", "i_d"},
		{"Name", "name"},
		{"IPAddress", "i_p_address"},
		{"", ""},
	}
	for _, tt := range tests {
		result := toSnakeCaseAudit(tt.input)
		if result != tt.expected {
			t.Errorf("toSnakeCaseAudit(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestStructToMapAudit(t *testing.T) {
	item := testModel{ID: 1, Name: "Test", Email: "test@test.com", Score: 42.5}
	m := structToMapAudit(item)

	if m["name"] != "Test" {
		t.Errorf("expected name=Test, got %v", m["name"])
	}
	if m["email"] != "test@test.com" {
		t.Errorf("expected email=test@test.com, got %v", m["email"])
	}
	if m["score"].(float64) != 42.5 {
		t.Errorf("expected score=42.5, got %v", m["score"])
	}
}

func TestStructToMapAudit_Pointer(t *testing.T) {
	item := &testModel{ID: 2, Name: "Ptr"}
	m := structToMapAudit(item)

	if m["name"] != "Ptr" {
		t.Errorf("expected name=Ptr, got %v", m["name"])
	}
}
