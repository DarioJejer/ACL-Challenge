package entity

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

// Notification is the canonical domain representation of a user notification.
type Notification struct {
	ID        uuid.UUID
	Recipient uuid.UUID
	Title     string
	Content   string
	Channel   Channel
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Channel string

const (
	ChannelEmail            Channel = "email"
	ChannelSMS              Channel = "sms"
	ChannelPushNotification Channel = "push_notification"
)

var ValidChannels = []Channel{
	ChannelEmail,
	ChannelSMS,
	ChannelPushNotification,
}

func IsSupportedChannel(channel Channel) bool {
	return slices.Contains(ValidChannels, channel)
}
