package persistence

import (
	"context"
	"errors"
	"fmt"

	"acl-challenge/internal/domain/entity"
	domainrepo "acl-challenge/internal/domain/repository"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domainrepo.IUserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	model := ToUserModel(*user)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("repository error: user create failed: %w", err)
	}

	*user = ToUserEntity(model)
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("repository error: user not found: %w", err)
		}
		return nil, fmt.Errorf("repository error: user find by id failed: %w", err)
	}

	domain := ToUserEntity(model)
	return &domain, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var model UserModel
	if err := r.db.WithContext(ctx).First(&model, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("repository error: user not found: %w", err)
		}
		return nil, fmt.Errorf("repository error: user find by email failed: %w", err)
	}

	domain := ToUserEntity(model)
	return &domain, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	model := ToUserModel(*user)
	result := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"email":         model.Email,
			"password_hash": model.PasswordHash,
			"updated_at":    model.UpdatedAt,
		})
	if result.Error != nil {
		return fmt.Errorf("repository error: user update failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("repository error: user not found: %w", gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&UserModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("repository error: user delete failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("repository error: user not found: %w", gorm.ErrRecordNotFound)
	}

	return nil
}
