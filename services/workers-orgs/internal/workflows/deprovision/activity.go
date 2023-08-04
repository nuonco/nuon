package deprovision

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/deprecated/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
	"k8s.io/client-go/rest"
)

type Activities struct {
	namespaceDestroyer
	helmUninstaller uninstaller

	Kubeconfig *rest.Config
}

func NewActivities() *Activities {
	return &Activities{
		namespaceDestroyer: &nsDestroyer{},
		helmUninstaller:    helm.NewUninstaller(),
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
