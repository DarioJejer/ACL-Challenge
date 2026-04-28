package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"acl-challenge/internal/api/dtos/request"
	"acl-challenge/internal/domain/entity"
	"acl-challenge/internal/domain/repository"

	"gorm.io/gorm"
)

type UserUseCase struct {
	repo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*entity.User, error) {
	user, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepositoryError("usecase: user: get", err)
	}

	return user, nil
}

func (uc *UserUseCase) UpdateUser(ctx context.Context, id string, input request.ResquestUserDTO) (*entity.User, error) {
	if strings.TrimSpace(input.Email) == "" && strings.TrimSpace(input.PasswordHash) == "" {
		return nil, ErrInvalidInput
	}

	user, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepositoryError("usecase: user: update: find by id", err)
	}

	if strings.TrimSpace(input.Email) != "" {
		user.Email = input.Email
	}
	if strings.TrimSpace(input.PasswordHash) != "" {
		user.PasswordHash = input.PasswordHash
	}

	if err := uc.repo.Update(ctx, user); err != nil {
		return nil, mapRepositoryError("usecase: user: update", err)
	}

	updatedUser, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepositoryError("usecase: user: update: reload", err)
	}

	return updatedUser, nil
}

func (uc *UserUseCase) DeleteUser(ctx context.Context, id string) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return mapRepositoryError("usecase: user: delete", err)
	}

	return nil
}

func mapRepositoryError(context string, err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return fmt.Errorf("%s: %v classification: %w", context, err, ErrNotFound)
	default:
		return fmt.Errorf("%s: %v: classification: %w", context, err, ErrDatabase)
	}
}
