package repository

import (
	"context"

	"acl-challenge/internal/domain/entity"
)

//go:generate mockery --config ../../../.mockery.yaml
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
}
