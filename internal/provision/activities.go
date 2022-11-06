package provision

import (
	"k8s.io/client-go/rest"

	"github.com/powertoolsdev/go-helm"
	"github.com/powertoolsdev/go-sender"
	"github.com/powertoolsdev/go-waypoint"
	workers "github.com/powertoolsdev/workers-installs/internal"
)

// NOTE: we alias this type so it doesn't conflict
type waypointProvider = waypoint.Provider

// ProvisionActivities is a type that wraps the set of provision activities that we'll be using to execute this
// workflow. It should only be a few activities, such as running terraform and installing the agent
type ProvisionActivities struct {
	terraformProvisioner terraformProvisioner
	helmInstaller        installer
	sender               sender.NotificationSender

	config workers.Config

	// this is exposed for testing and should not otherwise be used
	Kubeconfig *rest.Config

	// TODO(jm): refactor once we've finished all the waypoint setup work
	waypointProvider
	waypointProjectCreator
	waypointServerCookieGetter
	waypointRunnerAdopter
	waypointWorkspaceCreator
	roleBindingCreator
	waypointRunnerProfileCreator
	starter
	finisher
}

func NewProvisionActivities(cfg workers.Config, sender sender.NotificationSender) *ProvisionActivities {
	return &ProvisionActivities{
		terraformProvisioner: &tfProvisioner{},
		helmInstaller:        helm.NewInstaller(),
		config:               cfg,
		sender:               sender,

		// TODO(jm): refactor this once all runner activities are running
		waypointProvider:             waypoint.NewProvider(),
		waypointProjectCreator:       &wpProjectCreator{},
		waypointServerCookieGetter:   &wpServerCookieGetter{},
		waypointRunnerAdopter:        &wpRunnerAdopter{},
		waypointWorkspaceCreator:     &wpWorkspaceCreator{},
		waypointRunnerProfileCreator: &wpRunnerProfileCreator{},
		finisher:                     &finisherImpl{},
		starter:                      &starterImpl{sender},
		roleBindingCreator:           &roleBindingCreatorImpl{},
	}
}
