package runner

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/deprecated/helm"
	"k8s.io/client-go/rest"

	workers "github.com/powertoolsdev/mono/services/workers-installs/internal"
)

// Activities is a type that wraps the set of provision activities that we'll be using to execute this
// workflow. It should only be a few activities, such as running terraform and installing the agent
type Activities struct {
	v             *validator.Validate
	helmInstaller installer
	cfg           workers.Config

	// this is exposed for testing and should not otherwise be used
	Kubeconfig *rest.Config

	waypointProjectCreator
	waypointServerCookieGetter
	waypointRunnerAdopter
	waypointWorkspaceCreator
	roleBindingCreator
	waypointRunnerProfileCreator
}

func NewActivities(v *validator.Validate, cfg workers.Config) *Activities {
	return &Activities{
		v:                            v,
		cfg:                          cfg,
		helmInstaller:                helm.NewInstaller(),
		waypointProjectCreator:       &wpProjectCreator{},
		waypointServerCookieGetter:   &wpServerCookieGetter{},
		waypointRunnerAdopter:        &wpRunnerAdopter{},
		waypointWorkspaceCreator:     &wpWorkspaceCreator{},
		waypointRunnerProfileCreator: &wpRunnerProfileCreator{},
		roleBindingCreator:           &roleBindingCreatorImpl{},
	}
}
