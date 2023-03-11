package teardown

import (
	"github.com/powertoolsdev/mono/pkg/helm"
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
