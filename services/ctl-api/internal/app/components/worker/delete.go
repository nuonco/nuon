package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) delete(ctx workflow.Context, componentID string, dryRun bool) error {
	if err := w.defaultExecErrorActivity(ctx, w.acts.Delete, activities.DeleteRequest{
		ComponentID: componentID,
	}); err != nil {
		return fmt.Errorf("unable to delete component: %w", err)
	}

	return nil
}
