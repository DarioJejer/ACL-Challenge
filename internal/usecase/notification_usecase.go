package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"acl-challenge/internal/api/dtos/request"
	"acl-challenge/internal/domain/entity"
	domainnotification "acl-challenge/internal/domain/notification"
	"acl-challenge/internal/domain/repository"

	"github.com/google/uuid"
)

type NotificationUseCase struct {
	userRepo         repository.UserRepository
	notificationRepo repository.NotificationRepository
	senders          domainnotification.SenderRegistry
}

func NewNotificationUseCase(
	userRepo repository.UserRepository,
	notificationRepo repository.NotificationRepository,
	senders domainnotification.SenderRegistry,
) *NotificationUseCase {
	return &NotificationUseCase{
		userRepo:         userRepo,
		notificationRepo: notificationRepo,
		senders:          senders,
	}
}

// ListNotifications returns the notifications addressed to currentUserID.
// The query is already scoped by user, so no separate ownership check runs.
func (uc *NotificationUseCase) ListNotifications(ctx context.Context, currentUserID string) ([]entity.Notification, error) {
	notifications, err := uc.notificationRepo.FindAllByUserID(ctx, currentUserID)
	if err != nil {
		return nil, mapRepositoryError("usecase: notification: list", err)
	}
	return notifications, nil
}

// GetNotification fetches a notification by id and enforces that the caller
// owns it. Non-owners receive ErrForbidden, identical to a 403.
func (uc *NotificationUseCase) GetNotification(ctx context.Context, id, currentUserID string) (*entity.Notification, error) {
	notification, err := uc.notificationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepositoryError("usecase: notification: get", err)
	}
	if err := assertOwner(notification, currentUserID); err != nil {
		return nil, err
	}
	return notification, nil
}

// CreateNotification creates a notification owned by currentUserID. The
// recipient is always the authenticated caller — the request body cannot
// override it.
func (uc *NotificationUseCase) CreateNotification(ctx context.Context, currentUserID string, input request.ResquestNotificationDTO) (*entity.Notification, error) {
	if strings.TrimSpace(currentUserID) == "" {
		return nil, fmt.Errorf("usecase: notification: create: missing current user: %w", ErrUnauthorized)
	}
	if input.Channel == "" || strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Content) == "" {
		return nil, fmt.Errorf("usecase: notification: create: invalid input: %w", ErrInvalidInput)
	}

	ownerID, err := uuid.Parse(currentUserID)
	if err != nil {
		return nil, fmt.Errorf("usecase: notification: create: invalid user id: %w", ErrInvalidInput)
	}

	sender, err := uc.senders.GetSender(input.Channel)
	if err != nil {
		if errors.Is(err, ErrUnsupportedChannel) {
			return nil, fmt.Errorf("usecase: notification: create: sender registry: %v: %w", err, ErrUnsupportedChannel)
		}
		return nil, fmt.Errorf("usecase: notification: create: sender registry: %v: %w", err, ErrDatabase)
	}

	notification := &entity.Notification{
		Recipient: ownerID,
		Title:     input.Title,
		Content:   input.Content,
		Channel:   input.Channel,
	}

	if err := uc.notificationRepo.Create(ctx, notification); err != nil {
		return nil, mapRepositoryError("usecase: notification: create", err)
	}

	if err := sender.Send(ctx, notification); err != nil {
		slog.Warn(
			"notification dispatch failed",
			"notification_id", notification.ID,
			"channel", notification.Channel,
			"error", err.Error(),
		)
	}

	return notification, nil
}

// UpdateNotification fetches a notification by id, enforces ownership, then
// applies the supplied field updates. The recipient (owner) cannot be changed.
func (uc *NotificationUseCase) UpdateNotification(ctx context.Context, id, currentUserID string, input request.ResquestNotificationDTO) (*entity.Notification, error) {
	if strings.TrimSpace(input.Title) == "" &&
		strings.TrimSpace(input.Content) == "" &&
		strings.TrimSpace(string(input.Channel)) == "" {
		return nil, ErrInvalidInput
	}

	notification, err := uc.notificationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepositoryError("usecase: notification: update: find", err)
	}
	if err := assertOwner(notification, currentUserID); err != nil {
		return nil, err
	}

	if strings.TrimSpace(input.Title) != "" {
		notification.Title = input.Title
	}
	if strings.TrimSpace(input.Content) != "" {
		notification.Content = input.Content
	}
	if strings.TrimSpace(string(input.Channel)) != "" {
		if !entity.IsSupportedChannel(input.Channel) {
			return nil, fmt.Errorf("usecase: notification: update: invalid channel: %w", ErrUnsupportedChannel)
		}
		notification.Channel = input.Channel
	}

	if err := uc.notificationRepo.Update(ctx, notification); err != nil {
		return nil, mapRepositoryError("usecase: notification: update", err)
	}

	updated, err := uc.notificationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepositoryError("usecase: notification: update: reload", err)
	}
	return updated, nil
}

// DeleteNotification fetches a notification, enforces ownership, then deletes.
func (uc *NotificationUseCase) DeleteNotification(ctx context.Context, id, currentUserID string) error {
	notification, err := uc.notificationRepo.FindByID(ctx, id)
	if err != nil {
		return mapRepositoryError("usecase: notification: delete: find", err)
	}
	if err := assertOwner(notification, currentUserID); err != nil {
		return err
	}

	if err := uc.notificationRepo.Delete(ctx, id); err != nil {
		return mapRepositoryError("usecase: notification: delete", err)
	}
	return nil
}

// assertOwner returns ErrForbidden when the notification does not belong to
// the supplied user. The recipient field is treated as the owner.
func assertOwner(notification *entity.Notification, currentUserID string) error {
	if notification == nil {
		return fmt.Errorf("usecase: notification: ownership: nil notification: %w", ErrInternalServer)
	}
	if strings.TrimSpace(currentUserID) == "" {
		return fmt.Errorf("usecase: notification: ownership: missing current user: %w", ErrUnauthorized)
	}
	if notification.Recipient.String() != currentUserID {
		return fmt.Errorf("usecase: notification: ownership mismatch: %w", ErrForbidden)
	}
	return nil
}
