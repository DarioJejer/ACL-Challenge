package persistence

import (
	"context"
	"errors"
	"fmt"

	"acl-challenge/internal/domain/entity"
	domainrepo "acl-challenge/internal/domain/repository"
	"acl-challenge/internal/usecase"

	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) domainrepo.INotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *entity.Notification) error {
	model := toNotificationModel(*n)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("notification repository create: %w", usecase.ErrDatabase)
	}

	*n = toNotificationDomain(model)
	return nil
}

func (r *notificationRepository) FindByID(ctx context.Context, id string) (*entity.Notification, error) {
	var model NotificationModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, usecase.ErrNotFound
		}
		return nil, fmt.Errorf("notification repository find by id: %w", usecase.ErrDatabase)
	}

	domain := toNotificationDomain(model)
	return &domain, nil
}

func (r *notificationRepository) FindAllByUserID(ctx context.Context, userID string) ([]entity.Notification, error) {
	var models []NotificationModel
	if err := r.db.WithContext(ctx).Where("recipient = ?", userID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("notification repository find all by user id: %w", usecase.ErrDatabase)
	}

	notifications := make([]entity.Notification, 0, len(models))
	for _, model := range models {
		notifications = append(notifications, toNotificationDomain(model))
	}

	return notifications, nil
}

func (r *notificationRepository) Update(ctx context.Context, n *entity.Notification) error {
	model := toNotificationModel(*n)
	result := r.db.WithContext(ctx).
		Model(&NotificationModel{}).
		Where("id = ?", n.ID).
		Updates(map[string]interface{}{
			"recipient":  model.Recipient,
			"title":      model.Title,
			"content":    model.Content,
			"channel":    model.Channel,
			"updated_at": model.UpdatedAt,
		})
	if result.Error != nil {
		return fmt.Errorf("notification repository update: %w", usecase.ErrDatabase)
	}
	if result.RowsAffected == 0 {
		return usecase.ErrNotFound
	}

	return nil
}

func (r *notificationRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&NotificationModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("notification repository delete: %w", usecase.ErrDatabase)
	}
	if result.RowsAffected == 0 {
		return usecase.ErrNotFound
	}

	return nil
}

func toNotificationModel(notification entity.Notification) NotificationModel {
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

func toNotificationDomain(model NotificationModel) entity.Notification {
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
