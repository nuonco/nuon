package runner

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/release"

	"github.com/powertoolsdev/mono/pkg/helm"
	"github.com/powertoolsdev/mono/pkg/kube"
	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
)

type InstallOrUpgradeRequest struct {
	ClusterInfo *kube.ClusterInfo
	Image       executors.ProvisionRunnerRequestImage

	Namespace                string
	Timeout                  time.Duration
	RunnerID                 string
	RunnerIAMRole            string
	RunnerServiceAccountName string

	APIURL                 string
	APIToken               string
	SettingsRefreshTimeout time.Duration
}

type InstallOrUpgradeResponse struct{}

// @temporal-gen activity
// @schedule-to-close-timeout 10m
// @start-to-close-timeout 10m
// @max-retries 5
func (a *Activities) InstallOrUpgrade(ctx context.Context, req *InstallOrUpgradeRequest) (*InstallOrUpgradeResponse, error) {
	l := zap.L()
	l.Error("aws-cfg", zap.Any("cfg", req.ClusterInfo))

	kubeCfg, err := kube.ConfigForCluster(ctx, req.ClusterInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to get config for cluster: %w", err)
	}

	helmClient, err := helm.Client(l, kubeCfg, req.Namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to get helm config: %w", err)
	}

	releaseName := fmt.Sprintf("runner-%s", req.RunnerID)
	prevRel, err := helm.GetRelease(helmClient, releaseName)
	if err != nil {
		return nil, fmt.Errorf("unable to get previous helm release: %w", err)
	}

	var (
		rel *release.Release
		op  string
	)
	if prevRel == nil {
		op = "install"
		rel, err = a.install(ctx, helmClient, req)
	} else {
		op = "upgrade"
		rel, err = a.upgrade(ctx, helmClient, req)
	}
	if err != nil {
		return nil, fmt.Errorf("error on %s: %w", op, err)
	}

	l.Info("helm release", zap.Int("release", rel.Version))
	return &InstallOrUpgradeResponse{}, nil
}
