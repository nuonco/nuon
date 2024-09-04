package helm

import (
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"

	"github.com/powertoolsdev/mono/pkg/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
)

func (h *handler) actionInit(log *zap.Logger) (*action.Configuration, error) {
	kubeCfg, err := kube.GetKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get in-cluster config: %w", err)
	}

	helmCfg, err := helm.Client(h.log, kubeCfg, h.state.cfg.Namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get helm client: %w", err)
	}

	return helmCfg, nil
}
