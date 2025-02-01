package helm

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/rest"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
)

func (h *handler) actionInit(ctx context.Context, l *zap.Logger) (*action.Configuration, *rest.Config, error) {
	if os.Getenv("IS_NUONCTL") == "true" {
		l.Info("local runner using helm, so hard coding EKS creds for client",
			zap.Any("cluster-info", h.state.cfg.ClusterInfo))

		kubeCfg, err := kube.ConfigForCluster(ctx, h.state.cfg.ClusterInfo)
		if err != nil {
			return nil, nil, errors.Wrap(err, "unable to get kube config")
		}

		helmCfg, err := helm.Client(l, kubeCfg, h.state.cfg.Namespace)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to get helm client: %w", err)
		}

		return helmCfg, kubeCfg, nil
	}

	kubeCfg, err := kube.GetKubeConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get in-cluster config: %w", err)
	}

	helmCfg, err := helm.Client(l, kubeCfg, h.state.cfg.Namespace)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get helm client: %w", err)
	}

	return helmCfg, kubeCfg, nil
}
