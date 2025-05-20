package runner

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/workflows/types/executors"
)

func (w wkflow) DeprovisionRunner(ctx workflow.Context, req executors.DeprovisionRunnerRequest) (*executors.DeprovisionRunnerResponse, error) {
	clusterInfo := w.getClusterInfo()

	if _, err := AwaitUninstall(ctx, &UninstallRequest{
		ClusterInfo: clusterInfo,
		Namespace:   req.RunnerID,
                RunnerID: req.RunnerID,
	}); err != nil {
		return nil, fmt.Errorf("unable to uninstall runner: %w", err)
	}

	return &executors.DeprovisionRunnerResponse{}, nil
}
