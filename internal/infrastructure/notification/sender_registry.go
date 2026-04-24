package notification

import (
	"fmt"

	"acl-challenge/internal/domain/entity"
	domainnotification "acl-challenge/internal/domain/notification"
	"acl-challenge/internal/usecase"
)

type SenderRegistry map[entity.Channel]domainnotification.Sender

func (r SenderRegistry) GetSender(channel entity.Channel) (domainnotification.Sender, error) {
	sender, exists := r[channel]
	if !exists {
		return nil, fmt.Errorf("sender registry: get sender: channel=%s: %w", channel, usecase.ErrUnsupportedChannel)
	}

	return sender, nil
}
