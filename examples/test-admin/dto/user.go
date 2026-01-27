package dto

import "time"

// UserDTO excludes password hash for API responses.
type UserDTO struct {
	ID        uint      `json:"id" gux:"User.ID"`
	Email     string    `json:"email" gux:"User.Email"`
	Name      string    `json:"name" gux:"User.Name"`
	Role      string    `json:"role" gux:"User.Role"`
	CreatedAt time.Time `json:"created_at" gux:"User.CreatedAt"`
}
