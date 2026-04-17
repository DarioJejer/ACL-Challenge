package persistence

import "time"

type UserModel struct {
	ID           string `gorm:"type:uuid;primaryKey"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
