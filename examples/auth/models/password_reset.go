package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
)

// PasswordReset stores password reset tokens.
type PasswordReset struct {
	gorm.Model
	UserID    uint      `json:"user_id"`
	Token     string    `json:"token" gorm:"uniqueIndex"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used" gorm:"default:false"`
}

// GenerateToken creates a secure random token and sets expiry.
func (pr *PasswordReset) GenerateToken() error {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	pr.Token = hex.EncodeToString(bytes)
	pr.ExpiresAt = time.Now().Add(1 * time.Hour) // 1 hour expiry
	return nil
}

// IsValid checks if token is not expired and not used.
func (pr *PasswordReset) IsValid() bool {
	return !pr.Used && time.Now().Before(pr.ExpiresAt)
}
