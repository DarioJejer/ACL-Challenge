package notification

import (
	"context"

	"acl-challenge/internal/domain/entity"
)

type Sender interface {
	Send(ctx context.Context, n *entity.Notification) error
}
