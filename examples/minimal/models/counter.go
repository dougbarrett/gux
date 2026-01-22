package models

import "gorm.io/gorm"

// Counter represents a named counter with a value.
type Counter struct {
	gorm.Model
	Name  string `json:"name"`
	Value int    `json:"value"`
}
