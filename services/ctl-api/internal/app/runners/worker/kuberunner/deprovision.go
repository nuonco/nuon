package runner

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
)

type DeprovisionRunnerRequest struct {
	RunnerID string `validate:"required"`
}

type DeprovisionRunnerResponse struct{}

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
// @id-template {{ .CallerID }}-deprovision-runner
func (w Wkflow) DeprovisionRunner(ctx workflow.Context, req DeprovisionRunnerRequest) (*executors.DeprovisionRunnerResponse, error) {
	clusterInfo := w.getClusterInfo()

	if _, err := AwaitUninstall(ctx, &UninstallRequest{
		ClusterInfo: clusterInfo,
		Namespace:   req.RunnerID,
		RunnerID:    req.RunnerID,
	}); err != nil {
		return nil, fmt.Errorf("unable to uninstall runner: %w", err)
	}

	return &executors.DeprovisionRunnerResponse{}, nil
}
