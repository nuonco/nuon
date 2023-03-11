package server

import (
	"github.com/powertoolsdev/go-waypoint"
	"github.com/powertoolsdev/mono/pkg/helm"
	"k8s.io/client-go/rest"
)

// NOTE(jm): we alias this type here so that it embeds as WaypointProvider, but allows us to have waypoint.Provider as
// the actual package
type WaypointProvider = waypoint.Provider

type Activities struct {
	namespaceCreator
	serviceCreator
	helmInstaller    installer
	waypointProvider WaypointProvider
	waypointServerPinger
	waypointServerBootstrapper
	waypointProjectCreator

	Kubeconfig *rest.Config
}

func NewActivities() *Activities {
	return &Activities{
		namespaceCreator:           &nsCreator{},
		serviceCreator:             &svcCreator{},
		helmInstaller:              helm.NewInstaller(),
		waypointProvider:           waypoint.NewProvider(),
		waypointServerPinger:       &wpServerPinger{},
		waypointServerBootstrapper: &wpServerBootstrapper{},
		waypointProjectCreator:     &wpProjectCreator{},
	}
}
