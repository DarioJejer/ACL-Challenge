package persistence

import "time"

type NotificationModel struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	Recipient string `gorm:"column:recipient;type:uuid;not null;index"`
	Title     string `gorm:"not null"`
	Content   string `gorm:"not null"`
	Channel   string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	User      UserModel `gorm:"foreignKey:Recipient;references:ID"`
}

func (NotificationModel) TableName() string {
	return "notifications"
}
