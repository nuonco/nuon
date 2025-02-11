package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker/activities"
)

func (w *Workflows) restart(ctx workflow.Context, _ signals.RequestSignal) error {
	w.startReconcileEventLoopsWorkflowCron(ctx, activities.EnsureEventLoopsRequest{})
	if _, err := activities.AwaitEnsureEventLoop(ctx, activities.EnsureEventLoopsRequest{}); err != nil {
		return fmt.Errorf("unable to start EnsureEventLoop: %w", err)
	}
	return nil
}
