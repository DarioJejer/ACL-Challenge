package handler

import (
	apireponse "acl-challenge/internal/api/dtos/response"
	"acl-challenge/internal/domain/entity"
)

func toUserDTO(user *entity.User) apireponse.UserDTO {
	return apireponse.UserDTO{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func toNotificationDTO(notification *entity.Notification) apireponse.NotificationDTO {
	return apireponse.NotificationDTO{
		ID:        notification.ID.String(),
		Recipient: notification.Recipient.String(),
		Title:     notification.Title,
		Content:   notification.Content,
		Channel:   string(notification.Channel),
		CreatedAt: notification.CreatedAt,
		UpdatedAt: notification.UpdatedAt,
	}
}

func toNotificationDTOList(notifications []entity.Notification) []apireponse.NotificationDTO {
	items := make([]apireponse.NotificationDTO, 0, len(notifications))
	for i := range notifications {
		n := notifications[i]
		items = append(items, apireponse.NotificationDTO{
			ID:        n.ID.String(),
			Recipient: n.Recipient.String(),
			Title:     n.Title,
			Content:   n.Content,
			Channel:   string(n.Channel),
			CreatedAt: n.CreatedAt,
			UpdatedAt: n.UpdatedAt,
		})
	}
	return items
}
