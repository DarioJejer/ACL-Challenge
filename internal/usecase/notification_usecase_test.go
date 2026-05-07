package usecase_test

import (
	"errors"
	"strings"
	"testing"

	"acl-challenge/internal/api/dtos/request"
	"acl-challenge/internal/domain/entity"
	sendermocks "acl-challenge/internal/domain/notification/mocks"
	repomocks "acl-challenge/internal/domain/repository/mocks"
	notificationinfra "acl-challenge/internal/infrastructure/notification"
	"acl-challenge/internal/usecase"
	"acl-challenge/tests/testhelper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNotificationUseCase_ListNotifications(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		repoErr       error
		wantErrTarget error
	}{
		{name: "happy path"},
		{name: "not found", repoErr: gorm.ErrRecordNotFound, wantErrTarget: usecase.ErrNotFound},
		{name: "database error", repoErr: errors.New("find failed"), wantErrTarget: usecase.ErrDatabase},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			userRepo := repomocks.NewMockUserRepository(t)
			notifRepo := repomocks.NewMockNotificationRepository(t)
			sender := sendermocks.NewMockSender(t)
			reg := notificationinfra.SenderRegistry{entity.ChannelEmail: sender}

			uc := usecase.NewNotificationUseCase(userRepo, notifRepo, reg)
			ctx, cancel := testhelper.NewContextWithTimeout()
			defer cancel()

			expected := []entity.Notification{{ID: uuid.New()}}
			notifRepo.EXPECT().FindAllByUserID(ctx, "user-id").Return(expected, tt.repoErr).Once()

			got, err := uc.ListNotifications(ctx, "user-id")
			if tt.wantErrTarget != nil {
				require.Nil(t, got)
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}

			require.NoError(t, err)
			require.Len(t, got, 1)
		})
	}
}

func TestNotificationUseCase_GetNotification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		repoErr       error
		wantErrTarget error
	}{
		{name: "happy path"},
		{name: "not found", repoErr: gorm.ErrRecordNotFound, wantErrTarget: usecase.ErrNotFound},
		{name: "database error", repoErr: errors.New("query failed"), wantErrTarget: usecase.ErrDatabase},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			userRepo := repomocks.NewMockUserRepository(t)
			notifRepo := repomocks.NewMockNotificationRepository(t)
			sender := sendermocks.NewMockSender(t)
			reg := notificationinfra.SenderRegistry{entity.ChannelEmail: sender}
			uc := usecase.NewNotificationUseCase(userRepo, notifRepo, reg)
			ctx, cancel := testhelper.NewContextWithTimeout()
			defer cancel()

			expected := &entity.Notification{ID: uuid.New()}
			notifRepo.EXPECT().FindByID(ctx, "notif-id").Return(expected, tt.repoErr).Once()

			got, err := uc.GetNotification(ctx, "notif-id")
			if tt.wantErrTarget != nil {
				require.Nil(t, got)
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}

func TestNotificationUseCase_CreateNotification(t *testing.T) {
	t.Parallel()

	validInput := request.ResquestNotificationDTO{
		Recipient: uuid.NewString(),
		Title:     "Title",
		Content:   "Content",
		Channel:   entity.ChannelEmail,
	}

	tests := []struct {
		name               string
		input              request.ResquestNotificationDTO
		userFindErr        error
		repoCreateErr      error
		senderSendErr      error
		customRegistry     notificationinfra.SenderRegistry
		wantErrTarget      error
		expectNotification bool
	}{
		{
			name:          "invalid input",
			input:         request.ResquestNotificationDTO{},
			wantErrTarget: usecase.ErrInvalidInput,
		},
		{
			name:          "user not found",
			input:         validInput,
			userFindErr:   gorm.ErrRecordNotFound,
			wantErrTarget: usecase.ErrNotFound,
		},
		{
			name:          "unsupported channel",
			input:         request.ResquestNotificationDTO{Recipient: uuid.NewString(), Title: "T", Content: "C", Channel: entity.Channel("fax")},
			wantErrTarget: usecase.ErrUnsupportedChannel,
		},
		{
			name:          "repository database error",
			input:         validInput,
			repoCreateErr: errors.New("insert failed"),
			wantErrTarget: usecase.ErrDatabase,
		},
		{
			name:               "dispatch failure still succeeds",
			input:              validInput,
			senderSendErr:      errors.New("smtp timeout"),
			expectNotification: true,
		},
		{
			name:               "happy path",
			input:              validInput,
			expectNotification: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			userRepo := repomocks.NewMockUserRepository(t)
			notifRepo := repomocks.NewMockNotificationRepository(t)
			sender := sendermocks.NewMockSender(t)
			reg := notificationinfra.SenderRegistry{entity.ChannelEmail: sender}
			if tt.customRegistry != nil {
				reg = tt.customRegistry
			}

			uc := usecase.NewNotificationUseCase(userRepo, notifRepo, reg)
			ctx, cancel := testhelper.NewContextWithTimeout()
			defer cancel()

			if tt.wantErrTarget != usecase.ErrInvalidInput {
				userRepo.EXPECT().FindByID(ctx, tt.input.Recipient).Return(&entity.User{ID: uuid.MustParse(tt.input.Recipient)}, tt.userFindErr).Once()
			}
			if tt.userFindErr == nil && tt.wantErrTarget != usecase.ErrInvalidInput && tt.wantErrTarget != usecase.ErrUnsupportedChannel {
				notifRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entity.Notification")).Return(tt.repoCreateErr).Once()
			}
			if tt.userFindErr == nil && tt.repoCreateErr == nil && tt.wantErrTarget != usecase.ErrInvalidInput && tt.wantErrTarget != usecase.ErrUnsupportedChannel {
				sender.EXPECT().Send(ctx, mock.AnythingOfType("*entity.Notification")).Return(tt.senderSendErr).Once()
			}

			got, err := uc.CreateNotification(ctx, tt.input)
			if tt.wantErrTarget != nil {
				require.Nil(t, got)
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}

			require.NoError(t, err)
			if tt.expectNotification {
				require.NotNil(t, got)
			}
		})
	}
}

