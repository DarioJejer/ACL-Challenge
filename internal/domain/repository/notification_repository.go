package repository

import (
	"context"

	"acl-challenge/internal/domain/entity"
)

type INotificationRepository interface {
	Create(ctx context.Context, n *entity.Notification) error
	FindByID(ctx context.Context, id string) (*entity.Notification, error)
	FindAllByUserID(ctx context.Context, userID string) ([]entity.Notification, error)
	Update(ctx context.Context, n *entity.Notification) error
	Delete(ctx context.Context, id string) error
}
