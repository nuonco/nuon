package provision

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/sender"
)

type Activities struct {
	v *validator.Validate
	hostnameNotificationSender
}

func NewActivities(v *validator.Validate, sender sender.NotificationSender) *Activities {
	return &Activities{
		v: v,
		hostnameNotificationSender: &hostnameNotificationSenderImpl{
			sender: sender,
		},
	}
}