func TestNotificationUseCase_UpdateNotification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         request.ResquestNotificationDTO
		findErr       error
		userFindErr   error
		updateErr     error
		reloadErr     error
		wantErrTarget error
	}{
		{
			name:          "invalid input",
			input:         request.ResquestNotificationDTO{},
			wantErrTarget: usecase.ErrInvalidInput,
		},
		{
			name:          "not found",
			input:         request.ResquestNotificationDTO{Title: "New"},
			findErr:       gorm.ErrRecordNotFound,
			wantErrTarget: usecase.ErrNotFound,
		},
		{
			name:          "database error on update",
			input:         request.ResquestNotificationDTO{Title: "New"},
			updateErr:     errors.New("update failed"),
			wantErrTarget: usecase.ErrDatabase,
		},
		{
			name:          "unsupported channel",
			input:         request.ResquestNotificationDTO{Channel: entity.Channel("fax")},
			wantErrTarget: usecase.ErrUnsupportedChannel,
		},
		{
			name: "happy path",
			input: request.ResquestNotificationDTO{
				Recipient: uuid.NewString(),
				Title:     "Updated",
				Content:   "Updated content",
				Channel:   entity.ChannelSMS,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			userRepo := repomocks.NewMockUserRepository(t)
			notifRepo := repomocks.NewMockNotificationRepository(t)
			sender := sendermocks.NewMockSender(t)
			reg := notificationinfra.SenderRegistry{entity.ChannelEmail: sender}
			uc := usecase.NewNotificationUseCase(userRepo, notifRepo, reg)
			ctx, cancel := testhelper.NewContextWithTimeout()
			defer cancel()

			id := uuid.New()
			initial := &entity.Notification{ID: id, Recipient: uuid.New(), Title: "Old", Content: "Old", Channel: entity.ChannelEmail}
			reloaded := &entity.Notification{ID: id, Recipient: initial.Recipient, Title: "Updated", Content: "Updated content", Channel: entity.ChannelSMS}

			if tt.wantErrTarget == usecase.ErrInvalidInput {
				got, err := uc.UpdateNotification(ctx, id.String(), tt.input)
				require.Nil(t, got)
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}

			notifRepo.EXPECT().FindByID(ctx, id.String()).Return(initial, tt.findErr).Once()
			if tt.findErr == nil && strings.TrimSpace(tt.input.Recipient) != "" {
				userRepo.EXPECT().FindByID(ctx, tt.input.Recipient).Return(&entity.User{ID: uuid.MustParse(tt.input.Recipient)}, tt.userFindErr).Once()
			}
			if tt.findErr == nil && tt.userFindErr == nil && tt.wantErrTarget != usecase.ErrUnsupportedChannel {
				notifRepo.EXPECT().Update(ctx, initial).Return(tt.updateErr).Once()
			}
			if tt.findErr == nil && tt.userFindErr == nil && tt.updateErr == nil && tt.wantErrTarget == nil {
				notifRepo.EXPECT().FindByID(ctx, id.String()).Return(reloaded, tt.reloadErr).Once()
			}

			got, err := uc.UpdateNotification(ctx, id.String(), tt.input)
			if tt.wantErrTarget != nil {
				require.Nil(t, got)
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)

		})
	}
}

func TestNotificationUseCase_DeleteNotification(t *testing.T) {
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			userRepo := repomocks.NewMockUserRepository(t)
			notifRepo := repomocks.NewMockNotificationRepository(t)
			sender := sendermocks.NewMockSender(t)
			reg := notificationinfra.SenderRegistry{entity.ChannelEmail: sender}
			uc := usecase.NewNotificationUseCase(userRepo, notifRepo, reg)
			ctx, cancel := testhelper.NewContextWithTimeout()
			defer cancel()

			notifRepo.EXPECT().Delete(ctx, "notif-id").Return(tt.repoErr).Once()

			err := uc.DeleteNotification(ctx, "notif-id")
			if tt.wantErrTarget != nil {
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}
			require.NoError(t, err)
		})
	}
}

