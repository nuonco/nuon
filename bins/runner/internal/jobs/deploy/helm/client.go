package helm

import (
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"

	"github.com/powertoolsdev/mono/pkg/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
)

func (h *handler) actionInit(l *zap.Logger) (*action.Configuration, error) {
	kubeCfg, err := kube.GetKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get in-cluster config: %w", err)
	}

	helmCfg, err := helm.Client(l, kubeCfg, h.state.cfg.Namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get helm client: %w", err)
	}

	return helmCfg, nil
}
