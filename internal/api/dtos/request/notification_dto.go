package request

import "acl-challenge/internal/domain/entity"

// ResquestNotificationDTO carries notification fields supplied by the client.
// The owning user is derived from the authenticated context, never from the
// request body.
type ResquestNotificationDTO struct {
	Title   string         `json:"title"`
	Content string         `json:"content"`
	Channel entity.Channel `json:"channel"`
}
