package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
)

func (w *Workflows) delete(ctx workflow.Context, runnerID string, sandboxMode, force bool) error {
	err := w.deprovision(ctx, runnerID, sandboxMode)
	if err != nil {
		if !force {
			return fmt.Errorf("unable to deprovision runner: %w", err)
		}
	}

	if err := activities.AwaitDelete(ctx, activities.DeleteRequest{
		RunnerID: runnerID,
	}); err != nil {
		return fmt.Errorf("unable to delete runner: %w", err)
	}

	return nil
}
