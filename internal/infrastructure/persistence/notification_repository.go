package persistence

import (
	"context"
	"errors"
	"fmt"

	"acl-challenge/internal/domain/entity"
	domainrepo "acl-challenge/internal/domain/repository"

	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) domainrepo.INotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *entity.Notification) error {
	model := ToNotificationModel(*n)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("repository error: notification create failed: %w", err)
	}

	*n = ToNotificationEntity(model)
	return nil
}

func (r *notificationRepository) FindByID(ctx context.Context, id string) (*entity.Notification, error) {
	var model NotificationModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("repository error: notification not found: %w", err)
		}
		return nil, fmt.Errorf("repository error: notification find by id failed: %w", err)
	}

	domain := ToNotificationEntity(model)
	return &domain, nil
}

func (r *notificationRepository) FindAllByUserID(ctx context.Context, userID string) ([]entity.Notification, error) {
	var models []NotificationModel
	if err := r.db.WithContext(ctx).Where("recipient = ?", userID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("repository error: notification find all by user id failed: %w", err)
	}

	notifications := make([]entity.Notification, 0, len(models))
	for _, model := range models {
		notifications = append(notifications, ToNotificationEntity(model))
	}

	return notifications, nil
}

func (r *notificationRepository) Update(ctx context.Context, n *entity.Notification) error {
	model := ToNotificationModel(*n)
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
		return fmt.Errorf("repository error: notification update failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("repository error: notification not found: %w", gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *notificationRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&NotificationModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("repository error: notification delete failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("repository error: notification not found: %w", gorm.ErrRecordNotFound)
	}

	return nil
}
