package usecase_test

import (
	"context"
	"errors"
	"testing"

	"acl-challenge/internal/api/dtos/request"
	"acl-challenge/internal/domain/entity"
	repomocks "acl-challenge/internal/domain/repository/mocks"
	"acl-challenge/internal/usecase"
	"acl-challenge/tests/testhelper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestUserUseCase_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          usecase.RegisterInput
		findByEmail    *entity.User
		findByEmailErr error
		createErr      error
		wantErrTarget  error
	}{
		{
			name:          "invalid input empty fields",
			input:         usecase.RegisterInput{},
			wantErrTarget: usecase.ErrInvalidInput,
		},
		{
			name:          "invalid email format",
			input:         usecase.RegisterInput{Email: "not-an-email", Password: "password123"},
			wantErrTarget: usecase.ErrInvalidInput,
		},
		{
			name:          "email already exists",
			input:         usecase.RegisterInput{Email: "existing@example.com", Password: "password123"},
			findByEmail:   &entity.User{ID: uuid.New(), Email: "existing@example.com"},
			wantErrTarget: usecase.ErrConflict,
		},
		{
			name:           "find by email database error",
			input:          usecase.RegisterInput{Email: "new@example.com", Password: "password123"},
			findByEmailErr: errors.New("db down"),
			wantErrTarget:  usecase.ErrDatabase,
		},
		{
			name:           "create database error",
			input:          usecase.RegisterInput{Email: "new@example.com", Password: "password123"},
			findByEmailErr: gorm.ErrRecordNotFound,
			createErr:      errors.New("insert failed"),
			wantErrTarget:  usecase.ErrDatabase,
		},
		{
			name:           "happy path",
			input:          usecase.RegisterInput{Email: "new@example.com", Password: "password123"},
			findByEmailErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := repomocks.NewMockUserRepository(t)
			uc := usecase.NewUserUseCase(repo)
			ctx, cancel := testhelper.NewContextWithTimeout()
			defer cancel()

			if tt.wantErrTarget == usecase.ErrInvalidInput {
				got, err := uc.Register(ctx, tt.input)
				require.Nil(t, got)
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}

			repo.EXPECT().FindByEmail(ctx, tt.input.Email).Return(tt.findByEmail, tt.findByEmailErr).Once()
			if errors.Is(tt.findByEmailErr, gorm.ErrRecordNotFound) {
				repo.EXPECT().Create(ctx, mock.AnythingOfType("*entity.User")).RunAndReturn(func(_ context.Context, user *entity.User) error {
					if tt.createErr != nil {
						return tt.createErr
					}
					return nil
				}).Once()
			}

			got, err := uc.Register(ctx, tt.input)

			if tt.wantErrTarget != nil {
				require.Nil(t, got)
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.NotEqual(t, uuid.Nil, got.ID)
			require.Equal(t, tt.input.Email, got.Email)
			require.NotEmpty(t, got.PasswordHash)
			require.NotEqual(t, tt.input.Password, got.PasswordHash)
			require.NoError(t, bcrypt.CompareHashAndPassword([]byte(got.PasswordHash), []byte(tt.input.Password)))
		})
	}
}

func TestUserUseCase_GetUser(t *testing.T) {
	t.Parallel()

	baseUser := &entity.User{ID: uuid.New(), Email: "user@example.com"}

	tests := []struct {
		name          string
		repoErr       error
		repoUser      *entity.User
		wantErrTarget error
	}{
		{name: "happy path", repoUser: baseUser},
		{name: "not found", repoErr: gorm.ErrRecordNotFound, wantErrTarget: usecase.ErrNotFound},
		{name: "database error", repoErr: errors.New("db down"), wantErrTarget: usecase.ErrDatabase},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := repomocks.NewMockUserRepository(t)
			uc := usecase.NewUserUseCase(repo)
			ctx, cancel := testhelper.NewContextWithTimeout()
			defer cancel()

			repo.EXPECT().FindByID(ctx, "user-id").Return(tt.repoUser, tt.repoErr).Once()

			got, err := uc.GetUser(ctx, "user-id")

			if tt.wantErrTarget != nil {
				require.Nil(t, got)
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, tt.repoUser.ID, got.ID)
		})
	}
}

func TestUserUseCase_UpdateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         request.ResquestUserDTO
		findErr       error
		updateErr     error
		reloadErr     error
		wantErrTarget error
	}{
		{
			name:          "invalid input",
			input:         request.ResquestUserDTO{},
			wantErrTarget: usecase.ErrInvalidInput,
		},
		{
			name:          "not found",
			input:         request.ResquestUserDTO{Email: "updated@example.com"},
			findErr:       gorm.ErrRecordNotFound,
			wantErrTarget: usecase.ErrNotFound,
		},
		{
			name:          "database error on update",
			input:         request.ResquestUserDTO{Email: "updated@example.com"},
			updateErr:     errors.New("update failed"),
			wantErrTarget: usecase.ErrDatabase,
		},
		{
			name:          "happy path",
			input:         request.ResquestUserDTO{Email: "updated@example.com"},
			wantErrTarget: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := repomocks.NewMockUserRepository(t)
			uc := usecase.NewUserUseCase(repo)
			ctx, cancel := testhelper.NewContextWithTimeout()
			defer cancel()

			initial := &entity.User{ID: uuid.New(), Email: "old@example.com", PasswordHash: "hash"}
			updated := &entity.User{ID: initial.ID, Email: "updated@example.com", PasswordHash: "hash"}

			if tt.wantErrTarget == usecase.ErrInvalidInput {
				got, err := uc.UpdateUser(ctx, initial.ID.String(), tt.input)
				require.Nil(t, got)
				testhelper.AssertErrorIs(t, err, usecase.ErrInvalidInput)
				return
			}

			repo.EXPECT().FindByID(ctx, initial.ID.String()).Return(initial, tt.findErr).Once()
			if tt.findErr == nil {
				repo.EXPECT().Update(ctx, initial).Return(tt.updateErr).Once()
			}
			if tt.findErr == nil && tt.updateErr == nil {
				repo.EXPECT().FindByID(ctx, initial.ID.String()).Return(updated, tt.reloadErr).Once()
			}

			got, err := uc.UpdateUser(ctx, initial.ID.String(), tt.input)

			if tt.wantErrTarget != nil {
				require.Nil(t, got)
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, updated.Email, got.Email)
		})
	}
}

func TestUserUseCase_DeleteUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		repoErr       error
		wantErrTarget error
	}{
		{name: "happy path"},
		{name: "not found", repoErr: gorm.ErrRecordNotFound, wantErrTarget: usecase.ErrNotFound},
		{name: "database error", repoErr: errors.New("delete failed"), wantErrTarget: usecase.ErrDatabase},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := repomocks.NewMockUserRepository(t)
			uc := usecase.NewUserUseCase(repo)
			ctx, cancel := testhelper.NewContextWithTimeout()
			defer cancel()

			repo.EXPECT().Delete(ctx, "user-id").Return(tt.repoErr).Once()

			err := uc.DeleteUser(ctx, "user-id")

			if tt.wantErrTarget != nil {
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}

			require.NoError(t, err)
		})
	}
}

