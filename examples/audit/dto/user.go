package dto

import "time"

type UserList struct {
	ID        uint      `json:"id" gux:"User.ID"`
	Email     string    `json:"email" gux:"User.Email"`
	Name      string    `json:"name" gux:"User.Name"`
	Role      string    `json:"role" gux:"User.Role"`
	CreatedAt time.Time `json:"created_at" gux:"User.CreatedAt"`
}

type UserDetail struct {
	ID        uint      `json:"id" gux:"User.ID"`
	Email     string    `json:"email" gux:"User.Email"`
	Name      string    `json:"name" gux:"User.Name"`
	Role      string    `json:"role" gux:"User.Role"`
	CreatedAt time.Time `json:"created_at" gux:"User.CreatedAt"`
	UpdatedAt time.Time `json:"updated_at" gux:"User.UpdatedAt"`
}
