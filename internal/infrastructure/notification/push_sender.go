package notification

import (
	"context"
	"log/slog"

	"acl-challenge/internal/domain/entity"
)

type PushSender struct{}

func (s *PushSender) Send(ctx context.Context, n *entity.Notification) error {
	_ = ctx
	slog.Info("[stub] sending push notification",
		"id", n.ID,
		"channel", n.Channel,
	)
	return nil
}
