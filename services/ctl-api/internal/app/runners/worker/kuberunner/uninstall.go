package runner

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
)

type UninstallRequest struct {
	ClusterInfo *kube.ClusterInfo

	Namespace string
	RunnerID  string
}

type UninstallResponse struct{}

// @temporal-gen activity
// @schedule-to-close-timeout 10m
// @max-retries 5
func (a *Activities) Uninstall(ctx context.Context, req *UninstallRequest) (*InstallOrUpgradeResponse, error) {
	kubeCfg, err := kube.ConfigForCluster(ctx, req.ClusterInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to get config for cluster: %w", err)
	}

	l := zap.L()
	helmCfg, err := helm.Client(l, kubeCfg, req.Namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get helm config: %w", err)
	}

	if err := a.uninstall(ctx, helmCfg, req.RunnerID); err != nil {
		return nil, fmt.Errorf("unable to uninstall helm chart: %w", err)
	}

	return nil, nil
}
