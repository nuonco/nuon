package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

func (w *Workflows) updateSandbox(ctx workflow.Context, appID, sandboxReleaseID string, dryRun bool) error {
	l := workflow.GetLogger(ctx)
	l.Info("updating sandbox release", zap.String("app-id", appID), zap.String("release-id", sandboxReleaseID))
	return nil
}
