package sender

import "context"

type NotificationSender interface {
	Send(context.Context, string) error
}
