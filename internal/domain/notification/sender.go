package notification

import (
	"context"

	"acl-challenge/internal/domain/entity"
)

//go:generate mockery --config ../../../.mockery.yaml
type Sender interface {
	Send(ctx context.Context, n *entity.Notification) error
}
