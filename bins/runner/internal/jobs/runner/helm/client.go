package helm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/kubernetes"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
)

func (h *handler) getKubeConfig(ctx context.Context) (*kubernetes.Clientset, error) {
	kubeCfg, err := kube.ConfigForCluster(ctx, h.state.cfg.ClusterInfo)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get kube config")
	}

	client, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get kube client")
	}

	return client, nil
}

func (h *handler) actionInit(ctx context.Context, log *zap.Logger) (*action.Configuration, error) {
	kubeCfg, err := kube.ConfigForCluster(ctx, h.state.cfg.ClusterInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to get kube config: %w", err)
	}

	helmClient, err := helm.Client(log, kubeCfg, h.state.cfg.Namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get helm client: %w", err)
	}

	return helmClient, nil
}
