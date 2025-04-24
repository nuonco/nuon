package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/rest"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
)

func (h *handler) actionInit(ctx context.Context, l *zap.Logger) (*action.Configuration, *rest.Config, error) {
	kubeCfg, err := kube.ConfigForCluster(ctx, h.state.plan.HelmDeployPlan.ClusterInfo)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get kube config")
	}

	helmCfg, err := helm.Client(l, kubeCfg, h.state.plan.HelmDeployPlan.Namespace)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get helm client: %w", err)
	}

	return helmCfg, kubeCfg, nil
}
