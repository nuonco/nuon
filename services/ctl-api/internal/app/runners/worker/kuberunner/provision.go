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

type ProvisionRunnerRequestImage struct {
	URL string `validate:"required"`
	Tag string `validate:"required"`
}

type ProvisionRunnerRequest struct {
	RunnerID                 string                      `validate:"required"`
	APIURL                   string                      `validate:"required"`
	APIToken                 string                      `validate:"required"`
	Image                    ProvisionRunnerRequestImage `validate:"required"`
	RunnerIAMRole            string                      `validate:"required"`
	RunnerServiceAccountName string                      `validate:"required"`
}

type ProvisionRunnerResponse struct{}

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
// @id-template {{ .CallerID }}-provision-runner
func (w Wkflow) ProvisionRunner(ctx workflow.Context, req *ProvisionRunnerRequest) (*executors.ProvisionRunnerResponse, error) {
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
		return nil, fmt.Errorf("unable to provision runner: %w", err)
	}

	return &executors.ProvisionRunnerResponse{}, nil
}
