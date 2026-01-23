package models

import (
	"gorm.io/gorm"
)

// ActivityLog represents an audit trail entry for admin actions.
type ActivityLog struct {
	gorm.Model
	UserID      *uint   `json:"user_id"`                        // Nullable for system events
	User        *User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Action      string  `json:"action"`                         // created, updated, deleted, login, logout
	Entity      string  `json:"entity"`                         // user, setting
	EntityID    *uint   `json:"entity_id,omitempty"`            // ID of affected entity
	Description string  `json:"description"`                    // Human-readable description
	IPAddress   string  `json:"ip_address,omitempty"`           // Client IP address
	Metadata    string  `json:"metadata,omitempty"`             // Optional JSON metadata
}
