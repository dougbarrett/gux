package dto

import (
	"time"
)

// ActivityLogList is the DTO for activity log list responses.
// Includes UserName derived from the User relation via preload.
type ActivityLogList struct {
	ID          uint       `json:"id" gux:"ActivityLog.ID"`
	UserID      *uint      `json:"user_id,omitempty" gux:"ActivityLog.UserID"`
	UserName    string     `json:"user_name,omitempty"`   // Derived from User.Name via preload
	Action      string     `json:"action" gux:"ActivityLog.Action"`
	Entity      string     `json:"entity" gux:"ActivityLog.Entity"`
	EntityID    *uint      `json:"entity_id,omitempty" gux:"ActivityLog.EntityID"`
	Description string     `json:"description" gux:"ActivityLog.Description"`
	CreatedAt   time.Time  `json:"created_at" gux:"ActivityLog.CreatedAt"`
}
