package notification

import "acl-challenge/internal/domain/entity"

type SenderRegistry interface {
	GetSender(channel entity.Channel) (Sender, error)
}
