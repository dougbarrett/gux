package models

import "gorm.io/gorm"

type Document struct {
	gorm.Model
	Title    string `json:"title"`
	Content  string `json:"content" gorm:"type:text"`
	Status   string `json:"status" gorm:"default:draft"`
	AuthorID uint   `json:"author_id"`
	Author   User   `json:"author" gorm:"foreignKey:AuthorID"`
}
