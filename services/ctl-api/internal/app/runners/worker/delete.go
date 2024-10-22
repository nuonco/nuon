package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
)

// @temporal-gen workflow
// @execution-timeout 60m
// @task-timeout 30m
func (w *Workflows) Delete(ctx workflow.Context, sreq signals.RequestSignal) error {
	err := w.Deprovision(ctx, sreq)
	if err != nil {
		return fmt.Errorf("unable to deprovision runner: %w", err)
	}

	if err := activities.AwaitDelete(ctx, activities.DeleteRequest{
		RunnerID: sreq.ID,
	}); err != nil {
		return fmt.Errorf("unable to delete runner: %w", err)
	}

	return nil
}
