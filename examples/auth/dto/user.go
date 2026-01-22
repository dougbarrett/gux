package dto

import "time"

// UserDTO excludes password hash for API responses.
type UserDTO struct {
	ID        uint      `json:"id" gux:"User.ID"`
	Email     string    `json:"email" gux:"User.Email"`
	Name      string    `json:"name" gux:"User.Name"`
	Verified  bool      `json:"verified" gux:"User.Verified"`
	CreatedAt time.Time `json:"created_at" gux:"User.CreatedAt"`
}

// UserCreate is the DTO for creating a new user.
type UserCreate struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"` // Plain text, will be hashed
}

// UserUpdate is the DTO for updating a user.
type UserUpdate struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password,omitempty"` // Optional, only set if changing
}
