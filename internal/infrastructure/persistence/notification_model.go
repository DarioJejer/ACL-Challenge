package persistence

import (
	"time"

	uuid "github.com/google/uuid"
)

type NotificationModel struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Recipient uuid.UUID `gorm:"column:recipient;type:uuid;not null;index"`
	Title     string    `gorm:"not null"`
	Content   string    `gorm:"not null"`
	Channel   string    `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	User      UserModel `gorm:"foreignKey:Recipient;references:ID"`
}

func (NotificationModel) TableName() string {
	return "notifications"
}
