package models

import "gorm.io/gorm"

// Post represents a blog post written by a user.
type Post struct {
	gorm.Model
	Title   string `json:"title"`
	Content string `json:"content"`
	UserID  uint   `json:"user_id"`
	User    *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
