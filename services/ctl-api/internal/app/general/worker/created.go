package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker/activities"
)

func (w *Workflows) created(ctx workflow.Context, _ signals.RequestSignal) error {
	// our logic goes here
	w.startReconcileEventLoopsWorkflowCron(ctx, activities.EnsureEventLoopsRequest{})
	return nil
}
