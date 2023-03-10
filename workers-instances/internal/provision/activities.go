package provision

import (
	"github.com/powertoolsdev/go-sender"
	"github.com/powertoolsdev/go-waypoint"
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
