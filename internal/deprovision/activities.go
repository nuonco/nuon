package deprovision

import (
	"context"
)

// notificationSender defines the interface for sending messages
// There are implementation in the `sender` package but others may be created trivially.
type notificationSender interface {
	Send(context.Context, string) error
}

type Activities struct {
	sender notificationSender
	starter
	finisher
}

func NewActivities(sender notificationSender) *Activities {
	return &Activities{
		sender:   sender,
		starter:  &starterImpl{},
		finisher: &finisherImpl{},
	}
}
