package models

import (
	"gorm.io/gorm"
)

// Project represents a project in the SaaS application.
type Project struct {
	gorm.Model
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Status      string `json:"status" gorm:"default:active"` // active, completed, archived
}
