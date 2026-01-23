package dto

import (
	"time"
)

// ProjectList is the DTO for project list responses.
// Used in list views showing summary project information.
type ProjectList struct {
	ID        uint      `json:"id" gux:"Project.ID"`
	Name      string    `json:"name" gux:"Project.Name"`
	Status    string    `json:"status" gux:"Project.Status"`
	CreatedAt time.Time `json:"created_at" gux:"Project.CreatedAt"`
}

// ProjectDetail is the DTO for single project responses.
// Used in detail views showing full project information.
type ProjectDetail struct {
	ID          uint      `json:"id" gux:"Project.ID"`
	Name        string    `json:"name" gux:"Project.Name"`
	Description string    `json:"description" gux:"Project.Description"`
	Status      string    `json:"status" gux:"Project.Status"`
	CreatedAt   time.Time `json:"created_at" gux:"Project.CreatedAt"`
	UpdatedAt   time.Time `json:"updated_at" gux:"Project.UpdatedAt"`
}
