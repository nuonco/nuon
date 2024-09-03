package worker

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker/activities"
	"go.temporal.io/sdk/workflow"
)

func (w *Workflows) restart(ctx workflow.Context) error {
	w.startReconcileEventLoopsWorkflowCron(ctx, activities.EnsureEventLoopsRequest{})
	if _, err := activities.AwaitEnsureEventLoop(ctx, activities.EnsureEventLoopsRequest{}); err != nil {
		return fmt.Errorf("unable to start EnsureEventLoop: %w", err)
	}
	return nil
}
