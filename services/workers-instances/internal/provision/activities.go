package provision

import (
	"github.com/powertoolsdev/go-waypoint"
	"github.com/powertoolsdev/mono/pkg/sender"
)

type Activities struct {
	waypointProvider waypoint.Provider
	hostnameNotificationSender
}

func NewActivities(sender sender.NotificationSender) *Activities {
	return &Activities{
		waypointProvider: waypoint.NewProvider(),
		hostnameNotificationSender: &hostnameNotificationSenderImpl{
			sender: sender,
		},
	}
}
