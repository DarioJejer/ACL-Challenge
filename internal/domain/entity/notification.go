package entity

import "time"

type Channel string

const (
	ChannelEmail            Channel = "email"
	ChannelSMS              Channel = "sms"
	ChannelPushNotification Channel = "push_notification"
)

// Notification is the canonical domain representation of a user notification.
type Notification struct {
	ID        string
	Recipient string
	Title     string
	Content   string
	Channel   Channel
	CreatedAt time.Time
	UpdatedAt time.Time
}
