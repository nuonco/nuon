package server

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/helm"
	"k8s.io/client-go/rest"
)

type Activities struct {
	v *validator.Validate
	namespaceCreator
	serviceCreator
	helmInstaller installer
	waypointServerPinger
	waypointServerBootstrapper
	waypointProjectCreator

	Kubeconfig *rest.Config
}

func NewActivities(v *validator.Validate) *Activities {
	return &Activities{
		v: v,

		namespaceCreator:           &nsCreator{},
		serviceCreator:             &svcCreator{},
		helmInstaller:              helm.NewInstaller(),
		waypointServerPinger:       &wpServerPinger{},
		waypointServerBootstrapper: &wpServerBootstrapper{},
		waypointProjectCreator:     &wpProjectCreator{},
	}
}
