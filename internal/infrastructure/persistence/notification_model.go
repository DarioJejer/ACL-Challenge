package persistence

import "time"

type NotificationModel struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	UserID    string `gorm:"type:uuid;not null;index"`
	Title     string `gorm:"not null"`
	Content   string `gorm:"not null"`
	Channel   string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Recipient UserModel `gorm:"foreignKey:UserID"`
}
