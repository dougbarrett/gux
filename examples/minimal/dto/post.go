package dto

import (
	"time"
)

// PostList is the DTO for post list responses.
type PostList struct {
	ID        uint      `json:"id" gux:"Post.ID"`
	Title     string    `json:"title" gux:"Post.Title"`
	Content   string    `json:"content" gux:"Post.Content"`
	UserID    uint      `json:"user_id" gux:"Post.UserID"`
	Author    UserBrief `json:"author" gux:"Post.User" preload:"User"`
	CreatedAt time.Time `json:"created_at" gux:"Post.CreatedAt"`
}

// PostDetail is the DTO for single post responses with author info.
type PostDetail struct {
	ID        uint       `json:"id" gux:"Post.ID"`
	Title     string     `json:"title" gux:"Post.Title"`
	Content   string     `json:"content" gux:"Post.Content"`
	UserID    uint       `json:"user_id" gux:"Post.UserID"`
	Author    UserBrief  `json:"author" gux:"Post.User" preload:"User"`
	CreatedAt time.Time  `json:"created_at" gux:"Post.CreatedAt"`
	UpdatedAt time.Time  `json:"updated_at" gux:"Post.UpdatedAt"`
}

// UserBrief is a simplified user DTO for embedding in PostDetail.
type UserBrief struct {
	ID   uint   `json:"id" gux:"User.ID"`
	Name string `json:"name" gux:"User.Name"`
}

// PostCreate is the DTO for creating a new post.
type PostCreate struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	UserID  uint   `json:"user_id"`
}

// PostUpdate is the DTO for updating a post.
type PostUpdate struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
