package entity

import (
	"time"

	"github.com/google/uuid"
)

// User is the canonical domain representation of an application user.
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
