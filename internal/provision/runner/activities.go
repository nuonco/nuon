package runner

import (
	"github.com/powertoolsdev/go-helm"
	"github.com/powertoolsdev/go-waypoint"
	"k8s.io/client-go/rest"

	workers "github.com/powertoolsdev/workers-installs/internal"
)

// Activities is a type that wraps the set of provision activities that we'll be using to execute this
// workflow. It should only be a few activities, such as running terraform and installing the agent
type Activities struct {
	helmInstaller installer

	// this is exposed for testing and should not otherwise be used
	Kubeconfig *rest.Config

	// TODO(jm): refactor once we've finished all the waypoint setup work
	waypointProvider waypoint.Provider
	waypointProjectCreator
	waypointServerCookieGetter
	waypointRunnerAdopter
	waypointWorkspaceCreator
	roleBindingCreator
	waypointRunnerProfileCreator
}

func NewActivities(cfg workers.Config) *Activities {
	return &Activities{
		helmInstaller: helm.NewInstaller(),

		waypointProvider:             waypoint.NewProvider(),
		waypointProjectCreator:       &wpProjectCreator{},
		waypointServerCookieGetter:   &wpServerCookieGetter{},
		waypointRunnerAdopter:        &wpRunnerAdopter{},
		waypointWorkspaceCreator:     &wpWorkspaceCreator{},
		waypointRunnerProfileCreator: &wpRunnerProfileCreator{},
		roleBindingCreator:           &roleBindingCreatorImpl{},
	}
}
