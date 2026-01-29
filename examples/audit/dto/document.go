package dto

import "time"

type DocumentList struct {
	ID        uint      `json:"id" gux:"Document.ID"`
	Title     string    `json:"title" gux:"Document.Title"`
	Status    string    `json:"status" gux:"Document.Status"`
	AuthorID  uint      `json:"author_id" gux:"Document.AuthorID"`
	CreatedAt time.Time `json:"created_at" gux:"Document.CreatedAt"`
}

type DocumentDetail struct {
	ID        uint      `json:"id" gux:"Document.ID"`
	Title     string    `json:"title" gux:"Document.Title"`
	Content   string    `json:"content" gux:"Document.Content"`
	Status    string    `json:"status" gux:"Document.Status"`
	AuthorID  uint      `json:"author_id" gux:"Document.AuthorID"`
	CreatedAt time.Time `json:"created_at" gux:"Document.CreatedAt"`
	UpdatedAt time.Time `json:"updated_at" gux:"Document.UpdatedAt"`
}
