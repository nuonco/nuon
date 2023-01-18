package provision

import (
	"github.com/powertoolsdev/go-sender"
	"github.com/powertoolsdev/go-waypoint"
)

// NOTE(jm): we alias this type so it doesn't conflict
type waypointProvider = waypoint.Provider

type Activities struct {
	waypointProvider
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
