package dto

import "time"

// AuditEntryList is the DTO for audit entry list responses.
type AuditEntryList struct {
	ID         uint      `json:"id" gux:"AuditEntry.ID"`
	CreatedAt  time.Time `json:"created_at" gux:"AuditEntry.CreatedAt"`
	Action     string    `json:"action" gux:"AuditEntry.Action"`
	EntityType string    `json:"entity_type" gux:"AuditEntry.EntityType"`
	EntityID   uint      `json:"entity_id" gux:"AuditEntry.EntityID"`
	UserID     string    `json:"user_id" gux:"AuditEntry.UserID"`
	UserEmail  string    `json:"user_email" gux:"AuditEntry.UserEmail"`
	IPAddress  string    `json:"ip_address" gux:"AuditEntry.IPAddress"`
	Changes    string    `json:"changes" gux:"AuditEntry.Changes"`
}
