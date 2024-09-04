package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"

	"github.com/powertoolsdev/mono/pkg/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
)

func (h *handler) actionInit(ctx context.Context, log *zap.Logger) (*action.Configuration, error) {
	kubeCfg, err := kube.ConfigForCluster(ctx, h.state.cfg.ClusterInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to get kube config: %w", err)
	}

	helmClient, err := helm.Client(h.log, kubeCfg, h.state.cfg.Namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get helm client: %w", err)
	}

	return helmClient, nil
}
