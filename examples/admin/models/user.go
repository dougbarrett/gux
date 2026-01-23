package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents an admin panel user account.
type User struct {
	gorm.Model
	Email        string     `json:"email" gorm:"uniqueIndex"`
	Name         string     `json:"name"`
	PasswordHash string     `json:"-"` // Never expose via JSON
	Role         string     `json:"role" gorm:"default:user"`
	Status       string     `json:"status" gorm:"default:active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
}

// SetPassword hashes and sets the user's password.
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies if the provided password matches the stored hash.
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}
