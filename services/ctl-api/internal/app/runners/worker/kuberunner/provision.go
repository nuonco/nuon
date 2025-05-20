package runner

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
)

const (
	defaultHelmOperationTimeout time.Duration = time.Minute * 10
)

// @temporal-gen workflow
// @id-template {{ .CallerID }}-provision-runner
// @execution-timeout 30m
// @task-timeout 15m
func (w wkflow) ProvisionRunner(ctx workflow.Context, req executors.ProvisionRunnerRequest) (*executors.ProvisionRunnerResponse, error) {
	clusterInfo := w.getClusterInfo()

	if _, err := AwaitInstallOrUpgrade(ctx, &InstallOrUpgradeRequest{
		ClusterInfo: clusterInfo,
		Image:       req.Image,

		Namespace:                req.RunnerID,
		Timeout:                  defaultHelmOperationTimeout,
		RunnerID:                 req.RunnerID,
		RunnerServiceAccountName: req.RunnerServiceAccountName,
		RunnerIAMRole:            req.RunnerIAMRole,
		APIToken:                 req.APIToken,
		APIURL:                   req.APIURL,
	}); err != nil {
		return nil, fmt.Errorf("unable to uninstall runner: %w", err)
	}

	return &executors.ProvisionRunnerResponse{}, nil
}
