package persistence

import "acl-challenge/internal/domain/entity"

func ToUserModel(user entity.User) UserModel {
	return UserModel{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func ToUserEntity(model UserModel) entity.User {
	return entity.User{
		ID:           model.ID,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func ToNotificationModel(notification entity.Notification) NotificationModel {
	return NotificationModel{
		ID:        notification.ID,
		Recipient: notification.Recipient,
		Title:     notification.Title,
		Content:   notification.Content,
		Channel:   string(notification.Channel),
		CreatedAt: notification.CreatedAt,
		UpdatedAt: notification.UpdatedAt,
	}
}

func ToNotificationEntity(model NotificationModel) entity.Notification {
	return entity.Notification{
		ID:        model.ID,
		Recipient: model.Recipient,
		Title:     model.Title,
		Content:   model.Content,
		Channel:   entity.Channel(model.Channel),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
