package worker

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// @temporal-gen workflow
func (w *Workflows) UpdateSandbox(ctx workflow.Context, sreq signals.RequestSignal) error {
	l := workflow.GetLogger(ctx)
	// NOTE(sdboyer) note that this is whole behavior is a no-op right now and the signal can't carry a release-id, so we print an empty string
	l.Info("updating sandbox release", zap.String("app-id", sreq.ID), zap.String("release-id", ""))
	return nil
}
