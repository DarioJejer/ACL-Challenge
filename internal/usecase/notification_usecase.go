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

func (uc *NotificationUseCase) ListNotifications(ctx context.Context, userID string) ([]entity.Notification, error) {
	notifications, err := uc.notificationRepo.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, mapRepositoryError("usecase: notification: list", err)
	}
	return notifications, nil
}

func (uc *NotificationUseCase) GetNotification(ctx context.Context, id string) (*entity.Notification, error) {
	notification, err := uc.notificationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepositoryError("usecase: notification: get", err)
	}
	return notification, nil
}

func (uc *NotificationUseCase) CreateNotification(ctx context.Context, input request.ResquestNotificationDTO) (*entity.Notification, error) {
	if strings.TrimSpace(input.Recipient) == "" || input.Channel == "" || strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Content) == "" {
		return nil, fmt.Errorf("usecase: notification: create: invalid input: %w", ErrInvalidInput)
	}

	_, err := uc.userRepo.FindByID(ctx, input.Recipient)
	if err != nil {
		return nil, mapRepositoryError("usecase: notification: create: find user", err)
	}

	sender, err := uc.senders.GetSender(input.Channel)
	if err != nil {
		if errors.Is(err, ErrUnsupportedChannel) {
			return nil, fmt.Errorf("usecase: notification: create: sender registry: %v: %w", err, ErrUnsupportedChannel)
		}
		return nil, fmt.Errorf("usecase: notification: create: sender registry: %v: %w", err, ErrDatabase)
	}

	notification := &entity.Notification{
		Recipient: uuid.MustParse(input.Recipient),
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

func (uc *NotificationUseCase) UpdateNotification(ctx context.Context, id string, input request.ResquestNotificationDTO) (*entity.Notification, error) {
	if strings.TrimSpace(input.Recipient) == "" &&
		strings.TrimSpace(input.Title) == "" &&
		strings.TrimSpace(input.Content) == "" &&
		strings.TrimSpace(string(input.Channel)) == "" {
		return nil, ErrInvalidInput
	}

	notification, err := uc.notificationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, mapRepositoryError("usecase: notification: update: find", err)
	}

	if strings.TrimSpace(input.Recipient) != "" {
		_, err = uc.userRepo.FindByID(ctx, input.Recipient)
		if err != nil {
			return nil, mapRepositoryError("usecase: notification: update: find user", err)
		}
		notification.Recipient = uuid.MustParse(input.Recipient)
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

func (uc *NotificationUseCase) DeleteNotification(ctx context.Context, id string) error {
	if err := uc.notificationRepo.Delete(ctx, id); err != nil {
		return mapRepositoryError("usecase: notification: delete", err)
	}
	return nil
}
