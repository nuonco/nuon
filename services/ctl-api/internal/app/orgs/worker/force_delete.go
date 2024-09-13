package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// @temporal-gen workflow
// @execution-timeout 30m
// @task-timeout 15m
func (w *Workflows) ForceDelete(ctx workflow.Context, sreq signals.RequestSignal) error {
	l := workflow.GetLogger(ctx)
	if err := w.Deprovision(ctx, sreq); err != nil {
		l.Error("unable to deprovision org: %w", zap.Error(err))
	}

	if err := activities.AwaitDeleteByOrgID(ctx, sreq.ID); err != nil {
		l.Error("unable to delete org: %w", zap.Error(err))
	}
	return nil
}
