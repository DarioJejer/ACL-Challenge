package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"acl-challenge/internal/api/dtos/request"
	"acl-challenge/internal/domain/entity"
	"acl-challenge/internal/domain/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserUseCase struct {
	repo repository.UserRepository
}

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

func NewUserUseCase(repo repository.UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (uc *UserUseCase) Register(ctx context.Context, input RegisterInput) (*entity.User, error) {
	email := strings.TrimSpace(input.Email)
	password := strings.TrimSpace(input.Password)

	if email == "" || password == "" {
		return nil, fmt.Errorf("usecase: user: register: invalid input: %w", ErrInvalidInput)
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("usecase: user: register: invalid email format: %w", ErrInvalidInput)
	}

	_, err := uc.repo.FindByEmail(ctx, email)
	if err == nil {
		return nil, fmt.Errorf("usecase: user: register: email already exists: %w", ErrConflict)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, mapRepositoryError("usecase: user: register: find by email", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("usecase: user: register: hash password: %w", ErrInternalServer)
	}

	user := &entity.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(passwordHash),
	}

	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, mapRepositoryError("usecase: user: register: create", err)
	}

	return user, nil
}

// Login authenticates a user by email and password.
// Both "email not found" and "wrong password" return ErrUnauthorized so callers
// cannot distinguish between them and probe for valid emails.
func (uc *UserUseCase) Login(ctx context.Context, input LoginInput) (*entity.User, error) {
	email := strings.TrimSpace(input.Email)
	password := strings.TrimSpace(input.Password)

	if email == "" || password == "" {
		return nil, fmt.Errorf("usecase: user: login: invalid input: %w", ErrInvalidInput)
	}

	user, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("usecase: user: login: email not found: %w", ErrUnauthorized)
		}
		return nil, mapRepositoryError("usecase: user: login: find by email", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("usecase: user: login: password mismatch: %w", ErrUnauthorized)
	}

	return user, nil
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
