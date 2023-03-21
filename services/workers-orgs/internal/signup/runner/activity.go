package runner

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/helm"
	workers "github.com/powertoolsdev/mono/services/workers-orgs/internal"
	"k8s.io/client-go/rest"
)

type Activities struct {
	v *validator.Validate

	helmInstaller installer
	waypointServerCookieGetter
	waypointRunnerAdopter
	roleBindingCreator

	config     workers.Config
	Kubeconfig *rest.Config
}

func NewActivities(v *validator.Validate, cfg workers.Config) *Activities {
	return &Activities{
		v:                          v,
		waypointServerCookieGetter: &wpServerCookieGetter{},
		waypointRunnerAdopter:      &wpRunnerAdopter{},
		config:                     cfg,
		helmInstaller:              helm.NewInstaller(),
		roleBindingCreator:         &roleBindingCreatorImpl{},
	}
}
