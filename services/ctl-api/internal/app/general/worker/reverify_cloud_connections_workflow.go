package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func (w *Workflows) ReverifyCloudConnections(ctx workflow.Context) error {
	response, err := activities.AwaitReverifyCloudConnections(ctx, activities.ReverifyCloudConnectionsRequest{})
	if err != nil {
		return err
	}
	if logger, err := log.WorkflowLogger(ctx); err == nil {
		logger.Info("reverified cloud connections", zap.Int("probed", response.Probed), zap.Int("oidc", response.OIDC), zap.Int("legacy", response.Legacy), zap.Int("failures", response.Failures))
	}
	return nil
}
