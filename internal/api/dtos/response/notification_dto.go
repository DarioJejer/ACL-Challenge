package response

import "time"

type NotificationDTO struct {
	ID        string    `json:"id"`
	Recipient string    `json:"recipient"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Channel   string    `json:"channel"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
