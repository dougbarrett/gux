package dto

import (
	"time"
)

// UserList is the DTO for user list responses.
// Excludes sensitive fields like PasswordHash.
// The gux tags map DTO fields to model fields for automatic code generation.
type UserList struct {
	ID        uint      `json:"id" gux:"User.ID"`
	Email     string    `json:"email" gux:"User.Email"`
	Name      string    `json:"name" gux:"User.Name"`
	Role      string    `json:"role" gux:"User.Role"`
	CreatedAt time.Time `json:"created_at" gux:"User.CreatedAt"`
}

// UserDetail is the DTO for single user responses.
// Includes related posts but excludes PasswordHash.
type UserDetail struct {
	ID        uint        `json:"id" gux:"User.ID"`
	Email     string      `json:"email" gux:"User.Email"`
	Name      string      `json:"name" gux:"User.Name"`
	Role      string      `json:"role" gux:"User.Role"`
	CreatedAt time.Time   `json:"created_at" gux:"User.CreatedAt"`
	UpdatedAt time.Time   `json:"updated_at" gux:"User.UpdatedAt"`
	Posts     []PostBrief `json:"posts,omitempty" gux:"User.Posts" preload:"Posts"`
}

// PostBrief is a simplified post DTO for embedding in UserDetail.
type PostBrief struct {
	ID        uint      `json:"id" gux:"Post.ID"`
	Title     string    `json:"title" gux:"Post.Title"`
	CreatedAt time.Time `json:"created_at" gux:"Post.CreatedAt"`
}

// UserCreate is the DTO for creating a new user.
type UserCreate struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"` // Plain text, will be hashed
	Role     string `json:"role"`
}

// UserUpdate is the DTO for updating a user.
type UserUpdate struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password,omitempty"` // Optional, only set if changing
	Role     string `json:"role"`
}
