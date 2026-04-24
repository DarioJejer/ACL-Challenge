package request

import "acl-challenge/internal/domain/entity"

type ResquestNotificationDTO struct {
	Recipient string         `json:"recipient"`
	Title     string         `json:"title"`
	Content   string         `json:"content"`
	Channel   entity.Channel `json:"channel"`
}
