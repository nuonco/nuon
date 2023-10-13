package deprovision

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/workers-installs/internal"
)

// notificationSender defines the interface for sending messages
// There are implementation in the `sender` package but others may be created trivially.
type notificationSender interface {
	Send(context.Context, string) error
}

type Activities struct {
	v      *validator.Validate
	cfg    *internal.Config
	sender notificationSender
	starter
	finisher
}

func NewActivities(v *validator.Validate, sender notificationSender, cfg *internal.Config) *Activities {
	return &Activities{
		v:        v,
		cfg:      cfg,
		sender:   sender,
		starter:  &starterImpl{},
		finisher: &finisherImpl{},
	}
}
