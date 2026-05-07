package usecase_test

import (
	"errors"
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

func newNotificationUC(t *testing.T) (*usecase.NotificationUseCase, *repomocks.MockUserRepository, *repomocks.MockNotificationRepository, *sendermocks.MockSender) {
	t.Helper()
	userRepo := repomocks.NewMockUserRepository(t)
	notifRepo := repomocks.NewMockNotificationRepository(t)
	sender := sendermocks.NewMockSender(t)
	reg := notificationinfra.SenderRegistry{entity.ChannelEmail: sender, entity.ChannelSMS: sender}
	uc := usecase.NewNotificationUseCase(userRepo, notifRepo, reg)
	return uc, userRepo, notifRepo, sender
}

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
			uc, _, notifRepo, _ := newNotificationUC(t)
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

	owner := uuid.New()
	other := uuid.New()

	tests := []struct {
		name          string
		repoErr       error
		stored        *entity.Notification
		currentUserID string
		wantErrTarget error
	}{
		{
			name:          "happy path",
			stored:        &entity.Notification{ID: uuid.New(), Recipient: owner},
			currentUserID: owner.String(),
		},
		{
			name:          "forbidden when not owner",
			stored:        &entity.Notification{ID: uuid.New(), Recipient: owner},
			currentUserID: other.String(),
			wantErrTarget: usecase.ErrForbidden,
		},
		{
			name:          "not found",
			repoErr:       gorm.ErrRecordNotFound,
			currentUserID: owner.String(),
			wantErrTarget: usecase.ErrNotFound,
		},
		{
			name:          "database error",
			repoErr:       errors.New("query failed"),
			currentUserID: owner.String(),
			wantErrTarget: usecase.ErrDatabase,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, _, notifRepo, _ := newNotificationUC(t)
			ctx, cancel := testhelper.NewContextWithTimeout()
			defer cancel()

			notifRepo.EXPECT().FindByID(ctx, "notif-id").Return(tt.stored, tt.repoErr).Once()

			got, err := uc.GetNotification(ctx, "notif-id", tt.currentUserID)
			if tt.wantErrTarget != nil {
				require.Nil(t, got)
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, owner, got.Recipient)
		})
	}
}

func TestNotificationUseCase_CreateNotification(t *testing.T) {
	t.Parallel()

	currentUserID := uuid.NewString()
	validInput := request.ResquestNotificationDTO{
		Title:   "Title",
		Content: "Content",
		Channel: entity.ChannelEmail,
	}

	tests := []struct {
		name               string
		input              request.ResquestNotificationDTO
		currentUserID      string
		repoCreateErr      error
		senderSendErr      error
		wantErrTarget      error
		expectNotification bool
	}{
		{
			name:          "missing current user id",
			input:         validInput,
			currentUserID: "",
			wantErrTarget: usecase.ErrUnauthorized,
		},
		{
			name:          "invalid current user id",
			input:         validInput,
			currentUserID: "not-a-uuid",
			wantErrTarget: usecase.ErrInvalidInput,
		},
		{
			name:          "invalid input",
			input:         request.ResquestNotificationDTO{},
			currentUserID: currentUserID,
			wantErrTarget: usecase.ErrInvalidInput,
		},
		{
			name:          "unsupported channel",
			input:         request.ResquestNotificationDTO{Title: "T", Content: "C", Channel: entity.Channel("fax")},
			currentUserID: currentUserID,
			wantErrTarget: usecase.ErrUnsupportedChannel,
		},
		{
			name:          "repository database error",
			input:         validInput,
			currentUserID: currentUserID,
			repoCreateErr: errors.New("insert failed"),
			wantErrTarget: usecase.ErrDatabase,
		},
		{
			name:               "dispatch failure still succeeds",
			input:              validInput,
			currentUserID:      currentUserID,
			senderSendErr:      errors.New("smtp timeout"),
			expectNotification: true,
		},
		{
			name:               "happy path",
			input:              validInput,
			currentUserID:      currentUserID,
			expectNotification: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, _, notifRepo, sender := newNotificationUC(t)
			ctx, cancel := testhelper.NewContextWithTimeout()
			defer cancel()

			callsRepo := tt.wantErrTarget == nil ||
				tt.wantErrTarget == usecase.ErrDatabase
			if callsRepo {
				notifRepo.EXPECT().Create(ctx, mock.AnythingOfType("*entity.Notification")).Return(tt.repoCreateErr).Once()
			}
			if tt.wantErrTarget == nil && tt.repoCreateErr == nil {
				sender.EXPECT().Send(ctx, mock.AnythingOfType("*entity.Notification")).Return(tt.senderSendErr).Once()
			}

			got, err := uc.CreateNotification(ctx, tt.currentUserID, tt.input)
			if tt.wantErrTarget != nil {
				require.Nil(t, got)
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}

			require.NoError(t, err)
			if tt.expectNotification {
				require.NotNil(t, got)
				require.Equal(t, tt.currentUserID, got.Recipient.String())
			}
		})
	}
}

