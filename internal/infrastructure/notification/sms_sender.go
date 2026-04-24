package notification

import (
	"context"
	"log/slog"

	"acl-challenge/internal/domain/entity"
)

type SMSSender struct{}

func (s *SMSSender) Send(ctx context.Context, n *entity.Notification) error {
	_ = ctx
	slog.Info("[stub] sending sms notification",
		"id", n.ID,
		"channel", n.Channel,
	)
	return nil
}
