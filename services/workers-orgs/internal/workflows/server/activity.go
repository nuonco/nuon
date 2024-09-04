package server

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"k8s.io/client-go/rest"

	"github.com/powertoolsdev/mono/pkg/deprecated/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
)

type Activities struct {
	v *validator.Validate
	namespaceCreator
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
		helmInstaller:              helm.NewInstaller(),
		waypointServerPinger:       &wpServerPinger{},
		waypointServerBootstrapper: &wpServerBootstrapper{},
		waypointProjectCreator:     &wpProjectCreator{},
	}
}

func (a *Activities) getKubeConfig(ctx context.Context, info *kube.ClusterInfo) (*rest.Config, error) {
	if a.Kubeconfig != nil {
		return a.Kubeconfig, nil
	}

	kCfg, err := kube.ConfigForCluster(ctx, info)
	if err != nil {
		return nil, fmt.Errorf("failed to get config for cluster: %w", err)
	}

	return kCfg, nil
}
