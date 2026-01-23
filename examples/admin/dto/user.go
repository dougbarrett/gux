package dto

import (
	"time"
)

// UserList is the DTO for user list responses.
// Excludes PasswordHash for security.
type UserList struct {
	ID          uint       `json:"id" gux:"User.ID"`
	Email       string     `json:"email" gux:"User.Email"`
	Name        string     `json:"name" gux:"User.Name"`
	Role        string     `json:"role" gux:"User.Role"`
	Status      string     `json:"status" gux:"User.Status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" gux:"User.LastLoginAt"`
	CreatedAt   time.Time  `json:"created_at" gux:"User.CreatedAt"`
}

// UserDetail is the DTO for single user responses.
// Includes all list fields plus UpdatedAt.
type UserDetail struct {
	ID          uint       `json:"id" gux:"User.ID"`
	Email       string     `json:"email" gux:"User.Email"`
	Name        string     `json:"name" gux:"User.Name"`
	Role        string     `json:"role" gux:"User.Role"`
	Status      string     `json:"status" gux:"User.Status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" gux:"User.LastLoginAt"`
	CreatedAt   time.Time  `json:"created_at" gux:"User.CreatedAt"`
	UpdatedAt   time.Time  `json:"updated_at" gux:"User.UpdatedAt"`
}

// UserCreate is the DTO for creating a new user.
type UserCreate struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"` // Plain text, will be hashed
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// UserUpdate is the DTO for updating a user.
type UserUpdate struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password,omitempty"` // Optional, only set if changing
	Role     string `json:"role"`
	Status   string `json:"status"`
}
