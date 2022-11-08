package provision

import (
	"k8s.io/client-go/rest"

	"github.com/powertoolsdev/go-sender"
	workers "github.com/powertoolsdev/workers-installs/internal"
)

// ProvisionActivities is a type that wraps the set of provision activities that we'll be using to execute this
// workflow. It should only be a few activities, such as running terraform and installing the agent
type ProvisionActivities struct {
	terraformProvisioner terraformProvisioner
	sender               sender.NotificationSender

	config workers.Config

	// this is exposed for testing and should not otherwise be used
	Kubeconfig *rest.Config

	// TODO(jm): refactor once we've finished all the waypoint setup work
	starter
	finisher
}

func NewProvisionActivities(cfg workers.Config, sender sender.NotificationSender) *ProvisionActivities {
	return &ProvisionActivities{
		terraformProvisioner: &tfProvisioner{},
		config:               cfg,
		sender:               sender,

		finisher: &finisherImpl{},
		starter:  &starterImpl{sender},
	}
}