func TestNotificationUseCase_UpdateNotification(t *testing.T) {
	t.Parallel()

	owner := uuid.New()
	other := uuid.New()

	tests := []struct {
		name          string
		input         request.ResquestNotificationDTO
		stored        *entity.Notification
		currentUserID string
		findErr       error
		updateErr     error
		reloadErr     error
		wantErrTarget error
	}{
		{
			name:          "invalid input no fields",
			input:         request.ResquestNotificationDTO{},
			currentUserID: owner.String(),
			wantErrTarget: usecase.ErrInvalidInput,
		},
		{
			name:          "not found",
			input:         request.ResquestNotificationDTO{Title: "New"},
			currentUserID: owner.String(),
			findErr:       gorm.ErrRecordNotFound,
			wantErrTarget: usecase.ErrNotFound,
		},
		{
			name:          "forbidden when not owner",
			input:         request.ResquestNotificationDTO{Title: "New"},
			stored:        &entity.Notification{ID: uuid.New(), Recipient: owner, Title: "Old", Content: "Old", Channel: entity.ChannelEmail},
			currentUserID: other.String(),
			wantErrTarget: usecase.ErrForbidden,
		},
		{
			name:          "database error on update",
			input:         request.ResquestNotificationDTO{Title: "New"},
			stored:        &entity.Notification{ID: uuid.New(), Recipient: owner, Title: "Old", Content: "Old", Channel: entity.ChannelEmail},
			currentUserID: owner.String(),
			updateErr:     errors.New("update failed"),
			wantErrTarget: usecase.ErrDatabase,
		},
		{
			name:          "unsupported channel",
			input:         request.ResquestNotificationDTO{Channel: entity.Channel("fax")},
			stored:        &entity.Notification{ID: uuid.New(), Recipient: owner, Title: "Old", Content: "Old", Channel: entity.ChannelEmail},
			currentUserID: owner.String(),
			wantErrTarget: usecase.ErrUnsupportedChannel,
		},
		{
			name: "happy path",
			input: request.ResquestNotificationDTO{
				Title:   "Updated",
				Content: "Updated content",
				Channel: entity.ChannelSMS,
			},
			stored:        &entity.Notification{ID: uuid.New(), Recipient: owner, Title: "Old", Content: "Old", Channel: entity.ChannelEmail},
			currentUserID: owner.String(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, _, notifRepo, _ := newNotificationUC(t)
			ctx, cancel := testhelper.NewContextWithTimeout()
			defer cancel()

			id := uuid.New().String()

			if tt.wantErrTarget == usecase.ErrInvalidInput {
				got, err := uc.UpdateNotification(ctx, id, tt.currentUserID, tt.input)
				require.Nil(t, got)
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}

			notifRepo.EXPECT().FindByID(ctx, id).Return(tt.stored, tt.findErr).Once()

			callsUpdate := tt.findErr == nil &&
				tt.wantErrTarget != usecase.ErrForbidden &&
				tt.wantErrTarget != usecase.ErrUnsupportedChannel
			if callsUpdate {
				notifRepo.EXPECT().Update(ctx, mock.AnythingOfType("*entity.Notification")).Return(tt.updateErr).Once()
			}

			callsReload := callsUpdate && tt.updateErr == nil && tt.wantErrTarget == nil
			if callsReload {
				reloaded := &entity.Notification{
					ID:        tt.stored.ID,
					Recipient: tt.stored.Recipient,
					Title:     "Updated",
					Content:   "Updated content",
					Channel:   entity.ChannelSMS,
				}
				notifRepo.EXPECT().FindByID(ctx, id).Return(reloaded, tt.reloadErr).Once()
			}

			got, err := uc.UpdateNotification(ctx, id, tt.currentUserID, tt.input)
			if tt.wantErrTarget != nil {
				require.Nil(t, got)
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, "Updated", got.Title)
		})
	}
}

func TestNotificationUseCase_DeleteNotification(t *testing.T) {
	t.Parallel()

	owner := uuid.New()
	other := uuid.New()

	tests := []struct {
		name          string
		stored        *entity.Notification
		currentUserID string
		findErr       error
		deleteErr     error
		wantErrTarget error
	}{
		{
			name:          "happy path",
			stored:        &entity.Notification{ID: uuid.New(), Recipient: owner},
			currentUserID: owner.String(),
		},
		{
			name:          "forbidden when not owner",
			stored:        &entity.Notification{ID: uuid.New(), Recipient: owner},
			currentUserID: other.String(),
			wantErrTarget: usecase.ErrForbidden,
		},
		{
			name:          "not found on find",
			currentUserID: owner.String(),
			findErr:       gorm.ErrRecordNotFound,
			wantErrTarget: usecase.ErrNotFound,
		},
		{
			name:          "database error on delete",
			stored:        &entity.Notification{ID: uuid.New(), Recipient: owner},
			currentUserID: owner.String(),
			deleteErr:     errors.New("delete failed"),
			wantErrTarget: usecase.ErrDatabase,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, _, notifRepo, _ := newNotificationUC(t)
			ctx, cancel := testhelper.NewContextWithTimeout()
			defer cancel()

			id := "notif-id"
			notifRepo.EXPECT().FindByID(ctx, id).Return(tt.stored, tt.findErr).Once()
			if tt.findErr == nil && tt.wantErrTarget != usecase.ErrForbidden {
				notifRepo.EXPECT().Delete(ctx, id).Return(tt.deleteErr).Once()
			}

			err := uc.DeleteNotification(ctx, id, tt.currentUserID)
			if tt.wantErrTarget != nil {
				testhelper.AssertErrorIs(t, err, tt.wantErrTarget)
				return
			}
			require.NoError(t, err)
		})
	}
}
