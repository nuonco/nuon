package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) Forget(ctx workflow.Context, sreq signals.RequestSignal) error {
	installID := sreq.ID
	if err := activities.AwaitDeleteByInstallID(ctx, installID); err != nil {
		return fmt.Errorf("unable to delete install: %w", err)
	}

	return nil
}
