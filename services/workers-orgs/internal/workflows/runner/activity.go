package runner

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/deprecated/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
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

func (a *Activities) getKubeConfig(info *kube.ClusterInfo) (*rest.Config, error) {
	if a.Kubeconfig != nil {
		return a.Kubeconfig, nil
	}

	kCfg, err := kube.ConfigForCluster(info)
	if err != nil {
		return nil, fmt.Errorf("failed to get config for cluster: %w", err)
	}

	return kCfg, nil
}
