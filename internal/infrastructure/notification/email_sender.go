package notification

import (
	"context"
	"log/slog"

	"acl-challenge/internal/domain/entity"
)

type EmailSender struct{}

func (s *EmailSender) Send(ctx context.Context, n *entity.Notification) error {
	_ = ctx
	slog.Info("[stub] sending email notification",
		"id", n.ID,
		"channel", n.Channel,
	)
	return nil
}
