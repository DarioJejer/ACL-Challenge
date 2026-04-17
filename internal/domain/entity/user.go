package entity

import "time"

// User is the canonical domain representation of an application user.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
